package abc

// Hand-checked skeleton of src/sat/glucose/SolverTypes.h (namespace Gluco).
// Differs from the Minisat version by the LBD / canbedel / sizeWithoutSelectors clause fields.
// See skel_sat_bsat2_Vec_h.go for the C++ -> Go conventions.

// GlucoVar is `typedef int Var`.
type GlucoVar = int

// GlucoVarUndef is the macro `#define var_Undef (-1)`.
const GlucoVarUndef = -1

// GlucoLit is Gluco::Lit: a literal, x = 2*var + sign.
type GlucoLit struct {
	x int
}

// Eq is Lit::operator==.
// C: bool operator==(Lit p) const
func (l GlucoLit) Eq(p GlucoLit) bool {
	panic("TODO, stubbed")
}

// Ne is Lit::operator!=.
// C: bool operator!=(Lit p) const
func (l GlucoLit) Ne(p GlucoLit) bool {
	panic("TODO, stubbed")
}

// Lt is Lit::operator< (makes p, ~p adjacent in the ordering).
// C: bool operator<(Lit p) const
func (l GlucoLit) Lt(p GlucoLit) bool {
	panic("TODO, stubbed")
}

// GlucoMkLit is mkLit (inline; sign defaults to false).
// C: inline Lit mkLit(Var var, bool sign = false)
func GlucoMkLit(v GlucoVar, sign bool) GlucoLit {
	panic("TODO, stubbed")
}

// GlucoLitNot is `operator ~` on Lit (inline).
// C: inline Lit operator~(Lit p)
func GlucoLitNot(p GlucoLit) GlucoLit {
	panic("TODO, stubbed")
}

// GlucoLitXor is `operator ^` on (Lit, bool) (inline).
// C: inline Lit operator^(Lit p, bool b)
func GlucoLitXor(p GlucoLit, b bool) GlucoLit {
	panic("TODO, stubbed")
}

// GlucoSign is Gluco::sign (inline).
// C: inline bool sign(Lit p)
func GlucoSign(p GlucoLit) bool {
	panic("TODO, stubbed")
}

// GlucoLitVar is Gluco::var(Lit) (inline); renamed because GlucoVar is the type.
// C: inline int var(Lit p)
func GlucoLitVar(p GlucoLit) int {
	panic("TODO, stubbed")
}

// GlucoToIntVar is the overload toInt(Var) (inline).
// C: inline int toInt(Var v)
func GlucoToIntVar(v GlucoVar) int {
	panic("TODO, stubbed")
}

// GlucoToIntLit is the overload toInt(Lit) (inline).
// C: inline int toInt(Lit p)
func GlucoToIntLit(p GlucoLit) int {
	panic("TODO, stubbed")
}

// GlucoToLit is Gluco::toLit (inline).
// C: inline Lit toLit(int i)
func GlucoToLit(i int) GlucoLit {
	panic("TODO, stubbed")
}

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
func NewGlucoLboolV(v uint8) GlucoLbool {
	panic("TODO, stubbed")
}

// NewGlucoLbool is lbool(): value 0.
// C: lbool()
func NewGlucoLbool() GlucoLbool {
	panic("TODO, stubbed")
}

// NewGlucoLboolB is the explicit constructor lbool(bool x): value !x.
// C: explicit lbool(bool x)
func NewGlucoLboolB(x bool) GlucoLbool {
	panic("TODO, stubbed")
}

// Eq is lbool::operator==.
// C: bool operator==(lbool b) const
func (l GlucoLbool) Eq(b GlucoLbool) bool {
	panic("TODO, stubbed")
}

// Ne is lbool::operator!=.
// C: bool operator!=(lbool b) const
func (l GlucoLbool) Ne(b GlucoLbool) bool {
	panic("TODO, stubbed")
}

// Xor is lbool::operator^.
// C: lbool operator^(bool b) const
func (l GlucoLbool) Xor(b bool) GlucoLbool {
	panic("TODO, stubbed")
}

// And is lbool::operator&&.
// C: lbool operator&&(lbool b) const
func (l GlucoLbool) And(b GlucoLbool) GlucoLbool {
	panic("TODO, stubbed")
}

// Or is lbool::operator||.
// C: lbool operator||(lbool b) const
func (l GlucoLbool) Or(b GlucoLbool) GlucoLbool {
	panic("TODO, stubbed")
}

