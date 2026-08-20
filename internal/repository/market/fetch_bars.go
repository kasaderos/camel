package market

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

func (r *Repository) FetchBars(
	ctx context.Context,
	assetID string,
	start, end time.Time,
) ([]model.Bar, error) {
	if assetID == "" {
		return nil, errors.New("asset_id is required")
	}

	var rows []AssetBar
	err := r.db.WithContext(ctx).
		Where("asset_id = ? AND date >= ? AND date <= ?", assetID, start, end).
		Order("date").
		Find(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("fetch bars: %w", err)
	}

	out := make([]model.Bar, 0, len(rows))
	for _, row := range rows {
		out = append(out, toModel(row))
	}

	return out, nil
}
