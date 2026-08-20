package market

import "time"

type AssetBar struct {
	AssetID   string    `gorm:"column:asset_id;primaryKey"`
	Timestamp time.Time `gorm:"column:date;primaryKey"`
	Open      float64   `gorm:"column:open"`
	High      float64   `gorm:"column:high"`
	Low       float64   `gorm:"column:low"`
	Close     float64   `gorm:"column:close"`
	Volume    int64     `gorm:"column:volume"`
}

func (AssetBar) TableName() string {
	return "asset_bars"
}
