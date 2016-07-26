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

type Closer interface {
	Close() error
}

// Seeker 可读可写, 需要设置一个offset位置
// whence 0 表示相对于开始位置
//        1 表示相对于当前的offset位置
//        2 表示相对于结束位置, 因此可以直接读取最后几位的字节
type Seeker interface {
	Seek(offset int64, whence int) (int64, error)
}

// ReadWriter 是一个接口组合, 两个接口的所有方法都要被满足
type ReadWriter interface {
	Reader
	Writer
}

type ReadCloser interface {
	Reader
	Closer
}

type WriteCloser interface {
	Writer
	Closer
}

type ReadWriteCloser interface {
	Reader
	Writer
	Closer
}

type ReadSeeker interface {
	Reader
	Seeker
}

type WriteSeeker interface {
	Writer
	Seeker
}

type ReadWriteSeeker interface {
	Reader
	Writer
	Seeker
}

// ReadFrom 从r里读取数据, 返回读取到的字节长度或者错误
// Copy 会使用此接口用于从Reader里读取数据
type ReaderFrom interface {
	ReadFrom(r Reader) (n int64, err error)
}

// WriterTo 将数据写向提供的Writer
type WriterTo interface {
	WriteTo(w Writer) (n int64, err error)
}

// ReaderAt 从底层的Reader读取数据到p里
// 与Reader的区别是啥?
type ReaderAt interface {
	ReadAt(p []byte, off int64) (n int, err error)
}

// WriterAt 将p写入底层数据对象
type WriterAt interface {
	WriteAt(p []byte, off int64) (n int, err error)
}

// ByteReader 读取一个byte, 并且返回它
type ByteReader interface {
	ReadByte() (c byte, err error)
}

// ByteScanner 有实例么? Scanner一直都不太好理解, 也算是一种IO里的一大类别
type ByteScanner interface {
	ByteReader
	UnreadByte() error
}

// ByteWriter 将参数的一个字节写入object
type ByteWriter interface {
	WriteByte(c byte) error
}

// RuneReader 读取一个rune, rune是int32的别名
type RuneReader interface {
	ReadRune() (r rune, size int, err error)
}

// RuneScanner 表示一个可前后读取特征的对象
type RuneScanner interface {
	RuneReader
	UnreadRune() error
}

// 为何不public?
// 为了interface casting?
type stringWriter interface {
	WriteString(s string) (n int, err error)
}

// WriteString 向一个Writer写入string
// 如果writer实现了stringWriter, 则直接调用
func WriteString(w Writer, s string) (n int, err error) {
	if sw, ok := w.(stringWriter); ok {
		return sw.WriteString(s)
	}
	return w.Write([]byte(s))
}

func ReadAtLeast(r Reader, buf []byte, min int) (n int, err error) {

}

// func ReadFull(r Reader, buf []byte)(n int, err error)
