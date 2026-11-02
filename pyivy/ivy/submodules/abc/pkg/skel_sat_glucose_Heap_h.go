package abc

// Hand-checked skeleton of src/sat/glucose/Heap.h (namespace Gluco).

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
func glucoHeapLeft(i int) int {
	panic("TODO, stubbed")
}

// glucoHeapRight is Heap::right (private static inline).
// C: static inline int right(int i)
func glucoHeapRight(i int) int {
	panic("TODO, stubbed")
}

// glucoHeapParent is Heap::parent (private static inline).
// C: static inline int parent(int i)
func glucoHeapParent(i int) int {
	panic("TODO, stubbed")
}

// percolateUp is Heap::percolateUp (private).
// C: void percolateUp(int i)
func (h *GlucoHeap[Comp]) percolateUp(i int) {
	panic("TODO, stubbed")
}

// percolateDown is Heap::percolateDown (private).
// C: void percolateDown(int i)
func (h *GlucoHeap[Comp]) percolateDown(i int) {
	panic("TODO, stubbed")
}

// NewGlucoHeap is Heap(const Comp& c).
// C: Heap(const Comp& c)
func NewGlucoHeap[Comp GlucoHeapLess](c Comp) *GlucoHeap[Comp] {
	panic("TODO, stubbed")
}

// Size is Heap::size.
// C: int size() const
func (h *GlucoHeap[Comp]) Size() int {
	panic("TODO, stubbed")
}

// Empty is Heap::empty.
// C: bool empty() const
func (h *GlucoHeap[Comp]) Empty() bool {
	panic("TODO, stubbed")
}

// InHeap is Heap::inHeap.
// C: bool inHeap(int n) const
func (h *GlucoHeap[Comp]) InHeap(n int) bool {
	panic("TODO, stubbed")
}

// At is operator[] (const).
// C: int operator[](int index) const
func (h *GlucoHeap[Comp]) At(index int) int {
	panic("TODO, stubbed")
}

// Decrease is Heap::decrease.
// C: void decrease(int n)
func (h *GlucoHeap[Comp]) Decrease(n int) {
	panic("TODO, stubbed")
}

// Increase is Heap::increase.
// C: void increase(int n)
func (h *GlucoHeap[Comp]) Increase(n int) {
	panic("TODO, stubbed")
}

// Update is Heap::update.
// C: void update(int n)
func (h *GlucoHeap[Comp]) Update(n int) {
	panic("TODO, stubbed")
}

// Insert is Heap::insert.
// C: void insert(int n)
func (h *GlucoHeap[Comp]) Insert(n int) {
	panic("TODO, stubbed")
}

// RemoveMin is Heap::removeMin.
// C: int removeMin()
func (h *GlucoHeap[Comp]) RemoveMin() int {
	panic("TODO, stubbed")
}

// Build is Heap::build.
// C: void build(vec<int>& ns)
func (h *GlucoHeap[Comp]) Build(ns *GlucoVec[int]) {
	panic("TODO, stubbed")
}

// Clear is Heap::clear (dealloc defaults to false).
// C: void clear(bool dealloc = false)
func (h *GlucoHeap[Comp]) Clear(dealloc bool) {
	panic("TODO, stubbed")
}
