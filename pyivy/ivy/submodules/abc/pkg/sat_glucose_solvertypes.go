package abc

import "unsafe"

// Port of src/sat/glucose/SolverTypes.h (namespace Gluco).
// See sat_bsat2_vec.go for the C++ -> Go conventions.

// GlucoVar is `typedef int Var`.
type GlucoVar = int

// GlucoVarUndef is the macro `#define var_Undef (-1)`.
const GlucoVarUndef = -1

// GlucoLit is Gluco::Lit: a literal, x = 2*var + sign (a 32-bit int in C).
type GlucoLit struct {
	x int32
}

// Eq is Lit::operator==.
// C: bool operator==(Lit p) const
func (l GlucoLit) Eq(p GlucoLit) bool { return l.x == p.x }

// Ne is Lit::operator!=.
// C: bool operator!=(Lit p) const
func (l GlucoLit) Ne(p GlucoLit) bool { return l.x != p.x }

// Lt is Lit::operator< (makes p, ~p adjacent in the ordering).
// C: bool operator<(Lit p) const
func (l GlucoLit) Lt(p GlucoLit) bool { return l.x < p.x }

// GlucoMkLit is mkLit (inline; sign defaults to false).
// C: inline Lit mkLit(Var var, bool sign = false)
func GlucoMkLit(v GlucoVar, sign bool) GlucoLit {
	var p GlucoLit
	p.x = int32(v + v)
	if sign {
		p.x++
	}
	return p
}

// GlucoLitNot is `operator ~` on Lit (inline).
// C: inline Lit operator~(Lit p)
func GlucoLitNot(p GlucoLit) GlucoLit { return GlucoLit{x: p.x ^ 1} }

// GlucoLitXor is `operator ^` on (Lit, bool) (inline).
// C: inline Lit operator^(Lit p, bool b)
func GlucoLitXor(p GlucoLit, b bool) GlucoLit {
	if b {
		return GlucoLit{x: p.x ^ 1}
	}
	return GlucoLit{x: p.x}
}

// GlucoSign is Gluco::sign (inline).
// C: inline bool sign(Lit p)
func GlucoSign(p GlucoLit) bool { return p.x&1 != 0 }

// GlucoLitVar is Gluco::var(Lit) (inline); renamed because GlucoVar is the type.
// C: inline int var(Lit p)
func GlucoLitVar(p GlucoLit) int { return int(p.x >> 1) }

// GlucoToIntVar is the overload toInt(Var) (inline).
// C: inline int toInt(Var v)
func GlucoToIntVar(v GlucoVar) int { return v }

// GlucoToIntLit is the overload toInt(Lit) (inline).
// C: inline int toInt(Lit p)
func GlucoToIntLit(p GlucoLit) int { return int(p.x) }

// GlucoToLit is Gluco::toLit (inline).
// C: inline Lit toLit(int i)
func GlucoToLit(i int) GlucoLit { return GlucoLit{x: int32(i)} }

// GlucoLitUndef is `const Lit lit_Undef = { -2 }`.
var GlucoLitUndef = GlucoLit{x: -2}

// GlucoLitError is `const Lit lit_Error = { -1 }`.
var GlucoLitError = GlucoLit{x: -1}

// GlucoLbool is Gluco::lbool: a three-valued boolean (0 true, 1 false, 2 undef).
type GlucoLbool struct {
	value uint8
}

// NewGlucoLboolV is the explicit constructor lbool(uint8_t v).
// C: explicit lbool(uint8_t v)
func NewGlucoLboolV(v uint8) GlucoLbool { return glucoLboolNorm(v) }

// NewGlucoLbool is lbool(): value 0.
// C: lbool()
func NewGlucoLbool() GlucoLbool { return GlucoLbool{} }

// NewGlucoLboolB is the explicit constructor lbool(bool x): value !x.
// C: explicit lbool(bool x)
func NewGlucoLboolB(x bool) GlucoLbool {
	if x {
		return GlucoLbool{value: 0}
	}
	return GlucoLbool{value: 1}
}

