package clause

import (
	"fmt"
	"strings"
)

type SelectBuilder struct {
	columns []string
}

func NewSelectBuilder(columns []string) *SelectBuilder {
	return &SelectBuilder{
		columns: columns,
	}
}

func (s *SelectBuilder) Build() *Clause {
	return &Clause{
		sql: fmt.Sprintf("SELECT %s", strings.Join(s.columns, ", ")),
	}

}

func (s *SelectBuilder) ResetColumns(columns []string) {
	s.columns = columns
}

func (s *SelectBuilder) GetColumns() []string {
	return s.columns
}
