package abc

// Integer pair/triple interning from src/misc/vec/vecHash.h.

import "math"

type IntTupleHashLimits struct {
	MaxEntries int
	MaxBins    int
	Stop       func() bool
}

type IntTupleHashEntry struct {
	Data0 int
	Data1 int
	Data2 int
}

type intTupleHashNode struct {
	entry IntTupleHashEntry
	next  int
}

// IntTupleHash assigns stable one-based IDs in first-insertion order. Pair
// managers treat Data2 as mutable payload; triple managers treat it as key.
type IntTupleHash struct {
	arity  int
	bins   []int
	nodes  []intTupleHashNode
	limits IntTupleHashLimits
	closed bool
}

func normalizeIntTupleHashLimits(limits IntTupleHashLimits) (IntTupleHashLimits, error) {
	if limits.MaxEntries < 0 || limits.MaxBins < 0 {
		return limits, ErrDimension
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 1 << 24
	}
	if limits.MaxBins == 0 {
		limits.MaxBins = 1 << 26
	}
	return limits, nil
}

func vecHashCheckedInt32(value int) error {
	if int64(value) < math.MinInt32 || int64(value) > math.MaxInt32 {
		return ErrDimension
	}
	return nil
}

func vecHashPrimeAtLeast(value, maximum int, stop func() bool) (int, error) {
	if value < 1 || maximum < value {
		return 0, ErrLimit
	}
	if value <= 2 {
		return 2, nil
	}
	candidate := value | 1
	for candidate <= maximum {
		if stop != nil && stop() {
			return 0, ErrCancelled
		}
		prime := true
		for divisor := 3; divisor <= candidate/divisor; divisor += 2 {
			if candidate%divisor == 0 {
				prime = false
				break
			}
		}
		if prime {
			return candidate, nil
		}
		if candidate > maximum-2 {
			break
		}
		candidate += 2
	}
	return 0, ErrLimit
}

func newIntTupleHash(initialSize, arity int, limits IntTupleHashLimits) (*IntTupleHash, error) {
	limits, err := normalizeIntTupleHashLimits(limits)
	if err != nil || initialSize < 0 || arity < 2 || arity > 3 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	if initialSize > limits.MaxBins-100 {
		return nil, ErrLimit
	}
	binCount, err := vecHashPrimeAtLeast(initialSize+100, limits.MaxBins, limits.Stop)
	if err != nil {
		return nil, err
	}
	return &IntTupleHash{
		arity:  arity,
		bins:   make([]int, binCount),
		nodes:  make([]intTupleHashNode, 1),
		limits: limits,
	}, nil
}

func NewIntPairHash(initialSize int, limits IntTupleHashLimits) (*IntTupleHash, error) {
	return newIntTupleHash(initialSize, 2, limits)
}

func NewIntTripleHash(initialSize int, limits IntTupleHashLimits) (*IntTupleHash, error) {
	return newIntTupleHash(initialSize, 3, limits)
}

func (hash *IntTupleHash) check() error {
	if hash == nil {
		return ErrDimension
	}
	if hash.closed || hash.arity < 2 || hash.arity > 3 || len(hash.bins) == 0 || len(hash.nodes) == 0 || len(hash.nodes)-1 > hash.limits.MaxEntries {
		return ErrInvariant
	}
	return nil
}

// IntPairHashBin preserves Hash_Int2ManHash's uint32 overflow.
func IntPairHashBin(data0, data1, tableSize int) (int, error) {
	if tableSize <= 0 || uint64(tableSize) > math.MaxUint32 || vecHashCheckedInt32(data0) != nil || vecHashCheckedInt32(data1) != nil {
		return 0, ErrDimension
	}
	value := uint32(4177)*uint32(int32(data0)) + uint32(7873)*uint32(int32(data1))
	return int(value % uint32(tableSize)), nil
}

// IntTripleHashBin preserves Hsh_Int3ManHash's uint32 overflow.
func IntTripleHashBin(data0, data1, data2, tableSize int) (int, error) {
	if tableSize <= 0 || uint64(tableSize) > math.MaxUint32 || vecHashCheckedInt32(data0) != nil || vecHashCheckedInt32(data1) != nil || vecHashCheckedInt32(data2) != nil {
		return 0, ErrDimension
	}
	value := uint32(4177)*uint32(int32(data0)) + uint32(7873)*uint32(int32(data1)) + uint32(1699)*uint32(int32(data2))
	return int(value % uint32(tableSize)), nil
}

