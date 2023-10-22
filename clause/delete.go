package clause

import (
	"fmt"
	"github.com/WANGgbin/mini_gorm/model"
)

type DeleteBuilder struct {
	tableName string
}

func NewDeleteBuilder(tableName string) *DeleteBuilder {
	return &DeleteBuilder{tableName: tableName}
}

func (d *DeleteBuilder) Build(mi *model.Info, unscoped bool) *Clause {
	sdField := mi.GetSoftDeleteTag()
	if sdField == nil || unscoped {
		return &Clause{
			sql: fmt.Sprintf("DELETE FROM %s", d.tableName),
		}
	}
	// 如果存在软删除字段，则执行 Update 语句更新软删除字段
	// UPDATE table SET soft_delete = 1;
	return NewUpdateBuilder(map[string]interface{}{sdField.GetFieldName(): sdField.GetSoftDeleteValue()}).Build(mi)
}
