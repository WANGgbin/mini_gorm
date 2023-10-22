package gorm

import (
	"errors"
	"github.com/WANGgbin/mini_gorm/clause"
	error2 "github.com/WANGgbin/mini_gorm/error"
	"github.com/WANGgbin/mini_gorm/utils"
	"github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestDB_First(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
		)
		convey.So(err, convey.ShouldBeNil)
		var p person
		// Stmt cache
		for idx := 0; idx < 2; idx++ {
			if err := db.Debug().Lock(clause.LockModeUpdate).Model(&person{}).First(&p).err; err != nil {
				t.Fatalf("err: %v", err)
			}
			t.Logf("%#v", p)
		}

		// Select
		db.Debug().Select("Name", "Gender").First(&p)
	})
}

func TestDB_Count(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
		)
		convey.So(err, convey.ShouldBeNil)
		var count int64
		if err := db.Debug().Model(&person{}).Count(&count, false).err; err != nil {
			t.Fatalf("err: %v", err)
		}
		t.Logf("%#v", count)

		if err := db.Debug().Model(&person{}).Where(map[string]interface{}{"gender": "male"}).Count(&count, true, "gender", "is_alive").err; err != nil {
			t.Fatalf("err: %v", err)
		}
		t.Logf("%#v", count)
	})
}

func TestDB_Create(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
			//WithDryRun(),
		)
		convey.So(err, convey.ShouldBeNil)

		// 单记录插入
		p := &person{
			Name:     "test",
			Gender:   "male",
			BornTime: time.Now(),
		}
		if err := db.Debug().Create(p).err; err != nil {
			t.Fatalf("err: %v", err)
		}
		t.Logf("%v", *p)
	})
}

func TestDB_BatchCreate(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
			WithDryRun(),
		)
		convey.So(err, convey.ShouldBeNil)

		// 批量插入
		ps := []*person{
			{
				Name:     "xiaoqiang",
				Gender:   "male",
				BornTime: time.Now(),
			},
			{
				Name:     "xiaoqiang",
				Gender:   "male",
				BornTime: time.Now(),
			},
		}
		if err := db.Debug().Create(ps).err; err != nil {
			t.Fatalf("err: %v", err)
		}
		t.Logf("%d, %d", ps[0].ID, ps[1].ID)
	})
}

func TestDB_CreateWithSelect(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
			WithDryRun(),
		)
		convey.So(err, convey.ShouldBeNil)

		// 插入指定字段

		p := &person{
			Name:     "xiaoqiang",
			Gender:   "male",
			BornTime: time.Now(),
		}

		if err := db.Debug().Select("Name", "gender", []string{"BornTime"}).Create(p).err; err != nil {
			t.Fatalf("err: %v", err)
		}
		t.Logf("%d", p.ID)
	})
}

func TestCreateUsingDflValue(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
		)
		convey.So(err, convey.ShouldBeNil)

		// 插入指定字段
		p := &person{
			Name:     "xiaoqiang",
			BornTime: time.Now(),
		}

		if err := db.Debug().Select("Name", "gender", []string{"BornTime", "Age"}).Create(p).err; err != nil {
			t.Fatalf("err: %v", err)
		}
		t.Logf("%d", p.ID)
	})
}

func TestCreateUpsert(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
			WithDryRun(),
		)
		convey.So(err, convey.ShouldBeNil)

		p := &person{
			ID:       1,
			Name:     "huahua",
			Gender:   "male",
			Age:      utils.Ptr2Uint16(1),
			BornTime: time.Now(),
		}

		// case 1: 冲突时，不进行任何操作
		//result := db.Debug().OnConflict(clause.DoNothing()).Create(p)
		//convey.So(result.err, convey.ShouldBeNil)

		// case2: 只有 id 冲突时，更新特定列为指定值
		//result := db.Debug().OnConflict(clause.UpdateColWithSpecificVal(map[string]interface{}{
		//	"Name": "new_name",
		//	"Age": 10,
		//})).Create(p)
		//convey.So(result.err, convey.ShouldBeNil)

		// case3: 只有 id 冲突时，更新特定列为新值
		result := db.Debug().OnConflict(clause.UpdateColsWithNewVal([]string{"Name", "Age"})).Create(p)
		convey.So(result.err, convey.ShouldBeNil)
	})
}

func TestDB_Update(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
		)
		convey.So(err, convey.ShouldBeNil)

		// 没有 where clause
		err = db.Debug().Model(&person{}).Update("Name", "xiaoqiang").err
		convey.So(err, convey.ShouldNotBeNil)
		convey.So(errors.Is(err, error2.ErrMissingWhereClause), convey.ShouldBeTrue)

		err = db.Debug().Model(&person{}).Where("id = ?", 1).Update("Name", "xiaoqiang").err
		convey.So(err, convey.ShouldBeNil)
	})
}

// 更新多列
func TestDB_Updates(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
			WithDryRun(),
		)
		convey.So(err, convey.ShouldBeNil)

		err = db.Debug().Model(&person{}).Where("id = ?", 1).Updates(map[string]interface{}{"Secret": "private data"}).err
		convey.So(err, convey.ShouldBeNil)

		err = db.Debug().Unscoped().Model(&person{}).Select("Name").Where("id = ?", 1).Updates(&person{Name: "new name", Gender: "female"}).err
		convey.So(err, convey.ShouldBeNil)
	})

}

func TestDB_Delete(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
			WithDryRun(),
		)
		convey.So(err, convey.ShouldBeNil)

		// 禁止无 where 条件的删除
		err = db.Debug().Delete(&person{}).err
		convey.So(errors.Is(err, error2.ErrMissingWhereClause), convey.ShouldBeTrue)

		err = db.Debug().Where("id = ?", 1).Delete(&person{}).err
		convey.So(err, convey.ShouldBeNil)

		// 将 obj 中的主键作为 where
		err = db.Debug().Where("name = ?", "xiaowang").Delete(&person{ID: 1}).err
		convey.So(err, convey.ShouldBeNil)

		err = db.Debug().Where("name = ?", "xiaowang").Delete([]*person{{ID: 1}, {ID: 2}}).err
		convey.So(err, convey.ShouldBeNil)

		// 软删除
		err = db.Debug().Delete(&person{ID: 1}).err
		convey.So(err, convey.ShouldBeNil)

		// 使用 Unscoped() 物理删除
		err = db.Debug().Unscoped().Delete(&person{ID: 1}).err
		convey.So(err, convey.ShouldBeNil)
	})
}