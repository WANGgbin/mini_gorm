package tests

import (
	"errors"
	"gorm.io/gorm"
	"testing"
)

func (p *person) BeforeDelete(db *gorm.DB) error {
	return errors.New("no permission")
}

func TestHook(t *testing.T) {
	// 在执行删除操作之前，会先执行我们注册的 hook。
	// 我们只是定义了 hook 并没有执行任何注册的行为，那么 gorm 是如何感知这些 hook 的呢？
	// 通过接口断定的方式来判断。
	// if i, ok := value.(BeforeDeleteInterface); ok {
	//				db.AddError(i.BeforeDelete(tx))
	//				return true
	//			}
	if err := db.Session(&gorm.Session{DryRun: true}).Delete(&person{ID: 16}).Error; err != nil {
		panic(err)
	}
}