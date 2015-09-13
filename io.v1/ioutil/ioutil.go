// Package ioutil implements some I/O utility functions.
package ioutil

import (
	"bytes"
	"io"
	"os"
	"sort"
)

// 从参数r读取数据,函数返回读取到的数据, 第二个参数表示容量大小
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, capacity))
	//buffer溢出
	defer func() {
		e := recover() //recover只能在defer里用
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr //只处理读取过大的错误,将异常转变为错误, 为什么?
		} else {
			panic(e)
		}
	}()
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}

func ReadAll(r io.Reader) ([]byte, error) {
	return readAll(r, bytes.MinRead) // bytes.MinRead = 512
}

func ReadFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var n int64
	if fi, err := f.Stat(); err == nil {
		if size := fi.Size(); size < 1e9 { //根据file的大小来读取
			n = size
		}
	}
	return readAll(f, n+bytes.MinRead) //防止n为0
}

// perm 为0600
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	//此处为何不用defer f.Close()
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil { //这里close了
		err = err1
	}
	return err
}

//一个排序, 根据文字名字排序
type byName []os.FileInfo

func (f byName) Len() int           { return len(f) }
func (f byName) Less(i, j int) bool { return f[i].Name() < f[j].Name() }
func (f byName) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func ReadDir(dirname string) ([]os.FileInfo, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Sort(byName(list)) //强制类型转换, 因为底层类型是一样的
	return list, nil
}
