gorm 相关特性。参考: [GORM](https://gorm.io/zh_CN/docs/index.html)

# 基本 CRUD

详情可以参考 tests 包下的文件。

# 预加载

需要注意的是：事务的预加载是单独的。

# 事务

手动调用 db.Begin()、db.Commit()、db.Rollback() 操作一个事务。也可以直接调用 db.Transaction() 开启一个事务。

# DryRun

全局 db 或者 session 可以设置 dry run，这样不会真正执行 sql。

# 钩子思想

钩子运行在执行 crud 的前后，执行一些特定的操作。需要注意的是：钩子操作跟真正的操作都是放在一个事务中执行的。

# join

# 子查询

# 插件

# context

可以每次操作都绑定一个 ctx，也可以先给 session 初始化一个 ctx，然后基于 session 的每个 instance 都会使用该 ctx.

# 错误处理

gorm 的一个特点是可以进行链式调用，之所以能这样做的原因是，发生错误时并不是直接返回，而是设置成员变量 db.err.

# 软删除

gorm 支持软删除，即不执行真正的删除操作，只是更新一些特定的字段标记改行记录无效。后续的 crud 都会忽略该记录。

软删除字段支持 bool/time 等类型。

# 自定义数据类型

在 model 中我们可以自定义一些数据类型，前提是我们要实现 Value 和 Scan 接口，告诉 database/sql 如何转化数据。

# session

gorm 中有三个概念：db、session、instance。

db 与 session 本质上是一样的。本质上表示的与数据库操作的配置。当我们想开启一个不同的配置的时候， 就可以通过 Session() 方法来初始化一个新的配置对象。

而 instance 就是 基于 db/session 初始化的实例，每个 instance 对应一个单独的 statement，statement 对应一个 SQL。

# associations

如何确定一个对象都有哪些 association ?

如果一个 model 的某个字段可以读写且无法确定其 datatype，则认为该字段是一个 association.

# gorm 整体架构

gorm 依赖 connpool 执行 sql。

database/sql 中的 sql.DB 是 connpool 的一种实现。

sql.DB 又依赖具体的 driver 来执行 sql。

# gorm 如何屏蔽不同的 sql 实现的

# 慢查询

# 分表(Sharding)

# gorm  vs raw sql

gorm 跟 raw sql 相比有什么优缺点呢？

gorm 优点：

- 兼容多种数据库，在使用的时候屏蔽底层的数据库，在数据库迁移的时候很方便
- 支持 prepare 模式，性能更好
- 支持更多高级特性：字段权限、hook 等

gorm 缺点：

- 有一定学习成本