// Eq is lbool::operator==.
// C: bool operator==(lbool b) const
func (l GlucoLbool) Eq(b GlucoLbool) bool {
	return ((b.value&2)&(l.value&2)) != 0 || ((b.value&2) == 0 && l.value == b.value)
}

// Ne is lbool::operator!=.
// C: bool operator!=(lbool b) const
func (l GlucoLbool) Ne(b GlucoLbool) bool { return !l.Eq(b) }

// glucoLboolNorm canonicalises an lbool value: C++ lbool::operator== treats every value
// with bit 1 set as l_Undef (e.g. l_Undef ^ true == 3 still equals l_Undef), so the Go port
// stores all of them as 2 and the struct comparison == is then equivalent to operator==.
func glucoLboolNorm(v uint8) GlucoLbool {
	if v&2 != 0 {
		v = 2
	}
	return GlucoLbool{value: v}
}

// Xor is lbool::operator^.
// C: lbool operator^(bool b) const
func (l GlucoLbool) Xor(b bool) GlucoLbool {
	var u uint8
	if b {
		u = 1
	}
	return glucoLboolNorm(l.value ^ u)
}

// And is lbool::operator&&.
// C: lbool operator&&(lbool b) const
func (l GlucoLbool) And(b GlucoLbool) GlucoLbool {
	sel := (l.value << 1) | (b.value << 3)
	v := uint8((uint32(0xF7F755F4) >> sel) & 3)
	return glucoLboolNorm(v)
}

// Or is lbool::operator||.
// C: lbool operator||(lbool b) const
func (l GlucoLbool) Or(b GlucoLbool) GlucoLbool {
	sel := (l.value << 1) | (b.value << 3)
	v := uint8((uint32(0xFCFCF400) >> sel) & 3)
	return glucoLboolNorm(v)
}

// GlucoToIntLbool is the friend toInt(lbool) (inline).
// C: inline int toInt(lbool l)
func GlucoToIntLbool(l GlucoLbool) int { return int(l.value) }

// GlucoToLbool is the friend toLbool(int) (inline).
// C: inline lbool toLbool(int v)
func GlucoToLbool(v int) GlucoLbool { return glucoLboolNorm(uint8(v)) }

// The macros l_True, l_False and l_Undef.
var (
	GlucoLTrue  = GlucoLbool{value: 0} // #define l_True  (lbool((uint8_t)0))
	GlucoLFalse = GlucoLbool{value: 1} // #define l_False (lbool((uint8_t)1))
	GlucoLUndef = GlucoLbool{value: 2} // #define l_Undef (lbool((uint8_t)2))
)

// GlucoCRef is `typedef RegionAllocator<uint32_t>::Ref CRef`.
type GlucoCRef = uint32

// GlucoCRefUndef is `const CRef CRef_Undef = RegionAllocator<uint32_t>::Ref_Undef`.
const GlucoCRefUndef GlucoCRef = GlucoRegionAllocatorRefUndef

// GlucoLitSeq is what the Clause constructor and ClauseAllocator::alloc template
// parameters (V, Lits) require of the literal source: size() and operator[]
// (instantiated with vec<Lit> and Clause).
type GlucoLitSeq interface {
	Size() int
	At(i int) *GlucoLit
}

// Layout of the Clause header (gcc, little endian). The header is three 32-bit words:
//
//	w[0]: mark:2 (bits 0-1), learnt:1 (bit 2), has_extra:1 (bit 3), reloced:1 (bit 4), lbd:26 (bits 5-30), canbedel:1 (bit 31)
//	w[1]: size:32
//	w[2]: szWithoutSelectors:32
const (
	glucoClauseMarkMask    = 0x3
	glucoClauseLearntBit   = 1 << 2
	glucoClauseHasExtraBit = 1 << 3
	glucoClauseRelocedBit  = 1 << 4
	glucoClauseLbdShift    = 5
	glucoClauseLbdMask     = (1<<26 - 1) << glucoClauseLbdShift
	glucoClauseCanBeDelBit = 1 << 31
	glucoClauseHeaderWords = 3 // sizeof(Clause) / sizeof(uint32_t)
)

