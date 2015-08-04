//就像bytes, strings是用于操作string类型的函数库
package strings

import "io"

// Reader实现了io.Reader, io.ReaderAt, io.Seeker, io.WriterTo,
// io.ByteScanner(这是什么鬼?), and io.RuneScanner接口, 通过从一个string里读取数据
type Reader struct {
	s        string //这个就是数据源了
	i        int64  // 当前主读到的index
	prevRune int    // 因为rune可能会多个字节, index值会跳
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
