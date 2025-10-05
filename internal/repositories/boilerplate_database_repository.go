package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"go-boilerplate/datasources/boilerplate_database"
	"go-boilerplate/pkg/constants"
	custom_context "go-boilerplate/pkg/context"
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
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"args":      args,
	}

	_, err = r.stmt.ExecContext(ctx, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[boilerplateDatabaseStatement][Exec][ExecContext] failed to exec statement")
		log.Debug().
			Ctx(ctx).
			Err(err).
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
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"args":      args,
	}

	err = r.stmt.GetContext(ctx, dest, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[boilerplateDatabaseStatement][Get][GetContext] failed to get with statement")
		log.Debug().
			Ctx(ctx).
			Err(err).
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
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"args":      args,
	}

	err = r.stmt.SelectContext(ctx, dest, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[boilerplateDatabaseStatement][Select][SelectContext] failed to select with statement")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[boilerplateDatabaseStatement][Select][SelectContext] failed to select with statement")
		return err
	}

	return nil
}

func (r *boilerplateDatabaseStatement) Close() error {
	return r.stmt.Stmt.Close()
}

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
	return r.tx.Commit()
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
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"query":     query,
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
		return err
	}

	return nil
}

type IBoilerplateDatabaseRepository[TEntity interface{}] interface {
	BeginTransaction(ctx context.Context) (IBoilerplateDatabaseTransaction, error)

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
		}

		tag = field.Tag.Get("db")
		if tag != "" && tag != "-" {
			fields = append(fields, tag)
		}
	}

	return tableName, fields
}

func (r *BoilerplateDatabaseRepository[TEntity]) getTableNameAndMapFieldWithValueFrom(entity *TEntity) (string, map[string]interface{}) {
	var (
		tableName        string
		mapFieldAndValue map[string]interface{}
		entityType       reflect.Type
		field            reflect.StructField
		tag              string
	)

	mapFieldAndValue = map[string]interface{}{}

	entityType = reflect.TypeOf(entity).Elem()

	for i := range entityType.NumField() {
		field = entityType.Field(i)

		tag = field.Tag.Get("table")
		if tag != "" && tag != "-" {
			tableName = tag
			continue
		}

		tag = field.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue
		}

		mapFieldAndValue[tag] = reflect.ValueOf(entity).Elem().Field(i).Interface()
	}

	return tableName, mapFieldAndValue
}

