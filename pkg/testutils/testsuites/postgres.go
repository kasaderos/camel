// Package testsuites provides reusable test suite components for integration testing.
//
// # Usage Patterns
//
// ## Pattern 1: Direct Container Usage (for helper functions)
//
// Use NewPostgresContainer directly in test helper functions:
//
//	func setupTestDB(t *testing.T) (*Repository, *testsuites.Postgres) {
//	    ctx := context.Background()
//	    pg := testsuites.NewPostgresContainer(ctx, t)
//	    pg.RunMigrations(t, "../../../migrations")
//	    repo := NewRepository(pg.DB)
//	    t.Cleanup(func() { pg.Close(ctx, t) })
//	    return repo, pg
//	}
//
// ## Pattern 2: Suite Embedding (for testify/suite)
//
// Embed PostgresSuite in your test suite:
//
//	type MyRepoTestSuite struct {
//	    testsuites.PostgresSuite
//	    repo *Repository
//	}
//
//	func (s *MyRepoTestSuite) SetupSuite() {
//	    s.PostgresSuite.SetupSuite()
//	    s.RunMigrations("../../../migrations")
//	    s.repo = NewRepository(s.DB)
//	}
package testsuites

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/kasaderos/camel/pkg/testutils/containers"
	"github.com/stretchr/testify/suite"
)

// Postgres wraps containers.Postgres to provide it via testsuites package
type Postgres = containers.Postgres

// NewPostgresContainer creates a new PostgreSQL container for testing
func NewPostgresContainer(ctx context.Context, t *testing.T) *Postgres {
	return containers.NewPostgresContainer(ctx, t)
}

// PostgresSuite is a test suite that provides a PostgreSQL database
type PostgresSuite struct {
	suite.Suite
	Postgres *Postgres
	DB       *sqlx.DB
	ctx      context.Context
}

// SetupSuite runs once before all tests in the suite
func (s *PostgresSuite) SetupSuite() {
	s.ctx = context.Background()
	s.Postgres = NewPostgresContainer(s.ctx, s.T())
	s.DB = s.Postgres.DB
}

// TearDownSuite runs once after all tests in the suite
func (s *PostgresSuite) TearDownSuite() {
	if s.Postgres != nil {
		s.Postgres.Close(s.ctx, s.T())
	}
}

// RunMigrations is a helper to run migrations in tests
func (s *PostgresSuite) RunMigrations(migrationsPath string) {
	s.Postgres.RunMigrations(s.T(), migrationsPath)
}

// Truncate is a helper to truncate tables between tests
func (s *PostgresSuite) Truncate(tables ...string) {
	s.Postgres.Truncate(s.T(), tables...)
}
