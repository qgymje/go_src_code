// Package io 提供了基本的接口做I/O原语
package io

import (
	"errors"
	"fmt"
)

const (
	SeekStart = 0
	SeekCurrent = 1
	SeekEnd = 2
)
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

// ReadAtLeast 从r里读取最少min个字节到buf里
func ReadAtLeast(r Reader, buf []byte, min int) (n int, err error) {
	if len(buf) < min {
		return 0, ErrShortBuffer
	}
	for n < min && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])
		n += nn
	}
	if n >= min {
		err = nil
	} else if n > 0 && err == EOF {
		err = ErrUnexpectedEOF
	}
	return
}

func ReadFull(r Reader, buf []byte) (n int, err error) {
	return ReadAtLeast(r, buf, len(buf))
}

func CopyN(dst Writer, src Reader, n int64) (written int64, err error) {
	written, err = Copy(dst, LimitReader(src, n))
	if written == n {
		return n, nil
	}
	if written < n && err == nil {
		err = EOF
	}
	return
}

// Go里做io的函数，Writer始终是第一个参数，Reader是第二个参数，就像http.HandleFunc里
// 第一个参数是RespnseWriter, 第二个参数是Request, builtin里copy函数，也是dst, src的顺序
func Copy(dst Writer, src Reader) (written int64, err error) {
	return copyBuffer(dst, src, nil)
}

func CopyBuffer(dst Writer, src Reader, buf []byte) (written int64, err error) {
	if buf != nil && len(buf) == 0 {
		panic("empty buffer in io.CopyBuffer")
	}
	return copyBuffer(dst, src, buf)
}

func copyBuffer(dst Writer, src Reader, buf []byte) (written int64, err error) {
	if wt, ok := src.(WriterTo); ok {
		return wt.WriteTo(dst)
	}
	if rt, ok := dst.(ReaderFrom); ok {
		return rt.ReadFrom(src)
	}
	if buf == nil {
		buf = make([]byte, 32*1024)
	}
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

func LimitReader(r Reader, n int64) Reader { return &LimitedReader{ r, n}}

type LimitedReader struct{
	R Reader
	N int64
}

func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}

func NewSectionReader(r ReaderAt, off int64, n int64) *SectionReader {
	return &SectionReader{r, off, off, off + n}
}

type SectionReader struct{
	r ReaderAt
	base int64
	off int64
	limit int64
}

func (s *SectionReader) Read(p []byte) (n int, err error) {
	if s.off >= s.limit {
		return 0, EOF
	}
	if max := s.limit - s.off; int64(len(p)) > max {
		p = p[0:max]
	}
	n, err = s.r.ReadAt(p, s.off)
	s.off += int64(n)
	return
}

var errWhence = errors.New("Seek: invalid whence")
var errOffset = errors.New("Seek: invalid offset")

func (s *SectionReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	default:
		return 0, errWhence
	case SeekStart:
		offset += s.base
	case SeekCurrent:
		offset += s.off
	case SeekEnd:
		offset += s.limit
	}
	if offset < s.base {
		return 0, errOffset
	}
	return offset - s.base, nil
}

func (s *SectionReader) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 || off >= s.limit - s.base {
		return 0, EOF
	}
	off += s.base
	if max := s.limit - off; int64(len(p)) > max {
		p = p[0:max]
		n, err= s.r.ReadAt(p, off)
		if err == nil {
			err = EOF
		}
		return n, err
	}
	return s.r.ReadAt(p, off)
}

func (s *SectionReader) Size() int64 { return s.limit - s.base}

func TeeReader(r Reader, w Writer) Reader {
	return &teeReader{r, w}
}

// TeeReader returns a Reader that writes to w what it reads from r.
// All reads from r performed through it are matched with
// corresponding writes to w. There is no internal buffering
// the write must complete before the read completes.
// Any err encountered while writing is reported as a read error.
type teeReader struct{
	r Reader
	w Writer
}

func (t *teeReader) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	fmt.Printf("read from reader: r = %+v, n = %d, err =%v\n", t.r, n, err)
	if n > 0 {
		if n, err := t.w.Write(p[:n]); err != nil {
			return n, err
		}
	}
	return
}

