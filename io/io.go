// Package io 为I/O原语(primitives)提供了基本的接口
// 这很重要, 是基础,经常看到有人用这个包的接口来举例说明go的interface有多牛逼,
// 接口有多灵活
package io

import "errors"

var ErrShortWrite = errors.New("short write")

var ErrShortBuffer = errors.New("short buffer")

// 这是喜闻乐见的end of file标识
var EOF = errors.New("EOF")

var ErrUnexpectedEOF = errors.New("unexpected EOF")

var ErrNoProgress = errors.New("multiple Read calls return no data or error")

// Read从将参数p长度的数据放存到p里,那么数据流(data stream)呢?从哪里来的?
// 一般实现Reader接口的对象, 通常都是input方,比如stdin
//
type Reader interface {
	// Read会将底层数据流里的数据复制到参数p里去
	Read(p []byte) (n int, err error)
}

// Write将参数p长度的数据写进到底层的数据流
type Writer interface {
	Write(p []byte) (n int, err error)
}

//第一次调用是undefined, 何解?
type Closer interface {
	Close() error
}

// 将指向底层数据的偏移index改变
// whence 0表示相对于数据流头,1表示相对当当前偏移,2表示数据流尾, 设置偏移
// 用于下次读取或者写入
type Seeker interface {
	Seek(offset int64, whence int) (int64, error)
}

// 接口是可能组合的
type ReadWriter interface {
	Reader
	Writer
}

// 接口是可能组合的
type ReadCloser interface {
	Reader
	Closer
}

// 接口是可能组合的
type WriteCloser interface {
	Writer
	Closer
}

// 接口是可能组合的
type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

// 接口是可能组合的
type ReadSeeker interface {
	Reader
	Seeker
}

// 接口是可能组合的
type WriteSeeker interface {
	Writer
	Seeker
}

// 接口是可能组合的
// 重要的事情要说N遍
type ReadWriteSeeker interface {
	Reader
	Writer
	Seeker
}

// 这里是一对接口ReadFrom与WriteTo, 有一种对称美
// 从一个实现了Read方法的对象里读取数据, 走到EOF或者错误
type ReaderFrom interface {
	ReadFrom(r Reader) (n int64, err error)
}

// 任何实现了Write方法的对象, 将实现了WriteTo方法的对象底层的数据流
// 写进去(就是复制进去)
type WriterTo interface {
	WriteTo(w Writer) (n int64, err error)
}

type ReaderAt interface {
	ReadAt(p []byte, off int64) (n int, err error)
}

type WriterAt interface {
	WriteAt(p []byte, off int64) (n int, err error)
}

type ByteReader interface {
	ReadByte() (c byte, err error)
}

//一个Scanner通常就是指可进可退的读取操作
type ByteScanner interface {
	ByteReader
	UnreadByte() error
}

type ByteWriter interface {
	WriteByte(c byte) error
}

type RuneReader interface {
	ReadRune() (r rune, size int, err error)
}

type RuneScanner interface {
	RuneReader
	UnreadRune() error
}

type stringWriter interface {
	WriteString(s string) (n int, err error)
}
