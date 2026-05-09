package main

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/ilyakaznacheev/cleanenv"
)

type config struct {
	Postgres Postgres
	Alpaca   Alpaca
}

type Postgres struct {
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     int    `env:"POSTGRES_PORT" env-default:"5432"`
	User     string `env:"POSTGRES_USER" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD"`
	Database string `env:"POSTGRES_DATABASE" env-default:"postgres"`
	SSLMode  string `env:"POSTGRES_SSLMODE" env-default:"disable"`
}

func (p Postgres) DSN() (string, error) {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(p.User, p.Password),
		Host:   p.Host + ":" + strconv.Itoa(p.Port),
		Path:   p.Database,
	}

	q := url.Values{}
	if p.SSLMode != "" {
		q.Set("sslmode", p.SSLMode)
	}

	u.RawQuery = q.Encode()

	return u.String(), nil
}

type Alpaca struct {
	APIKey    string `env:"APCA_API_KEY_ID" env-description:"Alpaca API key"`
	Secret    string `env:"APCA_API_SECRET_KEY" env-description:"Alpaca API secret"`
	MarketURL string `env:"ALPACA_MARKETDATA_URL" env-default:""`
}

func loadConfig() (*config, error) {
	var cfg config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read env config: %w", err)
	}

	return &cfg, nil
}
