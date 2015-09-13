package io

import (
	"errors"
	"sync"
)

var ErrClosePipe = errors.New("io: read/write on closed pipe")

type pipeResult struct { //似乎这个结构体没有被调用
	n   int
	err error
}

// pipe就是一个管道, 想像成自来水管道, 流式体验
// 双向全时管道?
// 管道连接两头, 一头输入, 一头输出
// 这个对象是最重要的, 公开的方法都是其方法的包装
type pipe struct {
	rl    sync.Mutex //rl指的是readers lock
	wl    sync.Mutex // wl指的是writers lock
	l     sync.Mutex // 保护剩余的字段
	data  []byte     //在管道中的数据
	rwait sync.Cond
	wwait sync.Cond
	rerr  error
	werr  error
}

func (p *pipe) read(b []byte) (n int, err error) {
	p.rl.Lock() // 同一时间只能有一个reader工作
	defer p.rl.Unlock()

	p.l.Lock()   //下部分操作都被锁起来
	p.l.Unlock() //记得要解锁
	for {
		if p.rerr != nil { //如果Reader有错误
			return 0, ErrClosePipe
		}
		if p.data != nil { //如果pipe里已经流进了一部分数据
			break
		}
		if p.werr != nil {
			return 0, p.werr
		}
		p.rwait.Wait() //写等
	}
	n = copy(b, p.data)   //将管道里的数据复制到参数b里
	p.data = p.data[n:]   //将已经读取的字节清除
	if len(p.data) == 0 { //如果管道里的数据被读完了
		p.data = nil
		p.wwait.Signal() //这是神马意思
	}
	return
}

var zero [0]byte

func (p *pipe) write(b []byte) (n int, err error) {
	if b == nil { //pipe使用nil来表示当前不可用
		b = zero[:]
	}

	p.wl.Lock()
	defer p.l.Unlock()

	p.l.Lock()
	defer p.l.Unlock()
	if p.werr != nil { //如果Writer有错误
		err = ErrClosePipe
		return
	}
	p.data = b       //不理解
	p.rwait.Signal() //不理解
	for {
		if p.data == nil {
			break
		}
		if p.rerr != nil {
			err = p.rerr
			break
		}
		if p.werr != nil {
			err = ErrClosePipe
		}
		p.wwait.Wait()
	}
	n = len(b) - len(p.data) //什么意思
	p.data = nil
	return
}
