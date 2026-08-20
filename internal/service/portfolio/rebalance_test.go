package portfolio

import (
	"errors"
	"time"

	"github.com/kasaderos/camel/internal/model"
	"github.com/kasaderos/camel/internal/service/portfolio/mocks"
)

func (s *TestSuite) TestBuildTask() {
	const stockID model.StockID = "AAPL"
	const portfolioID = "portfolio-1"

	tests := []struct {
		name      string
		stock     model.PortfolioStock
		targetSum float64
		price     float64
		wantNil   bool
		wantSide  model.OrderSide
		wantQty   float64
	}{
		{
			name: "skip when sums are equal",
			stock: model.PortfolioStock{
				StockID:  stockID,
				Quantity: 10,
			},
			targetSum: 100,
			price:     10,
			wantNil:   true,
		},
		{
			name: "skip when abs difference is below 0.01",
			stock: model.PortfolioStock{
				StockID:  stockID,
				Quantity: 10,
			},
			targetSum: 100.009,
			price:     10,
			wantNil:   true,
		},
		{
			name: "buy when quantity is greater than 1",
			stock: model.PortfolioStock{
				StockID:  stockID,
				Quantity: 1,
			},
			targetSum: 300,
			price:     100,
			wantSide:  model.OrderSideBuy,
			wantQty:   2,
		},
		{
			name: "buy when quantity is exactly 1",
			stock: model.PortfolioStock{
				StockID:  stockID,
				Quantity: 1,
			},
			targetSum: 200,
			price:     100,
			wantSide:  model.OrderSideBuy,
			wantQty:   1,
		},
		{
			name: "skip when quantity change is below 1",
			stock: model.PortfolioStock{
				StockID:  stockID,
				Quantity: 1,
			},
			targetSum: 150,
			price:     100,
			wantNil:   true,
		},
		{
			name: "sell when current sum exceeds target",
			stock: model.PortfolioStock{
				StockID:  stockID,
				Quantity: 3,
			},
			targetSum: 100,
			price:     100,
			wantSide:  model.OrderSideSell,
			wantQty:   2,
		},
		{
			name: "sell all when target is zero",
			stock: model.PortfolioStock{
				StockID:  stockID,
				Quantity: 2,
			},
			targetSum: 0,
			price:     100,
			wantSide:  model.OrderSideSell,
			wantQty:   2,
		},
		{
			name: "skip sell all when target is zero and quantity is zero",
			stock: model.PortfolioStock{
				StockID:  stockID,
				Quantity: 0,
			},
			targetSum: 0,
			price:     100,
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := buildTask(portfolioID, tt.stock, tt.targetSum, tt.price)
			if tt.wantNil {
				s.Nil(got)
				return
			}

			s.Require().NotNil(got)
			s.Equal(portfolioID, got.PortfolioID)
			s.Equal(tt.stock.StockID, got.StockID)
			s.Equal(tt.wantSide, got.Side)
			s.InDelta(tt.wantQty, got.Quantity, 1e-9)
			s.Equal(model.TaskStatusCreated, got.Status)
		})
	}
}

