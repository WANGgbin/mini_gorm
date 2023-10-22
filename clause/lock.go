package clause

type LockMode string

const (
	LockModeUpdate LockMode = "UPDATE"
	LockModeShare  LockMode = "SHARE"
)

type LockBuilder struct {
	mode LockMode
}

func NewLockBuilder(mode LockMode) *LockBuilder {
	return &LockBuilder{
		mode: mode,
	}
}

func (l *LockBuilder) Build() *Clause {
	return &Clause{
		sql: "FOR" + " " + string(l.mode),
	}
}