func (hash *IntTupleHash) bin(entry IntTupleHashEntry, tableSize int) int {
	if hash.arity == 2 {
		value, _ := IntPairHashBin(entry.Data0, entry.Data1, tableSize)
		return value
	}
	value, _ := IntTripleHashBin(entry.Data0, entry.Data1, entry.Data2, tableSize)
	return value
}

func (hash *IntTupleHash) same(entry intTupleHashNode, data0, data1, data2 int) bool {
	return entry.entry.Data0 == data0 && entry.entry.Data1 == data1 && (hash.arity == 2 || entry.entry.Data2 == data2)
}

func (hash *IntTupleHash) lookup(data0, data1, data2 int) (int, bool, error) {
	if err := hash.check(); err != nil {
		return 0, false, err
	}
	if vecHashCheckedInt32(data0) != nil || vecHashCheckedInt32(data1) != nil || vecHashCheckedInt32(data2) != nil {
		return 0, false, ErrDimension
	}
	entry := IntTupleHashEntry{Data0: data0, Data1: data1, Data2: data2}
	for id := hash.bins[hash.bin(entry, len(hash.bins))]; id != 0; id = hash.nodes[id].next {
		if id < 1 || id >= len(hash.nodes) {
			return 0, false, ErrInvariant
		}
		if hash.same(hash.nodes[id], data0, data1, data2) {
			return id, true, nil
		}
	}
	return 0, false, nil
}

func (hash *IntTupleHash) resize() error {
	if len(hash.nodes) <= len(hash.bins) {
		return nil
	}
	if len(hash.bins) > hash.limits.MaxBins/2 {
		return ErrLimit
	}
	binCount, err := vecHashPrimeAtLeast(2*len(hash.bins), hash.limits.MaxBins, hash.limits.Stop)
	if err != nil {
		return err
	}
	bins := make([]int, binCount)
	next := make([]int, len(hash.nodes))
	for id := 1; id < len(hash.nodes); id++ {
		bin := hash.bin(hash.nodes[id].entry, binCount)
		cursor := &bins[bin]
		for *cursor != 0 {
			cursor = &next[*cursor]
		}
		*cursor = id
	}
	for id := 1; id < len(hash.nodes); id++ {
		hash.nodes[id].next = next[id]
	}
	hash.bins = bins
	return nil
}

func (hash *IntTupleHash) intern(data0, data1, data2 int) (int, bool, error) {
	if err := hash.check(); err != nil {
		return 0, false, err
	}
	if vecHashCheckedInt32(data0) != nil || vecHashCheckedInt32(data1) != nil || vecHashCheckedInt32(data2) != nil {
		return 0, false, ErrDimension
	}
	if hash.limits.Stop != nil && hash.limits.Stop() {
		return 0, false, ErrCancelled
	}
	if err := hash.resize(); err != nil {
		return 0, false, err
	}
	id, found, err := hash.lookup(data0, data1, data2)
	if err != nil || found {
		return id, false, err
	}
	if len(hash.nodes)-1 == hash.limits.MaxEntries {
		return 0, false, ErrLimit
	}
	entry := IntTupleHashEntry{Data0: data0, Data1: data1, Data2: data2}
	bin := hash.bin(entry, len(hash.bins))
	cursor := &hash.bins[bin]
	for *cursor != 0 {
		cursor = &hash.nodes[*cursor].next
	}
	id = len(hash.nodes)
	*cursor = id
	hash.nodes = append(hash.nodes, intTupleHashNode{entry: entry})
	return id, true, nil
}

func (hash *IntTupleHash) LookupPair(data0, data1 int) (int, bool, error) {
	if hash == nil || hash.arity != 2 {
		return 0, false, ErrDimension
	}
	return hash.lookup(data0, data1, 0)
}

func (hash *IntTupleHash) InternPair(data0, data1, payload int) (int, bool, error) {
	if hash == nil || hash.arity != 2 {
		return 0, false, ErrDimension
	}
	return hash.intern(data0, data1, payload)
}

