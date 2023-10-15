package server

import (
	"context"
	"exchanges/pkg/exchange"
)

func (s *Server) SetExchange(obj exchange.Exchange) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.exchanges[obj.GetID()] = obj
}

func (s *Server) Run(ctx context.Context, addr string) error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		errCh <- s.engine.Listen(addr)
	}()

	select {
	case <-ctx.Done():
		return s.engine.ShutdownWithTimeout(shutdownTimeout)
	case err := <-errCh:
		return err
	}
}
