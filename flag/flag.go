package flag

import "io"

type Value interface {
	String() string
	Set(string) error
}
type ErrorHandling int

const (
	ContinueOnError ErrorHandling = iota
	ExitOnError
	PanicOnError
)

type FlagSet struct {
	Usage func()

	name          string
	parsed        bool
	actual        map[string]*Flag
	formal        map[string]*Flag
	args          []string
	errorHandling ErrorHandling
	output        io.Writer
}

type Flag struct {
	Name     string
	Usage    string
	Value    Value
	DefValue string
}
