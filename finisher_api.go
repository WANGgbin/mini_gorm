package gorm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// First 根据主键正排，取第一个数据
// order by ID limit 1
func (db *DB) First(target interface{}) (instance *DB) {
	// 设置 model 信息，如果通过 Model 已经设置过，则忽略
	instance = db.Model(target)
	if instance.err != nil {
		return
	}

	if instance.parseHooks(target).hks.SetHooksOnQuery() {
		return instance.innerTransaction(buildQueryTransaction(target, instance), nil)
	}

	instance.doFirst(target)
	return
}

func buildQueryTransaction(obj interface{}, db *DB) func(*DB) error {
	return func(tx *DB) error {
		if bc := db.hks.GetBeforeQueryHook(); bc != nil {
			if err := bc.BeforeQuery(tx); err != nil {
				return err
			}
		}

		// 使用 sql.tx 执行
		db.setByTx(tx).doFirst(obj)
		if db.err != nil {
			return db.err
		}

		if ac := db.hks.GetAfterQueryHook(); ac != nil {
			if err := ac.AfterQuery(tx); err != nil {
				return err
			}
		}

		return nil
	}
}

func (db *DB) doFirst(target interface{}) {
	db.stmt.SetColumnsToSelect(db.stmt.mi.GetColumns())
	values, err := db.stmt.GetValuesToScan(target)
	if err != nil {
		db.addErr(err)
		return
	}
	db.stmt.AddOrderField(db.stmt.mi.GetPrimaryColumn())
	if err := db.stmt.SetLimitNum(1); err != nil {
		db.addErr(err)
		return
	}

	db.queryRow(values...)
	return
}

func (db *DB) queryRow(values ...interface{}) {
	result, err := db.doExecute(ExecModeQueryRow)
	if err != nil {
		db.addErr(err)
		return
	}

	if result == nil {
		return
	}

	if err := result.(*sql.Row).Scan(values...); err != nil {
		db.addErr(err)
	}
}

func (db *DB) Count(target interface{}, distinct bool, columns ...string) (tx *DB) {
	tx = db.new()
	tx.stmt.Count(distinct, columns...)
	tx.queryRow(target)
	return
}

// Create 执行后需要设置主键 id
func (db *DB) Create(obj interface{}) (tx *DB) {
	tx = db.new()
	tx.Model(obj)
	if tx.err != nil {
		return tx
	}
	if tx.parseHooks(obj).hks.SetHooksOnCreate() {
		// 将 hook 和 create 放在一个事务中执行
		return tx.innerTransaction(buildCreateTransaction(obj, tx), nil)
	}

	tx.doCreate(obj)
	return tx
}

func buildCreateTransaction(obj interface{}, db *DB) func(*DB) error {
	return func(tx *DB) error {
		if bc := db.hks.GetBeforeCreateHook(); bc != nil {
			if err := bc.BeforeCreate(tx); err != nil {
				return err
			}
		}

		// 使用 sql.tx 执行
		db.setByTx(tx).doCreate(obj)
		if db.err != nil {
			return db.err
		}

		if ac := db.hks.GetAfterCreateHook(); ac != nil {
			if err := ac.AfterCreate(tx); err != nil {
				return err
			}
		}

		return nil
	}
}

func (db *DB) doCreate(obj interface{}) {
	db.stmt.SetColumnsToInsert()
	values, err := db.stmt.GetValuesToInsert(obj)
	if err != nil {
		db.addErr(err)
		return
	}
	db.stmt.SetInsertValues(values)

	db.exec()
	if db.isError() {
		return
	}
	if !db.cfg.DryRun {
		db.SetPrimaryKey(obj, db.result.lastInsertID)
	}
}

// Update 更新单列
func (db *DB) Update(field string, val interface{}) (tx *DB) {
	tx = db.new()
	if err := tx.stmt.Update(map[string]interface{}{field: val}); err != nil {
		tx.addErr(err)
		return tx
	}
	tx.exec()
	return tx
}

// Updates 更新多列，通过结构体或者 map 更新。结构体更新时，忽略零值字段
func (db *DB) Updates(src interface{}) (instance *DB) {
	instance = db.new()

	if instance.parseHooks(src).hks.SetHooksOnUpdate() {
		return instance.innerTransaction(buildUpdateTransaction(src, instance), nil)
	}

	instance.doUpdate(src)
	return instance
}

func buildUpdateTransaction(obj interface{}, db *DB) func(*DB) error {
	return func(tx *DB) error {
		if bc := db.hks.GetBeforeUpdateHook(); bc != nil {
			if err := bc.BeforeUpdate(tx); err != nil {
				return err
			}
		}

		// 使用 sql.tx 执行
		db.setByTx(tx).doUpdate(obj)
		if db.err != nil {
			return db.err
		}

		if ac := db.hks.GetAfterUpdateHook(); ac != nil {
			if err := ac.AfterUpdate(tx); err != nil {
				return err
			}
		}

		return nil
	}
}