func (r *BoilerplateDatabaseRepository[TEntity]) exec(ctx context.Context, query string, args ...interface{}) error {
	var (
		span               trace.Span
		logFields          map[string]interface{}
		stmt               IBoilerplateDatabaseStatement
		sqlxStmt           *sqlx.Stmt
		queryExecStartTime time.Time
		queryExecEndTime   time.Time
		queryExecDuration  time.Duration
		err                error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][exec]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"query":     query,
		"args":      args,
	}

	if r.tx != nil {
		stmt, err = r.tx.Prepare(ctx, query)
		if err != nil {
			log.Err(err).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Str("query", query).
				Msg("[BoilerplateDatabaseRepository][exec][Prepare] failed to prepare statement")
			log.Debug().
				Ctx(ctx).
				Err(err).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][exec][Prepare] failed to prepare statement")
			err = gocerr.New(http.StatusInternalServerError, err.Error())
			return err
		}
	} else {
		sqlxStmt, err = r.db.Master.PreparexContext(ctx, query)
		if err != nil {
			log.Err(err).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Str("query", query).
				Msg("[BoilerplateDatabaseRepository][exec][PreparexContext] failed to prepare statement")
			log.Debug().
				Ctx(ctx).
				Err(err).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][exec][PreparexContext] failed to prepare statement")
			err = gocerr.New(http.StatusInternalServerError, err.Error())
			return err
		}

		stmt = newBoilerplateDatabaseStatement(sqlxStmt)
	}
	defer func() {
		var errCloseStmt = stmt.Close()
		if errCloseStmt != nil {
			log.Err(errCloseStmt).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Str("query", query).
				Msg("[BoilerplateDatabaseRepository][exec][Close] failed to close statement")
			log.Debug().
				Ctx(ctx).
				Err(errCloseStmt).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][exec][Close] failed to close statement")
		}
	}()

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[BoilerplateDatabaseRepository][exec] query execution")

	queryExecStartTime = time.Now()

	err = stmt.Exec(ctx, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Str("query", query).
			Msg("[BoilerplateDatabaseRepository][exec][Exec] failed to exec statement")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][exec][Exec] failed to exec statement")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	queryExecEndTime = time.Now()
	queryExecDuration = queryExecEndTime.Sub(queryExecStartTime)
	logFields["duration"] = fmt.Sprintf("%.3f ms", (float64(queryExecDuration) / float64(time.Millisecond)))

	if queryExecDuration > r.db.MasterMaxQueryDurationWarning {
		log.Warn().
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Str("query", query).
			Str("duration", fmt.Sprintf("%.3f ms", (float64(queryExecDuration)/float64(time.Millisecond)))).
			Msg("[BoilerplateDatabaseRepository][exec] slow query")
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[BoilerplateDatabaseRepository][exec] query executed")

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

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
	}

	sqlxTx, err = r.db.Master.BeginTxx(ctx, nil)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][BeginTransaction][BeginTxx] failed to begin transaction")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
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
		table              string
		selectQuery        *goqube.SelectQuery
		dialect            goqube.Dialect
		query              string
		args               []interface{}
		stmt               IBoilerplateDatabaseStatement
		sqlxStmt           *sqlx.Stmt
		queryExecStartTime time.Time
		queryExecEndTime   time.Time
		queryExecDuration  time.Duration
		count              uint64
		err                error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][Count]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"useMaster": useMaster,
	}

	table, _ = r.getTableNameAndFields()

	selectQuery = &goqube.SelectQuery{
		Fields: []goqube.Field{goqube.Field{Column: "COUNT(-1)"}},
		Table:  goqube.Table{Name: table},
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
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[BoilerplateDatabaseRepository][Count][ToSQLWithArgsWithAlias] failed to build select query")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][Count][ToSQLWithArgsWithAlias] failed to build select query")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return 0, err
	}
	logFields["query"] = query
	logFields["args"] = args

	if useMaster {
		if r.tx != nil {
			stmt, err = r.tx.Prepare(ctx, query)
			if err != nil {
				log.Err(err).
					Ctx(ctx).
					Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
					Str("query", query).
					Msg("[BoilerplateDatabaseRepository][Count][Prepare] failed to prepare statement")
				log.Debug().
					Ctx(ctx).
					Err(err).
					Fields(logFields).
					Msg("[BoilerplateDatabaseRepository][Count][Prepare] failed to prepare statement")
				err = gocerr.New(http.StatusInternalServerError, err.Error())
				return 0, err
			}
		} else {
			sqlxStmt, err = r.db.Master.PreparexContext(ctx, query)
			if err != nil {
				log.Err(err).
					Ctx(ctx).
					Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
					Str("query", query).
					Msg("[BoilerplateDatabaseRepository][Count][PreparexContext] failed to prepare statement")
				log.Debug().
					Ctx(ctx).
					Err(err).
					Fields(logFields).
					Msg("[BoilerplateDatabaseRepository][Count][PreparexContext] failed to prepare statement")
				err = gocerr.New(http.StatusInternalServerError, err.Error())
				return 0, err
			}

			stmt = newBoilerplateDatabaseStatement(sqlxStmt)
		}
	} else {
		sqlxStmt, err = r.db.Slave.PreparexContext(ctx, query)
		if err != nil {
			log.Err(err).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Str("query", query).
				Msg("[BoilerplateDatabaseRepository][Count][PreparexContext] failed to prepare statement")
			log.Debug().
				Ctx(ctx).
				Err(err).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][Count][PreparexContext] failed to prepare statement")
			err = gocerr.New(http.StatusInternalServerError, err.Error())
			return 0, err
		}

		stmt = newBoilerplateDatabaseStatement(sqlxStmt)
	}
	defer func() {
		var errCloseStmt = stmt.Close()
		if errCloseStmt != nil {
			log.Err(errCloseStmt).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Str("query", query).
				Msg("[BoilerplateDatabaseRepository][Count][Close] failed to close statement")
			log.Debug().
				Ctx(ctx).
				Err(errCloseStmt).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][Count][Close] failed to close statement")
		}
	}()

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[BoilerplateDatabaseRepository][Count] query execution")

	queryExecStartTime = time.Now()

	err = stmt.Get(ctx, &count, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			err = gocerr.New(http.StatusNotFound, "entity not found")
		} else {
			log.Err(err).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Str("query", query).
				Msg("[BoilerplateDatabaseRepository][Count][Get] failed to count entities")
			log.Debug().
				Ctx(ctx).
				Err(err).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][Count][Get] failed to count entities")
			err = gocerr.New(http.StatusInternalServerError, err.Error())
		}

		return 0, err
	}

	queryExecEndTime = time.Now()
	queryExecDuration = queryExecEndTime.Sub(queryExecStartTime)
	logFields["duration"] = fmt.Sprintf("%.3f ms", (float64(queryExecDuration) / float64(time.Millisecond)))

	if queryExecDuration > r.db.SlaveMaxQueryDurationWarning {
		log.Warn().
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Str("query", query).
			Str("duration", fmt.Sprintf("%.3f ms", (float64(queryExecDuration)/float64(time.Millisecond)))).
			Msg("[BoilerplateDatabaseRepository][Count] slow query")
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[BoilerplateDatabaseRepository][Count] query executed")

	return count, nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) Create(ctx context.Context, entity *TEntity) error {
	var (
		span              trace.Span
		logFields         map[string]interface{}
		table             string
		mapFieldWithValue map[string]interface{}
		insertQuery       *goqube.InsertQuery
		dialect           goqube.Dialect
		query             string
		args              []interface{}
		err               error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][Create]")
	defer span.End()

	if entity == nil {
		return nil
	}

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"entity":    entity,
	}

	table, mapFieldWithValue = r.getTableNameAndMapFieldWithValueFrom(entity)

	insertQuery = &goqube.InsertQuery{
		Table:  table,
		Values: []map[string]interface{}{mapFieldWithValue},
	}
	logFields["insertQuery"] = insertQuery

	dialect = goqube.Dialect(r.db.Master.DriverName())
	logFields["dialect"] = dialect

	query, args, err = insertQuery.BuildInsertQuery(dialect)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[BoilerplateDatabaseRepository][Create][ToSQLWithArgs] failed to build insert query")
		log.Debug().
			Ctx(ctx).
			Err(err).
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
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Str("query", query).
			Msg("[BoilerplateDatabaseRepository][Create][exec] failed to create entity")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][Create][exec] failed to create entity")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) Delete(ctx context.Context, filter *goqube.Filter) error {
	var (
		span        trace.Span
		logFields   map[string]interface{}
		table       string
		deleteQuery *goqube.DeleteQuery
		dialect     goqube.Dialect
		query       string
		args        []interface{}
		err         error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][Delete]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
	}

	table, _ = r.getTableNameAndFields()

	deleteQuery = &goqube.DeleteQuery{
		Table:  table,
		Filter: filter,
	}
	logFields["deleteQuery"] = deleteQuery

	dialect = goqube.Dialect(r.db.Master.DriverName())
	logFields["dialect"] = dialect

	query, args, err = deleteQuery.BuildDeleteQuery(dialect)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[BoilerplateDatabaseRepository][Delete][ToSQLWithArgs] failed to build delete query")
		log.Debug().
			Ctx(ctx).
			Err(err).
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
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Str("query", query).
			Msg("[BoilerplateDatabaseRepository][Delete][exec] failed to delete entity")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][Delete][exec] failed to delete entity")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
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
		table              string
		fields             []string
		selectFields       []goqube.Field
		selectQuery        *goqube.SelectQuery
		dialect            goqube.Dialect
		query              string
		args               []interface{}
		stmt               IBoilerplateDatabaseStatement
		sqlxStmt           *sqlx.Stmt
		queryExecStartTime time.Time
		queryExecEndTime   time.Time
		queryExecDuration  time.Duration
		entities           []TEntity
		err                error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][FindAll]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"useMaster": useMaster,
	}

	table, fields = r.getTableNameAndFields()

	selectFields = []goqube.Field{}
	for i := range fields {
		selectFields = append(selectFields, goqube.Field{Column: fields[i]})
	}

	selectQuery = &goqube.SelectQuery{
		Fields: selectFields,
		Table:  goqube.Table{Name: table},
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
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[BoilerplateDatabaseRepository][FindAll][ToSQLWithArgsWithAlias] failed to build select query")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][FindAll][ToSQLWithArgsWithAlias] failed to build select query")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return nil, err
	}
	logFields["query"] = query
	logFields["args"] = args

	if useMaster {
		if r.tx != nil {
			stmt, err = r.tx.Prepare(ctx, query)
			if err != nil {
				log.Err(err).
					Ctx(ctx).
					Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
					Str("query", query).
					Msg("[BoilerplateDatabaseRepository][FindAll][Prepare] failed to prepare statement")
				log.Debug().
					Ctx(ctx).
					Err(err).
					Fields(logFields).
					Msg("[BoilerplateDatabaseRepository][FindAll][Prepare] failed to prepare statement")
				err = gocerr.New(http.StatusInternalServerError, err.Error())
				return nil, err
			}
		} else {
			sqlxStmt, err = r.db.Master.PreparexContext(ctx, query)
			if err != nil {
				log.Err(err).
					Ctx(ctx).
					Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
					Str("query", query).
					Msg("[BoilerplateDatabaseRepository][FindAll][PreparexContext] failed to prepare statement")
				log.Debug().
					Ctx(ctx).
					Err(err).
					Fields(logFields).
					Msg("[BoilerplateDatabaseRepository][FindAll][PreparexContext] failed to prepare statement")
				err = gocerr.New(http.StatusInternalServerError, err.Error())
				return nil, err
			}

			stmt = newBoilerplateDatabaseStatement(sqlxStmt)
		}
	} else {
		sqlxStmt, err = r.db.Slave.PreparexContext(ctx, query)
		if err != nil {
			log.Err(err).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Str("query", query).
				Msg("[BoilerplateDatabaseRepository][FindAll][PreparexContext] failed to prepare statement")
			log.Debug().
				Ctx(ctx).
				Err(err).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][FindAll][PreparexContext] failed to prepare statement")
			err = gocerr.New(http.StatusInternalServerError, err.Error())
			return nil, err
		}

		stmt = newBoilerplateDatabaseStatement(sqlxStmt)
	}
	defer func() {
		var errCloseStmt = stmt.Close()
		if errCloseStmt != nil {
			log.Err(errCloseStmt).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Str("query", query).
				Msg("[BoilerplateDatabaseRepository][FindAll][Close] failed to close statement")
			log.Debug().
				Ctx(ctx).
				Err(errCloseStmt).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][FindAll][Close] failed to close statement")
		}
	}()

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[BoilerplateDatabaseRepository][FindAll] query execution")

	queryExecStartTime = time.Now()

	entities = []TEntity{}
	err = stmt.Select(ctx, &entities, args...)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Str("query", query).
			Msg("[BoilerplateDatabaseRepository][FindAll][Select] failed to select entities")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][FindAll][Select] failed to select entities")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return nil, err
	}

	queryExecEndTime = time.Now()
	queryExecDuration = queryExecEndTime.Sub(queryExecStartTime)
	logFields["duration"] = fmt.Sprintf("%.3f ms", (float64(queryExecDuration) / float64(time.Millisecond)))

	if queryExecDuration > r.db.SlaveMaxQueryDurationWarning {
		log.Warn().
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Str("query", query).
			Str("duration", fmt.Sprintf("%.3f ms", (float64(queryExecDuration)/float64(time.Millisecond)))).
			Msg("[BoilerplateDatabaseRepository][FindAll] slow query")
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[BoilerplateDatabaseRepository][FindAll] query executed")

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
		table              string
		fields             []string
		selectFields       []goqube.Field
		selectQuery        *goqube.SelectQuery
		dialect            goqube.Dialect
		query              string
		args               []interface{}
		stmt               IBoilerplateDatabaseStatement
		sqlxStmt           *sqlx.Stmt
		queryExecStartTime time.Time
		queryExecEndTime   time.Time
		queryExecDuration  time.Duration
		entity             *TEntity
		err                error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][FindOne]")
	defer span.End()

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"useMaster": useMaster,
	}

	table, fields = r.getTableNameAndFields()

	selectFields = []goqube.Field{}
	for i := range fields {
		selectFields = append(selectFields, goqube.Field{Column: fields[i]})
	}

	selectQuery = &goqube.SelectQuery{
		Fields: selectFields,
		Table:  goqube.Table{Name: table},
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
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[BoilerplateDatabaseRepository][FindOne][ToSQLWithArgsWithAlias] failed to build select query")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][FindOne][ToSQLWithArgsWithAlias] failed to build select query")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return nil, err
	}
	logFields["query"] = query
	logFields["args"] = args

	if useMaster {
		if r.tx != nil {
			stmt, err = r.tx.Prepare(ctx, query)
			if err != nil {
				log.Err(err).
					Ctx(ctx).
					Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
					Str("query", query).
					Msg("[BoilerplateDatabaseRepository][FindOne][Prepare] failed to prepare statement")
				log.Debug().
					Ctx(ctx).
					Err(err).
					Fields(fields).
					Msg("[BoilerplateDatabaseRepository][FindOne][Prepare] failed to prepare statement")
				err = gocerr.New(http.StatusInternalServerError, err.Error())
				return nil, err
			}
		} else {
			sqlxStmt, err = r.db.Master.PreparexContext(ctx, query)
			if err != nil {
				log.Err(err).
					Ctx(ctx).
					Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
					Str("query", query).
					Msg("[BoilerplateDatabaseRepository][FindOne][PreparexContext] failed to prepare statement")
				log.Debug().
					Ctx(ctx).
					Err(err).
					Fields(fields).
					Msg("[BoilerplateDatabaseRepository][FindOne][PreparexContext] failed to prepare statement")
				err = gocerr.New(http.StatusInternalServerError, err.Error())
				return nil, err
			}

			stmt = newBoilerplateDatabaseStatement(sqlxStmt)
		}
	} else {
		sqlxStmt, err = r.db.Slave.PreparexContext(ctx, query)
		if err != nil {
			log.Err(err).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Str("query", query).
				Msg("[BoilerplateDatabaseRepository][FindOne][PreparexContext] failed to prepare statement")
			log.Debug().
				Ctx(ctx).
				Err(err).
				Fields(fields).
				Msg("[BoilerplateDatabaseRepository][FindOne][PreparexContext] failed to prepare statement")
			err = gocerr.New(http.StatusInternalServerError, err.Error())
			return nil, err
		}

		stmt = newBoilerplateDatabaseStatement(sqlxStmt)
	}
	defer func() {
		var errCloseStmt = stmt.Close()
		if errCloseStmt != nil {
			log.Err(errCloseStmt).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Str("query", query).
				Msg("[BoilerplateDatabaseRepository][FindOne][Close] failed to close statement")
			log.Debug().
				Ctx(ctx).
				Err(errCloseStmt).
				Fields(logFields).
				Msg("[BoilerplateDatabaseRepository][FindOne][Close] failed to close statement")
		}
	}()

	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Str("query", query).
			Msg("[BoilerplateDatabaseRepository][FindOne][PreparexContext] failed to prepare statement")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(fields).
			Msg("[BoilerplateDatabaseRepository][FindOne][PreparexContext] failed to prepare statement")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return nil, err
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[BoilerplateDatabaseRepository][FindOne] query execution")

	queryExecStartTime = time.Now()

	entity = new(TEntity)
	err = stmt.Get(ctx, entity, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			err = gocerr.New(http.StatusNotFound, "entity not found")
		} else {
			log.Err(err).
				Ctx(ctx).
				Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
				Str("query", query).
				Msg("[BoilerplateDatabaseRepository][FindOne][GetContext] failed to select entity")
			log.Debug().
				Ctx(ctx).
				Err(err).
				Fields(fields).
				Msg("[BoilerplateDatabaseRepository][FindOne][GetContext] failed to select entity")
			err = gocerr.New(http.StatusInternalServerError, err.Error())
		}

		return nil, err
	}

	queryExecEndTime = time.Now()
	queryExecDuration = queryExecEndTime.Sub(queryExecStartTime)
	logFields["duration"] = fmt.Sprintf("%.3f ms", (float64(queryExecDuration) / float64(time.Millisecond)))

	if queryExecDuration > r.db.SlaveMaxQueryDurationWarning {
		log.Warn().
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Str("query", query).
			Str("duration", fmt.Sprintf("%.3f ms", (float64(queryExecDuration)/float64(time.Millisecond)))).
			Msg("[BoilerplateDatabaseRepository][FindOne] slow query")
	}

	log.Debug().
		Ctx(ctx).
		Fields(logFields).
		Msg("[BoilerplateDatabaseRepository][FindOne] query executed")

	return entity, nil
}

