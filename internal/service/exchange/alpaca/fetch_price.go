package alpaca

import (
	"context"
	"fmt"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
)

func (s *Service) FetchPrice(ctx context.Context, assetID string) (float64, time.Time, error) {
	trade, err := s.market.GetLatestTrade(assetID, marketdata.GetLatestTradeRequest{
		Feed: marketdata.IEX,
	})
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("get latest trade for %s: %w", assetID, err)
	}

	if trade == nil {
		return 0, time.Time{}, fmt.Errorf("no price data available for asset %s", assetID)
	}

	return trade.Price, trade.Timestamp, nil
}
