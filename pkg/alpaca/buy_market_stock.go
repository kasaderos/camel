package alpaca

import (
	"context"
	"fmt"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
)

func (s *Client) BuyMarketStock(
	ctx context.Context,
	symbol string,
	qty decimal.Decimal,
) (*alpaca.Order, error) {
	order, err := s.client.PlaceOrder(alpaca.PlaceOrderRequest{
		Symbol:      symbol,
		Qty:         &qty,
		Side:        "buy",
		Type:        "market",
		TimeInForce: "day",
	})
	if err != nil {
		return nil, fmt.Errorf("buy market stock: %w", err)
	}

	return order, nil
}
