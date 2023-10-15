package server

import (
	"context"
	"encoding/json"
	"errors"
	"exchanges/pkg/exchange"
	"github.com/gofiber/fiber/v2"
	"log"
	"sort"
)

func (s *Server) init() {
	cfg := fiber.Config{
		DisableStartupMessage: true,
		StrictRouting:         true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			var e *fiber.Error

			if errors.As(err, &e) {
				code = e.Code
			}

			if code == fiber.StatusInternalServerError {
				log.Printf("%v [path: %s]", err, c.Path())
			}

			return c.Status(code).Send(nil)
		},
	}

	engine := fiber.New(cfg)

	engine.Get("/exchanges", func(c *fiber.Ctx) error {
		s.mu.Lock()
		defer s.mu.Unlock()

		var list []string

		for exchangeID := range s.exchanges {
			list = append(list, exchangeID)
		}

		sort.Slice(list, func(i, j int) bool {
			return list[i] < list[j]
		})

		rsp, err := json.Marshal(list)
		if err != nil {
			return err
		}

		c.Set("Content-Type", "application/json")

		return c.Status(fiber.StatusOK).Send(rsp)
	})

	engine.Get("/:exchangeID/pairs", func(c *fiber.Ctx) error {
		obj := func() exchange.Exchange {
			s.mu.Lock()
			defer s.mu.Unlock()

			obj, _ := s.exchanges[c.Params("exchangeID")]

			return obj
		}()

		if obj == nil {
			return fiber.ErrNotFound
		}

		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		defer cancel()

		rsp, err := obj.GetPairs(ctx)
		if err != nil {
			return err
		}

		sort.Slice(rsp, func(i, j int) bool {
			return rsp[i].Id < rsp[j].Id
		})

		return c.Status(fiber.StatusOK).JSON(rsp)
	})

	engine.Get("/:exchangeID/orderbook/:pairID", func(c *fiber.Ctx) error {
		obj := func() exchange.Exchange {
			s.mu.Lock()
			defer s.mu.Unlock()

			obj, _ := s.exchanges[c.Params("exchangeID")]

			return obj
		}()

		if obj == nil {
			return fiber.ErrNotFound
		}

		ctx, cancel := context.WithTimeout(context.Background(), reqTimeout)
		defer cancel()

		rsp, err := obj.GetOrderBook(ctx, c.Params("pairID"))
		if err != nil {
			return err
		}

		sort.Slice(rsp.Ask, func(i, j int) bool {
			return rsp.Ask[i][0].LessThan(rsp.Ask[j][0])
		})

		sort.Slice(rsp.Bid, func(i, j int) bool {
			return rsp.Bid[i][0].GreaterThan(rsp.Bid[j][0])
		})

		return c.Status(fiber.StatusOK).JSON(rsp)
	})

	s.engine = engine
}
