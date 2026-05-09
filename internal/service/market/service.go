package market

import (
	"context"
	"fmt"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"github.com/kasaderos/camel/pkg/alpaca"
	"github.com/samber/lo"
)

type Service struct {
	market MarketProvider
	repo   Repository
}

func New(market MarketProvider, repo Repository) *Service {
	return &Service{
		market: market,
		repo:   repo,
	}
}

// gap is a half-open time interval that needs to be fetched from the upstream provider.
type gap struct{ start, end time.Time }

// missingGaps returns the contiguous intervals that [bars] does not cover within [start, end].
// At most two gaps are returned: a prefix and/or a suffix.
func missingGaps(bars []model.Bar, start, end time.Time) []gap {
	if len(bars) == 0 {
		return []gap{{start, end}}
	}

	var gaps []gap
	if bars[0].Timestamp.After(start) {
		gaps = append(gaps, gap{start, bars[0].Timestamp})
	}

	if bars[len(bars)-1].Timestamp.Before(end) {
		gaps = append(gaps, gap{bars[len(bars)-1].Timestamp, end})
	}

	return gaps
}

func (s *Service) FetchBars(
	ctx context.Context,
	symbol string,
	start, end time.Time,
) ([]model.Bar, error) {
	bars, err := s.repo.FetchBars(ctx, symbol, start, end)
	if err != nil {
		return nil, fmt.Errorf("fetch cached bars: %w", err)
	}

	if coversRange(bars, start, end) {
		return bars, nil
	}

	for _, g := range missingGaps(bars, start, end) {
		fetched, err := s.fetchFromProvider(ctx, symbol, g.start, g.end)
		if err != nil {
			return nil, err
		}

		if err := s.repo.SaveBars(ctx, symbol, fetched); err != nil {
			return nil, fmt.Errorf("save bars [%v, %v]: %w", g.start, g.end, err)
		}
	}

	// Re-read so the caller always gets a consistently ordered, merged slice.
	bars, err = s.repo.FetchBars(ctx, symbol, start, end)
	if err != nil {
		return nil, fmt.Errorf("fetch cached bars (post-fill): %w", err)
	}

	return bars, nil
}

func (s *Service) fetchFromProvider(
	ctx context.Context,
	symbol string,
	start, end time.Time,
) ([]model.Bar, error) {
	bars, err := s.market.FetchBars(ctx, symbol, start, end)
	if err != nil {
		return nil, fmt.Errorf("fetch bars: %w", err)
	}

	return lo.Map(bars, func(b alpaca.Bar, _ int) model.Bar {
		return mapAlpacaBarToBar(b)
	}), nil
}

func coversRange(bars []model.Bar, start, end time.Time) bool {
	return len(bars) > 0 &&
		!bars[0].Timestamp.After(start) &&
		!bars[len(bars)-1].Timestamp.Before(end)
}
