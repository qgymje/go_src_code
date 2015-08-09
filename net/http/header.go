package http

import "net/textproto"

var reaceEnabled = false // set by race.go

type Header map[string][]string // Header是map的别名, 但是有map的可操作的方法, 比如v,ok := map[key], delete(map, key), 这里体现了"一人千面"的个性特征

func (h Header) Add(key, value string) {
	//都是相同的map结构
	textproto.MIMEHeader(h).Add(key, value) //将header类型转换
}

func (h Header) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

func (h Header) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}
