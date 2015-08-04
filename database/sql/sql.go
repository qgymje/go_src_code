// Package sql provides a generic interface around SQL databases.
//
// The sql package must be used in conjuction with a database dirver.
// http://golang.org/s/sqlwiki.
package sql

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"sort"
	"sync"
)

var drivers = make(map[string]driver.Driver)

// Register makes a database driver available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panic.
// Register用于一个数据库驱动可使用通过参数提供的名字
// 如果Register被相同的名字的driver调用了两次, 或者驱动是nil, 它会恐慌.
// driver.Driver是一个接口, 方法列表为Open
func Register(name string, driver driver.Driver) {
	if driver == nil {
		panic("sql: Register driver is nil")
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

// Drivers returns a sorted list of the names of the registered drivers.
func Drivers() []string {
	var list []string
	for name := range drivers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// RawBytes is a byte slice that holds a reference to memory owned by
// the database itself. After a Scan into a RawBytes, the slice is only
// valid until the next call to Next, Scan, or CLose.
type RawBytes []byte

// NullString represetnts a string that may be null.
// NullString implements the Scanner interface so
// it can be used as a scan destination:
// var s NullString
// err := db.QueryRow("SELECT name FROM foo WHERE id=?", id).Scan(&s)
// ...
// if s.Valid {
//  // use s.String
//	} esle {
//    // NULL value
//	}
type NullString struct {
	String string
	Valid  bool
}

// Scan implements the Scanner interface.
func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		ns.String, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return convertAssign(&ns.String, value)
}

// Value implements the driver Valuer interface.
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

// NullInt64 represents an int64 that maybe null.
// NullInt64 implements the Scanner interface so
// it can be used as a scan destination, similar to NullString.
type NullInt64 struct {
	Int64 int64
	Valid bool // Valid is true if Int64 is no NULL
}

func (n *NullInt64) Scan(value interface{}) error {
	if value == nil {
		n.Int64, n.Valid = 0, false
		return nil
	}
	n.Valid = true
	return convertAssign(&n.Int64, value)
}

func (n NullInt64) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.Int64, nil
}

// Scanner is an interface used by Scan.
type Scanner interface {
	// Scan assigns a value from a database driver.
	//
	// The src value will be of one of the following restricted
	// set of types:
	//
	// int64
	// float64
	// bool
	// []byte
	// string
	// time.Time
	// nil - for NULL values
	//
	// An error should be returned if the value can not be stored
	// without loss of information.
	Scan(src interface{}) error
}

// ErrNoRows is returned by Scan when QueryRow doesn't return a
// row. In such case, QueryRow returns a placeholder *Row value that
// defers this error until a Scan.
var ErrNoRows = errors.New("sql: no rows in result set")

// DB is database handle representing a pool of zero or more
// underlying connections. It's safe for concurrent use by multiple
// goroutines.
//
// The sql package creates and frees connections automatically; it
// also maintains a free pool of idle connections. If the database has
// a concept of per-connection state, such state can only be reliably
// observed within a transaction. Once DB.Begin is called, the
// returned Tx is bound to a single connection. Once Commit or
// Rollback is called on the transaction, that transaction's
// connection is returned to DB's idle connection pool. The pool size
// can be controlled with SetMaxIdleConns.
// 说了这样几个事情:
// 1. sql.DB对象维护一个连接池, 这个池大小通过SetMaxIdleConns方法设置
type DB struct {
	driver driver.Driver
	dsn    string // data source name

	mu           sync.Mutex // protects following fields
	freeConn     []*driverConn
	connRequests []chan connRequest
	numOpen      int
	pendingOpens int

	//
	openerCh chan struct{}
	closed   bool
	dep      map[finalCloser]depSet
	lastPut  map[*driverConn]string
	maxIdle  int
	maxOpen  int
}

// driverConn wraps a driver.Conn with a mutext, to
// be held during all calls into the Conn. (including any calls onto
// interfaces returned via that Conn, such as calls on Tx, Stmt, Result, Rows)
type driverConn struct {
	db          *DB
	sync.Mutex  // guards following
	ci          driver.Conn
	closed      bool
	finalClosed bool // ci.Close has been called
	openStmt    map[driver.Stmt]bool

	// guarded by db.mu
	inUse      bool
	onPut      []func() // code (with db.mu held) run when conn is next returned
	dbmuClosed bool     // same as closed, but guarded by db.mu, for connIfFree
}

func (dc *driverConn) releaseConn(err error) {
	dc.db.putConn(dc, err)
}

// driverStmt associates a driver.Stmt with the
// *driverConn from which it came, so the driverConn's lock can be
// held during calls.
type driverStmt struct {
	sync.Locker // the *driverConn
	si          driver.Stmt
}

func (ds *driverStmt) Close() error {
	ds.Lock()
	defer ds.Unlock()
	return ds.si.CLose()
}

type depSet map[interface{}]bool

type finalCloser interface {
	finalClose() error
}

func (db *DB) addDep(x finalCloser, dep interface{}) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.addDepLocked(x, dep)
}

