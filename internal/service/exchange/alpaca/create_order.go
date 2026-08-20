package alpaca

import (
	"context"
	"fmt"

	alpacasdk "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/kasaderos/camel/internal/model"
	"github.com/shopspring/decimal"
)

func (s *Service) CreateMarketOrder(
	ctx context.Context,
	symbol string,
	qty float64,
	side model.OrderSide,
) (*model.Order, error) {
	orderSide, err := toAlpacaSide(side)
	if err != nil {
		return nil, fmt.Errorf("convert order side: %w", err)
	}

	qtyDecimal := decimal.NewFromFloat(qty).Floor()

	order, err := s.client.PlaceOrder(alpacasdk.PlaceOrderRequest{
		Symbol:      symbol,
		Qty:         &qtyDecimal,
		Side:        orderSide,
		Type:        alpacasdk.Market,
		TimeInForce: alpacasdk.Day,
	})
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	return toModelOrder(order)
}
