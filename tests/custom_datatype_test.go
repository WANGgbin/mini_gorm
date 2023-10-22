package tests

import (
	"testing"
)

// 自定义数据类型
type PersonExtra struct {
	LiveAt string `json:"live_at"`
	IsSingle string `json:"is_single"`
}

//func (e *PersonExtra) GormDataType() string {
//	return "string"
//}
//
//func (e *PersonExtra) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
//	b, err := json.Marshal(e)
//	if err != nil {
//		db.AddError(err)
//		return clause.Expr{}
//	}
//
//	return clause.Expr{
//		SQL: "?",
//		Vars: []interface{}{b},
//	}
//}

func TestCustomDataType(t *testing.T) {
	if err := db.Updates(&person{ID: 1, Extra: &PersonExtra{LiveAt: "BeiJing", IsSingle: "true"}}).Error; err != nil {
		panic(err)
	}
}

// 此外，自定义类型还可以实现 sql 标准包定义的 Scanner/Valuer 接口，实现序列/反序列化。
// json 是一种很常见的序列化方式，显然，对于使用 json 序列化方式的类型都定义 Scanner/Valuer
// 是有大量重复的。为此， gorm 引入了 serializer，字段只需要声明 tag: gorm:"serializer:json"
// 即可实现 json 格式的序列化，无需手动指定。