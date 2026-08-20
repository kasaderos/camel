package alpaca

import (
	"log/slog"

	alpacasdk "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

type Service struct {
	client *alpacasdk.Client
	market *marketdata.Client
}

func New(apikey, apisecret, tradingURL, marketURL string) *Service {
	slog.Info("alpaca trading", "baseURL", tradingURL)
	slog.Info("alpaca market", "baseURL", marketURL)

	return &Service{
		client: alpacasdk.NewClient(alpacasdk.ClientOpts{
			APIKey:    apikey,
			APISecret: apisecret,
			BaseURL:   tradingURL,
		}),
		market: marketdata.NewClient(marketdata.ClientOpts{
			APIKey:    apikey,
			APISecret: apisecret,
			BaseURL:   marketURL,
			Feed:      marketdata.IEX,
		}),
	}
}
