package io

import "errors"

var ErrShortWrite = errors.New("short write")

var ErrShortBuffer = errors.New("short buffer")

var EOF = errors.New("EOF")

var ErrUnexpectedEOF = errors.New("unexpected EOF")

var ErrNoProgress = errors.New("multiple Read calls return no data or error")

// Read从将参数p长度的数据放存到p里,那么数据流(data stream)呢?从哪里来的?
// 一般实现Reader接口的对象, 通常都是input方,比如stdin
//
type Reader interface {
	Read(p []byte) (n int, err error)
}

// Write将参数p长度的数据写进到底层的数据流
type Writer interface {
	Write(p []byte) (n int, err error)
}
