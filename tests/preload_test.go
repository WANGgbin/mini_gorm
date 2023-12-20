package tests

import (
	"encoding/json"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Order struct {
	ID       int64
	Price    float64
	PersonID int64
}

type Person struct {
	ID     int64
	Name   string
	Orders []Order
}

func TestPreload(t *testing.T) {
	var results []*Person
	// 通过第二个参数，自定义关联查询流程
	err := db.Debug().Preload(clause.Associations).Preload("Orders", func(db *gorm.DB) *gorm.DB {
		return db.Table("orders")
	}).Find(&results).Error

	if err != nil {
		t.Fatal(err)
	}

	beautyResults, _ := json.MarshalIndent(results, "", "\t")
	t.Logf("%s", string(beautyResults))
}
