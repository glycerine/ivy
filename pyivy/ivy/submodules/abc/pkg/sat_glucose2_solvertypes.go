package abc

import "unsafe"

// Port of src/sat/glucose2/SolverTypes.h (namespace Gluco2).
// See sat_bsat2_vec.go for the C++ -> Go conventions.

// Gluco2Var is `typedef int Var`.
type Gluco2Var = int

// Gluco2VarUndef is the macro `#define var_Undef (-1)`.
const Gluco2VarUndef = -1

// Gluco2Lit is Gluco2::Lit: a literal, x = 2*var + sign (a 32-bit int in C).
type Gluco2Lit struct {
	x int32
}

// Eq is Lit::operator==.
// C: bool operator==(Lit p) const
func (l Gluco2Lit) Eq(p Gluco2Lit) bool { return l.x == p.x }

// Ne is Lit::operator!=.
// C: bool operator!=(Lit p) const
func (l Gluco2Lit) Ne(p Gluco2Lit) bool { return l.x != p.x }

// Lt is Lit::operator< (makes p, ~p adjacent in the ordering).
// C: bool operator<(Lit p) const
func (l Gluco2Lit) Lt(p Gluco2Lit) bool { return l.x < p.x }

// Gluco2MkLit is mkLit (inline; sign defaults to false).
// C: inline Lit mkLit(Var var, bool sign = false)
func Gluco2MkLit(v Gluco2Var, sign bool) Gluco2Lit {
	var p Gluco2Lit
	p.x = int32(v + v)
	if sign {
		p.x++
	}
	return p
}

// Gluco2LitNot is `operator ~` on Lit (inline).
// C: inline Lit operator~(Lit p)
func Gluco2LitNot(p Gluco2Lit) Gluco2Lit { return Gluco2Lit{x: p.x ^ 1} }

// Gluco2LitXor is `operator ^` on (Lit, bool) (inline).
// C: inline Lit operator^(Lit p, bool b)
func Gluco2LitXor(p Gluco2Lit, b bool) Gluco2Lit {
	if b {
		return Gluco2Lit{x: p.x ^ 1}
	}
	return Gluco2Lit{x: p.x}
}

// Gluco2Sign is Gluco2::sign (inline).
// C: inline bool sign(Lit p)
func Gluco2Sign(p Gluco2Lit) bool { return p.x&1 != 0 }

// Gluco2LitVar is Gluco2::var(Lit) (inline); renamed because Gluco2Var is the type.
// C: inline int var(Lit p)
func Gluco2LitVar(p Gluco2Lit) int { return int(p.x >> 1) }

// Gluco2ToIntVar is the overload toInt(Var) (inline).
// C: inline int toInt(Var v)
func Gluco2ToIntVar(v Gluco2Var) int { return v }

// Gluco2ToIntLit is the overload toInt(Lit) (inline).
// C: inline int toInt(Lit p)
func Gluco2ToIntLit(p Gluco2Lit) int { return int(p.x) }

// Gluco2ToLit is Gluco2::toLit (inline).
// C: inline Lit toLit(int i)
func Gluco2ToLit(i int) Gluco2Lit { return Gluco2Lit{x: int32(i)} }

// Gluco2LitUndef is `const Lit lit_Undef = { -2 }`.
var Gluco2LitUndef = Gluco2Lit{x: -2}

// Gluco2LitError is `const Lit lit_Error = { -1 }`.
var Gluco2LitError = Gluco2Lit{x: -1}

// Gluco2Lbool is Gluco2::lbool: a three-valued boolean (0 true, 1 false, 2 undef).
type Gluco2Lbool struct {
	value uint8
}

// NewGluco2LboolV is the explicit constructor lbool(uint8_t v).
// C: explicit lbool(uint8_t v)
func NewGluco2LboolV(v uint8) Gluco2Lbool { return gluco2LboolNorm(v) }

// NewGluco2Lbool is lbool(): value 0.
// C: lbool()
func NewGluco2Lbool() Gluco2Lbool { return Gluco2Lbool{} }

