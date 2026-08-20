package alpaca

import (
	"testing"
	"time"

	alpacasdk "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/kasaderos/camel/internal/model"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestToAlpacaSide(t *testing.T) {
	tests := []struct {
		name    string
		side    model.OrderSide
		want    alpacasdk.Side
		wantErr string
	}{
		{
			name: "buy",
			side: model.OrderSideBuy,
			want: alpacasdk.Buy,
		},
		{
			name: "sell",
			side: model.OrderSideSell,
			want: alpacasdk.Sell,
		},
		{
			name:    "invalid",
			side:    model.OrderSideNone,
			wantErr: "invalid order side: none (must be 'buy' or 'sell')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toAlpacaSide(tt.side)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				require.Empty(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestToModelOrderSide(t *testing.T) {
	tests := []struct {
		name    string
		side    alpacasdk.Side
		want    model.OrderSide
		wantErr string
	}{
		{
			name: "buy",
			side: alpacasdk.Buy,
			want: model.OrderSideBuy,
		},
		{
			name: "sell",
			side: alpacasdk.Sell,
			want: model.OrderSideSell,
		},
		{
			name:    "invalid",
			side:    alpacasdk.Side("hold"),
			wantErr: "invalid order side: hold (must be 'buy' or 'sell')",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toModelOrderSide(tt.side)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				require.Empty(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestToModelOrderStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		want    model.OrderStatus
		wantErr string
	}{
		{
			name:   "filled",
			status: alpacaOrderStatusFilled,
			want:   model.OrderStatusFilled,
		},
		{
			name:   "partially filled",
			status: alpacaOrderStatusPartiallyFilled,
			want:   model.OrderStatusPartiallyFilled,
		},
		{
			name:   "canceled",
			status: alpacaOrderStatusCanceled,
			want:   model.OrderStatusCancelled,
		},
		{
			name:   "expired",
			status: alpacaOrderStatusExpired,
			want:   model.OrderStatusCancelled,
		},
		{
			name:   "rejected",
			status: alpacaOrderStatusRejected,
			want:   model.OrderStatusCancelled,
		},
		{
			name:   "replaced",
			status: alpacaOrderStatusReplaced,
			want:   model.OrderStatusCancelled,
		},
		{
			name:   "new",
			status: alpacaOrderStatusNew,
			want:   model.OrderStatusPending,
		},
		{
			name:   "accepted",
			status: alpacaOrderStatusAccepted,
			want:   model.OrderStatusPending,
		},
		{
			name:   "pending new",
			status: alpacaOrderStatusPendingNew,
			want:   model.OrderStatusPending,
		},
		{
			name:   "accepted for bidding",
			status: alpacaOrderStatusAcceptedForBidding,
			want:   model.OrderStatusPending,
		},
		{
			name:   "pending cancel",
			status: alpacaOrderStatusPendingCancel,
			want:   model.OrderStatusPending,
		},
		{
			name:   "pending replace",
			status: alpacaOrderStatusPendingReplace,
			want:   model.OrderStatusPending,
		},
		{
			name:   "pending review",
			status: alpacaOrderStatusPendingReview,
			want:   model.OrderStatusPending,
		},
		{
			name:   "stopped",
			status: alpacaOrderStatusStopped,
			want:   model.OrderStatusPending,
		},
		{
			name:   "suspended",
			status: alpacaOrderStatusSuspended,
			want:   model.OrderStatusPending,
		},
		{
			name:   "held",
			status: alpacaOrderStatusHeld,
			want:   model.OrderStatusPending,
		},
		{
			name:    "unknown",
			status:  "weird",
			wantErr: "unknown order status: weird",
		},
		{
			name:    "done for day unmapped",
			status:  alpacaOrderStatusDoneForDay,
			wantErr: "unknown order status: done_for_day",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toModelOrderStatus(tt.status)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				require.Empty(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestToModelBar(t *testing.T) {
	ts := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	got := toModelBar(marketdata.Bar{
		Timestamp: ts,
		Open:      1.1,
		High:      2.2,
		Low:       0.9,
		Close:     1.5,
		Volume:    1000,
	})

	require.Equal(t, model.Bar{
		Timestamp: ts,
		Open:      1.1,
		High:      2.2,
		Low:       0.9,
		Close:     1.5,
		Volume:    1000,
	}, got)
}

func TestToModelOrder(t *testing.T) {
	createdAt := time.Date(2026, 8, 21, 9, 0, 0, 0, time.UTC)
	filledAt := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	qty := decimal.NewFromFloat(2)
	avgPrice := decimal.NewFromFloat(50.25)

	tests := []struct {
		name    string
		order   *alpacasdk.Order
		want    *model.Order
		wantErr string
	}{
		{
			name: "maps filled order with qty and avg price",
			order: &alpacasdk.Order{
				ID:             "order-1",
				Symbol:         "AAPL",
				Side:           alpacasdk.Buy,
				Status:         alpacaOrderStatusFilled,
				Qty:            &qty,
				FilledQty:      decimal.NewFromFloat(2),
				FilledAvgPrice: &avgPrice,
				FilledAt:       &filledAt,
				CreatedAt:      createdAt,
			},
			want: &model.Order{
				ID:           "order-1",
				AssetID:      "AAPL",
				AvgFillPrice: 50.25,
				Qty:          2,
				FilledQty:    2,
				Side:         model.OrderSideBuy,
				Status:       model.OrderStatusFilled,
				FilledAt:     filledAt,
				CreatedAt:    createdAt,
			},
		},
		{
			name: "uses filled qty when order qty is nil",
			order: &alpacasdk.Order{
				ID:        "order-2",
				Symbol:    "AAPL",
				Side:      alpacasdk.Sell,
				Status:    alpacaOrderStatusNew,
				FilledQty: decimal.NewFromFloat(3),
				CreatedAt: createdAt,
			},
			want: &model.Order{
				ID:        "order-2",
				AssetID:   "AAPL",
				Qty:       3,
				FilledQty: 3,
				Side:      model.OrderSideSell,
				Status:    model.OrderStatusPending,
				CreatedAt: createdAt,
			},
		},
		{
			name: "uses filled qty when qty is zero",
			order: &alpacasdk.Order{
				ID:        "order-3",
				Symbol:    "AAPL",
				Side:      alpacasdk.Buy,
				Status:    alpacaOrderStatusPartiallyFilled,
				FilledQty: decimal.NewFromFloat(1.5),
				CreatedAt: createdAt,
			},
			want: &model.Order{
				ID:        "order-3",
				AssetID:   "AAPL",
				Qty:       1.5,
				FilledQty: 1.5,
				Side:      model.OrderSideBuy,
				Status:    model.OrderStatusPartiallyFilled,
				CreatedAt: createdAt,
			},
		},
		{
			name: "returns error for invalid side",
			order: &alpacasdk.Order{
				ID:     "order-4",
				Symbol: "AAPL",
				Side:   alpacasdk.Side("hold"),
				Status: alpacaOrderStatusNew,
			},
			wantErr: "invalid order side: hold (must be 'buy' or 'sell')",
		},
		{
			name: "returns error for unknown status",
			order: &alpacasdk.Order{
				ID:     "order-5",
				Symbol: "AAPL",
				Side:   alpacasdk.Buy,
				Status: "weird",
			},
			wantErr: "unknown order status: weird",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toModelOrder(tt.order)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				require.Nil(t, got)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