// GlucoClause is Gluco::Clause. The C++ object is a header followed by a
// flexible array `union { Lit lit; float act; uint32_t abs; CRef rel; } data[0]`
// of 32-bit words living inside a ClauseAllocator region. In Go a GlucoClause
// is a view onto those words: w[0] is the header, w[1+i] is data[i]. The
// allocator hands out views of memory[ref:], so the view's reference is
// len(memory)-len(w).
type GlucoClause struct {
	w []uint32
}

// newGlucoClause is the private template constructor Clause(const V& ps, bool use_extra, bool learnt),
// a placement new: w is the memory it constructs into.
// C: template<class V> Clause(const V& ps, bool use_extra, bool learnt)
func newGlucoClause(w []uint32, ps GlucoLitSeq, useExtra bool, learnt bool) *GlucoClause {
	c := &GlucoClause{w: w}
	var h uint32
	if learnt {
		h |= glucoClauseLearntBit
	}
	if useExtra {
		h |= glucoClauseHasExtraBit
	}
	h |= glucoClauseCanBeDelBit // header.lbd = 0; header.canbedel = 1
	c.w[0] = h
	c.w[1] = uint32(ps.Size())
	c.w[2] = 0
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
func (c *GlucoClause) CalcAbstraction() {
	if !c.HasExtra() {
		panic("GlucoClause.CalcAbstraction: no extra field")
	}
	var abstraction uint32
	for i := 0; i < c.Size(); i++ {
		abstraction |= 1 << (uint(GlucoLitVar(*c.At(i))) & 31)
	}
	c.w[glucoClauseHeaderWords+c.Size()] = abstraction
}

// Size is Clause::size.
// C: int size() const
func (c *GlucoClause) Size() int {
	return int(c.w[1])
}

func (c *GlucoClause) setSize(n int) {
	c.w[1] = uint32(n)
}

// Shrink is Clause::shrink.
// C: void shrink(int i)
func (c *GlucoClause) Shrink(i int) {
	if i > c.Size() {
		panic("GlucoClause.Shrink: i > size")
	}
	if c.HasExtra() {
		c.w[glucoClauseHeaderWords+c.Size()-i] = c.w[glucoClauseHeaderWords+c.Size()]
	}
	c.setSize(c.Size() - i)
}

// Pop is Clause::pop.
// C: void pop()
func (c *GlucoClause) Pop() { c.Shrink(1) }

// Learnt is Clause::learnt.
// C: bool learnt() const
func (c *GlucoClause) Learnt() bool { return c.w[0]&glucoClauseLearntBit != 0 }

// HasExtra is Clause::has_extra.
// C: bool has_extra() const
func (c *GlucoClause) HasExtra() bool { return c.w[0]&glucoClauseHasExtraBit != 0 }

// Mark is the getter overload mark().
// C: uint32_t mark() const
func (c *GlucoClause) Mark() uint32 { return c.w[0] & glucoClauseMarkMask }

// SetMark is the setter overload mark(uint32_t m).
// C: void mark(uint32_t m)
func (c *GlucoClause) SetMark(m uint32) {
	c.w[0] = (c.w[0] &^ glucoClauseMarkMask) | (m & glucoClauseMarkMask)
}

// Last is Clause::last.
// C: const Lit& last() const
func (c *GlucoClause) Last() GlucoLit { return *c.At(c.Size() - 1) }

// Reloced is Clause::reloced.
// C: bool reloced() const
func (c *GlucoClause) Reloced() bool { return c.w[0]&glucoClauseRelocedBit != 0 }

// Relocation is Clause::relocation.
// C: CRef relocation() const
func (c *GlucoClause) Relocation() GlucoCRef { return c.w[glucoClauseHeaderWords] }

// Relocate is Clause::relocate.
// C: void relocate(CRef c)
func (c *GlucoClause) Relocate(cr GlucoCRef) {
	c.w[0] |= glucoClauseRelocedBit
	c.w[glucoClauseHeaderWords] = cr
}

// At is operator[] (const and non-const overloads); a Lit and a uint32 are both
// 4 bytes, so the union member data[i].lit is the word w[1+i] viewed as a Lit.
// C: Lit& operator[](int i)
func (c *GlucoClause) At(i int) *GlucoLit {
	return (*GlucoLit)(unsafe.Pointer(&c.w[glucoClauseHeaderWords+i]))
}

// Lits is `operator const Lit*() const`: the literals as a slice.
// C: operator const Lit*(void) const
func (c *GlucoClause) Lits() []GlucoLit {
	return unsafe.Slice((*GlucoLit)(unsafe.Pointer(&c.w[glucoClauseHeaderWords])), c.Size())
}

// Activity is Clause::activity (the float member of the extra union).
// C: float& activity()
func (c *GlucoClause) Activity() *float32 {
	if !c.HasExtra() {
		panic("GlucoClause.Activity: no extra field")
	}
	return (*float32)(unsafe.Pointer(&c.w[glucoClauseHeaderWords+c.Size()]))
}

// Abstraction is Clause::abstraction.
// C: uint32_t abstraction() const
func (c *GlucoClause) Abstraction() uint32 {
	if !c.HasExtra() {
		panic("GlucoClause.Abstraction: no extra field")
	}
	return c.w[glucoClauseHeaderWords+c.Size()]
}

// SetLBD is Clause::setLBD.
// C: void setLBD(int i)
func (c *GlucoClause) SetLBD(i int) {
	c.w[0] = (c.w[0] &^ glucoClauseLbdMask) | (uint32(i)<<glucoClauseLbdShift)&glucoClauseLbdMask
}

// Lbd is Clause::lbd.
// C: unsigned int lbd() const
func (c *GlucoClause) Lbd() uint32 { return (c.w[0] & glucoClauseLbdMask) >> glucoClauseLbdShift }

// SetCanBeDel is Clause::setCanBeDel.
// C: void setCanBeDel(bool b)
func (c *GlucoClause) SetCanBeDel(b bool) {
	if b {
		c.w[0] |= glucoClauseCanBeDelBit
	} else {
		c.w[0] &^= glucoClauseCanBeDelBit
	}
}

// CanBeDel is Clause::canBeDel.
// C: bool canBeDel()
func (c *GlucoClause) CanBeDel() bool { return c.w[0]&glucoClauseCanBeDelBit != 0 }

// SetSizeWithoutSelectors is Clause::setSizeWithoutSelectors.
// C: void setSizeWithoutSelectors(unsigned int n)
func (c *GlucoClause) SetSizeWithoutSelectors(n uint32) { c.w[2] = n }

// SizeWithoutSelectors is Clause::sizeWithoutSelectors.
// C: unsigned int sizeWithoutSelectors() const
func (c *GlucoClause) SizeWithoutSelectors() uint32 { return c.w[2] }

// Subsumes is Clause::subsumes (defined inline after the classes): if *this
// subsumes 'other' (or 'other' with exactly one literal negated) it returns
// lit_Undef (resp. the literal that was negated), else lit_Error.
// C: inline Lit Clause::subsumes(const Clause& other) const
func (c *GlucoClause) Subsumes(other *GlucoClause) GlucoLit {
	if c.Learnt() || other.Learnt() {
		panic("GlucoClause.Subsumes: learnt clause")
	}
	if !c.HasExtra() || !other.HasExtra() {
		panic("GlucoClause.Subsumes: no extra field")
	}
	if other.Size() < c.Size() || (c.Abstraction()&^other.Abstraction()) != 0 {
		return GlucoLitError
	}
	ret := GlucoLitUndef
	cs := c.Lits()
	ds := other.Lits()
outer:
	for i := 0; i < c.Size(); i++ {
		// search for c[i] or ~c[i]
		for j := 0; j < other.Size(); j++ {
			if cs[i] == ds[j] {
				continue outer
			} else if ret == GlucoLitUndef && cs[i] == GlucoLitNot(ds[j]) {
				ret = cs[i]
				continue outer
			}
		}
		// did not find it
		return GlucoLitError
	}
	return ret
}

// Strengthen is Clause::strengthen (defined inline after the classes).
// C: inline void Clause::strengthen(Lit p)
func (c *GlucoClause) Strengthen(p GlucoLit) {
	glucoRemove[GlucoLit](c, p, glucoNe[GlucoLit])
	c.CalcAbstraction()
}

// GlucoClauseAllocator is Gluco::ClauseAllocator, a RegionAllocator<uint32_t>.
type GlucoClauseAllocator struct {
	GlucoRegionAllocator[uint32]
	ExtraClauseField bool // extra_clause_field
}

// clauseWord32Size is ClauseAllocator::clauseWord32Size (private static).
// C: static int clauseWord32Size(int size, bool has_extra)
func glucoClauseAllocatorClauseWord32Size(size int, hasExtra bool) int {
	extra := 0
	if hasExtra {
		extra = 1
	}
	return glucoClauseHeaderWords + size + extra
}

// NewGlucoClauseAllocatorCap is ClauseAllocator(uint32_t start_cap).
// C: ClauseAllocator(uint32_t start_cap)
func NewGlucoClauseAllocatorCap(startCap uint32) *GlucoClauseAllocator {
	a := &GlucoClauseAllocator{}
	a.capacity(startCap)
	return a
}

// NewGlucoClauseAllocator is ClauseAllocator(): the base allocator is
// default constructed (RegionAllocator's start_cap defaults to 1024*1024).
// C: ClauseAllocator()
func NewGlucoClauseAllocator() *GlucoClauseAllocator {
	a := &GlucoClauseAllocator{}
	a.capacity(1024 * 1024)
	return a
}

// MoveTo is ClauseAllocator::moveTo.
// C: void moveTo(ClauseAllocator& to)
func (a *GlucoClauseAllocator) MoveTo(to *GlucoClauseAllocator) {
	to.ExtraClauseField = a.ExtraClauseField
	a.GlucoRegionAllocator.MoveTo(&to.GlucoRegionAllocator)
}

// Alloc is the template ClauseAllocator::alloc (learnt defaults to false).
// C: template<class Lits> CRef alloc(const Lits& ps, bool learnt = false)
func (a *GlucoClauseAllocator) Alloc(ps GlucoLitSeq, learnt bool) GlucoCRef {
	useExtra := learnt || a.ExtraClauseField
	cid := a.GlucoRegionAllocator.Alloc(glucoClauseAllocatorClauseWord32Size(ps.Size(), useExtra))
	newGlucoClause(a.memory[cid:], ps, useExtra, learnt)
	return cid
}

// At is operator[] (const and non-const overloads).
// C: Clause& operator[](Ref r)
func (a *GlucoClauseAllocator) At(r GlucoCRef) *GlucoClause {
	if r >= a.sz {
		panic("GlucoClauseAllocator.At: reference out of range")
	}
	return &GlucoClause{w: a.memory[r:]}
}

// Lea is ClauseAllocator::lea.
// C: Clause* lea(Ref r)
func (a *GlucoClauseAllocator) Lea(r GlucoCRef) *GlucoClause { return a.At(r) }

// Ael is ClauseAllocator::ael: the reference of a clause view.
// C: Ref ael(const Clause* t)
func (a *GlucoClauseAllocator) Ael(t *GlucoClause) GlucoCRef {
	return GlucoCRef(len(a.memory) - len(t.w))
}

// Free_ is ClauseAllocator::free_ (renamed from _free in Minisat).
// C: void free_(CRef cid)
func (a *GlucoClauseAllocator) Free_(cid GlucoCRef) {
	c := a.At(cid)
	a.GlucoRegionAllocator.Free_(glucoClauseAllocatorClauseWord32Size(c.Size(), c.HasExtra()))
}

// Reloc is ClauseAllocator::reloc.
// C: void reloc(CRef& cr, ClauseAllocator& to)
func (a *GlucoClauseAllocator) Reloc(cr *GlucoCRef, to *GlucoClauseAllocator) {
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
		nc.SetLBD(int(c.Lbd()))
		nc.SetSizeWithoutSelectors(c.SizeWithoutSelectors())
		nc.SetCanBeDel(c.CanBeDel())
	} else if nc.HasExtra() {
		nc.CalcAbstraction()
	}
}

