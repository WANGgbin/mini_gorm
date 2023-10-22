package tests

import (
	"gorm.io/gorm"
	"testing"
)

func TestTransaction(t *testing.T) {
	// 使用 db.Transaction() 在出错的时候，无需手动 Rollback()，成功的时候
	// 也无需手动 Commit()。
	db.Transaction(func(tx *gorm.DB) error {
		// 通过 tx 执行一些 db 错误
		return nil
	})
}

func TestManualTransaction(t *testing.T) {
	// 使用 db.Begin() 开启事务
	session := db.Where("id = ?", 2).Session(&gorm.Session{DryRun: true})
	tx := session.Begin()
	var p person
	tx.First(&p)
	tx.Model(&person{}).Update("gender", "male")
	// 回滚事务
	//tx.Rollback()

	// 提交事务
	tx.Commit()

	tx1 := session.Begin()
	var p1 person
	tx1.First(&p1)
	tx1.Model(&person{}).Update("gender", "male")

	// 提交事务
	tx1.Commit()
}

