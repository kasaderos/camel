package model

import (
	"time"
)

const (
	emaChange = "ema_change"
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

func (s *State) init() {
	if s.data == nil {
		s.data = make(map[string]any)
	}
}

func (s *State) SetDate(t time.Time) {
	s.init()
	s.data["date"] = t.Format(time.DateOnly)
}

func (s *State) SetEmaChange(f float64) {
	s.init()
	s.data[emaChange] = f
}

func (s *State) EmaChange() (float64, bool) {
	value, ok := s.data[emaChange]
	if !ok {
		return 0.0, false
	}

	f, ok := value.(float64)
	if !ok {
		return 0.0, false
	}

	return f, true
}
