package abc

import "unsafe"

// Port of src/sat/bsat2/SolverTypes.h (namespace Minisat).
// See sat_bsat2_vec.go for the C++ -> Go conventions.

// MinisatVar is `typedef int Var`.
type MinisatVar = int

// MinisatLit is Minisat::Lit: a literal, x = 2*var + sign (a 32-bit int in C).
type MinisatLit struct {
	x int32
}

// Eq is Lit::operator==.
// C: bool operator==(Lit p) const
func (l MinisatLit) Eq(p MinisatLit) bool { return l.x == p.x }

// Ne is Lit::operator!=.
// C: bool operator!=(Lit p) const
func (l MinisatLit) Ne(p MinisatLit) bool { return l.x != p.x }

// Lt is Lit::operator< (makes p, ~p adjacent in the ordering).
// C: bool operator<(Lit p) const
func (l MinisatLit) Lt(p MinisatLit) bool { return l.x < p.x }

// MinisatMkLit is mkLit (inline; sign defaults to false).
// C: inline Lit mkLit(Var var, bool sign = false)
func MinisatMkLit(v MinisatVar, sign bool) MinisatLit {
	var p MinisatLit
	p.x = int32(v + v)
	if sign {
		p.x++
	}
	return p
}

// MinisatLitNot is `operator ~` on Lit (inline).
// C: inline Lit operator~(Lit p)
func MinisatLitNot(p MinisatLit) MinisatLit { return MinisatLit{x: p.x ^ 1} }

// MinisatLitXor is `operator ^` on (Lit, bool) (inline).
// C: inline Lit operator^(Lit p, bool b)
func MinisatLitXor(p MinisatLit, b bool) MinisatLit {
	if b {
		return MinisatLit{x: p.x ^ 1}
	}
	return MinisatLit{x: p.x}
}

// MinisatSign is Minisat::sign (inline).
// C: inline bool sign(Lit p)
func MinisatSign(p MinisatLit) bool { return p.x&1 != 0 }

// MinisatLitVar is Minisat::var(Lit) (inline); renamed because MinisatVar is the type.
// C: inline int var(Lit p)
func MinisatLitVar(p MinisatLit) int { return int(p.x >> 1) }

// MinisatToIntVar is the overload toInt(Var) (inline).
// C: inline int toInt(Var v)
func MinisatToIntVar(v MinisatVar) int { return v }

// MinisatToIntLit is the overload toInt(Lit) (inline).
// C: inline int toInt(Lit p)
func MinisatToIntLit(p MinisatLit) int { return int(p.x) }

// MinisatToLit is Minisat::toLit (inline).
// C: inline Lit toLit(int i)
func MinisatToLit(i int) MinisatLit { return MinisatLit{x: int32(i)} }

// MinisatLitUndef is `const Lit lit_Undef = { -2 }`.
var MinisatLitUndef = MinisatLit{x: -2}

// MinisatLitError is `const Lit lit_Error = { -1 }`.
var MinisatLitError = MinisatLit{x: -1}

// MinisatLbool is Minisat::lbool: a three-valued boolean (0 true, 1 false, 2 undef).
type MinisatLbool struct {
	value uint8
}

// NewMinisatLboolV is the explicit constructor lbool(uint8_t v).
// C: explicit lbool(uint8_t v)
func NewMinisatLboolV(v uint8) MinisatLbool { return minisatLboolNorm(v) }

// NewMinisatLbool is lbool(): value 0.
// C: lbool()
func NewMinisatLbool() MinisatLbool { return MinisatLbool{} }

// NewMinisatLboolB is the explicit constructor lbool(bool x): value !x.
// C: explicit lbool(bool x)
func NewMinisatLboolB(x bool) MinisatLbool {
	if x {
		return MinisatLbool{value: 0}
	}
	return MinisatLbool{value: 1}
}

// Eq is lbool::operator==.
// C: bool operator==(lbool b) const
func (l MinisatLbool) Eq(b MinisatLbool) bool {
	return ((b.value&2)&(l.value&2)) != 0 || ((b.value&2) == 0 && l.value == b.value)
}

// Ne is lbool::operator!=.
// C: bool operator!=(lbool b) const
func (l MinisatLbool) Ne(b MinisatLbool) bool { return !l.Eq(b) }

