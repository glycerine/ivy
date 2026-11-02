package abc

// Hand-checked skeleton of src/sat/bsat2/SolverTypes.h (namespace Minisat).
// See skel_sat_bsat2_Vec_h.go for the C++ -> Go conventions.

// MinisatVar is `typedef int Var`.
type MinisatVar = int

// MinisatLit is Minisat::Lit: a literal, x = 2*var + sign.
type MinisatLit struct {
	x int
}

// Eq is Lit::operator==.
// C: bool operator==(Lit p) const
func (l MinisatLit) Eq(p MinisatLit) bool {
	panic("TODO, stubbed")
}

// Ne is Lit::operator!=.
// C: bool operator!=(Lit p) const
func (l MinisatLit) Ne(p MinisatLit) bool {
	panic("TODO, stubbed")
}

// Lt is Lit::operator< (makes p, ~p adjacent in the ordering).
// C: bool operator<(Lit p) const
func (l MinisatLit) Lt(p MinisatLit) bool {
	panic("TODO, stubbed")
}

// MinisatMkLit is mkLit (inline; sign defaults to false).
// C: inline Lit mkLit(Var var, bool sign = false)
func MinisatMkLit(v MinisatVar, sign bool) MinisatLit {
	panic("TODO, stubbed")
}

// MinisatLitNot is `operator ~` on Lit (inline).
// C: inline Lit operator~(Lit p)
func MinisatLitNot(p MinisatLit) MinisatLit {
	panic("TODO, stubbed")
}

// MinisatLitXor is `operator ^` on (Lit, bool) (inline).
// C: inline Lit operator^(Lit p, bool b)
func MinisatLitXor(p MinisatLit, b bool) MinisatLit {
	panic("TODO, stubbed")
}

// MinisatSign is Minisat::sign (inline).
// C: inline bool sign(Lit p)
func MinisatSign(p MinisatLit) bool {
	panic("TODO, stubbed")
}

// MinisatLitVar is Minisat::var(Lit) (inline); renamed because MinisatVar is the type.
// C: inline int var(Lit p)
func MinisatLitVar(p MinisatLit) int {
	panic("TODO, stubbed")
}

// MinisatToIntVar is the overload toInt(Var) (inline).
// C: inline int toInt(Var v)
func MinisatToIntVar(v MinisatVar) int {
	panic("TODO, stubbed")
}

// MinisatToIntLit is the overload toInt(Lit) (inline).
// C: inline int toInt(Lit p)
func MinisatToIntLit(p MinisatLit) int {
	panic("TODO, stubbed")
}

// MinisatToLit is Minisat::toLit (inline).
// C: inline Lit toLit(int i)
func MinisatToLit(i int) MinisatLit {
	panic("TODO, stubbed")
}

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
func NewMinisatLboolV(v uint8) MinisatLbool {
	panic("TODO, stubbed")
}

// NewMinisatLbool is lbool(): value 0.
// C: lbool()
func NewMinisatLbool() MinisatLbool {
	panic("TODO, stubbed")
}

// NewMinisatLboolB is the explicit constructor lbool(bool x): value !x.
// C: explicit lbool(bool x)
func NewMinisatLboolB(x bool) MinisatLbool {
	panic("TODO, stubbed")
}

// Eq is lbool::operator==.
// C: bool operator==(lbool b) const
func (l MinisatLbool) Eq(b MinisatLbool) bool {
	panic("TODO, stubbed")
}

// Ne is lbool::operator!=.
// C: bool operator!=(lbool b) const
func (l MinisatLbool) Ne(b MinisatLbool) bool {
	panic("TODO, stubbed")
}

// Xor is lbool::operator^.
// C: lbool operator^(bool b) const
func (l MinisatLbool) Xor(b bool) MinisatLbool {
	panic("TODO, stubbed")
}

