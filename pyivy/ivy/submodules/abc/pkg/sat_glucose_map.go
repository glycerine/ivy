package abc

// Port of src/sat/glucose/Map.h (namespace Gluco).

// GlucoHasher is the requirement on Map's template argument H: an object with
// uint32_t operator()(const K&).
type GlucoHasher[K any] interface {
	Call(k K) uint32
}

// GlucoEqualer is the requirement on Map's template argument E: an object with
// bool operator()(const K&, const K&).
type GlucoEqualer[K any] interface {
	Call(k1 K, k2 K) bool
}

// GlucoHash is Hash<K>; its operator() calls the hash() overload for K
// (uint32, uint64, int32 and int64 keys exist in Map.h).
type GlucoHash[K any] struct{}

// Call is Hash::operator().
// C: uint32_t operator()(const K& k) const
func (h GlucoHash[K]) Call(k K) uint32 {
	switch v := any(k).(type) {
	case uint32:
		return glucoHashU32(v)
	case uint64:
		return glucoHashU64(v)
	case int32:
		return glucoHashI32(v)
	case int64:
		return glucoHashI64(v)
	case int:
		return glucoHashI64(int64(v))
	}
	panic("GlucoHash: no hash() overload for this key type")
}

// GlucoEqual is Equal<K>.
type GlucoEqual[K comparable] struct{}

// Call is Equal::operator().
// C: bool operator()(const K& k1, const K& k2) const
func (e GlucoEqual[K]) Call(k1 K, k2 K) bool { return k1 == k2 }

// GlucoDeepHash is DeepHash<K>; the key is a pointer, the pointee is hashed.
type GlucoDeepHash[K any] struct{}

// Call is DeepHash::operator().
// C: uint32_t operator()(const K* k) const
func (h GlucoDeepHash[K]) Call(k *K) uint32 {
	return GlucoHash[K]{}.Call(*k)
}

// GlucoDeepEqual is DeepEqual<K>; the keys are pointers, the pointees are compared.
type GlucoDeepEqual[K comparable] struct{}

// Call is DeepEqual::operator().
// C: bool operator()(const K* k1, const K* k2) const
func (e GlucoDeepEqual[K]) Call(k1 *K, k2 *K) bool { return *k1 == *k2 }

// glucoHashU32 is the overload hash(uint32_t) (static inline).
// C: static inline uint32_t hash(uint32_t x)
func glucoHashU32(x uint32) uint32 { return x }

// glucoHashU64 is the overload hash(uint64_t) (static inline).
// C: static inline uint32_t hash(uint64_t x)
func glucoHashU64(x uint64) uint32 { return uint32(x) }

// glucoHashI32 is the overload hash(int32_t) (static inline).
// C: static inline uint32_t hash(int32_t x)
func glucoHashI32(x int32) uint32 { return uint32(x) }

// glucoHashI64 is the overload hash(int64_t) (static inline).
// C: static inline uint32_t hash(int64_t x)
func glucoHashI64(x int64) uint32 { return uint32(x) }

// glucoNprimes is `static const int nprimes = 25`.
const glucoNprimes = 25

// glucoPrimes is `static const int primes[nprimes]`.
var glucoPrimes = [glucoNprimes]int{31, 73, 151, 313, 643, 1291, 2593, 5233, 10501, 21013, 42073, 84181, 168451, 337219, 674701, 1349473, 2699299, 5398891, 10798093, 21596719, 43193641, 86387383, 172775299, 345550609, 691101253}

// GlucoMapPair is Map<K,D,H,E>::Pair.
type GlucoMapPair[K any, D any] struct {
	key  K
	data D
}

// GlucoMap is Gluco::Map<K, D, H, E>: a chained hash map (default H = Hash<K>,
// E = Equal<K>). The private copy constructor and operator= only assert(0) and are
// not ported.
type GlucoMap[K any, D any, H GlucoHasher[K], E GlucoEqualer[K]] struct {
	hash   H
	equals E
	table  []GlucoVec[GlucoMapPair[K, D]] // vec<Pair>*
	cap    int
	size   int
}

// NewGlucoMap is Map().
// C: Map()
func NewGlucoMap[K any, D any, H GlucoHasher[K], E GlucoEqualer[K]]() *GlucoMap[K, D, H, E] {
	return &GlucoMap[K, D, H, E]{}
}