// minisatLboolNorm canonicalises an lbool value: C++ lbool::operator== treats every value
// with bit 1 set as l_Undef (e.g. l_Undef ^ true == 3 still equals l_Undef), so the Go port
// stores all of them as 2 and the struct comparison == is then equivalent to operator==.
func minisatLboolNorm(v uint8) MinisatLbool {
	if v&2 != 0 {
		v = 2
	}
	return MinisatLbool{value: v}
}

// Xor is lbool::operator^.
// C: lbool operator^(bool b) const
func (l MinisatLbool) Xor(b bool) MinisatLbool {
	var u uint8
	if b {
		u = 1
	}
	return minisatLboolNorm(l.value ^ u)
}

// And is lbool::operator&&.
// C: lbool operator&&(lbool b) const
func (l MinisatLbool) And(b MinisatLbool) MinisatLbool {
	sel := (l.value << 1) | (b.value << 3)
	v := uint8((uint32(0xF7F755F4) >> sel) & 3)
	return minisatLboolNorm(v)
}

// Or is lbool::operator||.
// C: lbool operator||(lbool b) const
func (l MinisatLbool) Or(b MinisatLbool) MinisatLbool {
	sel := (l.value << 1) | (b.value << 3)
	v := uint8((uint32(0xFCFCF400) >> sel) & 3)
	return minisatLboolNorm(v)
}

// MinisatToIntLbool is the friend toInt(lbool) (inline).
// C: inline int toInt(lbool l)
func MinisatToIntLbool(l MinisatLbool) int { return int(l.value) }

// MinisatToLbool is the friend toLbool(int) (inline).
// C: inline lbool toLbool(int v)
func MinisatToLbool(v int) MinisatLbool { return minisatLboolNorm(uint8(v)) }

// The macros l_True, l_False and l_Undef.
var (
	MinisatLTrue  = MinisatLbool{value: 0} // #define l_True  (lbool((uint8_t)0))
	MinisatLFalse = MinisatLbool{value: 1} // #define l_False (lbool((uint8_t)1))
	MinisatLUndef = MinisatLbool{value: 2} // #define l_Undef (lbool((uint8_t)2))
)

// MinisatCRef is `typedef RegionAllocator<uint32_t>::Ref CRef`.
type MinisatCRef = uint32

// MinisatCRefUndef is `const CRef CRef_Undef = RegionAllocator<uint32_t>::Ref_Undef`.
const MinisatCRefUndef MinisatCRef = MinisatRegionAllocatorRefUndef

// MinisatLitSeq is what the Clause constructor and ClauseAllocator::alloc template
// parameters (V, Lits) require of the literal source: size() and operator[]
// (instantiated with vec<Lit> and Clause).
type MinisatLitSeq interface {
	Size() int
	At(i int) *MinisatLit
}

// Bit layout of the Clause header word (gcc, little endian):
//
//	mark:2 (bits 0-1), learnt:1 (bit 2), has_extra:1 (bit 3), reloced:1 (bit 4), size:27 (bits 5-31)
const (
	minisatClauseMarkMask      = 0x3
	minisatClauseLearntBit     = 1 << 2
	minisatClauseHasExtraBit   = 1 << 3
	minisatClauseRelocedBit    = 1 << 4
	minisatClauseSizeShift     = 5
	minisatClauseHeaderWords   = 1 // sizeof(Clause) / sizeof(uint32_t)
	minisatClauseSizeFieldMask = (1 << 27) - 1
)

// MinisatClause is Minisat::Clause. The C++ object is a header followed by a
// flexible array `union { Lit lit; float act; uint32_t abs; CRef rel; } data[0]`
// of 32-bit words living inside a ClauseAllocator region. In Go a MinisatClause
// is a view onto those words: w[0] is the header, w[1+i] is data[i]. The
// allocator hands out views of memory[ref:], so the view's reference is
// len(memory)-len(w).
type MinisatClause struct {
	w []uint32
}

