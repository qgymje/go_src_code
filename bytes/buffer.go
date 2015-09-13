package bytes

import (
	"errors"
	"unicode/utf8"
)

// 一个可读可写的对象
// 比起Read来, 还需要操作Write
type Buffer struct {
	buf       []byte //底层数据, buf[off : len(buf)]
	off       int    //read at &buf[off], write at &buf[len(buf)]理解这句话就理解了buffer的原理
	runeBytes [utf8.UTFMax]byte
	boostrap  [64]byte
	lastRead  readOp
}

// 所以状态都不要直接用int类型, 不好, 还是包装一下, 更加可读
type readOp int

// 常量不一定非要大写, 大小写只适用于private/public
const (
	opInvalid readOp = iota
	opReadRune
	opRead
)

var ErrTooLarge = errors.New("bytes.Buffer: too large")

func (b *Buffer) Bytes() []byte { return b.buf[b.off:] }

// 所以, 这就是真相
func (b *Buffer) String() string {
	if b == nil {
		return "<nil>"
	}
	return string(b.buf[b.off:])
}

func (b *Buffer) Len() int { return len(b.buf) - b.off }

func (b *Buffer) Cap() int { return cap(b.buf) }

func (b *Buffer) Truncate(n int) {
	b.lastRead = opInvalid //有一个类型作为常量的好处是, 不需要知道1,2,3分别代表啥, less concern
	switch {
	case n < 0 || n > b.Len():
		// 这里为何不用package.struct.method格式?难道这就是panic与error的区别
		panic("bytes.Buffer: truncation out of range")
	case n == 0:
		b.off = 0
	}
	b.buf = b.buf[0 : b.off+n]
}

func (b *Buffer) Reset() { b.Truncate(0) }

// 这个就是Buffer的核心逻辑了
func (b *Buffer) grow(n int) int {

}
