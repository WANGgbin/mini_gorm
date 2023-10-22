package clause

import "fmt"

type OffsetBuilder struct {
	offset int
}

func NewOffsetBuilder(offset int) *OffsetBuilder {
	return &OffsetBuilder{
		offset: offset,
	}
}

func (o *OffsetBuilder) Build() *Clause {
	return &Clause{
		sql: fmt.Sprintf("OFFSET %d", o.offset),
	}
}