func (r *BoilerplateDatabaseRepository[TEntity]) Update(ctx context.Context, entity *TEntity, filter *goqube.Filter) error {
	var (
		span               trace.Span
		logFields          map[string]interface{}
		table              string
		mapFieldWithValues map[string]interface{}
		updateQuery        *goqube.UpdateQuery
		dialect            goqube.Dialect
		query              string
		args               []interface{}
		err                error
	)

	ctx, span = tracer.Start(ctx, "[BoilerplateDatabaseRepository][Update]")
	defer span.End()

	if entity == nil {
		return nil
	}

	logFields = map[string]interface{}{
		"requestid": custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID),
		"entity":    entity,
	}

	table, mapFieldWithValues = r.getTableNameAndMapFieldWithValueFrom(entity)

	updateQuery = &goqube.UpdateQuery{
		Table:       table,
		FieldsValue: mapFieldWithValues,
		Filter:      filter,
	}
	logFields["updateQuery"] = updateQuery

	dialect = goqube.Dialect(r.db.Master.DriverName())
	logFields["dialect"] = dialect

	query, args, err = updateQuery.BuildUpdateQuery(dialect)
	if err != nil {
		log.Err(err).
			Ctx(ctx).
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Msg("[BoilerplateDatabaseRepository][Update][ToSQLWithArgs] failed to build update query")
		log.Debug().
			Ctx(ctx).
			Err(err).
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
			Str("requestid", custom_context.GetCtxValueSafely[string](ctx, constants.ContextKeyRequestID)).
			Str("query", query).
			Msg("[BoilerplateDatabaseRepository][Update][exec] failed to update entity")
		log.Debug().
			Ctx(ctx).
			Err(err).
			Fields(logFields).
			Msg("[BoilerplateDatabaseRepository][Update][exec] failed to update entity")
		err = gocerr.New(http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}
