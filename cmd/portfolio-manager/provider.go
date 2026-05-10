package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/kasaderos/camel/internal/agents/asset"
	"github.com/kasaderos/camel/internal/agents/portfolio"
	agents "github.com/kasaderos/camel/internal/repository/agents"
	marketrepo "github.com/kasaderos/camel/internal/repository/market"
	marketservice "github.com/kasaderos/camel/internal/service/market"
	"github.com/kasaderos/camel/pkg/alpaca"
	"github.com/samber/do/v2"
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

	do.Provide(injector, func(i do.Injector) (*agents.AgentRepository, error) {
		db, err := do.Invoke[*sqlx.DB](i)
		if err != nil {
			return nil, err
		}
		return agents.New(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*marketrepo.Repository, error) {
		db, err := do.Invoke[*sqlx.DB](i)
		if err != nil {
			return nil, err
		}
		return marketrepo.New(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*alpaca.MarketDataClient, error) {
		cfg, err := do.Invoke[*config](i)
		if err != nil {
			return nil, err
		}
		return alpaca.NewMarketDataClient(cfg.Alpaca.APIKey, cfg.Alpaca.Secret, cfg.Alpaca.MarketURL)
	})

	do.Provide(injector, func(i do.Injector) (*marketservice.Service, error) {
		client, err := do.Invoke[*alpaca.MarketDataClient](i)
		if err != nil {
			return nil, err
		}
		repo, err := do.Invoke[*marketrepo.Repository](i)
		if err != nil {
			return nil, err
		}
		return marketservice.New(client, repo), nil
	})

	do.Provide(injector, func(i do.Injector) (*portfolio.Agent, error) {
		repo, err := do.Invoke[*agents.AgentRepository](i)
		if err != nil {
			return nil, err
		}
		market, err := do.Invoke[*marketservice.Service](i)
		if err != nil {
			return nil, err
		}

		assetAgentsInitFunc := func(ctx context.Context, id string) (portfolio.AssetAgent, error) {
			return asset.NewAgent(repo, market, asset.WithInitialize(ctx, id))
		}

		return portfolio.NewAgent(repo, assetAgentsInitFunc), nil
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
