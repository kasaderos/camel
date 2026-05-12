package alpaca

import (
	"log/slog"
	"net/http"
)

type MarketClient struct {
	client *http.Client

	baseURL string
	apikey  string
	secret  string
}

func NewMarketClient(
	apikey, apisecret string,
	baseURL string,
) (*MarketClient, error) {
	slog.Info("alpaca", "baseURL", baseURL)

	return &MarketClient{
		client: &http.Client{},

		baseURL: baseURL,
		apikey:  apikey,
		secret:  apisecret,
	}, nil
}
