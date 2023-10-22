package clause

import "fmt"

type FromBuilder struct {
	table string
}

func NewFromBuilder(table string) *FromBuilder {
	return &FromBuilder{table: table}
}

func (f *FromBuilder) Build() *Clause {
	return &Clause{
		sql: fmt.Sprintf("FROM `%s`", f.table),
	}
}
