package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"go-boilerplate/datasources/boilerplate_database"
	"go-boilerplate/pkg/tracer"
	"net/http"
	"reflect"
	"time"

	"github.com/fikri240794/gocerr"
	"github.com/fikri240794/goqube"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"
)

//mockery:generate: true
//mockery:structname: BoilerplateDatabaseStatementMock
//mockery:filename: boilerplate_database_statement_mock.go
//mockery:output: internal/repositories/mocks/
type IBoilerplateDatabaseStatement interface {
	Exec(ctx context.Context, args ...interface{}) error
	Get(ctx context.Context, dest interface{}, args ...interface{}) error
	Select(ctx context.Context, dest interface{}, args ...interface{}) error
	Close() error
}

type boilerplateDatabaseStatement struct {
	stmt *sqlx.Stmt
}

func newBoilerplateDatabaseStatement(stmt *sqlx.Stmt) *boilerplateDatabaseStatement {
	return &boilerplateDatabaseStatement{
		stmt: stmt,
	}
}

func (r *boilerplateDatabaseStatement) Exec(ctx context.Context, args ...interface{}) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		err       error
	)

	ctx, span = tracer.Start(ctx, "[boilerplateDatabaseStatement][Exec]")
	defer span.End()

	logFields = map[string]interface{}{
		"args": args,
	}

	_, err = r.stmt.ExecContext(ctx, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[boilerplateDatabaseStatement][Exec][ExecContext] failed to exec statement")
		return err
	}

	return nil
}

func (r *boilerplateDatabaseStatement) Get(ctx context.Context, dest interface{}, args ...interface{}) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		err       error
	)

	ctx, span = tracer.Start(ctx, "[boilerplateDatabaseStatement][Get]")
	defer span.End()

	logFields = map[string]interface{}{
		"args": args,
	}

	err = r.stmt.GetContext(ctx, dest, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[boilerplateDatabaseStatement][Get][GetContext] failed to get with statement")
		return err
	}

	return nil
}

func (r *boilerplateDatabaseStatement) Select(ctx context.Context, dest interface{}, args ...interface{}) error {
	var (
		span      trace.Span
		logFields map[string]interface{}
		err       error
	)

	ctx, span = tracer.Start(ctx, "[boilerplateDatabaseStatement][Select]")
	defer span.End()

	logFields = map[string]interface{}{
		"args": args,
	}

	err = r.stmt.SelectContext(ctx, dest, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[boilerplateDatabaseStatement][Select][SelectContext] failed to select with statement")
		return err
	}

	return nil
}

func (r *boilerplateDatabaseStatement) Close() error {
	return r.stmt.Stmt.Close()
}

//mockery:generate: true
//mockery:structname: BoilerplateDatabaseTransactionMock
//mockery:filename: boilerplate_database_transaction_mock.go
//mockery:output: internal/repositories/mocks/
type IBoilerplateDatabaseTransaction interface {
	Commit() error
	DriverName() string
	Prepare(ctx context.Context, query string) (IBoilerplateDatabaseStatement, error)
	Rollback() error
}

type boilerplateDatabaseTransaction struct {
	tx *sqlx.Tx
}

func newBoilerplateDatabaseTransaction(tx *sqlx.Tx) *boilerplateDatabaseTransaction {
	return &boilerplateDatabaseTransaction{
		tx: tx,
	}
}

func (r *boilerplateDatabaseTransaction) Commit() error {
	var err error = r.tx.Commit()
	if err != nil {
		log.Err(err).
			Msg("[boilerplateDatabaseTransaction][Commit] failed to commit transaction")
		err = gocerr.New(http.StatusInternalServerError, "error")
		return err
	}
	return nil
}

func (r *boilerplateDatabaseTransaction) DriverName() string {
	return r.tx.DriverName()
}

func (r *boilerplateDatabaseTransaction) Prepare(ctx context.Context, query string) (IBoilerplateDatabaseStatement, error) {
	var (
		span      trace.Span
		logFields map[string]interface{}
		stmt      *sqlx.Stmt
		err       error
	)

	ctx, span = tracer.Start(ctx, "[boilerplateDatabaseTransaction][Prepare]")
	defer span.End()

	logFields = map[string]interface{}{
		"query": query,
	}

	stmt, err = r.tx.PreparexContext(ctx, query)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[boilerplateDatabaseTransaction][Prepare][PreparexContext] failed to prepare statement")
		return nil, err
	}

	return newBoilerplateDatabaseStatement(stmt), nil
}

