package gorm

import (
	"github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func (p *person) AfterCreate(tx *DB) error {
	// just update
	return tx.Model(&person{}).Where(&person{ID: p.ID}).Select("ID").Updates(p).err
}

func TestHooks(t *testing.T) {
	convey.Convey("", t, func() {
		dsn := "test:123456@tcp(127.0.0.1:3306)/world?charset=utf8mb4&loc=Local&parseTime=true"
		db, err := Open(
			"mini_mysql", dsn,
			WithPrepareStmt(),
			//WithDryRun(),
			WithDebug(),
		)
		convey.So(err, convey.ShouldBeNil)

		// 单记录插入
		p := &person{
			Name:     "abc4",
			Gender:   "male",
			BornTime: time.Now(),
		}
		// FixMe: 进行了两次初始化
		if err := db.Create(p).err; err != nil {
			t.Fatalf("err: %v", err)
		}
		t.Logf("%v", *p)
	})
}