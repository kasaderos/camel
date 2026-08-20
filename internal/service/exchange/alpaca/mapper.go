package alpaca

import (
	"fmt"
	"time"

	alpacasdk "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/kasaderos/camel/internal/model"
)

const (
	alpacaOrderStatusNew                = "new"
	alpacaOrderStatusPartiallyFilled    = "partially_filled"
	alpacaOrderStatusFilled             = "filled"
	alpacaOrderStatusDoneForDay         = "done_for_day"
	alpacaOrderStatusCanceled           = "canceled"
	alpacaOrderStatusExpired            = "expired"
	alpacaOrderStatusReplaced           = "replaced"
	alpacaOrderStatusPendingCancel      = "pending_cancel"
	alpacaOrderStatusPendingReplace     = "pending_replace"
	alpacaOrderStatusAccepted           = "accepted"
	alpacaOrderStatusPendingNew         = "pending_new"
	alpacaOrderStatusAcceptedForBidding = "accepted_for_bidding"
	alpacaOrderStatusStopped            = "stopped"
	alpacaOrderStatusRejected           = "rejected"
	alpacaOrderStatusSuspended          = "suspended"
	alpacaOrderStatusCalculated         = "calculated"
	alpacaOrderStatusPendingReview      = "pending_review"
	alpacaOrderStatusHeld               = "held"
)

func toAlpacaSide(side model.OrderSide) (alpacasdk.Side, error) {
	switch side {
	case model.OrderSideBuy:
		return alpacasdk.Buy, nil

	case model.OrderSideSell:
		return alpacasdk.Sell, nil

	default:
		return "", fmt.Errorf("invalid order side: %s (must be 'buy' or 'sell')", side)
	}
}

func toModelOrderSide(side alpacasdk.Side) (model.OrderSide, error) {
	switch side {
	case alpacasdk.Buy:
		return model.OrderSideBuy, nil

	case alpacasdk.Sell:
		return model.OrderSideSell, nil

	default:
		return "", fmt.Errorf("invalid order side: %s (must be 'buy' or 'sell')", side)
	}
}

func toModelOrderStatus(status string) (model.OrderStatus, error) {
	switch status {
	case alpacaOrderStatusFilled:
		return model.OrderStatusFilled, nil

	case alpacaOrderStatusPartiallyFilled:
		return model.OrderStatusPartiallyFilled, nil

	case alpacaOrderStatusCanceled,
		alpacaOrderStatusExpired,
		alpacaOrderStatusRejected,
		alpacaOrderStatusReplaced:
		return model.OrderStatusCancelled, nil

	case alpacaOrderStatusNew,
		alpacaOrderStatusAccepted,
		alpacaOrderStatusPendingNew,
		alpacaOrderStatusAcceptedForBidding,
		alpacaOrderStatusPendingCancel,
		alpacaOrderStatusPendingReplace,
		alpacaOrderStatusPendingReview,
		alpacaOrderStatusStopped,
		alpacaOrderStatusSuspended,
		alpacaOrderStatusHeld:
		return model.OrderStatusPending, nil

	default:
		return "", fmt.Errorf("unknown order status: %s", status)
	}
}

func toModelBar(bar marketdata.Bar) model.Bar {
	return model.Bar{
		Timestamp: bar.Timestamp,
		Open:      bar.Open,
		High:      bar.High,
		Low:       bar.Low,
		Close:     bar.Close,
		Volume:    int64(bar.Volume),
	}
}

func toModelOrder(order *alpacasdk.Order) (*model.Order, error) {
	var qty float64
	if order.Qty != nil {
		qty, _ = order.Qty.Float64()
	}

	filledQty, _ := order.FilledQty.Float64()
	if qty == 0 {
		qty = filledQty
	}

	var price float64
	if order.FilledAvgPrice != nil {
		price, _ = order.FilledAvgPrice.Float64()
	}

	var filledAt time.Time
	if order.FilledAt != nil {
		filledAt = *order.FilledAt
	}

	side, err := toModelOrderSide(order.Side)
	if err != nil {
		return nil, err
	}

	status, err := toModelOrderStatus(order.Status)
	if err != nil {
		return nil, err
	}

	return &model.Order{
		ID:           order.ID,
		AssetID:      order.Symbol,
		AvgFillPrice: price,
		Qty:          qty,
		FilledQty:    filledQty,
		Side:         side,
		Status:       status,
		FilledAt:     filledAt,
		CreatedAt:    order.CreatedAt,
	}, nil
}
