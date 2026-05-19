package portfolio

import (
	"context"
	"testing"

	depmock "github.com/kasaderos/camel/gen/mocks/portfolio"
	"github.com/kasaderos/camel/internal/model"
	"github.com/stretchr/testify/suite"
)

type AgentTestSuite struct {
	suite.Suite

	ctx      context.Context
	repo     *depmock.AgentRepository
	exchange *depmock.Exchanger
	market   *depmock.MarketService
}

func (s *AgentTestSuite) SetupTest() {
	s.ctx = s.T().Context()
	s.repo = depmock.NewAgentRepository(s.T())
	s.exchange = depmock.NewExchanger(s.T())
	s.market = depmock.NewMarketService(s.T())
}

func TestPortfolioAgentTestSuite(t *testing.T) {
	suite.Run(t, new(AgentTestSuite))
}

func (s *AgentTestSuite) TestAdjustTargetSum() {
	tests := []struct {
		name      string
		assetID   string
		assetQty  float64
		targetSum float64
		mockPrice float64
		// expected order args
		expectedOrderSide string
		expectedOrderQty  float64
		expectedLeftQty   float64
	}{
		{
			name:      "no adjustment needed: over by less than one unit",
			assetID:   "AAPL",
			assetQty:  2.0,
			mockPrice: 600.0,
			// currentSum=1200, targetSum=700
			// currentSum-currentPrice=600 < targetSum=700 => skip
			targetSum:       700.0,
			expectedLeftQty: 2.0,
		},
		{
			name:      "no adjustment needed: sum diff < 0.01$",
			assetID:   "AAPL",
			assetQty:  2.0,
			mockPrice: 300.0,
			// currentSum=600, targetSum=600
			// currentSum-currentPrice = 0.0
			targetSum:       600.0,
			expectedLeftQty: 0.0,
		},
		{
			name:      "buy: current sum below target",
			assetID:   "AAPL",
			assetQty:  1.0,
			mockPrice: 500.0,
			// currentSum=500, targetSum=1500
			// deltaQty = 3 - 1 = 2 => buy ceil(2)=2
			targetSum:         1500.0,
			expectedOrderSide: model.OrderSideBuy,
			expectedOrderQty:  2.0,
			expectedLeftQty:   3.0,
		},
		{
			name:      "buy: fractional delta rounds up",
			assetID:   "AAPL",
			assetQty:  1.0,
			mockPrice: 300.0,
			// currentSum=300, targetSum=1000
			// deltaQty = ceil(1000/300 - 1) = ceil(2.33) = 3
			targetSum:         1000.0,
			expectedOrderSide: model.OrderSideBuy,
			expectedOrderQty:  3.0,
			expectedLeftQty:   4.0,
		},
		{
			name:      "sell: current sum above target by more than one unit",
			assetID:   "AAPL",
			assetQty:  5.0,
			mockPrice: 500.0,
			// currentSum=2500, targetSum=500
			// deltaQty = 1 - 5 = -4 => sell ceil(4)=4
			targetSum:         500.0,
			expectedOrderSide: model.OrderSideSell,
			expectedOrderQty:  4.0,
			expectedLeftQty:   1.0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			agentModel := model.PortfolioAgent{
				AssetID:  tc.assetID,
				AssetQty: tc.assetQty,
			}
			agent := &Agent{
				PortfolioAgent: agentModel,
				repo:           s.repo,
				market:         s.market,
				exchange:       s.exchange,
			}

			s.exchange.
				On("FetchPrice", s.ctx, tc.assetID).
				Return(tc.mockPrice, nil).
				Once()

			if tc.expectedOrderSide != "" {
				s.exchange.
					On("CreateMarketOrder",
						s.ctx,
						tc.assetID,
						tc.expectedOrderQty,
						tc.expectedOrderSide,
					).
					Return(&model.Order{
						AssetID: tc.assetID,
						Qty:     tc.expectedOrderQty,
						Side:    tc.expectedOrderSide,
					}, nil).
					Once()

				agentModel.AssetQty = tc.expectedLeftQty
				s.repo.
					On("UpdatePortfolioAgent",
						s.ctx,
						agentModel,
					).
					Return(nil).
					Once()
			}

			_, err := agent.AdjustTargetSum(s.ctx, tc.targetSum)
			s.Require().NoError(err)
		})
	}
}
