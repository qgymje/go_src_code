// Package sql provides a generic interface around SQL databases.
//
// The sql package must be used in conjuction with a database dirver.
// http://golang.org/s/sqlwiki.
package sql

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"sync"
)

var drivers = make(map[string]driver.Driver)

// Register用于一个数据库驱动可使用通过参数提供的名字
// 如果Register被相同的名字的driver调用了两次, 或者驱动是nil, 它会恐慌.
// driver.Driver是一个接口, 方法列表为Open
func Register(name string, driver driver.Driver) {
	if driver == nil {
		panic("sql: Register driver is nil") //重要的地方要用panic,表示如果这里出错了,下面的就根本没法做了
	}
	if _, dup := drivers[name]; dup {
		panic("sql: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

func unregisterAllDrivers() {
	// for test.
	drivers = make(map[string]driver.Driver)
}

//返回并排序所有的驱动, 有什么用呢? 当用于测试不同的数据库时,并且在同一程序里使用
func Drivers() []string {
	var list []string
	for name := range drivers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// 虽然它只是一个字节切片, 为何要取名为RawBytes呢?
// 注: 取别名的好处是, 底层类型可以用的地方, 它都可以用, 并且还可以扩展方法
// 呃, 说是holds a reference to memory owned by the database itself.
// 和Scan相关
type RawBytes []byte

// 一大堆处理Null的方法, 先跳过

// 用于QueryRow结果为空的情况,用于代替*Row, 延时到Scan结束(这第一次见)
var ErrNoRows = errors.New("sql: no rows in result set")

// 这货表示一个新的请求连接
// 如果空闲连接不够用了, DB.conn会创建一个新的connRequest
// 并且放入db.connRequests列表中
type connRequest struct {
	conn *driverConn
	err  error
}

// 这个对象表示连接池,并且是协程安全的
// sql.DB表示一个数据库的抽象, 会管理连接池, 也是goroutine安全的
// 自动创建与释放连接,维护连接池
type DB struct {
	driver driver.Driver
	dsn    string

	mu           sync.Mutex         // 为何需要保护下面的字段?
	freeConn     []*driverConn      //看名字应该是连接,而这个字段表示连接池里的空闲连接
	connRequests []chan connRequest //这是一个channel, 表示要从连接池里取一个连接来
	numOpen      int                //已经打开的连接数
	pendingOpens int                //待定的连接数?

	// 这里解释似乎很复杂
	openerCh chan struct{} //使用了空struct, 特点是size=0
	closed   bool
	dep      map[finalCloser]depSet
	lastPut  map[*driverConn]string
	maxIdle  int //保持的最大空闲连接数, 0表示黑夜空闲连接数, 负数表示无限
	maxOpen  int //最多连接数,但感觉会超过这个设置, 直到暴出too many connections错误, <=0表示无限
}

const debugGetPut = false

// for testing?
var putConnHook func(*DB, *driverConn)

// 将一个连接放放到空闲连接池中
func (db *DB) putConn(dc *driverConn, err error) {
	db.mu.Lock()
	if !dc.inUse {
		if debugGetPut {
			fmt.Printf("putConn(%v) DUPLICATE was : %s\n\nPREVIOUS was: %s", dc, stack(), db.lastPut[dc])
		}
		panic("sql: connection returned that was never out")
	}
	if debugGetPut {
		db.lastPut[dc] = stack()
	}
	db.inUse = false

	for _, fn := range dc.onPut {
		fn()
	}
	dc.onPut = nil

	if err == driver.ErrBadConn {
		db.maybeOpenNewConnections()
		db.mu.Unlock()
		dc.Close()
		return
	}
	if putConnHook != nil {
		putConnHook(db, dc)
	}
	added := db.putConnDBLocked(dc, nil)
	db.mu.Unlock()

	if !added {
		dc.Close()
	}
}

// 需要保证db.mu.Lock()
func (db *DB) maybeOpenNewConnections() {
	numRequests := len(db.connRequests) - db.pendingOpens
	if db.maxOpen > 0 {
		numCanOpen := db.maxOpen - (db.numOpen + db.pendingOpens)
		if numRequests > numCanOpen {
			numRequests = numCanOpen
		}
	}
	for numRequests > 0 {
		db.pendingOpens++
		numRequests--
		db.openerCh <- struct{}{} //向此字段发空struct
	}
}

func (db *DB) putConnDBLocked(dc *driverConn, err error) bool {
}

// 一个带锁的实现了driver.Conn接口的对象
// 总之很重要
type driverConn struct {
	db *DB //拥有数据库抽象层DB
	// notice这里不直接恋情匿名嵌入指针

	sync.Mutex
	ci          driver.Conn //接口 Prepare Close Begin
	closed      bool
	finalClosed bool
	openStmt    map[driver.Stmt]bool //表示打开的prepared statement, Stmt是一个接口: Close NumInput Exec Query

	// 被db的mu保护
	inUse      bool
	onPut      []func()
	dbmuClosed bool // db.mu是否关闭了
}

// 看名字: releaseConn, 释放连接
func (dc *driverConn) releaseConn(err error) {
	dc.db.putConn(dc, err) //将此driverConn放回数据库连接中
}

func (dc *driverConn) removeOpenStmt(si driver.Stmt) {
	dc.Lock()
	defer dc.Unlock()
	delete(dc.openStmt, si)
}

// 是指prepared statement locked?
// yes
func (dc *driverConn) preparedLocked(query string) (driver.Stmt, error) {
	si, err := dc.ci.Prepare(query)
	if err == nil {
		if dc.openStmt == nil {
			dc.openStmt = make(map[driver.Stmt]bool)
		}
		dc.openStmt[si] = true
	}
	return si, err
}

func (dc *driverConn) closeDBLocked() func() error {
	dc.Lock()
	defer dc.Unlock()
	if dc.closed {
		return func() error { return errors.New("sql: duplicate driverConn close") }
	}
	dc.closed = true
	return dc.db.rremoveDepLocked(dc, dc)
}

func (dc *driverConn) Close() error {
	dc.Lock()
	if dc.closed {
		dc.Unlock()
		return errors.New("sql: duplicate driverConn close")
	}
	dc.closed = true
	dc.Unlock()

	dc.db.mu.Lock()
	dc.dbmuClosed = true
	fn := dc.db.remoceDepLocked(dc, dc)
	dc.db.mu.Unlock()
	return fn()
}

func (dc *driverConn) finalClose() error {
	dc.Lock()

	for si := range dc.openStmt {
		si.Close()
	}
	dc.openStmt = nil

	err := dc.ci.Close()
	dc.ci = nil
	dc.finalClosed = true
	dc.Unlock()

	dc.db.mu.Lock()
	dc.db.numOpen--
	dc.db.maybeOpenNewConnections()
	dc.db.mu.Unlock()

	return err
}

type driverStmt struct {
	sync.Locker
	si driver.Stmt
}

func (ds *driverStmt) CLose() error {
	ds.Lock()
	defer ds.Unlock()
	return ds.si.Close()
}

// dependencies
type depSet map[interface{}]bool

// WTF...
type finalCloser interface {
	finalClose() error
}

func stack() string {
	var buf [2 << 10]byte
	return string(buf[:runtime.Stack(buf[:], false)])
}
