package market

import (
	"time"

	"github.com/kasaderos/camel/internal/model"
)

func (s *RepositorySuite) TestSaveBars_RequiresAssetID() {
	err := s.repo.SaveBars(s.T().Context(), "", []model.Bar{{
		Timestamp: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		Open:      1,
		High:      2,
		Low:       0.5,
		Close:     1.5,
	}})
	s.EqualError(err, "asset_id is required")
}

func (s *RepositorySuite) TestSaveBars_EmptyBarsNoOp() {
	err := s.repo.SaveBars(s.T().Context(), "AAPL", nil)
	s.Require().NoError(err)

	bars, err := s.repo.FetchBars(
		s.T().Context(),
		"AAPL",
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
	)
	s.Require().NoError(err)
	s.Empty(bars)
}

func (s *RepositorySuite) TestSaveBars() {
	ctx := s.T().Context()
	start := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)

	original := []model.Bar{
		{Timestamp: start, Open: 100, High: 110, Low: 95, Close: 105, Volume: 1000},
		{Timestamp: end, Open: 105, High: 115, Low: 100, Close: 112, Volume: 2000},
	}

	err := s.repo.SaveBars(ctx, "AAPL", original)
	s.Require().NoError(err)

	fetched, err := s.repo.FetchBars(ctx, "AAPL", start, end)
	s.Require().NoError(err)
	s.Require().Len(fetched, 2)
	s.Equal(start, fetched[0].Timestamp.UTC())
	s.Equal(100.0, fetched[0].Open)
	s.Equal(110.0, fetched[0].High)
	s.Equal(95.0, fetched[0].Low)
	s.Equal(105.0, fetched[0].Close)
	s.Equal(int64(1000), fetched[0].Volume)
	s.Equal(end, fetched[1].Timestamp.UTC())
	s.Equal(112.0, fetched[1].Close)
	s.Equal(int64(2000), fetched[1].Volume)
}

func (s *RepositorySuite) TestSaveBars_OnConflictDoNothing() {
	ctx := s.T().Context()
	ts := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	err := s.repo.SaveBars(ctx, "AAPL", []model.Bar{{
		Timestamp: ts,
		Open:      100,
		High:      110,
		Low:       95,
		Close:     105,
		Volume:    1000,
	}})
	s.Require().NoError(err)

	err = s.repo.SaveBars(ctx, "AAPL", []model.Bar{{
		Timestamp: ts,
		Open:      1,
		High:      2,
		Low:       0.5,
		Close:     1.5,
		Volume:    999,
	}})
	s.Require().NoError(err)

	fetched, err := s.repo.FetchBars(ctx, "AAPL", ts, ts)
	s.Require().NoError(err)
	s.Require().Len(fetched, 1)
	s.Equal(100.0, fetched[0].Open)
	s.Equal(110.0, fetched[0].High)
	s.Equal(95.0, fetched[0].Low)
	s.Equal(105.0, fetched[0].Close)
	s.Equal(int64(1000), fetched[0].Volume)
}