// GlucoOccIdx is what OccLists' index type may be (toInt is defined for Var and Lit).
type GlucoOccIdx interface {
	int | GlucoLit
}

// glucoOccToInt is the toInt() overload selected by OccLists' Idx.
func glucoOccToInt[Idx GlucoOccIdx](idx Idx) int {
	switch v := any(idx).(type) {
	case int:
		return GlucoToIntVar(v)
	case GlucoLit:
		return GlucoToIntLit(v)
	}
	panic("unreachable")
}

// GlucoDeleted is the requirement on OccLists' template argument Deleted: an
// object with bool operator()(const E&) over the element type E of the Vec.
type GlucoDeleted[E any] interface {
	Call(e E) bool
}

// GlucoOccLists is OccLists<Idx, Vec, Deleted> where Vec is vec<E>.
type GlucoOccLists[Idx GlucoOccIdx, E any, D GlucoDeleted[E]] struct {
	occs    GlucoVec[GlucoVec[E]]
	dirty   GlucoVec[int8]
	dirties GlucoVec[Idx]
	deleted D
}

// NewGlucoOccLists is OccLists(const Deleted& d).
// C: OccLists(const Deleted& d)
func NewGlucoOccLists[Idx GlucoOccIdx, E any, D GlucoDeleted[E]](d D) *GlucoOccLists[Idx, E, D] {
	return &GlucoOccLists[Idx, E, D]{deleted: d}
}