// And is lbool::operator&&.
// C: lbool operator&&(lbool b) const
func (l MinisatLbool) And(b MinisatLbool) MinisatLbool {
	panic("TODO, stubbed")
}

// Or is lbool::operator||.
// C: lbool operator||(lbool b) const
func (l MinisatLbool) Or(b MinisatLbool) MinisatLbool {
	panic("TODO, stubbed")
}

// MinisatToIntLbool is the friend toInt(lbool) (inline).
// C: inline int toInt(lbool l)
func MinisatToIntLbool(l MinisatLbool) int {
	panic("TODO, stubbed")
}

// MinisatToLbool is the friend toLbool(int) (inline).
// C: inline lbool toLbool(int v)
func MinisatToLbool(v int) MinisatLbool {
	panic("TODO, stubbed")
}

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

// minisatClauseHeader is the anonymous bit-field struct `header` of Clause.
type minisatClauseHeader struct {
	mark      uint32 // 2 bits
	learnt    uint32 // 1 bit
	has_extra uint32 // 1 bit
	reloced   uint32 // 1 bit
	size      uint32 // 27 bits
}

// MinisatClause is Minisat::Clause. The C++ object is a header followed by a
// flexible array `union { Lit lit; float act; uint32_t abs; CRef rel; } data[0]`
// of 32-bit words living inside a ClauseAllocator region.
type MinisatClause struct {
	header minisatClauseHeader
	data   []uint32
}

// newMinisatClause is the private template constructor Clause(const V& ps, bool use_extra, bool learnt).
// C: template<class V> Clause(const V& ps, bool use_extra, bool learnt)
func newMinisatClause(ps MinisatLitSeq, useExtra bool, learnt bool) *MinisatClause {
	panic("TODO, stubbed")
}

// CalcAbstraction is Clause::calcAbstraction.
// C: void calcAbstraction()
func (c *MinisatClause) CalcAbstraction() {
	panic("TODO, stubbed")
}

// Size is Clause::size.
// C: int size() const
func (c *MinisatClause) Size() int {
	panic("TODO, stubbed")
}

// Shrink is Clause::shrink.
// C: void shrink(int i)
func (c *MinisatClause) Shrink(i int) {
	panic("TODO, stubbed")
}

// Pop is Clause::pop.
// C: void pop()
func (c *MinisatClause) Pop() {
	panic("TODO, stubbed")
}

// Learnt is Clause::learnt.
// C: bool learnt() const
func (c *MinisatClause) Learnt() bool {
	panic("TODO, stubbed")
}

// HasExtra is Clause::has_extra.
// C: bool has_extra() const
func (c *MinisatClause) HasExtra() bool {
	panic("TODO, stubbed")
}

// Mark is the getter overload mark().
// C: uint32_t mark() const
func (c *MinisatClause) Mark() uint32 {
	panic("TODO, stubbed")
}

// SetMark is the setter overload mark(uint32_t m).
// C: void mark(uint32_t m)
func (c *MinisatClause) SetMark(m uint32) {
	panic("TODO, stubbed")
}

// Last is Clause::last.
// C: const Lit& last() const
func (c *MinisatClause) Last() MinisatLit {
	panic("TODO, stubbed")
}

// Reloced is Clause::reloced.
// C: bool reloced() const
func (c *MinisatClause) Reloced() bool {
	panic("TODO, stubbed")
}

// Relocation is Clause::relocation.
// C: CRef relocation() const
func (c *MinisatClause) Relocation() MinisatCRef {
	panic("TODO, stubbed")
}

