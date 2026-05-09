package alpaca

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

func (s *MarketDataClient) FetchBars(
	ctx context.Context,
	symbol string,
	start time.Time,
	end time.Time,
) ([]Bar, error) {
	values := url.Values{}

	values.Set("symbols", symbol)
	values.Set("timeframe", "1D")
	values.Set("start", start.Format(time.DateOnly))
	values.Set("end", end.Format(time.DateOnly))
	values.Set("limit", "100")
	values.Set("adjustment", "raw")
	values.Set("feed", "iex")
	values.Set("sort", "asc")

	reqURL := s.baseURL + "/v2/stocks/bars?" + values.Encode()

	req, err := http.NewRequest(
		http.MethodGet,
		reqURL,
		nil,
	)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("APCA-API-KEY-ID", s.apikey)
	req.Header.Set("APCA-API-SECRET-KEY", s.secret)

	resp, err := s.client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read all: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d, resp %s", resp.StatusCode, data)
	}

	slog.Info(
		"fetch bars",
		"symbol", symbol,
		"start", start,
		"end", end,
	)

	var barsResp BarsResponse

	err = json.Unmarshal(data, &barsResp)
	if err != nil {
		return nil, fmt.Errorf("unmarshal bar response: %w", err)
	}

	bars, exist := barsResp.Bars[symbol]
	if !exist {
		return nil, fmt.Errorf("symbol not found in bars response: %s", symbol)
	}

	return bars, nil
}

type BarsResponse struct {
	Bars          map[string][]Bar `json:"bars"`
	NextPageToken string           `json:"next_page_token"`
}

type Bar struct {
	Close     float64   `json:"c"`
	High      float64   `json:"h"`
	Low       float64   `json:"l"`
	Trades    int64     `json:"n"`
	Open      float64   `json:"o"`
	Timestamp time.Time `json:"t"`
	Volume    int64     `json:"v"`
	VWAP      float64   `json:"vw"`
}