// Init is OccLists::init.
// C: void init(const Idx& idx)
func (o *GlucoOccLists[Idx, E, D]) Init(idx Idx) {
	o.occs.GrowTo(glucoOccToInt(idx) + 1)
	o.dirty.GrowToPad(glucoOccToInt(idx)+1, 0)
}

// At is OccLists::operator[].
// C: Vec& operator[](const Idx& idx)
func (o *GlucoOccLists[Idx, E, D]) At(idx Idx) *GlucoVec[E] {
	return o.occs.At(glucoOccToInt(idx))
}

// Lookup is OccLists::lookup.
// C: Vec& lookup(const Idx& idx)
func (o *GlucoOccLists[Idx, E, D]) Lookup(idx Idx) *GlucoVec[E] {
	if *o.dirty.At(glucoOccToInt(idx)) != 0 {
		o.Clean(idx)
	}
	return o.occs.At(glucoOccToInt(idx))
}

// CleanAll is OccLists::cleanAll.
// C: void cleanAll()
func (o *GlucoOccLists[Idx, E, D]) CleanAll() {
	for i := 0; i < o.dirties.Size(); i++ {
		// Dirties may contain duplicates so check here if a variable is already cleaned:
		if *o.dirty.At(glucoOccToInt(*o.dirties.At(i))) != 0 {
			o.Clean(*o.dirties.At(i))
		}
	}
	o.dirties.Clear(false)
}

