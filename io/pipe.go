package io

import (
	"errors"
	"sync"
)

// ErrClosedPipe is the error used for read or wirte operations on a closed pipe.
var ErrClosedPipe = errors.New("io: read/write on closed pipe")

type pipeResult struct {
	n   int
	err error
}

// A pipe is the shared pipe structure underlying PipeReader and PipeWriter.
type pipe struct {
	rl    sync.Mutex // gates readers one at a time
	wl    sync.Mutex // gates writers one at a time
	l     sync.Mutex // protects remaining fields
	data  []byte     // data remaining in pending write
	rwait sync.Cond  // waiting reader
	wwait sync.Cond  // waiting writer
	rerr  error      // if reader closed, error to give writes
	werr  error      // if writer closed, error to give reads
}

func (p *pipe) read(b []byte) (n int, err error) {
	p.rl.Lock()
	defer p.rl.Unlock()

	p.l.Lock()
	defer p.l.Unlock()
	for {
		if p.rerr != nil {
			return 0, ErrClosedPipe
		}
		if p.data != nil {
			break
		}
		if p.werr != nil {
			return 0, p.werr
		}
		p.rwait.Wait()
	}
	n = copy(b, p.data)
	p.data = p.data[n:]
	if len(p.data) == 0 {
		p.data = nil
		p.wwait.Signal()
	}
	return
}

var zero [0]byte

func (p *pipe) write(b []byte) (n int, err error) {
	if b == nil {
		b = zero[:]
	}

	p.wl.Lock()
	defer p.wl.Unlock()

	p.l.Lock()
	defer p.l.Unlock()
	if p.werr != nil {
		err = ErrClosedPipe
		return
	}
	p.data = b
	p.rwait.Signal()
	for {
		if p.data == nil {
			break
		}
		if p.rerr != nil {
			err = p.rerr
			break
		}
		if p.werr != nil {
			err = ErrClosedPipe
		}
		p.wwait.Wait()
	}
	n = len(b) - len(p.data)
	p.data = nil
	return
}

func (p *pipe) rclose(err error) {
	if err == nil {
		err = ErrClosedPipe
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.rerr = err
	p.rwait.Signal()
	p.wwait.Signal()
}

func (p *pipe) wclose(err error) {
	if err == nil {
		err = EOF
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.werr = err
	p.rwait.Signal()
	p.wwait.Signal()
}

type PipeReader struct {
	p *pipe
}

func (r *PipeReader) Read(data []byte) (n int, err error) {
	return r.p.read(data)
}

func (r *PipeReader) Close() error {
	return r.CloseWithError(nil)
}

func (r *PipeReader) CloseWithError(err error) error {
	r.p.rclose(err)
	return nil
}

type PipeWriter struct {
	p *pipe
}

func (w *PipeWriter) Write(data []byte) (n int, err error) {
	return w.p.write(data)
}

func (w *PipeWriter) Close() error {
	return w.CloseWithError(nil)
}

func (w *PipeWriter) CloseWithError(err error) error {
	w.p.wclose(err)
	return nil
}

func Pipe() (*PipeReader, *PipeWriter) {
	p := new(pipe)
	p.rwait.L = &p.l
	p.wwait.L = &p.l
	r := &PipeReader{p}
	w := &PipeWriter{p}
	return r, w
}