func (r *boilerplateDatabaseTransaction) Rollback() error {
	var err error = r.tx.Rollback()
	if err != nil {
		log.Err(err).
			Msg("[boilerplateDatabaseTransaction][Rollback] failed to rollback transaction")
		err = gocerr.New(http.StatusInternalServerError, "error")
		return err
	}

	return nil
}

//mockery:generate: true
//mockery:structname: BoilerplateDatabaseRepositoryMock
//mockery:filename: boilerplate_database_repository_mock.go
//mockery:output: internal/repositories/mocks/
type IBoilerplateDatabaseRepository[TEntity interface{}] interface {
	BeginTransaction(ctx context.Context) (IBoilerplateDatabaseTransaction, error)

	BulkCreate(ctx context.Context, entities []TEntity) error

	BulkUpdate(ctx context.Context, entities []TEntity) error

	Count(
		ctx context.Context,
		filter *goqube.Filter,
		useMaster bool,
	) (uint64, error)

	Create(ctx context.Context, entity *TEntity) error

	Delete(ctx context.Context, filter *goqube.Filter) error

	FindAll(
		ctx context.Context,
		filter *goqube.Filter,
		sorts []goqube.Sort,
		take uint64,
		skip uint64,
		useMaster bool,
	) ([]TEntity, error)

	FindOne(
		ctx context.Context,
		filter *goqube.Filter,
		sorts []goqube.Sort,
		useMaster bool,
	) (*TEntity, error)

	Update(
		ctx context.Context,
		entity *TEntity,
		filter *goqube.Filter,
	) error
}

type BoilerplateDatabaseRepository[TEntity interface{}] struct {
	db *boilerplate_database.BoilerplateDatabase
	tx IBoilerplateDatabaseTransaction
}

func NewBoilerplateDatabaseRepository[TEntity interface{}](db *boilerplate_database.BoilerplateDatabase) *BoilerplateDatabaseRepository[TEntity] {
	return &BoilerplateDatabaseRepository[TEntity]{
		db: db,
	}
}

func (r *BoilerplateDatabaseRepository[TEntity]) getTableNameAndFields() (string, []string) {
	var (
		tableName  string
		fields     []string
		entityType reflect.Type
	)

	entityType = reflect.TypeOf(new(TEntity)).Elem()

	for i := range entityType.NumField() {
		var (
			field reflect.StructField
			tag   string
		)

		field = entityType.Field(i)

		tag = field.Tag.Get("table")
		if tag != "" && tag != "-" {
			tableName = tag
			continue
		}

		tag = field.Tag.Get("db")
		if tag != "" && tag != "-" {
			fields = append(fields, tag)
		}
	}

	return tableName, fields
}

type entityMeta struct {
	TableName     string
	PrimaryKey    string
	FieldTypeMap  map[string]string
	FieldValueMap map[string]interface{}
}

func (r *BoilerplateDatabaseRepository[TEntity]) getEntityMeta(entity *TEntity) entityMeta {
	var (
		entityType reflect.Type
		field      reflect.StructField
		tagValue   string
		dbTag      string
		meta       entityMeta
	)

	meta = entityMeta{
		FieldValueMap: map[string]interface{}{},
		FieldTypeMap:  map[string]string{},
	}

	entityType = reflect.TypeOf(entity).Elem()

	for i := range entityType.NumField() {
		field = entityType.Field(i)

		tagValue = field.Tag.Get("table")
		if tagValue != "" && tagValue != "-" {
			meta.TableName = tagValue
			continue
		}

		dbTag = field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		meta.FieldValueMap[dbTag] = reflect.ValueOf(entity).Elem().Field(i).Interface()

		tagValue = field.Tag.Get("db_type")
		if tagValue != "" {
			meta.FieldTypeMap[dbTag] = tagValue
		}

		tagValue = field.Tag.Get("primary_key")
		if tagValue == "true" {
			meta.PrimaryKey = dbTag
		}
	}

	return meta
}

