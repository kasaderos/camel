package alpaca

import (
	"context"
	"fmt"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/shopspring/decimal"
)

func (s *Client) SellMarketStock(
	ctx context.Context,
	symbol string,
	qty decimal.Decimal,
) (*alpaca.Order, error) {
	order, err := s.client.PlaceOrder(alpaca.PlaceOrderRequest{
		Symbol:      symbol,
		Qty:         &qty,
		Side:        "sell",
		Type:        "market",
		TimeInForce: "day",
	})
	if err != nil {
		return nil, fmt.Errorf("sell market stock: %w", err)
	}

	return order, nil
}
