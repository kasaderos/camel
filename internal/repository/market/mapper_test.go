package market

import (
	"testing"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"github.com/stretchr/testify/require"
)

func TestToModel(t *testing.T) {
	ts := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	got := toModel(AssetBar{
		AssetID:   "AAPL",
		Timestamp: ts,
		Open:      100,
		High:      110,
		Low:       95,
		Close:     105,
		Volume:    1000,
	})

	require.Equal(t, model.Bar{
		Timestamp: ts,
		Open:      100,
		High:      110,
		Low:       95,
		Close:     105,
		Volume:    1000,
	}, got)
}
