package server

import (
	"exchanges/pkg/exchange"
	"github.com/gofiber/fiber/v2"
	"sync"
)

func NewServer() *Server {
	obj := new(Server)
	obj.mu = new(sync.Mutex)
	obj.exchanges = make(map[string]exchange.Exchange)
	obj.init()

	return obj
}

type Server struct {
	mu        *sync.Mutex
	engine    *fiber.App
	exchanges map[string]exchange.Exchange
}
