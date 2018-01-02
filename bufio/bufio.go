// Package bufio 实现了缓冲的I/O. 它包装了io.Reader和io.Writer对象
package bufio

import (
	"errors"
	"io"
)

const (
	defaultBufSize = 4096
)

var (
	ErrInvalidUnreadByte = errors.New("bufio: invalid use of UnreadByte")
	ErrINvalidUnreadRune = errors.New("bufio: invalid use of UnreadRune")
	ErrBufferFull        = errors.New("bufio: buffer full")
	ErrNegativeCount     = errors.New("bufio: negative count")
)

// Buffered input.

// Reader 给一个io.Reader对象实现缓冲功能
// 通常一个Reader要实现Read方法，ReadByte, UnreadByte type Reader struct {
type Reader struct {
	buf          []byte
	rd           io.Reader // 由调用方提供
	r, w         int
	err          error
	lastByte     int
	lastRuneSize int
}

const minReadBufferSize = 16
const maxConsecutiveEmptyReads = 100

// NewReaderSize 指定缓冲区大小返回bufio.Reader
func NewReaderSize(rd io.Reader, size int) *Reader {
	// 通过type asseration提高容错能力
	// 判断一个接口是否是一个实例，与判断一个实例是否是一个接口是相同的语法
	if b, ok := rd.(*Reader)
	if ok && len(b.buf) >= size {
		return b
	}
	if size < minReadBufferSize {
		size = minReadBufferSize
	}
	r := new(Reader)
	r.reset(make([]byte, size), rd)
	return r
}

// NewReader 包装函数,将一个io.Reader对象包装成为一个bufio.Reader对象
// 为什么需要包装？因为包装一层，可以提供额外的功能
// 所有软件问题都可以通过添加一个层来解决
func NewReader(rd io.Reader) *Reader {
	return NewReaderSize(rd, defaultBufSize)
}

// Reset 丢弃任何缓冲,重置所有状态，重设reader
func (b *Reader) Reset(r io.Reader) {
	b.reset(b.buf, r)
}

// 如果一个对象的初始化值不是默认的go的初始化值，则提供一个Rest方法
func (b *Reader) reset(buf []byte, r io.Reader) {
	*b = Reader{
		buf:          buf,
		rd:           r,
		lastByte:     -1,
		lastRuneSize: -1,
	}
}

var errNegativeRead = errors.New("bufio: reader returned negative count from Read")

func (b *Reader) fill() {
	if b.r > 0 {
		copy(b.buf, b.buf[b.r:b.w])
		b.w -= b.r
		b.r = 0
	}
}

func (b *Reader) readErr() error {
	err := b.err
	b.err = nil
	return err
}

// Peek 偷看?
// wtf of without advancing the reader?
func (b *Reader) Peek(n int) ([]byte, error) {
	if n < 0 {
		return nil, ErrNegativeCount
	}

	for b.w-b.r < n && b.w-b.r < len(b.buf) && b.err == nil {
		b.fill()
	}

	if n > len(b.buf) {
		return b.buf[b.r:b.w], ErrBufferFull
	}

	var err error
	if avail := b.w - b.r; avail < n {
		n = avail
		err = b.readErr()
		if err == nil {
			err = ErrBufferFull
		}
	}
	return b.buf[b.r : b.r+n], err
}
