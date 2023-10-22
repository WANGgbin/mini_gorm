package gorm

import (
	"database/sql"
	"github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTransaction(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
			//WithDryRun(),
		)
		convey.So(err, convey.ShouldBeNil)

		tx := db.Begin(&sql.TxOptions{
			Isolation: sql.LevelRepeatableRead,
		})

		err = tx.Debug().Model(&person{}).Where("id = ?", 2).Select("Name").Updates(&person{Name: "name v1", Gender: "female"}).err
		convey.So(err, convey.ShouldBeNil)
		err = tx.Debug().Model(&person{}).Where("id = ?", 2).Select("Name").Updates(&person{Name: "name v2", Gender: "female"}).err
		convey.So(err, convey.ShouldBeNil)
		tx.Commit()
	})
}

func TestSessionTransaction(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
		)
		convey.So(err, convey.ShouldBeNil)

		tx := db.Session(&Session{DryRun: true}).Begin(nil)

		err = tx.Debug().Model(&person{}).Where("id = ?", 2).Select("Name").Updates(&person{Name: "name v1", Gender: "female"}).err
		convey.So(err, convey.ShouldBeNil)
		err = tx.Debug().Model(&person{}).Where("id = ?", 2).Select("Name").Updates(&person{Name: "name v2", Gender: "female"}).err
		convey.So(err, convey.ShouldBeNil)
		tx.Commit()
	})
}