// newMinisatClause is the private template constructor Clause(const V& ps, bool use_extra, bool learnt),
// a placement new: w is the memory it constructs into.
// C: template<class V> Clause(const V& ps, bool use_extra, bool learnt)
func newMinisatClause(w []uint32, ps MinisatLitSeq, useExtra bool, learnt bool) *MinisatClause {
	c := &MinisatClause{w: w}
	var h uint32
	if learnt {
		h |= minisatClauseLearntBit
	}
	if useExtra {
		h |= minisatClauseHasExtraBit
	}
	h |= uint32(ps.Size()) << minisatClauseSizeShift
	c.w[0] = h
	for i := 0; i < ps.Size(); i++ {
		c.At(i).x = ps.At(i).x
	}
	if c.HasExtra() {
		if c.Learnt() {
			*c.Activity() = 0
		} else {
			c.CalcAbstraction()
		}
	}
	return c
}

// CalcAbstraction is Clause::calcAbstraction.
// C: void calcAbstraction()
func (c *MinisatClause) CalcAbstraction() {
	if !c.HasExtra() {
		panic("MinisatClause.CalcAbstraction: no extra field")
	}
	var abstraction uint32
	for i := 0; i < c.Size(); i++ {
		abstraction |= 1 << (uint(MinisatLitVar(*c.At(i))) & 31)
	}
	c.w[minisatClauseHeaderWords+c.Size()] = abstraction
}

// Size is Clause::size.
// C: int size() const
func (c *MinisatClause) Size() int {
	return int(c.w[0] >> minisatClauseSizeShift)
}

func (c *MinisatClause) setSize(n int) {
	c.w[0] = (c.w[0] & (1<<minisatClauseSizeShift - 1)) | uint32(n)<<minisatClauseSizeShift
}

// Shrink is Clause::shrink.
// C: void shrink(int i)
func (c *MinisatClause) Shrink(i int) {
	if i > c.Size() {
		panic("MinisatClause.Shrink: i > size")
	}
	if c.HasExtra() {
		c.w[minisatClauseHeaderWords+c.Size()-i] = c.w[minisatClauseHeaderWords+c.Size()]
	}
	c.setSize(c.Size() - i)
}

// Pop is Clause::pop.
// C: void pop()
func (c *MinisatClause) Pop() { c.Shrink(1) }

// Learnt is Clause::learnt.
// C: bool learnt() const
func (c *MinisatClause) Learnt() bool { return c.w[0]&minisatClauseLearntBit != 0 }

// HasExtra is Clause::has_extra.
// C: bool has_extra() const
func (c *MinisatClause) HasExtra() bool { return c.w[0]&minisatClauseHasExtraBit != 0 }

// Mark is the getter overload mark().
// C: uint32_t mark() const
func (c *MinisatClause) Mark() uint32 { return c.w[0] & minisatClauseMarkMask }

// SetMark is the setter overload mark(uint32_t m).
// C: void mark(uint32_t m)
func (c *MinisatClause) SetMark(m uint32) {
	c.w[0] = (c.w[0] &^ minisatClauseMarkMask) | (m & minisatClauseMarkMask)
}

// Last is Clause::last.
// C: const Lit& last() const
func (c *MinisatClause) Last() MinisatLit { return *c.At(c.Size() - 1) }

// Reloced is Clause::reloced.
// C: bool reloced() const
func (c *MinisatClause) Reloced() bool { return c.w[0]&minisatClauseRelocedBit != 0 }

// Relocation is Clause::relocation.
// C: CRef relocation() const
func (c *MinisatClause) Relocation() MinisatCRef { return c.w[minisatClauseHeaderWords] }

// Relocate is Clause::relocate.
// C: void relocate(CRef c)
func (c *MinisatClause) Relocate(cr MinisatCRef) {
	c.w[0] |= minisatClauseRelocedBit
	c.w[minisatClauseHeaderWords] = cr
}

// At is operator[] (const and non-const overloads); a Lit and a uint32 are both
// 4 bytes, so the union member data[i].lit is the word w[1+i] viewed as a Lit.
// C: Lit& operator[](int i)
func (c *MinisatClause) At(i int) *MinisatLit {
	return (*MinisatLit)(unsafe.Pointer(&c.w[minisatClauseHeaderWords+i]))
}

// Lits is `operator const Lit*() const`: the literals as a slice.
// C: operator const Lit*(void) const
func (c *MinisatClause) Lits() []MinisatLit {
	return unsafe.Slice((*MinisatLit)(unsafe.Pointer(&c.w[minisatClauseHeaderWords])), c.Size())
}

