package gorm

import (
	"context"
	"errors"
	"fmt"
	"github.com/WANGgbin/mini_gorm/clause"
	error2 "github.com/WANGgbin/mini_gorm/error"
	"github.com/WANGgbin/mini_gorm/model"
	"github.com/WANGgbin/mini_gorm/utils"
	"reflect"
	"strings"
)

type statement struct {
	ctx context.Context
	mi  *model.Info
	// 支持 where group，本质是个树结构
	sb        *clause.SelectBuilder
	fb        *clause.FromBuilder
	wb        *clause.WhereBuilder
	ob        *clause.OrderBuilder
	lb        *clause.LimitBuilder
	offb      *clause.OffsetBuilder
	lockb     *clause.LockBuilder
	ib        *clause.InsertBuilder
	vb        *clause.ValueBuilder
	conflictB *clause.ConflictBuilder
	ub        *clause.UpdateBuilder
	db        *clause.DeleteBuilder
	css       [clause.Num]*clause.Clause

	// cfg
	raiseErrRecordNotFound bool
	debug                  bool
	// 不使用软删除
	unscoped bool

	selectedFields []string // select 对应的列
	query          string
	params         []interface{}
	tx             *DB
}

func newStmt(db *DB) *statement {
	return &statement{tx: db, ctx: context.Background()}
}

func (s *statement) clone(newDb *DB) *statement {
	if s == nil {
		return nil
	}

	return &statement{
		ctx:       s.ctx,
		mi:        s.mi,
		sb:        s.sb,
		fb:        s.fb,
		wb:        s.wb,
		ob:        s.ob,
		lb:        s.lb,
		offb:      s.offb,
		lockb:     s.lockb,
		ib:        s.ib,
		vb:        s.vb,
		conflictB: s.conflictB,
		ub:        s.ub,
		db:        s.db,

		tx:             newDb,
		selectedFields: s.selectedFields,
	}
}

// buildSQL 构建待执行 SQL
func (s *statement) buildSQL() error {
	if err := s.setAndValidateClause(); err != nil {
		return err
	}

	clauses := make([]string, 0, len(s.css))
	for _, cs := range s.css {
		if cs == nil {
			continue
		}
		clauses = append(clauses, cs.GetContent())
		s.params = append(s.params, cs.GetParams()...)
	}

	s.query = strings.Join(clauses, " ")
	if s.tx.cfg.Debug {
		fmt.Printf("SQL: %s\n", s.query)
	}
	return nil
}

func (s *statement) setAndValidateClause() error {
	s.setClauses()
	return s.validateClause()
}

func (s *statement) setClauses() {
	s.setSelectClause().
		setFromClause().
		setWhereClause().
		setOrderClause().
		setLimitClause().
		setLockClause().
		setInsertClause().
		setValuesClause().
		setConflictClause().
		setUpdateClause().
		setDeleteClause()
}

// validateClause 校验 clause，比如 update 必须要有对应的 where
func (s *statement) validateClause() error {
	if s.css[clause.KindUpdate] != nil {
		if !s.tx.cfg.AllowGlobalUpdate && s.css[clause.KindWhere] == nil {
			return error2.ErrMissingWhereClause
		}
	}

	if s.css[clause.KindDelete] != nil {
		if !s.tx.cfg.AllowGlobalDelete && s.css[clause.KindWhere] == nil {
			return error2.ErrMissingWhereClause
		}
	}

	// 添加其他校验
	return nil
}

func (s *statement) setSelectClause() *statement {
	if s.sb == nil {
		return s
	}
	return s.setClause(clause.KindSelect, s.sb.Build())
}

func (s *statement) setFromClause() *statement {
	if s.fb == nil {
		return s
	}
	return s.setClause(clause.KindFrom, s.fb.Build())
}

func (s *statement) setWhereClause() *statement {
	if s.wb == nil {
		return s
	}
	return s.setClause(clause.KindWhere, s.wb.Build(s.mi, s.unscoped))
}

