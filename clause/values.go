package clause

import (
	"fmt"
	"strings"
)

type ValueBuilder struct {
	cntOfCols int
	values    []interface{}
}

func NewValueBuilder(cntOfCols int, values []interface{}) *ValueBuilder {
	return &ValueBuilder{
		cntOfCols: cntOfCols,
		values:    values,
	}
}

// Build VALUES (?, ?), (?, ?) val1, val2, val3, val4
func (v *ValueBuilder) Build() *Clause {
	row := make([]string, 0, v.cntOfCols)
	for idx := 0; idx < v.cntOfCols; idx++ {
		row = append(row, "?")
	}

	cntOfRow := len(v.values) / v.cntOfCols
	rows := make([]string, 0, cntOfRow)

	for idx := 0; idx < cntOfRow; idx++ {
		rows = append(rows, fmt.Sprintf("(%s)", strings.Join(row, ", ")))
	}
	return &Clause{
		params:             v.values,
		sqlWithPlaceHolder: fmt.Sprintf("VALUES %s", strings.Join(rows, ", ")),
		sql:                fmt.Sprintf("VALUES %s", strings.Join(rows, ", ")),
	}
}
