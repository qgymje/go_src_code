package http

import (
	"crypto/tls"
	"io"
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
