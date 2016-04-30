package json

import (
	"bytes"
	"reflect"
	"runtime"
	"strconv"
	"sync"
)

// 如果一个对象实现了Marshaler接口(要有MarshalJSON)方法
// json.Marshal(v) 会调用其自定义的MarshalJSON来办出json encoding
// 将一个map或者struct转为json格式的字节数组
func Marshal(v interface{}) ([]byte, error) {
	e := &encodeState{} // 初始化未必用newStruct方法
	err := e.marshal(v)
	if err != nil {
		return nil, err
	}
	return e.Bytes(), nil
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	b, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = Indent(&buf, b, prefix, indent)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// 不是很明白
func HTMLEscape(dst *bytes.Buffer, src []byte) {
	start := 0
	for i, c := range src {
		if c == '<' || c == '>' || c == '&' {
			if start < i {
				dst.Write(src[start:i])
			}
		}
	}
}

// Marshaler 是用于对象自定义成JSON
type Marshaler interface {
	MarshalJSON() ([]byte, error)
}

type UnsupportedTypeError struct {
	Type reflect.Type
}

// 一个实现了Error()方法的对象, 就可以作为error
func (e *UnsupportedTypeError) Error() string {
	return "json: unsupported type: " + e.Type.String()
}

type UnsupportedValueError struct {
	Value reflect.Value
	Str   string
}

func (e *UnsupportedValueError) Error() string {
	return "json: unsupported value: " + e.Str
}

type InvalidUTF8Error struct {
	S string
}

func (e *InvalidUTF8Error) Error() string {
	return "json: invalid UTF-8 in string: " + strconv.Quote(e.S)
}

type MarshaleError struct {
	Type reflect.Type
	Err  error
}

func (e *MarshaleError) Error() string {
	return "json: error calling MarshalJSON for type " + e.Type.String() + ": " + e.Err.Error()
}

var hex = "0123456789abcdef"

type encodeState struct {
	bytes.Buffer
	scratch [64]byte
}

func (e *encodeState) marshal(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				panic(s)
			}
			err = r.(error)
		}
	}()
	e.reflectValue(reflect.ValueOf(v))
	return nil
}

func (e *encodeState) reflectValue(v reflect.Value) {
	valueEncoder(v)(e, v, false)
}

type encoderFunc func(e *encodeState, v reflect.Value, quoted bool)

var encoderCache struct {
	sync.RWMutex
	m map[reflect.Type]encoderFunc
}

func valueEncoder(v reflect.Value) encoderFunc {
	if !v.IsValid() {
		return invalidValueEncoder
	}
	return typeEncoder(v.Type())
}

func typeEncoder(t reflect.Type) encoderFunc {
	encoderCache.RLock()
	f := encoderCache.m[t]
	encoderCache.RUnlock()
	if f != nil {
		return f
	}
}
