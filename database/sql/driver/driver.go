// Package driver定义了一些接口, 它需要被数据库驱动实现, 并且被用于sql包中
// 这里的代码大多被用于sql包
package driver

import "errors"

// Value是一个值(它是一个空接口), 驱动必须有能力去处理
// 它要么是nil, 要么是以下类型的实例:
// int64
// float64
// bool
// []byte
// string 注:除了Rows.Next用到(啥意思啊?)
// time.Time
type Value interface{}

// Driver是数据库驱动必须要实现的接口
type Driver interface {
	// Open 返回一个新的数据库连接(其实是一个Conn接口)
	// name是特殊指定的, 通常为username:password@location/databasename?query=value
	//
	// 注意这里是说sql会维持一个sql连接池
	// 返回的连接只能在同一时间里被一个goroutine使用
	Open(name string) (Conn, error)
}

var ErrSkip = errors.New("driver: skip fast-path; continue as if unimplemented")

var ErrBadConn = errors.New("driver: bad connection")

// Execer是一个可选的接口, 它可能被一个Conn实现
// 如果Conn接口没有实现Execer, sql.DB.Exec会先生成一个prepared statement
// 然后关闭
var Execer interface {
	Exec(query string, args []Value) (Result, error)
}

type Queryer interface {
	Query(query string, args []Value) (Rows, error)
}

type Conn interface {
	Prepare(query string) (Stmt, error)

	Close() error

	Begin() (Tx, error)
}

type Result interface {
	LastInsertId() (int64, error)
	RowsAffected() (int64, error)
}

type Stmt interface {
	Close() error

	NumInput() int

	Exec(args []Value) (Result, error)

	Query(args []Value) (Rows, error)
}

type ColumnConverter interface {
	ColumnConverter(id int) ValueConverter
}

type Rows interface {
	Columns() []string

	Close() error

	Next(dest []Value) error
}

type Tx interface {
	Commit() error
	Rollback() error
}

type RowsAffected int64

var _ Result = RowsAffected(0)

func (RowsAffected) LastInsertId() (int64, error) {
	return 0, errors.New("no LastInsertId available")
}

func (v RowsAffected) RowsAffected() (int64, error) {
	return int64(v), nil
}

var ResultNoRows noRows

type noRows struct{}

var _ Result = noRows{}

func (noRows) LastInsertId() (int64, error) {
	return 0, errors.New("no LastInsertId available after DDL statement")
}

func (noRows) RowsAffected() (int64, error) {
	return 0, errors.New("no RowsAffected available after DDL statement")
}