func (db *DB) doUpdate(src interface{}) {
	if err := db.stmt.Update(src); err != nil {
		db.addErr(err)
		return
	}
	db.exec()
}

// Delete 批量删除
func (db *DB) Delete(src interface{}) (instance *DB) {
	instance = db.new()
	instance.Model(src)

	if instance.parseHooks(src).hks.SetHooksOnDelete() {
		return instance.innerTransaction(buildDeleteTransaction(src, instance), nil)
	}
	instance.doDelete(src)
	return instance
}

func buildDeleteTransaction(obj interface{}, db *DB) func(*DB) error {
	return func(tx *DB) error {
		if bc := db.hks.GetBeforeDeleteHook(); bc != nil {
			if err := bc.BeforeDelete(tx); err != nil {
				return err
			}
		}

		// 使用 sql.tx 执行
		db.setByTx(tx).doDelete(obj)
		if db.err != nil {
			return db.err
		}

		if ac := db.hks.GetAfterDeleteHook(); ac != nil {
			if err := ac.AfterDelete(tx); err != nil {
				return err
			}
		}

		return nil
	}
}

func (db *DB) doDelete(src interface{}) {
	db.buildWhereClauseByPrimaryKey(src)
	db.stmt.newDeleteBuilder()
	db.exec()
}

func (db *DB) buildWhereClauseByPrimaryKey(src interface{}) {
	refVal := reflect.ValueOf(src)
	if refVal.Kind() == reflect.Slice {
		primaryVals := make([]interface{}, 0, refVal.Len())
		quesionMark := make([]string, 0, refVal.Len())
		for idx := 0; idx < refVal.Len(); idx++ {
			primaryVal := refVal.Index(idx).Elem().FieldByName(db.stmt.mi.GetPrimaryField())
			if !primaryVal.IsZero() {
				primaryVals = append(primaryVals, primaryVal.Interface())
				quesionMark = append(quesionMark, "?")
			}
		}
		if len(primaryVals) > 0 {
			db.Where(fmt.Sprintf("%s in (%s)", db.stmt.mi.GetPrimaryColumn(), strings.Join(quesionMark, ",")), primaryVals...)
		}
		return
	}

	// 如果主键不是零值
	primaryVal := reflect.ValueOf(src).Elem().FieldByName(db.stmt.mi.GetPrimaryField())
	if !primaryVal.IsZero() {
		db.Where(fmt.Sprintf("%s=?", db.stmt.mi.GetPrimaryColumn()), primaryVal.Interface())
	}
	return
}

func (db *DB) exec() {
	result, err := db.doExecute(ExecModeExec)
	if err != nil {
		db.addErr(err)
		return
	}
	if result == nil {
		return
	}

	db.result = new(DBResult)
	db.result.rowsAffected, err = result.(sql.Result).RowsAffected()
	if err != nil {
		db.addErr(err)
		return
	}

	db.result.lastInsertID, err = result.(sql.Result).LastInsertId()
	if err != nil {
		db.addErr(err)
	}
}

func (db *DB) doExecute(em ExecMode) (interface{}, error) {
	if err := db.stmt.buildSQL(); err != nil {
		return nil, err
	}

	if !db.toExecute() {
		return nil, nil
	}

	if err := db.setSqlExecutor(); err != nil {
		return nil, err
	}

	switch em {
	case ExecModeQueryRow:
		return db.executor.QueryRowContext(db.stmt.ctx, db.stmt.query, db.stmt.params...), nil
	case ExecModeQuery:
		return db.executor.QueryContext(db.stmt.ctx, db.stmt.query, db.stmt.params...)
	case ExecModeExec:
		return db.executor.ExecContext(db.stmt.ctx, db.stmt.query, db.stmt.params...)
	default:
		return nil, fmt.Errorf("unknown exec mode: %d", em)
	}
}

func (db *DB) setSqlExecutor() error {
	if db.isSetPrepareStmt() {
		executor, err := NewStmtExecutor(db)
		if err != nil {
			return err
		}
		db.executor = executor
	}
	return nil
}

func (db *DB) toExecute() bool {
	if db.isError() {
		return false
	}

	select {
	case <-db.stmt.ctx.Done():
		db.addErr(db.stmt.ctx.Err())
		return false
	default:
	}

	if db.cfg.DryRun {
		return false
	}
	return true
}

func (db *DB) SetPrimaryKey(target interface{}, value int64) {
	// 只有 autoIncrement 主键才设置
	if !db.stmt.mi.ToSetPrimaryKey() {
		return
	}

	refVal := reflect.ValueOf(target)
	if reflect.TypeOf(target).Kind() == reflect.Slice {
		for idx := 0; idx < refVal.Len(); idx++ {
			elem := refVal.Index(idx)
			db.SetPrimaryKey(elem.Interface(), db.result.lastInsertID+int64(idx))
		}
		return
	}

	primaryValue := refVal.Elem().FieldByName(db.stmt.mi.GetPrimaryField())
	switch primaryValue.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		primaryValue.SetUint(uint64(value))
	default:
		primaryValue.SetInt(value)
	}
}
