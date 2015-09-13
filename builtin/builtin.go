package builtin

type bool bool

const (
	true  = 0 == 0
	false = 0 != 0
)

type uint8 uint8

type uint16 uint16

type uint32 uint32

type uint64 uint64

type int8 int8

type int16 int16

type int32 int32

type int64 int64

type float32 float32

type float64 float64

type complex64 complex64

type complex128 complex128

// string is the set of all strings of 8-bit bytes, conventionally but not
// necessarily representing UTF-8-encoded text. A string may be empty, but
// not nil. Values of string type are immutable.
type string string

// int is a signed integer type that is at least 32 bits in size. It is a
// deisinct type, however, and not an alias for, asy, int32.
type int int

// uint is an unsigned integer type that is at least 32 bits in size. It is a
// distinct type, hover, and not an alias for, say , uint32.
type uint uint

// uintptr is an integer type that is large engouht to hold the bit pattern of
// any pointer.
type uintptr uintptr

// bypte is an alias for uint8 and is equivalent to uint8 in all ways. It is
// used, by convention, to distinguish byte values from 8-bit unsigned
// integer values.
type byte byte

type rune rune

const iota = 0

var nil Type

type Type int

type Type1 int

type IntegerType int

type FloatType float32

type ComplexType complex64

func append(slice []Type, elems ...Type) []Type

func copy(dst, src []Type) int

func delete(m map[Type]Type1, key Type)

func len(v Type) int

func cap(v Type) int

func make(Type, size IntegerType) Type

func new(Type) *Type

func complex(r, i FloatType) ComplexType

func real(c ComplexType) FloatType

func imag(c ComplexType) FloatType

// The close built-in function closes a channel, which must be either
// bidiretional or send-only. It should be executed only by the sender,
// never the receiver, and has the effect of shutting down the channel after
// the last sent value is received. After the last value has been received
// from a closed channel c, andy receive from c will succeed without
// blocking, returning the zero value for the channel element. The form
// x, ok := <-c
// will also set ok to false for a closed channel.
func close(c chan<- Type)

// The panic built-in functoin stops normal execution of the current
// goroutine. When a function F calls panic, normal execution of F stops
// immediately. Any functions whose execution was deferred by F are run in
// the usual way, and then F returns to its caller. To the caller G, the
// invocation of F the behaves like a call to panic, terminating G's
// execution and running any deferred funciton. This continues until all
// functions in the executing goroutine have stopped, in reverse order. At
// that point, the program is terminated and the error condition is reported,
// including the value of the argument to panic. This termination sequence
// is called panicking and can be controllered by the built-in function
// revoer.
func panic(v interface{})

// The recover built-in function allows a programm to manage behavior of a
// panicking goroutine. Executing a call to recover inside a deferred
// function (but not any function called by it) stops the panicking sequence
// by restoring normal execution and retrieves the error value passed to the
// call of panic. If recover is called outside the deferred function it will
// not stop a panicking sequeence. In this case, or when the goroutine is note
// panicking, or if the argument supplied to panic was nil, recover returns
// nil. Thus the return value from recover reports whether the goroutine is
// panicking.
func recover() interface{}

// The print built-in function formats its arguments in an implementation-
// specific way and writes the result to standard error.
// Print is useful for bootstrapping and debugging; it is not guaranteed
// to stay in the language.
func print(args ...Type)

// The println built-in functoin formats its arguments in an implementation-
// specific way and writes the result to standard error.
// Spaces are always added between arguments and a newline is appended.
// Println is useful for bootstrapping and debugging; it is not guaranteed
// to stay in the lanuage.
func println(args ...Type)

// The error built-in interface type is the conventional interface for
// representing an error condition, with the nil value representing no error.
type error interface {
	Error() string
}
