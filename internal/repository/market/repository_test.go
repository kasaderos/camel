package market

import (
	"testing"

	"github.com/kasaderos/camel/pkg/testutils/testsuites"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type RepositorySuite struct {
	testsuites.PostgresSuite

	repo *Repository
}

func (s *RepositorySuite) SetupSuite() {
	s.PostgresSuite.SetupSuite()
	s.RunMigrations("../../../migrations")

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		Conn: s.DB.DB,
	}), &gorm.Config{})
	s.Require().NoError(err)

	s.repo = New(gdb)
}

func (s *RepositorySuite) SetupTest() {
	s.Truncate("asset_bars")
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}
