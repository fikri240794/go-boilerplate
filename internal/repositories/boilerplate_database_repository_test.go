package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-boilerplate/datasources/boilerplate_database"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/fikri240794/goqube"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

type testDBEntity struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type testEntityWithValidTags struct {
	ID        int    `db:"id" table:"test_table"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	CreatedAt string `db:"created_at"`
}

type testEntityWithDashTags struct {
	ID          int    `db:"id" table:"users"`
	Name        string `db:"name"`
	Password    string `db:"-"`
	InternalUse string `table:"-"`
}

type testEntityWithEmptyTags struct {
	ID       int    `db:"id" table:"products"`
	Name     string `db:"name"`
	Ignored1 string `db:""`
	Ignored2 string
}

type testEntityWithTableTag struct {
	ID   int    `db:"id" table:"orders"`
	Code string `db:"code"`
}

type testEntityWithValues struct {
	ID        int       `db:"id" table:"test_values_table"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	Age       int       `db:"age"`
	CreatedAt time.Time `db:"created_at"`
}

type testEntityWithDashTagValues struct {
	ID       int    `db:"id" table:"users_values"`
	Name     string `db:"name"`
	Password string `db:"-"`
	Internal string `table:"-"`
}

type testEntityWithEmptyTagValues struct {
	ID       int     `db:"id" table:"products_values"`
	Name     string  `db:"name"`
	Price    float64 `db:"price"`
	Ignored1 string  `db:""`
	Ignored2 string
}

type testEntityWithNilPointerValues struct {
	ID   int    `db:"id" table:"nullable_table"`
	Name string `db:"name"`
}

type testEntityForExec struct {
	ID   int    `db:"id" table:"exec_test_table"`
	Name string `db:"name"`
}

type testEntityForBulkExec struct {
	tableName string `table:"bulk_test_table"`
	ID        int    `db:"id" primary_key:"true" db_type:"int"`
	Name      string `db:"name" db_type:"text"`
}

func Test_newBoilerplateDatabaseStatement(t *testing.T) {
	tests := []struct {
		name     string
		stmt     *sqlx.Stmt
		validate func(t *testing.T, result *boilerplateDatabaseStatement)
	}{
		{
			name: "create statement with valid sqlx stmt",
			stmt: &sqlx.Stmt{},
			validate: func(t *testing.T, result *boilerplateDatabaseStatement) {
				assert.NotNil(t, result, "Expected result to be created, got nil")
				assert.NotNil(t, result.stmt, "Expected stmt to be initialized, got nil")
			},
		},
		{
			name: "create statement with nil stmt",
			stmt: nil,
			validate: func(t *testing.T, result *boilerplateDatabaseStatement) {
				assert.NotNil(t, result, "Expected result to be created, got nil")
				assert.Nil(t, result.stmt, "Expected stmt to be nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newBoilerplateDatabaseStatement(tt.stmt)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func Test_boilerplateDatabaseStatement_Exec(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		ctx         context.Context
		args        []interface{}
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "exec successfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("INSERT INTO test")
				mock.ExpectExec("INSERT INTO test").
					WithArgs("value1", "value2").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			ctx:         context.Background(),
			args:        []interface{}{"value1", "value2"},
			expectError: false,
		},
		{
			name: "exec with error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("INSERT INTO test")
				mock.ExpectExec("INSERT INTO test").
					WithArgs("value1").
					WillReturnError(errors.New("exec error"))
			},
			ctx:         context.Background(),
			args:        []interface{}{"value1"},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "Expected error, got nil")
				assert.Equal(t, "exec error", err.Error(), "Expected 'exec error', got '%s'", err.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			tt.setupMock(mock)

			sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
			stmt, err := sqlxDB.Preparex("INSERT INTO test")
			if err != nil {
				t.Fatalf("Failed to prepare statement: %v", err)
			}
			defer stmt.Close()

			dbStmt := newBoilerplateDatabaseStatement(stmt)
			err = dbStmt.Exec(tt.ctx, tt.args...)

			if tt.expectError {
				assert.NotNil(t, err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), "Unfulfilled expectations: %v", mock.ExpectationsWereMet())
		})
	}
}

func Test_boilerplateDatabaseStatement_Get(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		ctx         context.Context
		args        []interface{}
		expectError bool
		validate    func(t *testing.T, dest *testDBEntity, err error)
	}{
		{
			name: "get successfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test WHERE id = ?")
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "test_name")
				mock.ExpectQuery("SELECT (.+) FROM test WHERE id = ?").
					WithArgs(1).
					WillReturnRows(rows)
			},
			ctx:         context.Background(),
			args:        []interface{}{1},
			expectError: false,
			validate: func(t *testing.T, dest *testDBEntity, err error) {
				assert.NoError(t, err, "expected no error, got %v", err)
				assert.Equal(t, 1, dest.ID, "expected ID 1, got %d", dest.ID)
				assert.Equal(t, "test_name", dest.Name, "expected Name 'test_name', got %s", dest.Name)
			},
		},
		{
			name: "get with error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test WHERE id = ?")
				mock.ExpectQuery("SELECT (.+) FROM test WHERE id = ?").
					WithArgs(999).
					WillReturnError(errors.New("get error"))
			},
			ctx:         context.Background(),
			args:        []interface{}{999},
			expectError: true,
			validate: func(t *testing.T, dest *testDBEntity, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Equal(t, "get error", err.Error(), "expected 'get error', got %v", err.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			tt.setupMock(mock)

			sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
			stmt, err := sqlxDB.Preparex("SELECT * FROM test WHERE id = ?")
			if err != nil {
				t.Fatalf("Failed to prepare statement: %v", err)
			}
			defer stmt.Close()

			dbStmt := newBoilerplateDatabaseStatement(stmt)
			dest := &testDBEntity{}
			err = dbStmt.Get(tt.ctx, dest, tt.args...)

			if tt.expectError && err == nil {
				assert.NotNil(t, err, "Expected error but got nil")
			}

			if !tt.expectError && err != nil {
				assert.NoError(t, err, "Expected no error but got: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, dest, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				assert.NoError(t, err, "Unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_boilerplateDatabaseStatement_Select(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		ctx         context.Context
		args        []interface{}
		expectError bool
		validate    func(t *testing.T, dest *[]testDBEntity, err error)
	}{
		{
			name: "select successfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test")
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "test_name_1").
					AddRow(2, "test_name_2")
				mock.ExpectQuery("SELECT (.+) FROM test").
					WillReturnRows(rows)
			},
			ctx:         context.Background(),
			args:        []interface{}{},
			expectError: false,
			validate: func(t *testing.T, dest *[]testDBEntity, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error, got %v", err))
				assert.Equal(t, 2, len(*dest), fmt.Sprintf("expected 2 rows, got %d", len(*dest)))
				if len(*dest) >= 1 {
					assert.Equal(t, 1, (*dest)[0].ID, fmt.Sprintf("expected first row ID=1, got %d", (*dest)[0].ID))
					assert.Equal(t, "test_name_1", (*dest)[0].Name, fmt.Sprintf("expected first row Name='test_name_1', got %s", (*dest)[0].Name))
				}
				if len(*dest) >= 2 {
					assert.Equal(t, 2, (*dest)[1].ID, fmt.Sprintf("expected second row ID=2, got %d", (*dest)[1].ID))
					assert.Equal(t, "test_name_2", (*dest)[1].Name, fmt.Sprintf("expected second row Name='test_name_2', got %s", (*dest)[1].Name))
				}
			},
		},
		{
			name: "select with error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test")
				mock.ExpectQuery("SELECT (.+) FROM test").
					WillReturnError(errors.New("select error"))
			},
			ctx:         context.Background(),
			args:        []interface{}{},
			expectError: true,
			validate: func(t *testing.T, dest *[]testDBEntity, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Equal(t, "select error", err.Error(), fmt.Sprintf("expected 'select error', got %v", err.Error()))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			tt.setupMock(mock)

			sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
			stmt, err := sqlxDB.Preparex("SELECT * FROM test")
			if err != nil {
				t.Fatalf("Failed to prepare statement: %v", err)
			}
			defer stmt.Close()

			dbStmt := newBoilerplateDatabaseStatement(stmt)
			dest := &[]testDBEntity{}
			err = dbStmt.Select(tt.ctx, dest, tt.args...)

			if tt.expectError && err == nil {
				assert.NotNil(t, err, "Expected error but got nil")
			}

			if !tt.expectError && err != nil {
				assert.NoError(t, err, fmt.Sprintf("Expected no error but got: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, dest, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				assert.NoError(t, err, fmt.Sprintf("Unfulfilled expectations: %v", err))
			}
		})
	}
}

