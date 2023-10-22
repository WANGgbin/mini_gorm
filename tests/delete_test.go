package tests

import (
	"testing"
)

// 根据主键 id 删除一条记录
func TestDeleteByPrimaryKey(t *testing.T) {
	db.DryRun = true
	// 如果主键 id 不为零值，会自动作为 where 条件
	// DELETE FROM `person` WHERE `person`.`id` = 1
	db.Delete(&person{ID: 1})

	// DELETE FROM `person`
	db.Delete(&person{})

	// DELETE FROM `person` WHERE name = 'xiaohua'
	db.Where("name = ?", "xiaohua").Delete(&person{})
}

// 软删除
func TestSoftDelete(t *testing.T) {
	// 软删除只是更新记录中的某个字段，标记为已删除，并不会真正删除记录
	// 默认情况下，查询的时候，会自动忽略软删除的记录

	// UPDATE `person` SET `deleted_at`='2023-08-22 10:20:38.739' WHERE `person`.`id` = 16 AND `person`.`deleted_at` IS NULL
	//db.Delete(&person{ID: 16})

	// 查询的时候自动忽略已经删除的记录
	// SELECT * FROM `person` WHERE `person`.`id` = 16 AND `person`.`deleted_at` IS NULL
	//var p person
	//db.Model(&person{}).Find(&p, 16)

	// 可以使用 Unscoped() 方法获取所有匹配的记录，不管是否删除
	// SELECT * FROM `person` WHERE `person`.`id` = 16
	//var p []*person
	//db.Unscoped().Model(&person{}).Find(&p, 16)

	// 同样可以使用 Unscoped() 永久删除记录
	// DELETE FROM `person` WHERE `person`.`id` = 16
	db.Unscoped().Delete(&person{ID: 16})
}