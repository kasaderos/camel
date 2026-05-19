package market

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kasaderos/camel/internal/model"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

type barRow struct {
	AssetID   string    `db:"asset_id"`
	Timestamp time.Time `db:"date"`
	Open      float64   `db:"open"`
	High      float64   `db:"high"`
	Low       float64   `db:"low"`
	Close     float64   `db:"close"`
}

func (r barRow) toModel() model.Bar {
	return model.Bar{
		Timestamp: r.Timestamp,
		Open:      r.Open,
		High:      r.High,
		Low:       r.Low,
		Close:     r.Close,
	}
}

type insertBarRow struct {
	AssetID string    `db:"asset_id"`
	Date    time.Time `db:"date"`
	Open    float64   `db:"open"`
	High    float64   `db:"high"`
	Low     float64   `db:"low"`
	Close   float64   `db:"close"`
}

func (r *Repository) SaveBars(ctx context.Context, assetID string, bars []model.Bar) error {
	if assetID == "" {
		return errors.New("asset_id is required")
	}

	if len(bars) == 0 {
		return nil
	}

	// Use a transaction so the batch insert is atomic.
	return withTransaction(ctx, r.db, func(tx *sqlx.Tx) error {
		rows := make([]insertBarRow, 0, len(bars))
		for _, b := range bars {
			rows = append(rows, insertBarRow{
				AssetID: assetID,
				Date:    b.Timestamp,
				Open:    b.Open,
				High:    b.High,
				Low:     b.Low,
				Close:   b.Close,
			})
		}

		query := `
			INSERT INTO asset_bars (asset_id, date, open, high, low, close)
			VALUES (:asset_id, :date, :open, :high, :low, :close)
			ON CONFLICT DO NOTHING`

		_, err := tx.NamedExecContext(ctx, query, rows)
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

	query := `
		SELECT asset_id, date, open, high, low, close
		FROM asset_bars
		WHERE asset_id = $1
			AND date >= $2
			AND date <= $3
		ORDER BY date 
	`

	var rows []barRow

	err := r.db.SelectContext(ctx, &rows, query, assetID, start, end)
	if err != nil {
		return nil, fmt.Errorf("fetch bars: %w", err)
	}

	out := make([]model.Bar, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toModel())
	}

	return out, nil
}

func withTransaction(ctx context.Context, db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
