package strings

import (
	"errors"
	"io"
	"unicode/utf8"
)

type Reader struct {
	s        string
	i        int64
	prevRune int
}

func (r *Reader) Len() int {
	if r.i >= int64(len(r.s)) {
		return 0
	}
	return int(int64(len(r.s)) - r.i)
}

func (r *Reader) Size() int64 { return int64(len(r.s)) }

// Read是从底层数据读到数据到参数里
// 一个Reader是指能够实现被读取数据的对象
// Write是从参数p里数据写入到底层数据中
// 一个Writer是指能够实现被写入数据的对象
func (r *Reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF
	}
	r.prevRune = -1
	// copy(dst, src []Type) int
	n = copy(b, r.s[r.i:]) //这才是真正做工作的操作
	r.i += int64(n)
	return //返回参数指明变量的好处
}

// 所有的Readxx函数的参数, 都是容器, 将对象底层的数据写入到参数里
func (r *Reader) ReadAt(b []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.New("stirngs.Reader.ReadAt: negative offset")
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

// r.prevRuen为何不写一个方法
// func (r *Reader) resetRune() { r.prevRune = -1 }
// 这样省得每次都设置?
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
		return errors.New("strings.Reader.UnreadByte: at beginning of string")
	}
	r.i--
	return nil
}

func (r *Reader) ReadRune() (ch rune, size int, err error) {
	if r.i >= int64(len(r.s)) {
		r.prevRune = -1
		return 0, 0, io.EOF
	}
	r.prevRune = int(r.i) //通过此标志是ReadRune操作
	if c := r.s[r.i]; c < utf8.RuneSelf {
		r.i++
		return rune(c), 1, nil //所以就算有命名返回名, 也可以明确覆盖掉
	}
	ch, size = utf8.DecodeRuneInString(r.s[r.i:]) //将当前index后面的串发送过去,解释下一个unicode字符
	r.i += int64(size)
	return
}

func (r *Reader) UnreadRune() error {
	if r.prevRune < 0 {
		return errors.New("strings.Reader.UnreadRune: previous operation was not ReadRune")
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
		abs = offset //从给出的位置设置为当前index
	case 1:
		abs = int64(r.i) + offset //相对于当前index后的offset
	case 2:
		abs = int64(len(r.s)) + offset // ???反向?
	default:
		return 0, errors.New("strings.Reader.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("strings.Reader.Seek: negative position")
	}
	r.i = abs
	return abs, nil
}

func (r *Reader) WriteTo(w io.Writer) (n int64, err error) {
	r.prevRune = -1
	if r.i >= int64(len(r.s)) {
		return 0, nil
	}
	s := r.s[r.i:]
	m, err := io.WriteString(w, s)
	if m > len(s) {
		panic("strings.Reader.WriteTo: invalid WriteStrin count")
	}
	r.i += int64(m)
	n = int64(m)
	if m != len(s) && err == nil {
		err = io.ErrShortWrite
	}
	return
}

func NewReader(s string) *Reader { return &Reader{s, 0, -1} }
