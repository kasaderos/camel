package alpaca

import (
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

type MarketDataClient struct {
	client *marketdata.Client
}

func NewMarketDataClient(
	apikey, apisecret string,
	baseURL string,
) (*MarketDataClient, error) {
	client := marketdata.NewClient(
		marketdata.ClientOpts{
			APIKey:    apikey,
			APISecret: apisecret,
			BaseURL:   baseURL,
		},
	)

	return &MarketDataClient{
		client: client,
	}, nil
}
