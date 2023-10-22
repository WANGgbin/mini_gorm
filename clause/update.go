package clause

import (
	"fmt"
	"github.com/WANGgbin/mini_gorm/model"
	"strings"
	"time"
)

func NewUpdateBuilder(fieldValPairs map[string]interface{}) *UpdateBuilder {
	return &UpdateBuilder{fieldValPairs: fieldValPairs}
}

type UpdateBuilder struct {
	fieldValPairs map[string]interface{}
}

// Build UPDATE table SET field=val, updated_at=NOW()
func (u *UpdateBuilder) Build(mi *model.Info) *Clause {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("UPDATE `%s` SET ", mi.GetTableName()))
	// 更新时携带自动更新字段
	autoUpdateFields := mi.GetAutoUpdateTimeFields()
	pairs := make([]string, 0, len(autoUpdateFields)+len(u.fieldValPairs))
	params := make([]interface{}, 0, len(autoUpdateFields)+len(u.fieldValPairs))

	for field, val := range u.fieldValPairs {
		pairs = append(pairs, fmt.Sprintf("`%s`=?", mi.GetColumn(field)))
		params = append(params, val)
	}

	now := time.Now()
	for _, field := range autoUpdateFields {
		pairs = append(pairs, fmt.Sprintf("`%s`=?", field.GetColumn()))
		params = append(params, now)
	}

	sb.WriteString(strings.Join(pairs, ","))
	return &Clause{
		sql:    sb.String(),
		params: params,
	}
}