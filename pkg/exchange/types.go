package exchange

import (
	"github.com/shopspring/decimal"
)

type Pair struct {
	Id         string          `json:"id"`
	BaseAsset  string          `json:"base_asset"`
	QuoteAsset string          `json:"quote_asset"`
	Ask        decimal.Decimal `json:"ask"`
	Bid        decimal.Decimal `json:"bid"`
}

type OrderBook struct {
	Ask [][]decimal.Decimal `json:"ask"`
	Bid [][]decimal.Decimal `json:"bid"`
}