func (s *TestSuite) TestPrepareTasks() {
	tests := []struct {
		name          string
		portfolio     model.Portfolio
		currentPrices map[model.StockID]float64
		targetWeights map[model.StockID]float64
		wantErr       string
		want          []model.Task
	}{
		{
			name: "no stocks",
			portfolio: model.Portfolio{
				Cost: 1000,
			},
			currentPrices: map[model.StockID]float64{},
			targetWeights: map[model.StockID]float64{},
			want:          []model.Task{},
		},
		{
			name: "error when price is missing",
			portfolio: model.Portfolio{
				Cost: 1000,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 1},
				},
			},
			currentPrices: map[model.StockID]float64{},
			targetWeights: map[model.StockID]float64{"AAPL": 0.3},
			wantErr:       "price not found for AAPL",
		},
		{
			name: "skip when already at target",
			portfolio: model.Portfolio{
				Cost: 1000,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 1},
				},
			},
			currentPrices: map[model.StockID]float64{"AAPL": 100},
			targetWeights: map[model.StockID]float64{"AAPL": 0.1},
			want:          []model.Task{},
		},
		{
			name: "skip when quantity change is below 1",
			portfolio: model.Portfolio{
				Cost: 1000,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 1},
				},
			},
			currentPrices: map[model.StockID]float64{"AAPL": 100},
			targetWeights: map[model.StockID]float64{"AAPL": 0.15},
			want:          []model.Task{},
		},
		{
			name: "buy when underweight",
			portfolio: model.Portfolio{
				Cost: 1000,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 1},
				},
			},
			currentPrices: map[model.StockID]float64{"AAPL": 100},
			targetWeights: map[model.StockID]float64{"AAPL": 0.3},
			want: []model.Task{
				{
					StockID:  "AAPL",
					Side:     model.OrderSideBuy,
					Quantity: 2,
					Status:   model.TaskStatusCreated,
				},
			},
		},
		{
			name: "sell when overweight",
			portfolio: model.Portfolio{
				Cost: 1000,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 3},
				},
			},
			currentPrices: map[model.StockID]float64{"AAPL": 100},
			targetWeights: map[model.StockID]float64{"AAPL": 0.1},
			want: []model.Task{
				{
					StockID:  "AAPL",
					Side:     model.OrderSideSell,
					Quantity: 2,
					Status:   model.TaskStatusCreated,
				},
			},
		},
		{
			name: "sell all when weight is missing",
			portfolio: model.Portfolio{
				Cost: 1000,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 2},
				},
			},
			currentPrices: map[model.StockID]float64{"AAPL": 100},
			targetWeights: map[model.StockID]float64{},
			want: []model.Task{
				{
					StockID:  "AAPL",
					Side:     model.OrderSideSell,
					Quantity: 2,
					Status:   model.TaskStatusCreated,
				},
			},
		},
		{
			name: "sell all when target sum is zero and quantity is positive",
			portfolio: model.Portfolio{
				Cost: 1000,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 2},
				},
			},
			currentPrices: map[model.StockID]float64{"AAPL": 100},
			targetWeights: map[model.StockID]float64{"AAPL": 0},
			want: []model.Task{
				{
					StockID:  "AAPL",
					Side:     model.OrderSideSell,
					Quantity: 2,
					Status:   model.TaskStatusCreated,
				},
			},
		},
		{
			name: "buy sell and skip across stocks",
			portfolio: model.Portfolio{
				Cost: 1000,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 0},
					{StockID: "MSFT", Quantity: 5},
					{StockID: "GOOG", Quantity: 1},
				},
			},
			currentPrices: map[model.StockID]float64{
				"AAPL": 100,
				"MSFT": 100,
				"GOOG": 100,
			},
			targetWeights: map[model.StockID]float64{
				"AAPL": 0.3,
				"MSFT": 0.2,
				"GOOG": 0.1,
			},
			want: []model.Task{
				{
					StockID:  "AAPL",
					Side:     model.OrderSideBuy,
					Quantity: 3,
					Status:   model.TaskStatusCreated,
				},
				{
					StockID:  "MSFT",
					Side:     model.OrderSideSell,
					Quantity: 3,
					Status:   model.TaskStatusCreated,
				},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := prepareTasks(tt.portfolio, tt.currentPrices, tt.targetWeights)
			if tt.wantErr != "" {
				s.Require().EqualError(err, tt.wantErr)
				s.Nil(got)
				return
			}

			s.Require().NoError(err)
			s.Require().Len(got, len(tt.want))
			for i, want := range tt.want {
				s.Require().NotNil(got[i])
				s.Equal(want.StockID, got[i].StockID)
				s.Equal(want.Side, got[i].Side)
				s.InDelta(want.Quantity, got[i].Quantity, 1e-9)
				s.Equal(want.Status, got[i].Status)
			}
		})
	}
}

