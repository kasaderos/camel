package portfolio

type Service struct {
	exchange  Exchanger
	repo      PortfolioRepository
	analytics AnalyticsService
	taskRepo  TaskRepository
}

func New(
	exchange Exchanger,
	repo PortfolioRepository,
	analytics AnalyticsService,
	taskRepo TaskRepository,
) *Service {
	return &Service{
		exchange:  exchange,
		repo:      repo,
		analytics: analytics,
		taskRepo:  taskRepo,
	}
}