func Test_boilerplateDatabaseStatement_Close(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "close successfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT 1").WillBeClosed()
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error, got %v", err))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			tt.setupMock(mock)

			sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
			stmt, err := sqlxDB.Preparex("SELECT 1")
			if err != nil {
				t.Fatalf("Failed to prepare statement: %v", err)
			}

			dbStmt := newBoilerplateDatabaseStatement(stmt)
			err = dbStmt.Close()

			if tt.expectError {
				assert.NotNil(t, err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.NoError(t, err, fmt.Sprintf("Expected no error but got: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled expectations: %v", mock.ExpectationsWereMet()))
		})
	}
}

func Test_newBoilerplateDatabaseTransaction(t *testing.T) {
	tests := []struct {
		name     string
		tx       *sqlx.Tx
		validate func(t *testing.T, result *boilerplateDatabaseTransaction)
	}{
		{
			name: "create transaction with valid sqlx tx",
			tx:   &sqlx.Tx{},
			validate: func(t *testing.T, result *boilerplateDatabaseTransaction) {
				assert.NotNil(t, result, "Expected result to be created, got nil")
				assert.NotNil(t, result.tx, "Expected tx to be initialized, got nil")
			},
		},
		{
			name: "create transaction with nil tx",
			tx:   nil,
			validate: func(t *testing.T, result *boilerplateDatabaseTransaction) {
				assert.NotNil(t, result, "Expected result to be created, got nil")
				assert.Nil(t, result.tx, "Expected tx to be nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := newBoilerplateDatabaseTransaction(tt.tx)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func Test_boilerplateDatabaseTransaction_Commit(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "commit successfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit()
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error, got %v", err))
			},
		},
		{
			name: "commit with error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Equal(t, "error", err.Error(), fmt.Sprintf("expected 'error', got %v", err.Error()))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			tt.setupMock(mock)

			sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
			tx, err := sqlxDB.Beginx()
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}

			dbTx := newBoilerplateDatabaseTransaction(tx)
			err = dbTx.Commit()

			if tt.expectError {
				assert.NotNil(t, err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.NoError(t, err, fmt.Sprintf("Expected no error but got: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled expectations: %v", mock.ExpectationsWereMet()))
		})
	}
}

func Test_boilerplateDatabaseTransaction_DriverName(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(mock sqlmock.Sqlmock)
		driverName string
		validate   func(t *testing.T, result string)
	}{
		{
			name: "get driver name successfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
			},
			driverName: "sqlmock",
			validate: func(t *testing.T, result string) {
				assert.Equal(t, "sqlmock", result, fmt.Sprintf("expected driver name 'sqlmock', got %s", result))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			tt.setupMock(mock)

			sqlxDB := sqlx.NewDb(mockDB, tt.driverName)
			tx, err := sqlxDB.Beginx()
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}

			dbTx := newBoilerplateDatabaseTransaction(tx)
			result := dbTx.DriverName()

			if tt.validate != nil {
				tt.validate(t, result)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				assert.NoError(t, err, "Unfulfilled expectations: %v", err)
			}
		})
	}
}

func Test_boilerplateDatabaseTransaction_Prepare(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		ctx         context.Context
		query       string
		expectError bool
		validate    func(t *testing.T, stmt IBoilerplateDatabaseStatement, err error)
	}{
		{
			name: "prepare successfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("SELECT (.+) FROM test")
			},
			ctx:         context.Background(),
			query:       "SELECT * FROM test",
			expectError: false,
			validate: func(t *testing.T, stmt IBoilerplateDatabaseStatement, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error, got %v", err))
				assert.NotNil(t, stmt, "expected statement to be created, got nil")
			},
		},
		{
			name: "prepare with error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectPrepare("SELECT (.+) FROM test").
					WillReturnError(errors.New("prepare error"))
			},
			ctx:         context.Background(),
			query:       "SELECT * FROM test",
			expectError: true,
			validate: func(t *testing.T, stmt IBoilerplateDatabaseStatement, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Nil(t, stmt, "expected statement to be nil")
				assert.Equal(t, "prepare error", err.Error(), fmt.Sprintf("expected 'prepare error', got %v", err.Error()))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			tt.setupMock(mock)

			sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
			tx, err := sqlxDB.Beginx()
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}

			dbTx := newBoilerplateDatabaseTransaction(tx)
			stmt, err := dbTx.Prepare(tt.ctx, tt.query)

			if tt.expectError {
				assert.NotNil(t, err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.NoError(t, err, fmt.Sprintf("Expected no error but got: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, stmt, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled expectations: %v", mock.ExpectationsWereMet()))
		})
	}
}

func Test_boilerplateDatabaseTransaction_Rollback(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "rollback successfully",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback()
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error, got %v", err))
			},
		},
		{
			name: "rollback with error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectRollback().WillReturnError(errors.New("rollback error"))
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Equal(t, "error", err.Error(), fmt.Sprintf("expected 'error', got %v", err.Error()))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create sqlmock: %v", err)
			}
			defer mockDB.Close()

			tt.setupMock(mock)

			sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
			tx, err := sqlxDB.Beginx()
			if err != nil {
				t.Fatalf("Failed to begin transaction: %v", err)
			}

			dbTx := newBoilerplateDatabaseTransaction(tx)
			err = dbTx.Rollback()

			if tt.expectError {
				assert.NotNil(t, err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.NoError(t, err, fmt.Sprintf("Expected no error but got: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled expectations: %v", mock.ExpectationsWereMet()))
		})
	}
}

func Test_NewBoilerplateDatabaseRepository(t *testing.T) {
	tests := []struct {
		name     string
		setupDB  func() (*sqlx.DB, *sqlx.DB)
		validate func(t *testing.T, repo *BoilerplateDatabaseRepository[testDBEntity])
	}{
		{
			name: "create repository with valid database connections",
			setupDB: func() (*sqlx.DB, *sqlx.DB) {
				mockMasterDB, _, err := sqlmock.New()
				if err != nil {
					t.Fatalf("Failed to create master sqlmock: %v", err)
				}
				mockSlaveDB, _, err := sqlmock.New()
				if err != nil {
					t.Fatalf("Failed to create slave sqlmock: %v", err)
				}
				return sqlx.NewDb(mockMasterDB, "sqlmock"), sqlx.NewDb(mockSlaveDB, "sqlmock")
			},
			validate: func(t *testing.T, repo *BoilerplateDatabaseRepository[testDBEntity]) {
				assert.NotNil(t, repo, "Expected repository to be created, got nil")
				assert.NotNil(t, repo.db, "Expected db to be initialized, got nil")
				assert.NotNil(t, repo.db.Master, "Expected db.Master to be initialized, got nil")
				assert.NotNil(t, repo.db.Slave, "Expected db.Slave to be initialized, got nil")
				assert.Nil(t, repo.tx, "Expected tx to be nil for new repository")
			},
		},
		{
			name: "create repository with nil database",
			setupDB: func() (*sqlx.DB, *sqlx.DB) {
				return nil, nil
			},
			validate: func(t *testing.T, repo *BoilerplateDatabaseRepository[testDBEntity]) {
				assert.NotNil(t, repo, "Expected repository to be created, got nil")
				assert.Nil(t, repo.db, "Expected db to be nil")
				assert.Nil(t, repo.tx, "Expected tx to be nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masterDB, slaveDB := tt.setupDB()

			var boilerplateDB *boilerplate_database.BoilerplateDatabase
			if masterDB != nil && slaveDB != nil {
				boilerplateDB = &boilerplate_database.BoilerplateDatabase{
					Master:                        masterDB,
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         slaveDB,
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}
				defer masterDB.Close()
				defer slaveDB.Close()
			}

			repo := NewBoilerplateDatabaseRepository[testDBEntity](boilerplateDB)

			if tt.validate != nil {
				tt.validate(t, repo)
			}
		})
	}
}

func Test_BoilerplateDatabaseRepository_getTableNameAndFields(t *testing.T) {
	tests := []struct {
		name           string
		setupRepo      func() (string, []string)
		expectedTable  string
		expectedFields []string
		validate       func(t *testing.T, tableName string, fields []string)
	}{
		{
			name: "get table name and fields from entity with valid tags",
			setupRepo: func() (string, []string) {
				mockMasterDB, _, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)
				return repo.getTableNameAndFields()
			},
			expectedTable:  "test_table",
			expectedFields: []string{"name", "email", "created_at"},
			validate: func(t *testing.T, tableName string, fields []string) {
				assert.Equal(t, "test_table", tableName, fmt.Sprintf("expected table name 'test_table', got '%s'", tableName))
				assert.Len(t, fields, 3, fmt.Sprintf("expected 3 fields (ID has table tag, should be skipped), got %d", len(fields)))
				expectedFields := []string{"name", "email", "created_at"}
				assert.Equal(t, expectedFields, fields, fmt.Sprintf("expected fields %v, got %v", expectedFields, fields))
			},
		},
		{
			name: "get table name and fields with dash tag (should be ignored)",
			setupRepo: func() (string, []string) {
				mockMasterDB, _, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithDashTags](boilerplateDB)
				return repo.getTableNameAndFields()
			},
			expectedTable:  "users",
			expectedFields: []string{"name"},
			validate: func(t *testing.T, tableName string, fields []string) {
				assert.Equal(t, "users", tableName, fmt.Sprintf("expected table name 'users', got '%s'", tableName))
				assert.Len(t, fields, 1, fmt.Sprintf("expected 1 field (ID has table tag, dash tags should be ignored), got %d", len(fields)))
			},
		},
		{
			name: "get table name and fields with empty tags (should be ignored)",
			setupRepo: func() (string, []string) {
				mockMasterDB, _, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithEmptyTags](boilerplateDB)
				return repo.getTableNameAndFields()
			},
			expectedTable:  "products",
			expectedFields: []string{"name"},
			validate: func(t *testing.T, tableName string, fields []string) {
				assert.Equal(t, "products", tableName, fmt.Sprintf("expected table name 'products', got '%s'", tableName))
				assert.Len(t, fields, 1, fmt.Sprintf("expected 1 field (ID has table tag, empty tags should be ignored), got %d", len(fields)))
			},
		},
		{
			name: "get table name and fields with single field having table tag",
			setupRepo: func() (string, []string) {
				mockMasterDB, _, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithTableTag](boilerplateDB)
				return repo.getTableNameAndFields()
			},
			expectedTable:  "orders",
			expectedFields: []string{"code"},
			validate: func(t *testing.T, tableName string, fields []string) {
				assert.Equal(t, "orders", tableName, fmt.Sprintf("expected table name 'orders', got '%s'", tableName))
				assert.Len(t, fields, 1, fmt.Sprintf("expected 1 field (ID has table tag, should be skipped), got %d", len(fields)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName, fields := tt.setupRepo()

			if tt.validate != nil {
				tt.validate(t, tableName, fields)
			}
		})
	}
}

