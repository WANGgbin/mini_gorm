package tests

import (
	"testing"
	"time"
)

// save 会保存所有字段
func TestSave(t *testing.T) {
	// 如果保存的对象没有指定主键，则 Create 一条记录，负责更新记录。

	var p person
	db.First(&p)

	p.Age += 1
	// UPDATE `person` SET `id`=1,`name`='xiaoming',`gender`='male',`age`=2,`secret`='',`is_alive`=true,`born_time`='2023-08-14 12:20:38' WHERE `id` = 1
	db.Save(p)

	p1 := &person {
		Name: "huahua",
		Gender: "male",
		Age: 10,
		BornTime: time.Now(),
	}
	// INSERT INTO `person` (`name`,`gender`,`age`,`secret`,`is_alive`,`born_time`) VALUES ('huahua','male',10,'',false,'2023-08-21 08:58:35.664')
	db.Save(p1)
}

// Update 更新单个列
func TestUpdate(t *testing.T) {
	// 更新需要加 where clause，负责会导致全局更新，在没有打开全局更新开关的前提下，会报错：ErrMissingWhereClause
	var p person
	db.First(&p)

	// 如果 model 的参数包含主键 id，也会被用于构建条件从句
	//  UPDATE `person` SET `age`=18 WHERE `id` = 1
	db.Model(&p).Update("age", 18)
}

// Updates 更新多个列
func TestUpdates(t *testing.T) {
	// 可以使用结构体或者 map 来更新字段，对于结构体会忽略 0 值字段
	var p person
	db.First(&p)

	// UPDATE `person` SET `age`=0 WHERE `id` = 1
	db.Model(&p).Updates(map[string]interface{}{"age": 0})

	// 对于结构体，同样可以使用 Select() 选择要更新的字段
	p.Age = 0
	p.IsAlive = false
	// UPDATE `person` SET `age`=0,`is_alive`=false WHERE `id` = 1
	db.Model(&p).Select("age", "is_alive").Updates(p)
}

// 全局更新
func TestGlobalUpdate(t *testing.T) {
	// 学习完 session 后再来看这部分
}

// 如果要跳过 hook 并且不想自动更新 更新时间字段，可以使用 UpdateColumn/UpdateColumns 方法
// 使用方式跟 Update/Updates 方法类似，不再赘述
func TestUpdateColumn(t *testing.T) {}


