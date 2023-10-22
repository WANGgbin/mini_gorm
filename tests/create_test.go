package tests

import (
	"github.com/smartystreets/goconvey/convey"
	"gorm.io/gorm/clause"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	convey.Convey("", t, func() {
		// zero 值如何处理？zero值会被认为有效值
		p := &person{
			Name:     "test",
			BornTime: time.Now(),
		}

		// 主键 id 会更新到 p 中
		// result 是 db 类型，Error 表明是否发生错误，RowsAffected: 表示影响行数
		result := db.Create(p)
		convey.So(result.Error, convey.ShouldBeNil)
		t.Logf("id: %d", p.ID)
	})
}

// TestBatchCreate 批量插入
func TestBatchCreate(t *testing.T) {
	convey.Convey("", t, func() {
		ps := []*person{
			{
				Name:     "Jane",
				//Gender:   "female",
				Age:      18,
				BornTime: time.Now(),
			},
			{
				Name:     "Ai",
				//Gender:   "male",
				Age:      20,
				Secret:   []byte("secret"),
				IsAlive:  true,
				BornTime: time.Now(),
			},
		}

		// 批量插入，gorm 会通过一条 insert 语句，完成所有记录的插入。
		// 但是当记录条数很多的时候，就会导致 insert 语句很长，有的 sql server 会限制
		// insert 语句的长度，为此，gorm 会启动一个事务，通过分批的方式，完成记录的插入。

		// 一次性插入所有记录
		// 对应 SQL:
		// INSERT INTO `person` (`name`,`gender`,`age`,`secret`,`is_alive`,`born_time`) VALUES
		// (?,?,?,?,?,?),(?,?,?,?,?,?)
		result := db.Create(ps)

		// 开启事务，分批方式插入
		// 分两次，每次插入一条记录
		db.CreateInBatches(ps, 1)

		convey.So(result.Error, convey.ShouldBeNil)

		// 通过 OK_Packet 中的 lastInsertID & affectedRows 推导出每个记录的主键 ID
		for _, p := range ps {
			t.Logf("id: %d", p.ID)
		}
	})
}

// TestCreateUseSelect 为指定的字段分配值
func TestCreateUseSelect(t *testing.T) {
	convey.Convey("", t, func() {
		p := &person{
			Name:     "xiaoqiang",
			//Gender:   "male",
			BornTime: time.Now(),
		}

		// 使用 Select 指定特定的字段
		// 对应 SQL:
		// INSERT INTO `person` (`name`,`gender`) VALUES ('xiaoqiang','male')
		result := db.Select("name", "gender").Create(p)
		// 但是一定要注意，对于不为 NULL 且没有 Default 的字段，一定要指定值，不可忽略，否则会报错:
		// Error 1364 (HY000): Field 'born_time' doesn't have a default value
		convey.So(result.Error, convey.ShouldBeNil)
		t.Logf("id: %d", p.ID)
	})
}

// TestCreateUserOmit 忽略特定的字段
func TestCreateUseOmit(t *testing.T) {
	convey.Convey("", t, func() {
		p := &person{
			Name:     "xiaoqiang",
			//Gender:   "male",
			Age:      18,
			BornTime: time.Now(),
		}

		// 使用 Omit 忽略特定的字段
		// 对应 SQL:
		// INSERT INTO `person` (`name`,`gender`,`secret`,`is_alive`,`born_time`) VALUES (?,?,?,?,?)
		result := db.Omit("age").Create(p)
		convey.So(result.Error, convey.ShouldBeNil)
		t.Logf("id: %d", p.ID)
	})
}

// TestCreateUserDflValue 使用默认值创建记录
func TestCreateUseDflValue(t *testing.T) {
	convey.Convey("", t, func() {
		// 通过 tag: default:value 可以为某个字段指定默认值
		type person struct {
			ID       uint64
			Name     string
			Gender   string `gorm:"default:female"` // 默认值: female
			Age      uint16
			Secret   []byte
			IsAlive  bool
			BornTime time.Time
		}

		// 未指定 Gender，该字段将使用默认值: female
		// 特别注意：即使字段指定了 zero value，仍然会使用默认值，
		// 为了避免这种情况，可以将字段类型定义为指针类型
		p := &person{
			Name:     "xiaoqiang",
			Age:      18,
			BornTime: time.Now(),
		}

		result := db.Create(p)
		convey.So(result.Error, convey.ShouldBeNil)
		t.Logf("id: %d", p.ID)
	})
}

// TestCreateUpsert 插入时冲突的解决方案，对于 mysql 而言，就是实现：
// INSERT INTO () VALUES() ON DUPLICATE KEY UPDATE col = values(col) 语义
func TestCreateUpsert(t *testing.T) {
	convey.Convey("", t, func() {
		p := &person{
			Name:     "huahua",
			Gender:   "male",
			Age: 10,
			BornTime: time.Now(),
		}

		// case 1: 冲突时，不进行任何操作
		//result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(p)
		//convey.So(result.Error, convey.ShouldBeNil)
		//t.Logf("id: %d", p.ID)

		// case2: 只有 id 冲突时，更新特定列为指定值
		db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"gender": "female"}),
		}).Create(p)

		// case3: 只有 id 冲突时，更新特定列为新值
		//db.Clauses(clause.OnConflict{
		//	Columns: []clause.Column{{Name: "id"}},
		//	DoUpdates: clause.AssignmentColumns([]string{"gender", "born_time"}),
		//}).Create(p)

		// case4: 只有 id 冲突时，更新所有列(除了主键)为新值
		//db.Clauses(clause.OnConflict{
		//	Columns: []clause.Column{{Name: "id"}},
		//	UpdateAll: true,
		//}).Create(p)
	})

}

// join 查询
func TestJoin(t *testing.T) {
	// 通过 Joins 来指定 join 方式以及 join 表
	type result struct {
		CityName string
		CountryName string
	}
	// SELECT city.name as city_name,country.name as country_name FROM `city` inner join country on city.countrycode = country.code WHERE country.code = 'CHN' LIMIT 10
	var rs []*result
	db.Table("city").Select([]string{"city.name as city_name", "country.name as country_name"}).Joins("inner join country on city.countrycode = country.code").Where("country.code = ?", "CHN").Limit(10).Find(&rs)
	for _, r := range rs {
		t.Logf("%#v", r)
	}
}

// 子查询
func TestSubQuery(t *testing.T) {
	// from 子查询
	// 可以在 table 方法中，使用 from 子查询。派生表必须通过 as 指定 alias
	var ps []*person
	// SELECT * FROM (SELECT * FROM `person` WHERE gender = 'male') as u WHERE name like '%xiao%'
	//db.Table("(?) as u", db.Model(&person{}).Where("gender = ?", "male")).Where("name like ?", "%xiao%").Find(&ps)
	//for _, p := range ps {
	//	t.Logf("%#v", p)
	//}

	// where 子句
	// 通过将 DB 对象作为 where 子句参数实现子查询
	// SELECT * FROM `person` WHERE age > (SELECT AVG(age) FROM `person`)
	db.Where("age > (?)", db.Model(&person{}).Select("AVG(age)")).Find(&ps)
	for _, p := range ps {
		t.Logf("%#v", p)
	}
}