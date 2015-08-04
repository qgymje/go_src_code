// Package heap provides heap operations for any type that implements
// 这是一个最小堆结构
// heap.Interface. A heap is a tree with the property that each node is the
// 任何树结构都是链表结构的扩展
// minimum-valued node in its subtree.
// The minimum element in the tree is the root, at index 0.
// 最小堆的index为0的是最小数
// A heap is a common way to implement a proiority queue. To build a priority
// 优先队列结构
// queue, implement the Heap interface with the (negative) priority as the
// ordering for the Less method, so Push adds items while Pop removes the
// highest-priority item from the queue. The examples include such an
// implementation; the file example_pq_test.go has the complete source.
package heap

import "sort"

// Any type that implements heap.Interface may be used as a
// min-heap with the following invariants (established after
// Init has been called or if the data is empty or sorted):
//
// !h.Less(j, i) for 0 <= i < h.Len() and 2*i+1 <= j <= 2*i+2 and j < h.Len()
//
// Note that Push and Pop in this interface are for package heap's
// implementation to call. To add and remove things from the heap,
// use heap.Push and heap.Pop.
type Interface interface {
	sort.Interface
	Push(x interface{}) // add x as element Len()
	Pop() interface{}   // remove and return element Len() -1.
}

// A heap must be initialized before any of the heap operations
// 先要调用些方法初始化
// can be used. Init is idempotent with respect ot the heap invariants
// and may be called whenever the heap invariants may have been invalidated.
// Its complexity is O(n) when n - h.Len().
// 时间复杂度
func Init(h Interface) {
	n := h.Len()
	for i := n/2 - 1; i >= 0; i-- {
		down(h, i, n)
	}
}

// Push pushes the element x onto the heap. The complexity is
// O(log(n)) where n = h.Len()
//
func Push(h Interface, x interface{}) {
	h.Push(x)
	up(h, h.Len()-1)
}

// Pop removes the minimum element (according to Less) from the heap
// and returns it. The complexity is O(log(n)) where n = h.Len().
// It is equivalent to Remove(h, 0).
func Pop(h Interface) interface{} {
	n := h.Len() - 1
	h.Swap(0, n)
	down(h, 0, n)
	return h.Pop()
}

// Remove removes the element as index i from the heap.
// The complexity is O(log(n)) when n = h.Len().
func Remove(h Interface, i int) interface{} {
	n := h.Len() - 1
	if n != i {
		h.Swap(i, n)
		down(h, i, n)
		up(h, i)
	}
	return h.Pop()
}

func Fix(h Interface, i int) {
	down(h, i, h.Len())
	up(h, i)
}

func up(h Interface, j int) {
	for {
		i := (j - 1) / 2
		if i == j || !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		j = i
	}
}

func down(h Interface, i, n int) {
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 {
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && !h.Less(j1, j2) {
			j = j2 //right child
		}
		if !h.Less(j, i) {
			break
		}
		h.Swap(i, j)
		i = j
	}
}
