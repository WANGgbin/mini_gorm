package model

import (
	"errors"
	"fmt"
	error2 "github.com/WANGgbin/mini_gorm/error"
	"github.com/WANGgbin/mini_gorm/utils"
	"reflect"
	"strings"
	"time"
)

// Info 解析 model of go object
type Info struct {
	tableName            string
	primaryFieldTag      *FieldTag
	softDeleteFieldTag   *FieldTag
	autoUpdateTimeFields []*FieldTag
	FieldTags            []*FieldTag
}

func (i *Info) GetPrimaryField() string {
	return i.primaryFieldTag.fieldName
}

func (i *Info) GetPrimaryColumn() string {
	return i.primaryFieldTag.column
}

func (i *Info) GetTableName() string {
	return i.tableName
}

func (i *Info) GetColumns() []string {
	columns := make([]string, 0, len(i.FieldTags))
	for _, tag := range i.FieldTags {
		columns = append(columns, tag.column)
	}

	return columns
}

func (i *Info) GetFieldNames() []string {
	fields := make([]string, 0, len(i.FieldTags))
	for _, tag := range i.FieldTags {
		fields = append(fields, tag.fieldName)
	}

	return fields
}

func (i *Info) GetFieldTagsByFields(fields []string) ([]*FieldTag, error) {
	ret := make([]*FieldTag, 0, len(fields))
	for _, field := range fields {
		ft := i.GetFieldTagByField(field)
		if ft == nil{
			return nil, error2.ErrShouldUseFieldNameToSpecifyColumn
		}
		ret = append(ret, ft)
	}
	return ret, nil
}

func (i *Info) GetFieldTagByField(field string) *FieldTag {
	for _, ft := range i.FieldTags {
		if field == ft.fieldName {
			return ft
		}
	}

	return nil
}

func (i *Info) IsValidField(field string) bool {
	return i.GetFieldTagByField(field) != nil
}

func (i *Info) IsValidFields(fields []string) bool {
	for _, field := range fields {
		if !i.IsValidField(field) {
			return false
		}
	}
	return true
}

func (i *Info) GetColumn(name string) string {
	for _, field := range i.FieldTags {
		if name == field.fieldName || name == field.column {
			return field.column
		}
	}
	return ""
}

func (i *Info) ToSetPrimaryKey() bool {
	return i.primaryFieldTag.autoIncrement
}

// GetAutoUpdateTimeFields 获取需要自动更新为当前时间的字段
func (i *Info) GetAutoUpdateTimeFields() []*FieldTag {
	ret := make([]*FieldTag, 0, len(i.FieldTags))
	for _, field := range i.FieldTags {
		if field.autoUpdateTime || field.fieldName == "UpdatedAt" {
			ret = append(ret, field)
		}
	}
	return ret
}

func (i *Info) GetSoftDeleteTag() *FieldTag {
	return i.softDeleteFieldTag
}

type Parser struct {
	refTyp reflect.Type

	mi *Info
}

func Parse(obj interface{}) (*Info, error) {
	parser := &Parser{
		mi: &Info{},
	}
	return parser.Parse(obj)
}

func (m *Parser) Parse(obj interface{}) (*Info, error) {
	m.setReflectItem(reflect.TypeOf(obj))
	if err := m.doParse(); err != nil {
		return nil, err
	}
	return m.mi, nil
}

func (m *Parser) setReflectItem(refTyp reflect.Type) {
	if refTyp.Kind() == reflect.Slice {
		m.setReflectItem(refTyp.Elem())
		return
	}
	if refTyp.Kind() == reflect.Ptr {
		refTyp = refTyp.Elem()
	}
	utils.Assert(refTyp.Kind() == reflect.Struct, "should be struct, but got: %s", refTyp.Kind().String())

	m.refTyp = refTyp
}

func (m *Parser) doParse() error {
	// 表名默认就是蛇形
	m.parseTableName()
	return m.parseColumns()
}

func (m *Parser) parseTableName() {
	m.mi.tableName = utils.TransFromHumpToSnake(m.refTyp.Name())
}

func (m *Parser) parseColumns() error {
	for idx := 0; idx < m.refTyp.NumField(); idx++ {
		m.parseColumn(m.refTyp.Field(idx))
	}

	return m.validate()
}

// validate 校验模型 tag 等信息是否准确
func (m *Parser) validate() error {
	fns := []func() error{
		m.setPrimaryKey,
		m.setSoftDelete,
	}

	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	return nil
}