func (hash *IntTupleHash) LookupTriple(data0, data1, data2 int) (int, bool, error) {
	if hash == nil || hash.arity != 3 {
		return 0, false, ErrDimension
	}
	return hash.lookup(data0, data1, data2)
}

func (hash *IntTupleHash) InternTriple(data0, data1, data2 int) (int, bool, error) {
	if hash == nil || hash.arity != 3 {
		return 0, false, ErrDimension
	}
	return hash.intern(data0, data1, data2)
}

func (hash *IntTupleHash) EntryCount() (int, error) {
	if err := hash.check(); err != nil {
		return 0, err
	}
	return len(hash.nodes) - 1, nil
}

func (hash *IntTupleHash) Entry(id int) (IntTupleHashEntry, error) {
	if err := hash.check(); err != nil {
		return IntTupleHashEntry{}, err
	}
	if id <= 0 || id >= len(hash.nodes) {
		return IntTupleHashEntry{}, ErrDimension
	}
	return hash.nodes[id].entry, nil
}

func (hash *IntTupleHash) Entries() ([]IntTupleHashEntry, error) {
	if err := hash.check(); err != nil {
		return nil, err
	}
	result := make([]IntTupleHashEntry, len(hash.nodes)-1)
	for id := 1; id < len(hash.nodes); id++ {
		result[id-1] = hash.nodes[id].entry
	}
	return result, nil
}

func (hash *IntTupleHash) PairPayloadIncrement(id int) (int, error) {
	if hash == nil || hash.arity != 2 {
		return 0, ErrDimension
	}
	entry, err := hash.Entry(id)
	if err != nil {
		return 0, err
	}
	if entry.Data2 == math.MaxInt32 {
		return 0, ErrLimit
	}
	hash.nodes[id].entry.Data2++
	return entry.Data2, nil
}

func (hash *IntTupleHash) PairPayloadDecrement(id int) (int, error) {
	if hash == nil || hash.arity != 2 {
		return 0, ErrDimension
	}
	entry, err := hash.Entry(id)
	if err != nil {
		return 0, err
	}
	if entry.Data2 == math.MinInt32 {
		return 0, ErrLimit
	}
	hash.nodes[id].entry.Data2--
	return entry.Data2 - 1, nil
}

func (hash *IntTupleHash) SetPairPayload(id, payload int) error {
	if hash == nil || hash.arity != 2 || vecHashCheckedInt32(payload) != nil {
		return ErrDimension
	}
	if _, err := hash.Entry(id); err != nil {
		return err
	}
	hash.nodes[id].entry.Data2 = payload
	return nil
}

// Profile returns the source print routine's chain lengths in bin order.
func (hash *IntTupleHash) Profile() ([]int, error) {
	if err := hash.check(); err != nil {
		return nil, err
	}
	result := make([]int, len(hash.bins))
	for bin, id := range hash.bins {
		for ; id != 0; id = hash.nodes[id].next {
			if id < 1 || id >= len(hash.nodes) {
				return nil, ErrInvariant
			}
			result[bin]++
		}
	}
	return result, nil
}

func (hash *IntTupleHash) Validate() error {
	if err := hash.check(); err != nil {
		return err
	}
	seen := make([]bool, len(hash.nodes))
	keys := make(map[[3]int]int, len(hash.nodes)-1)
	for bin, id := range hash.bins {
		steps := 0
		for ; id != 0; id = hash.nodes[id].next {
			if id < 1 || id >= len(hash.nodes) || seen[id] || steps >= len(hash.nodes) {
				return ErrInvariant
			}
			seen[id] = true
			entry := hash.nodes[id].entry
			if hash.bin(entry, len(hash.bins)) != bin {
				return ErrInvariant
			}
			key := [3]int{entry.Data0, entry.Data1, entry.Data2}
			if hash.arity == 2 {
				key[2] = 0
			}
			if keys[key] != 0 {
				return ErrInvariant
			}
			keys[key] = id
			steps++
		}
	}
	for id := 1; id < len(seen); id++ {
		if !seen[id] {
			return ErrInvariant
		}
	}
	return nil
}

func (hash *IntTupleHash) Close() error {
	if err := hash.check(); err != nil {
		return err
	}
	hash.bins = nil
	hash.nodes = nil
	hash.closed = true
	return nil
}
