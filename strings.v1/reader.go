//就像bytes, strings是用于操作string类型的函数库
package strings

import (
	"errors"
	"io"
	"unicode/utf8"
)

// Reader实现了io.Reader, io.ReaderAt, io.Seeker, io.WriterTo,
// io.ByteScanner(这是什么鬼?), and io.RuneScanner接口, 通过从一个string里读取数据
type Reader struct {
	s        string //这个就是数据源了
	i        int64  // 当前主读到的index
	prevRune int    // 因为rune可能会多个字节, index值会跳,默认值为-1
}

// Len返回Reader里的s没有被读取过的字节长度
func (r *Reader) Len() int {
	if r.i >= int64(len(r.s)) { //这里表示index超过了string的长度
		return 0
	}
	return int(int64(len(r.s)) - r.i) //这里表示index指向后面的string长度
}

// Read方法就是实现了io.Reader接口
// 参数b []byte, 指的是要从string里读的数据保存的地方
func (r *Reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 { //如果参数b为空
		return 0, nil
	}
	if r.i >= int64(len(r.s)) { //如果i已经超越了s的长度,说明读完了
		return 0, io.EOF
	}
	r.prevRune = -1        //什么意思?
	n = copy(b, r.s[r.i:]) //将s里剩余的数据复制到参数b里 copy(dst, src),因为dst有多少容量,就只能copy过多少
	r.i += int64(n)        //将i偏移, 表示这一块数据被处理过了
	return
}

// 实现了ReaderAt接口
// 将底层数据流从off位置开始, 复制到b里
func (r *Reader) ReadAt(b []byte, off int64) (n int, err error) {
	// cannot modify state - see io.ReaderAt
	if off < 0 { // 嗯? 等会看, good staff takes time...
		return 0, errors.New("strings.Reader.ReadAt: negative offset") //这里指明"私有错误"的格式: packageName.Object.Method: blabla...
	}
	if off >= int64(len(r.s)) {
		return 0, io.EOF //表示超过了底层数据流的大小了
	}
	n = copy(b, r.s[off:]) //从源数据流复制一些数据到参数b里
	if n < len(b) {
		err = io.EOF
	}
	//这里为何不做偏移操作?
	return
}

// 读取一个byte,偏移向前走一位
func (r *Reader) ReadByte() (b byte, err error) {
	r.prevRune = -1
	if r.i >= int64(len(r.s)) {
		return 0, io.EOF //这个判断反复出现啊
	}
	b = r.s[r.i] //只读一个byte
	r.i++        //偏移+1
	return
}

// 偏移向后走一位
func (r *Reader) UnreadByte() error {
	r.prevRune = -1
	if r.i <= 0 {
		return errors.New("strings.Reader.UnreadByte: at beginning of string")
	}
	r.i--
	return nil
}

// ReadRune与UnredRune同上
func (r *Reader) ReadRune() (ch rune, size int, err error) {
	if r.i >= int64(len(r.s)) {
		r.prevRune = -1
		return 0, 0, io.EOF
	}
	r.prevRune = int(r.i)
	if c := r.s[r.i]; c < utf8.RuneSelf { //如果当前index指向的是一个单字节
		r.i++
		return rune(c), 1, nil
	}
	ch, size = utf8.DecodeRuneInString(r.s[r.i:]) //大约是从这个index读取下一个rune字符
	r.i += int64(size)
	return
}

func (r *Reader) UnreadRune() error {
	if r.prevRune < 0 { //如果之前没有操作过ReadRune, 则要报错
		return errors.New("strings.Reader.UnreadRune: previous operation wat not ReadRune")
	}
	r.i = int64(r.prevRune)
	r.prevRune = -1
	return nil
}

// Seek将偏移放到参数offset位置
// whence是什么鬼?
func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	r.prevRune = -1
	var abs int64
	switch whence {
	case 0: //表示offset不管底层偏移在哪里
		abs = offset
	case 1: //表示offset需要和底层偏移的位置加起来
		abs = int64(r.i) + offset
	case 2: //这是几个意思啊?
		abs = int64(len(r.s)) + offset
	default:
		return 0, errors.New("strings.Reader.Seek: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("strings.Reader.Seek: negative position")
	}
	r.i = abs //将底层偏移设置一下
	return abs, nil
}

// WriteTo 将底层数据流写到实现了io.Writer的对象的底层数据流中
func (r *Reader) WriteTo(w io.Writer) (n int64, err error) {
	r.prevRune = -1
	if r.i >= int64(len(r.s)) { //说明数据出错了?
		return 0, nil
	}
	s := r.s[r.i:] //将还没有被读走的数据保存下,这明显不是线程安全嘛
	m, err := io.WriteString(w, s)
	if m > len(s) {
		panic("strings.Reader.WriteTo: invalid WriteString count") //在这种意想不到的情况下, 果断用panic
	}
	r.i += int64(m)
	n = int64(m)
	if m != len(s) && err == nil {
		err = io.ErrShortWrite //?
	}
	return
}

//用于生成一个strings.Reader, 它比bytes.NewBufferString更有效, 因为它是只读的
func NewReader(s string) *Reader { return &Reader{s, 0, -1} }