// GlucoToIntLbool is the friend toInt(lbool) (inline).
// C: inline int toInt(lbool l)
func GlucoToIntLbool(l GlucoLbool) int {
	panic("TODO, stubbed")
}

// GlucoToLbool is the friend toLbool(int) (inline).
// C: inline lbool toLbool(int v)
func GlucoToLbool(v int) GlucoLbool {
	panic("TODO, stubbed")
}

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

// glucoClauseHeader is the anonymous bit-field struct `header` of Clause.
type glucoClauseHeader struct {
	mark               uint32 // 2 bits
	learnt             uint32 // 1 bit
	has_extra          uint32 // 1 bit
	reloced            uint32 // 1 bit
	lbd                uint32 // 26 bits
	canbedel           uint32 // 1 bit
	size               uint32 // 32 bits
	szWithoutSelectors uint32 // 32 bits
}

// GlucoClause is Gluco::Clause. The C++ object is a header followed by a
// flexible array `union { Lit lit; float act; uint32_t abs; CRef rel; } data[0]`
// of 32-bit words living inside a ClauseAllocator region.
type GlucoClause struct {
	header glucoClauseHeader
	data   []uint32
}

// newGlucoClause is the private template constructor Clause(const V& ps, bool use_extra, bool learnt).
// C: template<class V> Clause(const V& ps, bool use_extra, bool learnt)
func newGlucoClause(ps GlucoLitSeq, useExtra bool, learnt bool) *GlucoClause {
	panic("TODO, stubbed")
}

// CalcAbstraction is Clause::calcAbstraction.
// C: void calcAbstraction()
func (c *GlucoClause) CalcAbstraction() {
	panic("TODO, stubbed")
}

// Size is Clause::size.
// C: int size() const
func (c *GlucoClause) Size() int {
	panic("TODO, stubbed")
}

// Shrink is Clause::shrink.
// C: void shrink(int i)
func (c *GlucoClause) Shrink(i int) {
	panic("TODO, stubbed")
}

// Pop is Clause::pop.
// C: void pop()
func (c *GlucoClause) Pop() {
	panic("TODO, stubbed")
}

// Learnt is Clause::learnt.
// C: bool learnt() const
func (c *GlucoClause) Learnt() bool {
	panic("TODO, stubbed")
}

// HasExtra is Clause::has_extra.
// C: bool has_extra() const
func (c *GlucoClause) HasExtra() bool {
	panic("TODO, stubbed")
}

// Mark is the getter overload mark().
// C: uint32_t mark() const
func (c *GlucoClause) Mark() uint32 {
	panic("TODO, stubbed")
}

// SetMark is the setter overload mark(uint32_t m).
// C: void mark(uint32_t m)
func (c *GlucoClause) SetMark(m uint32) {
	panic("TODO, stubbed")
}

// Last is Clause::last.
// C: const Lit& last() const
func (c *GlucoClause) Last() GlucoLit {
	panic("TODO, stubbed")
}

// Reloced is Clause::reloced.
// C: bool reloced() const
func (c *GlucoClause) Reloced() bool {
	panic("TODO, stubbed")
}

// Relocation is Clause::relocation.
// C: CRef relocation() const
func (c *GlucoClause) Relocation() GlucoCRef {
	panic("TODO, stubbed")
}

