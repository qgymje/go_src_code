package http

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/url"
)

type Request struct {
	Method string
	URL    *url.URL

	Proto      string
	ProtoMajor int
	ProtoMinor int

	Header Header // 这里为啥要一样?而不是直接嵌入?

	Body io.ReadCloser

	ContentLength int64

	TransferEncoding []string

	Close bool

	Host string

	Form url.Values

	PostForm url.Values

	MultipartForm *multipart.Form //mime/multipart

	Trailer Header //额外的header,再看, 预告片的意思

	RemoteAddr string

	RequestURI string

	TLS *tls.ConnectionState
}

func (r *Request) ProtoAtLeast(major, minor int) bool {
	return r.ProtoMajor > major || r.ProtoMajor == major && r.ProtoMinor >= minor
}

func (r *Request) UserAgent() string {
	return r.Header.Get("User-Agent")
}

func (r *Request) Cookies() []*Cookie {
	return readCookies(r.Header, "")
}

var ErrNoCookie = errors.New("http: named cookie not present")

func (r *Request) Cookie(name string) (*Cookie, error) {
	for _, c := range readCookies(r.Header, name) {
		return c, nil
	}
	return nil, ErrNoCookie
}

// for http client
func (r *Request) AddCookie(c *Cookie) {
	s := fmt.Sprintf("%s=%s", sanitizeCookieName(c.Name), santizeCookieValue(c.Value))
	if c := r.Header.Get("Cookie"); c != "" {
		r.Header.Set("Cookie", c+"; "+s)
	} else {
		r.Header.Set("Cookie", s)
	}
}

func (r *Request) Referer() string {
	return r.Header.Get("Referer")
}

var multipartByReader = &multipart.Form{
	Value: make(map[string][]string),
	File:  make(map[string][]*multipart.FileHeader),
}

func (r *Request) MultipartReader() (*multipart.Reader, error) {
	if r.MultipartForm == multipartByReader {
		return nil, errors.New("http: MultipartReader called twice")
	}
	if r.MultipartForm != nil {
		return nil, errors.New("http: multipart handled by ParseMultipartForm")
	}
	r.MultipartForm = multipartByReader
	return r.multipartReader()
}

func (r *Request) multipartReader() (*multipart.Reader, error) {
	v := r.Header.Get("Content-Type")
	if v == "" {
		return nil, ErrNotMultipart
	}
	d, params, err := mime.ParseMediaType(v)
	if err != nil || d != "multipart/form-data" {
		return nil, ErrNotMultipart
	}
	boundary, ok := params["boundary"]
	if !ok {
		return nil, ErrMissingBoundary
	}
	return multipart.NewReader(r.Body, boundary), nil
}
