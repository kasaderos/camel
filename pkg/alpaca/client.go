package alpaca

import (
	"context"
	"log/slog"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
)

type Client struct {
	account *alpaca.Account
	client  *alpaca.Client
}

func NewClient(
	apikey, apisecret string,
	baseURL string,
) (*Client, error) {
	client := alpaca.NewClient(alpaca.ClientOpts{
		// Alternatively you can set your key and secret using the
		// APCA_API_KEY_ID and APCA_API_SECRET_KEY environment variables
		APIKey:    apikey,
		APISecret: apisecret,
		BaseURL:   baseURL,
	})

	acct, err := client.GetAccount()
	if err != nil {
		return nil, err
	}

	return &Client{
		account: acct,
		client:  client,
	}, nil
}

func (c *Client) LogTradeUpdates(ctx context.Context) {
	// Listen to trade updates in the background (with unlimited reconnect)
	alpaca.StreamTradeUpdatesInBackground(context.TODO(), func(tu alpaca.TradeUpdate) {
		slog.Info("TRADE UPDATE", slog.Any("update", tu))
	})
}
