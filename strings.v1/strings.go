//提供了一些简单的方法用于操作stirng
package strings

import "unicode/utf8"

//分割字符串
func explode(s string, n int) []string {
	if n == 0 {
		return nil
	}
	l := utf8.RuneCountInString(s)
	if n <= 0 || n > l {
		n = l
	}
	a := make([]string, n)
	var size int
	var ch rune
	i, cur := 0, 0
	for ; i+1 < n; i++ {
		ch, size = utf8.DecodeRuneInString(s[cur:])
		if ch == utf8.RuneError {
			a[i] = string(utf8.RuneError)
		} else {
			a[i] = s[cur : cur+size]
		}
		cur += size
	}
	if cur < len(s) {
		a[i] = s[cur:]
	}
	return a
}

// Rabin-Karp algorithm
const primePK = 16777619

// 一直看成了hasStr...
func hashStr(sep string) (uint32, uint32) {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*primePK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, primePK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

func hashStrRev(sep string) (uint32, uint32) {
	hash := uint32(0)
	for i := len(sep) - 1; i >= 0; i-- {
		hash = hash*primePK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, primePK
	for i := len(sep); i > 0; i >>= 1 {
		if 1&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}