// NewGlucoMapHE is Map(const H& h, const E& e).
// C: Map(const H& h, const E& e)
func NewGlucoMapHE[K any, D any, H GlucoHasher[K], E GlucoEqualer[K]](h H, e E) *GlucoMap[K, D, H, E] {
	return &GlucoMap[K, D, H, E]{hash: h, equals: e}
}

// Destroy is ~Map().
// C: ~Map()
func (m *GlucoMap[K, D, H, E]) Destroy() {
	m.table = nil
}

// checkCap is Map::checkCap (private).
// C: bool checkCap(int new_size) const
func (m *GlucoMap[K, D, H, E]) checkCap(newSize int) bool { return newSize > m.cap }

// index is Map::index (private).
// C: int32_t index(const K& k) const
func (m *GlucoMap[K, D, H, E]) index(k K) int32 {
	return int32(m.hash.Call(k) % uint32(m.cap))
}

// insert is Map::_insert (private).
// C: void _insert(const K& k, const D& d)
func (m *GlucoMap[K, D, H, E]) insert(k K, d D) {
	ps := &m.table[m.index(k)]
	ps.Push()
	ps.Last().key = k
	ps.Last().data = d
}

// rehash is Map::rehash (private).
// C: void rehash()
func (m *GlucoMap[K, D, H, E]) rehash() {
	old := m.table
	oldCap := m.cap
	newsize := glucoPrimes[0]
	for i := 1; newsize <= m.cap && i < glucoNprimes; i++ {
		newsize = glucoPrimes[i]
	}
	m.table = make([]GlucoVec[GlucoMapPair[K, D]], newsize)
	m.cap = newsize
	for i := 0; i < oldCap; i++ {
		for j := 0; j < old[i].Size(); j++ {
			m.insert(old[i].At(j).key, old[i].At(j).data)
		}
	}
}

// At is operator[] (const and non-const overloads); the key must be present.
// C: D& operator[](const K& k)
func (m *GlucoMap[K, D, H, E]) At(k K) *D {
	if m.size == 0 {
		panic("GlucoMap.At: empty map")
	}
	var res *D
	ps := &m.table[m.index(k)]
	for i := 0; i < ps.Size(); i++ {
		if m.equals.Call(ps.At(i).key, k) {
			res = &ps.At(i).data
		}
	}
	if res == nil {
		panic("GlucoMap.At: key not found")
	}
	return res
}

// Insert is Map::insert.
// C: void insert(const K& k, const D& d)
func (m *GlucoMap[K, D, H, E]) Insert(k K, d D) {
	if m.checkCap(m.size + 1) {
		m.rehash()
	}
	m.insert(k, d)
	m.size++
}

// Peek is Map::peek.
// C: bool peek(const K& k, D& d) const
func (m *GlucoMap[K, D, H, E]) Peek(k K, d *D) bool {
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
func (m *GlucoMap[K, D, H, E]) Has(k K) bool {
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
func (m *GlucoMap[K, D, H, E]) Remove(k K) {
	if m.table == nil {
		panic("GlucoMap.Remove: empty map")
	}
	ps := &m.table[m.index(k)]
	j := 0
	for ; j < ps.Size() && !m.equals.Call(ps.At(j).key, k); j++ {
	}
	if j >= ps.Size() {
		panic("GlucoMap.Remove: key not found")
	}
	*ps.At(j) = *ps.Last()
	ps.Pop()
	m.size--
}

// Clear is Map::clear.
// C: void clear()
func (m *GlucoMap[K, D, H, E]) Clear() {
	m.cap = 0
	m.size = 0
	m.table = nil
}

// Elems is Map::elems.
// C: int elems() const
func (m *GlucoMap[K, D, H, E]) Elems() int { return m.size }

// BucketCount is Map::bucket_count.
// C: int bucket_count() const
func (m *GlucoMap[K, D, H, E]) BucketCount() int { return m.cap }

// MoveTo is Map::moveTo.
// C: void moveTo(Map& other)
func (m *GlucoMap[K, D, H, E]) MoveTo(other *GlucoMap[K, D, H, E]) {
	other.table = m.table
	other.cap = m.cap
	other.size = m.size
	m.table = nil
	m.size = 0
	m.cap = 0
}

// Bucket is Map::bucket.
// C: const vec<Pair>& bucket(int i) const
func (m *GlucoMap[K, D, H, E]) Bucket(i int) *GlucoVec[GlucoMapPair[K, D]] {
	return &m.table[i]
}
