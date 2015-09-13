Errors are values
===

A common point of discusstion among Go programmers, especially those new to the language, is how to handle errors. The conversation often turns into a lament at the number of times the sequence
```
if err != nil {
    return error
}
```
shows up. We recently scanned all the open source projects we could find and discovered that this snippet occurs only once per page or two, less often than some would have you believe. Still, if the perception persists that one must type
```
if err != nil
```
all the time, something must be wrong, and the obvious target is Go itself.
> 如果一直出现error判断, 似乎有些地方出错了.

This is unfortunate, misleading, and easily corrected. Perhaps what is happening is that programmers new to Go ask, "How does one handle errors?", learn this pattern, and stop there. In other languages, one might use a try-catch block or other such mechanism to handle errors. Therefore, the programmer thinks, when I would have used a try-catch in my old language, I will just type if err != nil in Go. Over time the Go code collects many such snippets, and the result feels clumsy.
> 在其它语言里可能使用try-catch来处理错误, 在Go里就是使用if err != nil

Regardless of whether this explanation fits, it is clear that these Go programmers miss a fundamental point about errors: *Errors are values*.
> Errors are values.

Values can be programmed, and since errors are values, errors can be programmed.
> 说是值是可被编程的, 错误是值, 错误是可被编程的.这三段论真是深入骨子里了


Of course a common statement involving an error value is to test whether it is nil, but there are countless other things one can do with an error value, and application of some of those other things can *make your program better*, eliminating much of the boilerplate that arises if every error is checked with a rote if statement.
> 说可被编程的错误让程序更好理解,每一次错误都要处理

Here's a simple example from the bufio package's Scanner type. Its Scan method performs the underlying I/O, which can of course lead to an error. Yet the Scan method does not expose an error at all. Instead, it returns a boolean, and a seprate method, to be run at the end of the scan, reports whether an error occurred. Client code looks like this:
```
scanner := bufio.NewScanner(input)
for scanner.Scan() {
    token := scanner.Text()
    // process token
}
if err := scanner.Err(); err != nil {
    // process the error
}
```
Sure, there is a nil check for an error, but it appears and excutes only once. The Scan method could instead have been defined as
```
func (s *Scanner) Scan() (token []byte, error)
```
and then the example user code might be (depending on how the token is retrieved),

