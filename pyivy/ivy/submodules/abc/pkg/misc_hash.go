package abc

// Concrete deterministic hash maps from src/misc/hash.

type HashMapLimits struct {
	MaxEntries int
}

func normalizeHashMapLimits(limits HashMapLimits) (HashMapLimits, error) {
	if limits.MaxEntries < 0 {
		return limits, ErrDimension
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 1 << 24
	}
	return limits, nil
}

// HashDefaultBin preserves Hash_DefaultHashFunc's 32-bit overflow arithmetic.
func HashDefaultBin(key, bins int) (int, error) {
	if bins <= 0 || bins > 1<<31-1 {
		return 0, ErrDimension
	}
	value := int32((uint32(key)+11)*uint32(key)*7 + 3)
	remainder := value % int32(bins)
	if remainder < 0 {
		remainder = -remainder
	}
	return int(remainder), nil
}

func HashStringBin(key string, bins int) (int, error) {
	if bins <= 0 {
		return 0, ErrDimension
	}
	var hash uint32
	for index := 0; index < len(key); index++ {
		hash = hash*6 + uint32(key[index])
	}
	return int(hash % uint32(bins)), nil
}

type hashIntNode struct {
	key, data int
	next      *hashIntNode
	alive     bool
}

type HashIntEntry struct{ Key, Data int }

type HashIntMap struct {
	bins   []*hashIntNode
	size   int
	limits HashMapLimits
	closed bool
}

func NewHashIntMap(bins int, limits HashMapLimits) (*HashIntMap, error) {
	limits, err := normalizeHashMapLimits(limits)
	if err != nil || bins <= 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	return &HashIntMap{bins: make([]*hashIntNode, bins), limits: limits}, nil
}

func (m *HashIntMap) check() error {
	if m == nil {
		return ErrDimension
	}
	if m.closed || len(m.bins) == 0 || m.size < 0 || m.size > m.limits.MaxEntries {
		return ErrInvariant
	}
	return nil
}

func (m *HashIntMap) find(key int) (**hashIntNode, *hashIntNode, error) {
	if err := m.check(); err != nil {
		return nil, nil, err
	}
	bin, _ := HashDefaultBin(key, len(m.bins))
	link := &m.bins[bin]
	for *link != nil {
		if (*link).key == key {
			return link, *link, nil
		}
		link = &(*link).next
	}
	return link, nil, nil
}

func (m *HashIntMap) getOrCreate(key int, create bool) (*hashIntNode, bool, error) {
	link, node, err := m.find(key)
	if err != nil || node != nil || !create {
		return node, false, err
	}
	if m.size == m.limits.MaxEntries {
		return nil, false, ErrLimit
	}
	node = &hashIntNode{key: key, alive: true}
	*link = node
	m.size++
	return node, true, nil
}

func (m *HashIntMap) Exists(key int) (bool, error) {
	_, node, err := m.find(key)
	return node != nil, err
}

func (m *HashIntMap) Write(key, data int) error {
	node, _, err := m.getOrCreate(key, true)
	if err != nil {
		return err
	}
	node.data = data
	return nil
}

func (m *HashIntMap) Entry(key int, create bool) (int, bool, error) {
	node, created, err := m.getOrCreate(key, create)
	if err != nil || node == nil {
		return 0, false, err
	}
	return node.data, created, nil
}

type HashIntCell struct {
	owner *HashIntMap
	node  *hashIntNode
}

func (m *HashIntMap) EntryCell(key int) (HashIntCell, error) {
	node, _, err := m.getOrCreate(key, true)
	if err != nil {
		return HashIntCell{}, err
	}
	return HashIntCell{m, node}, nil
}

func (cell HashIntCell) Get() (int, error) {
	if cell.owner == nil || cell.node == nil || !cell.node.alive {
		return 0, ErrInvariant
	}
	if err := cell.owner.check(); err != nil {
		return 0, err
	}
	return cell.node.data, nil
}

func (cell HashIntCell) Set(value int) error {
	if cell.owner == nil || cell.node == nil || !cell.node.alive {
		return ErrInvariant
	}
	if err := cell.owner.check(); err != nil {
		return err
	}
	cell.node.data = value
	return nil
}

