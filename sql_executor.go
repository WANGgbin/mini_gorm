package gorm

import (
	"context"
	"database/sql"
)

type ExecMode uint8

const (
	ExecModeQueryRow ExecMode = iota
	ExecModeQuery
	ExecModeExec
)

type SqlExecutor interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// 3 中实现，db、tx、stmt

type DBExecutorImpl struct {
	db *sql.DB
}

func NewDBExecutor(db *sql.DB) SqlExecutor {
	return &DBExecutorImpl{
		db: db,
	}
}

func (d *DBExecutorImpl) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *DBExecutorImpl) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *DBExecutorImpl) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

var _ SqlExecutor = (*StmtExecutorImpl)(nil)

type StmtExecutorImpl struct {
	stmt *sql.Stmt
}

func NewStmtExecutor(db *DB) (SqlExecutor, error) {
	var err error
	var stmt *sql.Stmt

	stmtI, exist := db.stmtCache.Load(db.stmt.query)
	if !exist {
		// 如果处在事务中，使用事务创建 stmt
		if db.isInTx() {
			stmt, err = db.getRawSqlTx().PrepareContext(db.stmt.ctx, db.stmt.query)
		} else {
			stmt, err = db.db.PrepareContext(db.stmt.ctx, db.stmt.query)
		}
		if err != nil {
			return nil, err
		}
		db.stmtCache.Store(db.stmt.query, stmt)
	} else {
		stmt = stmtI.(*sql.Stmt)
	}

	return &StmtExecutorImpl{stmt: stmt}, nil
}

func (s *StmtExecutorImpl) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return s.stmt.QueryContext(ctx, args...)
}

func (s *StmtExecutorImpl) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return s.stmt.QueryRowContext(ctx, args...)
}

func (s *StmtExecutorImpl) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return s.stmt.ExecContext(ctx, args...)
}

var _ SqlExecutor = (*TxExecutorImpl)(nil)
type TxExecutorImpl struct {
	tx *sql.Tx
}

func NewTxExecutor(tx *sql.Tx) *TxExecutorImpl {
	return &TxExecutorImpl{
		tx: tx,
	}
}

func (t *TxExecutorImpl) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *TxExecutorImpl) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *TxExecutorImpl) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}