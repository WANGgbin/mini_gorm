package tests

import (
	"testing"
)

// 检索单个对象
func TestQueryOneRecord(t *testing.T) {
	// 获取第一条记录(主键升序)
	var p person
	//db.First(&p)
	//t.Logf("%#v\n", p)

	// 获取一条记录，没有指定排序字段
	//db.Take(&p)
	//t.Logf("%#v\n", p)

	// 获取第一条记录(主键降序)
	type obj struct {
		name string
	}
	var o obj
	db.Model(&person{}).Last(&o)
	t.Logf("%#v\n", p)
}

// 根据主键 id 进行检索
func TestQueryByPrimaryKey(t *testing.T) {
	// 当目标对象有一个主键值，使用该主键值进行查询
	p := person{
		ID: 2,
	}

	// SELECT * FROM person WHERE id = 2;
	//db.First(&p)
	//t.Logf("%#v\n", p)

	// 同样可以通过下面方式根据主键查询
	//db.Model(&person{ID: 2}).First(&p)
	//t.Logf("%#v\n", p)

	// p 中的 id 会覆盖 model 中的 id，因此查到的结果还是 id = 2 对应的记录
	db.Model(person{ID: 3}).First(&p)
	t.Logf("%#v\n", p)
}

// 查询全部数据
func TestFind(t *testing.T) {
	//var ps []*person
	//
	//// SELECT * FROM person;
	//db.Find(&ps)
	//for _, p := range ps {
	//	t.Logf("%#v", p)
	//}

	// 如果查询的条数很多，如果一次性 load 到内存中会很占内存，为此，我们可以通过 Rows 方法一条条获取记录
	rows, _ := db.Table("person").Rows()
	defer func(){
		// 一定要 Close()，否则会导致连接不可复用，进一步导致连接泄露
		_ = rows.Close()
	}()

	for rows.Next() {
		var p person
		err := db.ScanRows(rows, &p)
		if err != nil {
			t.Fatalf("err: %v", err)
		} else {
			t.Logf("%#v", p)
		}
	}
}

// 根据 where 子句进行查询
func TestQueryUsingStringWhere(t *testing.T) {
	// 通过 string 指定条件

	var p person
	var ps []*person

	// SELECT * FROM person WHERE gender = male ORDER BY id LIMIT 1;
	db.Where("gender = ?", "male").First(&p)
	t.Logf("%#v", p)

	// SELECT * FROM person WHERE gender = male;
	db.Where("gender = ?", "male").Find(&ps)

	// IN
	db.Where("id in ?", []int64{2, 3}).Find(&ps)
	//for _, p := range ps {
	//	t.Logf("%#v", p)
	//}

	// LIKE
	db.Where("name like ?", "%xiao%").Find(&ps)
	for _, p := range ps {
		t.Logf("%#v", p)
	}

	// 如果对象的主键已经被设置了，其会被作为查询条件一部分，通过 and 连接起来
	p.ID = 1
	// SELECT * FROM person WHERE id = 2 and id = 1 ORDER BY id LIMIT 1;
	result := db.Where("id = ?", 2).Find(&p)

	t.Logf("%d", result.RowsAffected) // output: 0
}

func TestQueryUsingStructOrMapWhere(t *testing.T) {
	var p person
	//var ps []*person

	// 通过结构体指定条件，会忽略零值字段。因为无法区分是指定为0值还是未指定。
	// WHERE name = "xiaoming"
	//db.Where(&person{Name: "xiaoming"}).First(&p)
	//t.Logf("%#v", p)

	// 还可以指定结构体中的特定字段作为查询条件
	// WHERE name = "xiaoming" AND age = 0;
	db.Where(&person{Name: "xiaoming"}, "name", "age").First(&p)

	// 通过 map 指定条件，所有 kv 会作为条件。
	// WHERE name = "xiaoming" AND age = 0;
	db.Where(map[string]interface{}{"name": "xiaoming", "age": 0}).First(&p)
	t.Logf("%#v", p)
}

func TestQueryUsingNot(t *testing.T) {
	var p person
	// not 使用方式跟 where 类似
	// WHERE name != "xiaoming" AND gender != "male";
	db.Debug().Not(person{Name: "xiaoming", Gender: "male"}).First(&p)
	db.Debug().Or(person{Name: "xiaoming"}).Where(map[string]interface{}{"name": "xiaoming", "age": 0}).Not(db.Where(&person{Name: "xiaoming"}, "name", "age").Or(person{Name: "xiaoming", Gender: "male"})).First(&p)
	// WHERE name NOT IN ("xiaoli");
	db.Debug().Not(map[string]interface{}{"name": []string{"xiaoli"}}).First(&p)
}

// 只查询特定字段
func TestQuerySpecField(t *testing.T) {
	var p person

	// SELECT `name`,`gender` FROM `person` ORDER BY `person`.`id` LIMIT 1
	db.Debug().Select([]string{"name", "gender"}).First(&p)
	db.Debug().Select("name", "gender").First(&p)
	db.Debug().Select("name, gender", "age").First(&p)
}

// 使用 order 排序
func TestQueryUsingOrder(t *testing.T) {
	var p person
	// 注意 First 中的 ORDER BY id 仍然生效
	// SELECT * FROM `person` WHERE `person`.`gender` = 'male' ORDER BY name,`person`.`id` LIMIT 1
	db.Debug().Where(person{Gender: "male"}).Order("name").First(&p)
	t.Logf("%#v", p)

	// 还可以通过多个 Order clause 指定 order
	// SELECT * FROM `person` WHERE `person`.`gender` = 'male' AND `person`.`id` = 14 ORDER BY name desc,age,`person`.`id` LIMIT 1
	db.Debug().Where(person{Gender: "male"}).Order("name desc").Order("age").First(&p)
}

// 使用 limit && offset 子句查询
func TestQueryUsingLimitAndOffset(t *testing.T) {
	var ps []*person

	db.Offset(3).Limit(3).Find(&ps)
	for _, p := range ps {
		t.Logf("%#v", p)
	}
}

// 使用 distinct
func TestDistinct(t *testing.T) {
	type result struct {
		Name string
	}

	var rs []*result
	// 需要通过 table 指定表名，否则会通过 Find(&rs) 解析表名为：result
	db.Table("person").Distinct("name").Find(&rs)
	for _, r := range rs {
		t.Logf("%#v", r)
	}
}

// 使用 group by and having
func TestQueryUsingGroupByAndHaving(t *testing.T) {
	type result struct {
		Name string
		Total int64
	}

	var rs []*result
	// SELECT `name`,count(*) as total FROM `person` GROUP BY `name`
	//db.Model(&person{}).Select([]string{"name", "count(*) as total"}).Group("name").Find(&rs)
	//for _, r := range rs {
	//	t.Logf("%#v", r)
	//}

	// SELECT name, count(*) as total FROM `person` GROUP BY `name` HAVING count(*) > 1
	db.Table("person").Select("name, count(*) as total").Group("name").Having("count(*) > ?", 1).Find(&rs)
	for _, r := range rs {
		t.Logf("%#v", r)
	}
}