// Activity is Clause::activity (the float member of the extra union).
// C: float& activity()
func (c *MinisatClause) Activity() *float32 {
	if !c.HasExtra() {
		panic("MinisatClause.Activity: no extra field")
	}
	return (*float32)(unsafe.Pointer(&c.w[minisatClauseHeaderWords+c.Size()]))
}

// Abstraction is Clause::abstraction.
// C: uint32_t abstraction() const
func (c *MinisatClause) Abstraction() uint32 {
	if !c.HasExtra() {
		panic("MinisatClause.Abstraction: no extra field")
	}
	return c.w[minisatClauseHeaderWords+c.Size()]
}

// Subsumes is Clause::subsumes (defined inline after the classes): if *this
// subsumes 'other' (or 'other' with exactly one literal negated) it returns
// lit_Undef (resp. the literal that was negated), else lit_Error.
// C: inline Lit Clause::subsumes(const Clause& other) const
func (c *MinisatClause) Subsumes(other *MinisatClause) MinisatLit {
	if c.Learnt() || other.Learnt() {
		panic("MinisatClause.Subsumes: learnt clause")
	}
	if !c.HasExtra() || !other.HasExtra() {
		panic("MinisatClause.Subsumes: no extra field")
	}
	if other.Size() < c.Size() || (c.Abstraction()&^other.Abstraction()) != 0 {
		return MinisatLitError
	}
	ret := MinisatLitUndef
	cs := c.Lits()
	ds := other.Lits()
outer:
	for i := 0; i < c.Size(); i++ {
		// search for c[i] or ~c[i]
		for j := 0; j < other.Size(); j++ {
			if cs[i] == ds[j] {
				continue outer
			} else if ret == MinisatLitUndef && cs[i] == MinisatLitNot(ds[j]) {
				ret = cs[i]
				continue outer
			}
		}
		// did not find it
		return MinisatLitError
	}
	return ret
}

// Strengthen is Clause::strengthen (defined inline after the classes).
// C: inline void Clause::strengthen(Lit p)
func (c *MinisatClause) Strengthen(p MinisatLit) {
	minisatRemove[MinisatLit](c, p, minisatNe[MinisatLit])
	c.CalcAbstraction()
}

// MinisatClauseAllocator is Minisat::ClauseAllocator, a RegionAllocator<uint32_t>.
type MinisatClauseAllocator struct {
	MinisatRegionAllocator[uint32]
	ExtraClauseField bool // extra_clause_field
}

// clauseWord32Size is ClauseAllocator::clauseWord32Size (private static).
// C: static int clauseWord32Size(int size, bool has_extra)
func minisatClauseAllocatorClauseWord32Size(size int, hasExtra bool) int {
	extra := 0
	if hasExtra {
		extra = 1
	}
	return minisatClauseHeaderWords + size + extra
}

// NewMinisatClauseAllocatorCap is ClauseAllocator(uint32_t start_cap).
// C: ClauseAllocator(uint32_t start_cap)
func NewMinisatClauseAllocatorCap(startCap uint32) *MinisatClauseAllocator {
	a := &MinisatClauseAllocator{}
	a.capacity(startCap)
	return a
}

// NewMinisatClauseAllocator is ClauseAllocator(): the base allocator is
// default constructed (RegionAllocator's start_cap defaults to 1024*1024).
// C: ClauseAllocator()
func NewMinisatClauseAllocator() *MinisatClauseAllocator {
	a := &MinisatClauseAllocator{}
	a.capacity(1024 * 1024)
	return a
}

// MoveTo is ClauseAllocator::moveTo.
// C: void moveTo(ClauseAllocator& to)
func (a *MinisatClauseAllocator) MoveTo(to *MinisatClauseAllocator) {
	to.ExtraClauseField = a.ExtraClauseField
	a.MinisatRegionAllocator.MoveTo(&to.MinisatRegionAllocator)
}

// Alloc is the template ClauseAllocator::alloc (learnt defaults to false).
// C: template<class Lits> CRef alloc(const Lits& ps, bool learnt = false)
func (a *MinisatClauseAllocator) Alloc(ps MinisatLitSeq, learnt bool) MinisatCRef {
	useExtra := learnt || a.ExtraClauseField
	cid := a.MinisatRegionAllocator.Alloc(minisatClauseAllocatorClauseWord32Size(ps.Size(), useExtra))
	newMinisatClause(a.memory[cid:], ps, useExtra, learnt)
	return cid
}

