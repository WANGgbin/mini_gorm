package gorm

import (
	"errors"
	"github.com/WANGgbin/mini_gorm/clause"
	"github.com/WANGgbin/mini_gorm/model"
)

func (db *DB) Model(obj interface{}) (tx *DB) {
	tx = db.new()
	if tx.stmt.mi != nil {
		return
	}
	mi, err := model.Parse(obj)
	if err != nil {
		tx.addErr(err)
		return
	}

	tx.stmt.mi = mi

	return
}

// Where 指定查询条件，可以为字符串格式，结构体格式，map 格式。
func (db *DB) Where(query interface{}, args ...interface{}) (tx *DB) {
	return db.addCond(query, clause.CondKindWhere, args...)
}

func (db *DB) Or(query interface{}, args ...interface{}) (tx *DB) {
	return db.addCond(query, clause.CondKindOr, args...)
}

func (db *DB) Not(query interface{}, args ...interface{}) (tx *DB) {
	return db.addCond(query, clause.CondKindNot, args...)
}

func (db *DB) addCond(query interface{}, kind clause.CondKind, args ...interface{}) (tx *DB) {
	tx = db.new()
	var err error
	switch q := query.(type) {
	case map[string]interface{}:
		err = tx.stmt.AddCond(clause.BuildCondByMap(q, kind))
	case string:
		err = tx.stmt.AddCond(clause.BuildCondByString(q, kind, args...))
	case *DB:
		// 条件 Group
		err = tx.stmt.AddCond(q.stmt.wb.GetRootCond(kind))
	default:
		// 为了性能考虑，只接受结构体指针，不接受结构体。
		cd, e := clause.BuildCondByStruct(query, kind, args...)
		if e != nil {
			tx.addErr(e)
			return
		}
		err = tx.stmt.AddCond(cd)
	}

	if err != nil {
		tx.addErr(err)
	}
	return
}

// Select 指定查询字段
func (db *DB) Select(args ...interface{}) (tx *DB) {
	tx = db.new()
	columns := make([]string, 0, len(args))
	for _, arg := range args {
		switch val := arg.(type) {
		case string:
			columns = append(columns, val)
		case []string:
			columns = append(columns, val...)
		default:
			tx.addErr(errors.New("arg of select must be either string or []string"))
		}
	}
	tx.stmt.SetSelectedColumns(columns)
	return
}

func (db *DB) Order(field string) (tx *DB) {
	tx = db.new()
	tx.stmt.AddOrderField(field)
	return
}

func (db *DB) Limit(num int) (tx *DB) {
	tx = db.new()
	err := tx.stmt.SetLimitNum(num)
	if err != nil {
		tx.addErr(err)
	}
	return tx
}

func (db *DB) Offset(offset int) (tx *DB) {
	tx = db.new()
	err := tx.stmt.SetOffset(offset)
	if err != nil {
		tx.addErr(err)
	}
	return tx
}

func (db *DB) Debug() (tx *DB) {
	tx = db.new()
	tx.cfg.Debug = true
	return tx
}

func (db *DB) Lock(mode clause.LockMode) (tx *DB) {
	tx = db.new()
	if err := tx.stmt.SetLock(mode); err != nil {
		tx.addErr(err)
	}
	return tx
}

func (db *DB) OnConflict(how clause.OnConflictOptional) (tx *DB) {
	tx = db.new()
	tx.stmt.OnConflict(how)
	return tx
}

func (db *DB) Unscoped() (tx *DB) {
	tx = db.new()
	tx.stmt.Unscoped()
	return tx
}