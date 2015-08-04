package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {
	n := 1000
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			log.Println(randomStr(26))
			wg.Done()
		}()
	}
	wg.Wait()
}

func randomStr(l int64) string {
	rand.Seed(time.Now().UnixNano())
	str := "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, l)
	var i int64
	for i = 0; i < l; i++ {
		var index = rand.Intn(int(l))
		result[i] = str[index]
	}
	return string(result)
}
