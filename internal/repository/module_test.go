package repository_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"

	"github.com/profmoo/mnemosyne/db"
	"github.com/profmoo/mnemosyne/internal/listens"
	"github.com/profmoo/mnemosyne/internal/repository"
)

type ModuleSuite struct {
	suite.Suite
}

func TestModuleSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ModuleSuite))
}

func (s *ModuleSuite) TestProvidesSeparateListenCapabilities() {
	database, err := db.Open(s.T().Context(), ":memory:")
	s.Require().NoError(err)
	s.T().Cleanup(func() { s.NoError(database.Close()) })

	var creator listens.ListenCreator
	var getter listens.ListenGetter
	var lister listens.ListenLister
	app := fx.New(
		fx.NopLogger,
		fx.Supply(database),
		repository.Module,
		fx.Populate(&creator, &getter, &lister),
	)
	s.Require().NoError(app.Err())

	expected := repositoryListen(uuid.New())
	s.Require().NoError(creator.Create(s.T().Context(), expected))
	actual, err := getter.Get(s.T().Context(), *expected.ID)
	s.Require().NoError(err)
	s.Equal(expected, actual)
	items, err := lister.List(s.T().Context())
	s.Require().NoError(err)
	s.Equal([]listens.Listen{*expected}, items)
}
