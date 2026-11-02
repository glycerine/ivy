package abc

// Port of src/sat/glucose2/Map.h (namespace Gluco2).

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

// Gluco2Hash is Hash<K>; its operator() calls the hash() overload for K
// (uint32, uint64, int32 and int64 keys exist in Map.h).
type Gluco2Hash[K any] struct{}

// Call is Hash::operator().
// C: uint32_t operator()(const K& k) const
func (h Gluco2Hash[K]) Call(k K) uint32 {
	switch v := any(k).(type) {
	case uint32:
		return gluco2HashU32(v)
	case uint64:
		return gluco2HashU64(v)
	case int32:
		return gluco2HashI32(v)
	case int64:
		return gluco2HashI64(v)
	case int:
		return gluco2HashI64(int64(v))
	}
	panic("Gluco2Hash: no hash() overload for this key type")
}

// Gluco2Equal is Equal<K>.
type Gluco2Equal[K comparable] struct{}

// Call is Equal::operator().
// C: bool operator()(const K& k1, const K& k2) const
func (e Gluco2Equal[K]) Call(k1 K, k2 K) bool { return k1 == k2 }

// Gluco2DeepHash is DeepHash<K>; the key is a pointer, the pointee is hashed.
type Gluco2DeepHash[K any] struct{}

// Call is DeepHash::operator().
// C: uint32_t operator()(const K* k) const
func (h Gluco2DeepHash[K]) Call(k *K) uint32 {
	return Gluco2Hash[K]{}.Call(*k)
}

// Gluco2DeepEqual is DeepEqual<K>; the keys are pointers, the pointees are compared.
type Gluco2DeepEqual[K comparable] struct{}

// Call is DeepEqual::operator().
// C: bool operator()(const K* k1, const K* k2) const
func (e Gluco2DeepEqual[K]) Call(k1 *K, k2 *K) bool { return *k1 == *k2 }

// gluco2HashU32 is the overload hash(uint32_t) (static inline).
// C: static inline uint32_t hash(uint32_t x)
func gluco2HashU32(x uint32) uint32 { return x }

// gluco2HashU64 is the overload hash(uint64_t) (static inline).
// C: static inline uint32_t hash(uint64_t x)
func gluco2HashU64(x uint64) uint32 { return uint32(x) }

// gluco2HashI32 is the overload hash(int32_t) (static inline).
// C: static inline uint32_t hash(int32_t x)
func gluco2HashI32(x int32) uint32 { return uint32(x) }

// gluco2HashI64 is the overload hash(int64_t) (static inline).
// C: static inline uint32_t hash(int64_t x)
func gluco2HashI64(x int64) uint32 { return uint32(x) }

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
	return &Gluco2Map[K, D, H, E]{}
}

// NewGluco2MapHE is Map(const H& h, const E& e).
// C: Map(const H& h, const E& e)
func NewGluco2MapHE[K any, D any, H Gluco2Hasher[K], E Gluco2Equaler[K]](h H, e E) *Gluco2Map[K, D, H, E] {
	return &Gluco2Map[K, D, H, E]{hash: h, equals: e}
}

// Destroy is ~Map().
// C: ~Map()
func (m *Gluco2Map[K, D, H, E]) Destroy() {
	m.table = nil
}

// checkCap is Map::checkCap (private).
// C: bool checkCap(int new_size) const
func (m *Gluco2Map[K, D, H, E]) checkCap(newSize int) bool { return newSize > m.cap }

// index is Map::index (private).
// C: int32_t index(const K& k) const
func (m *Gluco2Map[K, D, H, E]) index(k K) int32 {
	return int32(m.hash.Call(k) % uint32(m.cap))
}

// insert is Map::_insert (private).
// C: void _insert(const K& k, const D& d)
func (m *Gluco2Map[K, D, H, E]) insert(k K, d D) {
	ps := &m.table[m.index(k)]
	ps.Push()
	ps.Last().key = k
	ps.Last().data = d
}

// rehash is Map::rehash (private).
// C: void rehash()
func (m *Gluco2Map[K, D, H, E]) rehash() {
	old := m.table
	oldCap := m.cap
	newsize := gluco2Primes[0]
	for i := 1; newsize <= m.cap && i < gluco2Nprimes; i++ {
		newsize = gluco2Primes[i]
	}
	m.table = make([]Gluco2Vec[Gluco2MapPair[K, D]], newsize)
	m.cap = newsize
	for i := 0; i < oldCap; i++ {
		for j := 0; j < old[i].Size(); j++ {
			m.insert(old[i].At(j).key, old[i].At(j).data)
		}
	}
}

// At is operator[] (const and non-const overloads); the key must be present.
// C: D& operator[](const K& k)
func (m *Gluco2Map[K, D, H, E]) At(k K) *D {
	if m.size == 0 {
		panic("Gluco2Map.At: empty map")
	}
	var res *D
	ps := &m.table[m.index(k)]
	for i := 0; i < ps.Size(); i++ {
		if m.equals.Call(ps.At(i).key, k) {
			res = &ps.At(i).data
		}
	}
	if res == nil {
		panic("Gluco2Map.At: key not found")
	}
	return res
}

// Insert is Map::insert.
// C: void insert(const K& k, const D& d)
func (m *Gluco2Map[K, D, H, E]) Insert(k K, d D) {
	if m.checkCap(m.size + 1) {
		m.rehash()
	}
	m.insert(k, d)
	m.size++
}

// Peek is Map::peek.
// C: bool peek(const K& k, D& d) const
func (m *Gluco2Map[K, D, H, E]) Peek(k K, d *D) bool {
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
func (m *Gluco2Map[K, D, H, E]) Has(k K) bool {
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
func (m *Gluco2Map[K, D, H, E]) Remove(k K) {
	if m.table == nil {
		panic("Gluco2Map.Remove: empty map")
	}
	ps := &m.table[m.index(k)]
	j := 0
	for ; j < ps.Size() && !m.equals.Call(ps.At(j).key, k); j++ {
	}
	if j >= ps.Size() {
		panic("Gluco2Map.Remove: key not found")
	}
	*ps.At(j) = *ps.Last()
	ps.Pop()
	m.size--
}

// Clear is Map::clear.
// C: void clear()
func (m *Gluco2Map[K, D, H, E]) Clear() {
	m.cap = 0
	m.size = 0
	m.table = nil
}

// Elems is Map::elems.
// C: int elems() const
func (m *Gluco2Map[K, D, H, E]) Elems() int { return m.size }

// BucketCount is Map::bucket_count.
// C: int bucket_count() const
func (m *Gluco2Map[K, D, H, E]) BucketCount() int { return m.cap }

// MoveTo is Map::moveTo.
// C: void moveTo(Map& other)
func (m *Gluco2Map[K, D, H, E]) MoveTo(other *Gluco2Map[K, D, H, E]) {
	other.table = m.table
	other.cap = m.cap
	other.size = m.size
	m.table = nil
	m.size = 0
	m.cap = 0
}

// Bucket is Map::bucket.
// C: const vec<Pair>& bucket(int i) const
func (m *Gluco2Map[K, D, H, E]) Bucket(i int) *Gluco2Vec[Gluco2MapPair[K, D]] {
	return &m.table[i]
}
