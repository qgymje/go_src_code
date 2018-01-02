// HTTP Server. See RFC 2616.
package http

import (
	"crypto/tls"
	"log"
	"net"
	"sync"
	"time"
)

type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}

type ResponseWriter interface {
	Header() Header
	Write([]byte) (int, error)
	WriteHeader(int)
}

type Flusher interface {
	Flush()
}

type HandlerFunc func(ResponseWriter, *Request)

type ServeMux struct {
	mu    sync.RWMutex
	m     map[string]muxEntry
	hosts bool
}

type muxEntry struct {
	explicit bool
	h        Handler
	pattern  string
}

func NewServeMux() *ServeMux {
	return &ServeMux{m: make(map[string]muxEntry)}
}

var DefaultServeMux = NewServeMux()

type Server struct {
	Addr           string
	Handler        Handler
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxHeaderBytes int
	TLSConfig      *tls.Config

	TLSNextProto map[string]func(*Server, *tls.Conn, Handler)

	ConnState func(net.Conn, ConnState)

	ErrorLog          *log.Logger
	disableKeepAlives int32
}
