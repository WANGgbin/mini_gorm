package clause

import "fmt"

type LimitBuilder struct {
	num int
}

func NewLimitBuilder(num int) *LimitBuilder {
	return &LimitBuilder{
		num: num,
	}
}

func (l *LimitBuilder) Build() *Clause {
	if l.num == 0 {
		return nil
	}

	return &Clause{
		sql: fmt.Sprintf("LIMIT %d", l.num),
	}
}