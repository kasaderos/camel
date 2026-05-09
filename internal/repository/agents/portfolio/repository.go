package portfolio

import (
	"github.com/jmoiron/sqlx"
)

type AgentRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *AgentRepository {
	return &AgentRepository{db: db}
}
