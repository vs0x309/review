package okx

import (
	"context"
	"exchanges/pkg/exchange"
	"fmt"
	"github.com/shopspring/decimal"
	"net/url"
	"time"
)

func (a *API) GetID() string {
	return "okx"
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

	if cache, ok := a.cacheDB.Get(cacheKey).([]exchange.Pair); ok {
		return cache, nil
	}

	endpoint := "/api/v5/public/instruments"

	payload := url.Values{}
	payload.Set("instType", "SPOT")

	var temp struct {
		Data []struct {
			InstId   string `json:"instId"`
			BaseCcy  string `json:"baseCcy"`
			QuoteCcy string `json:"quoteCcy"`
			State    string `json:"state"`
		} `json:"data"`
	}

	if err := a.doPublicGET(ctx, endpoint, payload, &temp); err != nil {
		return nil, err
	}

	var result []exchange.Pair

	for _, row := range temp.Data {
		if row.State != "live" {
			continue
		}

		if len(row.InstId) == 0 {
			continue
		}

		if len(row.BaseCcy) == 0 {
			continue
		}

		if len(row.QuoteCcy) == 0 {
			continue
		}

		result = append(result, exchange.Pair{
			Id:         row.InstId,
			BaseAsset:  row.BaseCcy,
			QuoteAsset: row.QuoteCcy,
		})
	}

	a.cacheDB.Set(cacheKey, cacheTimeout, result)

	return result, nil
}

func (a *API) getTickers(ctx context.Context) (map[string][]decimal.Decimal, error) {
	endpoint := "/api/v5/market/tickers"

	payload := url.Values{}
	payload.Set("instType", "SPOT")

	var temp struct {
		Data []struct {
			InstId string          `json:"instId"`
			AskPx  decimal.Decimal `json:"askPx"`
			BidPx  decimal.Decimal `json:"bidPx"`
		} `json:"data"`
	}

	if err := a.doPublicGET(ctx, endpoint, payload, &temp); err != nil {
		return nil, err
	}

	result := make(map[string][]decimal.Decimal)

	for _, row := range temp.Data {
		if len(row.InstId) == 0 {
			continue
		}

		if row.AskPx.LessThanOrEqual(decimal.Zero) {
			continue
		}

		if row.BidPx.LessThanOrEqual(decimal.Zero) {
			continue
		}

		result[row.InstId] = []decimal.Decimal{row.AskPx, row.BidPx}
	}

	return result, nil
}

func (a *API) GetOrderBook(ctx context.Context, pairID string) (exchange.OrderBook, error) {
	endpoint := "/api/v5/market/books"

	payload := url.Values{}
	payload.Set("instId", pairID)
	payload.Set("sz", "100")

	var temp struct {
		Data []struct {
			Asks [][]decimal.Decimal `json:"asks"`
			Bids [][]decimal.Decimal `json:"bids"`
		} `json:"data"`
	}

	if err := a.doPublicGET(ctx, endpoint, payload, &temp); err != nil {
		return exchange.OrderBook{}, err
	}

	if len(temp.Data) != 1 {
		return exchange.OrderBook{}, fmt.Errorf("json parse error: %v", temp.Data)
	}

	var asks, bids [][]decimal.Decimal

	for _, row := range temp.Data[0].Asks {
		if len(row) != 4 {
			return exchange.OrderBook{}, fmt.Errorf("json parse error: %v", temp.Data)
		}

		asks = append(asks, []decimal.Decimal{row[0], row[1]})
	}

	for _, row := range temp.Data[0].Asks {
		if len(row) != 4 {
			return exchange.OrderBook{}, fmt.Errorf("json parse error: %v", temp.Data)
		}

		bids = append(bids, []decimal.Decimal{row[0], row[1]})
	}

	return exchange.OrderBook{Ask: asks, Bid: bids}, nil
}