// Relocate is Clause::relocate.
// C: void relocate(CRef c)
func (c *GlucoClause) Relocate(cr GlucoCRef) {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: Lit& operator[](int i)
func (c *GlucoClause) At(i int) *GlucoLit {
	panic("TODO, stubbed")
}

// Lits is `operator const Lit*() const`.
// C: operator const Lit*(void) const
func (c *GlucoClause) Lits() []GlucoLit {
	panic("TODO, stubbed")
}

// Activity is Clause::activity.
// C: float& activity()
func (c *GlucoClause) Activity() *float32 {
	panic("TODO, stubbed")
}

// Abstraction is Clause::abstraction.
// C: uint32_t abstraction() const
func (c *GlucoClause) Abstraction() uint32 {
	panic("TODO, stubbed")
}

// SetLBD is Clause::setLBD.
// C: void setLBD(int i)
func (c *GlucoClause) SetLBD(i int) {
	panic("TODO, stubbed")
}

// Lbd is Clause::lbd.
// C: unsigned int lbd() const
func (c *GlucoClause) Lbd() uint32 {
	panic("TODO, stubbed")
}

// SetCanBeDel is Clause::setCanBeDel.
// C: void setCanBeDel(bool b)
func (c *GlucoClause) SetCanBeDel(b bool) {
	panic("TODO, stubbed")
}

// CanBeDel is Clause::canBeDel.
// C: bool canBeDel()
func (c *GlucoClause) CanBeDel() bool {
	panic("TODO, stubbed")
}

// SetSizeWithoutSelectors is Clause::setSizeWithoutSelectors.
// C: void setSizeWithoutSelectors(unsigned int n)
func (c *GlucoClause) SetSizeWithoutSelectors(n uint32) {
	panic("TODO, stubbed")
}

// SizeWithoutSelectors is Clause::sizeWithoutSelectors.
// C: unsigned int sizeWithoutSelectors() const
func (c *GlucoClause) SizeWithoutSelectors() uint32 {
	panic("TODO, stubbed")
}

// Subsumes is Clause::subsumes (defined inline after the classes).
// C: inline Lit Clause::subsumes(const Clause& other) const
func (c *GlucoClause) Subsumes(other *GlucoClause) GlucoLit {
	panic("TODO, stubbed")
}

// Strengthen is Clause::strengthen (defined inline after the classes).
// C: inline void Clause::strengthen(Lit p)
func (c *GlucoClause) Strengthen(p GlucoLit) {
	panic("TODO, stubbed")
}

// GlucoClauseAllocator is Gluco::ClauseAllocator, a RegionAllocator<uint32_t>.
type GlucoClauseAllocator struct {
	GlucoRegionAllocator[uint32]
	ExtraClauseField bool // extra_clause_field
}

// clauseWord32Size is ClauseAllocator::clauseWord32Size (private static).
// C: static int clauseWord32Size(int size, bool has_extra)
func glucoClauseAllocatorClauseWord32Size(size int, hasExtra bool) int {
	panic("TODO, stubbed")
}

// NewGlucoClauseAllocatorCap is ClauseAllocator(uint32_t start_cap).
// C: ClauseAllocator(uint32_t start_cap)
func NewGlucoClauseAllocatorCap(startCap uint32) *GlucoClauseAllocator {
	panic("TODO, stubbed")
}

// NewGlucoClauseAllocator is ClauseAllocator().
// C: ClauseAllocator()
func NewGlucoClauseAllocator() *GlucoClauseAllocator {
	panic("TODO, stubbed")
}

// MoveTo is ClauseAllocator::moveTo.
// C: void moveTo(ClauseAllocator& to)
func (a *GlucoClauseAllocator) MoveTo(to *GlucoClauseAllocator) {
	panic("TODO, stubbed")
}

// Alloc is the template ClauseAllocator::alloc (learnt defaults to false).
// C: template<class Lits> CRef alloc(const Lits& ps, bool learnt = false)
func (a *GlucoClauseAllocator) Alloc(ps GlucoLitSeq, learnt bool) GlucoCRef {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: Clause& operator[](Ref r)
func (a *GlucoClauseAllocator) At(r GlucoCRef) *GlucoClause {
	panic("TODO, stubbed")
}

// Lea is ClauseAllocator::lea.
// C: Clause* lea(Ref r)
func (a *GlucoClauseAllocator) Lea(r GlucoCRef) *GlucoClause {
	panic("TODO, stubbed")
}

// Ael is ClauseAllocator::ael.
// C: Ref ael(const Clause* t)
func (a *GlucoClauseAllocator) Ael(t *GlucoClause) GlucoCRef {
	panic("TODO, stubbed")
}

// Free_ is ClauseAllocator::free_ (renamed from _free in Minisat).
// C: void free_(CRef cid)
func (a *GlucoClauseAllocator) Free_(cid GlucoCRef) {
	panic("TODO, stubbed")
}

// Reloc is ClauseAllocator::reloc.
// C: void reloc(CRef& cr, ClauseAllocator& to)
func (a *GlucoClauseAllocator) Reloc(cr *GlucoCRef, to *GlucoClauseAllocator) {
	panic("TODO, stubbed")
}

// GlucoOccIdx is what OccLists' index type may be (toInt is defined for Var and Lit).
type GlucoOccIdx interface {
	int | GlucoLit
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
	panic("TODO, stubbed")
}

// Init is OccLists::init.
// C: void init(const Idx& idx)
func (o *GlucoOccLists[Idx, E, D]) Init(idx Idx) {
	panic("TODO, stubbed")
}

// At is OccLists::operator[].
// C: Vec& operator[](const Idx& idx)
func (o *GlucoOccLists[Idx, E, D]) At(idx Idx) *GlucoVec[E] {
	panic("TODO, stubbed")
}

// Lookup is OccLists::lookup.
// C: Vec& lookup(const Idx& idx)
func (o *GlucoOccLists[Idx, E, D]) Lookup(idx Idx) *GlucoVec[E] {
	panic("TODO, stubbed")
}

// CleanAll is OccLists::cleanAll.
// C: void cleanAll()
func (o *GlucoOccLists[Idx, E, D]) CleanAll() {
	panic("TODO, stubbed")
}

// Clean is OccLists::clean.
// C: void clean(const Idx& idx)
func (o *GlucoOccLists[Idx, E, D]) Clean(idx Idx) {
	panic("TODO, stubbed")
}

// Smudge is OccLists::smudge.
// C: void smudge(const Idx& idx)
func (o *GlucoOccLists[Idx, E, D]) Smudge(idx Idx) {
	panic("TODO, stubbed")
}

// Clear is OccLists::clear (free defaults to true).
// C: void clear(bool free = true)
func (o *GlucoOccLists[Idx, E, D]) Clear(free bool) {
	panic("TODO, stubbed")
}

// GlucoCRefHash is CMap<T>::CRefHash.
type GlucoCRefHash struct{}

// Call is CRefHash::operator().
// C: uint32_t operator()(CRef cr) const
func (h GlucoCRefHash) Call(cr GlucoCRef) uint32 {
	panic("TODO, stubbed")
}

// GlucoCMap is Gluco::CMap<T>: a Map<CRef, T, CRefHash>.
type GlucoCMap[T any] struct {
	map_ GlucoMap[GlucoCRef, T, GlucoCRefHash, GlucoEqual[GlucoCRef]]
}

// Clear is CMap::clear.
// C: void clear()
func (m *GlucoCMap[T]) Clear() {
	panic("TODO, stubbed")
}

// Size is CMap::size.
// C: int size() const
func (m *GlucoCMap[T]) Size() int {
	panic("TODO, stubbed")
}

// Insert is CMap::insert.
// C: void insert(CRef cr, const T& t)
func (m *GlucoCMap[T]) Insert(cr GlucoCRef, t T) {
	panic("TODO, stubbed")
}

// GrowTo is CMap::growTo (for compatibility; same as insert).
// C: void growTo(CRef cr, const T& t)
func (m *GlucoCMap[T]) GrowTo(cr GlucoCRef, t T) {
	panic("TODO, stubbed")
}

// Remove is CMap::remove.
// C: void remove(CRef cr)
func (m *GlucoCMap[T]) Remove(cr GlucoCRef) {
	panic("TODO, stubbed")
}

// Has is CMap::has.
// C: bool has(CRef cr, T& t)
func (m *GlucoCMap[T]) Has(cr GlucoCRef, t *T) bool {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](CRef cr)
func (m *GlucoCMap[T]) At(cr GlucoCRef) *T {
	panic("TODO, stubbed")
}

// BucketCount is CMap::bucket_count.
// C: int bucket_count() const
func (m *GlucoCMap[T]) BucketCount() int {
	panic("TODO, stubbed")
}

// Bucket is CMap::bucket.
// C: const vec<typename HashTable::Pair>& bucket(int i) const
func (m *GlucoCMap[T]) Bucket(i int) *GlucoVec[GlucoMapPair[GlucoCRef, T]] {
	panic("TODO, stubbed")
}

// MoveTo is CMap::moveTo.
// C: void moveTo(CMap& other)
func (m *GlucoCMap[T]) MoveTo(other *GlucoCMap[T]) {
	panic("TODO, stubbed")
}

// Debug is CMap::debug.
// C: void debug()
func (m *GlucoCMap[T]) Debug() {
	panic("TODO, stubbed")
}
