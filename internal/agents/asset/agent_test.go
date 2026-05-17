package asset

import (
	"context"
	"errors"
	"testing"
	"time"

	depmock "github.com/kasaderos/camel/gen/mocks/asset"
	"github.com/kasaderos/camel/internal/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AssetAgentTestSuite struct {
	suite.Suite
	ctx      context.Context
	repo     *depmock.AgentRepository
	market   *depmock.MarketService
	exchange *depmock.Exchanger
	agent    *Agent
}

func (s *AssetAgentTestSuite) SetupTest() {
	s.ctx = context.Background()
	s.repo = depmock.NewAgentRepository(s.T())
	s.market = depmock.NewMarketService(s.T())
	s.exchange = depmock.NewExchanger(s.T())

	assetAgent := model.AssetAgent{
		ID:       "test-agent-1",
		AssetID:  "AAPL",
		AssetQty: 10.0,
		Cash:     1000.0,
		State:    model.State{},
	}

	s.agent = NewAgent(assetAgent, s.repo, s.market, s.exchange)
}

func (s *AssetAgentTestSuite) TestFetchInfo() {
	info := s.agent.FetchInfo(s.ctx)

	s.Equal("test-agent-1", info.ID)
	s.Equal("AAPL", info.AssetID)
	s.Equal(10.0, info.AssetQty)
	s.Equal(1000.0, info.Cash)
}

func (s *AssetAgentTestSuite) TestFetchState() {
	state := s.agent.FetchState(s.ctx)
	s.NotNil(state)
}

func (s *AssetAgentTestSuite) TestBuyAsset_Success() {
	currentPrice := 150.0
	amount := 300.0
	expectedQty := 2.0 // 300 / 150 = 2.0

	// Mock FetchPrice
	s.exchange.EXPECT().
		FetchPrice(s.ctx, "AAPL").
		Return(currentPrice, nil).
		Once()

	// Mock CreateMarketOrder
	order := &model.Order{
		AssetID: "AAPL",
		Price:   150.0,
		Qty:     expectedQty,
		Status:  "filled",
	}
	s.exchange.EXPECT().
		CreateMarketOrder(s.ctx, "AAPL", expectedQty, model.OrderSideBuy).
		Return(order, nil).
		Once()

	// Mock Withdraw
	s.repo.EXPECT().
		Withdraw(s.ctx, "test-agent-1", 300.0).
		Return(nil).
		Once()

	// Execute
	err := s.agent.BuyAsset(s.ctx, amount)

	// Assert
	s.NoError(err)
	s.Equal(700.0, s.agent.Cash)    // 1000 - 300
	s.Equal(12.0, s.agent.AssetQty) // 10 + 2
}

func (s *AssetAgentTestSuite) TestBuyAsset_InvalidAmount() {
	err := s.agent.BuyAsset(s.ctx, 0)
	s.ErrorIs(err, model.ErrInvalidAmount)

	err = s.agent.BuyAsset(s.ctx, -100)
	s.ErrorIs(err, model.ErrInvalidAmount)
}

func (s *AssetAgentTestSuite) TestBuyAsset_InsufficientCash() {
	err := s.agent.BuyAsset(s.ctx, 2000.0)
	s.Error(err)
	s.Contains(err.Error(), "insufficient cash")
}

func (s *AssetAgentTestSuite) TestBuyAsset_FetchPriceError() {
	s.exchange.EXPECT().
		FetchPrice(s.ctx, "AAPL").
		Return(0.0, errors.New("price fetch failed")).
		Once()

	err := s.agent.BuyAsset(s.ctx, 100.0)
	s.Error(err)
	s.Contains(err.Error(), "failed to fetch current price")
}

func (s *AssetAgentTestSuite) TestBuyAsset_OrderCreationError() {
	s.exchange.EXPECT().
		FetchPrice(s.ctx, "AAPL").
		Return(150.0, nil).
		Once()

	s.exchange.EXPECT().
		CreateMarketOrder(s.ctx, "AAPL", mock.Anything, model.OrderSideBuy).
		Return(nil, errors.New("order failed")).
		Once()

	err := s.agent.BuyAsset(s.ctx, 100.0)
	s.Error(err)
	s.Contains(err.Error(), "failed to create buy order")
}