func Test_BoilerplateDatabaseRepository_getEntityMeta(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name                 string
		setupRepo            func() entityMeta
		expectedTable        string
		expectedMapLen       int
		expectedMapContains  map[string]interface{}
		expectedMapNotHasKey []string
		validate             func(t *testing.T, tableName string, mapFieldWithValue map[string]interface{})
	}{
		{
			name: "get table name and map field with value from entity with valid tags and values",
			setupRepo: func() entityMeta {
				mockMasterDB, _, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValues](boilerplateDB)
				entity := &testEntityWithValues{
					ID:        1,
					Name:      "John Doe",
					Email:     "john@example.com",
					Age:       30,
					CreatedAt: now,
				}
				return repo.getEntityMeta(entity)
			},
			expectedTable:  "test_values_table",
			expectedMapLen: 4,
			expectedMapContains: map[string]interface{}{
				"name":       "John Doe",
				"email":      "john@example.com",
				"age":        30,
				"created_at": now,
			},
			validate: func(t *testing.T, tableName string, mapFieldWithValue map[string]interface{}) {
				assert.Equal(t, "test_values_table", tableName, fmt.Sprintf("expected table name 'test_values_table', got '%s'", tableName))
				assert.Len(t, mapFieldWithValue, 4, fmt.Sprintf("expected 4 fields in map (ID field with table tag is skipped), got %d", len(mapFieldWithValue)))
				assert.NotContains(t, mapFieldWithValue, "id", "expected 'id' field to be skipped (has table tag)")
				assert.Equal(t, "John Doe", mapFieldWithValue["name"], fmt.Sprintf("expected name = 'John Doe', got %v", mapFieldWithValue["name"]))
				assert.Equal(t, "john@example.com", mapFieldWithValue["email"], fmt.Sprintf("expected email = 'john@example.com', got %v", mapFieldWithValue["email"]))
				assert.Equal(t, 30, mapFieldWithValue["age"], fmt.Sprintf("expected age = 30, got %v", mapFieldWithValue["age"]))
				assert.Equal(t, now, mapFieldWithValue["created_at"], fmt.Sprintf("expected created_at = %v, got %v", now, mapFieldWithValue["created_at"]))
			},
		},
		{
			name: "get table name and map with dash tag (should be ignored)",
			setupRepo: func() entityMeta {
				mockMasterDB, _, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithDashTagValues](boilerplateDB)
				entity := &testEntityWithDashTagValues{
					ID:       2,
					Name:     "Jane Doe",
					Password: "secret123",
					Internal: "internal_data",
				}
				return repo.getEntityMeta(entity)
			},
			expectedTable:        "users_values",
			expectedMapLen:       1,
			expectedMapNotHasKey: []string{"Password", "Internal"},
			validate: func(t *testing.T, tableName string, mapFieldWithValue map[string]interface{}) {
				assert.Equal(t, "users_values", tableName, fmt.Sprintf("expected table name 'users_values', got '%s'", tableName))
				assert.Len(t, mapFieldWithValue, 1, fmt.Sprintf("expected 1 field in map (ID has table tag, dash tags should be ignored), got %d", len(mapFieldWithValue)))
				assert.NotContains(t, mapFieldWithValue, "id", "expected 'id' field to be skipped (has table tag)")
				assert.Equal(t, "Jane Doe", mapFieldWithValue["name"], fmt.Sprintf("expected name = 'Jane Doe', got %v", mapFieldWithValue["name"]))
				assert.NotContains(t, mapFieldWithValue, "password", "expected 'password' field to be ignored (dash tag)")
				assert.NotContains(t, mapFieldWithValue, "internal", "expected 'internal' field to be ignored (dash table tag)")
			},
		},
		{
			name: "get table name and map with empty tags (should be ignored)",
			setupRepo: func() entityMeta {
				mockMasterDB, _, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithEmptyTagValues](boilerplateDB)
				entity := &testEntityWithEmptyTagValues{
					ID:       3,
					Name:     "Product A",
					Price:    99.99,
					Ignored1: "should_be_ignored",
					Ignored2: "also_ignored",
				}
				return repo.getEntityMeta(entity)
			},
			expectedTable:        "products_values",
			expectedMapLen:       2,
			expectedMapNotHasKey: []string{"Ignored1", "Ignored2"},
			validate: func(t *testing.T, tableName string, mapFieldWithValue map[string]interface{}) {
				assert.Equal(t, "products_values", tableName, fmt.Sprintf("expected table name 'products_values', got '%s'", tableName))
				assert.Len(t, mapFieldWithValue, 2, fmt.Sprintf("expected 2 fields in map (ID has table tag, empty tags should be ignored), got %d", len(mapFieldWithValue)))
				assert.NotContains(t, mapFieldWithValue, "id", "expected 'id' field to be skipped (has table tag)")
				assert.Equal(t, "Product A", mapFieldWithValue["name"], fmt.Sprintf("expected name = 'Product A', got %v", mapFieldWithValue["name"]))
				assert.Equal(t, 99.99, mapFieldWithValue["price"], fmt.Sprintf("expected price = 99.99, got %v", mapFieldWithValue["price"]))
				assert.NotContains(t, mapFieldWithValue, "ignored1", "expected 'ignored1' field to be ignored (empty tag)")
				assert.NotContains(t, mapFieldWithValue, "ignored2", "expected 'ignored2' field to be ignored (no tag)")
			},
		},
		{
			name: "get table name and map with zero values",
			setupRepo: func() entityMeta {
				mockMasterDB, _, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithNilPointerValues](boilerplateDB)
				entity := &testEntityWithNilPointerValues{
					ID:   0,
					Name: "",
				}
				return repo.getEntityMeta(entity)
			},
			expectedTable:  "nullable_table",
			expectedMapLen: 1,
			validate: func(t *testing.T, tableName string, mapFieldWithValue map[string]interface{}) {
				assert.Equal(t, "nullable_table", tableName, fmt.Sprintf("expected table name 'nullable_table', got '%s'", tableName))
				assert.Len(t, mapFieldWithValue, 1, fmt.Sprintf("expected 1 field in map (ID has table tag), got %d", len(mapFieldWithValue)))
				assert.NotContains(t, mapFieldWithValue, "id", "expected 'id' field to be skipped (has table tag)")
				assert.Equal(t, "", mapFieldWithValue["name"], fmt.Sprintf("expected name = '' (empty string), got %v", mapFieldWithValue["name"]))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := tt.setupRepo()
			if tt.validate != nil {
				tt.validate(t, meta.TableName, meta.FieldValueMap)
			}
		})
	}
}

