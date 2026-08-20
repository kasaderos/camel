package market

import "github.com/kasaderos/camel/internal/model"

func toModel(r AssetBar) model.Bar {
	return model.Bar{
		Timestamp: r.Timestamp,
		Open:      r.Open,
		High:      r.High,
		Low:       r.Low,
		Close:     r.Close,
		Volume:    r.Volume,
	}
}