func (s *statement) setOrderClause() *statement {
	if s.ob == nil {
		return s
	}
	return s.setClause(clause.KindOrder, s.ob.Build())
}

func (s *statement) setLimitClause() *statement {
	if s.lb == nil {
		return s
	}
	return s.setClause(clause.KindLimit, s.lb.Build())
}

func (s *statement) setLockClause() *statement {
	if s.lockb == nil {
		return s
	}

	return s.setClause(clause.KindLock, s.lockb.Build())
}

func (s *statement) setInsertClause() *statement {
	if s.ib == nil {
		return s
	}

	return s.setClause(clause.KindInsert, s.ib.Build(s.mi.GetTableName()))
}

func (s *statement) setValuesClause() *statement {
	if s.vb == nil {
		return s
	}

	return s.setClause(clause.KindValues, s.vb.Build())
}

func (s *statement) setConflictClause() *statement {
	if s.conflictB == nil {
		return s
	}

	return s.setClause(clause.KindConflict, s.conflictB.Build(s.mi))
}

func (s *statement) setUpdateClause() *statement {
	if s.ub == nil {
		return s
	}

	return s.setClause(clause.KindUpdate, s.ub.Build(s.mi))
}

func (s *statement) setDeleteClause() *statement {
	if s.db == nil {
		return s
	}
	return s.setClause(clause.KindDelete, s.db.Build(s.mi, s.unscoped))
}

func (s *statement) setClause(idx clause.Kind, c *clause.Clause) *statement {
	s.css[idx] = c
	return s
}

func (s *statement) AddOrderField(field string) {
	if s.ob == nil {
		s.ob = clause.NewOrderBuilder()
	}

	s.ob.AddOrderField(field)
}

func (s *statement) SetLimitNum(num int) error {
	if s.lb == nil {
		s.lb = clause.NewLimitBuilder(num)
		return nil
	}

	return errors.New("reset limit clause")
}

func (s *statement) SetOffset(offset int) error {
	if s.offb == nil {
		s.offb = clause.NewOffsetBuilder(offset)
	}

	return errors.New("reset offset clause")
}

func (s *statement) Count(distinct bool, columns ...string) {
	toCount := "*"
	if len(columns) != 0 {
		toCount = strings.Join(columns, ", ")
	}
	item := fmt.Sprintf("COUNT(%s)", toCount)
	if distinct {
		item = fmt.Sprintf("COUNT(%s %s)", "DISTINCT", toCount)
	}
	s.SetColumnsToSelect([]string{item})
}

func (s *statement) SetSelectedColumns(columns []string) {
	s.selectedFields = columns
}

func (s *statement) SetColumnsToSelect(cols []string) {
	if s.selectedFields == nil {
		s.selectedFields = cols
	}

	colsToSelect := make([]string, 0, len(s.selectedFields))
	for idx, col := range s.selectedFields {
		// 如果是表的列，则 format(`table`.`column`)
		realCol := s.mi.GetColumn(col)
		if realCol != "" {
			s.selectedFields[idx] = realCol
			colsToSelect = append(colsToSelect, fmt.Sprintf("`%s`.`%s`", s.mi.GetTableName(), realCol))
		} else {
			colsToSelect = append(colsToSelect, col)
		}
	}
	s.sb = clause.NewSelectBuilder(colsToSelect)
	s.fb = clause.NewFromBuilder(s.mi.GetTableName())
}

func (s *statement) SetColumnsToInsert() {
	if s.selectedFields == nil {
		s.selectedFields = s.mi.GetFieldNames()
	}

	colsToInsert := make([]string, 0, len(s.selectedFields))
	for _, col := range s.selectedFields {
		// 如果是表的列，则 format(`column`)
		realCol := s.mi.GetColumn(col)
		if realCol != "" {
			colsToInsert = append(colsToInsert, fmt.Sprintf("`%s`", realCol))
		} else {
			colsToInsert = append(colsToInsert, col)
		}
	}
	s.ib = clause.NewInsertBuilder(colsToInsert)
}

