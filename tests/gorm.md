描述 gorm 的实现。

# 整体架构

# statement

statement 对应一条 SQL 语句。

# clause

clause 即子句的意思。像 SELECT、WHERE 等都是子句。

statement 中的字段 `Clauses map[string]clause.Clause` 表示 statement 的所有子句。

Clause 定义如下：
```go
type Clause struct {
	Name                string // WHERE
	BeforeExpression    Expression
	AfterNameExpression Expression
	AfterExpression     Expression
	Expression          Expression
}
```
其中，最关键的成员就是 Expression，类型 Expression 是一个接口，定义如下：

```go
type Expression interface {
	Build(builder Builder) // gorm 中 Builder 其实就是 statement，每个 Expr 调用 statement 的一些 Write* 方法，来构造 SQL
}
```

当执行 SQL 的时候，就会通过执行每个 Clause 的 Build 方法来构造 SQL，然后执行。

## clause.Interface

gorm 中的每个具体的 clause 实现了该接口，我们看看该接口的定义：

```go
type Interface interface {
	Name() string // clause 的名字，比如：WHERE
	Build(Builder) // SQL 
	MergeClause(*Clause) // 在 db 的整个链式调用中，一个 clause 可能出现多次，MergeClause() 表示如何将当前的 clause 合并到 statement 中已有的 Clause 中。
}

// MergeClause 的调用如下
// 在很多 chain method 中都会调用该方法
func (stmt *Statement) AddClause(v clause.Interface) {
	if optimizer, ok := v.(StatementModifier); ok {
		optimizer.ModifyStatement(stmt)
	} else {
		name := v.Name()
        // 获取当前的 Clause
		c := stmt.Clauses[name]
		c.Name = name
        将 v 合并到 Clause 中
		v.MergeClause(&c)
		stmt.Clauses[name] = c
	}
}
```

站在 statement 角度看，statement 不关心 clause.Interface，statement 只关心每一类 clause 的最终结果 Clause。Clause 是 clause.Interface 的一个上层概念，由一个或者多个同类的 clause.Interface Merge 得到。

## 构造 SQL

既然一个 statement 有了 map[string]clause.Clause，那么是如何根据 该成员构造 SQL 的呢？不同的 DB 操作，需要的 clause 不同，且 clause 也是有顺序的。

关键在于两点：

- 在 callbacks/callbacks.go 中定义了每种 db 操作所涉及的 clause

定义如下：
```go
var (
	createClauses = []string{"INSERT", "VALUES", "ON CONFLICT"}
	queryClauses  = []string{"SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "LIMIT", "FOR"}
	updateClauses = []string{"UPDATE", "SET", "WHERE"}
	deleteClauses = []string{"DELETE", "FROM", "WHERE"}
)
```

- 通过调用 stmt.Build 方法完成 SQL 构造

```go
// clauses 就是上述成员之一
func (stmt *Statement) Build(clauses ...string) {
	var firstClauseWritten bool

	for _, name := range clauses {
		if c, ok := stmt.Clauses[name]; ok {
			if firstClauseWritten {
				stmt.WriteByte(' ')
			}

			firstClauseWritten = true
			if b, ok := stmt.DB.ClauseBuilders[name]; ok {
				b(c, stmt)
			} else {
				c.Build(stmt) // 调用每个 Clause.Build() 完成 SQL 构造
			}
		}
	}
}
```

# callback

gorm 中 sql 的执行，是通过 callback 的方式实现的。

怎么理解 callback 呢？实际上，就是每种 sql 操作，可能需要执行若干操作（不仅仅是执行 sql），包括执行 sql 前完成一些操作，执行 sql 后完成一些操作。

而这些操作就是通过 Registe callback 的方式实现的。这种实现方式的好处就是灵活，扩展性好，如果要增加新的操作，只需 registe callback 即可。无须更改 gorm 本身的代码。

callback 的原型如下：
```go
func(*DB)
```

## 注册 callback

那么 callback 是如何注册的呢？

有两种方式。

- dialector 指定

第一种是，在初始化 db 对象的时候，指定的 dialector 对象在 `Initialize` 方法中完成 callback 的调用。