func (m *HashIntMap) Size() (int, error) {
	if err := m.check(); err != nil {
		return 0, err
	}
	return m.size, nil
}

func (m *HashIntMap) Entries() ([]HashIntEntry, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	result := make([]HashIntEntry, 0, m.size)
	for _, head := range m.bins {
		for node := head; node != nil; node = node.next {
			result = append(result, HashIntEntry{node.key, node.data})
		}
	}
	return result, nil
}

func (m *HashIntMap) Close() error {
	if err := m.check(); err != nil {
		return err
	}
	for bin := range m.bins {
		for node := m.bins[bin]; node != nil; node = node.next {
			node.alive = false
		}
		m.bins[bin] = nil
	}
	m.size, m.closed = 0, true
	return nil
}

type hashFloatNode struct {
	key   int
	data  float32
	next  *hashFloatNode
	alive bool
}

type HashFloatEntry struct {
	Key  int
	Data float32
}

type HashFloatMap struct {
	bins   []*hashFloatNode
	size   int
	limits HashMapLimits
	closed bool
}

func NewHashFloatMap(bins int, limits HashMapLimits) (*HashFloatMap, error) {
	limits, err := normalizeHashMapLimits(limits)
	if err != nil || bins <= 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	return &HashFloatMap{bins: make([]*hashFloatNode, bins), limits: limits}, nil
}

func (m *HashFloatMap) check() error {
	if m == nil {
		return ErrDimension
	}
	if m.closed || len(m.bins) == 0 || m.size < 0 || m.size > m.limits.MaxEntries {
		return ErrInvariant
	}
	return nil
}

func (m *HashFloatMap) find(key int) (**hashFloatNode, *hashFloatNode, error) {
	if err := m.check(); err != nil {
		return nil, nil, err
	}
	bin, _ := HashDefaultBin(key, len(m.bins))
	link := &m.bins[bin]
	for *link != nil {
		if (*link).key == key {
			return link, *link, nil
		}
		link = &(*link).next
	}
	return link, nil, nil
}

func (m *HashFloatMap) getOrCreate(key int, create bool) (*hashFloatNode, bool, error) {
	link, node, err := m.find(key)
	if err != nil || node != nil || !create {
		return node, false, err
	}
	if m.size == m.limits.MaxEntries {
		return nil, false, ErrLimit
	}
	node = &hashFloatNode{key: key, alive: true}
	*link = node
	m.size++
	return node, true, nil
}

func (m *HashFloatMap) Exists(key int) (bool, error) {
	_, node, err := m.find(key)
	return node != nil, err
}
func (m *HashFloatMap) Write(key int, data float32) error {
	node, _, err := m.getOrCreate(key, true)
	if err == nil {
		node.data = data
	}
	return err
}
func (m *HashFloatMap) Entry(key int, create bool) (float32, bool, error) {
	node, created, err := m.getOrCreate(key, create)
	if err != nil || node == nil {
		return 0, false, err
	}
	return node.data, created, nil
}

type HashFloatCell struct {
	owner *HashFloatMap
	node  *hashFloatNode
}

func (m *HashFloatMap) EntryCell(key int) (HashFloatCell, error) {
	node, _, err := m.getOrCreate(key, true)
	return HashFloatCell{m, node}, err
}
func (cell HashFloatCell) Get() (float32, error) {
	if cell.owner == nil || cell.node == nil || !cell.node.alive {
		return 0, ErrInvariant
	}
	if err := cell.owner.check(); err != nil {
		return 0, err
	}
	return cell.node.data, nil
}
func (cell HashFloatCell) Set(value float32) error {
	if cell.owner == nil || cell.node == nil || !cell.node.alive {
		return ErrInvariant
	}
	if err := cell.owner.check(); err != nil {
		return err
	}
	cell.node.data = value
	return nil
}

