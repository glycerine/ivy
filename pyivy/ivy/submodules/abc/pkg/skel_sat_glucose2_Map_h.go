package abc

// Hand-checked skeleton of src/sat/glucose2/Map.h (namespace Gluco2).

// Gluco2Hasher is the requirement on Map's template argument H: an object with
// uint32_t operator()(const K&).
type Gluco2Hasher[K any] interface {
	Call(k K) uint32
}

// Gluco2Equaler is the requirement on Map's template argument E: an object with
// bool operator()(const K&, const K&).
type Gluco2Equaler[K any] interface {
	Call(k1 K, k2 K) bool
}

// Gluco2Hash is Hash<K>; its operator() calls the hash() overload for K.
type Gluco2Hash[K any] struct{}

// Call is Hash::operator().
// C: uint32_t operator()(const K& k) const
func (h Gluco2Hash[K]) Call(k K) uint32 {
	panic("TODO, stubbed")
}

// Gluco2Equal is Equal<K>.
type Gluco2Equal[K comparable] struct{}

// Call is Equal::operator().
// C: bool operator()(const K& k1, const K& k2) const
func (e Gluco2Equal[K]) Call(k1 K, k2 K) bool {
	panic("TODO, stubbed")
}

// Gluco2DeepHash is DeepHash<K>; the key is a pointer.
type Gluco2DeepHash[K any] struct{}

// Call is DeepHash::operator().
// C: uint32_t operator()(const K* k) const
func (h Gluco2DeepHash[K]) Call(k *K) uint32 {
	panic("TODO, stubbed")
}

// Gluco2DeepEqual is DeepEqual<K>; the keys are pointers.
type Gluco2DeepEqual[K comparable] struct{}

// Call is DeepEqual::operator().
// C: bool operator()(const K* k1, const K* k2) const
func (e Gluco2DeepEqual[K]) Call(k1 *K, k2 *K) bool {
	panic("TODO, stubbed")
}

// gluco2HashU32 is the overload hash(uint32_t) (static inline).
// C: static inline uint32_t hash(uint32_t x)
func gluco2HashU32(x uint32) uint32 {
	panic("TODO, stubbed")
}

// gluco2HashU64 is the overload hash(uint64_t) (static inline).
// C: static inline uint32_t hash(uint64_t x)
func gluco2HashU64(x uint64) uint32 {
	panic("TODO, stubbed")
}

// gluco2HashI32 is the overload hash(int32_t) (static inline).
// C: static inline uint32_t hash(int32_t x)
func gluco2HashI32(x int32) uint32 {
	panic("TODO, stubbed")
}

// gluco2HashI64 is the overload hash(int64_t) (static inline).
// C: static inline uint32_t hash(int64_t x)
func gluco2HashI64(x int64) uint32 {
	panic("TODO, stubbed")
}

// gluco2Nprimes is `static const int nprimes = 25`.
const gluco2Nprimes = 25

// gluco2Primes is `static const int primes[nprimes]`.
var gluco2Primes = [gluco2Nprimes]int{31, 73, 151, 313, 643, 1291, 2593, 5233, 10501, 21013, 42073, 84181, 168451, 337219, 674701, 1349473, 2699299, 5398891, 10798093, 21596719, 43193641, 86387383, 172775299, 345550609, 691101253}

// Gluco2MapPair is Map<K,D,H,E>::Pair.
type Gluco2MapPair[K any, D any] struct {
	key  K
	data D
}

// Gluco2Map is Gluco2::Map<K, D, H, E>: a chained hash map (default H = Hash<K>,
// E = Equal<K>). The private copy constructor and operator= only assert(0) and are
// not ported.
type Gluco2Map[K any, D any, H Gluco2Hasher[K], E Gluco2Equaler[K]] struct {
	hash   H
	equals E
	table  []Gluco2Vec[Gluco2MapPair[K, D]] // vec<Pair>*
	cap    int
	size   int
}

// NewGluco2Map is Map().
// C: Map()
func NewGluco2Map[K any, D any, H Gluco2Hasher[K], E Gluco2Equaler[K]]() *Gluco2Map[K, D, H, E] {
	panic("TODO, stubbed")
}

// NewGluco2MapHE is Map(const H& h, const E& e).
// C: Map(const H& h, const E& e)
func NewGluco2MapHE[K any, D any, H Gluco2Hasher[K], E Gluco2Equaler[K]](h H, e E) *Gluco2Map[K, D, H, E] {
	panic("TODO, stubbed")
}

// Destroy is ~Map().
// C: ~Map()
func (m *Gluco2Map[K, D, H, E]) Destroy() {
	panic("TODO, stubbed")
}

// checkCap is Map::checkCap (private).
// C: bool checkCap(int new_size) const
func (m *Gluco2Map[K, D, H, E]) checkCap(newSize int) bool {
	panic("TODO, stubbed")
}

// index is Map::index (private).
// C: int32_t index(const K& k) const
func (m *Gluco2Map[K, D, H, E]) index(k K) int32 {
	panic("TODO, stubbed")
}

// insert is Map::_insert (private).
// C: void _insert(const K& k, const D& d)
func (m *Gluco2Map[K, D, H, E]) insert(k K, d D) {
	panic("TODO, stubbed")
}

// rehash is Map::rehash (private).
// C: void rehash()
func (m *Gluco2Map[K, D, H, E]) rehash() {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: D& operator[](const K& k)
func (m *Gluco2Map[K, D, H, E]) At(k K) *D {
	panic("TODO, stubbed")
}

// Insert is Map::insert.
// C: void insert(const K& k, const D& d)
func (m *Gluco2Map[K, D, H, E]) Insert(k K, d D) {
	panic("TODO, stubbed")
}

// Peek is Map::peek.
// C: bool peek(const K& k, D& d) const
func (m *Gluco2Map[K, D, H, E]) Peek(k K, d *D) bool {
	panic("TODO, stubbed")
}

// Has is Map::has.
// C: bool has(const K& k) const
func (m *Gluco2Map[K, D, H, E]) Has(k K) bool {
	panic("TODO, stubbed")
}

// Remove is Map::remove.
// C: void remove(const K& k)
func (m *Gluco2Map[K, D, H, E]) Remove(k K) {
	panic("TODO, stubbed")
}

// Clear is Map::clear.
// C: void clear()
func (m *Gluco2Map[K, D, H, E]) Clear() {
	panic("TODO, stubbed")
}

// Elems is Map::elems.
// C: int elems() const
func (m *Gluco2Map[K, D, H, E]) Elems() int {
	panic("TODO, stubbed")
}

// BucketCount is Map::bucket_count.
// C: int bucket_count() const
func (m *Gluco2Map[K, D, H, E]) BucketCount() int {
	panic("TODO, stubbed")
}

// MoveTo is Map::moveTo.
// C: void moveTo(Map& other)
func (m *Gluco2Map[K, D, H, E]) MoveTo(other *Gluco2Map[K, D, H, E]) {
	panic("TODO, stubbed")
}

// Bucket is Map::bucket.
// C: const vec<Pair>& bucket(int i) const
func (m *Gluco2Map[K, D, H, E]) Bucket(i int) *Gluco2Vec[Gluco2MapPair[K, D]] {
	panic("TODO, stubbed")
}
