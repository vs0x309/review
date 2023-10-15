package bybit

import (
	"context"
	"exchanges/pkg/exchange"
	"fmt"
	"github.com/shopspring/decimal"
	"net/url"
	"time"
)

func (a *API) GetID() string {
	return "bybit"
}

func (a *API) GetPairs(ctx context.Context) ([]exchange.Pair, error) {
	pairs, err := a.getPairs(ctx)
	if err != nil {
		return nil, err
	}

	tickers, err := a.getTickers(ctx)
	if err != nil {
		return nil, err
	}

	var result []exchange.Pair

	for _, row := range pairs {
		if askBid, ok := tickers[row.Id]; ok {
			result = append(result, exchange.Pair{
				Id:         row.Id,
				BaseAsset:  row.BaseAsset,
				QuoteAsset: row.QuoteAsset,
				Ask:        askBid[0],
				Bid:        askBid[1],
			})
		}
	}

	return result, nil
}

func (a *API) getPairs(ctx context.Context) ([]exchange.Pair, error) {
	cacheKey := "getPairs"
	cacheTimeout := time.Minute * 5

	if cache, ok := a.db.Get(cacheKey).([]exchange.Pair); ok {
		return cache, nil
	}

	endpoint := "/v5/market/instruments-info"

	payload := url.Values{}
	payload.Set("category", "spot")

	var temp struct {
		Result struct {
			List []struct {
				Symbol    string `json:"symbol"`
				BaseCoin  string `json:"baseCoin"`
				QuoteCoin string `json:"quoteCoin"`
				Status    string `json:"status"`
			} `json:"list"`
		} `json:"result"`
	}

	if err := a.doPublicGET(ctx, endpoint, payload, &temp); err != nil {
		return nil, err
	}

	var result []exchange.Pair

	for _, row := range temp.Result.List {
		if row.Status != "Trading" {
			continue
		}

		if len(row.Symbol) == 0 {
			continue
		}

		if len(row.BaseCoin) == 0 {
			continue
		}

		if len(row.QuoteCoin) == 0 {
			continue
		}

		result = append(result, exchange.Pair{
			Id:         row.Symbol,
			BaseAsset:  row.BaseCoin,
			QuoteAsset: row.QuoteCoin,
		})
	}

	a.db.Set(cacheKey, cacheTimeout, result)

	return result, nil
}

func (a *API) getTickers(ctx context.Context) (map[string][]decimal.Decimal, error) {
	endpoint := "/v5/market/tickers"

	payload := url.Values{}
	payload.Set("category", "spot")

	var temp struct {
		Result struct {
			List []struct {
				Symbol string          `json:"symbol"`
				Ask    decimal.Decimal `json:"ask1Price"`
				Bid    decimal.Decimal `json:"bid1Price"`
			} `json:"list"`
		} `json:"result"`
	}

	if err := a.doPublicGET(ctx, endpoint, payload, &temp); err != nil {
		return nil, err
	}

	result := make(map[string][]decimal.Decimal)

	for _, row := range temp.Result.List {
		if len(row.Symbol) == 0 {
			continue
		}

		if row.Ask.LessThanOrEqual(decimal.Zero) {
			continue
		}

		if row.Bid.LessThanOrEqual(decimal.Zero) {
			continue
		}

		result[row.Symbol] = []decimal.Decimal{row.Ask, row.Bid}
	}

	return result, nil
}

func (a *API) GetOrderBook(ctx context.Context, pairID string) (exchange.OrderBook, error) {
	endpoint := "/v5/market/orderbook"

	payload := url.Values{}
	payload.Set("category", "spot")
	payload.Set("symbol", pairID)
	payload.Set("limit", "50")

	var temp struct {
		Result struct {
			Symbol string              `json:"s"`
			Ask    [][]decimal.Decimal `json:"a"`
			Bid    [][]decimal.Decimal `json:"b"`
		} `json:"result"`
	}

	if err := a.doPublicGET(ctx, endpoint, payload, &temp); err != nil {
		return exchange.OrderBook{}, err
	}

	for _, asks := range temp.Result.Ask {
		if len(asks) != 2 {
			return exchange.OrderBook{}, fmt.Errorf("json parse error: %v", temp)
		}
	}

	for _, bids := range temp.Result.Bid {
		if len(bids) != 2 {
			return exchange.OrderBook{}, fmt.Errorf("json parse error: %v", temp)
		}
	}

	return exchange.OrderBook{Ask: temp.Result.Ask, Bid: temp.Result.Bid}, nil
}