// At is operator[] (const and non-const overloads).
// C: Clause& operator[](Ref r)
func (a *MinisatClauseAllocator) At(r MinisatCRef) *MinisatClause {
	if r >= a.sz {
		panic("MinisatClauseAllocator.At: reference out of range")
	}
	return &MinisatClause{w: a.memory[r:]}
}

// Lea is ClauseAllocator::lea.
// C: Clause* lea(Ref r)
func (a *MinisatClauseAllocator) Lea(r MinisatCRef) *MinisatClause { return a.At(r) }

// Ael is ClauseAllocator::ael: the reference of a clause view.
// C: Ref ael(const Clause* t)
func (a *MinisatClauseAllocator) Ael(t *MinisatClause) MinisatCRef {
	return MinisatCRef(len(a.memory) - len(t.w))
}

// Free is ClauseAllocator::_free.
// C: void _free(CRef cid)
func (a *MinisatClauseAllocator) Free(cid MinisatCRef) {
	c := a.At(cid)
	a.MinisatRegionAllocator.Free(minisatClauseAllocatorClauseWord32Size(c.Size(), c.HasExtra()))
}

// Reloc is ClauseAllocator::reloc.
// C: void reloc(CRef& cr, ClauseAllocator& to)
func (a *MinisatClauseAllocator) Reloc(cr *MinisatCRef, to *MinisatClauseAllocator) {
	c := a.At(*cr)
	if c.Reloced() {
		*cr = c.Relocation()
		return
	}
	*cr = to.Alloc(c, c.Learnt())
	c.Relocate(*cr)
	// Copy extra data-fields:
	// (This could be cleaned-up. Generalize Clause-constructor to be applicable here instead?)
	nc := to.At(*cr)
	nc.SetMark(c.Mark())
	if nc.Learnt() {
		*nc.Activity() = *c.Activity()
	} else if nc.HasExtra() {
		nc.CalcAbstraction()
	}
}

// MinisatOccIdx is what OccLists' index type may be (toInt is defined for Var and Lit).
type MinisatOccIdx interface {
	int | MinisatLit
}

// minisatOccToInt is the toInt() overload selected by OccLists' Idx.
func minisatOccToInt[Idx MinisatOccIdx](idx Idx) int {
	switch v := any(idx).(type) {
	case int:
		return MinisatToIntVar(v)
	case MinisatLit:
		return MinisatToIntLit(v)
	}
	panic("unreachable")
}

// MinisatDeleted is the requirement on OccLists' template argument Deleted: an
// object with bool operator()(const E&) over the element type E of the Vec.
type MinisatDeleted[E any] interface {
	Call(e E) bool
}

// MinisatOccLists is OccLists<Idx, Vec, Deleted> where Vec is vec<E>.
type MinisatOccLists[Idx MinisatOccIdx, E any, D MinisatDeleted[E]] struct {
	occs    MinisatVec[MinisatVec[E]]
	dirty   MinisatVec[int8]
	dirties MinisatVec[Idx]
	deleted D
}

// NewMinisatOccLists is OccLists(const Deleted& d).
// C: OccLists(const Deleted& d)
func NewMinisatOccLists[Idx MinisatOccIdx, E any, D MinisatDeleted[E]](d D) *MinisatOccLists[Idx, E, D] {
	return &MinisatOccLists[Idx, E, D]{deleted: d}
}

// Init is OccLists::init.
// C: void init(const Idx& idx)
func (o *MinisatOccLists[Idx, E, D]) Init(idx Idx) {
	o.occs.GrowTo(minisatOccToInt(idx) + 1)
	o.dirty.GrowToPad(minisatOccToInt(idx)+1, 0)
}

// At is OccLists::operator[].
// C: Vec& operator[](const Idx& idx)
func (o *MinisatOccLists[Idx, E, D]) At(idx Idx) *MinisatVec[E] {
	return o.occs.At(minisatOccToInt(idx))
}

// Lookup is OccLists::lookup.
// C: Vec& lookup(const Idx& idx)
func (o *MinisatOccLists[Idx, E, D]) Lookup(idx Idx) *MinisatVec[E] {
	if *o.dirty.At(minisatOccToInt(idx)) != 0 {
		o.Clean(idx)
	}
	return o.occs.At(minisatOccToInt(idx))
}

