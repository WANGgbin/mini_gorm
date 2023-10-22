package clause

type Kind uint8

// Clause 根据这里的序号排序
const (
	KindInsert Kind = iota
	KindValues
	KindConflict
	KindDelete
	KindUpdate
	KindSelect
	KindFrom
	KindWhere
	KindOrder
	KindLimit
	KindOffset
	KindGroup
	KindLock
	Num
)

type Clause struct {
	params             []interface{}
	sqlWithPlaceHolder string
	sql                string
}

func (c *Clause) GetParams() []interface{} {
	return c.params
}

func (c *Clause) GetContentWithPlaceHolder() string {
	return c.sqlWithPlaceHolder
}

func (c *Clause) GetContent() string {
	return c.sql
}