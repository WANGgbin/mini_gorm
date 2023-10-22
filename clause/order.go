package clause

import (
	"fmt"
	"strings"
)

type OrderBuilder struct {
	fields []string
}

func NewOrderBuilder() *OrderBuilder {
	return new(OrderBuilder)
}

func (o *OrderBuilder) Build() *Clause {
	if len(o.fields) == 0 {
		return nil
	}

	return &Clause{
		sql: fmt.Sprintf("ORDER BY %s", strings.Join(o.fields, ", ")),
	}
}

func (o *OrderBuilder) AddOrderField(field string) {
	o.fields = append(o.fields, field)
}
