package utils

import "reflect"

// GetNoZeroFields 获取结构体非零值成员
func GetNoZeroFields(obj interface{}) map[string]interface{} {
	ptrRefVal := reflect.ValueOf(obj)
	refVal := ptrRefVal.Elem()
	Assert(ptrRefVal.Kind() == reflect.Ptr && refVal.Kind() == reflect.Struct, "obj must be an pointer to struct")

	ret := make(map[string]interface{}, refVal.NumField())
	for idx := 0; idx < refVal.NumField(); idx++ {
		fieldVal := refVal.Field(idx)
		if fieldVal.IsZero() {
			continue
		}
		fieldTyp := refVal.Type().Field(idx)
		ret[fieldTyp.Name] = fieldVal.Interface()
	}

	return ret
}