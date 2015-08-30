The Laws of Reflect
====

Introduction
---

Reflection in computing is *the ability of a program to examine(检查) its own structure, particularly throught types*; it's a form of *metaprogramming*. It's also a great source of confusion.

In this article we attempt to clarify things by explaining how reflection works in Go. Each language's reflection model is different(and may languages don't support it at all(Go团队很傲娇)), but this article is about Go, so for the rest of this artichle the word "reflect" should be taken to mean "reflection in Go".

> 反射是能够让程序判断出数据的类型; 它是一种元编程形式; 我从来没真正理解元编程

Types and interfaces
----

>我的预感, 这又将刷新我对Go类型系统的理解, 第一次是通过许式伟的书, 第二次是[博客的博文](http://blog.sina.com.cn/s/articlelist_2615392497_1_1.html), 这次希望这篇blog以及学习第三方库,让我加深认识

Because reflection builds on the *type system*, let's start with a refresher about types in Go.

Go is statically typed. Every variable has a static type, that is, exactly one type know and fixed *at compile time*: int float32, \*MyType, []byte, and so on. If we declare
```
type MyInt int

var i int
var j MyInt
```

then i has type int and j has type MyInt. The variables i and j have *distinct* static types and, although they have the same *underlying type*, the cannot be assigned to one another without a conversion.
> 类型必须要转换, MyInt与\*MyInt不是同一个类型, []MyInt与[]\*MyInt不是同一个类型, [3]MyInt与[]MyInt不是同一个类型

One important category of type is *interface types*, which represent *fixed sets of methods*.(interface首先是methods集合) An *interface variable* can store any concrete(具体) (non-interface) value as long as that value implements the interface's methods. A well-known pair of examples is io.Reader and io.Writer, the types Reader and Writer from the [io package](http://golang.org/pkg/io/):

> interface types and interface variable

```
// Reader is the interface that wraps the basic Read method.
type Reader interface {
    Read(p []byte) (n int, err error)
}

// Writer is the interface that wraps the basic Write method.
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

Any type that implements a Read (or Write) method with this signature is said to implement io.Reader (or io.Writer). For the purposes of this discussion, that means that a variable of type io.Reader can hold any value whose type has a read method:

```
var r io.Reader
r = os.Stdin
r = bufio.NewReader(r)
r = new(bytes.Buffer)
// and so on
```

It's important to be clear that whatever concrete value r may hold, `r`'s type is always io.Reader: Go is statically typed and the static type of r is io.Reader.

An extremely important example of an interface type is the empty interface:

```
interface{}
```

It represents the empty set of methods and is satisfiled by value at all, since any value has zero or more methods.

Some people say that Go's interfaces are dynamically typed, but that is misleading. They are statically typed: a variable of interface type always has the same static type, and even though at run time the value stored in the interface variable may change type, that value will always satisfy the interface.

We need to be *precise* about all this because reflection and interfaces are closely related.(反射通常是用来反射interface类型的)

The representation of an interface
---

Russ Cox has written a [detailed blog post](http://research.swtch.com/interfaces) about the representation of interface values in Go. It's not necessary to repeat the full story here, but a simplified summary is in order.

A variable of interface type stores a pair: the concrete(具体) value assigned to the variable, and that value's type descriptor. To be more precise, the value is the underlying concrete data item that implements the interface and the type describes the full type of that item. For instance, after

```
var r io.Reader
tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
if err != nil {
    return nil, err
}
r = tty
```

r contains, schematically, the (value, type) pair, (tty, \*os.File). Notice that the type \*os.File implements methods other than Read; even though the interface value provides access only to the Read mehtod, the value inside carries all the type information about that value. That's why we can do things like this:
```
var w io.Writer
w = r.(io.Writer)
```
The expression in this assignment is a *type assertion*; that is asserts is that the item insider r also implements io.Writer, and so we can assign it to w. After the assignment, w will contain the pair(tty, \*os.File). That's the same pair as was held in r. The static type of the interface determines what methods may be invoked with an interface variable, even though the concrete value inside may have a larger set of methods.

Continuing, we can do this:
```
var empty interface{}
empty = w
```
and out empty interface value empty will again contain that same pair, (tty, \*os.File). That's handy: an empty interface can hold any value and contains all the information we could ever need aboute that value.

(We don't need a type assertion here because it's known statically that w satisfies the empty interface. In the example where we moved a value from a Reader to a Write, we needed to be explicit and use a type assertion because `Writer`'s methods are not a subset of `Reader`'s.)

One important detail is that the pair inside an interface always has the form (value, concrete type) and connot have the form (value, interface type). *Interfaces do not hold interface values*.

Now we're ready to reflect.

The first law of reflect
----

1. Reflection goes from interface value to reflection object.
----

At the basic level, reflection is just a mechanism to examine the type and value pair stored inside an interface variable. To get started, there are two types we need to know aboute in package reflect: Type and Value. Those two types give access to the contents of an interface variable, and two simple functions, called reflect.TypeOf and reflect.ValueOf, retrieve reflect.Type and reflect.Value pieces out of an interface value. (Also, from the reflect.Value it's easy to get to the
reflect.Type, but let's keep the Value and Type concepts separate for now.)

Let's start with TypeOf:
```
package maino

import (
    "fmt"
    "reflect"
)

func main() {
    var x float64 = 3.4
    fmt.Println("type:", reflect.TypeOf(x))
}
```

This program prints
```
type: float64
```

You might be wondering where the interface is here, since the program looks like it's passing the float64 variable x, not an interface value(我当时就这么想...), to reflect.TypeOf. But it's there; as godoc reports, the signature of reflect.TypeOf includes an empty interface:
```
// TypeOf returns the reflection Type of the value in the interface{}.
func TypeOf(i interface{}) Type
```

When we call reflect.TypeOf(x), x is first stored in an empty interface, which is then passed as the argument; reflect.TypeOf unpacks that empty interface to recover the type information.

The reflect.ValueOf function, of course, recovers the value (from here on we'll elide the boilerplate(样板) and focus just on the executable code):

```
var x float64 = 3.4
fmt.Println("value:", reflect.ValueOf(x))
```

prints:
```
value: <float64 Value>
```

There are also methods like SetInt and SetFloat but to use them we need to understand settability, the subject of the third law of reflection, discussed below.

The reflection library has a couple of properties worth *singling out*(单独挑出来说). First, to keep the API simple, the "getter" and "setter" methods of Value operate on the largest type that can hold the value: int64 for all the signed integers, for instance. That is, the Int method of Value returns an int64 and the SetInt value takes an int64; it may be necessary to convert to the actual type involved:
```
var x uint8 = 'x'
v := reflect.ValueOf(x)
fmt.Println("type:", v.Type())
fmt.Println("kind is uint8: ", v.Kind() == reflect.Uint8)
x = uint8(v.Uint())
```

The second property is that the Kind of a reflection object describes the underlying type, not the static type. If a reflection object contains a value of a user-defined integer type, as in 
```
type MyInt int
var x MyInt = 7
v := reflect.ValueOf(x)
```

the Kind of v is still reflect.Int, even though the static type of x is MyInt, not int. In other words, the Kind cannot discriminate an int from a MyInt even though the Type can.

The second law of reflection
----

2. Reflection goes from reflection object to interface value.
----

Like physical reflection, reflection in Go generates its own inverse.

Give a reflect.Value we can recover an interface value using the interface mehtod; in effect the method packs the type and value information back into an interface representation and returns the result:
