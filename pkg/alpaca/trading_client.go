package alpaca

import (
	"log/slog"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

type TradingClient struct {
	client       *alpaca.Client
	marketClient *MarketClient
	baseURL      string
}

func NewTradingClient(
	apikey, apisecret string,
	baseURL string,
	marketClient *MarketClient,
) (*TradingClient, error) {
	slog.Info("alpaca trading", "baseURL", baseURL)

	// Initialize the Alpaca SDK client
	alpacaSDK := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    apikey,
		APISecret: apisecret,
		BaseURL:   baseURL,
	})

	return &TradingClient{
		client:       alpacaSDK,
		marketClient: marketClient,
		baseURL:      baseURL,
	}, nil
}
