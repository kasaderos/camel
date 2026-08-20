package market

import (
	"context"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

type MarketProvider interface {
	FetchBars(
		ctx context.Context,
		symbol string,
		start, end time.Time,
	) ([]model.Bar, error)
}

type Repository interface {
	SaveBars(ctx context.Context, assetID string, bars []model.Bar) error
	FetchBars(ctx context.Context, assetID string, start, end time.Time) ([]model.Bar, error)
}