func Test_BoilerplateDatabaseRepository_exec(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock)
		ctx         context.Context
		query       string
		args        []interface{}
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "exec successfully without transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)
				return repo, mockMaster
			},
			ctx:         context.Background(),
			query:       "INSERT INTO exec_test_table (name) VALUES (?)",
			args:        []interface{}{"test_name"},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error, got %v", err))
			},
		},
		{
			name: "exec with prepare error without transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)
				return repo, mockMaster
			},
			ctx:         context.Background(),
			query:       "INSERT INTO exec_test_table (name) VALUES (?)",
			args:        []interface{}{"test_name"},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Equal(t, "error", err.Error(), fmt.Sprintf("expected error to contain 'error', got %v", err.Error()))
			},
		},
		{
			name: "exec with statement exec error without transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)
				return repo, mockMaster
			},
			ctx:         context.Background(),
			query:       "INSERT INTO exec_test_table (name) VALUES (?)",
			args:        []interface{}{"test_name"},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Equal(t, "error", err.Error(), fmt.Sprintf("expected error to contain 'error', got %v", err.Error()))
			},
		},
		{
			name: "exec successfully with transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)

				mockMaster.ExpectBegin()
				tx, _ := boilerplateDB.Master.BeginTxx(context.Background(), nil)
				repo.tx = newBoilerplateDatabaseTransaction(tx)

				return repo, mockMaster
			},
			ctx:         context.Background(),
			query:       "INSERT INTO exec_test_table (name) VALUES (?)",
			args:        []interface{}{"test_name"},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error, got %v", err))
			},
		},
		{
			name: "exec with transaction prepare error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)

				mockMaster.ExpectBegin()
				tx, _ := boilerplateDB.Master.BeginTxx(context.Background(), nil)
				repo.tx = newBoilerplateDatabaseTransaction(tx)

				return repo, mockMaster
			},
			ctx:         context.Background(),
			query:       "INSERT INTO exec_test_table (name) VALUES (?)",
			args:        []interface{}{"test_name"},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Equal(t, "error", err.Error(), fmt.Sprintf("expected error to contain 'error', got %v", err.Error()))
			},
		},
		{
			name: "exec with slow query warning",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 5 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)
				return repo, mockMaster
			},
			ctx:         context.Background(),
			query:       "INSERT INTO exec_test_table (name) VALUES (?)",
			args:        []interface{}{"test_name"},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error (slow query should only warn), got %v", err))
			},
		},
		{
			name: "exec with close statement error after success",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)
				return repo, mockMaster
			},
			ctx:         context.Background(),
			query:       "INSERT INTO exec_test_table (name) VALUES (?)",
			args:        []interface{}{"test_name"},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error (close error should only be logged), got %v", err))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := tt.setupRepo()

			switch tt.name {
			case "exec successfully without transaction":
				mock.ExpectPrepare("INSERT INTO exec_test_table").WillBeClosed()
				mock.ExpectExec("INSERT INTO exec_test_table").
					WithArgs("test_name").
					WillReturnResult(sqlmock.NewResult(1, 1))
			case "exec with prepare error without transaction":
				mock.ExpectPrepare("INSERT INTO exec_test_table").
					WillReturnError(errors.New("prepare error"))
			case "exec with statement exec error without transaction":
				mock.ExpectPrepare("INSERT INTO exec_test_table").WillBeClosed()
				mock.ExpectExec("INSERT INTO exec_test_table").
					WithArgs("test_name").
					WillReturnError(errors.New("exec error"))
			case "exec successfully with transaction":
				mock.ExpectPrepare("INSERT INTO exec_test_table").WillBeClosed()
				mock.ExpectExec("INSERT INTO exec_test_table").
					WithArgs("test_name").
					WillReturnResult(sqlmock.NewResult(1, 1))
			case "exec with transaction prepare error":
				mock.ExpectPrepare("INSERT INTO exec_test_table").
					WillReturnError(errors.New("tx prepare error"))
			case "exec with slow query warning":
				mock.ExpectPrepare("INSERT INTO exec_test_table").WillBeClosed()
				mock.ExpectExec("INSERT INTO exec_test_table").
					WithArgs("test_name").
					WillDelayFor(10 * time.Millisecond).
					WillReturnResult(sqlmock.NewResult(1, 1))
			case "exec with close statement error after success":
				mockPrepare := mock.ExpectPrepare("INSERT INTO exec_test_table")
				mockPrepare.WillReturnCloseError(errors.New("close stmt error"))
				mock.ExpectExec("INSERT INTO exec_test_table").
					WithArgs("test_name").
					WillReturnResult(sqlmock.NewResult(1, 1))
			}

			err := repo.exec(tt.ctx, tt.query, tt.args...)

			if tt.expectError {
				assert.NotNil(t, err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.NoError(t, err, fmt.Sprintf("Expected no error but got: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled expectations: %v", mock.ExpectationsWereMet()))
		})
	}
}

func Test_BoilerplateDatabaseRepository_BeginTransaction(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock)
		ctx         context.Context
		expectError bool
		validate    func(t *testing.T, tx IBoilerplateDatabaseTransaction, err error)
	}{
		{
			name: "begin transaction successfully",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)
				return repo, mockMaster
			},
			ctx:         context.Background(),
			expectError: false,
			validate: func(t *testing.T, tx IBoilerplateDatabaseTransaction, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error, got %v", err))
				assert.NotNil(t, tx, "expected transaction to be created, got nil")
			},
		},
		{
			name: "begin transaction with error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)
				return repo, mockMaster
			},
			ctx:         context.Background(),
			expectError: true,
			validate: func(t *testing.T, tx IBoilerplateDatabaseTransaction, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Nil(t, tx, fmt.Sprintf("expected transaction to be nil on error, got %v", tx))
				assert.Equal(t, "error", err.Error(), fmt.Sprintf("expected error message 'error', got %v", err.Error()))
			},
		},
		{
			name: "begin transaction when transaction already exists",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, _, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "sqlmock"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "sqlmock"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)

				mockMaster.ExpectBegin()
				existingTx, _ := boilerplateDB.Master.BeginTxx(context.Background(), nil)
				repo.tx = newBoilerplateDatabaseTransaction(existingTx)

				return repo, mockMaster
			},
			ctx:         context.Background(),
			expectError: false,
			validate: func(t *testing.T, tx IBoilerplateDatabaseTransaction, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error, got %v", err))
				assert.NotNil(t, tx, "expected existing transaction to be returned, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := tt.setupRepo()

			switch tt.name {
			case "begin transaction successfully":
				mock.ExpectBegin()
			case "begin transaction with error":
				mock.ExpectBegin().WillReturnError(errors.New("begin transaction error"))
			}

			tx, err := repo.BeginTransaction(tt.ctx)

			if tt.expectError {
				assert.NotNil(t, err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.NoError(t, err, fmt.Sprintf("Expected no error but got: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, tx, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled expectations: %v", mock.ExpectationsWereMet()))
		})
	}
}

