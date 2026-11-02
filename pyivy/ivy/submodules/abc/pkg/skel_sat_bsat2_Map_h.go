package abc

// Hand-checked skeleton of src/sat/bsat2/Map.h (namespace Minisat).

// MinisatHasher is the requirement on Map's template argument H: an object with
// uint32_t operator()(const K&).
type MinisatHasher[K any] interface {
	Call(k K) uint32
}

// MinisatEqualer is the requirement on Map's template argument E: an object with
// bool operator()(const K&, const K&).
type MinisatEqualer[K any] interface {
	Call(k1 K, k2 K) bool
}

// MinisatHash is Hash<K>; its operator() calls the hash() overload for K.
type MinisatHash[K any] struct{}

// Call is Hash::operator().
// C: uint32_t operator()(const K& k) const
func (h MinisatHash[K]) Call(k K) uint32 {
	panic("TODO, stubbed")
}

// MinisatEqual is Equal<K>.
type MinisatEqual[K comparable] struct{}

// Call is Equal::operator().
// C: bool operator()(const K& k1, const K& k2) const
func (e MinisatEqual[K]) Call(k1 K, k2 K) bool {
	panic("TODO, stubbed")
}

// MinisatDeepHash is DeepHash<K>; the key is a pointer.
type MinisatDeepHash[K any] struct{}

// Call is DeepHash::operator().
// C: uint32_t operator()(const K* k) const
func (h MinisatDeepHash[K]) Call(k *K) uint32 {
	panic("TODO, stubbed")
}

// MinisatDeepEqual is DeepEqual<K>; the keys are pointers.
type MinisatDeepEqual[K comparable] struct{}

// Call is DeepEqual::operator().
// C: bool operator()(const K* k1, const K* k2) const
func (e MinisatDeepEqual[K]) Call(k1 *K, k2 *K) bool {
	panic("TODO, stubbed")
}

// minisatHashU32 is the overload hash(uint32_t) (static inline).
// C: static inline uint32_t hash(uint32_t x)
func minisatHashU32(x uint32) uint32 {
	panic("TODO, stubbed")
}

// minisatHashU64 is the overload hash(uint64_t) (static inline).
// C: static inline uint32_t hash(uint64_t x)
func minisatHashU64(x uint64) uint32 {
	panic("TODO, stubbed")
}

// minisatHashI32 is the overload hash(int32_t) (static inline).
// C: static inline uint32_t hash(int32_t x)
func minisatHashI32(x int32) uint32 {
	panic("TODO, stubbed")
}

// minisatHashI64 is the overload hash(int64_t) (static inline).
// C: static inline uint32_t hash(int64_t x)
func minisatHashI64(x int64) uint32 {
	panic("TODO, stubbed")
}

// minisatNprimes is `static const int nprimes = 25`.
const minisatNprimes = 25

// minisatPrimes is `static const int primes[nprimes]`.
var minisatPrimes = [minisatNprimes]int{31, 73, 151, 313, 643, 1291, 2593, 5233, 10501, 21013, 42073, 84181, 168451, 337219, 674701, 1349473, 2699299, 5398891, 10798093, 21596719, 43193641, 86387383, 172775299, 345550609, 691101253}

// MinisatMapPair is Map<K,D,H,E>::Pair.
type MinisatMapPair[K any, D any] struct {
	key  K
	data D
}

// MinisatMap is Minisat::Map<K, D, H, E>: a chained hash map (default H = Hash<K>,
// E = Equal<K>). The private copy constructor and operator= only assert(0) and are
// not ported.
type MinisatMap[K any, D any, H MinisatHasher[K], E MinisatEqualer[K]] struct {
	hash   H
	equals E
	table  []MinisatVec[MinisatMapPair[K, D]] // vec<Pair>*
	cap    int
	size   int
}

// NewMinisatMap is Map().
// C: Map()
func NewMinisatMap[K any, D any, H MinisatHasher[K], E MinisatEqualer[K]]() *MinisatMap[K, D, H, E] {
	panic("TODO, stubbed")
}

// NewMinisatMapHE is Map(const H& h, const E& e).
// C: Map(const H& h, const E& e)
func NewMinisatMapHE[K any, D any, H MinisatHasher[K], E MinisatEqualer[K]](h H, e E) *MinisatMap[K, D, H, E] {
	panic("TODO, stubbed")
}

// Destroy is ~Map().
// C: ~Map()
func (m *MinisatMap[K, D, H, E]) Destroy() {
	panic("TODO, stubbed")
}

// checkCap is Map::checkCap (private).
// C: bool checkCap(int new_size) const
func (m *MinisatMap[K, D, H, E]) checkCap(newSize int) bool {
	panic("TODO, stubbed")
}

// index is Map::index (private).
// C: int32_t index(const K& k) const
func (m *MinisatMap[K, D, H, E]) index(k K) int32 {
	panic("TODO, stubbed")
}

// insert is Map::_insert (private).
// C: void _insert(const K& k, const D& d)
func (m *MinisatMap[K, D, H, E]) insert(k K, d D) {
	panic("TODO, stubbed")
}

// rehash is Map::rehash (private).
// C: void rehash()
func (m *MinisatMap[K, D, H, E]) rehash() {
	panic("TODO, stubbed")
}

// At is operator[] (const and non-const overloads).
// C: D& operator[](const K& k)
func (m *MinisatMap[K, D, H, E]) At(k K) *D {
	panic("TODO, stubbed")
}

// Insert is Map::insert.
// C: void insert(const K& k, const D& d)
func (m *MinisatMap[K, D, H, E]) Insert(k K, d D) {
	panic("TODO, stubbed")
}

// Peek is Map::peek.
// C: bool peek(const K& k, D& d) const
func (m *MinisatMap[K, D, H, E]) Peek(k K, d *D) bool {
	panic("TODO, stubbed")
}

// Has is Map::has.
// C: bool has(const K& k) const
func (m *MinisatMap[K, D, H, E]) Has(k K) bool {
	panic("TODO, stubbed")
}

// Remove is Map::remove.
// C: void remove(const K& k)
func (m *MinisatMap[K, D, H, E]) Remove(k K) {
	panic("TODO, stubbed")
}

// Clear is Map::clear.
// C: void clear()
func (m *MinisatMap[K, D, H, E]) Clear() {
	panic("TODO, stubbed")
}

// Elems is Map::elems.
// C: int elems() const
func (m *MinisatMap[K, D, H, E]) Elems() int {
	panic("TODO, stubbed")
}

// BucketCount is Map::bucket_count.
// C: int bucket_count() const
func (m *MinisatMap[K, D, H, E]) BucketCount() int {
	panic("TODO, stubbed")
}

// MoveTo is Map::moveTo.
// C: void moveTo(Map& other)
func (m *MinisatMap[K, D, H, E]) MoveTo(other *MinisatMap[K, D, H, E]) {
	panic("TODO, stubbed")
}

// Bucket is Map::bucket.
// C: const vec<Pair>& bucket(int i) const
func (m *MinisatMap[K, D, H, E]) Bucket(i int) *MinisatVec[MinisatMapPair[K, D]] {
	panic("TODO, stubbed")
}
