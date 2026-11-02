package abc

// Port of src/sat/glucose/Heap.h (namespace Gluco).

// GlucoHeapLess is the requirement on Heap's template argument Comp: a
// comparator object with bool operator()(int, int).
type GlucoHeapLess interface {
	Call(x int, y int) bool
}

// GlucoHeap is Gluco::Heap<Comp>: a binary min-heap of ints with respect to Comp.
type GlucoHeap[Comp GlucoHeapLess] struct {
	lt      Comp          // the heap is a minimum-heap with respect to this comparator
	heap    GlucoVec[int] // heap of integers
	indices GlucoVec[int] // each integer's position (index) in the heap
}

// glucoHeapLeft is Heap::left (private static inline).
// C: static inline int left(int i)
func glucoHeapLeft(i int) int { return i*2 + 1 }

// glucoHeapRight is Heap::right (private static inline).
// C: static inline int right(int i)
func glucoHeapRight(i int) int { return (i + 1) * 2 }

// glucoHeapParent is Heap::parent (private static inline).
// C: static inline int parent(int i)
func glucoHeapParent(i int) int { return (i - 1) >> 1 }

// percolateUp is Heap::percolateUp (private).
// C: void percolateUp(int i)
func (h *GlucoHeap[Comp]) percolateUp(i int) {
	x := *h.heap.At(i)
	p := glucoHeapParent(i)
	for i != 0 && h.lt.Call(x, *h.heap.At(p)) {
		*h.heap.At(i) = *h.heap.At(p)
		*h.indices.At(*h.heap.At(p)) = i
		i = p
		p = glucoHeapParent(p)
	}
	*h.heap.At(i) = x
	*h.indices.At(x) = i
}

// percolateDown is Heap::percolateDown (private).
// C: void percolateDown(int i)
func (h *GlucoHeap[Comp]) percolateDown(i int) {
	x := *h.heap.At(i)
	for glucoHeapLeft(i) < h.heap.Size() {
		child := glucoHeapLeft(i)
		if glucoHeapRight(i) < h.heap.Size() && h.lt.Call(*h.heap.At(glucoHeapRight(i)), *h.heap.At(glucoHeapLeft(i))) {
			child = glucoHeapRight(i)
		}
		if !h.lt.Call(*h.heap.At(child), x) {
			break
		}
		*h.heap.At(i) = *h.heap.At(child)
		*h.indices.At(*h.heap.At(i)) = i
		i = child
	}
	*h.heap.At(i) = x
	*h.indices.At(x) = i
}

// NewGlucoHeap is Heap(const Comp& c).
// C: Heap(const Comp& c)
func NewGlucoHeap[Comp GlucoHeapLess](c Comp) *GlucoHeap[Comp] {
	return &GlucoHeap[Comp]{lt: c}
}

// Size is Heap::size.
// C: int size() const
func (h *GlucoHeap[Comp]) Size() int { return h.heap.Size() }

// Empty is Heap::empty.
// C: bool empty() const
func (h *GlucoHeap[Comp]) Empty() bool { return h.heap.Size() == 0 }

// InHeap is Heap::inHeap.
// C: bool inHeap(int n) const
func (h *GlucoHeap[Comp]) InHeap(n int) bool {
	return n < h.indices.Size() && *h.indices.At(n) >= 0
}

// At is operator[] (const).
// C: int operator[](int index) const
func (h *GlucoHeap[Comp]) At(index int) int {
	if index >= h.heap.Size() {
		panic("GlucoHeap.At: index out of range")
	}
	return *h.heap.At(index)
}

// Decrease is Heap::decrease.
// C: void decrease(int n)
func (h *GlucoHeap[Comp]) Decrease(n int) {
	if !h.InHeap(n) {
		panic("GlucoHeap.Decrease: not in heap")
	}
	h.percolateUp(*h.indices.At(n))
}

// Increase is Heap::increase.
// C: void increase(int n)
func (h *GlucoHeap[Comp]) Increase(n int) {
	if !h.InHeap(n) {
		panic("GlucoHeap.Increase: not in heap")
	}
	h.percolateDown(*h.indices.At(n))
}

// Update is Heap::update (call after a change in either direction).
// C: void update(int n)
func (h *GlucoHeap[Comp]) Update(n int) {
	if !h.InHeap(n) {
		h.Insert(n)
	} else {
		h.percolateUp(*h.indices.At(n))
		h.percolateDown(*h.indices.At(n))
	}
}

// Insert is Heap::insert.
// C: void insert(int n)
func (h *GlucoHeap[Comp]) Insert(n int) {
	h.indices.GrowToPad(n+1, -1)
	if h.InHeap(n) {
		panic("GlucoHeap.Insert: already in heap")
	}
	*h.indices.At(n) = h.heap.Size()
	h.heap.PushElem(n)
	h.percolateUp(*h.indices.At(n))
}

// RemoveMin is Heap::removeMin.
// C: int removeMin()
func (h *GlucoHeap[Comp]) RemoveMin() int {
	x := *h.heap.At(0)
	*h.heap.At(0) = *h.heap.Last()
	*h.indices.At(*h.heap.At(0)) = 0
	*h.indices.At(x) = -1
	h.heap.Pop()
	if h.heap.Size() > 1 {
		h.percolateDown(0)
	}
	return x
}

// Build is Heap::build (build a heap of the variables in 'ns' in linear time).
// C: void build(vec<int>& ns)
func (h *GlucoHeap[Comp]) Build(ns *GlucoVec[int]) {
	var i int
	for i = 0; i < h.heap.Size(); i++ {
		*h.indices.At(*h.heap.At(i)) = -1
	}
	h.heap.Clear(false)
	for i = 0; i < ns.Size(); i++ {
		*h.indices.At(*ns.At(i)) = i
		h.heap.PushElem(*ns.At(i))
	}
	for i = h.heap.Size()/2 - 1; i >= 0; i-- {
		h.percolateDown(i)
	}
}

// Clear is Heap::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (h *GlucoHeap[Comp]) Clear(dealloc bool) {
	for i := 0; i < h.heap.Size(); i++ {
		*h.indices.At(*h.heap.At(i)) = -1
	}
	h.heap.Clear(dealloc)
}
