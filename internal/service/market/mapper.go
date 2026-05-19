package market

import (
	"github.com/kasaderos/camel/internal/model"
	"github.com/kasaderos/camel/pkg/alpaca"
)

func mapAlpacaBarToBar(item alpaca.Bar) model.Bar {
	return model.Bar{
		Timestamp: item.Timestamp,
		Open:      item.Open,
		High:      item.High,
		Low:       item.Low,
		Close:     item.Close,
		Volume:    item.Volume,
	}
}
