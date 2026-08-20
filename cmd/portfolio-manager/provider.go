package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/samber/do/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	marketrepo "github.com/kasaderos/camel/internal/repository/market"
	portfolioRepo "github.com/kasaderos/camel/internal/repository/portfolio"
	analyticsService "github.com/kasaderos/camel/internal/service/analytics"
	exchangealpaca "github.com/kasaderos/camel/internal/service/exchange/alpaca"
	marketservice "github.com/kasaderos/camel/internal/service/market"
	portfolioService "github.com/kasaderos/camel/internal/service/portfolio"
)

func provide() (do.Injector, error) {
	injector := do.New()

	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	do.ProvideValue(injector, cfg)

	do.Provide(injector, func(i do.Injector) (*gorm.DB, error) {
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

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("connect db: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("sql db: %w", err)
		}

		sqlDB.SetConnMaxLifetime(5 * time.Minute)
		sqlDB.SetMaxIdleConns(4)
		sqlDB.SetMaxOpenConns(10)

		return db, nil
	})

	do.Provide(injector, func(i do.Injector) (*portfolioRepo.Repository, error) {
		db, err := do.Invoke[*gorm.DB](i)
		if err != nil {
			return nil, err
		}

		return portfolioRepo.New(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*marketrepo.Repository, error) {
		db, err := do.Invoke[*gorm.DB](i)
		if err != nil {
			return nil, err
		}

		return marketrepo.New(db), nil
	})

	do.Provide(injector, func(i do.Injector) (*exchangealpaca.Service, error) {
		cfg, err := do.Invoke[*config](i)
		if err != nil {
			return nil, err
		}

		return exchangealpaca.New(
			cfg.Alpaca.APIKey,
			cfg.Alpaca.Secret,
			cfg.Alpaca.TradingURL,
			cfg.Alpaca.MarketURL,
		), nil
	})

	do.Provide(injector, func(i do.Injector) (*marketservice.Service, error) {
		client, err := do.Invoke[*exchangealpaca.Service](i)
		if err != nil {
			return nil, err
		}

		repo, err := do.Invoke[*marketrepo.Repository](i)
		if err != nil {
			return nil, err
		}

		return marketservice.New(client, repo), nil
	})

	do.Provide(injector, func(i do.Injector) (*analyticsService.Service, error) {
		market, err := do.Invoke[*marketservice.Service](i)
		if err != nil {
			return nil, err
		}

		return analyticsService.NewService(market), nil
	})

	do.Provide(injector, func(i do.Injector) (*portfolioService.Service, error) {
		exchange, err := do.Invoke[*exchangealpaca.Service](i)
		if err != nil {
			return nil, err
		}

		repo, err := do.Invoke[*portfolioRepo.Repository](i)
		if err != nil {
			return nil, err
		}

		analytics, err := do.Invoke[*analyticsService.Service](i)
		if err != nil {
			return nil, err
		}

		return portfolioService.New(exchange, repo, analytics, repo), nil
	})

	return injector, nil
}

func terminate(injector do.Injector) error {
	db, err := do.Invoke[*gorm.DB](injector)
	if err == nil && db != nil {
		if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	}

	return nil
}