func (m *Parser) setPrimaryKey() error {
	// 校验：是否存在多个主键等
	for _, tag := range m.mi.FieldTags {
		if tag.primaryKey {
			// 定义多个主键
			if m.mi.primaryFieldTag != nil {
				return fmt.Errorf("duplicate primaryKey: %s and %s", m.mi.primaryFieldTag.column, tag.column)
			}
			m.mi.primaryFieldTag = tag
		}
	}

	// 使用 ID 作为主键，如果还未找到，报错
	if m.mi.primaryFieldTag == nil {
		for _, tag := range m.mi.FieldTags {
			if tag.column == "id" {
				m.mi.primaryFieldTag = tag
				break
			}
		}
	}

	if m.mi.primaryFieldTag == nil {
		return errors.New("cant find primary key")
	}

	return nil
}

func (m *Parser) setSoftDelete() error {
	for _, field := range m.mi.FieldTags {
		if field.softDelete != nil {
			if m.mi.softDeleteFieldTag != nil {
				return fmt.Errorf("both %s and %s are soft deleted", m.mi.softDeleteFieldTag.fieldName, field.fieldName)
			}
			m.mi.softDeleteFieldTag = field
		}
	}
	return nil
}

func (m *Parser) parseColumn(fieldTyp reflect.StructField) {
	ft := newFieldTag(fieldTyp)
	m.mi.FieldTags = append(m.mi.FieldTags, ft)
}

// FieldTag 每列 gorm tag 的结构化表达
type FieldTag struct {
	fieldName      string
	column         string
	primaryKey     bool
	autoIncrement  bool
	autoUpdateTime bool
	softDelete     *SoftDeleteTag
	defaultValue   string
}

type SoftDeleteTag struct {
	sdType SoftDeleteType
}

type SoftDeleteType string

const (
	SoftDeleteTime SoftDeleteType = ""
	SoftDeleteMill SoftDeleteType = "mill"
	SoftDeleteNano SoftDeleteType = "nano"
	SoftDeleteFlag SoftDeleteType = "flag"
)

func newFieldTag(fieldTyp reflect.StructField) *FieldTag {
	ret := &FieldTag{
		fieldName:  fieldTyp.Name,
		column:     utils.TransFromHumpToSnake(fieldTyp.Name),
		primaryKey: false,
	}

	tag := fieldTyp.Tag.Get("gorm")
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		kvPair := strings.SplitN(part, ":", 2)
		switch kvPair[0] {
		case "primaryKey":
			ret.primaryKey = true
		case "autoIncrement":
			ret.autoIncrement = true
		case "autoUpdateTime":
			ret.autoUpdateTime = true
		case "softDelete":
			ret.softDelete = &SoftDeleteTag{}
			if len(kvPair) > 1 {
				ret.softDelete.sdType = SoftDeleteType(kvPair[1])
			}
		case "column":
			ret.column = kvPair[1]
		case "default":
			ret.defaultValue = kvPair[1]
		}
	}

	return ret
}

func (ft *FieldTag) GetValue(target reflect.Value) (interface{}, error) {
	if !target.IsZero() {
		return target.Interface(), nil
	}

	// 如果设置默认值，使用默认值
	if ft.defaultValue != "" {
		err := utils.SetRefValueUsingString(target, ft.defaultValue)
		if err != nil {
			return nil, fmt.Errorf("set dfl value %s to field %s error: %v", ft.defaultValue, ft.fieldName, err)
		}
		return target.Interface(), nil
	}

	switch ft.column {
	// 这两个字段需要设置为当前时间
	case "updated_at", "created_at":
		switch target.Interface().(type) {
		case *time.Time, time.Time:
			return time.Now(), nil
		case int:
			return time.Now().Unix(), nil
		}
	}

	return target.Interface(), nil
}

func (ft *FieldTag) GetColumn() string {
	return ft.column
}

func (ft *FieldTag) GetFieldName() string {
	return ft.fieldName
}

// GetSoftDeleteValue 调用者保证 ft 为软删除字段
func (ft *FieldTag) GetSoftDeleteValue() interface{} {
	switch ft.softDelete.sdType {
	case SoftDeleteTime:
		return time.Now()
	case SoftDeleteMill:
		// no UnixMill() ?
		return int64(float64(time.Now().UnixNano()) / float64(1000))
	case SoftDeleteNano:
		return time.Now().UnixNano()
	case SoftDeleteFlag:
		return 1
	default:
		utils.Assert(false, "invalid soft delete type: %v", ft.softDelete.sdType)
		return nil
	}
}
