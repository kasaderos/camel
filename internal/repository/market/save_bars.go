package market

import (
	"context"
	"errors"
	"fmt"

	"github.com/kasaderos/camel/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *Repository) SaveBars(ctx context.Context, assetID string, bars []model.Bar) error {
	if assetID == "" {
		return errors.New("asset_id is required")
	}

	if len(bars) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		rows := make([]AssetBar, 0, len(bars))
		for _, b := range bars {
			rows = append(rows, AssetBar{
				AssetID:   assetID,
				Timestamp: b.Timestamp,
				Open:      b.Open,
				High:      b.High,
				Low:       b.Low,
				Close:     b.Close,
				Volume:    b.Volume,
			})
		}

		err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
		if err != nil {
			return fmt.Errorf("save bars: %w", err)
		}

		return nil
	})
}
