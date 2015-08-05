package io

// 从名字上来看就是多个Reader的集合, 问题是这有什么卵用?
type multiReader struct {
	readers []Reader
}

// 和普通Read比, 外面套了一层for
func (mr *multiReader) Read(p []byte) (n int, err error) {
	for len(mr.readers) > 0 { //此处使用for
		n, err = mr.readers[0].Read(p)
		if n > 0 || err != EOF {
			if err == EOF {
				err = nil
			}
			return
		}
		mr.readers = mr.readers[1:] //这个操作可以看作简单的shift操作
	}
	return 0, EOF
}

func MultiReader(readers ...Reader) Reader {
	r := make([]Reader, len(readers))
	copy(r, readers) // ...Reader是slice的语法糖
	return &multiReader{r}
}

type multiWriter struct {
	writers []Writer
}

// 将多个multiWriter的数据写到p里去, 汇总数据流
func (t *multiWriter) Write(p []byte) (n int, err error) {
	for _, w := range t.writers { //此处为何不使用for?
		n, err = w.Write(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = ErrShortWrite
			return
		}
	}
	return len(p), nil
}

// MultiWriter creates a writer that duplicates its writes to all the
// provided writers, similar to the Unix tee(1) command.
func MultiWriter(writers ...Writer) Writer {
	w := make([]Writer, len(writers))
	copy(w, writers)
	return &multiWriter{w}
}
