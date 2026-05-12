package alpaca

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/kasaderos/camel/internal/model"
	"github.com/shopspring/decimal"
)

// CreateOrder creates a new order with the given parameters
func (c *TradingClient) CreateOrder(ctx context.Context, req alpaca.PlaceOrderRequest) (*alpaca.Order, error) {
	order, err := c.client.PlaceOrder(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return order, nil
}

// FetchOrders retrieves orders based on the given filter
func (c *TradingClient) FetchOrders(
	ctx context.Context,
	filter *alpaca.GetOrdersRequest,
) ([]alpaca.Order, error) {
	if filter == nil {
		filter = &alpaca.GetOrdersRequest{
			Status: "all",
			Limit:  100,
		}
	}

	orders, err := c.client.GetOrders(*filter)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch orders: %w", err)
	}
	return orders, nil
}

// FetchPositions retrieves all open positions
func (c *TradingClient) FetchPositions(ctx context.Context) ([]alpaca.Position, error) {
	positions, err := c.client.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch positions: %w", err)
	}

	return positions, nil
}

// CreateMarketOrderString creates a market order for buying or selling with string side parameter
// This method implements the Exchanger interface used by asset agents
func (c *TradingClient) CreateMarketOrder(
	ctx context.Context,
	symbol string,
	qty float64,
	side string,
) (*model.Order, error) {
	var orderSide alpaca.Side
	switch side {
	case "buy":
		orderSide = alpaca.Buy
	case "sell":
		orderSide = alpaca.Sell
	default:
		return nil, fmt.Errorf("invalid order side: %s (must be 'buy' or 'sell')", side)
	}

	qtyDecimal := decimal.NewFromFloat(qty).Floor()

	req := alpaca.PlaceOrderRequest{
		Symbol:      symbol,
		Qty:         &qtyDecimal,
		Side:        orderSide,
		Type:        alpaca.Market,
		TimeInForce: alpaca.Day,
	}

	order, err := c.CreateOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	slog.Info(
		"order created",
		"order_id", order.ID,
		"symbol", order.Symbol,
		"side", order.Side,
		"qty", order.Qty,
		"status", order.Status,
	)

	// Convert Alpaca order to model.Order
	var price float64
	if order.FilledAvgPrice != nil {
		price, _ = order.FilledAvgPrice.Float64()
	}

	modelOrder := &model.Order{
		AssetID:   order.Symbol,
		Price:     price,
		Qty:       qty,
		Status:    string(order.Status),
		CreatedAt: order.CreatedAt,
	}

	return modelOrder, nil
}

// FetchPrice retrieves the current market price for an asset
func (c *TradingClient) FetchPrice(ctx context.Context, assetID string) (float64, error) {
	// Fetch the most recent bar to get current price
	bars, err := c.marketClient.FetchBars(
		ctx,
		assetID,
		time.Now().Add(-24*time.Hour),
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch bars for %s: %w", assetID, err)
	}

	if len(bars) == 0 {
		return 0, fmt.Errorf("no price data available for asset %s", assetID)
	}

	// Return the close price of the most recent bar
	return bars[len(bars)-1].Close, nil
}
