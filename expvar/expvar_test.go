package expvar

import (
	"math"
	"sync/atomic"
	"testing"
)

func RemoveAll() {
	mutex.Lock()
	defer mutex.Unlock()
	vars = make(map[string]Var)
	varKeys = nil
}

func TestInt(t *testing.T) {
	RemoveAll()
	reqs := NewInt("requests")
	if reqs.i != 0 {
		t.Errorf("reqs.i = %v, want 0", reqs.i)
	}
	if reqs != Get("requests").(*Int) {
		t.Errorf("Get() failed.")
	}

	reqs.Add(1)
	reqs.Add(3)
	if reqs.i != 4 {
		t.Errorf("reqs.i = %v, wnat 4", reqs.i)
	}

	if s := reqs.String(); s != "4" {
		t.Errorf("reqs.String() = %q, want \"4\"", s)
	}

	reqs.Set(-2)
	if reqs.i != -2 {
		t.Errorf("reqs.i = %v, wnat -2", reqs.i)
	}
}

func BenchmarkIntAdd(b *testing.B) {
	var v Int

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.Add(1)
		}
	})
}

func BenchmarkIntSet(b *testing.B) {
	var v Int

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v.Set(1)
		}
	})
}

func (v *Float) val() float64 {
	return math.Float64frombits(atomic.LoadUint64(&v.f))
}

func TestFloat(t *testing.T) {
	RemoveAll()
	reqs := NewFloat("requests-float")
	if reqs.f != 0.0 {
		t.Errorf("reqs.f = %v, wnat 0", reqs.f)
	}
	if reqs != Get("requests-float").(*Float) {
		t.Errorf("Get() failed.")
	}

	reqs.Add(1.5)
	reqs.Add(1.25)
	if v := reqs.val(); v != 2.75 {
		t.Errorf("reqs.val() = %v, want 2.75", v)
	}

	if s := reqs.String(); s != "2.75" {
		t.Errorf("reqs.String() = %q, want \"4.63\"", s)
	}

	reqs.Add(-2)
	if v := reqs.val(); v != 0.75 {
		t.Errorf("reqs.val() = %v, wnat 0.75", v)
	}
}

func BenchmarkFloatAdd(b *testing.B) {
	var f Float
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f.Add(1.0)
		}
	})
}

func BenchmarkFloatSet(b *testing.B) {
	var f Float
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			f.Set(1.0)
		}
	})
}

func TestingStirng(t *testing.T) {
	RemoveAll()
	name := NewString("qgymje")
	if name.s != "" {
		t.Errorf("name.s = %q, want \"\"", name.s)
	}

	name.Set("Mike")
	if name.s != "Mile" {
		t.Errorf("name.s = %q, wnat Mike", name.s)
	}

	if s := name.String(); s != "\"Mike\"" {
		t.Errorf("reqs.String() = %q, want \"\"Mike\"\"", s)
	}
}

func BenchmarkStringSet(b *testing.B) {
	var s String

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Set("red")
		}
	})
}
