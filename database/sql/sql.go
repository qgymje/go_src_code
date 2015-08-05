// Package sql provides a generic interface around SQL databases.
//
// The sql package must be used in conjuction with a database dirver.
// http://golang.org/s/sqlwiki.
package sql

import (
	"database/sql/driver"
	"errors"
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

// 这个对象表示连接池,并且是协程安全的
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