// NewGluco2LboolB is the explicit constructor lbool(bool x): value !x.
// C: explicit lbool(bool x)
func NewGluco2LboolB(x bool) Gluco2Lbool {
	if x {
		return Gluco2Lbool{value: 0}
	}
	return Gluco2Lbool{value: 1}
}

// Eq is lbool::operator==.
// C: bool operator==(lbool b) const
func (l Gluco2Lbool) Eq(b Gluco2Lbool) bool {
	return ((b.value&2)&(l.value&2)) != 0 || ((b.value&2) == 0 && l.value == b.value)
}

// Ne is lbool::operator!=.
// C: bool operator!=(lbool b) const
func (l Gluco2Lbool) Ne(b Gluco2Lbool) bool { return !l.Eq(b) }

// gluco2LboolNorm canonicalises an lbool value: C++ lbool::operator== treats every value
// with bit 1 set as l_Undef (e.g. l_Undef ^ true == 3 still equals l_Undef), so the Go port
// stores all of them as 2 and the struct comparison == is then equivalent to operator==.
func gluco2LboolNorm(v uint8) Gluco2Lbool {
	if v&2 != 0 {
		v = 2
	}
	return Gluco2Lbool{value: v}
}

// Xor is lbool::operator^.
// C: lbool operator^(bool b) const
func (l Gluco2Lbool) Xor(b bool) Gluco2Lbool {
	var u uint8
	if b {
		u = 1
	}
	return gluco2LboolNorm(l.value ^ u)
}

// And is lbool::operator&&.
// C: lbool operator&&(lbool b) const
func (l Gluco2Lbool) And(b Gluco2Lbool) Gluco2Lbool {
	sel := (l.value << 1) | (b.value << 3)
	v := uint8((uint32(0xF7F755F4) >> sel) & 3)
	return gluco2LboolNorm(v)
}

// Or is lbool::operator||.
// C: lbool operator||(lbool b) const
func (l Gluco2Lbool) Or(b Gluco2Lbool) Gluco2Lbool {
	sel := (l.value << 1) | (b.value << 3)
	v := uint8((uint32(0xFCFCF400) >> sel) & 3)
	return gluco2LboolNorm(v)
}

// Gluco2ToIntLbool is the friend toInt(lbool) (inline).
// C: inline int toInt(lbool l)
func Gluco2ToIntLbool(l Gluco2Lbool) int { return int(l.value) }

// Gluco2ToLbool is the friend toLbool(int) (inline).
// C: inline lbool toLbool(int v)
func Gluco2ToLbool(v int) Gluco2Lbool { return gluco2LboolNorm(uint8(v)) }

// The macros l_True, l_False and l_Undef.
var (
	Gluco2LTrue  = Gluco2Lbool{value: 0} // #define l_True  (lbool((uint8_t)0))
	Gluco2LFalse = Gluco2Lbool{value: 1} // #define l_False (lbool((uint8_t)1))
	Gluco2LUndef = Gluco2Lbool{value: 2} // #define l_Undef (lbool((uint8_t)2))
)

// Gluco2CRef is `typedef RegionAllocator<uint32_t>::Ref CRef`.
type Gluco2CRef = uint32

// Gluco2CRefUndef is `const CRef CRef_Undef = RegionAllocator<uint32_t>::Ref_Undef`.
const Gluco2CRefUndef Gluco2CRef = Gluco2RegionAllocatorRefUndef

// Gluco2LitSeq is what the Clause constructor and ClauseAllocator::alloc template
// parameters (V, Lits) require of the literal source: size() and operator[]
// (instantiated with vec<Lit> and Clause).
type Gluco2LitSeq interface {
	Size() int
	At(i int) *Gluco2Lit
}

