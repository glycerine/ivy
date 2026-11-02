package abc

// Port of src/sat/glucose2/Heap.h (namespace Gluco2).

// Gluco2HeapLess is the requirement on Heap's template argument Comp: a
// comparator object with bool operator()(int, int).
type Gluco2HeapLess interface {
	Call(x int, y int) bool
}

// Gluco2Heap is Gluco2::Heap<Comp>: a binary min-heap of ints with respect to Comp.
type Gluco2Heap[Comp Gluco2HeapLess] struct {
	lt      Comp           // the heap is a minimum-heap with respect to this comparator
	heap    Gluco2Vec[int] // heap of integers
	indices Gluco2Vec[int] // each integer's position (index) in the heap
}

// gluco2HeapLeft is Heap::left (private static inline).
// C: static inline int left(int i)
func gluco2HeapLeft(i int) int { return i*2 + 1 }

// gluco2HeapRight is Heap::right (private static inline).
// C: static inline int right(int i)
func gluco2HeapRight(i int) int { return (i + 1) * 2 }

// gluco2HeapParent is Heap::parent (private static inline).
// C: static inline int parent(int i)
func gluco2HeapParent(i int) int { return (i - 1) >> 1 }

// percolateUp is Heap::percolateUp (private).
// C: void percolateUp(int i)
func (h *Gluco2Heap[Comp]) percolateUp(i int) {
	x := *h.heap.At(i)
	p := gluco2HeapParent(i)
	for i != 0 && h.lt.Call(x, *h.heap.At(p)) {
		*h.heap.At(i) = *h.heap.At(p)
		*h.indices.At(*h.heap.At(p)) = i
		i = p
		p = gluco2HeapParent(p)
	}
	*h.heap.At(i) = x
	*h.indices.At(x) = i
}

// percolateDown is Heap::percolateDown (private).
// C: void percolateDown(int i)
func (h *Gluco2Heap[Comp]) percolateDown(i int) {
	x := *h.heap.At(i)
	for gluco2HeapLeft(i) < h.heap.Size() {
		child := gluco2HeapLeft(i)
		if gluco2HeapRight(i) < h.heap.Size() && h.lt.Call(*h.heap.At(gluco2HeapRight(i)), *h.heap.At(gluco2HeapLeft(i))) {
			child = gluco2HeapRight(i)
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

// NewGluco2Heap is Heap(const Comp& c).
// C: Heap(const Comp& c)
func NewGluco2Heap[Comp Gluco2HeapLess](c Comp) *Gluco2Heap[Comp] {
	return &Gluco2Heap[Comp]{lt: c}
}

// Size is Heap::size.
// C: int size() const
func (h *Gluco2Heap[Comp]) Size() int { return h.heap.Size() }

// Empty is Heap::empty.
// C: bool empty() const
func (h *Gluco2Heap[Comp]) Empty() bool { return h.heap.Size() == 0 }

// InHeap is Heap::inHeap.
// C: bool inHeap(int n) const
func (h *Gluco2Heap[Comp]) InHeap(n int) bool {
	return n < h.indices.Size() && *h.indices.At(n) >= 0
}

// At is operator[] (const).
// C: int operator[](int index) const
func (h *Gluco2Heap[Comp]) At(index int) int {
	if index >= h.heap.Size() {
		panic("Gluco2Heap.At: index out of range")
	}
	return *h.heap.At(index)
}

// Decrease is Heap::decrease.
// C: void decrease(int n)
func (h *Gluco2Heap[Comp]) Decrease(n int) {
	if !h.InHeap(n) {
		panic("Gluco2Heap.Decrease: not in heap")
	}
	h.percolateUp(*h.indices.At(n))
}

// Increase is Heap::increase.
// C: void increase(int n)
func (h *Gluco2Heap[Comp]) Increase(n int) {
	if !h.InHeap(n) {
		panic("Gluco2Heap.Increase: not in heap")
	}
	h.percolateDown(*h.indices.At(n))
}

// Update is Heap::update (call after a change in either direction).
// C: void update(int n)
func (h *Gluco2Heap[Comp]) Update(n int) {
	if !h.InHeap(n) {
		h.Insert(n)
	} else {
		h.percolateUp(*h.indices.At(n))
		h.percolateDown(*h.indices.At(n))
	}
}

// Insert is Heap::insert.
// C: void insert(int n)
func (h *Gluco2Heap[Comp]) Insert(n int) {
	h.indices.GrowToPad(n+1, -1)
	if h.InHeap(n) {
		panic("Gluco2Heap.Insert: already in heap")
	}
	*h.indices.At(n) = h.heap.Size()
	h.heap.PushElem(n)
	h.percolateUp(*h.indices.At(n))
}

// RemoveMin is Heap::removeMin.
// C: int removeMin()
func (h *Gluco2Heap[Comp]) RemoveMin() int {
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
func (h *Gluco2Heap[Comp]) Build(ns *Gluco2Vec[int]) {
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
func (h *Gluco2Heap[Comp]) Clear(dealloc bool) {
	for i := 0; i < h.heap.Size(); i++ {
		*h.indices.At(*h.heap.At(i)) = -1
	}
	h.heap.Clear(dealloc)
}
