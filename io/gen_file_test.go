package main

import "testing"

func Benchmark_randomStr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		randomStr(26)
	}
}