func (m *HashFloatMap) Remove(key int) (bool, error) {
	link, node, err := m.find(key)
	if err != nil || node == nil {
		return false, err
	}
	*link = node.next
	node.next = nil
	node.alive = false
	m.size--
	return true, nil
}
func (m *HashFloatMap) Size() (int, error) {
	if err := m.check(); err != nil {
		return 0, err
	}
	return m.size, nil
}
func (m *HashFloatMap) Entries() ([]HashFloatEntry, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	result := make([]HashFloatEntry, 0, m.size)
	for _, head := range m.bins {
		for node := head; node != nil; node = node.next {
			result = append(result, HashFloatEntry{node.key, node.data})
		}
	}
	return result, nil
}
func (m *HashFloatMap) Close() error {
	if err := m.check(); err != nil {
		return err
	}
	for bin := range m.bins {
		for node := m.bins[bin]; node != nil; node = node.next {
			node.alive = false
		}
		m.bins[bin] = nil
	}
	m.size, m.closed = 0, true
	return nil
}

type hashPointerNode struct {
	key   int
	data  uintptr
	next  *hashPointerNode
	alive bool
}
type HashPointerEntry struct {
	Key  int
	Data uintptr
}
type HashPointerMap struct {
	bins   []*hashPointerNode
	size   int
	limits HashMapLimits
	closed bool
}

func NewHashPointerMap(bins int, limits HashMapLimits) (*HashPointerMap, error) {
	limits, err := normalizeHashMapLimits(limits)
	if err != nil || bins <= 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	return &HashPointerMap{bins: make([]*hashPointerNode, bins), limits: limits}, nil
}
func (m *HashPointerMap) check() error {
	if m == nil {
		return ErrDimension
	}
	if m.closed || len(m.bins) == 0 || m.size < 0 || m.size > m.limits.MaxEntries {
		return ErrInvariant
	}
	return nil
}
func (m *HashPointerMap) find(key int) (**hashPointerNode, *hashPointerNode, error) {
	if err := m.check(); err != nil {
		return nil, nil, err
	}
	bin, _ := HashDefaultBin(key, len(m.bins))
	link := &m.bins[bin]
	for *link != nil {
		if (*link).key == key {
			return link, *link, nil
		}
		link = &(*link).next
	}
	return link, nil, nil
}
func (m *HashPointerMap) getOrCreate(key int, create bool) (*hashPointerNode, bool, error) {
	link, node, err := m.find(key)
	if err != nil || node != nil || !create {
		return node, false, err
	}
	if m.size == m.limits.MaxEntries {
		return nil, false, ErrLimit
	}
	node = &hashPointerNode{key: key, alive: true}
	*link = node
	m.size++
	return node, true, nil
}
func (m *HashPointerMap) Exists(key int) (bool, error) {
	_, node, err := m.find(key)
	return node != nil, err
}
func (m *HashPointerMap) Write(key int, data uintptr) error {
	node, _, err := m.getOrCreate(key, true)
	if err == nil {
		node.data = data
	}
	return err
}
func (m *HashPointerMap) Entry(key int, create bool) (uintptr, bool, error) {
	node, created, err := m.getOrCreate(key, create)
	if err != nil || node == nil {
		return 0, false, err
	}
	return node.data, created, nil
}

type HashPointerCell struct {
	owner *HashPointerMap
	node  *hashPointerNode
}

func (m *HashPointerMap) EntryCell(key int) (HashPointerCell, error) {
	node, _, err := m.getOrCreate(key, true)
	return HashPointerCell{m, node}, err
}
func (cell HashPointerCell) Get() (uintptr, error) {
	if cell.owner == nil || cell.node == nil || !cell.node.alive {
		return 0, ErrInvariant
	}
	if err := cell.owner.check(); err != nil {
		return 0, err
	}
	return cell.node.data, nil
}
func (cell HashPointerCell) Set(value uintptr) error {
	if cell.owner == nil || cell.node == nil || !cell.node.alive {
		return ErrInvariant
	}
	if err := cell.owner.check(); err != nil {
		return err
	}
	cell.node.data = value
	return nil
}
func (m *HashPointerMap) Remove(key int) (uintptr, bool, error) {
	link, node, err := m.find(key)
	if err != nil || node == nil {
		return 0, false, err
	}
	*link = node.next
	node.next = nil
	node.alive = false
	m.size--
	return node.data, true, nil
}
func (m *HashPointerMap) Size() (int, error) {
	if err := m.check(); err != nil {
		return 0, err
	}
	return m.size, nil
}
func (m *HashPointerMap) Entries() ([]HashPointerEntry, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	result := make([]HashPointerEntry, 0, m.size)
	for _, head := range m.bins {
		for node := head; node != nil; node = node.next {
			result = append(result, HashPointerEntry{node.key, node.data})
		}
	}
	return result, nil
}
func (m *HashPointerMap) Close() error {
	if err := m.check(); err != nil {
		return err
	}
	for bin := range m.bins {
		for node := m.bins[bin]; node != nil; node = node.next {
			node.alive = false
		}
		m.bins[bin] = nil
	}
	m.size, m.closed = 0, true
	return nil
}

