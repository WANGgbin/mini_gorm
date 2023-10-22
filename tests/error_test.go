package tests

import "testing"

func TestError(t *testing.T) {
	// Gorm 的错误处理与普通的 go 代码不同，因为 Gorm 提供的是链式 api，
	// api 在出错的时候，会设置 db.Error 成员。
	// 在执行 finisher_api 的时候，会首先检查 db.Error，如果非空则直接返回。
	// 对于使用者而言，在每次调用 finisher_api 之后，都应该检查下 db.Error。
}

