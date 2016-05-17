package http

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"net/textproto"
	"net/url"
)

var respExcludeHeader = map[string]bool{
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
}

type Response struct {
	Status     string
	StatusCode int
	Proto      string
	ProtoMajor int
	ProtoMinor int

	Header Header // Header就个map结构, 因此不用pointer

	Body io.ReadCloser // 包含Read([]byte)(int, error) Close() error两个方法

	ContentLenght int64

	TransferEncoding []string

	Close bool

	Trailer Header

	Request *Request

	TLS *tls.ConnectionState
}

func (r *Response) Cookies() []*Cookie {
	return readSetCookies(r.Header)
}

var ErrNoLocation = errors.New("http: no Location header in response")

func (r *Response) Location() (*url.URL, error) {
	lv := r.Header.Get("Location")
	if lv == "" {
		return nil, ErrNoLocation
	}
	if r.Request != nil && r.Request.URL != nil {
		return r.Request.URL.Parse(lv)
	}
	return url.Parse(lv)
}

func ReadResponse(r *bufio.Reader, req *Request) (*Response, error) {
	tp := textproto.NewReader(r)
	resp := &Response{
		Request: req,
	}

	line, err := tp.ReadLine()
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}

	return resp, nil
}
