package bytes

import (
	"errors"
	"io"
	"unicode/utf8"
)

// A Buffer is a variable-sized buffer of bytes with Read and Write method.
// The zero value for Buffer is an empty buffer ready to use.
type Buffer struct {
	buf       []byte            // contents are the bytes buf[off:len(buf)]
	off       int               // read at &buf[off], write at &buf[len(buf)]
	runeBytes [utf8.UTFMax]byte // avoid allocation of slice on each WriteByte or Rune
	bootstrap [64]byte          // memory to hold first slice; helps small buffers (Printf) avoid allocation.
	lastRead  readOp            // last read operation, so that Unread* can work correctly.
}

// The readOp constants describe the last action performed on
// the buffer, so that UnreadRune and UnreadByte can
// check for invalid usage.
type readOp int

const (
	opInvalid  readOp = iota // Non-read operation.
	opReadRune               // Read rune.
	opRead                   // Any other read operation.
)

// ErrTooLarge is passed to panic if memory cannot be allocated to store data in a buffer.
var ErrTooLarge = errors.New("bytes.Buffer: too large")

// Bytes returns a slice of the contents of the unread portion of the buffer;
// len(b.Bytes()) == b.Len(). If the caller changes the contents of the
// returned slice, the contents of the buffer will change provided there
// are no intervening method calls on the Buffer.
func (b *Buffer) Bytes() []byte { return b.buf[b.off:] }

// String returns the contents of the unread portion of the buffer
// as a string. If the Buffer is a nil pointer, it returns "<nil>".
func (b *Buffer) String() string {
	if b == nil {
		// Special case, useful in debugging.
		return "<nil>"
	}
	return string(b.buf[b.off:])
}

// Len returns the number of bytes of the unread portion of the buffer;
// b.Len() == len(b.Bytes()).
func (b *Buffer) Len() int { return len(b.buf) - b.off }

// Truncate discards all but the first n unread bytes from the buffer.
// It panics if n is negative or greater than the length of the buffer.
func (b *Buffer) Truncate(n int) {
	b.lastRead = opInvalid
	switch {
	case n < 0 || n > b.Len():
		panic("bytes.Buffer: truncation out of range")
	case n == 0:
		b.off = 0
	}
	b.buf = b.buf[0 : b.off+n]
}

// Rest resets the buffer so it has no content.
// b.Reset() is the same as b.Truncate(0).
func (b *Buffer) Reset() { b.Truncate(0) }

// grow grows the buffer to guarantee space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *Buffer) grow(n int) int {
	m := b.Len()
	if m == 0 && b.off != 0 {
		b.Truncate(0)
	}
	if len(b.buf)+n > cap(b.buf) {
		var buf []byte
		if b.buf == nil && n <= len(b.bootstrap) {
			buf = b.bootstrap[0]
		} else if m+n <= cap(b.buf)/2 {
			// We can slide things down insted of allocating a new
			//slice. We only need m+n <= cap(b.buf) to slide, but
			// we insted let capacity get twice as large so we
			// don't spend all out time copying.
			copy(b.buf[:], b.buf[b.off:])
			buf = b.buf[:m]
		} else {
			// not enough space anywhere
			buf = makeSlice(2*cap(b.buf) + n)
			copy(buf, b.buf[b.off:])
		}
		b.buf = buf
		b.off = 0
	}
	b.buf = b.buf[0 : b.off+m+n]
	return b.off + m
}

// Grow grows the buffer's capacity, if necessary, to guarantee space for
// another n bytes. After Grow(n), at least n bytes can be written to the
// buffer without another allocatin.
// If n is negative, Grow will panic.
// If the buffer can't grow it will panic with ErrTooLarge.
func (b *Buffer) Grow(n int) {
	if n < 0 {
		panic("bytes.Buffer.Grow: negative count")
	}
	m := b.grow(n)
	b.buf = b.buf[0:m]
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.lastRead = opInvalid
	m := b.grow(len(p))
	return copy(b.buf[m:], p), nil
}

// WriteString appends the contents of s to the buffer, growing the buffer as
// needed. The returnvalue n is the length ofs; err is always nil. If the
// buffer becomes too large, WriteString will panic with ErrTooLarge.
func (b *Buffer) WriteString(s string) (n int, err error) {
	b.lastRead = opInvalid
	m := b.grow(len(s))
	return copy(b.buf[m:], s), nil
}

const MinRead = 512

func (b *Buffer) ReadFrom(r io.Reader) (n int64, err error) {

}
