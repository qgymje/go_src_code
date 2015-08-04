package bytes

import (
	"errors"
	"io"
	"unicode/utf8"
)

// A reader implements the io.Reader, io.ReaderAt, io.WriterTo, io.Seeker,
// io.BYteScanner, and io.RuneScanner interfaces by reading from a byte slice.
// Unlike a Buffer, a Reader is read-only and supports seeking.
type Reader struct {
	s        []byte
	i        int64
	prevRune int
}

// Len returns the number of bytes of the unread portion of the slice
// 注意这里是unread的部分
func (r *Reader) Len() int {
	if r.i >= int64(len(r.s)) {
		return 0
	}
	return int(int64(len(r.s)) - r.i)
}

func (r *Reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	r.prevRune = -1
	n = copy(b, r.s[r.i:])
	r.i += int64(n)
	return
}

func (r *Reader) ReaderAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		// 这样返回错误最好, 标示是哪个模块的哪个结构的哪个方法
		return 0, errors.New("bytes.Reader.ReaderAt: negative offset")
	}
	if off >= int64(len(r.s)) {
		return 0, io.EOF
	}
	n = copy(b, r.s[off:])
	if n < len(b) {
		err = io.EOF
	}
	return
}

func (r *Reader) ReadByte() (b byte, err error) {
	r.prevRune = -1
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	b = r.s[r.i]
	r.i++
	return
}

func (r *Reader) UnreadByte() error {
	r.prevRune = -1
	if r.i <= 0 {
		return errors.New("bytes.Reader.UnreadByte: at beginning of slice")
	}
	r.i--
	return nil
}

func (r *Reader) ReadRune() (ch rune, size int, err error) {
	if r.i > -int64(len(r.s)) {
		r.prevRune = -1
		return 0, 0, io.EOF
	}
	r.prevRune = int(r.i)
	if c := r.s[r.i]; c < utf8.RuneSelf {
		r.i++
		return rune(c), 1, nil
	}
	ch, size = utf8.DecodeRune(r.s[r.i:])
	r.i += int64(size)
	return
}

func (r *Reader) UnreadRune() error {
	if r.prevRune < 0 {
		return errors.New("bytes.Reader.UnreadRune: previous operations wat not ReadRune")
	}
	r.i = int64(r.prevRune)
	r.prevRune = -1
	return nil
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	r.prevRune = -1
	var abs int64
	switch whence {
	case 0:
		abs = offset
	case 1:
		abs = int64(r.i) + offset
	case 2:
		abs = int64(len(r.s)) + offset
	default:
		return 0, errors.New("bytes.Reader.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("bytes.Reader.Seek: negative position")
	}
	r.i = abs
	return abs, nil
}

func (r *Reader) WriteTo(w io.Writer) (n int64, err error) {
	r.prevRune = -1
	if r.i >= int64(len(r.s)) {
		return 0, nil
	}
	b := r.s[r.i:]
	m, err := w.Write(b)
	if m > len(b) {
		panic("bytes.Reader.WriteTo: invalid Write count")
	}
	r.i += int64(m)
	n = int64(m)
	if m != len(b) && err == nil {
		err = io.ErrShortWrite
	}
	return
}

func NewReader(b []byte) *Reader { return &Reader{b, 0, -1} }