func (r *BoilerplateDatabaseRepository[TEntity]) prepareQueryStatement(
	ctx context.Context,
	logFields map[string]interface{},
	query string,
	useMaster bool,
	fnName string,
) (IBoilerplateDatabaseStatement, error) {
	var (
		stmt     IBoilerplateDatabaseStatement
		sqlxStmt *sqlx.Stmt
		err      error
	)

	if useMaster {
		if r.tx != nil {
			stmt, err = r.tx.Prepare(ctx, query)
			if err != nil {
				log.Err(err).
					Ctx(ctx).
					Fields(logFields).
					Msg(fmt.Sprintf("[BoilerplateDatabaseRepository][%s][Prepare] failed to prepare statement", fnName))
				err = gocerr.New(http.StatusInternalServerError, "error")
				return nil, err
			}
		} else {
			sqlxStmt, err = r.db.Master.PreparexContext(ctx, query)
			if err != nil {
				log.Err(err).
					Ctx(ctx).
					Fields(logFields).
					Msg(fmt.Sprintf("[BoilerplateDatabaseRepository][%s][PreparexContext] failed to prepare statement", fnName))
				err = gocerr.New(http.StatusInternalServerError, "error")
				return nil, err
			}

			stmt = newBoilerplateDatabaseStatement(sqlxStmt)
		}
	} else {
		sqlxStmt, err = r.db.Slave.PreparexContext(ctx, query)
		if err != nil {
			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg(fmt.Sprintf("[BoilerplateDatabaseRepository][%s][PreparexContext] failed to prepare statement", fnName))
			err = gocerr.New(http.StatusInternalServerError, "error")
			return nil, err
		}

		stmt = newBoilerplateDatabaseStatement(sqlxStmt)
	}

	return stmt, nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) logSlowQuery(
	logFields map[string]interface{},
	duration time.Duration,
	threshold time.Duration,
	fnName string,
) {
	logFields["duration"] = fmt.Sprintf("%.3f ms", (float64(duration) / float64(time.Millisecond)))

	if duration > threshold {
		log.Warn().
			Ctx(context.Background()).
			Fields(logFields).
			Msg(fmt.Sprintf("[BoilerplateDatabaseRepository][%s] slow query", fnName))
	}
}