func Test_BoilerplateDatabaseRepository_Count(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock)
		ctx         context.Context
		filter      *goqube.Filter
		useMaster   bool
		expectError bool
		validate    func(t *testing.T, count uint64, err error)
	}{
		{
			name: "count successfully using slave",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)
				return repo, mockMaster, mockSlave
			},
			ctx:         context.Background(),
			filter:      nil,
			useMaster:   false,
			expectError: false,
			validate: func(t *testing.T, count uint64, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error, got %v", err))
				assert.Equal(t, uint64(5), count, fmt.Sprintf("expected count to be 5, got %d", count))
			},
		},
		{
			name: "count successfully using master without transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)
				return repo, mockMaster, mockSlave
			},
			ctx:         context.Background(),
			filter:      nil,
			useMaster:   true,
			expectError: false,
			validate: func(t *testing.T, count uint64, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error, got %v", err))
				assert.Equal(t, uint64(10), count, fmt.Sprintf("expected count to be 10, got %d", count))
			},
		},
		{
			name: "count successfully using master with transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				mockMaster.ExpectBegin()
				tx, _ := boilerplateDB.Master.BeginTxx(context.Background(), nil)
				repo.tx = newBoilerplateDatabaseTransaction(tx)

				return repo, mockMaster, mockSlave
			},
			ctx:         context.Background(),
			filter:      nil,
			useMaster:   true,
			expectError: false,
			validate: func(t *testing.T, count uint64, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error, got %v", err))
				assert.Equal(t, uint64(15), count, fmt.Sprintf("expected count to be 15, got %d", count))
			},
		},
		{
			name: "count with prepare error on slave",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)
				return repo, mockMaster, mockSlave
			},
			ctx:         context.Background(),
			filter:      nil,
			useMaster:   false,
			expectError: true,
			validate: func(t *testing.T, count uint64, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Equal(t, uint64(0), count, fmt.Sprintf("expected count to be 0 on error, got %d", count))
			},
		},
		{
			name: "count with get error (no rows)",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)
				return repo, mockMaster, mockSlave
			},
			ctx:         context.Background(),
			filter:      nil,
			useMaster:   false,
			expectError: true,
			validate: func(t *testing.T, count uint64, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Equal(t, uint64(0), count, fmt.Sprintf("expected count to be 0 on error, got %d", count))
				assert.Equal(t, "entity not found", err.Error(), fmt.Sprintf("expected 'entity not found' error, got %v", err.Error()))
			},
		},
		{
			name: "count with get error (other database error)",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)
				return repo, mockMaster, mockSlave
			},
			ctx:         context.Background(),
			filter:      nil,
			useMaster:   false,
			expectError: true,
			validate: func(t *testing.T, count uint64, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Equal(t, uint64(0), count, fmt.Sprintf("expected count to be 0 on error, got %d", count))
			},
		},
		{
			name: "count with slow query warning",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  5 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)
				return repo, mockMaster, mockSlave
			},
			ctx:         context.Background(),
			filter:      nil,
			useMaster:   false,
			expectError: false,
			validate: func(t *testing.T, count uint64, err error) {
				assert.NoError(t, err, fmt.Sprintf("expected no error (slow query should only warn), got %v", err))
				assert.Equal(t, uint64(20), count, fmt.Sprintf("expected count to be 20, got %d", count))
			},
		},
		{
			name: "count with BuildSelectQuery error (invalid dialect)",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "invalid_dialect"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "invalid_dialect"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)
				return repo, mockMaster, mockSlave
			},
			ctx:         context.Background(),
			filter:      nil,
			useMaster:   false,
			expectError: true,
			validate: func(t *testing.T, count uint64, err error) {
				assert.NotNil(t, err, "expected error from BuildSelectQuery, got nil")
				assert.Equal(t, uint64(0), count, fmt.Sprintf("expected count to be 0 on error, got %d", count))
			},
		},
		{
			name: "count with prepare error on master with transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				mockMaster.ExpectBegin()
				tx, _ := boilerplateDB.Master.BeginTxx(context.Background(), nil)
				repo.tx = newBoilerplateDatabaseTransaction(tx)

				return repo, mockMaster, mockSlave
			},
			ctx:         context.Background(),
			filter:      nil,
			useMaster:   true,
			expectError: true,
			validate: func(t *testing.T, count uint64, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Equal(t, uint64(0), count, fmt.Sprintf("expected count to be 0 on error, got %d", count))
			},
		},
		{
			name: "count with prepare error on master without transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:         context.Background(),
			filter:      nil,
			useMaster:   true,
			expectError: true,
			validate: func(t *testing.T, count uint64, err error) {
				assert.NotNil(t, err, "expected error, got nil")
				assert.Equal(t, uint64(0), count, fmt.Sprintf("expected count to be 0 on error, got %d", count))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mockMaster, mockSlave := tt.setupRepo()

			switch tt.name {
			case "count successfully using slave":
				mockSlave.ExpectPrepare("SELECT (.+) FROM test_table").WillBeClosed()
				rows := sqlmock.NewRows([]string{"COUNT(-1)"}).AddRow(uint64(5))
				mockSlave.ExpectQuery("SELECT (.+) FROM test_table").WillReturnRows(rows)
			case "count successfully using master without transaction":
				mockMaster.ExpectPrepare("SELECT (.+) FROM test_table").WillBeClosed()
				rows := sqlmock.NewRows([]string{"COUNT(-1)"}).AddRow(uint64(10))
				mockMaster.ExpectQuery("SELECT (.+) FROM test_table").WillReturnRows(rows)
			case "count successfully using master with transaction":
				mockMaster.ExpectPrepare("SELECT (.+) FROM test_table").WillBeClosed()
				rows := sqlmock.NewRows([]string{"COUNT(-1)"}).AddRow(uint64(15))
				mockMaster.ExpectQuery("SELECT (.+) FROM test_table").WillReturnRows(rows)
			case "count with prepare error on slave":
				mockSlave.ExpectPrepare("SELECT (.+) FROM test_table").
					WillReturnError(errors.New("prepare error"))
			case "count with get error (no rows)":
				mockSlave.ExpectPrepare("SELECT (.+) FROM test_table").WillBeClosed()
				mockSlave.ExpectQuery("SELECT (.+) FROM test_table").
					WillReturnError(sql.ErrNoRows)
			case "count with get error (other database error)":
				mockSlave.ExpectPrepare("SELECT (.+) FROM test_table").WillBeClosed()
				mockSlave.ExpectQuery("SELECT (.+) FROM test_table").
					WillReturnError(errors.New("database connection error"))
			case "count with slow query warning":
				mockSlave.ExpectPrepare("SELECT (.+) FROM test_table").WillBeClosed()
				rows := sqlmock.NewRows([]string{"COUNT(-1)"}).AddRow(uint64(20))
				mockSlave.ExpectQuery("SELECT (.+) FROM test_table").
					WillDelayFor(10 * time.Millisecond).
					WillReturnRows(rows)
			case "count with prepare error on master with transaction":
				mockMaster.ExpectPrepare("SELECT (.+) FROM test_table").
					WillReturnError(errors.New("tx prepare error"))
			case "count with prepare error on master without transaction":
				mockMaster.ExpectPrepare("SELECT (.+) FROM test_table").
					WillReturnError(errors.New("master prepare error"))
			}

			count, err := repo.Count(tt.ctx, tt.filter, tt.useMaster)

			if tt.expectError {
				assert.NotNil(t, err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.NoError(t, err, fmt.Sprintf("Expected no error but got: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, count, err)
			}

			if tt.useMaster || tt.name == "count successfully using master with transaction" {
				assert.NoError(t, mockMaster.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled master expectations: %v", mockMaster.ExpectationsWereMet()))
			} else {
				assert.NoError(t, mockSlave.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled slave expectations: %v", mockSlave.ExpectationsWereMet()))
			}
		})
	}
}

func Test_BoilerplateDatabaseRepository_Create(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock)
		ctx         context.Context
		entity      *testEntityForExec
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "create successfully",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
				}
				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)
				return repo, mock
			},
			ctx: context.Background(),
			entity: &testEntityForExec{
				ID:   123,
				Name: "Test Entity",
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "expected no error, got %v", err)
			},
		},
		{
			name: "create with nil entity",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
				}
				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)
				return repo, mock
			},
			ctx:         context.Background(),
			entity:      nil,
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "expected no error for nil entity, got %v", err)
			},
		},
		{
			name: "create with BuildInsertQuery error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockDB, "invalid_dialect"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
				}
				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)
				return repo, mock
			},
			ctx: context.Background(),
			entity: &testEntityForExec{
				ID:   456,
				Name: "Test Entity",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error from BuildInsertQuery, got nil")
			},
		},
		{
			name: "create with exec error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
				}
				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)
				return repo, mock
			},
			ctx: context.Background(),
			entity: &testEntityForExec{
				ID:   789,
				Name: "Test Entity",
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error from exec, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := tt.setupRepo()

			switch tt.name {
			case "create successfully":
				mock.ExpectPrepare("INSERT INTO (.+)").WillBeClosed()
				mock.ExpectExec("INSERT INTO (.+)").
					WillReturnResult(sqlmock.NewResult(1, 1))
			case "create with exec error":
				mock.ExpectPrepare("INSERT INTO (.+)").WillBeClosed()
				mock.ExpectExec("INSERT INTO (.+)").
					WillReturnError(errors.New("exec error"))
			}

			err := repo.Create(tt.ctx, tt.entity)

			if tt.expectError {
				assert.NotNil(t, err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.NoError(t, err, fmt.Sprintf("Expected no error but got: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}

			if tt.entity != nil && tt.name != "create with BuildInsertQuery error" {
				assert.NoError(t, mock.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled expectations: %v", mock.ExpectationsWereMet()))
			}
		})
	}
}

func Test_BoilerplateDatabaseRepository_Delete(t *testing.T) {
	tests := []struct {
		name        string
		filter      *goqube.Filter
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "delete successfully",
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "id"},
						Operator: goqube.OperatorEqual,
						Value:    goqube.FilterValue{Value: 1},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("DELETE FROM exec_test_table WHERE.*").
					WillBeClosed().
					ExpectExec().
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
		},
		{
			name: "delete with BuildDeleteQuery error",
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "id"},
						Operator: goqube.OperatorEqual,
						Value:    goqube.FilterValue{Value: 1},
					},
				},
			},
			setupMock:   func(mock sqlmock.Sqlmock) {},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error from BuildDeleteQuery, got nil")
			},
		},
		{
			name: "delete with exec error",
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "id"},
						Operator: goqube.OperatorEqual,
						Value:    goqube.FilterValue{Value: 1},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("DELETE FROM exec_test_table WHERE.*").
					WillBeClosed().
					ExpectExec().
					WithArgs(1).
					WillReturnError(errors.New("exec error"))
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error from exec, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				mock   sqlmock.Sqlmock
				db     *sql.DB
				sqlxDB *sqlx.DB
				err    error
			)

			db, mock, err = sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock: %v", err)
			}
			defer db.Close()

			sqlxDB = sqlx.NewDb(db, "mysql")

			var boilerplateDB *boilerplate_database.BoilerplateDatabase
			switch tt.name {
			case "delete with BuildDeleteQuery error":

				invalidDB := sqlx.NewDb(db, "invalid_dialect")
				boilerplateDB = &boilerplate_database.BoilerplateDatabase{
					Master:                        invalidDB,
					Slave:                         sqlxDB,
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}
			default:
				boilerplateDB = &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlxDB,
					Slave:                         sqlxDB,
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}
			}
			tt.setupMock(mock)

			repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)

			ctx := context.Background()
			err = repo.Delete(ctx, tt.filter)

			if tt.expectError {
				assert.NotNil(t, err, "expected error, got nil")
			}
			if !tt.expectError {
				assert.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}

			if tt.name != "delete with BuildDeleteQuery error" {
				assert.NoError(t, mock.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled expectations: %v", mock.ExpectationsWereMet()))
			}
		})
	}
}