// Layout of the Clause header (gcc, little endian). The header is three 32-bit words:
//
//	w[0]: mark:2 (bits 0-1), learnt:1 (bit 2), has_extra:1 (bit 3), reloced:1 (bit 4), lbd:26 (bits 5-30), canbedel:1 (bit 31)
//	w[1]: size:32
//	w[2]: szWithoutSelectors:32
const (
	gluco2ClauseMarkMask    = 0x3
	gluco2ClauseLearntBit   = 1 << 2
	gluco2ClauseHasExtraBit = 1 << 3
	gluco2ClauseRelocedBit  = 1 << 4
	gluco2ClauseLbdShift    = 5
	gluco2ClauseLbdMask     = (1<<26 - 1) << gluco2ClauseLbdShift
	gluco2ClauseCanBeDelBit = 1 << 31
	gluco2ClauseHeaderWords = 3 // sizeof(Clause) / sizeof(uint32_t)
)

// Gluco2Clause is Gluco2::Clause. The C++ object is a header followed by a
// flexible array `union { Lit lit; float act; uint32_t abs; CRef rel; } data[0]`
// of 32-bit words living inside a ClauseAllocator region. In Go a Gluco2Clause
// is a view onto those words: w[0] is the header, w[1+i] is data[i]. The
// allocator hands out views of memory[ref:], so the view's reference is
// len(memory)-len(w).
type Gluco2Clause struct {
	w []uint32
}

