package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/samber/do/v2"

	assetrepo "github.com/kasaderos/camel/internal/repository/agents/asset"
	portfoliorepo "github.com/kasaderos/camel/internal/repository/agents/portfolio"
	marketrepo "github.com/kasaderos/camel/internal/repository/market"
	assetservice "github.com/kasaderos/camel/internal/service/agents/asset"
	portfolioservice "github.com/kasaderos/camel/internal/service/agents/portfolio"
	marketservice "github.com/kasaderos/camel/internal/service/market"
	"github.com/kasaderos/camel/pkg/alpaca"
)

func provide() (do.Injector, error) {
	injector := do.New()

	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	do.ProvideValue(injector, cfg)

	do.Provide(injector, func(i do.Injector) (*sqlx.DB, error) {
		cfg, err := do.Invoke[*config](i)
		if err != nil {
			return nil, err
		}

		dsn, err := cfg.Postgres.DSN()
		if err != nil {
			return nil, fmt.Errorf("postgres dsn: %w", err)
		}
		if strings.TrimSpace(dsn) == "" {
			return nil, errors.New("postgres config is required (set DATABASE_URL or POSTGRES_* env vars)")
		}

		// Register pgx driver for sqlx.
		_ = stdlib.GetDefaultDriver()

		db, err := sqlx.Connect("pgx", dsn)
		if err != nil {
			return nil, fmt.Errorf("connect db: %w", err)
		}

		db.SetConnMaxLifetime(5 * time.Minute)
		db.SetMaxIdleConns(4)
		db.SetMaxOpenConns(10)

		return db, nil
	})

	do.Provide(injector, func(i do.Injector) (marketservice.MarketProvider, error) {
		cfg, err := do.Invoke[*config](i)
		if err != nil {
			return nil, err
		}
		if strings.TrimSpace(cfg.Alpaca.APIKey) == "" || strings.TrimSpace(cfg.Alpaca.Secret) == "" {
			return nil, errors.New("alpaca-key and alpaca-secret are required (or set APCA_API_KEY_ID/APCA_API_SECRET_KEY)")
		}

		client, err := alpaca.NewMarketDataClient(cfg.Alpaca.APIKey, cfg.Alpaca.Secret, cfg.Alpaca.MarketURL)
		if err != nil {
			return nil, fmt.Errorf("create alpaca market client: %w", err)
		}

		return client, nil
	})

	do.Provide(injector, func(i do.Injector) (*marketrepo.Repository, error) {
		db, err := do.Invoke[*sqlx.DB](i)
		if err != nil {
			return nil, err
		}
		return marketrepo.New(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*marketservice.Service, error) {
		provider, err := do.Invoke[marketservice.MarketProvider](i)
		if err != nil {
			return nil, err
		}
		repo, err := do.Invoke[*marketrepo.Repository](i)
		if err != nil {
			return nil, err
		}
		return marketservice.New(provider, repo), nil
	})

	do.Provide(injector, func(i do.Injector) (*assetrepo.AgentRepository, error) {
		db := do.MustInvoke[*sqlx.DB](i)
		return assetrepo.New(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*assetservice.AssetAgentService, error) {
		repo, err := do.Invoke[*assetrepo.AgentRepository](i)
		if err != nil {
			return nil, err
		}
		mkt, err := do.Invoke[*marketservice.Service](i)
		if err != nil {
			return nil, err
		}
		return assetservice.New(repo, mkt), nil
	})

	do.Provide(injector, func(i do.Injector) (*portfoliorepo.AgentRepository, error) {
		db, err := do.Invoke[*sqlx.DB](i)
		if err != nil {
			return nil, err
		}
		return portfoliorepo.New(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*portfolioservice.PortfolioAgentService, error) {
		assetSvc, err := do.Invoke[*assetservice.AssetAgentService](i)
		if err != nil {
			return nil, err
		}
		repo, err := do.Invoke[*portfoliorepo.AgentRepository](i)
		if err != nil {
			return nil, err
		}
		return portfolioservice.New(assetSvc, repo), nil
	})

	return injector, nil
}

func terminate(injector do.Injector) error {
	db, err := do.Invoke[*sqlx.DB](injector)
	if err == nil && db != nil {
		_ = db.Close()
	}

	return nil
}