func (s *TestSuite) TestBuildTargetWeights() {
	const threshold = 0.01

	tests := []struct {
		name      string
		portfolio model.Portfolio
		scores    map[model.StockID]float64
		wantNil   bool
		want      map[model.StockID]float64
	}{
		{
			name:      "no stocks",
			portfolio: model.Portfolio{},
			scores:    map[model.StockID]float64{},
			wantNil:   true,
		},
		{
			name: "all scores at or below threshold",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL"},
					{StockID: "MSFT"},
				},
			},
			scores: map[model.StockID]float64{
				"AAPL": 0.01,
				"MSFT": 0.009,
			},
			wantNil: true,
		},
		{
			name: "single stock above threshold is capped",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL"},
				},
			},
			scores: map[model.StockID]float64{
				"AAPL": 1,
			},
			want: map[model.StockID]float64{
				"AAPL": 0.3,
			},
		},
		{
			name: "proportional weights for two stocks",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL"},
					{StockID: "MSFT"},
				},
			},
			scores: map[model.StockID]float64{
				"AAPL": 1,
				"MSFT": 3,
			},
			want: map[model.StockID]float64{
				"AAPL": 0.25,
				"MSFT": 0.75,
			},
		},
		{
			name: "below-threshold stock gets zero weight",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL"},
					{StockID: "MSFT"},
					{StockID: "GOOG"},
				},
			},
			scores: map[model.StockID]float64{
				"AAPL": 1,
				"MSFT": 1,
				"GOOG": 0.005,
			},
			want: map[model.StockID]float64{
				"AAPL": 0.5,
				"MSFT": 0.5,
				"GOOG": 0,
			},
		},
		{
			name: "missing score is treated as zero",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL"},
					{StockID: "MSFT"},
				},
			},
			scores: map[model.StockID]float64{
				"AAPL": 2,
			},
			want: map[model.StockID]float64{
				"AAPL": 0.3,
				"MSFT": 0,
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := buildTargetWeights(tt.portfolio, tt.scores, threshold)
			s.Require().NoError(err)
			if tt.wantNil {
				s.Nil(got)
				return
			}

			s.Require().NotNil(got)
			s.Require().Len(got, len(tt.want))
			for stockID, wantWeight := range tt.want {
				s.InDelta(wantWeight, got[stockID], 1e-9)
			}
		})
	}
}

func (s *TestSuite) TestCalcPortfolioCost() {
	tests := []struct {
		name                string
		portfolio           model.Portfolio
		currentPrices       map[model.StockID]float64
		possibleDailyChange float64
		want                float64
	}{
		{
			name: "cash only",
			portfolio: model.Portfolio{
				Cash: 1000,
			},
			currentPrices:       map[model.StockID]float64{},
			possibleDailyChange: 0.01,
			want:                990,
		},
		{
			name:                "empty portfolio",
			portfolio:           model.Portfolio{},
			currentPrices:       map[model.StockID]float64{},
			possibleDailyChange: 0.01,
			want:                0,
		},
		{
			name: "stocks only",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 2},
					{StockID: "MSFT", Quantity: 5},
				},
			},
			currentPrices: map[model.StockID]float64{
				"AAPL": 100,
				"MSFT": 80,
			},
			possibleDailyChange: 0.01,
			want:                594,
		},
		{
			name: "cash and stocks",
			portfolio: model.Portfolio{
				Cash: 300,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 10},
					{StockID: "MSFT", Quantity: 5},
				},
			},
			currentPrices: map[model.StockID]float64{
				"AAPL": 100,
				"MSFT": 80,
			},
			possibleDailyChange: 0.01,
			want:                1683,
		},
		{
			name: "zero quantity stock adds nothing",
			portfolio: model.Portfolio{
				Cash: 500,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 0},
				},
			},
			currentPrices: map[model.StockID]float64{
				"AAPL": 100,
			},
			possibleDailyChange: 0.01,
			want:                495,
		},
		{
			name: "stored cost is ignored",
			portfolio: model.Portfolio{
				Cash: 200,
				Cost: 9999,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 3},
				},
			},
			currentPrices: map[model.StockID]float64{
				"AAPL": 50,
			},
			possibleDailyChange: 0.01,
			want:                346.5,
		},
		{
			name: "missing price is treated as zero",
			portfolio: model.Portfolio{
				Cash: 100,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 2},
				},
			},
			currentPrices:       map[model.StockID]float64{},
			possibleDailyChange: 0.01,
			want:                99,
		},
		{
			name: "no haircut when daily change is zero",
			portfolio: model.Portfolio{
				Cash: 300,
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL", Quantity: 10},
				},
			},
			currentPrices: map[model.StockID]float64{
				"AAPL": 100,
			},
			possibleDailyChange: 0,
			want:                1300,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			got := calcPortfolioCost(tt.portfolio, tt.currentPrices, tt.possibleDailyChange)
			s.InDelta(tt.want, got, 1e-9)
		})
	}
}