type hashStringNode struct {
	key   string
	data  uintptr
	next  *hashStringNode
	alive bool
}
type HashStringEntry struct {
	Key  string
	Data uintptr
}
type HashStringMap struct {
	bins   []*hashStringNode
	size   int
	limits HashMapLimits
	closed bool
}

func NewHashStringMap(bins int, limits HashMapLimits) (*HashStringMap, error) {
	limits, err := normalizeHashMapLimits(limits)
	if err != nil || bins <= 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	return &HashStringMap{bins: make([]*hashStringNode, bins), limits: limits}, nil
}
func (m *HashStringMap) check() error {
	if m == nil {
		return ErrDimension
	}
	if m.closed || len(m.bins) == 0 || m.size < 0 || m.size > m.limits.MaxEntries {
		return ErrInvariant
	}
	return nil
}
func (m *HashStringMap) find(key string) (**hashStringNode, *hashStringNode, error) {
	if err := m.check(); err != nil {
		return nil, nil, err
	}
	bin, _ := HashStringBin(key, len(m.bins))
	link := &m.bins[bin]
	for *link != nil {
		if (*link).key == key {
			return link, *link, nil
		}
		link = &(*link).next
	}
	return link, nil, nil
}
func (m *HashStringMap) getOrCreate(key string, create bool) (*hashStringNode, bool, error) {
	link, node, err := m.find(key)
	if err != nil || node != nil || !create {
		return node, false, err
	}
	if m.size == m.limits.MaxEntries {
		return nil, false, ErrLimit
	}
	node = &hashStringNode{key: string([]byte(key)), alive: true}
	*link = node
	m.size++
	return node, true, nil
}
func (m *HashStringMap) Exists(key string) (bool, error) {
	_, node, err := m.find(key)
	return node != nil, err
}
func (m *HashStringMap) Write(key string, data uintptr) error {
	node, _, err := m.getOrCreate(key, true)
	if err == nil {
		node.data = data
	}
	return err
}
func (m *HashStringMap) Entry(key string, create bool) (HashStringEntry, bool, bool, error) {
	node, created, err := m.getOrCreate(key, create)
	if err != nil || node == nil {
		return HashStringEntry{}, false, false, err
	}
	return HashStringEntry{node.key, node.data}, true, created, nil
}
func (m *HashStringMap) Remove(key string) (uintptr, bool, error) {
	link, node, err := m.find(key)
	if err != nil || node == nil {
		return 0, false, err
	}
	*link = node.next
	node.next = nil
	node.alive = false
	m.size--
	return node.data, true, nil
}
func (m *HashStringMap) Size() (int, error) {
	if err := m.check(); err != nil {
		return 0, err
	}
	return m.size, nil
}
func (m *HashStringMap) Entries() ([]HashStringEntry, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	result := make([]HashStringEntry, 0, m.size)
	for _, head := range m.bins {
		for node := head; node != nil; node = node.next {
			result = append(result, HashStringEntry{node.key, node.data})
		}
	}
	return result, nil
}
func (m *HashStringMap) Close() error {
	if err := m.check(); err != nil {
		return err
	}
	for bin := range m.bins {
		for node := m.bins[bin]; node != nil; node = node.next {
			node.alive = false
		}
		m.bins[bin] = nil
	}
	m.size, m.closed = 0, true
	return nil
}

