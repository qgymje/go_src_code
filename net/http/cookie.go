package http

import "time"

// yeah, what is a cookie? 如果没有一个类型去定义, 是很难掌握它的所有特性的
// 一个类型是一个定义, 无法想像我将来还要写弱类型语言
type Cookie struct {
	Name       string
	Value      string
	Path       string
	Domain     string
	Expires    time.Time
	RawExpires string

	MaxAge   int
	Secure   bool
	HttpOnly bool
	Raw      string
	Unparsed []string
}
