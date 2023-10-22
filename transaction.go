package gorm

import (
	"database/sql"
)

// Begin
// 事务隔离级别事务维度设置
func (db *DB) Begin(opts *sql.TxOptions) (ret *DB) {
	return db.begin(opts, &DBCloneConfig{newStmt: false})
}

func (db *DB) begin(opts *sql.TxOptions, cloneCfg *DBCloneConfig) (ret *DB) {
	ret = db.newTx(cloneCfg)
	if !db.toExecute() {
		return ret
	}
	tx, err := ret.db.BeginTx(ret.stmt.ctx, opts)
	if err != nil {
		ret.addErr(err)
		return ret
	}
	ret.executor = NewTxExecutor(tx)
	return ret
}

func (db *DB) Commit() {
	if !db.toExecute() {
		return
	}
	if err := db.getRawSqlTx().Commit(); err != nil {
		db.addErr(err)
	}
	return
}

func (db *DB) Rollback() {
	if !db.toExecute() {
		return
	}
	if err := db.getRawSqlTx().Rollback(); err != nil {
		db.addErr(err)
	}
	return
}

func (db *DB) getRawSqlTx() *sql.Tx {
	 return db.executor.(*TxExecutorImpl).tx
}

// SavePoint
// TODO: database 并不支持 savePoint 和 RollbackTo，需要直接执行 SQL
func (db *DB) SavePoint(sp string) (tx *DB) {
	return nil
}

func (db *DB) RollbackTo(sp string) (tx *DB) {
	return nil
}

// Transaction 供外部调用的事务
func (db *DB) Transaction(ops func(db *DB) error, opts *sql.TxOptions) (tx *DB) {
	return db.transaction(ops, opts, &DBCloneConfig{
		newStmt: false,
	})
}

// innerTransaction 内部使用的事务，比如 hooks、association 场景下的事务
func (db *DB) innerTransaction(ops func(db *DB) error, opts *sql.TxOptions) (tx *DB) {
	return db.transaction(ops, opts, &DBCloneConfig{
		newStmt: true,
	})
}

// transaction 事务创建/提交/回滚交由 Transaction 完成，用户只需关心数据库操作
func (db *DB) transaction(ops func(db *DB) error, opts *sql.TxOptions, cloneCfg *DBCloneConfig) (tx *DB){
	tx = db.begin(opts, cloneCfg)
	var err error
	defer func() {
		if err != nil {
			tx.Rollback()
			tx.addErr(err)
		}
	}()

	err = ops(tx)
	if err == nil {
		// 事务 commit 错误也需要回滚
		tx.Commit()
		err = tx.err
		return tx
	}

	return
}
