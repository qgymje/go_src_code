package dirver

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

type ValueConverter interface {
	ConvertValue(v interface{}) (Value, error)
}

type Valuer interface {
	Value() (Value, error)
}

var Bool boolType

type boolType struct{}

func (boolType) String() string { return "Bool" }

func (boolType) ConvertValue(src interface{}) (Value, error) {
	switch s := src.(type) {
	case bool:
		return s, nil
	case string:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, fmt.Errorf("sql/driver: couldn't convert %q into type bool", s)
		}
		return b, nil
	case []byte:
		b, err := strconv.ParseBool(string(s))
		if err != nil {
			return nil, fmt.Errorf("sql/driver: couldn't convert %q into type bool", s)
		}
		return b, nil
	}

	sv := reflect.ValueOf(src)
	switch sv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		iv := sv.Int()
		if iv == 1 || iv == 0 {
			return iv == 1, nil
		}
		return nil, fmt.Errorf("sql/driver: couldn't convert %d into type bool", iv)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uv := sv.Uint()
		if uv == 1 || uv == 0 {
			return uv == 1, nil
		}
		return nil, fmt.Errorf("sql/driver: couldn't convert %d into type bool", uv)
	}
	return nil, fmt.Errorf("sql/driver: couldn't convert %d (%T) into type bool", src, src)
}

var Int32 int32Type

type int32Type struct{}

// 判断int32Type实现了ValueConverter接口里的方法, 如果没有实现, 则编译会提示错误
var _ ValueConverter = int32Type{}

func (int32Type) ConvertValue(v interface{}) (Value, error) {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64 := rv.Int()
		if i64 > (1<<31)-1 || i64 < -(1<<31) {
			return nil, fmt.Errorf("sql/driver: value %d overflows int32", v)
		}
		return i64, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u64 := rv.Uint()
		if u64 > (1<<31)-1 {
			return nil, fmt.Errorf("sql/driver: value %d overflows int32", v)
		}
		return int64(u64), nil
	case reflect.String:
		i, err := strconv.Atoi(rv.String())
		if err != nil {
			return nil, fmt.Errorf("sql/driver: value %q can't be converted to int32", v)
		}
		return int64(i), nil
	}
	return nil, fmt.Errorf("sql/driver: unsupported value %v (type %T) converting to int32", v, v)
}

var String stringType

type stringType struct{}

func (stringType) ConvertValue(v interface{}) (Value, error) {
	switch v.(type) {
	case string, []byte:
		return v, nil
	}
	return fmt.Sprintf("%v", v), nil
}

type Null struct {
	Converter ValueConverter
}

func (n Null) ConvertValue(v interface{}) (Value, error) {
	if v == nil {
		return nil, nil
	}
	return n.Converter.ConvertValue(v)
}

type NotNull struct {
	Converter ValueConverter
}

func (n NotNull) ConvertValue(v interface{}) (Value, error) {
	if v == nil {
		return nil, fmt.Errorf("nil value not allowed")
	}
	return n.Converter.ConvertValue(v)
}

func IsValue(v interface{}) bool {
	if IsScanValue(v) {
		return true
	}
	if _, ok := v.(string); ok {
		return true
	}
	return false
}

func IsScanValue(v interface{}) bool {
	if v == nil {
		return true
	}
	switch v.(type) {
	case int64, float64, []byte, bool, time.Time:
		return true
	}
	return false
}

var DefaultParameterConverter defaultConverter

type defaultConverter struct{}

var _ ValueConverter = defaultConverter{}

func (defaultConverter) ConvertValue(v interface{}) (Value, error) {
	if IsValue(v) {
		return v, nil
	}

	if svi, ok := v.(Valuer); ok {
		sv, err := svi.Value()
		if err != nil {
			return nil, err
		}
		if !IsValue(sv) {
			return nil, fmt.Errorf("non-Value type %T returned from Value", sv)
		}
		return sv, nil
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return nil, nil
		} else {
			return defaultConverter{}.ConvertValue(rv.Elem().Interface())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return int64(rv.Uint()), nil
	case reflect.Uint64:
		u64 := rv.Uint()
		if u64 >= 1<<63 {
			return nil, fmt.Errorf("uint64 values with hight bit set are not supported")
		}
		return int64(u64), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	}
	return nil, fmt.Errorf("unsupported type %T, a %s", v, rv.Kind())
}
