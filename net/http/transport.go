package http

import (
	"net"
	"sync"
	"time"
)

var DefaultTransport RoundTripper = &Transport{
	Proxy: ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DailContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

const DefaultMaxIdleConnsPerHost = 2

// 默认会缓存连接，所以会导致很多tcp连接都无法断开，可以使用ColoseIdleConnections方法
// 以及MaxIdleConnsPerHost，DisableKeepAlives来管理
// 应该能够复用，不然浪费性能,并用是thread-safe的，可以被多个goroutine使用，棒
// 它是HTTP的传输层，较为底层，如果使用cookie, redirect，应该使用http.Client
// 支持HTTP/2
type Transport struct {
	idleMu     sync.Mutex
	wantIdle   bool
	idleConn   map[connectMethodKey][]*persistConn
	idleConnCh map[connectMethodKey]chan *persistConn
	idleLRU    connLRU
}
