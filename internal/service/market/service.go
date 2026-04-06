package market

import (
	"context"
	"fmt"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"

	"github.com/kasaderos/camel/internal/model"
	"github.com/kasaderos/camel/pkg/slices"
)

type MarketDataProvider interface {
	FetchBars(
		ctx context.Context,
		symbol string,
		start, end time.Time,
	) ([]marketdata.Bar, error)
}

type Service struct {
	market MarketDataProvider
}

func New(market MarketDataProvider) *Service {
	return &Service{
		market: market,
	}
}

func (s *Service) FetchBars(
	ctx context.Context,
	symbol string,
	start, end time.Time,
) ([]model.Bar, error) {
	bars, err := s.market.FetchBars(ctx, symbol, start, end)
	if err != nil {
		return nil, fmt.Errorf("fetch bars: %w", err)
	}

	return slices.Map(bars, func(bar marketdata.Bar) (model.Bar, error) {
		return mapAlpacaBarToBar(bar), nil
	})
}
