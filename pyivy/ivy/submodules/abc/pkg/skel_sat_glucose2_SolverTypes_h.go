package abc

// Hand-checked skeleton of src/sat/glucose2/SolverTypes.h (namespace Gluco2).
// Differs from the Minisat version by the LBD / canbedel / sizeWithoutSelectors clause fields (and, under CGLUCOSE_EXP, friend class Solver).
// See skel_sat_bsat2_Vec_h.go for the C++ -> Go conventions.

// Gluco2Var is `typedef int Var`.
type Gluco2Var = int

// Gluco2VarUndef is the macro `#define var_Undef (-1)`.
const Gluco2VarUndef = -1

// Gluco2Lit is Gluco2::Lit: a literal, x = 2*var + sign.
type Gluco2Lit struct {
	x int
}

// Eq is Lit::operator==.
// C: bool operator==(Lit p) const
func (l Gluco2Lit) Eq(p Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// Ne is Lit::operator!=.
// C: bool operator!=(Lit p) const
func (l Gluco2Lit) Ne(p Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// Lt is Lit::operator< (makes p, ~p adjacent in the ordering).
// C: bool operator<(Lit p) const
func (l Gluco2Lit) Lt(p Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// Gluco2MkLit is mkLit (inline; sign defaults to false).
// C: inline Lit mkLit(Var var, bool sign = false)
func Gluco2MkLit(v Gluco2Var, sign bool) Gluco2Lit {
	panic("TODO, stubbed")
}

// Gluco2LitNot is `operator ~` on Lit (inline).
// C: inline Lit operator~(Lit p)
func Gluco2LitNot(p Gluco2Lit) Gluco2Lit {
	panic("TODO, stubbed")
}

// Gluco2LitXor is `operator ^` on (Lit, bool) (inline).
// C: inline Lit operator^(Lit p, bool b)
func Gluco2LitXor(p Gluco2Lit, b bool) Gluco2Lit {
	panic("TODO, stubbed")
}

// Gluco2Sign is Gluco2::sign (inline).
// C: inline bool sign(Lit p)
func Gluco2Sign(p Gluco2Lit) bool {
	panic("TODO, stubbed")
}

// Gluco2LitVar is Gluco2::var(Lit) (inline); renamed because Gluco2Var is the type.
// C: inline int var(Lit p)
func Gluco2LitVar(p Gluco2Lit) int {
	panic("TODO, stubbed")
}

// Gluco2ToIntVar is the overload toInt(Var) (inline).
// C: inline int toInt(Var v)
func Gluco2ToIntVar(v Gluco2Var) int {
	panic("TODO, stubbed")
}

// Gluco2ToIntLit is the overload toInt(Lit) (inline).
// C: inline int toInt(Lit p)
func Gluco2ToIntLit(p Gluco2Lit) int {
	panic("TODO, stubbed")
}

// Gluco2ToLit is Gluco2::toLit (inline).
// C: inline Lit toLit(int i)
func Gluco2ToLit(i int) Gluco2Lit {
	panic("TODO, stubbed")
}

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
func NewGluco2LboolV(v uint8) Gluco2Lbool {
	panic("TODO, stubbed")
}

// NewGluco2Lbool is lbool(): value 0.
// C: lbool()
func NewGluco2Lbool() Gluco2Lbool {
	panic("TODO, stubbed")
}

// NewGluco2LboolB is the explicit constructor lbool(bool x): value !x.
// C: explicit lbool(bool x)
func NewGluco2LboolB(x bool) Gluco2Lbool {
	panic("TODO, stubbed")
}

// Eq is lbool::operator==.
// C: bool operator==(lbool b) const
func (l Gluco2Lbool) Eq(b Gluco2Lbool) bool {
	panic("TODO, stubbed")
}

// Ne is lbool::operator!=.
// C: bool operator!=(lbool b) const
func (l Gluco2Lbool) Ne(b Gluco2Lbool) bool {
	panic("TODO, stubbed")
}

// Xor is lbool::operator^.
// C: lbool operator^(bool b) const
func (l Gluco2Lbool) Xor(b bool) Gluco2Lbool {
	panic("TODO, stubbed")
}

// And is lbool::operator&&.
// C: lbool operator&&(lbool b) const
func (l Gluco2Lbool) And(b Gluco2Lbool) Gluco2Lbool {
	panic("TODO, stubbed")
}

// Or is lbool::operator||.
// C: lbool operator||(lbool b) const
func (l Gluco2Lbool) Or(b Gluco2Lbool) Gluco2Lbool {
	panic("TODO, stubbed")
}

// Gluco2ToIntLbool is the friend toInt(lbool) (inline).
// C: inline int toInt(lbool l)
func Gluco2ToIntLbool(l Gluco2Lbool) int {
	panic("TODO, stubbed")
}

// Gluco2ToLbool is the friend toLbool(int) (inline).
// C: inline lbool toLbool(int v)
func Gluco2ToLbool(v int) Gluco2Lbool {
	panic("TODO, stubbed")
}

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

// gluco2ClauseHeader is the anonymous bit-field struct `header` of Clause.
type gluco2ClauseHeader struct {
	mark               uint32 // 2 bits
	learnt             uint32 // 1 bit
	has_extra          uint32 // 1 bit
	reloced            uint32 // 1 bit
	lbd                uint32 // 26 bits
	canbedel           uint32 // 1 bit
	size               uint32 // 32 bits
	szWithoutSelectors uint32 // 32 bits
}

// Gluco2Clause is Gluco2::Clause. The C++ object is a header followed by a
// flexible array `union { Lit lit; float act; uint32_t abs; CRef rel; } data[0]`
// of 32-bit words living inside a ClauseAllocator region.
type Gluco2Clause struct {
	header gluco2ClauseHeader
	data   []uint32
}

// newGluco2Clause is the private template constructor Clause(const V& ps, bool use_extra, bool learnt).
// C: template<class V> Clause(const V& ps, bool use_extra, bool learnt)
func newGluco2Clause(ps Gluco2LitSeq, useExtra bool, learnt bool) *Gluco2Clause {
	panic("TODO, stubbed")
}

// CalcAbstraction is Clause::calcAbstraction.
// C: void calcAbstraction()
func (c *Gluco2Clause) CalcAbstraction() {
	panic("TODO, stubbed")
}

// Size is Clause::size.
// C: int size() const
func (c *Gluco2Clause) Size() int {
	panic("TODO, stubbed")
}

// Shrink is Clause::shrink.
// C: void shrink(int i)
func (c *Gluco2Clause) Shrink(i int) {
	panic("TODO, stubbed")
}

// Pop is Clause::pop.
// C: void pop()
func (c *Gluco2Clause) Pop() {
	panic("TODO, stubbed")
}

// Learnt is Clause::learnt.
// C: bool learnt() const
func (c *Gluco2Clause) Learnt() bool {
	panic("TODO, stubbed")
}

// HasExtra is Clause::has_extra.
// C: bool has_extra() const
func (c *Gluco2Clause) HasExtra() bool {
	panic("TODO, stubbed")
}

// Mark is the getter overload mark().
// C: uint32_t mark() const
func (c *Gluco2Clause) Mark() uint32 {
	panic("TODO, stubbed")
}

// SetMark is the setter overload mark(uint32_t m).
// C: void mark(uint32_t m)
func (c *Gluco2Clause) SetMark(m uint32) {
	panic("TODO, stubbed")
}

// Last is Clause::last.
// C: const Lit& last() const
func (c *Gluco2Clause) Last() Gluco2Lit {
	panic("TODO, stubbed")
}

// Reloced is Clause::reloced.
// C: bool reloced() const
func (c *Gluco2Clause) Reloced() bool {
	panic("TODO, stubbed")
}

// Relocation is Clause::relocation.
// C: CRef relocation() const
func (c *Gluco2Clause) Relocation() Gluco2CRef {
	panic("TODO, stubbed")
}

// Relocate is Clause::relocate.
// C: void relocate(CRef c)
func (c *Gluco2Clause) Relocate(cr Gluco2CRef) {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: Lit& operator[](int i)
func (c *Gluco2Clause) At(i int) *Gluco2Lit {
	panic("TODO, stubbed")
}

// Lits is `operator const Lit*() const`.
// C: operator const Lit*(void) const
func (c *Gluco2Clause) Lits() []Gluco2Lit {
	panic("TODO, stubbed")
}

// Activity is Clause::activity.
// C: float& activity()
func (c *Gluco2Clause) Activity() *float32 {
	panic("TODO, stubbed")
}

// Abstraction is Clause::abstraction.
// C: uint32_t abstraction() const
func (c *Gluco2Clause) Abstraction() uint32 {
	panic("TODO, stubbed")
}

// SetLBD is Clause::setLBD.
// C: void setLBD(int i)
func (c *Gluco2Clause) SetLBD(i int) {
	panic("TODO, stubbed")
}

// Lbd is Clause::lbd.
// C: unsigned int lbd() const
func (c *Gluco2Clause) Lbd() uint32 {
	panic("TODO, stubbed")
}

// SetCanBeDel is Clause::setCanBeDel.
// C: void setCanBeDel(bool b)
func (c *Gluco2Clause) SetCanBeDel(b bool) {
	panic("TODO, stubbed")
}

// CanBeDel is Clause::canBeDel.
// C: bool canBeDel()
func (c *Gluco2Clause) CanBeDel() bool {
	panic("TODO, stubbed")
}

// SetSizeWithoutSelectors is Clause::setSizeWithoutSelectors.
// C: void setSizeWithoutSelectors(unsigned int n)
func (c *Gluco2Clause) SetSizeWithoutSelectors(n uint32) {
	panic("TODO, stubbed")
}

// SizeWithoutSelectors is Clause::sizeWithoutSelectors.
// C: unsigned int sizeWithoutSelectors() const
func (c *Gluco2Clause) SizeWithoutSelectors() uint32 {
	panic("TODO, stubbed")
}

// Subsumes is Clause::subsumes (defined inline after the classes).
// C: inline Lit Clause::subsumes(const Clause& other) const
func (c *Gluco2Clause) Subsumes(other *Gluco2Clause) Gluco2Lit {
	panic("TODO, stubbed")
}

// Strengthen is Clause::strengthen (defined inline after the classes).
// C: inline void Clause::strengthen(Lit p)
func (c *Gluco2Clause) Strengthen(p Gluco2Lit) {
	panic("TODO, stubbed")
}

// Gluco2ClauseAllocator is Gluco2::ClauseAllocator, a RegionAllocator<uint32_t>.
type Gluco2ClauseAllocator struct {
	Gluco2RegionAllocator[uint32]
	ExtraClauseField bool // extra_clause_field
}

// clauseWord32Size is ClauseAllocator::clauseWord32Size (private static).
// C: static int clauseWord32Size(int size, bool has_extra)
func gluco2ClauseAllocatorClauseWord32Size(size int, hasExtra bool) int {
	panic("TODO, stubbed")
}

// NewGluco2ClauseAllocatorCap is ClauseAllocator(uint32_t start_cap).
// C: ClauseAllocator(uint32_t start_cap)
func NewGluco2ClauseAllocatorCap(startCap uint32) *Gluco2ClauseAllocator {
	panic("TODO, stubbed")
}

// NewGluco2ClauseAllocator is ClauseAllocator().
// C: ClauseAllocator()
func NewGluco2ClauseAllocator() *Gluco2ClauseAllocator {
	panic("TODO, stubbed")
}

// MoveTo is ClauseAllocator::moveTo.
// C: void moveTo(ClauseAllocator& to)
func (a *Gluco2ClauseAllocator) MoveTo(to *Gluco2ClauseAllocator) {
	panic("TODO, stubbed")
}

// Alloc is the template ClauseAllocator::alloc (learnt defaults to false).
// C: template<class Lits> CRef alloc(const Lits& ps, bool learnt = false)
func (a *Gluco2ClauseAllocator) Alloc(ps Gluco2LitSeq, learnt bool) Gluco2CRef {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: Clause& operator[](Ref r)
func (a *Gluco2ClauseAllocator) At(r Gluco2CRef) *Gluco2Clause {
	panic("TODO, stubbed")
}

// Lea is ClauseAllocator::lea.
// C: Clause* lea(Ref r)
func (a *Gluco2ClauseAllocator) Lea(r Gluco2CRef) *Gluco2Clause {
	panic("TODO, stubbed")
}

// Ael is ClauseAllocator::ael.
// C: Ref ael(const Clause* t)
func (a *Gluco2ClauseAllocator) Ael(t *Gluco2Clause) Gluco2CRef {
	panic("TODO, stubbed")
}

// Free_ is ClauseAllocator::free_ (renamed from _free in Minisat).
// C: void free_(CRef cid)
func (a *Gluco2ClauseAllocator) Free_(cid Gluco2CRef) {
	panic("TODO, stubbed")
}

// Reloc is ClauseAllocator::reloc.
// C: void reloc(CRef& cr, ClauseAllocator& to)
func (a *Gluco2ClauseAllocator) Reloc(cr *Gluco2CRef, to *Gluco2ClauseAllocator) {
	panic("TODO, stubbed")
}

// Gluco2OccIdx is what OccLists' index type may be (toInt is defined for Var and Lit).
type Gluco2OccIdx interface {
	int | Gluco2Lit
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
	panic("TODO, stubbed")
}

// Init is OccLists::init.
// C: void init(const Idx& idx)
func (o *Gluco2OccLists[Idx, E, D]) Init(idx Idx) {
	panic("TODO, stubbed")
}

// At is OccLists::operator[].
// C: Vec& operator[](const Idx& idx)
func (o *Gluco2OccLists[Idx, E, D]) At(idx Idx) *Gluco2Vec[E] {
	panic("TODO, stubbed")
}

// Lookup is OccLists::lookup.
// C: Vec& lookup(const Idx& idx)
func (o *Gluco2OccLists[Idx, E, D]) Lookup(idx Idx) *Gluco2Vec[E] {
	panic("TODO, stubbed")
}

// CleanAll is OccLists::cleanAll.
// C: void cleanAll()
func (o *Gluco2OccLists[Idx, E, D]) CleanAll() {
	panic("TODO, stubbed")
}

// Clean is OccLists::clean.
// C: void clean(const Idx& idx)
func (o *Gluco2OccLists[Idx, E, D]) Clean(idx Idx) {
	panic("TODO, stubbed")
}

// Smudge is OccLists::smudge.
// C: void smudge(const Idx& idx)
func (o *Gluco2OccLists[Idx, E, D]) Smudge(idx Idx) {
	panic("TODO, stubbed")
}

// Clear is OccLists::clear (free defaults to true).
// C: void clear(bool free = true)
func (o *Gluco2OccLists[Idx, E, D]) Clear(free bool) {
	panic("TODO, stubbed")
}

// Gluco2CRefHash is CMap<T>::CRefHash.
type Gluco2CRefHash struct{}

// Call is CRefHash::operator().
// C: uint32_t operator()(CRef cr) const
func (h Gluco2CRefHash) Call(cr Gluco2CRef) uint32 {
	panic("TODO, stubbed")
}

// Gluco2CMap is Gluco2::CMap<T>: a Map<CRef, T, CRefHash>.
type Gluco2CMap[T any] struct {
	map_ Gluco2Map[Gluco2CRef, T, Gluco2CRefHash, Gluco2Equal[Gluco2CRef]]
}

// Clear is CMap::clear.
// C: void clear()
func (m *Gluco2CMap[T]) Clear() {
	panic("TODO, stubbed")
}

// Size is CMap::size.
// C: int size() const
func (m *Gluco2CMap[T]) Size() int {
	panic("TODO, stubbed")
}

// Insert is CMap::insert.
// C: void insert(CRef cr, const T& t)
func (m *Gluco2CMap[T]) Insert(cr Gluco2CRef, t T) {
	panic("TODO, stubbed")
}

// GrowTo is CMap::growTo (for compatibility; same as insert).
// C: void growTo(CRef cr, const T& t)
func (m *Gluco2CMap[T]) GrowTo(cr Gluco2CRef, t T) {
	panic("TODO, stubbed")
}

// Remove is CMap::remove.
// C: void remove(CRef cr)
func (m *Gluco2CMap[T]) Remove(cr Gluco2CRef) {
	panic("TODO, stubbed")
}

// Has is CMap::has.
// C: bool has(CRef cr, T& t)
func (m *Gluco2CMap[T]) Has(cr Gluco2CRef, t *T) bool {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: T& operator[](CRef cr)
func (m *Gluco2CMap[T]) At(cr Gluco2CRef) *T {
	panic("TODO, stubbed")
}

// BucketCount is CMap::bucket_count.
// C: int bucket_count() const
func (m *Gluco2CMap[T]) BucketCount() int {
	panic("TODO, stubbed")
}

// Bucket is CMap::bucket.
// C: const vec<typename HashTable::Pair>& bucket(int i) const
func (m *Gluco2CMap[T]) Bucket(i int) *Gluco2Vec[Gluco2MapPair[Gluco2CRef, T]] {
	panic("TODO, stubbed")
}

// MoveTo is CMap::moveTo.
// C: void moveTo(CMap& other)
func (m *Gluco2CMap[T]) MoveTo(other *Gluco2CMap[T]) {
	panic("TODO, stubbed")
}

// Debug is CMap::debug.
// C: void debug()
func (m *Gluco2CMap[T]) Debug() {
	panic("TODO, stubbed")
}
