package alpaca

import (
	"context"
	"fmt"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/kasaderos/camel/internal/model"
)

func (s *Service) FetchBars(
	ctx context.Context,
	symbol string,
	start time.Time,
	end time.Time,
) ([]model.Bar, error) {
	bars, err := s.market.GetBars(symbol, marketdata.GetBarsRequest{
		TimeFrame: marketdata.OneDay,
		Start:     start,
		End:       end,
		Feed:      marketdata.IEX,
	})
	if err != nil {
		return nil, fmt.Errorf("get bars for %s: %w", symbol, err)
	}

	out := make([]model.Bar, 0, len(bars))
	for _, bar := range bars {
		out = append(out, toModelBar(bar))
	}

	return out, nil
}