func Test_BoilerplateDatabaseRepository_FindAll(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock)
		ctx         context.Context
		filter      *goqube.Filter
		sorts       []goqube.Sort
		take        uint64
		skip        uint64
		useMaster   bool
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
		validate    func(t *testing.T, entities []testEntityWithValidTags, err error)
	}{
		{
			name: "find all successfully using slave",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			take:      10,
			skip:      0,
			useMaster: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Entity 1").
					AddRow(2, "Entity 2")

				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillBeClosed().
					ExpectQuery().
					WillReturnRows(rows)
			},
			expectError: false,
			validate: func(t *testing.T, entities []testEntityWithValidTags, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
				assert.Len(t, entities, 2, "expected 2 entities, got %d", len(entities))
			},
		},
		{
			name: "find all successfully using master without transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			take:      10,
			skip:      0,
			useMaster: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Entity 1")

				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillBeClosed().
					ExpectQuery().
					WillReturnRows(rows)
			},
			expectError: false,
			validate: func(t *testing.T, entities []testEntityWithValidTags, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
				assert.Len(t, entities, 1, "expected 1 entity, got %d", len(entities))
			},
		},
		{
			name: "find all successfully using master with transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				mockMaster.ExpectBegin()
				tx, _ := boilerplateDB.Master.BeginTxx(context.Background(), nil)
				repo.tx = newBoilerplateDatabaseTransaction(tx)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			take:      10,
			skip:      0,
			useMaster: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Entity 1")

				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillBeClosed().
					ExpectQuery().
					WillReturnRows(rows)
			},
			expectError: false,
			validate: func(t *testing.T, entities []testEntityWithValidTags, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
				assert.Len(t, entities, 1, "expected 1 entity, got %d", len(entities))
			},
		},
		{
			name: "find all with BuildSelectQuery error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "invalid_dialect"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:         context.Background(),
			filter:      nil,
			sorts:       nil,
			take:        10,
			skip:        0,
			useMaster:   false,
			setupMock:   func(mock sqlmock.Sqlmock) {},
			expectError: true,
			validate: func(t *testing.T, entities []testEntityWithValidTags, err error) {
				assert.NotNil(t, err, "expected error from BuildSelectQuery, got nil")
			},
		},
		{
			name: "find all with prepare error on slave",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			take:      10,
			skip:      0,
			useMaster: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillReturnError(errors.New("prepare error"))
			},
			expectError: true,
			validate: func(t *testing.T, entities []testEntityWithValidTags, err error) {
				assert.NotNil(t, err, "expected error from prepare, got nil")
			},
		},
		{
			name: "find all with prepare error on master with transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				mockMaster.ExpectBegin()
				tx, _ := boilerplateDB.Master.BeginTxx(context.Background(), nil)
				repo.tx = newBoilerplateDatabaseTransaction(tx)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			take:      10,
			skip:      0,
			useMaster: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillReturnError(errors.New("prepare error"))
			},
			expectError: true,
			validate: func(t *testing.T, entities []testEntityWithValidTags, err error) {
				assert.NotNil(t, err, "expected error from prepare, got nil")
			},
		},
		{
			name: "find all with prepare error on master without transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			take:      10,
			skip:      0,
			useMaster: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillReturnError(errors.New("prepare error"))
			},
			expectError: true,
			validate: func(t *testing.T, entities []testEntityWithValidTags, err error) {
				assert.NotNil(t, err, "expected error from prepare, got nil")
			},
		},
		{
			name: "find all with select error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			take:      10,
			skip:      0,
			useMaster: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillBeClosed().
					ExpectQuery().
					WillReturnError(errors.New("select error"))
			},
			expectError: true,
			validate: func(t *testing.T, entities []testEntityWithValidTags, err error) {
				assert.NotNil(t, err, "expected error from select, got nil")
			},
		},
		{
			name: "find all with slow query warning",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  1 * time.Nanosecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			take:      10,
			skip:      0,
			useMaster: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "Entity 1")

				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillBeClosed().
					ExpectQuery().
					WillDelayFor(10 * time.Millisecond).
					WillReturnRows(rows)
			},
			expectError: false,
			validate: func(t *testing.T, entities []testEntityWithValidTags, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
				assert.Len(t, entities, 1, "expected 1 entity, got %d", len(entities))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repo, mockMaster, mockSlave := tt.setupRepo()

			if tt.useMaster {
				tt.setupMock(mockMaster)
			} else {
				tt.setupMock(mockSlave)
			}

			entities, err := repo.FindAll(tt.ctx, tt.filter, tt.sorts, tt.take, tt.skip, tt.useMaster)

			if tt.expectError {
				assert.NotNil(t, err, "expected error, got nil")
			}
			if !tt.expectError {
				assert.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, entities, err)
			}

			if tt.useMaster {
				assert.NoError(t, mockMaster.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled master expectations: %v", mockMaster.ExpectationsWereMet()))
			} else {
				assert.NoError(t, mockSlave.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled slave expectations: %v", mockSlave.ExpectationsWereMet()))
			}
		})
	}
}

