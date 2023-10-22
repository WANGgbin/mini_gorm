package tests

import "testing"

// Scopes 允许我们定义一些通用的逻辑，便于复用。
// 逻辑签名为 func(db *gorm.DB) *gorm.DB，可以通过闭包的方式传递额外的参数。
func TestScopes(t *testing.T) {
	db.Scopes()
}