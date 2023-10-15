package gateio

import (
	"exchanges/pkg/cache"
	"net/http"
	"sync"
)

func NewAPI() *API {
	return &API{
		mu:  new(sync.Mutex),
		cli: new(http.Client),
		db:  cache.NewDB(),
	}
}

type API struct {
	mu  *sync.Mutex
	cli *http.Client
	db  *cache.DB
}
