package market

import (
	"context"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/kasaderos/camel/internal/model"
)

type MarketProvider interface {
	FetchBars(
		ctx context.Context,
		symbol string,
		start, end time.Time,
	) ([]marketdata.Bar, error)
}

type BarRepository interface {
	SaveBars(ctx context.Context, assetID string, bars []model.Bar) error
	FetchBars(ctx context.Context, assetID string, start, end time.Time) ([]model.Bar, error)
}
