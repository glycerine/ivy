package abc

// Port of src/sat/bsat2/Map.h (namespace Minisat).

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

// MinisatHash is Hash<K>; its operator() calls the hash() overload for K
// (uint32, uint64, int32 and int64 keys exist in Map.h).
type MinisatHash[K any] struct{}

// Call is Hash::operator().
// C: uint32_t operator()(const K& k) const
func (h MinisatHash[K]) Call(k K) uint32 {
	switch v := any(k).(type) {
	case uint32:
		return minisatHashU32(v)
	case uint64:
		return minisatHashU64(v)
	case int32:
		return minisatHashI32(v)
	case int64:
		return minisatHashI64(v)
	case int:
		return minisatHashI64(int64(v))
	}
	panic("MinisatHash: no hash() overload for this key type")
}

// MinisatEqual is Equal<K>.
type MinisatEqual[K comparable] struct{}

// Call is Equal::operator().
// C: bool operator()(const K& k1, const K& k2) const
func (e MinisatEqual[K]) Call(k1 K, k2 K) bool { return k1 == k2 }

// MinisatDeepHash is DeepHash<K>; the key is a pointer, the pointee is hashed.
type MinisatDeepHash[K any] struct{}

// Call is DeepHash::operator().
// C: uint32_t operator()(const K* k) const
func (h MinisatDeepHash[K]) Call(k *K) uint32 {
	return MinisatHash[K]{}.Call(*k)
}

// MinisatDeepEqual is DeepEqual<K>; the keys are pointers, the pointees are compared.
type MinisatDeepEqual[K comparable] struct{}

// Call is DeepEqual::operator().
// C: bool operator()(const K* k1, const K* k2) const
func (e MinisatDeepEqual[K]) Call(k1 *K, k2 *K) bool { return *k1 == *k2 }

// minisatHashU32 is the overload hash(uint32_t) (static inline).
// C: static inline uint32_t hash(uint32_t x)
func minisatHashU32(x uint32) uint32 { return x }

// minisatHashU64 is the overload hash(uint64_t) (static inline).
// C: static inline uint32_t hash(uint64_t x)
func minisatHashU64(x uint64) uint32 { return uint32(x) }

// minisatHashI32 is the overload hash(int32_t) (static inline).
// C: static inline uint32_t hash(int32_t x)
func minisatHashI32(x int32) uint32 { return uint32(x) }

// minisatHashI64 is the overload hash(int64_t) (static inline).
// C: static inline uint32_t hash(int64_t x)
func minisatHashI64(x int64) uint32 { return uint32(x) }

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
	return &MinisatMap[K, D, H, E]{}
}

// NewMinisatMapHE is Map(const H& h, const E& e).
// C: Map(const H& h, const E& e)
func NewMinisatMapHE[K any, D any, H MinisatHasher[K], E MinisatEqualer[K]](h H, e E) *MinisatMap[K, D, H, E] {
	return &MinisatMap[K, D, H, E]{hash: h, equals: e}
}

// Destroy is ~Map().
// C: ~Map()
func (m *MinisatMap[K, D, H, E]) Destroy() {
	m.table = nil
}

// checkCap is Map::checkCap (private).
// C: bool checkCap(int new_size) const
func (m *MinisatMap[K, D, H, E]) checkCap(newSize int) bool { return newSize > m.cap }

// index is Map::index (private).
// C: int32_t index(const K& k) const
func (m *MinisatMap[K, D, H, E]) index(k K) int32 {
	return int32(m.hash.Call(k) % uint32(m.cap))
}

// insert is Map::_insert (private).
// C: void _insert(const K& k, const D& d)
func (m *MinisatMap[K, D, H, E]) insert(k K, d D) {
	ps := &m.table[m.index(k)]
	ps.Push()
	ps.Last().key = k
	ps.Last().data = d
}

// rehash is Map::rehash (private).
// C: void rehash()
func (m *MinisatMap[K, D, H, E]) rehash() {
	old := m.table
	oldCap := m.cap
	newsize := minisatPrimes[0]
	for i := 1; newsize <= m.cap && i < minisatNprimes; i++ {
		newsize = minisatPrimes[i]
	}
	m.table = make([]MinisatVec[MinisatMapPair[K, D]], newsize)
	m.cap = newsize
	for i := 0; i < oldCap; i++ {
		for j := 0; j < old[i].Size(); j++ {
			m.insert(old[i].At(j).key, old[i].At(j).data)
		}
	}
}

// At is operator[] (const and non-const overloads); the key must be present.
// C: D& operator[](const K& k)
func (m *MinisatMap[K, D, H, E]) At(k K) *D {
	if m.size == 0 {
		panic("MinisatMap.At: empty map")
	}
	var res *D
	ps := &m.table[m.index(k)]
	for i := 0; i < ps.Size(); i++ {
		if m.equals.Call(ps.At(i).key, k) {
			res = &ps.At(i).data
		}
	}
	if res == nil {
		panic("MinisatMap.At: key not found")
	}
	return res
}

// Insert is Map::insert.
// C: void insert(const K& k, const D& d)
func (m *MinisatMap[K, D, H, E]) Insert(k K, d D) {
	if m.checkCap(m.size + 1) {
		m.rehash()
	}
	m.insert(k, d)
	m.size++
}

// Peek is Map::peek.
// C: bool peek(const K& k, D& d) const
func (m *MinisatMap[K, D, H, E]) Peek(k K, d *D) bool {
	if m.size == 0 {
		return false
	}
	ps := &m.table[m.index(k)]
	for i := 0; i < ps.Size(); i++ {
		if m.equals.Call(ps.At(i).key, k) {
			*d = ps.At(i).data
			return true
		}
	}
	return false
}

// Has is Map::has.
// C: bool has(const K& k) const
func (m *MinisatMap[K, D, H, E]) Has(k K) bool {
	if m.size == 0 {
		return false
	}
	ps := &m.table[m.index(k)]
	for i := 0; i < ps.Size(); i++ {
		if m.equals.Call(ps.At(i).key, k) {
			return true
		}
	}
	return false
}

// Remove is Map::remove.
// C: void remove(const K& k)
func (m *MinisatMap[K, D, H, E]) Remove(k K) {
	if m.table == nil {
		panic("MinisatMap.Remove: empty map")
	}
	ps := &m.table[m.index(k)]
	j := 0
	for ; j < ps.Size() && !m.equals.Call(ps.At(j).key, k); j++ {
	}
	if j >= ps.Size() {
		panic("MinisatMap.Remove: key not found")
	}
	*ps.At(j) = *ps.Last()
	ps.Pop()
	m.size--
}

// Clear is Map::clear.
// C: void clear()
func (m *MinisatMap[K, D, H, E]) Clear() {
	m.cap = 0
	m.size = 0
	m.table = nil
}

// Elems is Map::elems.
// C: int elems() const
func (m *MinisatMap[K, D, H, E]) Elems() int { return m.size }

// BucketCount is Map::bucket_count.
// C: int bucket_count() const
func (m *MinisatMap[K, D, H, E]) BucketCount() int { return m.cap }

// MoveTo is Map::moveTo.
// C: void moveTo(Map& other)
func (m *MinisatMap[K, D, H, E]) MoveTo(other *MinisatMap[K, D, H, E]) {
	other.table = m.table
	other.cap = m.cap
	other.size = m.size
	m.table = nil
	m.size = 0
	m.cap = 0
}

// Bucket is Map::bucket.
// C: const vec<Pair>& bucket(int i) const
func (m *MinisatMap[K, D, H, E]) Bucket(i int) *MinisatVec[MinisatMapPair[K, D]] {
	return &m.table[i]
}
