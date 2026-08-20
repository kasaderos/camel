package market

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

type AssetBar struct {
	AssetID   string    `gorm:"column:asset_id;primaryKey"`
	Timestamp time.Time `gorm:"column:date;primaryKey"`
	Open      float64   `gorm:"column:open"`
	High      float64   `gorm:"column:high"`
	Low       float64   `gorm:"column:low"`
	Close     float64   `gorm:"column:close"`
}

func (AssetBar) TableName() string {
	return "asset_bars"
}

func (r AssetBar) toModel() model.Bar {
	return model.Bar{
		Timestamp: r.Timestamp,
		Open:      r.Open,
		High:      r.High,
		Low:       r.Low,
		Close:     r.Close,
	}
}

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
			})
		}

		err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
		if err != nil {
			return fmt.Errorf("save bars: %w", err)
		}

		return nil
	})
}

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
		out = append(out, row.toModel())
	}

	return out, nil
}
