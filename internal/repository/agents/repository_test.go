package agents

import (
	"context"
	"testing"

	"github.com/kasaderos/camel/pkg/testutils/testsuites"
)

// setupTestDB creates a test database with migrations applied
func setupTestDB(t *testing.T) (*AgentRepository, *testsuites.PostgresContainer) {
	t.Helper()

	ctx := context.Background()
	pgContainer := testsuites.NewPostgresContainer(ctx, t)

	// Run migrations
	pgContainer.RunMigrations(t, "../../../migrations")

	repo := New(pgContainer.DB)

	t.Cleanup(func() {
		pgContainer.Close(ctx, t)
	})

	return repo, pgContainer
}