func Test_BoilerplateDatabaseRepository_FindOne(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock)
		ctx         context.Context
		filter      *goqube.Filter
		sorts       []goqube.Sort
		useMaster   bool
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
		validate    func(t *testing.T, entity *testEntityWithValidTags, err error)
	}{
		{
			name: "find one successfully using slave",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			useMaster: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
					AddRow(1, "Entity 1", "test@example.com", time.Now())

				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillBeClosed().
					ExpectQuery().
					WillReturnRows(rows)
			},
			expectError: false,
			validate: func(t *testing.T, entity *testEntityWithValidTags, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
				assert.NotNil(t, entity, "expected entity, got nil")
			},
		},
		{
			name: "find one successfully using master without transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			useMaster: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
					AddRow(1, "Entity 1", "test@example.com", time.Now())

				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillBeClosed().
					ExpectQuery().
					WillReturnRows(rows)
			},
			expectError: false,
			validate: func(t *testing.T, entity *testEntityWithValidTags, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
				assert.NotNil(t, entity, "expected entity, got nil")
			},
		},
		{
			name: "find one successfully using master with transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				mockMaster.ExpectBegin()
				tx, _ := boilerplateDB.Master.BeginTxx(context.Background(), nil)
				repo.tx = newBoilerplateDatabaseTransaction(tx)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			useMaster: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
					AddRow(1, "Entity 1", "test@example.com", time.Now())

				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillBeClosed().
					ExpectQuery().
					WillReturnRows(rows)
			},
			expectError: false,
			validate: func(t *testing.T, entity *testEntityWithValidTags, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
				assert.NotNil(t, entity, "expected entity, got nil")
			},
		},
		{
			name: "find one with BuildSelectQuery error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "invalid_dialect"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:         context.Background(),
			filter:      nil,
			sorts:       nil,
			useMaster:   false,
			setupMock:   func(mock sqlmock.Sqlmock) {},
			expectError: true,
			validate: func(t *testing.T, entity *testEntityWithValidTags, err error) {
				assert.NotNil(t, err, "expected error from BuildSelectQuery, got nil")
				assert.Nil(t, entity, "expected nil entity on error")
			},
		},
		{
			name: "find one with prepare error on slave",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			useMaster: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillReturnError(errors.New("prepare error"))
			},
			expectError: true,
			validate: func(t *testing.T, entity *testEntityWithValidTags, err error) {
				assert.NotNil(t, err, "expected error from prepare, got nil")
			},
		},
		{
			name: "find one with prepare error on master with transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				mockMaster.ExpectBegin()
				tx, _ := boilerplateDB.Master.BeginTxx(context.Background(), nil)
				repo.tx = newBoilerplateDatabaseTransaction(tx)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			useMaster: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillReturnError(errors.New("prepare error"))
			},
			expectError: true,
			validate: func(t *testing.T, entity *testEntityWithValidTags, err error) {
				assert.NotNil(t, err, "expected error from prepare, got nil")
			},
		},
		{
			name: "find one with prepare error on master without transaction",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			useMaster: true,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillReturnError(errors.New("prepare error"))
			},
			expectError: true,
			validate: func(t *testing.T, entity *testEntityWithValidTags, err error) {
				assert.NotNil(t, err, "expected error from prepare, got nil")
			},
		},
		{
			name: "find one with no rows error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			useMaster: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillBeClosed().
					ExpectQuery().
					WillReturnError(sql.ErrNoRows)
			},
			expectError: true,
			validate: func(t *testing.T, entity *testEntityWithValidTags, err error) {
				assert.NotNil(t, err, "expected entity not found error, got nil")
				assert.Nil(t, entity, "expected nil entity when not found")
			},
		},
		{
			name: "find one with other database error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			useMaster: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillBeClosed().
					ExpectQuery().
					WillReturnError(errors.New("database connection error"))
			},
			expectError: true,
			validate: func(t *testing.T, entity *testEntityWithValidTags, err error) {
				assert.NotNil(t, err, "expected database error, got nil")
			},
		},
		{
			name: "find one with slow query warning",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityWithValidTags], sqlmock.Sqlmock, sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()
				mockSlaveDB, mockSlave, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
					Slave:                         sqlx.NewDb(mockSlaveDB, "mysql"),
					SlaveMaxQueryDurationWarning:  1 * time.Nanosecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityWithValidTags](boilerplateDB)

				return repo, mockMaster, mockSlave
			},
			ctx:       context.Background(),
			filter:    nil,
			sorts:     nil,
			useMaster: false,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
					AddRow(1, "Entity 1", "test@example.com", time.Now())

				mock.ExpectPrepare("SELECT (.+) FROM test_table").
					WillBeClosed().
					ExpectQuery().
					WillDelayFor(10 * time.Millisecond).
					WillReturnRows(rows)
			},
			expectError: false,
			validate: func(t *testing.T, entity *testEntityWithValidTags, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
				assert.NotNil(t, entity, "expected entity, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repo, mockMaster, mockSlave := tt.setupRepo()

			if tt.useMaster {
				tt.setupMock(mockMaster)
			} else {
				tt.setupMock(mockSlave)
			}

			entity, err := repo.FindOne(tt.ctx, tt.filter, tt.sorts, tt.useMaster)

			if tt.expectError {
				assert.NotNil(t, err, "expected error, got nil")
			}
			if !tt.expectError {
				assert.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, entity, err)
			}

			if tt.useMaster {
				assert.NoError(t, mockMaster.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled master expectations: %v", mockMaster.ExpectationsWereMet()))
			} else {
				assert.NoError(t, mockSlave.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled slave expectations: %v", mockSlave.ExpectationsWereMet()))
			}
		})
	}
}

