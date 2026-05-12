package testsuites

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer wraps a PostgreSQL testcontainer instance
type PostgresContainer struct {
	container testcontainers.Container
	DB        *sqlx.DB
	ConnStr   string
}

// NewPostgresContainer creates and starts a new PostgreSQL container for testing
func NewPostgresContainer(ctx context.Context, t *testing.T) *PostgresContainer {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get container port: %v", err)
	}

	connStr := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	return &PostgresContainer{
		container: container,
		DB:        db,
		ConnStr:   connStr,
	}
}

// RunMigrations runs database migrations from the given migrations directory
func (pc *PostgresContainer) RunMigrations(t *testing.T, migrationsPath string) {
	t.Helper()

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	m, err := migrate.New(
		fmt.Sprintf("file://%s", absPath),
		pc.ConnStr,
	)
	if err != nil {
		t.Fatalf("failed to create migrate instance: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("failed to run migrations: %v", err)
	}
}

// Close terminates the container and closes the database connection
func (pc *PostgresContainer) Close(ctx context.Context, t *testing.T) {
	t.Helper()

	if pc.DB != nil {
		if err := pc.DB.Close(); err != nil {
			t.Errorf("failed to close database: %v", err)
		}
	}

	if pc.container != nil {
		if err := pc.container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	}
}

// Truncate truncates all tables in the database for cleanup between tests
func (pc *PostgresContainer) Truncate(t *testing.T, tables ...string) {
	t.Helper()

	for _, table := range tables {
		_, err := pc.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Fatalf("failed to truncate table %s: %v", table, err)
		}
	}
}
