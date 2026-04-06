package model

import "time"

type Portfolio struct {
	ID   int64
	Name string

	Weights map[string]float64
	Cache   float64

	PolicyID string

	CreatedAt time.Time
	UpdatedAt time.Time
}
