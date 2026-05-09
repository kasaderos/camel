package alpaca

import (
	"log/slog"
	"net/http"
)

type MarketDataClient struct {
	client *http.Client

	baseURL string
	apikey  string
	secret  string
}

func NewMarketDataClient(
	apikey, apisecret string,
	baseURL string,
) (*MarketDataClient, error) {
	slog.Info("alpaca", "baseURL", baseURL, "apikey", apikey, "apisecret", apisecret)

	return &MarketDataClient{
		client: &http.Client{},

		baseURL: baseURL,
		apikey:  apikey,
		secret:  apisecret,
	}, nil
}
