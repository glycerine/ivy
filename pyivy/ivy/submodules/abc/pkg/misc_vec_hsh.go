package abc

// Integer and vector interning from src/misc/vec/vecHsh.h.

import "math"

type VectorHashLimits struct {
	MaxEntries  int
	MaxBins     int
	MaxDataInts int
	Stop        func() bool
}

func normalizeVectorHashLimits(limits VectorHashLimits) (VectorHashLimits, error) {
	if limits.MaxEntries < 0 || limits.MaxBins < 0 || limits.MaxDataInts < 0 {
		return limits, ErrDimension
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 1 << 24
	}
	if limits.MaxBins == 0 {
		limits.MaxBins = 1 << 26
	}
	if limits.MaxDataInts == 0 {
		limits.MaxDataInts = 1 << 28
	}
	return limits, nil
}

// hshPrimeAtLeast preserves Abc_PrimeCudd, including its treatment of one as
// prime and its omission of the even prime two.
func hshPrimeAtLeast(value, maximum int, stop func() bool) (int, error) {
	if value < 1 || value > maximum {
		return 0, ErrLimit
	}
	if value == 1 {
		return 1, nil
	}
	candidate := value
	if candidate&1 == 0 {
		candidate++
	}
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

func hshCheckTableSize(tableSize int) error {
	if tableSize <= 0 || uint64(tableSize) > math.MaxUint32 {
		return ErrDimension
	}
	return nil
}

func hshJenkinsInt32(values []int32) uint32 {
	var hash uint32
	for _, value := range values {
		word := uint32(value)
		for shift := uint(0); shift < 32; shift += 8 {
			hash += uint32(byte(word >> shift))
			hash += hash << 10
			hash ^= hash >> 6
		}
	}
	hash += hash << 3
	hash ^= hash >> 11
	hash += hash << 15
	return hash
}

func Int32JenkinsHashBin(value int32, tableSize int) (int, error) {
	if err := hshCheckTableSize(tableSize); err != nil {
		return 0, err
	}
	return int(hshJenkinsInt32([]int32{value}) % uint32(tableSize)), nil
}

func FixedIntRecordHashBin(values []int32, tableSize int) (int, error) {
	if err := hshCheckTableSize(tableSize); err != nil {
		return 0, err
	}
	return int(hshJenkinsInt32(values) % uint32(tableSize)), nil
}

func FixedIntRecordWeightedHashBin(values []int32, tableSize int) (int, error) {
	if err := hshCheckTableSize(tableSize); err != nil {
		return 0, err
	}
	primes := [...]uint32{4177, 5147, 5647, 6343, 7103, 7873, 8147}
	var key uint32
	character := 0
	for _, value := range values {
		word := uint32(value)
		for shift := uint(0); shift < 32; shift += 8 {
			key += uint32(byte(word>>shift)) * primes[character%len(primes)]
			character++
		}
	}
	return int(key % uint32(tableSize)), nil
}

func IntVectorHashBin(values []int32, tableSize int) (int, error) {
	if err := hshCheckTableSize(tableSize); err != nil {
		return 0, err
	}
	primes := [...]uint32{4177, 5147, 5647, 6343, 7103, 7873, 8147}
	var key uint32
	for index, value := range values {
		key += uint32(value) * primes[index%len(primes)]
	}
	return int(key % uint32(tableSize)), nil
}

type SingleIntInterner struct {
	values []int32
	next   []int
	bins   []int
	limits VectorHashLimits
	closed bool
}

func NewSingleIntInterner(initialEntries int, limits VectorHashLimits) (*SingleIntInterner, error) {
	limits, err := normalizeVectorHashLimits(limits)
	if err != nil || initialEntries <= 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if initialEntries > limits.MaxBins {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	binsCount, err := hshPrimeAtLeast(initialEntries, limits.MaxBins, limits.Stop)
	if err != nil {
		return nil, err
	}
	bins := make([]int, binsCount)
	for index := range bins {
		bins[index] = -1
	}
	return &SingleIntInterner{bins: bins, limits: limits}, nil
}

func (interner *SingleIntInterner) check() error {
	if interner == nil {
		return ErrDimension
	}
	if interner.closed || len(interner.bins) == 0 || len(interner.values) != len(interner.next) || len(interner.values) > interner.limits.MaxEntries {
		return ErrInvariant
	}
	return nil
}

func (interner *SingleIntInterner) resize() error {
	if len(interner.values) <= len(interner.bins) {
		return nil
	}
	if len(interner.bins) > interner.limits.MaxBins/2 {
		return ErrLimit
	}
	binsCount, err := hshPrimeAtLeast(2*len(interner.bins), interner.limits.MaxBins, interner.limits.Stop)
	if err != nil {
		return err
	}
	bins := make([]int, binsCount)
	for index := range bins {
		bins[index] = -1
	}
	next := make([]int, len(interner.next))
	for id, value := range interner.values {
		bin, _ := Int32JenkinsHashBin(value, binsCount)
		next[id] = bins[bin]
		bins[bin] = id
	}
	interner.bins, interner.next = bins, next
	return nil
}

func (interner *SingleIntInterner) lookup(value int32) (int, bool, error) {
	if err := interner.check(); err != nil {
		return 0, false, err
	}
	bin, _ := Int32JenkinsHashBin(value, len(interner.bins))
	steps := 0
	for id := interner.bins[bin]; id >= 0; id = interner.next[id] {
		if id >= len(interner.values) || steps >= len(interner.values) {
			return 0, false, ErrInvariant
		}
		if interner.values[id] == value {
			return id, true, nil
		}
		steps++
	}
	return 0, false, nil
}

func (interner *SingleIntInterner) Lookup(value int32) (int, bool, error) {
	return interner.lookup(value)
}

func (interner *SingleIntInterner) Intern(value int32) (int, bool, error) {
	if err := interner.check(); err != nil {
		return 0, false, err
	}
	if interner.limits.Stop != nil && interner.limits.Stop() {
		return 0, false, ErrCancelled
	}
	if err := interner.resize(); err != nil {
		return 0, false, err
	}
	if id, found, err := interner.lookup(value); err != nil || found {
		return id, false, err
	}
	if len(interner.values) == interner.limits.MaxEntries {
		return 0, false, ErrLimit
	}
	bin, _ := Int32JenkinsHashBin(value, len(interner.bins))
	cursor := &interner.bins[bin]
	for *cursor >= 0 {
		cursor = &interner.next[*cursor]
	}
	id := len(interner.values)
	*cursor = id
	interner.values = append(interner.values, value)
	interner.next = append(interner.next, -1)
	return id, true, nil
}

func (interner *SingleIntInterner) EntryCount() (int, error) {
	if err := interner.check(); err != nil {
		return 0, err
	}
	return len(interner.values), nil
}

func (interner *SingleIntInterner) Values() ([]int32, error) {
	if err := interner.check(); err != nil {
		return nil, err
	}
	return append([]int32(nil), interner.values...), nil
}

func (interner *SingleIntInterner) Validate() error {
	if err := interner.check(); err != nil {
		return err
	}
	seen := make([]bool, len(interner.values))
	for bin, id := range interner.bins {
		steps := 0
		for ; id >= 0; id = interner.next[id] {
			if id >= len(interner.values) || seen[id] || steps >= len(interner.values) {
				return ErrInvariant
			}
			actual, _ := Int32JenkinsHashBin(interner.values[id], len(interner.bins))
			if actual != bin {
				return ErrInvariant
			}
			seen[id] = true
			steps++
		}
	}
	for _, value := range seen {
		if !value {
			return ErrInvariant
		}
	}
	return nil
}

func (interner *SingleIntInterner) Close() error {
	if err := interner.check(); err != nil {
		return err
	}
	interner.values, interner.next, interner.bins = nil, nil, nil
	interner.closed = true
	return nil
}

type fixedRecordHashObject struct {
	source int
	next   int
}

type FixedIntRecordInterner struct {
	width   int
	data    []int32
	objects []fixedRecordHashObject
	bins    []int
	limits  VectorHashLimits
	closed  bool
}

func NewFixedIntRecordInterner(data []int32, width, initialEntries int, limits VectorHashLimits) (*FixedIntRecordInterner, error) {
	limits, err := normalizeVectorHashLimits(limits)
	if err != nil || width <= 0 || initialEntries <= 0 || len(data)%width != 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if len(data) > limits.MaxDataInts || initialEntries > limits.MaxBins {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	binsCount, err := hshPrimeAtLeast(initialEntries, limits.MaxBins, limits.Stop)
	if err != nil {
		return nil, err
	}
	bins := make([]int, binsCount)
	for index := range bins {
		bins[index] = -1
	}
	return &FixedIntRecordInterner{width: width, data: append([]int32(nil), data...), bins: bins, limits: limits}, nil
}

func (interner *FixedIntRecordInterner) check() error {
	if interner == nil {
		return ErrDimension
	}
	if interner.closed || interner.width <= 0 || len(interner.data)%interner.width != 0 || len(interner.bins) == 0 || len(interner.objects) > interner.limits.MaxEntries || len(interner.data) > interner.limits.MaxDataInts {
		return ErrInvariant
	}
	return nil
}

func (interner *FixedIntRecordInterner) record(index int) ([]int32, error) {
	if err := interner.check(); err != nil {
		return nil, err
	}
	if index < 0 || index >= len(interner.data)/interner.width {
		return nil, ErrDimension
	}
	start := index * interner.width
	return interner.data[start : start+interner.width], nil
}

func (interner *FixedIntRecordInterner) resize() error {
	if len(interner.objects) <= len(interner.bins) {
		return nil
	}
	if len(interner.bins) > interner.limits.MaxBins/2 {
		return ErrLimit
	}
	binsCount, err := hshPrimeAtLeast(2*len(interner.bins), interner.limits.MaxBins, interner.limits.Stop)
	if err != nil {
		return err
	}
	bins := make([]int, binsCount)
	for index := range bins {
		bins[index] = -1
	}
	next := make([]int, len(interner.objects))
	for id, object := range interner.objects {
		record, _ := interner.record(object.source)
		bin, _ := FixedIntRecordHashBin(record, binsCount)
		next[id] = bins[bin]
		bins[bin] = id
	}
	for id := range interner.objects {
		interner.objects[id].next = next[id]
	}
	interner.bins = bins
	return nil
}

func (interner *FixedIntRecordInterner) lookup(record []int32) (int, bool, error) {
	if err := interner.check(); err != nil {
		return 0, false, err
	}
	if len(record) != interner.width {
		return 0, false, ErrDimension
	}
	bin, _ := FixedIntRecordHashBin(record, len(interner.bins))
	steps := 0
	for id := interner.bins[bin]; id >= 0; id = interner.objects[id].next {
		if id >= len(interner.objects) || steps >= len(interner.objects) {
			return 0, false, ErrInvariant
		}
		stored, _ := interner.record(interner.objects[id].source)
		equal := true
		for index := range record {
			equal = equal && record[index] == stored[index]
		}
		if equal {
			return id, true, nil
		}
		steps++
	}
	return 0, false, nil
}

func (interner *FixedIntRecordInterner) Lookup(record []int32) (int, bool, error) {
	return interner.lookup(record)
}

func (interner *FixedIntRecordInterner) InternIndex(index int) (int, bool, error) {
	record, err := interner.record(index)
	if err != nil {
		return 0, false, err
	}
	if interner.limits.Stop != nil && interner.limits.Stop() {
		return 0, false, ErrCancelled
	}
	if err = interner.resize(); err != nil {
		return 0, false, err
	}
	if id, found, lookupErr := interner.lookup(record); lookupErr != nil || found {
		return id, false, lookupErr
	}
	if len(interner.objects) == interner.limits.MaxEntries {
		return 0, false, ErrLimit
	}
	bin, _ := FixedIntRecordHashBin(record, len(interner.bins))
	cursor := &interner.bins[bin]
	for *cursor >= 0 {
		cursor = &interner.objects[*cursor].next
	}
	id := len(interner.objects)
	*cursor = id
	interner.objects = append(interner.objects, fixedRecordHashObject{source: index, next: -1})
	return id, true, nil
}

func (interner *FixedIntRecordInterner) AppendRecord(record []int32) (int, error) {
	if err := interner.check(); err != nil {
		return 0, err
	}
	if len(record) != interner.width {
		return 0, ErrDimension
	}
	if len(record) > interner.limits.MaxDataInts-len(interner.data) {
		return 0, ErrLimit
	}
	if interner.limits.Stop != nil && interner.limits.Stop() {
		return 0, ErrCancelled
	}
	index := len(interner.data) / interner.width
	interner.data = append(interner.data, record...)
	return index, nil
}

func (interner *FixedIntRecordInterner) AppendAndIntern(record []int32) (int, int, bool, error) {
	index, err := interner.AppendRecord(record)
	if err != nil {
		return 0, 0, false, err
	}
	id, created, err := interner.InternIndex(index)
	return index, id, created, err
}

func (interner *FixedIntRecordInterner) Entry(id int) ([]int32, int, error) {
	if err := interner.check(); err != nil {
		return nil, 0, err
	}
	if id < 0 || id >= len(interner.objects) {
		return nil, 0, ErrDimension
	}
	source := interner.objects[id].source
	record, _ := interner.record(source)
	return append([]int32(nil), record...), source, nil
}

func (interner *FixedIntRecordInterner) EntryCount() (int, error) {
	if err := interner.check(); err != nil {
		return 0, err
	}
	return len(interner.objects), nil
}

func (interner *FixedIntRecordInterner) Data() ([]int32, error) {
	if err := interner.check(); err != nil {
		return nil, err
	}
	return append([]int32(nil), interner.data...), nil
}

func (interner *FixedIntRecordInterner) Validate() error {
	if err := interner.check(); err != nil {
		return err
	}
	seen := make([]bool, len(interner.objects))
	for bin, id := range interner.bins {
		steps := 0
		for ; id >= 0; id = interner.objects[id].next {
			if id >= len(interner.objects) || seen[id] || steps >= len(interner.objects) {
				return ErrInvariant
			}
			record, err := interner.record(interner.objects[id].source)
			if err != nil {
				return ErrInvariant
			}
			actual, _ := FixedIntRecordHashBin(record, len(interner.bins))
			if actual != bin {
				return ErrInvariant
			}
			seen[id] = true
			steps++
		}
	}
	for _, value := range seen {
		if !value {
			return ErrInvariant
		}
	}
	return nil
}

func (interner *FixedIntRecordInterner) Close() error {
	if err := interner.check(); err != nil {
		return err
	}
	interner.data, interner.objects, interner.bins = nil, nil, nil
	interner.closed = true
	return nil
}

func HashFixedIntRecords(data []int32, width int, limits VectorHashLimits) ([]int, error) {
	if width <= 0 || len(data)%width != 0 {
		return nil, ErrDimension
	}
	entries := len(data) / width
	if entries == 0 {
		return []int{}, nil
	}
	interner, err := NewFixedIntRecordInterner(data, width, entries, limits)
	if err != nil {
		return nil, err
	}
	result := make([]int, entries)
	for index := range result {
		result[index], _, err = interner.InternIndex(index)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func wordsAsInt32(words []uint64) []int32 {
	result := make([]int32, 2*len(words))
	for index, word := range words {
		result[2*index] = int32(uint32(word))
		result[2*index+1] = int32(uint32(word >> 32))
	}
	return result
}

func HashFixedWordRecords(data []uint64, width int, limits VectorHashLimits) ([]int, error) {
	if width <= 0 || len(data)%width != 0 {
		return nil, ErrDimension
	}
	return HashFixedIntRecords(wordsAsInt32(data), 2*width, limits)
}

func NewUniqueWordRecordInterner(data []uint64, width int, limits VectorHashLimits) (*FixedIntRecordInterner, error) {
	if width <= 0 || len(data)%width != 0 || len(data) == 0 {
		return nil, ErrDimension
	}
	entries := len(data) / width
	interner, err := NewFixedIntRecordInterner(wordsAsInt32(data), 2*width, entries, limits)
	if err != nil {
		return nil, err
	}
	for index := 0; index < entries; index++ {
		_, created, internErr := interner.InternIndex(index)
		if internErr != nil {
			return nil, internErr
		}
		if !created {
			return nil, ErrInvariant
		}
	}
	return interner, nil
}

type intVectorHashRecord struct {
	values []int32
	next   int
}

type IntVectorInterner struct {
	records    []intVectorHashRecord
	bins       []int
	storedInts int
	limits     VectorHashLimits
	closed     bool
}

func NewIntVectorInterner(initialEntries int, limits VectorHashLimits) (*IntVectorInterner, error) {
	limits, err := normalizeVectorHashLimits(limits)
	if err != nil || initialEntries <= 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if initialEntries > limits.MaxBins {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	binsCount, err := hshPrimeAtLeast(initialEntries, limits.MaxBins, limits.Stop)
	if err != nil {
		return nil, err
	}
	bins := make([]int, binsCount)
	for index := range bins {
		bins[index] = -1
	}
	return &IntVectorInterner{bins: bins, limits: limits}, nil
}

func (interner *IntVectorInterner) check() error {
	if interner == nil {
		return ErrDimension
	}
	if interner.closed || len(interner.bins) == 0 || len(interner.records) > interner.limits.MaxEntries || interner.storedInts < 0 || interner.storedInts > interner.limits.MaxDataInts {
		return ErrInvariant
	}
	return nil
}

func (interner *IntVectorInterner) resize() error {
	if len(interner.records) <= len(interner.bins) {
		return nil
	}
	if len(interner.bins) > interner.limits.MaxBins/2 {
		return ErrLimit
	}
	binsCount, err := hshPrimeAtLeast(2*len(interner.bins), interner.limits.MaxBins, interner.limits.Stop)
	if err != nil {
		return err
	}
	bins := make([]int, binsCount)
	for index := range bins {
		bins[index] = -1
	}
	next := make([]int, len(interner.records))
	for id, record := range interner.records {
		bin, _ := IntVectorHashBin(record.values, binsCount)
		next[id] = bins[bin]
		bins[bin] = id
	}
	for id := range interner.records {
		interner.records[id].next = next[id]
	}
	interner.bins = bins
	return nil
}

func equalInt32Slices(first, second []int32) bool {
	if len(first) != len(second) {
		return false
	}
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return true
}

func (interner *IntVectorInterner) lookup(values []int32) (int, bool, error) {
	if err := interner.check(); err != nil {
		return 0, false, err
	}
	bin, _ := IntVectorHashBin(values, len(interner.bins))
	steps := 0
	for id := interner.bins[bin]; id >= 0; id = interner.records[id].next {
		if id >= len(interner.records) || steps >= len(interner.records) {
			return 0, false, ErrInvariant
		}
		if equalInt32Slices(interner.records[id].values, values) {
			return id, true, nil
		}
		steps++
	}
	return 0, false, nil
}

func (interner *IntVectorInterner) Lookup(values []int32) (int, bool, error) {
	return interner.lookup(values)
}

func (interner *IntVectorInterner) Intern(values []int32) (int, bool, error) {
	if err := interner.check(); err != nil {
		return 0, false, err
	}
	if interner.limits.Stop != nil && interner.limits.Stop() {
		return 0, false, ErrCancelled
	}
	if err := interner.resize(); err != nil {
		return 0, false, err
	}
	if id, found, err := interner.lookup(values); err != nil || found {
		return id, false, err
	}
	if len(interner.records) == interner.limits.MaxEntries {
		return 0, false, ErrLimit
	}
	storage := 2 + len(values) + len(values)%2
	if storage > interner.limits.MaxDataInts-interner.storedInts {
		return 0, false, ErrLimit
	}
	bin, _ := IntVectorHashBin(values, len(interner.bins))
	cursor := &interner.bins[bin]
	for *cursor >= 0 {
		cursor = &interner.records[*cursor].next
	}
	id := len(interner.records)
	*cursor = id
	interner.records = append(interner.records, intVectorHashRecord{values: append([]int32(nil), values...), next: -1})
	interner.storedInts += storage
	return id, true, nil
}

func (interner *IntVectorInterner) Entry(id int) ([]int32, error) {
	if err := interner.check(); err != nil {
		return nil, err
	}
	if id < 0 || id >= len(interner.records) {
		return nil, ErrDimension
	}
	return append([]int32(nil), interner.records[id].values...), nil
}

func (interner *IntVectorInterner) EntryCount() (int, error) {
	if err := interner.check(); err != nil {
		return 0, err
	}
	return len(interner.records), nil
}

func (interner *IntVectorInterner) StoredInts() (int, error) {
	if err := interner.check(); err != nil {
		return 0, err
	}
	return interner.storedInts, nil
}

func (interner *IntVectorInterner) Validate() error {
	if err := interner.check(); err != nil {
		return err
	}
	seen := make([]bool, len(interner.records))
	storage := 0
	for bin, id := range interner.bins {
		steps := 0
		for ; id >= 0; id = interner.records[id].next {
			if id >= len(interner.records) || seen[id] || steps >= len(interner.records) {
				return ErrInvariant
			}
			actual, _ := IntVectorHashBin(interner.records[id].values, len(interner.bins))
			if actual != bin {
				return ErrInvariant
			}
			seen[id] = true
			storage += 2 + len(interner.records[id].values) + len(interner.records[id].values)%2
			steps++
		}
	}
	for _, value := range seen {
		if !value {
			return ErrInvariant
		}
	}
	if storage != interner.storedInts {
		return ErrInvariant
	}
	return nil
}

func (interner *IntVectorInterner) Close() error {
	if err := interner.check(); err != nil {
		return err
	}
	interner.records, interner.bins = nil, nil
	interner.storedInts = 0
	interner.closed = true
	return nil
}