func (r *BoilerplateDatabaseRepository[TEntity]) exec(ctx context.Context, query string, args ...interface{}) error {
	var (
		span               trace.Span
		logFields          map[string]interface{}
		stmt               IBoilerplateDatabaseStatement
		queryExecStartTime time.Time
		queryExecEndTime   time.Time
		queryExecDuration  time.Duration
		err                error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][exec]")
	defer span.End()

	logFields = map[string]interface{}{
		"query": query,
		"args":  args,
	}

	stmt, err = r.prepareQueryStatement(ctx, logFields, query, true, "exec")
	if err != nil {
		return err
	}
	defer func() {
		var errCloseStmt = stmt.Close()
		if errCloseStmt != nil {
			log.Err(errCloseStmt).
				Ctx(ctx).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][exec][Close] failed to close statement")
		}
	}()

	queryExecStartTime = time.Now()

	err = stmt.Exec(ctx, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][exec][Exec] failed to exec statement")
		err = gocerr.New(http.StatusInternalServerError, "error")
		return err
	}

	queryExecEndTime = time.Now()
	queryExecDuration = queryExecEndTime.Sub(queryExecStartTime)
	logFields["duration"] = fmt.Sprintf("%.3f ms", (float64(queryExecDuration) / float64(time.Millisecond)))

	if queryExecDuration > r.db.MasterMaxQueryDurationWarning {
		log.Warn().
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][exec] slow query")
	}

	return nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) BeginTransaction(ctx context.Context) (IBoilerplateDatabaseTransaction, error) {
	var (
		span      trace.Span
		logFields map[string]interface{}
		sqlxTx    *sqlx.Tx
		tx        IBoilerplateDatabaseTransaction
		err       error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][BeginTransaction]")
	defer span.End()

	if r.tx != nil {
		return r.tx, nil
	}

	logFields = map[string]interface{}{}

	sqlxTx, err = r.db.Master.BeginTxx(ctx, nil)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][BeginTransaction][BeginTxx] failed to begin transaction")
		err = gocerr.New(http.StatusInternalServerError, "error")
		return nil, err
	}

	tx = newBoilerplateDatabaseTransaction(sqlxTx)

	return tx, nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) Count(
	ctx context.Context,
	filter *goqube.Filter,
	useMaster bool,
) (uint64, error) {
	var (
		span               trace.Span
		logFields          map[string]interface{}
		tableName          string
		selectQuery        *goqube.SelectQuery
		dialect            goqube.Dialect
		query              string
		args               []interface{}
		stmt               IBoilerplateDatabaseStatement
		queryExecStartTime time.Time
		queryExecEndTime   time.Time
		queryExecDuration  time.Duration
		count              uint64
		err                error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][Count]")
	defer span.End()

	logFields = map[string]interface{}{
		"useMaster": useMaster,
	}

	tableName, _ = r.getTableNameAndFields()

	selectQuery = &goqube.SelectQuery{
		Fields: []goqube.Field{{Column: "COUNT(-1)"}},
		Table:  goqube.Table{Name: tableName},
		Filter: filter,
	}
	logFields["selectQuery"] = selectQuery

	dialect = goqube.Dialect(r.db.Slave.DriverName())
	if useMaster {
		if r.tx != nil {
			dialect = goqube.Dialect(r.tx.DriverName())
		} else {
			dialect = goqube.Dialect(r.db.Master.DriverName())
		}
	}
	logFields["dialect"] = dialect

	query, args, err = selectQuery.BuildSelectQuery(dialect)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][Count][ToSQLWithArgsWithAlias] failed to build select query")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return 0, err
	}
	logFields["query"] = query
	logFields["args"] = args

	stmt, err = r.prepareQueryStatement(ctx, logFields, query, useMaster, "Count")
	if err != nil {
		return 0, err
	}
	defer func() {
		var errCloseStmt = stmt.Close()
		if errCloseStmt != nil {
			log.Err(errCloseStmt).
				Ctx(ctx).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][Count][Close] failed to close statement")
		}
	}()

	queryExecStartTime = time.Now()

	err = stmt.Get(ctx, &count, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			err = gocerr.New(http.StatusNotFound, "entity not found")
		} else {
			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][Count][Get] failed to count entities")
			err = gocerr.New(http.StatusInternalServerError, "error")
		}

		return 0, err
	}

	queryExecEndTime = time.Now()
	queryExecDuration = queryExecEndTime.Sub(queryExecStartTime)
	r.logSlowQuery(logFields, queryExecDuration, r.db.SlaveMaxQueryDurationWarning, "Count")

	return count, nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) Create(ctx context.Context, entity *TEntity) error {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		meta        entityMeta
		insertQuery *goqube.InsertQuery
		dialect     goqube.Dialect
		query       string
		args        []interface{}
		err         error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][Create]")
	defer span.End()

	if entity == nil {
		return nil
	}

	logFields = map[string]interface{}{
		"entity": entity,
	}

	meta = r.getEntityMeta(entity)

	insertQuery = &goqube.InsertQuery{
		Table:  meta.TableName,
		Values: []map[string]interface{}{meta.FieldValueMap},
	}
	logFields["insertQuery"] = insertQuery

	dialect = goqube.Dialect(r.db.Master.DriverName())
	logFields["dialect"] = dialect

	query, args, err = insertQuery.BuildInsertQuery(dialect)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][Create][ToSQLWithArgs] failed to build insert query")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}
	logFields["query"] = query
	logFields["args"] = args

	err = r.exec(ctx, query, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][Create][exec] failed to create entity")
		return err
	}

	return nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) Delete(ctx context.Context, filter *goqube.Filter) error {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		tableName   string
		deleteQuery *goqube.DeleteQuery
		dialect     goqube.Dialect
		query       string
		args        []interface{}
		err         error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][Delete]")
	defer span.End()

	logFields = map[string]interface{}{}

	tableName, _ = r.getTableNameAndFields()

	deleteQuery = &goqube.DeleteQuery{
		Table:  tableName,
		Filter: filter,
	}
	logFields["deleteQuery"] = deleteQuery

	dialect = goqube.Dialect(r.db.Master.DriverName())
	logFields["dialect"] = dialect

	query, args, err = deleteQuery.BuildDeleteQuery(dialect)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][Delete][ToSQLWithArgs] failed to build delete query")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}
	logFields["query"] = query
	logFields["args"] = args

	err = r.exec(ctx, query, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][Delete][exec] failed to delete entity")
		return err
	}

	return nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) FindAll(
	ctx context.Context,
	filter *goqube.Filter,
	sorts []goqube.Sort,
	take uint64,
	skip uint64,
	useMaster bool,
) ([]TEntity, error) {
	var (
		span               trace.Span
		logFields          map[string]interface{}
		tableName          string
		fields             []string
		selectFields       []goqube.Field
		selectQuery        *goqube.SelectQuery
		dialect            goqube.Dialect
		query              string
		args               []interface{}
		stmt               IBoilerplateDatabaseStatement
		queryExecStartTime time.Time
		queryExecEndTime   time.Time
		queryExecDuration  time.Duration
		entities           []TEntity
		err                error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][FindAll]")
	defer span.End()

	logFields = map[string]interface{}{
		"useMaster": useMaster,
	}

	tableName, fields = r.getTableNameAndFields()

	selectFields = []goqube.Field{}
	for i := range fields {
		selectFields = append(selectFields, goqube.Field{Column: fields[i]})
	}

	selectQuery = &goqube.SelectQuery{
		Fields: selectFields,
		Table:  goqube.Table{Name: tableName},
		Filter: filter,
		Sorts:  sorts,
		Take:   take,
		Skip:   skip,
	}
	logFields["selectQuery"] = selectQuery

	dialect = goqube.Dialect(r.db.Slave.DriverName())
	if useMaster {
		if r.tx != nil {
			dialect = goqube.Dialect(r.tx.DriverName())
		} else {
			dialect = goqube.Dialect(r.db.Master.DriverName())
		}
	}
	logFields["dialect"] = dialect

	query, args, err = selectQuery.BuildSelectQuery(dialect)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][FindAll][ToSQLWithArgsWithAlias] failed to build select query")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return nil, err
	}
	logFields["query"] = query
	logFields["args"] = args

	stmt, err = r.prepareQueryStatement(ctx, logFields, query, useMaster, "FindAll")
	if err != nil {
		return nil, err
	}
	defer func() {
		var errCloseStmt = stmt.Close()
		if errCloseStmt != nil {
			log.Err(errCloseStmt).
				Ctx(ctx).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][FindAll][Close] failed to close statement")
		}
	}()

	queryExecStartTime = time.Now()

	entities = []TEntity{}
	err = stmt.Select(ctx, &entities, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][FindAll][Select] failed to select entities")
		err = gocerr.New(http.StatusInternalServerError, "error")
		return nil, err
	}

	queryExecEndTime = time.Now()
	queryExecDuration = queryExecEndTime.Sub(queryExecStartTime)
	r.logSlowQuery(logFields, queryExecDuration, r.db.SlaveMaxQueryDurationWarning, "FindAll")

	return entities, nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) FindOne(
	ctx context.Context,
	filter *goqube.Filter,
	sorts []goqube.Sort,
	useMaster bool,
) (*TEntity, error) {
	var (
		span               trace.Span
		logFields          map[string]interface{}
		tableName          string
		fields             []string
		selectFields       []goqube.Field
		selectQuery        *goqube.SelectQuery
		dialect            goqube.Dialect
		query              string
		args               []interface{}
		stmt               IBoilerplateDatabaseStatement
		queryExecStartTime time.Time
		queryExecEndTime   time.Time
		queryExecDuration  time.Duration
		entity             *TEntity
		err                error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][FindOne]")
	defer span.End()

	logFields = map[string]interface{}{
		"useMaster": useMaster,
	}

	tableName, fields = r.getTableNameAndFields()

	selectFields = []goqube.Field{}
	for i := range fields {
		selectFields = append(selectFields, goqube.Field{Column: fields[i]})
	}

	selectQuery = &goqube.SelectQuery{
		Fields: selectFields,
		Table:  goqube.Table{Name: tableName},
		Filter: filter,
		Sorts:  sorts,
		Take:   1,
	}
	logFields["selectQuery"] = selectQuery

	dialect = goqube.Dialect(r.db.Slave.DriverName())
	if useMaster {
		if r.tx != nil {
			dialect = goqube.Dialect(r.tx.DriverName())
		} else {
			dialect = goqube.Dialect(r.db.Master.DriverName())
		}
	}
	logFields["dialect"] = dialect

	query, args, err = selectQuery.BuildSelectQuery(dialect)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][FindOne][ToSQLWithArgsWithAlias] failed to build select query")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return nil, err
	}
	logFields["query"] = query
	logFields["args"] = args

	stmt, err = r.prepareQueryStatement(ctx, logFields, query, useMaster, "FindOne")
	if err != nil {
		return nil, err
	}
	defer func() {
		var errCloseStmt = stmt.Close()
		if errCloseStmt != nil {
			log.Err(errCloseStmt).
				Ctx(ctx).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][FindOne][Close] failed to close statement")
		}
	}()

	queryExecStartTime = time.Now()

	entity = new(TEntity)
	err = stmt.Get(ctx, entity, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			err = gocerr.New(http.StatusNotFound, "entity not found")
		} else {
			log.Err(err).
				Ctx(ctx).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][FindOne][GetContext] failed to select entity")
			err = gocerr.New(http.StatusInternalServerError, "error")
		}

		return nil, err
	}

	queryExecEndTime = time.Now()
	queryExecDuration = queryExecEndTime.Sub(queryExecStartTime)
	r.logSlowQuery(logFields, queryExecDuration, r.db.SlaveMaxQueryDurationWarning, "FindOne")

	return entity, nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) Update(ctx context.Context, entity *TEntity, filter *goqube.Filter) error {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		meta        entityMeta
		updateQuery *goqube.UpdateQuery
		dialect     goqube.Dialect
		query       string
		args        []interface{}
		err         error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][Update]")
	defer span.End()

	if entity == nil {
		return nil
	}

	logFields = map[string]interface{}{
		"entity": entity,
	}

	meta = r.getEntityMeta(entity)

	updateQuery = &goqube.UpdateQuery{
		Table:       meta.TableName,
		FieldsValue: meta.FieldValueMap,
		Filter:      filter,
	}
	logFields["updateQuery"] = updateQuery

	dialect = goqube.Dialect(r.db.Master.DriverName())
	logFields["dialect"] = dialect

	query, args, err = updateQuery.BuildUpdateQuery(dialect)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][Update][ToSQLWithArgs] failed to build update query")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err

	}
	logFields["query"] = query
	logFields["args"] = args

	err = r.exec(ctx, query, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][Update][exec] failed to update entity")
		return err
	}

	return nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) BulkCreate(ctx context.Context, entities []TEntity) error {
	var (
		span         trace.Span
		logFields    map[string]interface{}
		fieldsValues []map[string]interface{}
		meta         entityMeta
		insertQuery  *goqube.InsertQuery
		dialect      goqube.Dialect
		query        string
		args         []interface{}
		err          error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][BulkCreate]")
	defer span.End()

	if len(entities) <= 0 {
		return nil
	}

	logFields = map[string]interface{}{
		"entities": entities,
	}

	fieldsValues = []map[string]interface{}{}
	for i := range entities {
		meta = r.getEntityMeta(&entities[i])
		fieldsValues = append(fieldsValues, meta.FieldValueMap)
	}

	insertQuery = &goqube.InsertQuery{
		Table:  meta.TableName,
		Values: fieldsValues,
	}
	logFields["insertQuery"] = insertQuery

	dialect = goqube.Dialect(r.db.Master.DriverName())
	logFields["dialect"] = dialect

	query, args, err = insertQuery.BuildInsertQuery(dialect)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][BulkCreate][ToSQLWithArgs] failed to build insert query")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}
	logFields["query"] = query
	logFields["args"] = args

	err = r.exec(ctx, query, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][BulkCreate][exec] failed to bulk create entities")
		return err
	}

	return nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) BulkUpdate(ctx context.Context, entities []TEntity) error {
	var (
		span            trace.Span
		logFields       map[string]interface{}
		fieldsValues    []map[string]interface{}
		meta            entityMeta
		bulkUpdateQuery *goqube.BulkUpdateQuery
		dialect         goqube.Dialect
		query           string
		args            []interface{}
		err             error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][BulkUpdate]")
	defer span.End()

	if len(entities) <= 0 {
		return nil
	}
	logFields = map[string]interface{}{
		"entities": entities,
	}

	fieldsValues = []map[string]interface{}{}
	for i := range entities {
		meta = r.getEntityMeta(&entities[i])
		fieldsValues = append(fieldsValues, meta.FieldValueMap)
	}

	bulkUpdateQuery = &goqube.BulkUpdateQuery{
		Table:        meta.TableName,
		PrimaryKey:   meta.PrimaryKey,
		FieldsValues: fieldsValues,
		ColumnsType:  meta.FieldTypeMap,
	}
	logFields["bulkUpdateQuery"] = bulkUpdateQuery

	dialect = goqube.Dialect(r.db.Master.DriverName())
	logFields["dialect"] = dialect

	query, args, err = bulkUpdateQuery.BuildBulkUpdateQuery(dialect)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][BulkUpdate][BuildBulkUpdateQuery] failed to build bulk update query")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}
	logFields["query"] = query
	logFields["args"] = args

	err = r.exec(ctx, query, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][BulkUpdate][exec] failed to bulk update entities")
		return err
	}

	return nil
}
