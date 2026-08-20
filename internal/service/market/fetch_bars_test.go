package market

import (
	"errors"
	"time"

	"github.com/kasaderos/camel/internal/model"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func bar(ts time.Time, close float64) model.Bar {
	return model.Bar{
		Timestamp: ts,
		Open:      close - 1,
		High:      close + 1,
		Low:       close - 2,
		Close:     close,
	}
}

func (s *TestSuite) TestFetchBars_CacheHit() {
	start := day(2024, 1, 2)
	end := day(2024, 1, 4)
	cached := []model.Bar{
		bar(start, 100),
		bar(day(2024, 1, 3), 101),
		bar(end, 102),
	}

	s.repo.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(cached, nil)

	got, err := s.svc.FetchBars(s.ctx, "AAPL", start, end)
	s.Require().NoError(err)
	s.Equal(cached, got)
}

func (s *TestSuite) TestFetchBars_EmptyCacheFillsFromProvider() {
	start := day(2024, 1, 2)
	end := day(2024, 1, 4)
	fetched := []model.Bar{
		bar(start, 100),
		bar(end, 102),
	}
	merged := []model.Bar{
		bar(start, 100),
		bar(day(2024, 1, 3), 101),
		bar(end, 102),
	}

	s.repo.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(nil, nil).
		Once()
	s.market.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(fetched, nil)
	s.repo.EXPECT().
		SaveBars(s.ctx, "AAPL", fetched).
		Return(nil)
	s.repo.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(merged, nil).
		Once()

	got, err := s.svc.FetchBars(s.ctx, "AAPL", start, end)
	s.Require().NoError(err)
	s.Equal(merged, got)
}

func (s *TestSuite) TestFetchBars_PrefixAndSuffixGaps() {
	start := day(2024, 1, 1)
	midStart := day(2024, 1, 3)
	midEnd := day(2024, 1, 5)
	end := day(2024, 1, 7)

	cached := []model.Bar{
		bar(midStart, 103),
		bar(midEnd, 105),
	}
	prefix := []model.Bar{bar(start, 100), bar(day(2024, 1, 2), 101)}
	suffix := []model.Bar{bar(day(2024, 1, 6), 106), bar(end, 107)}
	merged := append(append(prefix, cached...), suffix...)

	s.repo.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(cached, nil).
		Once()
	s.market.EXPECT().
		FetchBars(s.ctx, "AAPL", start, midStart).
		Return(prefix, nil)
	s.repo.EXPECT().
		SaveBars(s.ctx, "AAPL", prefix).
		Return(nil)
	s.market.EXPECT().
		FetchBars(s.ctx, "AAPL", midEnd, end).
		Return(suffix, nil)
	s.repo.EXPECT().
		SaveBars(s.ctx, "AAPL", suffix).
		Return(nil)
	s.repo.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(merged, nil).
		Once()

	got, err := s.svc.FetchBars(s.ctx, "AAPL", start, end)
	s.Require().NoError(err)
	s.Equal(merged, got)
}

func (s *TestSuite) TestFetchBars_RepoFetchError() {
	start := day(2024, 1, 2)
	end := day(2024, 1, 4)

	s.repo.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(nil, errors.New("db down"))

	got, err := s.svc.FetchBars(s.ctx, "AAPL", start, end)
	s.Nil(got)
	s.EqualError(err, "fetch cached bars: db down")
}

func (s *TestSuite) TestFetchBars_ProviderError() {
	start := day(2024, 1, 2)
	end := day(2024, 1, 4)

	s.repo.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(nil, nil)
	s.market.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(nil, errors.New("provider down"))

	got, err := s.svc.FetchBars(s.ctx, "AAPL", start, end)
	s.Nil(got)
	s.EqualError(err, "fetch bars: provider down")
}

func (s *TestSuite) TestFetchBars_SaveError() {
	start := day(2024, 1, 2)
	end := day(2024, 1, 4)
	fetched := []model.Bar{bar(start, 100)}

	s.repo.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(nil, nil)
	s.market.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(fetched, nil)
	s.repo.EXPECT().
		SaveBars(s.ctx, "AAPL", fetched).
		Return(errors.New("save failed"))

	got, err := s.svc.FetchBars(s.ctx, "AAPL", start, end)
	s.Nil(got)
	s.EqualError(err, "save bars [2024-01-02 00:00:00 +0000 UTC, 2024-01-04 00:00:00 +0000 UTC]: save failed")
}

func (s *TestSuite) TestFetchBars_PostFillFetchError() {
	start := day(2024, 1, 2)
	end := day(2024, 1, 4)
	fetched := []model.Bar{bar(start, 100)}

	s.repo.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(nil, nil).
		Once()
	s.market.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(fetched, nil)
	s.repo.EXPECT().
		SaveBars(s.ctx, "AAPL", fetched).
		Return(nil)
	s.repo.EXPECT().
		FetchBars(s.ctx, "AAPL", start, end).
		Return(nil, errors.New("db down")).
		Once()

	got, err := s.svc.FetchBars(s.ctx, "AAPL", start, end)
	s.Nil(got)
	s.EqualError(err, "fetch cached bars (post-fill): db down")
}

func (s *TestSuite) TestCoversRange() {
	start := day(2024, 1, 2)
	end := day(2024, 1, 4)

	tests := []struct {
		name string
		bars []model.Bar
		want bool
	}{
		{
			name: "empty",
			want: false,
		},
		{
			name: "covers exactly",
			bars: []model.Bar{bar(start, 1), bar(end, 2)},
			want: true,
		},
		{
			name: "covers with margin",
			bars: []model.Bar{bar(day(2024, 1, 1), 1), bar(day(2024, 1, 5), 2)},
			want: true,
		},
		{
			name: "missing prefix",
			bars: []model.Bar{bar(day(2024, 1, 3), 1), bar(end, 2)},
			want: false,
		},
		{
			name: "missing suffix",
			bars: []model.Bar{bar(start, 1), bar(day(2024, 1, 3), 2)},
			want: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, coversRange(tt.bars, start, end))
		})
	}
}

func (s *TestSuite) TestMissingGaps() {
	start := day(2024, 1, 1)
	end := day(2024, 1, 7)

	tests := []struct {
		name string
		bars []model.Bar
		want []gap
	}{
		{
			name: "empty",
			want: []gap{{start, end}},
		},
		{
			name: "prefix only",
			bars: []model.Bar{bar(day(2024, 1, 3), 1), bar(end, 2)},
			want: []gap{{start, day(2024, 1, 3)}},
		},
		{
			name: "suffix only",
			bars: []model.Bar{bar(start, 1), bar(day(2024, 1, 5), 2)},
			want: []gap{{day(2024, 1, 5), end}},
		},
		{
			name: "prefix and suffix",
			bars: []model.Bar{bar(day(2024, 1, 3), 1), bar(day(2024, 1, 5), 2)},
			want: []gap{
				{start, day(2024, 1, 3)},
				{day(2024, 1, 5), end},
			},
		},
		{
			name: "fully covered",
			bars: []model.Bar{bar(start, 1), bar(end, 2)},
			want: nil,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, missingGaps(tt.bars, start, end))
		})
	}
}
