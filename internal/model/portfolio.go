package model

type Portfolio struct {
	ID   int64
	Name string

	Weights map[string]float64
	Cache   float64

	PolicyID string
}