func Test_BoilerplateDatabaseRepository_Update(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock)
		ctx         context.Context
		entity      *testEntityForExec
		filter      *goqube.Filter
		setupMock   func(mock sqlmock.Sqlmock)
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "update successfully",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)

				return repo, mockMaster
			},
			ctx: context.Background(),
			entity: &testEntityForExec{
				ID:   1,
				Name: "Updated Entity",
			},
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "id"},
						Operator: goqube.OperatorEqual,
						Value:    goqube.FilterValue{Value: 1},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("UPDATE exec_test_table SET (.+) WHERE (.+)").
					WillBeClosed().
					ExpectExec().
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "unexpected error: %v", err)
			},
		},
		{
			name: "update with nil entity",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)

				return repo, mockMaster
			},
			ctx:         context.Background(),
			entity:      nil,
			filter:      nil,
			setupMock:   func(mock sqlmock.Sqlmock) {},
			expectError: false,
			validate: func(t *testing.T, err error) {
				assert.NoError(t, err, "expected nil error for nil entity, got: %v", err)
			},
		},
		{
			name: "update with BuildUpdateQuery error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "invalid_dialect"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)

				return repo, mockMaster
			},
			ctx: context.Background(),
			entity: &testEntityForExec{
				ID:   1,
				Name: "Test Entity",
			},
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "id"},
						Operator: goqube.OperatorEqual,
						Value:    goqube.FilterValue{Value: 1},
					},
				},
			},
			setupMock:   func(mock sqlmock.Sqlmock) {},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error from BuildUpdateQuery, got nil")
			},
		},
		{
			name: "update with exec error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForExec], sqlmock.Sqlmock) {
				mockMasterDB, mockMaster, _ := sqlmock.New()

				boilerplateDB := &boilerplate_database.BoilerplateDatabase{
					Master:                        sqlx.NewDb(mockMasterDB, "mysql"),
					MasterMaxQueryDurationWarning: 100 * time.Millisecond,
				}

				repo := NewBoilerplateDatabaseRepository[testEntityForExec](boilerplateDB)

				return repo, mockMaster
			},
			ctx: context.Background(),
			entity: &testEntityForExec{
				ID:   1,
				Name: "Test Entity",
			},
			filter: &goqube.Filter{
				Logic: goqube.LogicAnd,
				Filters: []goqube.Filter{
					{
						Field:    goqube.Field{Column: "id"},
						Operator: goqube.OperatorEqual,
						Value:    goqube.FilterValue{Value: 1},
					},
				},
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectPrepare("UPDATE exec_test_table SET (.+) WHERE (.+)").
					WillBeClosed().
					ExpectExec().
					WillReturnError(errors.New("exec error"))
			},
			expectError: true,
			validate: func(t *testing.T, err error) {
				assert.NotNil(t, err, "expected error from exec, got nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			repo, mockMaster := tt.setupRepo()

			tt.setupMock(mockMaster)

			err := repo.Update(tt.ctx, tt.entity, tt.filter)

			if tt.expectError {
				assert.NotNil(t, err, "expected error, got nil")
			}
			if !tt.expectError {
				assert.NoError(t, err, fmt.Sprintf("unexpected error: %v", err))
			}

			if tt.validate != nil {
				tt.validate(t, err)
			}

			assert.NoError(t, mockMaster.ExpectationsWereMet(), fmt.Sprintf("Unfulfilled master expectations: %v", mockMaster.ExpectationsWereMet()))
		})
	}
}
func Test_BoilerplateDatabaseRepository_BulkCreate(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() (*BoilerplateDatabaseRepository[testEntityForBulkExec], sqlmock.Sqlmock)
		ctx         context.Context
		entities    []testEntityForBulkExec
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "bulk create successfully with multiple entities",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForBulkExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{Master: sqlx.NewDb(mockDB, "mysql"), MasterMaxQueryDurationWarning: 100 * time.Millisecond}
				repo := NewBoilerplateDatabaseRepository[testEntityForBulkExec](boilerplateDB)
				return repo, mock
			},
			ctx:         context.Background(),
			entities:    []testEntityForBulkExec{{ID: 1, Name: "Entity 1"}, {ID: 2, Name: "Entity 2"}},
			expectError: false,
			validate:    func(t *testing.T, err error) { assert.NoError(t, err) },
		},
		{
			name: "bulk create with empty entities",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForBulkExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{Master: sqlx.NewDb(mockDB, "mysql"), MasterMaxQueryDurationWarning: 100 * time.Millisecond}
				repo := NewBoilerplateDatabaseRepository[testEntityForBulkExec](boilerplateDB)
				return repo, mock
			},
			ctx:         context.Background(),
			entities:    []testEntityForBulkExec{},
			expectError: false,
			validate:    func(t *testing.T, err error) { assert.NoError(t, err) },
		},
		{
			name: "bulk create with BuildInsertQuery error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForBulkExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{Master: sqlx.NewDb(mockDB, "invalid_dialect"), MasterMaxQueryDurationWarning: 100 * time.Millisecond}
				repo := NewBoilerplateDatabaseRepository[testEntityForBulkExec](boilerplateDB)
				return repo, mock
			},
			ctx:         context.Background(),
			entities:    []testEntityForBulkExec{{ID: 1, Name: "Entity 1"}},
			expectError: true,
			validate:    func(t *testing.T, err error) { assert.NotNil(t, err) },
		},
		{
			name: "bulk create with exec error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForBulkExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{Master: sqlx.NewDb(mockDB, "mysql"), MasterMaxQueryDurationWarning: 100 * time.Millisecond}
				repo := NewBoilerplateDatabaseRepository[testEntityForBulkExec](boilerplateDB)
				return repo, mock
			},
			ctx:         context.Background(),
			entities:    []testEntityForBulkExec{{ID: 1, Name: "Entity 1"}},
			expectError: true,
			validate:    func(t *testing.T, err error) { assert.NotNil(t, err) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := tt.setupRepo()
			switch tt.name {
			case "bulk create successfully with multiple entities":
				mock.ExpectPrepare("INSERT INTO (.+)").WillBeClosed()
				mock.ExpectExec("INSERT INTO (.+)").WillReturnResult(sqlmock.NewResult(1, 2))
			case "bulk create with exec error":
				mock.ExpectPrepare("INSERT INTO (.+)").WillBeClosed()
				mock.ExpectExec("INSERT INTO (.+)").WillReturnError(errors.New("exec error"))
			}
			err := repo.BulkCreate(tt.ctx, tt.entities)
			if tt.expectError {
				assert.NotNil(t, err)
			}
			if !tt.expectError {
				assert.NoError(t, err)
			}
			if tt.validate != nil {
				tt.validate(t, err)
			}
			if len(tt.entities) > 0 && tt.name != "bulk create with BuildInsertQuery error" && tt.name != "bulk create with exec error" {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
func Test_BoilerplateDatabaseRepository_BulkUpdate(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func() (*BoilerplateDatabaseRepository[testEntityForBulkExec], sqlmock.Sqlmock)
		ctx         context.Context
		entities    []testEntityForBulkExec
		expectError bool
		validate    func(t *testing.T, err error)
	}{
		{
			name: "bulk update successfully with multiple entities",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForBulkExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{Master: sqlx.NewDb(mockDB, "mysql"), MasterMaxQueryDurationWarning: 100 * time.Millisecond}
				repo := NewBoilerplateDatabaseRepository[testEntityForBulkExec](boilerplateDB)
				return repo, mock
			},
			ctx:         context.Background(),
			entities:    []testEntityForBulkExec{{ID: 1, Name: "Updated 1"}, {ID: 2, Name: "Updated 2"}},
			expectError: false,
			validate:    func(t *testing.T, err error) { assert.NoError(t, err) },
		},
		{
			name: "bulk update with empty entities",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForBulkExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{Master: sqlx.NewDb(mockDB, "mysql"), MasterMaxQueryDurationWarning: 100 * time.Millisecond}
				repo := NewBoilerplateDatabaseRepository[testEntityForBulkExec](boilerplateDB)
				return repo, mock
			},
			ctx:         context.Background(),
			entities:    []testEntityForBulkExec{},
			expectError: false,
			validate:    func(t *testing.T, err error) { assert.NoError(t, err) },
		},
		{
			name: "bulk update with BuildBulkUpdateQuery error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForBulkExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{Master: sqlx.NewDb(mockDB, "invalid_dialect"), MasterMaxQueryDurationWarning: 100 * time.Millisecond}
				repo := NewBoilerplateDatabaseRepository[testEntityForBulkExec](boilerplateDB)
				return repo, mock
			},
			ctx:         context.Background(),
			entities:    []testEntityForBulkExec{{ID: 1, Name: "Test"}},
			expectError: true,
			validate:    func(t *testing.T, err error) { assert.NotNil(t, err) },
		},
		{
			name: "bulk update with exec error",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForBulkExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{Master: sqlx.NewDb(mockDB, "mysql"), MasterMaxQueryDurationWarning: 100 * time.Millisecond}
				repo := NewBoilerplateDatabaseRepository[testEntityForBulkExec](boilerplateDB)
				return repo, mock
			},
			ctx:         context.Background(),
			entities:    []testEntityForBulkExec{{ID: 1, Name: "Test"}},
			expectError: true,
			validate:    func(t *testing.T, err error) { assert.NotNil(t, err) },
		},
		{
			name: "bulk update with postgres dialect",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForBulkExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{Master: sqlx.NewDb(mockDB, "postgres"), MasterMaxQueryDurationWarning: 100 * time.Millisecond}
				repo := NewBoilerplateDatabaseRepository[testEntityForBulkExec](boilerplateDB)
				return repo, mock
			},
			ctx:         context.Background(),
			entities:    []testEntityForBulkExec{{ID: 1, Name: "Entity 1"}, {ID: 2, Name: "Entity 2"}},
			expectError: false,
			validate:    func(t *testing.T, err error) { assert.NoError(t, err) },
		},
		{
			name: "bulk update with sqlserver dialect",
			setupRepo: func() (*BoilerplateDatabaseRepository[testEntityForBulkExec], sqlmock.Sqlmock) {
				mockDB, mock, _ := sqlmock.New()
				boilerplateDB := &boilerplate_database.BoilerplateDatabase{Master: sqlx.NewDb(mockDB, "sqlserver"), MasterMaxQueryDurationWarning: 100 * time.Millisecond}
				repo := NewBoilerplateDatabaseRepository[testEntityForBulkExec](boilerplateDB)
				return repo, mock
			},
			ctx:         context.Background(),
			entities:    []testEntityForBulkExec{{ID: 1, Name: "Entity 1"}, {ID: 2, Name: "Entity 2"}},
			expectError: false,
			validate:    func(t *testing.T, err error) { assert.NoError(t, err) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := tt.setupRepo()
			switch tt.name {
			case "bulk update successfully with multiple entities":
				mock.ExpectPrepare("UPDATE (.+)").WillBeClosed()
				mock.ExpectExec("UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 2))
			case "bulk update with exec error":
				mock.ExpectPrepare("UPDATE (.+)").WillBeClosed()
				mock.ExpectExec("UPDATE (.+)").WillReturnError(errors.New("exec error"))
			case "bulk update with postgres dialect":
				mock.ExpectPrepare("UPDATE (.+)").WillBeClosed()
				mock.ExpectExec("UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 2))
			case "bulk update with sqlserver dialect":
				mock.ExpectPrepare("UPDATE (.+)").WillBeClosed()
				mock.ExpectExec("UPDATE (.+)").WillReturnResult(sqlmock.NewResult(0, 2))
			}
			err := repo.BulkUpdate(tt.ctx, tt.entities)
			if tt.expectError {
				assert.NotNil(t, err)
			}
			if !tt.expectError {
				assert.NoError(t, err)
			}
			if tt.validate != nil {
				tt.validate(t, err)
			}
			if len(tt.entities) > 0 && tt.name != "bulk update with BuildBulkUpdateQuery error" && tt.name != "bulk update with exec error" {
				assert.NoError(t, mock.ExpectationsWereMet())
			}
		})
	}
}
