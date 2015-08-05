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

//这是一个私有接口
type stringWriter interface {
	WriteString(s string) (n int, err error)
}

//将s写进实现Writer接口的对象里
func WriteString(w Writer, s string) (n int, err error) {
	if sw, ok := w.(stringWriter); ok { //接口的断言,将接口转为对象
		return sw.WriteString(s)
	}
	return w.Write([]byte(s))
}

// 每次看到使用接口做参数的函数/方法, 都很难用语言表达, 只能说实现Reader接口的对象, 如果直接说Reader对象, 似乎也对, 但忽略了它是接口参数的事实
// 但参数还是使用接口类型好,因为灵活, 不写死,面向接口,抽象程度高
// 从实现Reader接口的对象的底层数据流里读取min个字节, 复制到buf里
func ReadAtLeast(r Reader, buf []byte, min int) (n int, err error) {
	if len(buf) < min {
		return 0, ErrShortBuffer //对参数进行基本的判断, 这种可能预料到的错误, 只使用error即可, 不必非要弄个excpetion
	}
	for n < min && err == nil { //n初始化为0, err初始化为nil
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

// 将buf读满, 不含糊?
func ReadFull(r Reader, buf []byte) (n int, err error) {
	return ReadAtLeast(r, buf, len(buf)) //len(buf)可能是0
}

// copy的第一个参数都是dst,第二个参数是src
func CopyN(dst Writer, src Reader, n int64) (written int64, err error) {
	written, err = Copy(dst, LimitReader(src, n)) //Reader被写成了LimitedReader
	if written == n {
		return n, nil
	}
	if written < n && err == nil {
		err = EOF
	}
	return
}

func Copy(dst Writer, src Reader) (written int64, err error) {
	if wt, ok := src.(WriterTo); ok { //wt指writeTo?应该是wirter type
		return wt.WriteTo(dst)
	}
	if rt, ok := dst.(ReaderFrom); ok { //这样看上去很灵活啊
		return rt.ReadFrom(src)
	}
	buf := make([]byte, 32*1024) //默认32kb, 这么看起来一个byte slice是很廉价的
	for {
		nr, er := src.Read(buf) //实现了Reader接口就有Read方法
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr]) //实现了Writer接口就有Write方法
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil { //如果写入出错了, 就退了吧
				err = ew
				break
			}
			if nr != nw { //如果读取的长度与写入的长度不一致
				err = ErrShortWrite
				break
			}
			if err == EOF {
				break
			}
			if er != nil {
				err = er
				break
			}
		}
	}
	return written, err
}

// 做为一个函数, 取名尽量往动词取
func LimitReader(r Reader, n int64) Reader { return &LimitedReader{r, n} }

//作为一个类型, 是一个名词, 取名尽量向描述方向取
// 它是从0到N到
type LimitedReader struct {
	R Reader // underlying reader
	N int64  // max bytes remaining
}

// 覆写Reader里带的Read方法, 至少是包装一下
func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, EOF //上来还是对参数的判断,将预料到的错误检查出来, 而不是都用鬼畜的exception
	}
	if int64(len(p)) > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Read(p)
	l.N -= int64(n)
	return
}

// off指seek的位置, n指长度
func NewSectionReader(r ReaderAt, off int64, n int64) *SectionReader {
	return &SectionReader{r, off, off, off + n}
}

//它要实现Read, Seek, ReadAt
type SectionReader struct {
	r     ReaderAt //为什么ReaderAt也能作为一个底层数据源?虽然名字比较怪, 但是与Reader是一样的原理,这是一个接口,可以使用任何实现了ReadAt方法的对象
	base  int64    // base为何要和off一样的值?这就是起始偏移位置
	off   int64
	limit int64
}

func (s *SectionReader) Read(p []byte) (n int, err error) {
	if s.off >= s.limit {
		return 0, EOF
	}
	if max := s.limit - s.off; int64(len(p)) > max { //?
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
	default: //这也是醉了, default放这里的原因是好看?
		return 0, errWhence
	case 0:
		offset += s.base
	case 1:
		offset += s.off
	case 2:
		offset += s.limit
	}
	if offset < s.base {
		return 0, errOffset
	}
	s.off = offset
	return offset - s.base, nil
}

// ReadAt是一个copy操作,和Read不一样的地方, 仅仅是开始位置可以设置
// 因此如果ReadAt的第二个参数设置为0, 它就是一个Read
func (s *SectionReader) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 || off >= s.limit-s.base { //也就是说off必须在base与limit-base之间
		return 0, EOF
	}
	off += s.base
	if max := s.limit - off; int64(len(p)) > max {
		p = p[0:max]
		n, err = s.r.ReadAt(p, off)
		if err == nil {
			err = EOF
		}
		return n, err
	}
	return s.r.ReadAt(p, off)
}

// 因为limit是off + n, 就是NewSectionReader里的第三个参数的值
func (s *SectionReader) Size() int64 { return s.limit - s.base }