// CleanAll is OccLists::cleanAll.
// C: void cleanAll()
func (o *MinisatOccLists[Idx, E, D]) CleanAll() {
	for i := 0; i < o.dirties.Size(); i++ {
		// Dirties may contain duplicates so check here if a variable is already cleaned:
		if *o.dirty.At(minisatOccToInt(*o.dirties.At(i))) != 0 {
			o.Clean(*o.dirties.At(i))
		}
	}
	o.dirties.Clear(false)
}

// Clean is OccLists::clean.
// C: void clean(const Idx& idx)
func (o *MinisatOccLists[Idx, E, D]) Clean(idx Idx) {
	vec := o.occs.At(minisatOccToInt(idx))
	var i, j int
	for i = 0; i < vec.Size(); i++ {
		if !o.deleted.Call(*vec.At(i)) {
			*vec.At(j) = *vec.At(i)
			j++
		}
	}
	vec.Shrink(i - j)
	*o.dirty.At(minisatOccToInt(idx)) = 0
}

// Smudge is OccLists::smudge.
// C: void smudge(const Idx& idx)
func (o *MinisatOccLists[Idx, E, D]) Smudge(idx Idx) {
	if *o.dirty.At(minisatOccToInt(idx)) == 0 {
		*o.dirty.At(minisatOccToInt(idx)) = 1
		o.dirties.PushElem(idx)
	}
}

// Clear is OccLists::clear (free defaults to true).
// C: void clear(bool free = true)
func (o *MinisatOccLists[Idx, E, D]) Clear(free bool) {
	o.occs.Clear(free)
	o.dirty.Clear(free)
	o.dirties.Clear(free)
}

// MinisatCRefHash is CMap<T>::CRefHash.
type MinisatCRefHash struct{}

// Call is CRefHash::operator().
// C: uint32_t operator()(CRef cr) const
func (h MinisatCRefHash) Call(cr MinisatCRef) uint32 { return uint32(cr) }

// MinisatCMap is Minisat::CMap<T>: a Map<CRef, T, CRefHash>.
type MinisatCMap[T any] struct {
	map_ MinisatMap[MinisatCRef, T, MinisatCRefHash, MinisatEqual[MinisatCRef]]
}

// Clear is CMap::clear.
// C: void clear()
func (m *MinisatCMap[T]) Clear() { m.map_.Clear() }

// Size is CMap::size.
// C: int size() const
func (m *MinisatCMap[T]) Size() int { return m.map_.Elems() }

// Insert is CMap::insert.
// C: void insert(CRef cr, const T& t)
func (m *MinisatCMap[T]) Insert(cr MinisatCRef, t T) { m.map_.Insert(cr, t) }

// GrowTo is CMap::growTo (for compatibility; same as insert).
// C: void growTo(CRef cr, const T& t)
func (m *MinisatCMap[T]) GrowTo(cr MinisatCRef, t T) { m.map_.Insert(cr, t) }

// Remove is CMap::remove.
// C: void remove(CRef cr)
func (m *MinisatCMap[T]) Remove(cr MinisatCRef) { m.map_.Remove(cr) }

// Has is CMap::has.
// C: bool has(CRef cr, T& t)
func (m *MinisatCMap[T]) Has(cr MinisatCRef, t *T) bool { return m.map_.Peek(cr, t) }

// At is operator[] (const and non-const overloads).
// C: T& operator[](CRef cr)
func (m *MinisatCMap[T]) At(cr MinisatCRef) *T { return m.map_.At(cr) }

// BucketCount is CMap::bucket_count.
// C: int bucket_count() const
func (m *MinisatCMap[T]) BucketCount() int { return m.map_.BucketCount() }

// Bucket is CMap::bucket.
// C: const vec<typename HashTable::Pair>& bucket(int i) const
func (m *MinisatCMap[T]) Bucket(i int) *MinisatVec[MinisatMapPair[MinisatCRef, T]] {
	return m.map_.Bucket(i)
}

// MoveTo is CMap::moveTo.
// C: void moveTo(CMap& other)
func (m *MinisatCMap[T]) MoveTo(other *MinisatCMap[T]) { m.map_.MoveTo(&other.map_) }

// Debug is CMap::debug.
// C: void debug()
func (m *MinisatCMap[T]) Debug() {
	println(" --- size =", m.Size(), ", bucket_count =", m.map_.BucketCount())
}
