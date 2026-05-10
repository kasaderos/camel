package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type State struct {
	data map[string]any
}

func (s *State) Load(data map[string]any) {
	s.data = data
}

func (s *State) Data() map[string]any {
	return s.data
}

func (s *State) SetDate(t time.Time) {
	s.data["date"] = t
}

func (s *State) SetEmaChange(d decimal.Decimal) {
	s.data["ema_change"] = d
}

func (s *State) EmaChange() (float64, bool) {
	value, ok := s.data["ema_change"]
	if !ok {
		return 0.0, false
	}

	d, ok := value.(decimal.Decimal)
	if !ok {
		return 0.0, false
	}

	return d.Float64()
}
