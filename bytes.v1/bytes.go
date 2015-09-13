// Package bytes实现了用于操作byte slice的方法
// 与strings包提供的方法差不多
package bytes

import "unicode/utf8"

//判断两个[]byte是否一致
func equalPortable(a, b []byte) bool {
	if len(a) != len(b) { //先判断长度是否一致
		return false
	}
	for i, c := range a { //在长度一致的情况下,遍历比较
		if c != b[i] {
			return false
		}
	}
	return true
}

func explode(s []byte, n int) [][]byte {
	if n <= 0 {
		n = len(s)
	}
	a := make([][]byte, n)
	var size int
	na := 0
	for len(s) > 0 {
		if na+1 >= n {
			a[na] = s
			na++
			break
		}
		_, size = utf8.DecodeRune(s)
		a[na] = s[0:size]
		s = s[size:]
		na++
	}
	return a[0:na]
}

func Count(s, sep []byte) int {
	n := len(sep)
	if n == 0 {
		return utf8.RuneCount(s) + 1
	}
	if n > len(s) {
		return 0
	}
	count := 0
	c := sep[0]
	i := 0
	t := s[:len(s)-n+1]
	for i < len(t) {
		if t[i] != c {
			o := IndexByte(t[i:], c)
			if o < 0 {
				break
			}
			i += o
		}
		if n == 1 || Equal(s[i:i+n], sep) {
			count++
			i += n
			continue
		}
		i++
	}
	return count
}

func Contains(b, subslice []byte) bool {
	return Index(b, subslice) != -1
}

func Index(s, sep []byte) int {
	n := len(sep)
	if n == 0 {
		return 0
	}
	if n > len(s) {
		return -1
	}
	c := sep[0]
	if n == 1 {
		return IndexByte(s, c)
	}
	i := 0
	t := s[:len(s)-n+1]
	for i < len(t) {
		if t[i] != c {
			o := IndexByte(t[i:], c)
			if o < 0 {
				break
			}
			i += o
		}
		if Equal(s[i:i+n], sep) {
			return i
		}
		i++
	}
	return -1
}

func indexByteProtable(s []byte, c byte) int {
	for i, b := range s {
		if b == c {
			return i
		}
	}
	return -1
}

func LastIndex(s, sep []byte) int {
	n := len(sep)
	if n == 0 {
		return len(s)
	}
	c := sep[0]
	for i := len(s) - n; i >= 0; i-- {
		if s[i] == c && (n == 1 || Equal(s[i:i+n], sep)) {
			return i
		}
	}
	return -1
}

func IndexRune(s []byte, r rune) int {
	for i := 0; i < len(s); {
		r1, size := utf8.DecodeRune(s[i:])
		if r == r1 {
			return i
		}
		i += size
	}
	return -1
}

func IndexAny(s []byte, chars string) int {
	if len(chars) > 0 {
		var r rune
		var width int
		for i := 0; i < len(s); i += width {
			r = rune(s[i])
			if r < utf8.RuneSelf {
				width = 1
			} else {
				r, width = utf8.DecodeRune(s[i:])
			}
			for _, ch := range chars {
				if r == ch {
					return i
				}
			}
		}
	}
	return -1
}

func LastIndexAny(s []byte, chars string) int {
	if len(chars) > 0 {
		for i := len(s); i > 0; {
			r, size := utf8.DecodeLastRune(s[0:i])
			i -= size
			for _, ch := range chars {
				if r == ch {
					return i
				}
			}
		}
	}
	return -1
}
