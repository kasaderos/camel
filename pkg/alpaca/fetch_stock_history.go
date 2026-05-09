package alpaca

import (
	"context"
	"fmt"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

func (s *MarketDataClient) FetchBars(
	ctx context.Context,
	symbol string,
	start time.Time,
	end time.Time,
) ([]marketdata.Bar, error) {
	req := marketdata.GetBarsRequest{
		TimeFrame: marketdata.OneDay,
		Start:     start,
		End:       end,
	}

	bars, err := s.client.GetBars(symbol, req)
	if err != nil {
		return nil, fmt.Errorf("get bars: %w", err)
	}

	return bars, nil
}
