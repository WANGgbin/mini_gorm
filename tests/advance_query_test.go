package tests

import (
	"gorm.io/gorm/clause"
	"testing"
)

// 描述如果通过 gorm 执行一些高级查询

// 当前读
func TestForUpdate(t *testing.T) {
	var p person

	// SELECT * FROM `person` ORDER BY `person`.`id` LIMIT 1 FOR UPDATE
	db.Clauses(clause.Locking{
		Strength: "UPDATE",
	}).First(&p)

	t.Logf("%#v", p)
}

// 智能选择字段
func TestSelectFieldSmartly(t *testing.T) {
	// 如果经常需要使用 select() 来选择指定的字段，我们可以定义一个小的结构体 同时指定 QueryFields = true

	type smallPerson struct {
		Name string
		Gender string
	}

	var rs []*smallPerson
	db.QueryFields = true
	// SELECT `person`.`name`,`person`.`gender` FROM `person`
	db.Table("person").Find(&rs)
	for _, r := range rs {
		t.Logf("%#v", r)
	}
}

// 查询匹配的第一条记录，如果找不到，根据给定的条件(Attrs/Assign)初始化一个实例
func TestFirstOrInit(t *testing.T) {

	var p person
	// 因为没有查找到记录，使用 where and attrs 来初始化 p
	// p = person{Name: "non_existing", Age: 18}
	db.Where(person{Name: "non_existing"}).Attrs(person{Age: 18}).FirstOrInit(&p)
	t.Logf("%#v", p)

	// 如果查找到，直接使用数据库中的记录初始化 p
	db.Where(person{Name: "xiaoming"}).Attrs(person{Age: 1}).FirstOrInit(&p)
	t.Logf("%#v", p)

	// 无论是否查找到记录，Assign 都会初始化实例
	// p.Age = 1
	db.Where(person{Name: "xiaoming"}).Assign(person{Age: 1}).FirstOrInit(&p)
	t.Logf("%#v", p)
}

// 查找匹配的第一条记录，如果找不到，根据给定的条件创建一条新的记录
func TestFirstOrCreate(t *testing.T) {
	var p person
	// 为找到记录，则创建一条记录
	// 需要注意，其他未指定字段，则使用字段零值进行创建
	//db.Where(person{Name: "non_existing"}).Attrs(map[string]interface{}{"age": 18, "born_time": time.Now()}).FirstOrCreate(&p)
	//t.Logf("%#v", p)

	// 如果查找到，直接使用数据库中的记录初始化 p
	//db.Where(person{Name: "xiaoming"}).Attrs(person{Age: 1}).FirstOrCreate(&p)
	//t.Logf("%#v", p)

	// 无论是否查找到记录，Assign 都会初始化实例 并更新数据库
	// p.Age = 1
	// SELECT * FROM `person` WHERE `person`.`name` = 'xiaoming' ORDER BY `person`.`id` LIMIT 1
	// UPDATE `person` SET `age`=1 WHERE `person`.`name` = 'xiaoming' AND `id` = 1
	db.Where(person{Name: "xiaoming"}).Assign(person{Age: 1}).FirstOrCreate(&p)
	t.Logf("%#v", p)
}

// 使用 pluck 查询单个列
func TestPluck(t *testing.T) {
	var names []string
	db.Model(new(person)).Where("gender = ?", "male").Pluck("name", &names)
	t.Logf("%#v", names)
}

// 使用 count 获取匹配记录数
func TestCount(t *testing.T) {
	var count int64
	// SELECT count(*) FROM `person`
	db.Table("person").Count(&count)
	t.Logf("count: %d", count)

	// SELECT COUNT(DISTINCT(`name`)) FROM `person`
	db.Model(new(person)).Distinct([]string{"name"}).Count(&count)
	t.Logf("count: %d", count)
}

// 使用 group 执行复杂的条件查询
func TestUsingGroupCondition(t *testing.T) {
	var ps []*person

	// 我们可以通过在 where/or/not 等子句里面再嵌套子句来构造复杂的 sql
	// where 之间通过 AND 连接
	// SELECT * FROM `person` WHERE (name like '%xiao%' OR name like '%小%') AND (gender = 'male' AND age < 20) AND NOT name = 'xiaohua'
	db.Model(&person{}).
		Where(
			db.Where("name like ?", "%xiao%").Or("name like ?", "%小%"),
			).
		Where(
			db.Where("gender = ?", "male").Where("age < ?", 20),
			).
		Not(
			db.Where("name = ?", "xiaohua"),
			).
		Find(&ps)

	for _, p := range ps {
		t.Logf("%#v", p)
	}
}