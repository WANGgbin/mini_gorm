package clause

import (
	"errors"
	"fmt"
	"github.com/WANGgbin/mini_gorm/model"
	"github.com/WANGgbin/mini_gorm/utils"
	"reflect"
	"strings"
)

type WhereBuilder struct {
	ct  *CondTree
	cds *Conds
}

func NewWhereBuilder() *WhereBuilder {
	return &WhereBuilder{
		ct:  new(CondTree),
		cds: new(Conds),
	}
}

func (w *WhereBuilder) Build(mi *model.Info, unscoped bool) *Clause {
	// 如果存在软删除字段，需要过滤已经被删除的行
	if sdField := mi.GetSoftDeleteTag(); sdField != nil && !unscoped {
		_ = w.AddCond(BuildCondByString(fmt.Sprintf("%s IS NULL", utils.WrapWithBackQuote(sdField.GetColumn())), CondKindWhere))
	}

	w.setCondTree(CondKindWhere)
	return w.ct.buildWhere()
}

func (w *WhereBuilder) setCondTree(kind CondKind) {
	w.ct = newCondTree(w.cds, kind)
}

func (w *WhereBuilder) AddCond(cd *Cond) error {
	return w.cds.addCond(cd)
}

func (w *WhereBuilder) GetRootCond(kind CondKind) *Cond {
	w.setCondTree(kind)
	return w.ct.getRoot()
}

type CondKind uint8

const (
	CondKindWhere CondKind = iota
	CondKindOr
	CondKindNot
)

type Cond struct {
	kind     CondKind
	children []*Cond
	// 叶子节点 以下两个参数非空
	queryWithPlaceHolder string
	// param 合法性(能否转化为 driver.Value)交给 database/sql 判断
	params []interface{}
}

func (c *Cond) buildWhere() *Clause {
	c.setQueryAndParams()
	return &Clause{
		params:             c.params,
		sqlWithPlaceHolder: "WHERE " + c.queryWithPlaceHolder,
		sql:                "WHERE " + c.queryWithPlaceHolder,
	}
}

func (c *Cond) setQueryAndParams() {
	if len(c.children) == 0 {
		if c.kind == CondKindNot {
			c.queryWithPlaceHolder = fmt.Sprintf("NOT (%s)", c.queryWithPlaceHolder)
		}
		return
	}
	var s strings.Builder

	for idx, child := range c.children {
		child.setQueryAndParams()
		if idx != 0 {
			switch child.kind {
			case CondKindOr:
				s.WriteString(" OR ")
			default:
				s.WriteString(" AND ")
			}
		}
		c.params = append(c.params, child.params...)
		if len(c.children) > 1 {
			s.WriteString(fmt.Sprintf("(%s)", child.queryWithPlaceHolder))
		} else {
			// 如果只有一个 Cond 无需 ()
			s.WriteString(child.queryWithPlaceHolder)
		}
	}
	if c.kind == CondKindNot {
		c.queryWithPlaceHolder = fmt.Sprintf("NOT (%s)", s.String())
	} else {
		c.queryWithPlaceHolder = s.String()
	}
}

func BuildCondByString(query string, kind CondKind, args ...interface{}) *Cond {
	return &Cond{
		kind:                 kind,
		queryWithPlaceHolder: query,
		params:               args,
	}
}

func BuildCondByMap(q map[string]interface{}, kind CondKind) *Cond {
	exprs := make([]string, 0, len(q))
	params := make([]interface{}, 0, len(q))
	for key, value := range q {
		params = append(params, value)
		exprs = append(exprs, fmt.Sprintf("%s = ?", key))
	}

	return &Cond{
		kind:                 kind,
		queryWithPlaceHolder: strings.Join(exprs, " AND "),
		params:               params,
	}
}

func BuildCondByStruct(obj interface{}, kind CondKind, args ...interface{}) (*Cond, error) {
	// fields 只能是 []string 或者 string
	objTyp := reflect.TypeOf(obj)
	objVal := reflect.ValueOf(obj)

	utils.Assert(objTyp.Kind() == reflect.Ptr && objTyp.Elem().Kind() == reflect.Struct, "must be a pointer to struct")
	objTyp = objTyp.Elem()
	objVal = objVal.Elem()

	fields := make([]string, 0, len(args))
	for _, arg := range args {
		switch val := arg.(type) {
		case []string:
			fields = append(fields, val...)
		case string:
			fields = append(fields, val)
		default:
			return nil, errors.New("using string or []string to specify query fields of struct")
		}
	}

	expr := make([]string, 0, objTyp.NumField())
	params := make([]interface{}, 0, objTyp.NumField())
	for _, field := range fields {
		fieldVal := objVal.FieldByName(field)
		if !fieldVal.IsValid() {
			return nil, fmt.Errorf("%s is not a valid field of struct", field)
		}
		params = append(params, fieldVal.Interface())
		expr = append(expr, fmt.Sprintf("%s = ?", field))
	}

	// 使用结构体非零字段
	if len(fields) == 0 {
		for idx := 0; idx < objTyp.NumField(); idx++ {
			fieldVal := objVal.Field(idx)
			fieldTyp := objTyp.Field(idx)

			if fieldVal.IsZero() {
				continue
			}

			params = append(params, fieldVal.Interface())
			expr = append(expr, fmt.Sprintf("%s = ?", fieldTyp.Name))
		}
	}

	return &Cond{
		kind:                 kind,
		params:               params,
		queryWithPlaceHolder: strings.Join(expr, " AND "),
	}, nil
}

type Conds struct {
	cs      []*Cond
	hasOrCd bool
}

func (cs *Conds) buildRootCond(kind CondKind) *Cond {
	parent := &Cond{
		kind:     kind,
		children: cs.cs,
	}
	return parent
}

func (cs *Conds) reset() {
	cs.cs = nil
	cs.hasOrCd = false
}

func (cs *Conds) addCond(cd *Cond) error {
	if cd.kind != CondKindOr && cs.hasOrCd {
		return errors.New("[where/not] clause must be in front of [or] clause")
	}

	if cd.kind == CondKindOr && len(cs.cs) == 0 {
		return errors.New("at least one [where/not] clause before [or] clause")
	}

	if cd.kind == CondKindOr {
		cs.hasOrCd = true
	}
	cs.cs = append(cs.cs, cd)
	return nil
}

type CondTree struct {
	root *Cond
}

func newCondTree(cds *Conds, kind CondKind) *CondTree {
	return &CondTree{
		root: cds.buildRootCond(kind),
	}
}

func (ct *CondTree) getRoot() *Cond {
	return ct.root
}

func (ct *CondTree) buildWhere() *Clause {
	return ct.root.buildWhere()
}