我们以 mysql 这个 dialector 举例：
```go
func (dialector Dialector) Initialize(db *gorm.DB) (err error) {
	if dialector.DriverName == "" {
		dialector.DriverName = "mysql"
	}

	// register callbacks
	callbackConfig := &callbacks.Config{
		CreateClauses: CreateClauses,
		QueryClauses:  QueryClauses,
		UpdateClauses: UpdateClauses,
		DeleteClauses: DeleteClauses,
	}

    // 通过 RegisterDefaultCallbacks 完成 callback 的注册
	callbacks.RegisterDefaultCallbacks(db, callbackConfig)

	return
}

func RegisterDefaultCallbacks(db *gorm.DB, config *Config) {
	enableTransaction := func(db *gorm.DB) bool {
		return !db.SkipDefaultTransaction
	}

	createCallback := db.Callback().Create()
    // 只有默认打开事务，才会执行 begin/commit/rollback 操作
	createCallback.Match(enableTransaction).Register("gorm:begin_transaction", BeginTransaction)
	createCallback.Register("gorm:before_create", BeforeCreate)
	createCallback.Register("gorm:save_before_associations", SaveBeforeAssociations(true))
	createCallback.Register("gorm:create", Create(config))
	createCallback.Register("gorm:save_after_associations", SaveAfterAssociations(true))
	createCallback.Register("gorm:after_create", AfterCreate)
	createCallback.Match(enableTransaction).Register("gorm:commit_or_rollback_transaction", CommitOrRollbackTransaction)
	createCallback.Clauses = config.CreateClauses

    // 其他操作的 callback 注册类似，不再赘述。
}
```

- db.Callback()

第二种方式就是通过 Callback() 的方式指定 callback. 注意 callback 是有调用顺序的。

```go
// before gorm:create
db.Callback().Create().Before("gorm:create").Register("update_created_at", updateCreated)

// after gorm:create
db.Callback().Create().After("gorm:create").Register("update_created_at", updateCreated)
```

## callback 调用

callback 的调用实在 finisher_api 中调用的。比如：Find() 方法：
```go
func (db *DB) Find(dest interface{}, conds ...interface{}) (tx *DB) {
	tx = db.getInstance()
	tx.Statement.Dest = dest
	return tx.callbacks.Query().Execute(tx)
}
```

# Dialector

gorm 一个优秀的特性是跟具体的 db 实现解耦，通过 gorm 可以访问不同的 db。

解耦就是通过 Dialector 这个接口实现的。Dialector 即方言的意思，意为每一个 db 实现的具体行为。该接口的定义如下：

```go
type Dialector interface {
	Name() string
	Initialize(*DB) error
	Migrator(db *DB) Migrator
	DataTypeOf(*schema.Field) string
	DefaultValueOf(*schema.Field) clause.Expression
	BindVarTo(writer clause.Writer, stmt *Statement, v interface{})
	QuoteTo(clause.Writer, string)
	Explain(sql string, vars ...interface{}) string
}
```

# join

## 原理

gorm 中 join 是存放在 from 子句中的，通过 from 的 builder 来构造 sql.

```go
type From struct {
	Tables []Table
	Joins  []Join // 一个 from 子句，可以有多个 Join
}

type Join struct {
	Type       JoinType
	Table      Table
	ON         Where
	Using      []string
	Expression Expression
}
```

## join 跟 association 区别

join 执行一条 sql。association 在一个事务中执行多条 sql。

# association

## 原理

## Eager Loading(贪婪加载)

Eager loading 指的是将相关的数据都加载出来。这样就不需要用户一个个手动查询相关的数据。

要执行 Eager Loading，流程如下：

- model 中定义相关的 association 字段
- 通过 `Preload("fieldName")` 指定要 Eager Loading 的字段
- 可以通过 `Preload("fieldName, func(*gorm.DB)*gorm.DB)` 的方式自定义关联查询。
- 可以通过多次调用 Preload 的方式执行多个关联查询

# subQuery

其实就是单独拉起一个 db，然后作为主查询的 参数，注入。

GORM can generate subquery when using a *gorm.DB object as param

```go
db.Where("amount > (?)", db.Table("orders").Select("AVG(amount)")).Find(&orders)
// SELECT * FROM "orders" WHERE amount > (SELECT AVG(amount) FROM "orders");
```

# model

一些复杂的查询场景，返回结果并不是某个特定的跟 table 关联的 model，那么这种场景下，应该如何定义 model 呢？

## join 场景

join 场景，需要我们自定义 model.

## association 场景

可以使用 gorm/gen 生成 association model。具体参考：[associations](https://gorm.io/gen/associations.html)

# hooks

hook 是针对于特定 model 的方法，由 callback 负责调用。

hook 原型如下：
```go
func(db *gorm.DB) error
```