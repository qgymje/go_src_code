// Package io 提供了基本的接口做I/O原语
package io

import "errors"

// ErrShortWrite 表示
var ErrShortWrite = errors.New("short write")

// ErrShortBuffer 表示读取数据比提供的buffer容器要多
var ErrShortBuffer = errors.New("short buffer")

// EOF 当无法读取到更多的输入时, 会返回的错误
var EOF = errors.New("EOF")

// ErrUnexpectedEOF 在读取过程中断了
var ErrUnexpectedEOF = errors.New("unexpected EOF")

var ErrNoProgress = errors.New("multiple Read calls return no data or error")

// 对象都是围绕接口作设计
type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}
