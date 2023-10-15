package gateio

import (
	"context"
	"exchanges/pkg/exchange"
	"fmt"
	"github.com/shopspring/decimal"
	"net/url"
	"time"
)

func (a *API) GetID() string {
	return "gateio"
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

	endpoint := "/spot/currency_pairs"

	var temp []struct {
		Id          string `json:"id"`
		Base        string `json:"base"`
		Quote       string `json:"quote"`
		TradeStatus string `json:"trade_status"`
	}

	if err := a.doPublicGET(ctx, endpoint, nil, &temp); err != nil {
		return nil, err
	}

	var result []exchange.Pair

	for _, row := range temp {
		if row.TradeStatus != "tradable" {
			continue
		}

		if len(row.Id) == 0 {
			continue
		}

		if len(row.Base) == 0 {
			continue
		}

		if len(row.Quote) == 0 {
			continue
		}

		result = append(result, exchange.Pair{
			Id:         row.Id,
			BaseAsset:  row.Base,
			QuoteAsset: row.Quote,
		})
	}

	a.db.Set(cacheKey, cacheTimeout, result)

	return result, nil
}

func (a *API) getTickers(ctx context.Context) (map[string][]decimal.Decimal, error) {
	endpoint := "/spot/tickers"

	var temp []struct {
		Id  string `json:"currency_pair"`
		Ask string `json:"lowest_ask"`
		Bid string `json:"highest_bid"`
	}

	if err := a.doPublicGET(ctx, endpoint, nil, &temp); err != nil {
		return nil, err
	}

	result := make(map[string][]decimal.Decimal)

	for _, row := range temp {
		if len(row.Id) == 0 {
			continue
		}

		ask, err := decimal.NewFromString(row.Ask)
		if err != nil {
			continue
		}

		bid, err := decimal.NewFromString(row.Bid)
		if err != nil {
			continue
		}

		if ask.LessThanOrEqual(decimal.Zero) {
			continue
		}

		if bid.LessThanOrEqual(decimal.Zero) {
			continue
		}

		result[row.Id] = []decimal.Decimal{ask, bid}
	}

	return result, nil
}

func (a *API) GetOrderBook(ctx context.Context, pairID string) (exchange.OrderBook, error) {
	endpoint := "/spot/order_book"

	payload := url.Values{}
	payload.Set("currency_pair", pairID)
	payload.Set("limit", "100")

	var temp struct {
		Asks [][]decimal.Decimal `json:"asks"`
		Bids [][]decimal.Decimal `json:"bids"`
	}

	if err := a.doPublicGET(ctx, endpoint, payload, &temp); err != nil {
		return exchange.OrderBook{}, err
	}

	for _, asks := range temp.Asks {
		if len(asks) != 2 {
			return exchange.OrderBook{}, fmt.Errorf("json parse error: %v", temp)
		}
	}

	for _, bids := range temp.Asks {
		if len(bids) != 2 {
			return exchange.OrderBook{}, fmt.Errorf("json parse error: %v", temp)
		}
	}

	return exchange.OrderBook{Ask: temp.Asks, Bid: temp.Bids}, nil
}
