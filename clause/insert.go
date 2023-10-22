package clause

import (
	"fmt"
	"strings"
)

type InsertBuilder struct {
	columns []string
}

func NewInsertBuilder(columns []string) *InsertBuilder {
	return &InsertBuilder{
		columns: columns,
	}
}

// Build INSERT INTO table_name (col1, col2, ...) VALUES(val1, val2), (,..,)
func (i *InsertBuilder) Build(table string) *Clause {
	return &Clause{
		sql: fmt.Sprintf("INSERT INTO `%s` (%s)", table, strings.Join(i.columns, ", ")),
	}
}

func (i *InsertBuilder) ResetColumns(columns []string) {
	i.columns = columns
}