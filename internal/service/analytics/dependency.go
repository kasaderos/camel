package analytics

import (
	"context"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

type MarketService interface {
	FetchBars(
		ctx context.Context,
		assetID string,
		start time.Time,
		end time.Time,
	) ([]model.Bar, error)
}
