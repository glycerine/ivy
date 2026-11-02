package abc

// Hand-checked skeleton of src/sat/bsat2/Heap.h (namespace Minisat).

// MinisatHeapLess is the requirement on Heap's template argument Comp: a
// comparator object with bool operator()(int, int).
type MinisatHeapLess interface {
	Call(x int, y int) bool
}

// MinisatHeap is Minisat::Heap<Comp>: a binary min-heap of ints with respect to Comp.
type MinisatHeap[Comp MinisatHeapLess] struct {
	lt      Comp            // the heap is a minimum-heap with respect to this comparator
	heap    MinisatVec[int] // heap of integers
	indices MinisatVec[int] // each integer's position (index) in the heap
}

// minisatHeapLeft is Heap::left (private static inline).
// C: static inline int left(int i)
func minisatHeapLeft(i int) int {
	panic("TODO, stubbed")
}

// minisatHeapRight is Heap::right (private static inline).
// C: static inline int right(int i)
func minisatHeapRight(i int) int {
	panic("TODO, stubbed")
}

// minisatHeapParent is Heap::parent (private static inline).
// C: static inline int parent(int i)
func minisatHeapParent(i int) int {
	panic("TODO, stubbed")
}

// percolateUp is Heap::percolateUp (private).
// C: void percolateUp(int i)
func (h *MinisatHeap[Comp]) percolateUp(i int) {
	panic("TODO, stubbed")
}

// percolateDown is Heap::percolateDown (private).
// C: void percolateDown(int i)
func (h *MinisatHeap[Comp]) percolateDown(i int) {
	panic("TODO, stubbed")
}

// NewMinisatHeap is Heap(const Comp& c).
// C: Heap(const Comp& c)
func NewMinisatHeap[Comp MinisatHeapLess](c Comp) *MinisatHeap[Comp] {
	panic("TODO, stubbed")
}

// Size is Heap::size.
// C: int size() const
func (h *MinisatHeap[Comp]) Size() int {
	panic("TODO, stubbed")
}

// Empty is Heap::empty.
// C: bool empty() const
func (h *MinisatHeap[Comp]) Empty() bool {
	panic("TODO, stubbed")
}

// InHeap is Heap::inHeap.
// C: bool inHeap(int n) const
func (h *MinisatHeap[Comp]) InHeap(n int) bool {
	panic("TODO, stubbed")
}

// At is operator[] (const).
// C: int operator[](int index) const
func (h *MinisatHeap[Comp]) At(index int) int {
	panic("TODO, stubbed")
}

// Decrease is Heap::decrease.
// C: void decrease(int n)
func (h *MinisatHeap[Comp]) Decrease(n int) {
	panic("TODO, stubbed")
}

// Increase is Heap::increase.
// C: void increase(int n)
func (h *MinisatHeap[Comp]) Increase(n int) {
	panic("TODO, stubbed")
}

// Update is Heap::update.
// C: void update(int n)
func (h *MinisatHeap[Comp]) Update(n int) {
	panic("TODO, stubbed")
}

// Insert is Heap::insert.
// C: void insert(int n)
func (h *MinisatHeap[Comp]) Insert(n int) {
	panic("TODO, stubbed")
}

// RemoveMin is Heap::removeMin.
// C: int removeMin()
func (h *MinisatHeap[Comp]) RemoveMin() int {
	panic("TODO, stubbed")
}

// Build is Heap::build.
// C: void build(vec<int>& ns)
func (h *MinisatHeap[Comp]) Build(ns *MinisatVec[int]) {
	panic("TODO, stubbed")
}

// Clear is Heap::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (h *MinisatHeap[Comp]) Clear(dealloc bool) {
	panic("TODO, stubbed")
}
