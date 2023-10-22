package gorm

import (
	"context"
	"github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestDB_WithContext(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
			WithDryRun(),
		)
		convey.So(err, convey.ShouldBeNil)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		db = db.WithContext(ctx)
		err = db.Debug().Model(&person{}).Where("id = ?", 1).Updates(map[string]interface{}{"Secret": "private data"}).err
		convey.So(err, convey.ShouldBeNil)

		time.Sleep(2*time.Second)
		err = db.Debug().Unscoped().Model(&person{}).Select("Name").Where("id = ?", 1).Updates(&person{Name: "new name", Gender: "female"}).err
		convey.So(err, convey.ShouldNotBeNil)
		t.Log(err)
	})
}

func TestDB_Session(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
			WithDryRun(),
		)
		convey.So(err, convey.ShouldBeNil)

		tx := db.Where("id = ?", 1).Session(&Session{})
		err = tx.Debug().Model(&person{}).Updates(map[string]interface{}{"Secret": "private data"}).err
		convey.So(err, convey.ShouldBeNil)

		err = tx.Debug().Unscoped().Model(&person{}).Select("Name").Updates(&person{Name: "new name", Gender: "female"}).err
		convey.So(err, convey.ShouldBeNil)

		tx1 := db.Select("Name").Where(map[string]interface{}{"Gender": "female"}).Session(&Session{})
		err = tx1.Debug().Model(&person{}).Updates(map[string]interface{}{"Name": "private data", "Gender": "female"}).err
		convey.So(err, convey.ShouldBeNil)

		err = tx1.Debug().Unscoped().Model(&person{}).Updates(&person{Name: "new name", Gender: "female"}).err
		convey.So(err, convey.ShouldBeNil)
	})
}