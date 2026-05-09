package market

import (
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/kasaderos/camel/internal/model"
)

func mapAlpacaBarToBar(item marketdata.Bar) model.Bar {
	return model.Bar{
		Timestamp: item.Timestamp,
		Open:      item.Open,
		High:      item.High,
		Low:       item.Low,
		Close:     item.Close,
		Volume:    item.Volume,
	}
}