// Clean is OccLists::clean.
// C: void clean(const Idx& idx)
func (o *GlucoOccLists[Idx, E, D]) Clean(idx Idx) {
	vec := o.occs.At(glucoOccToInt(idx))
	var i, j int
	for i = 0; i < vec.Size(); i++ {
		if !o.deleted.Call(*vec.At(i)) {
			*vec.At(j) = *vec.At(i)
			j++
		}
	}
	vec.Shrink(i - j)
	*o.dirty.At(glucoOccToInt(idx)) = 0
}

// Smudge is OccLists::smudge.
// C: void smudge(const Idx& idx)
func (o *GlucoOccLists[Idx, E, D]) Smudge(idx Idx) {
	if *o.dirty.At(glucoOccToInt(idx)) == 0 {
		*o.dirty.At(glucoOccToInt(idx)) = 1
		o.dirties.PushElem(idx)
	}
}

// Clear is OccLists::clear (free defaults to true).
// C: void clear(bool free = true)
func (o *GlucoOccLists[Idx, E, D]) Clear(free bool) {
	if free {
		o.occs.Clear(free)
		o.dirty.Clear(free)
		o.dirties.Clear(free)
	} else {
		o.occs.Shrink_(o.occs.Size())
		o.dirty.Shrink_(o.dirty.Size())
		o.dirties.Shrink_(o.dirties.Size())
	}
}

