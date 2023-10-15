package exchange

import (
	"context"
)

type Exchange interface {
	GetID() string
	GetPairs(ctx context.Context) ([]Pair, error)
	GetOrderBook(ctx context.Context, pairID string) (OrderBook, error)
}