func (db *DB) addDepLocked(x finalCloser, dep interface{}) {

}

// This is the size of the connectionOpener request chan (dn.openerCh).
// This value should be larger than the maximum typical value
// used for db.maxOpen. If maxOpen is significantly larger than
// conncetionRequestQueueSize then it is possible for ALL calls into the *DB
// to block until the connectionOpener can satisfy the backlog of request.
var connectionRequestQueueSize = 1000000

// Open opens a database specified by its database driver name and a
// driver-specific data source name, usually consisting of at least a
// database name and connection information.
//
// Most users will open a database via a driver-specific connection
// helper function that returnsa *DB. No database drivers are included
// in the Go standard library. See ...
// a list of third-party drivers.
//
// Open may just validate its arguments without creating a connection
// to the database. To verify that the data source name is valid, call
// Ping.
// 有必要调用db.Ping()
//
// The returned DB is safe for concurrent use by multiple goroutines
// and maintains its own pool of idle connection. Thus, the Open
// function should be called just once. It is rarely necessary to close a DB.
// 这里说到很少会去关闭db.Close()
func Open(driverName, dataSourceName string) (*DB, error) {
	driveri, ok := driver[driverName]
	if !ok {
		return nil, fmt.Errorf("sql: unknow driver %q (forgotten import?)", driverName)
	}

	db := &DB{
		driver:   driveri,
		dsn:      dataSourceName,
		openerCh: make(chan struct{}, connectionRequestQueuesize),
		lastPut:  make(map[*driverConn]string),
	}
	// what's going on?
	go db.connectionOpener()
	return db, nil
}

// Ping verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (db *DB) Ping() error {
	dc, err := db.conn()
	if err != nil {
		return err
	}
	db.putConn(dc, nil)
	return nil
}

// CLose closes the database, releasing any open resources.
//
// It is rare to Close a DB, as the DB handle is meant to be
// long-lived and shared between many goroutines.
func (db *DB) Close() error {
	db.mu.Lock()
	if db.closed {
		db.mu.Unlock()
		return nil
	}
	close(db.openerCh)
	var err error
	// 创新空数组,如果知道长度, 记得要将cap设置为长度
	fns := make([]func() error, 0, len(db.freeConn))
	for _, dc := range db.freeConn {
		fns = append(fns, dc.closeDBLocked())
	}
	db.freeConn = nil
	db.closed = true
	for _, req := range db.connRequests {
		close(req)
	}
	db.mu.Unlock()
	for _, fn := range fns {
		err1 := fn()
		if err1 != nil {
			err = err1
		}
	}
	return err
}

// 默认空闲连接数
const defaultMaxIdleConns = 2

func (db *DB) maxIdleConnsLocked() int {
	n := db.maxIdle
	switch {
	case n == 0:
		return defaultMaxIdleConns
	case n < 0:
		return 0
	default:
		return n
	}
}

// SetMaxIdleConns sets the maximum number of connections in the idle
// connection pool.
//
// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns
// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit
//
// If n<= 0, no idle connections are retained.
func (db *DB) SetMaxIdleConns(n int) {
	db.mu.Lock()
	if n > 0 {
		db.maxIdle = n
	} else {
		db.maxIdle = -1
	}

	if db.maxOpen > 0 && db.maxIdleConnsLocked() > db.maxOpen {
		db.maxIdle = db.maxOpen
	}

	var closing []*driverConn
	idleCount := len(db.freeConn)
	maxIdle := db.maxIdleConnsLocked()
	if idleCount > maxIdle {
		closing = db.freeConn[maxIdle:]
		db.freeConn = db.freeConn[:maxIdle]
	}
	db.mu.Unlock()
	for _, c := range closing {
		c.Close()
	}
}

// SetMaxOpenConns sets the maximum number of open connections to the database.
//
// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
// MaxIdleConns, then MaxIdleConns will be reduced to match the new
// MaxOpenConns limit
// 如果MaxIdleConns大于0,并且新的MaxOpenConns小于MaxIdleConns, 那么MaxIdleConns将减少到匹配参数n的MaxOpenConns值
//
// If n <= 0, then there is no limit on the number of open connecitons.
// 如果n <=0, 那么,打开的连接数是无限的
// The default is 0 (unlimited).
// 默认是0(无限的)
func (db *DB) SetMaxOpenConns(n int) {
	db.mu.Lock()
	db.maxOpen = n
	if n < 0 {
		db.maxOpen = 0
	}
	syncMaxIdle := db.maxOpen > 0 && db.maxIdleConnsLocked() > db.maxOpen
	db.mu.Unlock()
	if syncMaxIdle {
		db.SetMaxIdleConns(n)
	}
}

func (db *DB) maybeOpenNewConnections() {
	numRequests := len(db.connRequests) - db.pendingOpens
	if db.maxOpen > 0 {

	}
}