type hashLongNode struct {
	key, data uintptr
	next      *hashLongNode
	alive     bool
}
type HashLongEntry struct{ Key, Data uintptr }
type HashLongMap struct {
	bins   []*hashLongNode
	size   int
	limits HashMapLimits
	closed bool
}

func NewHashLongMap(bins int, limits HashMapLimits) (*HashLongMap, error) {
	limits, err := normalizeHashMapLimits(limits)
	if err != nil || bins <= 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	return &HashLongMap{bins: make([]*hashLongNode, bins), limits: limits}, nil
}
func (m *HashLongMap) check() error {
	if m == nil {
		return ErrDimension
	}
	if m.closed || len(m.bins) == 0 || m.size < 0 || m.size > m.limits.MaxEntries {
		return ErrInvariant
	}
	return nil
}
func (m *HashLongMap) find(key uintptr) (**hashLongNode, *hashLongNode, error) {
	if err := m.check(); err != nil {
		return nil, nil, err
	}
	link := &m.bins[key%uintptr(len(m.bins))]
	for *link != nil {
		if (*link).key == key {
			return link, *link, nil
		}
		link = &(*link).next
	}
	return link, nil, nil
}
func (m *HashLongMap) getOrCreate(key uintptr, create bool) (*hashLongNode, bool, error) {
	link, node, err := m.find(key)
	if err != nil || node != nil || !create {
		return node, false, err
	}
	if m.size == m.limits.MaxEntries {
		return nil, false, ErrLimit
	}
	node = &hashLongNode{key: key, alive: true}
	*link = node
	m.size++
	return node, true, nil
}
func (m *HashLongMap) Exists(key uintptr) (bool, error) {
	_, node, err := m.find(key)
	return node != nil, err
}
func (m *HashLongMap) Write(key, data uintptr) error {
	node, _, err := m.getOrCreate(key, true)
	if err == nil {
		node.data = data
	}
	return err
}
func (m *HashLongMap) Entry(key uintptr, create bool) (HashLongEntry, bool, bool, error) {
	node, created, err := m.getOrCreate(key, create)
	if err != nil || node == nil {
		return HashLongEntry{}, false, false, err
	}
	return HashLongEntry{node.key, node.data}, true, created, nil
}
func (m *HashLongMap) Remove(key uintptr) (uintptr, bool, error) {
	link, node, err := m.find(key)
	if err != nil || node == nil {
		return 0, false, err
	}
	*link = node.next
	node.next = nil
	node.alive = false
	m.size--
	return node.data, true, nil
}
func (m *HashLongMap) Size() (int, error) {
	if err := m.check(); err != nil {
		return 0, err
	}
	return m.size, nil
}
func (m *HashLongMap) Entries() ([]HashLongEntry, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	result := make([]HashLongEntry, 0, m.size)
	for _, head := range m.bins {
		for node := head; node != nil; node = node.next {
			result = append(result, HashLongEntry{node.key, node.data})
		}
	}
	return result, nil
}
func (m *HashLongMap) Close() error {
	if err := m.check(); err != nil {
		return err
	}
	for bin := range m.bins {
		for node := m.bins[bin]; node != nil; node = node.next {
			node.alive = false
		}
		m.bins[bin] = nil
	}
	m.size, m.closed = 0, true
	return nil
}

// HashGenericHash and HashGenericCompare represent hashGen.h's optional
// callback mode without introducing an interface or untyped Go value.
type HashGenericHash func(key uintptr, bins int) int
type HashGenericCompare func(first, second uintptr) int

type hashGenericNode struct {
	key, data uintptr
	next      *hashGenericNode
	alive     bool
}

type HashGenericEntry struct{ Key, Data uintptr }

type HashGenericMap struct {
	bins    []*hashGenericNode
	size    int
	hash    HashGenericHash
	compare HashGenericCompare
	limits  HashMapLimits
	closed  bool
}

