package expvar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
)

type Var interface {
	String() string
}

type Int struct {
	i int64
}

func (v *Int) String() string {
	return strconv.FormatInt(atomic.LoadInt64(&v.i), 10)
}

func (v *Int) Add(delta int64) {
	atomic.AddInt64(&v.i, delta)
}

func (v *Int) Set(value int64) {
	atomic.StoreInt64(&v.i, value)
}

type Float struct {
	f uint64
}

func (v *Float) String() string {
	return strconv.FormatFloat(math.Float64frombits(atomic.LoadUint64(&v.f)), 'g', -1, 64)
}

func (v *Float) Add(delta float64) {
	for {
		cur := atomic.LoadUint64(&v.f)
		curVal := math.Float64frombits(cur)
		nxtVal := curVal + delta
		nxt := math.Float64bits(nxtVal)
		if atomic.CompareAndSwapUint64(&v.f, cur, nxt) {
			return
		}
	}
}

func (v *Float) Set(value float64) {
	atomic.StoreUint64(&v.f, math.Float64bits(value))
}

type Map struct {
	mu   sync.RWMutex
	m    map[string]Var
	keys []string
}

type KeyValue struct {
	Key   string
	Value Var
}

func (v *Map) String() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	var b bytes.Buffer
	fmt.Fprintf(&b, "{")
	first := true
	v.doLocked(func(kv KeyValue) {
		if !first {
			fmt.Fprintf(&b, ", ")
		}
		fmt.Fprintf(&b, "%q: %v", kv.Key, kv.Value)
		first = false
	})
	fmt.Fprintf(&b, "}")
	return b.String()
}

func (v *Map) Init() *Map {
	v.m = make(map[string]Var)
	return v
}

func (v *Map) updateKeys() {
	if len(v.m) == len(v.keys) {
		return
	}
	v.keys = v.keys[:0]
	for k := range v.m {
		v.keys = append(v.keys, k)
	}
	sort.Strings(v.keys)
}

func (v *Map) Get(key string, av Var) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.m[key] = av
	v.updateKeys()
}

func (v *Map) Add(key string, delta int64) {
	v.mu.RLock()
	av, ok := v.m[key]
	v.mu.RUnlock()
	if !ok {
		v.mu.Lock()
		av, ok = v.m[key]
		if !ok {
			av = new(Int)
			v.m[key] = av
			v.updateKeys()
		}
		v.mu.Unlock()
	}

	if iv, ok := av.(*Int); ok {
		iv.Add(delta)
	}
}

func (v *Map) AddFloat(key string, delta float64) {
	v.mu.RLock()
	av, ok := v.m[key]
	v.mu.RUnlock()
	if !ok {
		v.mu.Lock()
		av, ok = v.m[key]
		if !ok {
			av = new(Float)
			v.m[key] = av
			v.updateKeys()
		}
		v.mu.Unlock()
	}

	if iv, ok := av.(*Float); ok {
		iv.Add(delta)
	}
}

func (v *Map) Do(f func(KeyValue)) {
	v.mu.RLock()
	defer v.mu.RUnlock()
	v.doLocked(f)
}

func (v *Map) doLocked(f func(KeyValue)) {
	for _, k := range v.keys {
		f(KeyValue{k, v.m[k]})
	}
}

type String struct {
	mu sync.RWMutex
	s  string
}

func (v *String) String() string {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return strconv.Quote(v.s)
}

func (v *String) Set(value string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.s = value
}

type Func func() interface{}

func (f Func) String() string {
	v, _ := json.Marshal(f())
	return string(v)
}

var (
	mutex   sync.RWMutex
	vars    = make(map[string]Var)
	varKeys []string
)

func Publish(name string, v Var) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, existing := vars[name]; existing {
		log.Panicln("Reuse of exported var name:", name)
	}
	vars[name] = v
	varKeys = append(varKeys, name)
	sort.Strings(varKeys)
}

func Get(name string) Var {
	mutex.RLock()
	defer mutex.RUnlock()
	return vars[name]
}

func NewInt(name string) *Int {
	v := new(Int)
	Publish(name, v)
	return v
}

func NewFloat(name string) *Float {
	v := new(Float)
	Publish(name, v)
	return v
}

func NewMap(name string) *Map {
	v := new(Map).Init()
	Publish(name, v)
	return v
}

func NewString(name string) *String {
	v := new(String)
	Publish(name, v)
	return v
}

func Do(f func(KeyValue)) {
	mutex.RLock()
	defer mutex.RUnlock()
	for _, k := range varKeys {
		f(KeyValue{k, vars[k]})
	}
}

func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	Do(func(kv KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}

func cmdline() interface{} {
	return os.Args
}

func memstats() interface{} {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	return *stats
}

func init() {
	http.HandleFunc("/debug/vars", expvarHandler)
	Publish("cmdline", Func(cmdline))
	Publish("memstats", Func(memstats))
}
