package repositories

import (
	"context"
	"go-boilerplate/datasources/boilerplate_database"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func Test_NewGuestRepository(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	assert.NoError(t, err, "failed to create mock db")
	defer mockDB.Close()

	tests := []struct {
		name               string
		databaseConnection *boilerplate_database.BoilerplateDatabase
		expectNil          bool
	}{
		{
			name: "create guest repository with database connection",
			databaseConnection: &boilerplate_database.BoilerplateDatabase{
				Master: sqlx.NewDb(mockDB, "sqlmock"),
				Slave:  sqlx.NewDb(mockDB, "sqlmock"),
			},
			expectNil: false,
		},
		{
			name:               "create guest repository without database connection",
			databaseConnection: nil,
			expectNil:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewGuestRepository(tt.databaseConnection)

			if tt.expectNil {
				assert.Nil(t, repo, "NewGuestRepository() expected nil, got %v", repo)
			} else {
				assert.NotNil(t, repo, "NewGuestRepository() expected non-nil repository, got nil")
				assert.Equal(t, tt.databaseConnection, repo.db, "NewGuestRepository() db mismatch")
			}
		})
	}
}

func Test_GuestRepository_WithTransaction(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	assert.NoError(t, err, "failed to create mock db")
	defer mockDB.Close()

	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	databaseConnection := &boilerplate_database.BoilerplateDatabase{
		Master: sqlxDB,
		Slave:  sqlxDB,
	}

	tests := []struct {
		name      string
		setupRepo func() *GuestRepository
		setupMock func()
		tx        IBoilerplateDatabaseTransaction
		expectNil bool
		validate  func(t *testing.T, repo IGuestRepository)
	}{
		{
			name: "with transaction creates new repository with transaction",
			setupRepo: func() *GuestRepository {
				return NewGuestRepository(databaseConnection)
			},
			setupMock: func() {
				mock.ExpectBegin()
			},
			tx:        nil,
			expectNil: false,
			validate: func(t *testing.T, repo IGuestRepository) {
				assert.NotNil(t, repo, "WithTransaction() expected non-nil repository")
				guestRepo, ok := repo.(*GuestRepository)
				assert.True(t, ok, "WithTransaction() expected *GuestRepository type")
				assert.NotNil(t, guestRepo.tx, "WithTransaction() expected non-nil transaction")
				assert.NotNil(t, guestRepo.db, "WithTransaction() expected non-nil database connection")
			},
		},
		{
			name: "with nil transaction",
			setupRepo: func() *GuestRepository {
				return NewGuestRepository(databaseConnection)
			},
			setupMock: func() {},
			tx:        nil,
			expectNil: false,
			validate: func(t *testing.T, repo IGuestRepository) {
				assert.NotNil(t, repo, "WithTransaction() expected non-nil repository")
				guestRepo, ok := repo.(*GuestRepository)
				assert.True(t, ok, "WithTransaction() expected *GuestRepository type")
				assert.Nil(t, guestRepo.tx, "WithTransaction() expected nil transaction")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			repo := tt.setupRepo()

			var tx IBoilerplateDatabaseTransaction
			if tt.name == "with transaction creates new repository with transaction" {
				var err error
				tx, err = repo.BeginTransaction(context.Background())
				assert.NoError(t, err, "BeginTransaction() error")
			} else {
				tx = tt.tx
			}

			newRepo := repo.WithTransaction(tx)

			if tt.expectNil {
				assert.Nil(t, newRepo, "WithTransaction() expected nil, got %v", newRepo)
			} else {
				assert.NotNil(t, newRepo, "WithTransaction() expected non-nil repository, got nil")
			}

			if tt.validate != nil {
				tt.validate(t, newRepo)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "unfulfilled expectations")
		})
	}
}
