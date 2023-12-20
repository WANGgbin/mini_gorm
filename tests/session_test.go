package tests

import (
	"gorm.io/gorm"
	"testing"
)

func TestSession(t *testing.T) {
	// gorm 提供了三个 session 方法：
	// 1: Session()
	// 2: WithContext()
	// 3: Debug()
	// 其中，WithContext() 与 Debug() 本质上调用的是 Session() 方法

	// 比如对于某一系列操作，有着公共的查询条件，我们就可以先创建一个 Session，基于此 Session 进行后续操作

	newDB := db.Where("gender = ?", "male").Session(&gorm.Session{})

	var p person
	// SELECT * FROM `person` WHERE gender = 'male' AND age > 0 AND `person`.`deleted_at` IS NULL ORDER BY `person`.`id` LIMIT 1
	newDB.Where("age > ?", 0).First(&p)

	var p1 person
	// SELECT * FROM `person` WHERE gender = 'male' AND age > 18 AND `person`.`deleted_at` IS NULL ORDER BY `person`.`id` LIMIT 1
	newDB.Where("age > ?", 18).First(&p1)

}