// newGluco2Clause is the private template constructor Clause(const V& ps, bool use_extra, bool learnt),
// a placement new: w is the memory it constructs into.
// C: template<class V> Clause(const V& ps, bool use_extra, bool learnt)
func newGluco2Clause(w []uint32, ps Gluco2LitSeq, useExtra bool, learnt bool) *Gluco2Clause {
	c := &Gluco2Clause{w: w}
	var h uint32
	if learnt {
		h |= gluco2ClauseLearntBit
	}
	if useExtra {
		h |= gluco2ClauseHasExtraBit
	}
	h |= gluco2ClauseCanBeDelBit // header.lbd = 0; header.canbedel = 1
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
func (c *Gluco2Clause) CalcAbstraction() {
	if !c.HasExtra() {
		panic("Gluco2Clause.CalcAbstraction: no extra field")
	}
	var abstraction uint32
	for i := 0; i < c.Size(); i++ {
		abstraction |= 1 << (uint(Gluco2LitVar(*c.At(i))) & 31)
	}
	c.w[gluco2ClauseHeaderWords+c.Size()] = abstraction
}

// Size is Clause::size.
// C: int size() const
func (c *Gluco2Clause) Size() int {
	return int(c.w[1])
}

func (c *Gluco2Clause) setSize(n int) {
	c.w[1] = uint32(n)
}

// Shrink is Clause::shrink.
// C: void shrink(int i)
func (c *Gluco2Clause) Shrink(i int) {
	if i > c.Size() {
		panic("Gluco2Clause.Shrink: i > size")
	}
	if c.HasExtra() {
		c.w[gluco2ClauseHeaderWords+c.Size()-i] = c.w[gluco2ClauseHeaderWords+c.Size()]
	}
	c.setSize(c.Size() - i)
}

// Pop is Clause::pop.
// C: void pop()
func (c *Gluco2Clause) Pop() { c.Shrink(1) }

// Learnt is Clause::learnt.
// C: bool learnt() const
func (c *Gluco2Clause) Learnt() bool { return c.w[0]&gluco2ClauseLearntBit != 0 }

// HasExtra is Clause::has_extra.
// C: bool has_extra() const
func (c *Gluco2Clause) HasExtra() bool { return c.w[0]&gluco2ClauseHasExtraBit != 0 }

// Mark is the getter overload mark().
// C: uint32_t mark() const
func (c *Gluco2Clause) Mark() uint32 { return c.w[0] & gluco2ClauseMarkMask }

// SetMark is the setter overload mark(uint32_t m).
// C: void mark(uint32_t m)
func (c *Gluco2Clause) SetMark(m uint32) {
	c.w[0] = (c.w[0] &^ gluco2ClauseMarkMask) | (m & gluco2ClauseMarkMask)
}

// Last is Clause::last.
// C: const Lit& last() const
func (c *Gluco2Clause) Last() Gluco2Lit { return *c.At(c.Size() - 1) }

// Reloced is Clause::reloced.
// C: bool reloced() const
func (c *Gluco2Clause) Reloced() bool { return c.w[0]&gluco2ClauseRelocedBit != 0 }

// Relocation is Clause::relocation.
// C: CRef relocation() const
func (c *Gluco2Clause) Relocation() Gluco2CRef { return c.w[gluco2ClauseHeaderWords] }

// Relocate is Clause::relocate.
// C: void relocate(CRef c)
func (c *Gluco2Clause) Relocate(cr Gluco2CRef) {
	c.w[0] |= gluco2ClauseRelocedBit
	c.w[gluco2ClauseHeaderWords] = cr
}

// At is operator[] (const and non-const overloads); a Lit and a uint32 are both
// 4 bytes, so the union member data[i].lit is the word w[1+i] viewed as a Lit.
// C: Lit& operator[](int i)
func (c *Gluco2Clause) At(i int) *Gluco2Lit {
	return (*Gluco2Lit)(unsafe.Pointer(&c.w[gluco2ClauseHeaderWords+i]))
}

// Lits is `operator const Lit*() const`: the literals as a slice.
// C: operator const Lit*(void) const
func (c *Gluco2Clause) Lits() []Gluco2Lit {
	return unsafe.Slice((*Gluco2Lit)(unsafe.Pointer(&c.w[gluco2ClauseHeaderWords])), c.Size())
}

// Activity is Clause::activity (the float member of the extra union).
// C: float& activity()
func (c *Gluco2Clause) Activity() *float32 {
	if !c.HasExtra() {
		panic("Gluco2Clause.Activity: no extra field")
	}
	return (*float32)(unsafe.Pointer(&c.w[gluco2ClauseHeaderWords+c.Size()]))
}

// Abstraction is Clause::abstraction.
// C: uint32_t abstraction() const
func (c *Gluco2Clause) Abstraction() uint32 {
	if !c.HasExtra() {
		panic("Gluco2Clause.Abstraction: no extra field")
	}
	return c.w[gluco2ClauseHeaderWords+c.Size()]
}

// SetLBD is Clause::setLBD.
// C: void setLBD(int i)
func (c *Gluco2Clause) SetLBD(i int) {
	c.w[0] = (c.w[0] &^ gluco2ClauseLbdMask) | (uint32(i)<<gluco2ClauseLbdShift)&gluco2ClauseLbdMask
}

// Lbd is Clause::lbd.
// C: unsigned int lbd() const
func (c *Gluco2Clause) Lbd() uint32 { return (c.w[0] & gluco2ClauseLbdMask) >> gluco2ClauseLbdShift }

// SetCanBeDel is Clause::setCanBeDel.
// C: void setCanBeDel(bool b)
func (c *Gluco2Clause) SetCanBeDel(b bool) {
	if b {
		c.w[0] |= gluco2ClauseCanBeDelBit
	} else {
		c.w[0] &^= gluco2ClauseCanBeDelBit
	}
}

// CanBeDel is Clause::canBeDel.
// C: bool canBeDel()
func (c *Gluco2Clause) CanBeDel() bool { return c.w[0]&gluco2ClauseCanBeDelBit != 0 }

// SetSizeWithoutSelectors is Clause::setSizeWithoutSelectors.
// C: void setSizeWithoutSelectors(unsigned int n)
func (c *Gluco2Clause) SetSizeWithoutSelectors(n uint32) { c.w[2] = n }

// SizeWithoutSelectors is Clause::sizeWithoutSelectors.
// C: unsigned int sizeWithoutSelectors() const
func (c *Gluco2Clause) SizeWithoutSelectors() uint32 { return c.w[2] }

// Subsumes is Clause::subsumes (defined inline after the classes): if *this
// subsumes 'other' (or 'other' with exactly one literal negated) it returns
// lit_Undef (resp. the literal that was negated), else lit_Error.
// C: inline Lit Clause::subsumes(const Clause& other) const
func (c *Gluco2Clause) Subsumes(other *Gluco2Clause) Gluco2Lit {
	if c.Learnt() || other.Learnt() {
		panic("Gluco2Clause.Subsumes: learnt clause")
	}
	if !c.HasExtra() || !other.HasExtra() {
		panic("Gluco2Clause.Subsumes: no extra field")
	}
	if other.Size() < c.Size() || (c.Abstraction()&^other.Abstraction()) != 0 {
		return Gluco2LitError
	}
	ret := Gluco2LitUndef
	cs := c.Lits()
	ds := other.Lits()
outer:
	for i := 0; i < c.Size(); i++ {
		// search for c[i] or ~c[i]
		for j := 0; j < other.Size(); j++ {
			if cs[i] == ds[j] {
				continue outer
			} else if ret == Gluco2LitUndef && cs[i] == Gluco2LitNot(ds[j]) {
				ret = cs[i]
				continue outer
			}
		}
		// did not find it
		return Gluco2LitError
	}
	return ret
}

// Strengthen is Clause::strengthen (defined inline after the classes).
// C: inline void Clause::strengthen(Lit p)
func (c *Gluco2Clause) Strengthen(p Gluco2Lit) {
	gluco2Remove[Gluco2Lit](c, p, gluco2Ne[Gluco2Lit])
	c.CalcAbstraction()
}

// Gluco2ClauseAllocator is Gluco2::ClauseAllocator, a RegionAllocator<uint32_t>.
type Gluco2ClauseAllocator struct {
	Gluco2RegionAllocator[uint32]
	ExtraClauseField bool // extra_clause_field
}

// clauseWord32Size is ClauseAllocator::clauseWord32Size (private static).
// C: static int clauseWord32Size(int size, bool has_extra)
func gluco2ClauseAllocatorClauseWord32Size(size int, hasExtra bool) int {
	extra := 0
	if hasExtra {
		extra = 1
	}
	return gluco2ClauseHeaderWords + size + extra
}

// NewGluco2ClauseAllocatorCap is ClauseAllocator(uint32_t start_cap).
// C: ClauseAllocator(uint32_t start_cap)
func NewGluco2ClauseAllocatorCap(startCap uint32) *Gluco2ClauseAllocator {
	a := &Gluco2ClauseAllocator{}
	a.capacity(startCap)
	return a
}

// NewGluco2ClauseAllocator is ClauseAllocator(): the base allocator is
// default constructed (RegionAllocator's start_cap defaults to 1024*1024).
// C: ClauseAllocator()
func NewGluco2ClauseAllocator() *Gluco2ClauseAllocator {
	a := &Gluco2ClauseAllocator{}
	a.capacity(1024 * 1024)
	return a
}

// MoveTo is ClauseAllocator::moveTo.
// C: void moveTo(ClauseAllocator& to)
func (a *Gluco2ClauseAllocator) MoveTo(to *Gluco2ClauseAllocator) {
	to.ExtraClauseField = a.ExtraClauseField
	a.Gluco2RegionAllocator.MoveTo(&to.Gluco2RegionAllocator)
}

// Alloc is the template ClauseAllocator::alloc (learnt defaults to false).
// C: template<class Lits> CRef alloc(const Lits& ps, bool learnt = false)
func (a *Gluco2ClauseAllocator) Alloc(ps Gluco2LitSeq, learnt bool) Gluco2CRef {
	useExtra := learnt || a.ExtraClauseField
	cid := a.Gluco2RegionAllocator.Alloc(gluco2ClauseAllocatorClauseWord32Size(ps.Size(), useExtra))
	newGluco2Clause(a.memory[cid:], ps, useExtra, learnt)
	return cid
}

// At is operator[] (const and non-const overloads).
// C: Clause& operator[](Ref r)
func (a *Gluco2ClauseAllocator) At(r Gluco2CRef) *Gluco2Clause {
	if r >= a.sz {
		panic("Gluco2ClauseAllocator.At: reference out of range")
	}
	return &Gluco2Clause{w: a.memory[r:]}
}

// Lea is ClauseAllocator::lea.
// C: Clause* lea(Ref r)
func (a *Gluco2ClauseAllocator) Lea(r Gluco2CRef) *Gluco2Clause { return a.At(r) }

// Ael is ClauseAllocator::ael: the reference of a clause view.
// C: Ref ael(const Clause* t)
func (a *Gluco2ClauseAllocator) Ael(t *Gluco2Clause) Gluco2CRef {
	return Gluco2CRef(len(a.memory) - len(t.w))
}

// Free_ is ClauseAllocator::free_ (renamed from _free in Minisat).
// C: void free_(CRef cid)
func (a *Gluco2ClauseAllocator) Free_(cid Gluco2CRef) {
	c := a.At(cid)
	a.Gluco2RegionAllocator.Free_(gluco2ClauseAllocatorClauseWord32Size(c.Size(), c.HasExtra()))
}

// Reloc is ClauseAllocator::reloc.
// C: void reloc(CRef& cr, ClauseAllocator& to)
func (a *Gluco2ClauseAllocator) Reloc(cr *Gluco2CRef, to *Gluco2ClauseAllocator) {
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

// Gluco2OccIdx is what OccLists' index type may be (toInt is defined for Var and Lit).
type Gluco2OccIdx interface {
	int | Gluco2Lit
}

// gluco2OccToInt is the toInt() overload selected by OccLists' Idx.
func gluco2OccToInt[Idx Gluco2OccIdx](idx Idx) int {
	switch v := any(idx).(type) {
	case int:
		return Gluco2ToIntVar(v)
	case Gluco2Lit:
		return Gluco2ToIntLit(v)
	}
	panic("unreachable")
}

// Gluco2Deleted is the requirement on OccLists' template argument Deleted: an
// object with bool operator()(const E&) over the element type E of the Vec.
type Gluco2Deleted[E any] interface {
	Call(e E) bool
}

// Gluco2OccLists is OccLists<Idx, Vec, Deleted> where Vec is vec<E>.
type Gluco2OccLists[Idx Gluco2OccIdx, E any, D Gluco2Deleted[E]] struct {
	occs    Gluco2Vec[Gluco2Vec[E]]
	dirty   Gluco2Vec[int8]
	dirties Gluco2Vec[Idx]
	deleted D
}

// NewGluco2OccLists is OccLists(const Deleted& d).
// C: OccLists(const Deleted& d)
func NewGluco2OccLists[Idx Gluco2OccIdx, E any, D Gluco2Deleted[E]](d D) *Gluco2OccLists[Idx, E, D] {
	return &Gluco2OccLists[Idx, E, D]{deleted: d}
}

// Init is OccLists::init.
// C: void init(const Idx& idx)
func (o *Gluco2OccLists[Idx, E, D]) Init(idx Idx) {
	o.occs.GrowTo(gluco2OccToInt(idx) + 1)
	o.dirty.GrowToPad(gluco2OccToInt(idx)+1, 0)
}

// At is OccLists::operator[].
// C: Vec& operator[](const Idx& idx)
func (o *Gluco2OccLists[Idx, E, D]) At(idx Idx) *Gluco2Vec[E] {
	return o.occs.At(gluco2OccToInt(idx))
}

// Lookup is OccLists::lookup.
// C: Vec& lookup(const Idx& idx)
func (o *Gluco2OccLists[Idx, E, D]) Lookup(idx Idx) *Gluco2Vec[E] {
	if *o.dirty.At(gluco2OccToInt(idx)) != 0 {
		o.Clean(idx)
	}
	return o.occs.At(gluco2OccToInt(idx))
}

// CleanAll is OccLists::cleanAll.
// C: void cleanAll()
func (o *Gluco2OccLists[Idx, E, D]) CleanAll() {
	for i := 0; i < o.dirties.Size(); i++ {
		// Dirties may contain duplicates so check here if a variable is already cleaned:
		if *o.dirty.At(gluco2OccToInt(*o.dirties.At(i))) != 0 {
			o.Clean(*o.dirties.At(i))
		}
	}
	o.dirties.Clear(false)
}

// Clean is OccLists::clean.
// C: void clean(const Idx& idx)
func (o *Gluco2OccLists[Idx, E, D]) Clean(idx Idx) {
	vec := o.occs.At(gluco2OccToInt(idx))
	var i, j int
	for i = 0; i < vec.Size(); i++ {
		if !o.deleted.Call(*vec.At(i)) {
			*vec.At(j) = *vec.At(i)
			j++
		}
	}
	vec.Shrink(i - j)
	*o.dirty.At(gluco2OccToInt(idx)) = 0
}

// Smudge is OccLists::smudge.
// C: void smudge(const Idx& idx)
func (o *Gluco2OccLists[Idx, E, D]) Smudge(idx Idx) {
	if *o.dirty.At(gluco2OccToInt(idx)) == 0 {
		*o.dirty.At(gluco2OccToInt(idx)) = 1
		o.dirties.PushElem(idx)
	}
}

// Clear is OccLists::clear (free defaults to true).
// C: void clear(bool free = true)
func (o *Gluco2OccLists[Idx, E, D]) Clear(free bool) {
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

// Gluco2CRefHash is CMap<T>::CRefHash.
type Gluco2CRefHash struct{}

// Call is CRefHash::operator().
// C: uint32_t operator()(CRef cr) const
func (h Gluco2CRefHash) Call(cr Gluco2CRef) uint32 { return uint32(cr) }

// Gluco2CMap is Gluco2::CMap<T>: a Map<CRef, T, CRefHash>.
type Gluco2CMap[T any] struct {
	map_ Gluco2Map[Gluco2CRef, T, Gluco2CRefHash, Gluco2Equal[Gluco2CRef]]
}

// Clear is CMap::clear.
// C: void clear()
func (m *Gluco2CMap[T]) Clear() { m.map_.Clear() }

// Size is CMap::size.
// C: int size() const
func (m *Gluco2CMap[T]) Size() int { return m.map_.Elems() }

// Insert is CMap::insert.
// C: void insert(CRef cr, const T& t)
func (m *Gluco2CMap[T]) Insert(cr Gluco2CRef, t T) { m.map_.Insert(cr, t) }

// GrowTo is CMap::growTo (for compatibility; same as insert).
// C: void growTo(CRef cr, const T& t)
func (m *Gluco2CMap[T]) GrowTo(cr Gluco2CRef, t T) { m.map_.Insert(cr, t) }

// Remove is CMap::remove.
// C: void remove(CRef cr)
func (m *Gluco2CMap[T]) Remove(cr Gluco2CRef) { m.map_.Remove(cr) }

// Has is CMap::has.
// C: bool has(CRef cr, T& t)
func (m *Gluco2CMap[T]) Has(cr Gluco2CRef, t *T) bool { return m.map_.Peek(cr, t) }

// At is operator[] (const and non-const overloads).
// C: T& operator[](CRef cr)
func (m *Gluco2CMap[T]) At(cr Gluco2CRef) *T { return m.map_.At(cr) }

// BucketCount is CMap::bucket_count.
// C: int bucket_count() const
func (m *Gluco2CMap[T]) BucketCount() int { return m.map_.BucketCount() }

// Bucket is CMap::bucket.
// C: const vec<typename HashTable::Pair>& bucket(int i) const
func (m *Gluco2CMap[T]) Bucket(i int) *Gluco2Vec[Gluco2MapPair[Gluco2CRef, T]] {
	return m.map_.Bucket(i)
}

// MoveTo is CMap::moveTo.
// C: void moveTo(CMap& other)
func (m *Gluco2CMap[T]) MoveTo(other *Gluco2CMap[T]) { m.map_.MoveTo(&other.map_) }

// Debug is CMap::debug.
// C: void debug()
func (m *Gluco2CMap[T]) Debug() {
	println(" --- size =", m.Size(), ", bucket_count =", m.map_.BucketCount())
}