func (s *statement) AddCond(cd *clause.Cond) error {
	if s.wb == nil {
		s.wb = clause.NewWhereBuilder()
	}

	return s.wb.AddCond(cd)
}

func (s *statement) GetRootCond(kind clause.CondKind) *clause.Cond {
	return s.wb.GetRootCond(kind)
}

func (s *statement) GetValuesToScan(target interface{}) ([]interface{}, error) {
	selectFields, err := s.mi.GetFieldTagsByFields(s.selectedFields)
	if err != nil {
		return nil, err
	}
	refVal := reflect.ValueOf(target).Elem()
	ret := make([]interface{}, 0, len(selectFields))
	for _, field := range selectFields {
		val := refVal.FieldByName(field.GetFieldName())
		if !val.IsValid() {
			return nil, fmt.Errorf("no matched field with column `%s`", field.GetColumn())
		}
		// value -> Interface() 必须调用 Interface() 函数
		ret = append(ret, val.Addr().Interface())
	}

	return ret, nil
}

// GetValuesToInsert 获取插入的 value
func (s *statement) GetValuesToInsert(target interface{}) ([]interface{}, error) {
	refVal := reflect.ValueOf(target)
	if reflect.TypeOf(target).Kind() == reflect.Slice {
		var ret []interface{}
		for idx := 0; idx < reflect.ValueOf(target).Len(); idx++ {
			part, err := s.GetValuesToInsert(reflect.ValueOf(target).Index(idx).Interface())
			if err != nil {
				return nil, err
			}
			ret = append(ret, part...)
		}

		return ret, nil
	}

	insertFields, err := s.mi.GetFieldTagsByFields(s.selectedFields)
	if err != nil {
		return nil, err
	}
	refVal = reflect.ValueOf(target).Elem()
	ret := make([]interface{}, 0, len(insertFields))

	for _, field := range insertFields {
		val := refVal.FieldByName(field.GetFieldName())
		if !val.IsValid() {
			return nil, fmt.Errorf("no matched field with column `%s`", field.GetColumn())
		}
		i, err := field.GetValue(val)
		if err != nil {
			return nil, err
		}
		ret = append(ret, i)
	}

	return ret, nil
}

func (s *statement) SetInsertValues(values []interface{}) {
	s.vb = clause.NewValueBuilder(len(s.selectedFields), values)
}

func (s *statement) SetLock(mode clause.LockMode) error {
	if s.lockb == nil {
		s.lockb = clause.NewLockBuilder(mode)
		return nil
	}
	return errors.New("reset lock mode ")
}

func (s *statement) OnConflict(how clause.OnConflictOptional) {
	s.conflictB = clause.NewConflictBuilder(how)
}

func (s *statement) Update(src interface{}) error {
	if !s.mi.IsValidFields(s.selectedFields) {
		return error2.ErrShouldUseFieldNameToSpecifyColumn
	}

	fieldValPairs := make(map[string]interface{})
	switch v := src.(type) {
	case map[string]interface{}:
		for field := range v {
			if !s.mi.IsValidField(field) {
				return error2.ErrShouldUseFieldNameToSpecifyColumn
			}
		}
		if len(s.selectedFields) == 0 {
			fieldValPairs = v
		} else {
			for _, sc := range s.selectedFields {
				val, exist := v[sc]
				if !exist {
					// 如果 Select() 指定的列不在 map 中报错
					return fmt.Errorf("%s specified by Select does not exist in map", sc)
				}
				fieldValPairs[sc] = val
			}
		}
	default:
		if len(s.selectedFields) == 0 {
			fieldValPairs = utils.GetNoZeroFields(src)
		} else {
			for _, sc := range s.selectedFields {
				fieldValPairs[sc] = reflect.ValueOf(v).Elem().FieldByName(sc).Interface()
			}
		}
	}

	s.ub = clause.NewUpdateBuilder(fieldValPairs)
	return nil
}

func (s *statement) newDeleteBuilder() {
	s.db = clause.NewDeleteBuilder(s.mi.GetTableName())
}

func (s *statement) Unscoped() {
	s.unscoped = true
}
