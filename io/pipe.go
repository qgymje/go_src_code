package io

import (
	"errors"
	"sync"
)

var ErrClosePipe = errors.New("io: read/write on closed pipe")

// A pipe is the shared pipe structure underlying PipeReader and PipeWriter.
type pipe struct{
	rl sync.Mutex
	wl sync.Mutex
	l sync.Mutex
	data []byte
	rwait sync.Cond
	wwait sync.Cond
	rerr error
	werr error
}

func (p *pipe) read(b []byte) (n int, err error) {
	p.rl.Lock()
	defer p.rl.Unlock()

	p.l.Lock()
	defer p.l.Unlock()
	for {
		if p.rerr != nil {
			return 0, ErrClosePipe
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
		err = ErrClosePipe
		return
	}
	p.data = b
	p.rwait.Signal()
	for {
		if p.data == nil {
			break
		}
		if p.rerr != nil {
			err = ErrClosePipe
			break
		}
		p.wwait.Wait()
	}
	n = len(b) - len(p.data)
	p.data = nil
	return
}

func (p *pipe) rclose(err error) {
	if err == nil {
		err = ErrClosePipe
	}
	p.l.Lock()
	defer p.l.Unlock()
	p.rerr= err
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

type PipeReader struct{
	p *pipe
}

func (r *PipeReader) Read(data []byte) (n int, err error) {
	return r.p.read(data)
}

func (r *PipeReader) Close() error {
	return r.CloseWithError(nil)
}

func (r *PipeReader) CloseWithError(err error) error {
	r.p.rlcose(err)
	return nil
}

type PipeWriter struct{
	p *pipe
}

func(w *PipeWriter) Write(data []byte) (n int, err error) {
	return w.p.write(data)
}

func (w *PipeWriter) Close() error {
	return w.CloseWithError(nil)
}

func (w *PipeWriter) CloseWithError(err error) error {
	w.p.wclose(err)
	return nil
}

// Pipe creates a synchronous in-memory pipe.
// It can be used to connect code expecting an io.Reader
// with code expecting an io.Writer
//
// Reads and Writes on the pipe are matched one to one
// except when multiple Reads are needed to consume a single Write.
// That is, each Write to the PipeWriter blocks until it has satisfied
// one or more Reads from the PipeReader that fully consume
// the written data.
// That data is copied directly from the Write to the corresponding
// Read (or Reads); there is no internal buffering.
//
// It is safe to call Read and Write in a parallel with each other or with Close.
// Parallel calls to Read and parallel calls to Write are also safe;
// the individual calls will be gated sequentially.
func Pipe() (*PipeReader, *PipeWriter) {
	p := new(pipe)
	p.rwait.L = &p.l
	p.wwait.L = &p.l
	r := &PipeReader{p}
	w := &PipeWriter{p}
	return r,w
}