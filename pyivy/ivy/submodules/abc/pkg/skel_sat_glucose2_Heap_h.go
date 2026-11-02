package abc

// Hand-checked skeleton of src/sat/glucose2/Heap.h (namespace Gluco2).

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
func gluco2HeapLeft(i int) int {
	panic("TODO, stubbed")
}

// gluco2HeapRight is Heap::right (private static inline).
// C: static inline int right(int i)
func gluco2HeapRight(i int) int {
	panic("TODO, stubbed")
}

// gluco2HeapParent is Heap::parent (private static inline).
// C: static inline int parent(int i)
func gluco2HeapParent(i int) int {
	panic("TODO, stubbed")
}

// percolateUp is Heap::percolateUp (private).
// C: void percolateUp(int i)
func (h *Gluco2Heap[Comp]) percolateUp(i int) {
	panic("TODO, stubbed")
}

// percolateDown is Heap::percolateDown (private).
// C: void percolateDown(int i)
func (h *Gluco2Heap[Comp]) percolateDown(i int) {
	panic("TODO, stubbed")
}

// NewGluco2Heap is Heap(const Comp& c).
// C: Heap(const Comp& c)
func NewGluco2Heap[Comp Gluco2HeapLess](c Comp) *Gluco2Heap[Comp] {
	panic("TODO, stubbed")
}

// Size is Heap::size.
// C: int size() const
func (h *Gluco2Heap[Comp]) Size() int {
	panic("TODO, stubbed")
}

// Empty is Heap::empty.
// C: bool empty() const
func (h *Gluco2Heap[Comp]) Empty() bool {
	panic("TODO, stubbed")
}

// InHeap is Heap::inHeap.
// C: bool inHeap(int n) const
func (h *Gluco2Heap[Comp]) InHeap(n int) bool {
	panic("TODO, stubbed")
}

// At is operator[] (const).
// C: int operator[](int index) const
func (h *Gluco2Heap[Comp]) At(index int) int {
	panic("TODO, stubbed")
}

// Decrease is Heap::decrease.
// C: void decrease(int n)
func (h *Gluco2Heap[Comp]) Decrease(n int) {
	panic("TODO, stubbed")
}

// Increase is Heap::increase.
// C: void increase(int n)
func (h *Gluco2Heap[Comp]) Increase(n int) {
	panic("TODO, stubbed")
}

// Update is Heap::update.
// C: void update(int n)
func (h *Gluco2Heap[Comp]) Update(n int) {
	panic("TODO, stubbed")
}

// Insert is Heap::insert.
// C: void insert(int n)
func (h *Gluco2Heap[Comp]) Insert(n int) {
	panic("TODO, stubbed")
}

// RemoveMin is Heap::removeMin.
// C: int removeMin()
func (h *Gluco2Heap[Comp]) RemoveMin() int {
	panic("TODO, stubbed")
}

// Build is Heap::build.
// C: void build(vec<int>& ns)
func (h *Gluco2Heap[Comp]) Build(ns *Gluco2Vec[int]) {
	panic("TODO, stubbed")
}

// Clear is Heap::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (h *Gluco2Heap[Comp]) Clear(dealloc bool) {
	panic("TODO, stubbed")
}
