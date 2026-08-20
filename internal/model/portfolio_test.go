package model

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPortfolioPrint(t *testing.T) {
	updatedAt := time.Date(2026, 8, 21, 10, 30, 0, 0, time.UTC)
	p := Portfolio{
		ID:        "portfolio-1",
		Cash:      1234.567,
		Cost:      890.1,
		UpdatedAt: updatedAt,
		Stocks: []PortfolioStock{
			{StockID: "AAPL", Quantity: 1},
			{StockID: "MSFT", Quantity: 0},
			{StockID: "NVDA", Quantity: 3.5},
			{StockID: "TSLA", Quantity: -2},
		},
	}

	var buf bytes.Buffer
	p.Print(&buf)

	want := "" +
		"----------------\n" +
		"ID:       portfolio-1\n" +
		"Updated:  2026-08-21T10:30:00Z\n" +
		"----------------\n" +
		"Cash:     1234.57\n" +
		"Cost:     890.10\n" +
		"----------------\n" +
		"Positions\n" +
		"AAPL         1\n" +
		"NVDA         3.5\n"

	require.Equal(t, want, buf.String())
}

func TestAddStockQuantity(t *testing.T) {
	tests := []struct {
		name    string
		p       Portfolio
		stockID StockID
		qty     float64
		wantQty float64
		wantErr string
	}{
		{
			name: "increases quantity for existing stock",
			p: Portfolio{
				Stocks: []PortfolioStock{
					{StockID: "AAPL", Quantity: 1},
					{StockID: "MSFT", Quantity: 5},
				},
			},
			stockID: "AAPL",
			qty:     2,
			wantQty: 3,
		},
		{
			name: "decreases quantity for existing stock",
			p: Portfolio{
				Stocks: []PortfolioStock{
					{StockID: "AAPL", Quantity: 5},
				},
			},
			stockID: "AAPL",
			qty:     -2,
			wantQty: 3,
		},
		{
			name: "returns error when stock is missing",
			p: Portfolio{
				Stocks: []PortfolioStock{
					{StockID: "AAPL", Quantity: 1},
				},
			},
			stockID: "MSFT",
			qty:     1,
			wantErr: "stock not found in portfolio: MSFT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.p.AddStockQuantity(tt.stockID, tt.qty)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantQty, tt.p.Stocks[0].Quantity)
		})
	}
}

func TestApplyOrder(t *testing.T) {
	tests := []struct {
		name     string
		p        Portfolio
		order    *Order
		wantCash float64
		wantQty  float64
		wantErr  string
	}{
		{
			name: "returns error when order is nil",
			p: Portfolio{
				Cash: 1000,
				Stocks: []PortfolioStock{
					{StockID: "AAPL", Quantity: 1},
				},
			},
			wantErr: "order id is empty",
		},
		{
			name: "returns error when stock is missing",
			p: Portfolio{
				Cash: 1000,
				Stocks: []PortfolioStock{
					{StockID: "AAPL", Quantity: 1},
				},
			},
			order: &Order{
				ID:           "order-1",
				AssetID:      "MSFT",
				Qty:          2,
				AvgFillPrice: 50,
				Side:         OrderSideBuy,
			},
			wantErr: "stock not found in portfolio: MSFT",
		},
		{
			name: "applies buy order",
			p: Portfolio{
				Cash: 1000,
				Stocks: []PortfolioStock{
					{StockID: "AAPL", Quantity: 1},
				},
			},
			order: &Order{
				ID:           "order-1",
				AssetID:      "AAPL",
				Qty:          2,
				AvgFillPrice: 50,
				Side:         OrderSideBuy,
			},
			wantCash: 900,
			wantQty:  3,
		},
		{
			name: "applies sell order",
			p: Portfolio{
				Cash: 1000,
				Stocks: []PortfolioStock{
					{StockID: "AAPL", Quantity: 3},
				},
			},
			order: &Order{
				ID:           "order-1",
				AssetID:      "AAPL",
				Qty:          1,
				AvgFillPrice: 50,
				Side:         OrderSideSell,
			},
			wantCash: 1050,
			wantQty:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.p.ApplyOrder(tt.order)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantCash, tt.p.Cash)
			require.Equal(t, tt.wantQty, tt.p.Stocks[0].Quantity)
		})
	}
}