// GlucoCRefHash is CMap<T>::CRefHash.
type GlucoCRefHash struct{}

// Call is CRefHash::operator().
// C: uint32_t operator()(CRef cr) const
func (h GlucoCRefHash) Call(cr GlucoCRef) uint32 { return uint32(cr) }

// GlucoCMap is Gluco::CMap<T>: a Map<CRef, T, CRefHash>.
type GlucoCMap[T any] struct {
	map_ GlucoMap[GlucoCRef, T, GlucoCRefHash, GlucoEqual[GlucoCRef]]
}

// Clear is CMap::clear.
// C: void clear()
func (m *GlucoCMap[T]) Clear() { m.map_.Clear() }

// Size is CMap::size.
// C: int size() const
func (m *GlucoCMap[T]) Size() int { return m.map_.Elems() }

// Insert is CMap::insert.
// C: void insert(CRef cr, const T& t)
func (m *GlucoCMap[T]) Insert(cr GlucoCRef, t T) { m.map_.Insert(cr, t) }

// GrowTo is CMap::growTo (for compatibility; same as insert).
// C: void growTo(CRef cr, const T& t)
func (m *GlucoCMap[T]) GrowTo(cr GlucoCRef, t T) { m.map_.Insert(cr, t) }

// Remove is CMap::remove.
// C: void remove(CRef cr)
func (m *GlucoCMap[T]) Remove(cr GlucoCRef) { m.map_.Remove(cr) }

// Has is CMap::has.
// C: bool has(CRef cr, T& t)
func (m *GlucoCMap[T]) Has(cr GlucoCRef, t *T) bool { return m.map_.Peek(cr, t) }

// At is operator[] (const and non-const overloads).
// C: T& operator[](CRef cr)
func (m *GlucoCMap[T]) At(cr GlucoCRef) *T { return m.map_.At(cr) }

// BucketCount is CMap::bucket_count.
// C: int bucket_count() const
func (m *GlucoCMap[T]) BucketCount() int { return m.map_.BucketCount() }

// Bucket is CMap::bucket.
// C: const vec<typename HashTable::Pair>& bucket(int i) const
func (m *GlucoCMap[T]) Bucket(i int) *GlucoVec[GlucoMapPair[GlucoCRef, T]] {
	return m.map_.Bucket(i)
}

// MoveTo is CMap::moveTo.
// C: void moveTo(CMap& other)
func (m *GlucoCMap[T]) MoveTo(other *GlucoCMap[T]) { m.map_.MoveTo(&other.map_) }

// Debug is CMap::debug.
// C: void debug()
func (m *GlucoCMap[T]) Debug() {
	println(" --- size =", m.Size(), ", bucket_count =", m.map_.BucketCount())
}
