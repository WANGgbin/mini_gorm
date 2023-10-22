package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Assert(expr bool, msg string, args ...interface{}) {
	if !expr {
		panic(fmt.Sprintf(msg, args...))
	}
}

// TransFromHumpToSnake 将 name 从驼峰格式转化为蛇形格式
func TransFromHumpToSnake(name string) string {
	cvs := []converter{
		CommonConverter,
	}

	for _, cv := range cvs {
		if ret := cv(name); ret != "" {
			return ret
		}
	}

	return RegularConverter(name)
}

type converter func(string) string

func RegularConverter(name string) string {
	var ret strings.Builder
	ret.Grow(len(name))

	for idx := range name {
		if name[idx] < 'A' || name[idx] > 'Z' {
			ret.WriteByte(name[idx])
			continue
		}

		ch := name[idx] - 'A' + 'a'
		if idx == 0 {
			ret.WriteByte(ch)
		} else {
			ret.WriteByte('_')
			ret.WriteByte(ch)
		}
	}

	return ret.String()
}

var CommonConvert = map[string]string{
	"ID": "id",
}

func CommonConverter(name string) string {
	return CommonConvert[name]
}

func SetRefValueUsingString(target reflect.Value, val string) error {
	switch target.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, _ := strconv.ParseInt(val, 10, 64)
		target.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, _ := strconv.ParseUint(val, 10, 64)
		target.SetUint(v)
	case reflect.Bool:
		v, _ := strconv.ParseBool(val)
		target.SetBool(v)
	case reflect.Float32, reflect.Float64:
		v, _ := strconv.ParseFloat(val, 64)
		target.SetFloat(v)
	case reflect.String:
		target.SetString(val)
	case reflect.Slice:
		// []byte 类型
		if target.Type().Elem().Kind() == reflect.Uint8 {
			target.SetBytes([]byte(val))
		} else {
			return fmt.Errorf("unknown type: []%s", target.Type().Elem().Kind().String())
		}
	case reflect.Ptr:
		target.Set(reflect.New(target.Type().Elem()))
		return SetRefValueUsingString(target.Elem(), val)
	default:
		return fmt.Errorf("unknown type %s", target.Kind().String())
	}

	return nil
}

func WrapWithBackQuote(src string) string {
	return fmt.Sprintf("`%s`", src)
}