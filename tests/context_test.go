package tests

import (
	"context"
	"testing"
	"time"
)

// 描述 gorm 对 ctx 的支持，本质上在执行 SQL 之前，会判断下 ctx 是否被取消等。
// 底层是通过 sql 标准包 + 标准包实现的。

func TestCtxTimeout(t *testing.T) {
	// 对于长 SQL 查询，可以传入一个带超时的 ctx 设置超时时间

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	var p person
	db.WithContext(ctx).First(&p)
}