// Relocate is Clause::relocate.
// C: void relocate(CRef c)
func (c *MinisatClause) Relocate(cr MinisatCRef) {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: Lit& operator[](int i)
func (c *MinisatClause) At(i int) *MinisatLit {
	panic("TODO, stubbed")
}

// Lits is `operator const Lit*() const`.
// C: operator const Lit*(void) const
func (c *MinisatClause) Lits() []MinisatLit {
	panic("TODO, stubbed")
}

// Activity is Clause::activity.
// C: float& activity()
func (c *MinisatClause) Activity() *float32 {
	panic("TODO, stubbed")
}

// Abstraction is Clause::abstraction.
// C: uint32_t abstraction() const
func (c *MinisatClause) Abstraction() uint32 {
	panic("TODO, stubbed")
}

// Subsumes is Clause::subsumes (defined inline after the classes).
// C: inline Lit Clause::subsumes(const Clause& other) const
func (c *MinisatClause) Subsumes(other *MinisatClause) MinisatLit {
	panic("TODO, stubbed")
}

// Strengthen is Clause::strengthen (defined inline after the classes).
// C: inline void Clause::strengthen(Lit p)
func (c *MinisatClause) Strengthen(p MinisatLit) {
	panic("TODO, stubbed")
}

// MinisatClauseAllocator is Minisat::ClauseAllocator, a RegionAllocator<uint32_t>.
type MinisatClauseAllocator struct {
	MinisatRegionAllocator[uint32]
	ExtraClauseField bool // extra_clause_field
}

// clauseWord32Size is ClauseAllocator::clauseWord32Size (private static).
// C: static int clauseWord32Size(int size, bool has_extra)
func minisatClauseAllocatorClauseWord32Size(size int, hasExtra bool) int {
	panic("TODO, stubbed")
}

// NewMinisatClauseAllocatorCap is ClauseAllocator(uint32_t start_cap).
// C: ClauseAllocator(uint32_t start_cap)
func NewMinisatClauseAllocatorCap(startCap uint32) *MinisatClauseAllocator {
	panic("TODO, stubbed")
}

// NewMinisatClauseAllocator is ClauseAllocator().
// C: ClauseAllocator()
func NewMinisatClauseAllocator() *MinisatClauseAllocator {
	panic("TODO, stubbed")
}

// MoveTo is ClauseAllocator::moveTo.
// C: void moveTo(ClauseAllocator& to)
func (a *MinisatClauseAllocator) MoveTo(to *MinisatClauseAllocator) {
	panic("TODO, stubbed")
}

// Alloc is the template ClauseAllocator::alloc (learnt defaults to false).
// C: template<class Lits> CRef alloc(const Lits& ps, bool learnt = false)
func (a *MinisatClauseAllocator) Alloc(ps MinisatLitSeq, learnt bool) MinisatCRef {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: Clause& operator[](Ref r)
func (a *MinisatClauseAllocator) At(r MinisatCRef) *MinisatClause {
	panic("TODO, stubbed")
}

// Lea is ClauseAllocator::lea.
// C: Clause* lea(Ref r)
func (a *MinisatClauseAllocator) Lea(r MinisatCRef) *MinisatClause {
	panic("TODO, stubbed")
}

// Ael is ClauseAllocator::ael.
// C: Ref ael(const Clause* t)
func (a *MinisatClauseAllocator) Ael(t *MinisatClause) MinisatCRef {
	panic("TODO, stubbed")
}

// Free is ClauseAllocator::_free.
// C: void _free(CRef cid)
func (a *MinisatClauseAllocator) Free(cid MinisatCRef) {
	panic("TODO, stubbed")
}

// Reloc is ClauseAllocator::reloc.
// C: void reloc(CRef& cr, ClauseAllocator& to)
func (a *MinisatClauseAllocator) Reloc(cr *MinisatCRef, to *MinisatClauseAllocator) {
	panic("TODO, stubbed")
}

// MinisatOccIdx is what OccLists' index type may be (toInt is defined for Var and Lit).
type MinisatOccIdx interface {
	int | MinisatLit
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
	panic("TODO, stubbed")
}

// Init is OccLists::init.
// C: void init(const Idx& idx)
func (o *MinisatOccLists[Idx, E, D]) Init(idx Idx) {
	panic("TODO, stubbed")
}

// At is OccLists::operator[].
// C: Vec& operator[](const Idx& idx)
func (o *MinisatOccLists[Idx, E, D]) At(idx Idx) *MinisatVec[E] {
	panic("TODO, stubbed")
}

// Lookup is OccLists::lookup.
// C: Vec& lookup(const Idx& idx)
func (o *MinisatOccLists[Idx, E, D]) Lookup(idx Idx) *MinisatVec[E] {
	panic("TODO, stubbed")
}

// CleanAll is OccLists::cleanAll.
// C: void cleanAll()
func (o *MinisatOccLists[Idx, E, D]) CleanAll() {
	panic("TODO, stubbed")
}

// Clean is OccLists::clean.
// C: void clean(const Idx& idx)
func (o *MinisatOccLists[Idx, E, D]) Clean(idx Idx) {
	panic("TODO, stubbed")
}

// Smudge is OccLists::smudge.
// C: void smudge(const Idx& idx)
func (o *MinisatOccLists[Idx, E, D]) Smudge(idx Idx) {
	panic("TODO, stubbed")
}

// Clear is OccLists::clear (free defaults to true).
// C: void clear(bool free = true)
func (o *MinisatOccLists[Idx, E, D]) Clear(free bool) {
	panic("TODO, stubbed")
}

// MinisatCRefHash is CMap<T>::CRefHash.
type MinisatCRefHash struct{}

// Call is CRefHash::operator().
// C: uint32_t operator()(CRef cr) const
func (h MinisatCRefHash) Call(cr MinisatCRef) uint32 {
	panic("TODO, stubbed")
}

// MinisatCMap is Minisat::CMap<T>: a Map<CRef, T, CRefHash>.
type MinisatCMap[T any] struct {
	map_ MinisatMap[MinisatCRef, T, MinisatCRefHash, MinisatEqual[MinisatCRef]]
}

// Clear is CMap::clear.
// C: void clear()
func (m *MinisatCMap[T]) Clear() {
	panic("TODO, stubbed")
}

// Size is CMap::size.
// C: int size() const
func (m *MinisatCMap[T]) Size() int {
	panic("TODO, stubbed")
}

// Insert is CMap::insert.
// C: void insert(CRef cr, const T& t)
func (m *MinisatCMap[T]) Insert(cr MinisatCRef, t T) {
	panic("TODO, stubbed")
}

// GrowTo is CMap::growTo (for compatibility; same as insert).
// C: void growTo(CRef cr, const T& t)
func (m *MinisatCMap[T]) GrowTo(cr MinisatCRef, t T) {
	panic("TODO, stubbed")
}

// Remove is CMap::remove.
// C: void remove(CRef cr)
func (m *MinisatCMap[T]) Remove(cr MinisatCRef) {
	panic("TODO, stubbed")
}

// Has is CMap::has.
// C: bool has(CRef cr, T& t)
func (m *MinisatCMap[T]) Has(cr MinisatCRef, t *T) bool {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](CRef cr)
func (m *MinisatCMap[T]) At(cr MinisatCRef) *T {
	panic("TODO, stubbed")
}

// BucketCount is CMap::bucket_count.
// C: int bucket_count() const
func (m *MinisatCMap[T]) BucketCount() int {
	panic("TODO, stubbed")
}

// Bucket is CMap::bucket.
// C: const vec<typename HashTable::Pair>& bucket(int i) const
func (m *MinisatCMap[T]) Bucket(i int) *MinisatVec[MinisatMapPair[MinisatCRef, T]] {
	panic("TODO, stubbed")
}

// MoveTo is CMap::moveTo.
// C: void moveTo(CMap& other)
func (m *MinisatCMap[T]) MoveTo(other *MinisatCMap[T]) {
	panic("TODO, stubbed")
}

// Debug is CMap::debug.
// C: void debug()
func (m *MinisatCMap[T]) Debug() {
	panic("TODO, stubbed")
}