func (s *AssetAgentTestSuite) TestSellAsset_Success() {
	sellQty := 5.0
	price := 160.0

	// Mock CreateMarketOrder
	order := &model.Order{
		AssetID: "AAPL",
		Price:   price,
		Qty:     sellQty,
		Status:  "filled",
	}
	s.exchange.EXPECT().
		CreateMarketOrder(s.ctx, "AAPL", sellQty, model.OrderSideSell).
		Return(order, nil).
		Once()

	// Mock Deposit
	proceeds := price * sellQty // 800.0
	s.repo.EXPECT().
		Deposit(s.ctx, "test-agent-1", proceeds).
		Return(nil).
		Once()

	// Execute
	err := s.agent.SellAsset(s.ctx, sellQty)

	// Assert
	s.NoError(err)
	s.Equal(1800.0, s.agent.Cash)  // 1000 + 800
	s.Equal(5.0, s.agent.AssetQty) // 10 - 5
}

func (s *AssetAgentTestSuite) TestSellAsset_InvalidAmount() {
	err := s.agent.SellAsset(s.ctx, 0)
	s.ErrorIs(err, model.ErrInvalidAmount)

	err = s.agent.SellAsset(s.ctx, -5)
	s.ErrorIs(err, model.ErrInvalidAmount)
}

func (s *AssetAgentTestSuite) TestSellAsset_InsufficientQuantity() {
	err := s.agent.SellAsset(s.ctx, 20.0)
	s.Error(err)
	s.Contains(err.Error(), "insufficient asset quantity")
}

func (s *AssetAgentTestSuite) TestWithdraw() {
	amount := 200.0

	s.repo.EXPECT().
		Withdraw(s.ctx, "test-agent-1", amount).
		Return(nil).
		Once()

	_, err := s.agent.Withdraw(s.ctx, amount)
	s.NoError(err)
}

func (s *AssetAgentTestSuite) TestWithdraw_InvalidAmount() {
	_, err := s.agent.Withdraw(s.ctx, 0)
	s.ErrorIs(err, model.ErrInvalidAmount)

	_, err = s.agent.Withdraw(s.ctx, -100)
	s.ErrorIs(err, model.ErrInvalidAmount)
}

func (s *AssetAgentTestSuite) TestDeposit() {
	amount := 500.0

	s.repo.EXPECT().
		Deposit(s.ctx, "test-agent-1", amount).
		Return(nil).
		Once()

	_, err := s.agent.Deposit(s.ctx, amount)
	s.NoError(err)
}

func (s *AssetAgentTestSuite) TestDeposit_InvalidAmount() {
	_, err := s.agent.Deposit(s.ctx, 0)
	s.ErrorIs(err, model.ErrInvalidAmount)

	_, err = s.agent.Deposit(s.ctx, -100)
	s.ErrorIs(err, model.ErrInvalidAmount)
}

func (s *AssetAgentTestSuite) TestUpdateState() {
	now := time.Now()
	// Need at least 20 bars for EMA calculation (window = 20)
	bars := make([]model.Bar, 30)
	for i := 0; i < 30; i++ {
		bars[i] = model.Bar{
			Close:     100.0 + float64(i), // Increasing price trend
			Timestamp: now.Add(-time.Duration(30-i) * 24 * time.Hour),
		}
	}

	s.market.EXPECT().
		FetchBars(s.ctx, "AAPL", mock.Anything, mock.Anything).
		Return(bars, nil).
		Once()

	s.repo.EXPECT().
		UpdateState(s.ctx, "test-agent-1", mock.Anything).
		Return(nil).
		Once()

	_, err := s.agent.UpdateState(s.ctx)
	s.NoError(err)

	// Verify state was updated
	emaChange, ok := s.agent.State.EmaChange()
	s.True(ok)
	s.NotZero(emaChange)
}

func (s *AssetAgentTestSuite) TestUpdateState_FetchBarsError() {
	s.market.EXPECT().
		FetchBars(s.ctx, "AAPL", mock.Anything, mock.Anything).
		Return(nil, errors.New("market data unavailable")).
		Once()

	_, err := s.agent.UpdateState(s.ctx)
	s.Error(err)
	s.Contains(err.Error(), "failed to compute state")
}

func TestAssetAgentTestSuite(t *testing.T) {
	suite.Run(t, new(AssetAgentTestSuite))
}
