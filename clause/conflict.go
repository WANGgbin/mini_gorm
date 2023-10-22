package clause

import (
	"fmt"
	"github.com/WANGgbin/mini_gorm/model"
	"github.com/WANGgbin/mini_gorm/utils"
	"strings"
)

type OnConflictOptional func(o *ConflictBuilder)

func UpdateColWithSpecificVal(colValPairs map[string]interface{}) OnConflictOptional {
	return func(o *ConflictBuilder) {
		o.toUpdateColValPairs = colValPairs
	}
}

func UpdateColsWithNewVal(cols []string) OnConflictOptional {
	return func(o *ConflictBuilder) {
		o.toUpdateColsWithNewVal = cols
	}
}

func DoNothing() OnConflictOptional {
	return func(o *ConflictBuilder) {
		o.doNothing = true
	}
}

type ConflictBuilder struct {
	doNothing              bool
	toUpdateColValPairs    map[string]interface{}
	toUpdateColsWithNewVal []string
}

func NewConflictBuilder(how OnConflictOptional) *ConflictBuilder {
	cb := new(ConflictBuilder)
	how(cb)
	return cb
}

func (c *ConflictBuilder) Build(mi *model.Info) *Clause {
	var sb strings.Builder
	var params []interface{}
	// 对于 mysql，执行 update id = id
	sb.WriteString("ON DUPLICATE KEY UPDATE ")
	if c.doNothing {
		primaryKey := utils.WrapWithBackQuote(mi.GetPrimaryColumn())
		sb.WriteString(fmt.Sprintf("%s=%s", primaryKey, primaryKey))
	}

	var parts []string
	if len(c.toUpdateColValPairs) > 0 {
		for col, val := range c.toUpdateColValPairs {
			parts = append(parts, fmt.Sprintf("%s=?", utils.WrapWithBackQuote(mi.GetColumn(col))))
			params = append(params, val)
		}
	}

	if len(c.toUpdateColsWithNewVal) > 0 {
		for _, col := range c.toUpdateColsWithNewVal {
			col = utils.WrapWithBackQuote(mi.GetColumn(col))
			parts = append(parts, fmt.Sprintf("%s=VALUES(%s)",  col, col))
		}
	}

	sb.WriteString(strings.Join(parts, ","))
	return &Clause{sql: sb.String(), params: params}
}