func NewHashGenericMap(bins int, hash HashGenericHash, compare HashGenericCompare, limits HashMapLimits) (*HashGenericMap, error) {
	limits, err := normalizeHashMapLimits(limits)
	if err != nil || bins <= 0 || hash == nil || compare == nil {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	return &HashGenericMap{bins: make([]*hashGenericNode, bins), hash: hash, compare: compare, limits: limits}, nil
}

func (m *HashGenericMap) check() error {
	if m == nil {
		return ErrDimension
	}
	if m.closed || len(m.bins) == 0 || m.hash == nil || m.compare == nil || m.size < 0 || m.size > m.limits.MaxEntries {
		return ErrInvariant
	}
	return nil
}

func (m *HashGenericMap) find(key uintptr) (**hashGenericNode, *hashGenericNode, error) {
	if err := m.check(); err != nil {
		return nil, nil, err
	}
	bin := m.hash(key, len(m.bins))
	if bin < 0 || bin >= len(m.bins) {
		return nil, nil, ErrInvariant
	}
	link := &m.bins[bin]
	for *link != nil {
		if m.compare((*link).key, key) == 0 {
			return link, *link, nil
		}
		link = &(*link).next
	}
	return link, nil, nil
}

func (m *HashGenericMap) getOrCreate(key uintptr, create bool) (*hashGenericNode, bool, error) {
	link, node, err := m.find(key)
	if err != nil || node != nil || !create {
		return node, false, err
	}
	if m.size == m.limits.MaxEntries {
		return nil, false, ErrLimit
	}
	node = &hashGenericNode{key: key, alive: true}
	*link = node
	m.size++
	return node, true, nil
}

func (m *HashGenericMap) Exists(key uintptr) (bool, error) {
	_, node, err := m.find(key)
	return node != nil, err
}

func (m *HashGenericMap) Write(key, data uintptr) error {
	node, _, err := m.getOrCreate(key, true)
	if err == nil {
		node.data = data
	}
	return err
}

type HashGenericCell struct {
	owner *HashGenericMap
	node  *hashGenericNode
}

func (m *HashGenericMap) EntryCell(key uintptr, create bool) (HashGenericCell, bool, error) {
	node, created, err := m.getOrCreate(key, create)
	if err != nil || node == nil {
		return HashGenericCell{}, false, err
	}
	return HashGenericCell{m, node}, created, nil
}

func (cell HashGenericCell) Key() (uintptr, error) {
	if cell.owner == nil || cell.node == nil || !cell.node.alive {
		return 0, ErrInvariant
	}
	if err := cell.owner.check(); err != nil {
		return 0, err
	}
	return cell.node.key, nil
}

func (cell HashGenericCell) Get() (uintptr, error) {
	if cell.owner == nil || cell.node == nil || !cell.node.alive {
		return 0, ErrInvariant
	}
	if err := cell.owner.check(); err != nil {
		return 0, err
	}
	return cell.node.data, nil
}

func (cell HashGenericCell) Set(value uintptr) error {
	if cell.owner == nil || cell.node == nil || !cell.node.alive {
		return ErrInvariant
	}
	if err := cell.owner.check(); err != nil {
		return err
	}
	cell.node.data = value
	return nil
}

func (m *HashGenericMap) Remove(key uintptr) (uintptr, bool, error) {
	link, node, err := m.find(key)
	if err != nil || node == nil {
		return 0, false, err
	}
	*link = node.next
	node.next = nil
	node.alive = false
	m.size--
	return node.data, true, nil
}

func (m *HashGenericMap) Size() (int, error) {
	if err := m.check(); err != nil {
		return 0, err
	}
	return m.size, nil
}

func (m *HashGenericMap) Entries() ([]HashGenericEntry, error) {
	if err := m.check(); err != nil {
		return nil, err
	}
	result := make([]HashGenericEntry, 0, m.size)
	for _, head := range m.bins {
		for node := head; node != nil; node = node.next {
			result = append(result, HashGenericEntry{node.key, node.data})
		}
	}
	return result, nil
}

func (m *HashGenericMap) Close() error {
	if err := m.check(); err != nil {
		return err
	}
	for bin := range m.bins {
		for node := m.bins[bin]; node != nil; node = node.next {
			node.alive = false
		}
		m.bins[bin] = nil
	}
	m.size, m.closed = 0, true
	return nil
}