func (s *TestSuite) TestFetchStockPrices() {
	tests := []struct {
		name      string
		portfolio model.Portfolio
		setup     func(exchange *mocks.Exchanger)
		want      map[model.StockID]float64
		wantErr   string
	}{
		{
			name:      "no stocks",
			portfolio: model.Portfolio{},
			setup:     func(*mocks.Exchanger) {},
			want:      map[model.StockID]float64{},
		},
		{
			name: "one stock",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL"},
				},
			},
			setup: func(exchange *mocks.Exchanger) {
				exchange.EXPECT().
					FetchPrice(s.ctx, "AAPL").
					Return(150.0, time.Time{}, nil)
			},
			want: map[model.StockID]float64{
				"AAPL": 150,
			},
		},
		{
			name: "multiple stocks",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL"},
					{StockID: "MSFT"},
				},
			},
			setup: func(exchange *mocks.Exchanger) {
				exchange.EXPECT().
					FetchPrice(s.ctx, "AAPL").
					Return(100.0, time.Time{}, nil)
				exchange.EXPECT().
					FetchPrice(s.ctx, "MSFT").
					Return(80.0, time.Time{}, nil)
			},
			want: map[model.StockID]float64{
				"AAPL": 100,
				"MSFT": 80,
			},
		},
		{
			name: "fetch error",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL"},
					{StockID: "MSFT"},
				},
			},
			setup: func(exchange *mocks.Exchanger) {
				exchange.EXPECT().
					FetchPrice(s.ctx, "AAPL").
					Return(0.0, time.Time{}, errors.New("unavailable"))
			},
			wantErr: "fetch price: unavailable",
		},
		{
			name: "zero price",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL"},
				},
			},
			setup: func(exchange *mocks.Exchanger) {
				exchange.EXPECT().
					FetchPrice(s.ctx, "AAPL").
					Return(0.0, time.Time{}, nil)
			},
			wantErr: "price is not positive: AAPL",
		},
		{
			name: "negative price",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL"},
				},
			},
			setup: func(exchange *mocks.Exchanger) {
				exchange.EXPECT().
					FetchPrice(s.ctx, "AAPL").
					Return(-1.0, time.Time{}, nil)
			},
			wantErr: "price is not positive: AAPL",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			exchange := mocks.NewExchanger(s.T())
			svc := &Service{exchange: exchange}
			tt.setup(exchange)

			got, err := svc.fetchStockPrices(s.ctx, tt.portfolio)
			if tt.wantErr != "" {
				s.Require().EqualError(err, tt.wantErr)
				s.Nil(got)
				return
			}

			s.Require().NoError(err)
			s.Equal(tt.want, got)
		})
	}
}

func (s *TestSuite) TestFetchStockScores() {
	tests := []struct {
		name      string
		portfolio model.Portfolio
		setup     func(analytics *mocks.AnalyticsService)
		want      map[model.StockID]float64
		wantErr   string
	}{
		{
			name:      "no stocks",
			portfolio: model.Portfolio{},
			setup: func(analytics *mocks.AnalyticsService) {
				analytics.EXPECT().
					FetchStockScores(s.ctx, []model.StockID{}).
					Return(map[model.StockID]float64{}, nil)
			},
			want: map[model.StockID]float64{},
		},
		{
			name: "one stock",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL"},
				},
			},
			setup: func(analytics *mocks.AnalyticsService) {
				analytics.EXPECT().
					FetchStockScores(s.ctx, []model.StockID{"AAPL"}).
					Return(map[model.StockID]float64{"AAPL": 0.42}, nil)
			},
			want: map[model.StockID]float64{
				"AAPL": 0.42,
			},
		},
		{
			name: "multiple stocks in portfolio order",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "MSFT"},
					{StockID: "AAPL"},
				},
			},
			setup: func(analytics *mocks.AnalyticsService) {
				analytics.EXPECT().
					FetchStockScores(s.ctx, []model.StockID{"MSFT", "AAPL"}).
					Return(map[model.StockID]float64{
						"MSFT": 0.2,
						"AAPL": 0.8,
					}, nil)
			},
			want: map[model.StockID]float64{
				"MSFT": 0.2,
				"AAPL": 0.8,
			},
		},
		{
			name: "fetch error",
			portfolio: model.Portfolio{
				Stocks: []model.PortfolioStock{
					{StockID: "AAPL"},
				},
			},
			setup: func(analytics *mocks.AnalyticsService) {
				analytics.EXPECT().
					FetchStockScores(s.ctx, []model.StockID{"AAPL"}).
					Return(nil, errors.New("unavailable"))
			},
			wantErr: "fetch stock scores: unavailable",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			analytics := mocks.NewAnalyticsService(s.T())
			svc := &Service{analytics: analytics}
			tt.setup(analytics)

			got, err := svc.fetchStockScores(s.ctx, tt.portfolio)
			if tt.wantErr != "" {
				s.Require().EqualError(err, tt.wantErr)
				s.Nil(got)
				return
			}

			s.Require().NoError(err)
			s.Equal(tt.want, got)
		})
	}
}
