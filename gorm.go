package gorm

import (
	"database/sql"
	"fmt"
	_ "github.com/WANGgbin/mini_mysql_driver"
	"sync"
)

type DBExecutor struct {
	executor SqlExecutor
	db       *sql.DB
	sqlStmt  *sql.Stmt
	sqlTx    *sql.Tx
}

type DBResult struct {
	rows         *sql.Rows
	rowsAffected int64
	lastInsertID int64
}

type DBCloneConfig struct {
	// 初始化一个新的 stmt
	newStmt bool
}

// DB 一条 SQL 执行的上下文
type DB struct {
	db        *sql.DB
	inTx      bool
	cfg       *DBConfig
	stmt      *statement
	err       error
	executor  SqlExecutor
	result    *DBResult
	stmtCache *sync.Map // TODO: 缓存淘汰策略？
	hks       *hooks
	cloneStmt bool
}

func Open(driver, dsn string, options ...DBOption) (*DB, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	cfg := newDBConfig()
	for _, opt := range options {
		opt(cfg)
	}

	return &DB{
		db:        db,
		cfg:       cfg,
		executor:  NewDBExecutor(db),
		stmtCache: new(sync.Map),
	}, nil
}

func (db *DB) new() *DB {
	if db.stmt != nil && !db.cloneStmt {
		return db
	}

	ret := &DB{
		db:        db.db,
		inTx:      db.inTx,
		cfg:       db.cfg,
		err:       db.err,
		stmtCache: db.stmtCache,
		executor:  db.executor,
	}

	if db.stmt == nil {
		ret.stmt = newStmt(ret)
	} else {
		ret.stmt = db.stmt.clone(ret)
	}

	return ret
}

// newTxDB 基于 db 创建一个 tx 的上下文
func (db *DB) newTx(cloneCfg *DBCloneConfig) *DB {
	ret := &DB{
		db:   db.db,
		inTx: true,
		cfg:  db.cfg,
		err:  db.err,
		hks:  db.hks,
		// 事务使用事务内的 stmt 缓存
		stmtCache: new(sync.Map),
		cloneStmt: true,
	}

	if db.stmt != nil && !cloneCfg.newStmt {
		ret.stmt = db.stmt.clone(ret)
	} else {
		ret.stmt = newStmt(ret)
	}

	return ret
}

// newSession 基于 db 创建一个 session
func (db *DB) newSession() *DB {
	ret := &DB{
		db:        db.db,
		cfg:       db.cfg.clone(),
		err:       db.err,
		stmtCache: db.stmtCache,
		executor:  db.executor,
		cloneStmt: true,
	}

	// session 拷贝一份 stmt
	ret.stmt = db.stmt.clone(ret)

	return ret
}

func (db *DB) setByTx(tx *DB) *DB {
	db.executor = tx.executor
	db.inTx = true
	return db
}

func (db *DB) isInTx() bool {
	return db.inTx
}

func (db *DB) isSetPrepareStmt() bool {
	return db.cfg.PrepareStmt
}

func (db *DB) addErr(e error) {
	if db.err == nil {
		db.err = e
		return
	}

	db.err = fmt.Errorf("%v; %w", db.err, e)
}

func (db *DB) isError() bool {
	return db.err != nil
}

/*
********** DBConfig **********
 */

type DBConfig struct {
	PrepareStmt       bool // 以 Prepare 方式执行 sql
	DryRun            bool
	AllowGlobalUpdate bool
	AllowGlobalDelete bool
	Debug             bool
}

func newDBConfig() *DBConfig {
	return &DBConfig{
		PrepareStmt: true,
	}
}

// clone 深拷贝一份配置
func (cfg *DBConfig) clone() *DBConfig {
	cp := *cfg
	return &cp
}

type DBOption func(cfg *DBConfig)

func WithPrepareStmt() DBOption {
	return func(cfg *DBConfig) {
		cfg.PrepareStmt = true
	}
}

func WithDryRun() DBOption {
	return func(cfg *DBConfig) {
		cfg.DryRun = true
	}
}

func WithAllowGlobalUpdate() DBOption {
	return func(cfg *DBConfig) {
		cfg.AllowGlobalUpdate = true
	}
}

func WithDebug() DBOption {
	return func(cfg *DBConfig) {
		cfg.Debug = true
	}
}
