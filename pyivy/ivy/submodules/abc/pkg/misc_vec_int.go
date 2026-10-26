package abc

// Integer vectors from src/misc/vec/vecInt.h.

import (
	"encoding/binary"
	"math"
	"slices"
	"strconv"
	"strings"
)

type IntVectorLimits struct {
	MaxEntries int
	Stop       func() bool
}

func normalizeIntVectorLimits(limits IntVectorLimits) (IntVectorLimits, error) {
	if limits.MaxEntries < 0 {
		return limits, ErrDimension
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 1 << 30
	}
	return limits, nil
}

// IntVector owns capacity-sized storage and a separate logical size. Values
// hidden by Shrink and Clear remain dormant, matching Vec_Int_t storage.
type IntVector struct {
	data   []int
	size   int
	limits IntVectorLimits
	closed bool
}

func newIntVector(capacity int, exact bool, limits IntVectorLimits) (*IntVector, error) {
	limits, err := normalizeIntVectorLimits(limits)
	if err != nil || capacity < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if !exact && capacity > 0 && capacity < 16 {
		capacity = 16
	}
	if capacity > limits.MaxEntries {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &IntVector{data: make([]int, capacity), limits: limits}, nil
}

func NewIntVector(capacity int, limits IntVectorLimits) (*IntVector, error) {
	return newIntVector(capacity, false, limits)
}

func NewExactIntVector(capacity int, limits IntVectorLimits) (*IntVector, error) {
	return newIntVector(capacity, true, limits)
}

func NewZeroIntVector(size int, limits IntVectorLimits) (*IntVector, error) {
	vector, err := NewIntVector(size, limits)
	if err != nil {
		return nil, err
	}
	vector.size = size
	return vector, nil
}

func NewFullIntVector(size int, limits IntVectorLimits) (*IntVector, error) {
	vector, err := NewIntVector(size, limits)
	if err != nil {
		return nil, err
	}
	for i := 0; i < size; i++ {
		vector.data[i] = -1
	}
	vector.size = size
	return vector, nil
}

func NewRangeIntVector(first, count int, limits IntVectorLimits) (*IntVector, error) {
	if count < 0 || count > 0 && (first > math.MaxInt-count+1 || first < math.MinInt+count-1) {
		return nil, ErrDimension
	}
	vector, err := NewIntVector(count, limits)
	if err != nil {
		return nil, err
	}
	for i := 0; i < count; i++ {
		vector.data[i] = first + i
	}
	vector.size = count
	return vector, nil
}

func NewNaturalIntVector(size int, limits IntVectorLimits) (*IntVector, error) {
	return NewRangeIntVector(0, size, limits)
}

// NewIntVectorFrom replaces both borrowing and copying constructors with
// caller-independent owned storage.
func NewIntVectorFrom(values []int, limits IntVectorLimits) (*IntVector, error) {
	vector, err := NewExactIntVector(len(values), limits)
	if err != nil {
		return nil, err
	}
	copy(vector.data, values)
	vector.size = len(values)
	return vector, nil
}

func (vector *IntVector) check() error {
	if vector == nil {
		return ErrDimension
	}
	if vector.limits.MaxEntries == 0 {
		vector.limits.MaxEntries = 1 << 30
	}
	if vector.closed || vector.size < 0 || vector.size > len(vector.data) || len(vector.data) > vector.limits.MaxEntries {
		return ErrInvariant
	}
	return nil
}

func (vector *IntVector) stopped() bool { return vector.limits.Stop != nil && vector.limits.Stop() }

func (vector *IntVector) Size() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return vector.size, nil
}

func (vector *IntVector) Capacity() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return len(vector.data), nil
}

func (vector *IntVector) MemoryBytes() int {
	if vector == nil {
		return 0
	}
	return 16 + 4*len(vector.data)
}

func (vector *IntVector) Entries() ([]int, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	return slices.Clone(vector.data[:vector.size]), nil
}

func (vector *IntVector) Entry(index int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if index < 0 || index >= vector.size {
		return 0, ErrDimension
	}
	return vector.data[index], nil
}

func (vector *IntVector) Last() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if vector.size == 0 {
		return 0, ErrDimension
	}
	return vector.data[vector.size-1], nil
}

func (vector *IntVector) Write(index, value int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if index < 0 || index >= vector.size {
		return ErrDimension
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.data[index] = value
	return nil
}

func (vector *IntVector) AddTo(index, addition int) (int, error) {
	value, err := vector.Entry(index)
	if err != nil {
		return 0, err
	}
	if addition > 0 && value > math.MaxInt-addition || addition < 0 && value < math.MinInt-addition {
		return 0, ErrLimit
	}
	if err = vector.Write(index, value+addition); err != nil {
		return 0, err
	}
	return value + addition, nil
}

func (vector *IntVector) UpdateMax(index, value int) error {
	old, err := vector.Entry(index)
	if err != nil || old >= value {
		return err
	}
	return vector.Write(index, value)
}

func (vector *IntVector) UpdateMin(index, value int) error {
	old, err := vector.Entry(index)
	if err != nil || old <= value {
		return err
	}
	return vector.Write(index, value)
}

func (vector *IntVector) growUnchecked(capacity int) {
	if capacity <= len(vector.data) {
		return
	}
	data := make([]int, capacity)
	copy(data, vector.data)
	vector.data = data
}

func (vector *IntVector) Grow(capacity int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if capacity < 0 {
		return ErrDimension
	}
	if capacity <= len(vector.data) {
		return nil
	}
	if capacity > vector.limits.MaxEntries {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	return nil
}

func (vector *IntVector) Resize(size int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 {
		return ErrDimension
	}
	if size > vector.limits.MaxEntries {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(size)
	vector.size = size
	return nil
}

func (vector *IntVector) Fill(size, value int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 {
		return ErrDimension
	}
	if size > vector.limits.MaxEntries {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(size)
	for i := 0; i < size; i++ {
		vector.data[i] = value
	}
	vector.size = size
	return nil
}

func (vector *IntVector) FillTwo(size, even, odd int) error {
	if err := vector.Fill(size, even); err != nil {
		return err
	}
	for i := 1; i < size; i += 2 {
		vector.data[i] = odd
	}
	return nil
}

func (vector *IntVector) FillNatural(size int) error {
	if err := vector.Fill(size, 0); err != nil {
		return err
	}
	for i := 0; i < size; i++ {
		vector.data[i] = i
	}
	return nil
}

func (vector *IntVector) FillExtra(size, value int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 {
		return ErrDimension
	}
	if size <= vector.size {
		return nil
	}
	capacity := len(vector.data)
	if size > 2*capacity {
		capacity = size
	} else if size > capacity {
		if capacity > math.MaxInt/2 {
			return ErrLimit
		}
		capacity *= 2
	}
	if capacity > vector.limits.MaxEntries || size > vector.limits.MaxEntries {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	for i := vector.size; i < size; i++ {
		vector.data[i] = value
	}
	vector.size = size
	return nil
}

func (vector *IntVector) Get(index, fill int) (int, error) {
	if index < 0 || index == math.MaxInt {
		return 0, ErrDimension
	}
	if err := vector.FillExtra(index+1, fill); err != nil {
		return 0, err
	}
	return vector.data[index], nil
}

func (vector *IntVector) SetSparse(index, value, fill int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if index < 0 || index == math.MaxInt {
		return ErrDimension
	}
	capacity := len(vector.data)
	if index >= vector.size {
		size := index + 1
		if size > 2*capacity {
			capacity = size
		} else if size > capacity {
			if capacity > math.MaxInt/2 {
				return ErrLimit
			}
			capacity *= 2
		}
		if capacity > vector.limits.MaxEntries || size > vector.limits.MaxEntries {
			return ErrLimit
		}
	}
	if vector.stopped() {
		return ErrCancelled
	}
	if index >= vector.size {
		oldSize := vector.size
		vector.growUnchecked(capacity)
		for i := oldSize; i <= index; i++ {
			vector.data[i] = fill
		}
		vector.size = index + 1
	}
	vector.data[index] = value
	return nil
}

func (vector *IntVector) Shrink(size int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 || size > vector.size {
		return ErrDimension
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.size = size
	return nil
}

func (vector *IntVector) Clear() error { return vector.Shrink(0) }

func (vector *IntVector) ensurePush(count int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if count < 0 || vector.size > vector.limits.MaxEntries-count {
		return ErrLimit
	}
	want := vector.size + count
	capacity := len(vector.data)
	for capacity < want {
		if capacity < 16 {
			capacity = 16
		} else if capacity > vector.limits.MaxEntries/2 {
			capacity = vector.limits.MaxEntries
		} else {
			capacity *= 2
		}
		if capacity < want && capacity == vector.limits.MaxEntries {
			return ErrLimit
		}
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	return nil
}

func (vector *IntVector) Push(values ...int) error {
	if err := vector.ensurePush(len(values)); err != nil {
		return err
	}
	copy(vector.data[vector.size:], values)
	vector.size += len(values)
	return nil
}

func (vector *IntVector) Prepend(value int) error { return vector.Insert(0, value) }

func (vector *IntVector) Insert(index, value int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if index < 0 || index > vector.size {
		return ErrDimension
	}
	if err := vector.ensurePush(1); err != nil {
		return err
	}
	copy(vector.data[index+1:], vector.data[index:vector.size])
	vector.data[index] = value
	vector.size++
	return nil
}

func (vector *IntVector) PushOrdered(value int, reverse bool) error {
	if err := vector.check(); err != nil {
		return err
	}
	index := vector.size
	for index > 0 && ((!reverse && vector.data[index-1] > value) || (reverse && vector.data[index-1] < value)) {
		index--
	}
	return vector.Insert(index, value)
}

func (vector *IntVector) PushOrderedByCost(value int, costs *IntVector) error {
	if err := vector.check(); err != nil {
		return err
	}
	valueCost, err := costs.Entry(value)
	if err != nil {
		return err
	}
	index := vector.size
	for index > 0 {
		cost, costErr := costs.Entry(vector.data[index-1])
		if costErr != nil {
			return costErr
		}
		if cost <= valueCost {
			break
		}
		index--
	}
	return vector.Insert(index, value)
}

func (vector *IntVector) PushUnique(value int) (bool, error) {
	index, err := vector.Find(value)
	if err != nil || index >= 0 {
		return index >= 0, err
	}
	return false, vector.Push(value)
}

func (vector *IntVector) PushUniqueOrdered(value int, reverse bool) (bool, error) {
	index, err := vector.Find(value)
	if err != nil || index >= 0 {
		return index >= 0, err
	}
	return false, vector.PushOrdered(value, reverse)
}

func (vector *IntVector) PushUniqueOrderedByCost(value int, costs *IntVector) (bool, error) {
	index, err := vector.Find(value)
	if err != nil || index >= 0 {
		return index >= 0, err
	}
	return false, vector.PushOrderedByCost(value, costs)
}

// Fetch atomically reserves existing capacity and returns its starting index.
// The native routine advances size before detecting insufficient capacity;
// checked Go leaves the vector unchanged instead.
func (vector *IntVector) Fetch(count int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if count < 0 {
		return 0, ErrDimension
	}
	if count == 0 {
		return -1, nil
	}
	if count > len(vector.data)-vector.size {
		return 0, ErrLimit
	}
	if vector.stopped() {
		return 0, ErrCancelled
	}
	start := vector.size
	vector.size += count
	return start, nil
}

func (vector *IntVector) Pop() (int, error) {
	value, err := vector.Last()
	if err != nil {
		return 0, err
	}
	if vector.stopped() {
		return 0, ErrCancelled
	}
	vector.size--
	return value, nil
}

func (vector *IntVector) Find(value int) (int, error) {
	if err := vector.check(); err != nil {
		return -1, err
	}
	for i, entry := range vector.data[:vector.size] {
		if entry == value {
			return i, nil
		}
	}
	return -1, nil
}

func (vector *IntVector) Remove(value int, afterFirst bool) (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	start := 0
	if afterFirst {
		start = 1
	}
	index := -1
	for i := start; i < vector.size; i++ {
		if vector.data[i] == value {
			index = i
			break
		}
	}
	if index < 0 {
		return false, nil
	}
	return true, vector.Drop(index)
}

func (vector *IntVector) Drop(index int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if index < 0 || index >= vector.size {
		return ErrDimension
	}
	if vector.stopped() {
		return ErrCancelled
	}
	copy(vector.data[index:], vector.data[index+1:vector.size])
	vector.size--
	return nil
}

func (vector *IntVector) MinMax(maximum bool) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if vector.size == 0 {
		return 0, nil
	}
	best := vector.data[0]
	for _, value := range vector.data[1:vector.size] {
		if maximum && value > best || !maximum && value < best {
			best = value
		}
	}
	return best, nil
}

func (vector *IntVector) Reverse() error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	slices.Reverse(vector.data[:vector.size])
	return nil
}

func (vector *IntVector) RemoveParity(removeOdd bool) error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.size&1 != 0 {
		return ErrDimension
	}
	if vector.stopped() {
		return ErrCancelled
	}
	offset := 0
	if !removeOdd {
		offset = 1
	}
	for i := 0; i < vector.size/2; i++ {
		vector.data[i] = vector.data[2*i+offset]
	}
	vector.size /= 2
	return nil
}

func (vector *IntVector) Invert(fill int) (*IntVector, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	maximum := -1
	for _, entry := range vector.data[:vector.size] {
		if entry < 0 && entry != fill {
			return nil, ErrDimension
		}
		if entry >= 0 {
			maximum = max(maximum, entry)
		}
	}
	if maximum == math.MaxInt {
		return nil, ErrLimit
	}
	// Vec_IntInvert starts with capacity zero and grows exactly through
	// Vec_IntFill, rather than using Vec_IntAlloc's minimum capacity.
	result, err := NewExactIntVector(maximum+1, vector.limits)
	if err != nil {
		return nil, err
	}
	if err = result.Fill(maximum+1, fill); err != nil {
		return nil, err
	}
	for i, entry := range vector.data[:vector.size] {
		if entry == fill {
			continue
		}
		result.data[entry] = i
	}
	return result, nil
}

func (vector *IntVector) Condense(fill int) (*IntVector, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	// Vec_IntCondense reserves the source size, including Vec_IntAlloc's
	// minimum-capacity rule, even when many entries are removed.
	result, err := NewIntVector(vector.size, vector.limits)
	if err != nil {
		return nil, err
	}
	for _, entry := range vector.data[:vector.size] {
		if entry != fill {
			result.data[result.size] = entry
			result.size++
		}
	}
	return result, nil
}

func (vector *IntVector) Sum(positiveOnly bool) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	total := 0
	for _, value := range vector.data[:vector.size] {
		if positiveOnly && value <= 0 {
			continue
		}
		if value > 0 && total > math.MaxInt-value || value < 0 && total < math.MinInt-value {
			return 0, ErrLimit
		}
		total += value
	}
	return total, nil
}

func (vector *IntVector) Count(value int) (int, error) {
	return vector.CountBy(func(entry int) bool { return entry == value })
}

func (vector *IntVector) CountLarger(value int) (int, error) {
	return vector.CountBy(func(entry int) bool { return entry > value })
}

func (vector *IntVector) CountSmaller(value int) (int, error) {
	return vector.CountBy(func(entry int) bool { return entry < value })
}

func (vector *IntVector) CountPositive() (int, error) {
	return vector.CountBy(func(entry int) bool { return entry > 0 })
}

func (vector *IntVector) CountZero() (int, error) {
	return vector.Count(0)
}

func (vector *IntVector) CountBy(predicate func(int) bool) (int, error) {
	if err := vector.check(); err != nil || predicate == nil {
		if err != nil {
			return 0, err
		}
		return 0, ErrDimension
	}
	count := 0
	for _, value := range vector.data[:vector.size] {
		if predicate(value) {
			count++
		}
	}
	return count, nil
}

func (vector *IntVector) Equal(other *IntVector) (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	if err := other.check(); err != nil {
		return false, err
	}
	return slices.Equal(vector.data[:vector.size], other.data[:other.size]), nil
}

func (vector *IntVector) CountCommon(other *IntVector) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if err := other.check(); err != nil {
		return 0, err
	}
	longer, shorter := vector, other
	if longer.size < shorter.size {
		longer, shorter = shorter, longer
	}
	present := make(map[int]struct{}, shorter.size)
	for _, value := range shorter.data[:shorter.size] {
		if value < 0 {
			return 0, ErrDimension
		}
		present[value] = struct{}{}
	}
	count := 0
	for _, value := range longer.data[:longer.size] {
		if value < 0 {
			return 0, ErrDimension
		}
		if _, ok := present[value]; ok {
			count++
		}
	}
	return count, nil
}

func (vector *IntVector) Sort(reverse bool) error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	slices.Sort(vector.data[:vector.size])
	if reverse {
		slices.Reverse(vector.data[:vector.size])
	}
	return nil
}

func (vector *IntVector) SortRecords(width int, reverse bool) error {
	if err := vector.check(); err != nil {
		return err
	}
	if width <= 0 || vector.size%width != 0 {
		return ErrDimension
	}
	if vector.stopped() {
		return ErrCancelled
	}
	records := make([][]int, vector.size/width)
	for i := range records {
		records[i] = slices.Clone(vector.data[i*width : (i+1)*width])
	}
	slices.SortStableFunc(records, func(a, b []int) int {
		comparison := 0
		if a[0] < b[0] {
			comparison = -1
		} else if a[0] > b[0] {
			comparison = 1
		}
		if reverse {
			return -comparison
		}
		return comparison
	})
	for i, record := range records {
		copy(vector.data[i*width:], record)
	}
	return nil
}

func (vector *IntVector) IsSorted(reverse bool) (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	for i := 1; i < vector.size; i++ {
		if reverse && vector.data[i-1] < vector.data[i] || !reverse && vector.data[i-1] > vector.data[i] {
			return false, nil
		}
	}
	return true, nil
}

func (vector *IntVector) Uniqify() (int, error) {
	if err := vector.Sort(false); err != nil {
		return 0, err
	}
	if vector.size < 2 {
		return 0, nil
	}
	old := vector.size
	write := 1
	for read := 1; read < old; read++ {
		if vector.data[read] != vector.data[read-1] {
			vector.data[write] = vector.data[read]
			write++
		}
	}
	vector.size = write
	return old - write, nil
}

func (vector *IntVector) CountDuplicates() (int, error) {
	clone, err := vector.Clone()
	if err != nil {
		return 0, err
	}
	return clone.Uniqify()
}

func (vector *IntVector) Unique() (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	seen := make(map[int]struct{}, vector.size)
	for _, value := range vector.data[:vector.size] {
		if _, exists := seen[value]; exists {
			return false, nil
		}
		seen[value] = struct{}{}
	}
	return true, nil
}

func (vector *IntVector) CountUnique() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	seen := make(map[int]struct{}, vector.size)
	for _, value := range vector.data[:vector.size] {
		if value < 0 {
			return 0, ErrDimension
		}
		seen[value] = struct{}{}
	}
	return len(seen), nil
}

func (vector *IntVector) UniqifyPairs() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if vector.size&1 != 0 {
		return 0, ErrDimension
	}
	if err := vector.SortRecords(2, false); err != nil {
		return 0, err
	}
	pairs := vector.size / 2
	if pairs < 2 {
		return 0, nil
	}
	write := 1
	for read := 1; read < pairs; read++ {
		if vector.data[2*read] != vector.data[2*(read-1)] || vector.data[2*read+1] != vector.data[2*(read-1)+1] {
			copy(vector.data[2*write:2*write+2], vector.data[2*read:2*read+2])
			write++
		}
	}
	vector.size = 2 * write
	return pairs - write, nil
}

func intRecordKey(values []int) string {
	bytes := make([]byte, 4*len(values))
	for i, value := range values {
		binary.LittleEndian.PutUint32(bytes[4*i:], uint32(int32(value)))
	}
	return string(bytes)
}

func validInt32Record(values []int) bool {
	for _, value := range values {
		if value < math.MinInt32 || value > math.MaxInt32 {
			return false
		}
	}
	return true
}

type IntVectorHashStep struct {
	ByteIndex int
	Byte      byte
	Key       uint32
	Bucket    uint32
}

// IntVectorRecordHashDebug returns the native debug routine's information as
// owned data instead of writing diagnostic text to process-global stdout.
func IntVectorRecordHashDebug(values []int, tableMask uint32) (uint32, []IntVectorHashStep, error) {
	if !validInt32Record(values) {
		return 0, nil, ErrDimension
	}
	primes := [4]uint32{12582917, 25165843, 50331653, 100663319}
	steps := make([]IntVectorHashStep, 0, 4*len(values))
	var key uint32
	for i, value := range values {
		var bytes [4]byte
		binary.LittleEndian.PutUint32(bytes[:], uint32(int32(value)))
		for j, b := range bytes {
			byteIndex := 4*i + j
			key += uint32(b) * primes[byteIndex&3]
			steps = append(steps, IntVectorHashStep{ByteIndex: byteIndex, Byte: b, Key: key, Bucket: key & tableMask})
		}
	}
	return key, steps, nil
}

// IntVectorRecordHash2 is Vec_IntUniqueHashKey2 with explicit little-endian
// four-byte source integers and uint32 wraparound.
func IntVectorRecordHash2(values []int) (uint32, error) {
	if !validInt32Record(values) {
		return 0, ErrDimension
	}
	primes := [4]uint32{12582917, 25165843, 50331653, 100663319}
	var key uint32
	for i, value := range values {
		var bytes [4]byte
		binary.LittleEndian.PutUint32(bytes[:], uint32(int32(value)))
		for j, b := range bytes {
			key += uint32(b) * primes[(4*i+j)&3]
		}
	}
	return key, nil
}

// IntVectorRecordHash is Vec_IntUniqueHashKey with explicit source width.
func IntVectorRecordHash(values []int) (uint32, error) {
	if !validInt32Record(values) {
		return 0, ErrDimension
	}
	primes := [16]uint32{0x984b6ad9, 0x18a6eed3, 0x950353e2, 0x6222f6eb, 0xdfbedd47, 0xef0f9023, 0xac932a26, 0x590eaf55, 0x97d0a034, 0xdc36cd2e, 0x22736b37, 0xdc9066b0, 0x2eb2f98b, 0x5d9c7baf, 0x85747c9e, 0x8aca1055}
	weights := [16]uint32{0x8d8a5ebe, 0x1e6a15dc, 0x197d49db, 0x5bab9c89, 0x4b55dea7, 0x55dede49, 0x9a6a8080, 0xe5e51035, 0xe148d658, 0x8a17eb3b, 0xe22e4b38, 0xe5be2a9a, 0xbe938cbb, 0x3b981069, 0x7f9c0c8e, 0xf756df10}
	var key uint32
	byteIndex := 0
	for _, value := range values {
		var bytes [4]byte
		binary.LittleEndian.PutUint32(bytes[:], uint32(int32(value)))
		for _, b := range bytes {
			key += weights[(2*byteIndex)&15]*primes[b&15] + weights[(2*byteIndex+1)&15]*primes[b>>4]
			byteIndex++
		}
	}
	return key, nil
}

// IntVectorRecordClasses preserves first occurrence IDs and source order.
func (vector *IntVector) IntVectorRecordClasses(width int) (int, []int, error) {
	if err := vector.check(); err != nil {
		return 0, nil, err
	}
	if width <= 0 || vector.size%width != 0 {
		return 0, nil, ErrDimension
	}
	classes := make([]int, vector.size/width)
	first := make(map[string]int, len(classes))
	unique := 0
	for i := range classes {
		record := vector.data[i*width : (i+1)*width]
		if !validInt32Record(record) {
			return 0, nil, ErrDimension
		}
		key := intRecordKey(record)
		id, ok := first[key]
		if !ok {
			id = i
			first[key] = id
			unique++
		}
		classes[i] = id
	}
	return unique, classes, nil
}

// IntVectorRecordProfile replaces Vec_IntUniqueProfile's stdout diagnostics
// with deterministic bucket occupancies. The bucket count is the smallest
// power of two that can hold the record count, as in the source hash table.
func (vector *IntVector) IntVectorRecordProfile(width int) ([]int, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if width <= 0 || vector.size%width != 0 {
		return nil, ErrDimension
	}
	count := vector.size / width
	buckets := 1
	for buckets < count {
		if buckets > math.MaxInt/2 {
			return nil, ErrLimit
		}
		buckets *= 2
	}
	profile := make([]int, buckets)
	for i := 0; i < count; i++ {
		hash, err := IntVectorRecordHash(vector.data[i*width : (i+1)*width])
		if err != nil {
			return nil, err
		}
		profile[int(hash&uint32(buckets-1))]++
	}
	return profile, nil
}

func (vector *IntVector) UniqifyRecords(width int) (*IntVector, error) {
	unique, classes, err := vector.IntVectorRecordClasses(width)
	if err != nil {
		return nil, err
	}
	if unique > vector.limits.MaxEntries/width {
		return nil, ErrLimit
	}
	result, err := NewIntVector(unique*width, vector.limits)
	if err != nil {
		return nil, err
	}
	for i, first := range classes {
		if first == i {
			if err = result.Push(vector.data[i*width : (i+1)*width]...); err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

func (vector *IntVector) SortUnsigned() error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	slices.SortStableFunc(vector.data[:vector.size], func(a, b int) int {
		ua, ub := uint32(a), uint32(b)
		if ua < ub {
			return -1
		}
		if ua > ub {
			return 1
		}
		return 0
	})
	return nil
}

func requireSorted(vector *IntVector, reverse bool) error {
	sorted, err := vector.IsSorted(reverse)
	if err != nil {
		return err
	}
	if !sorted {
		return ErrDimension
	}
	return nil
}

func CommonSorted(first, second *IntVector, reverse bool) ([]int, error) {
	if err := requireSorted(first, reverse); err != nil {
		return nil, err
	}
	if err := requireSorted(second, reverse); err != nil {
		return nil, err
	}
	a, b := first.data[:first.size], second.data[:second.size]
	result := make([]int, 0, min(len(a), len(b)))
	for i, j := 0, 0; i < len(a) && j < len(b); {
		if a[i] == b[j] {
			result = append(result, a[i])
			i++
			j++
		} else if (!reverse && a[i] < b[j]) || (reverse && a[i] > b[j]) {
			i++
		} else {
			j++
		}
	}
	return result, nil
}

func RemoveCommonSorted(first, second *IntVector) ([]int, error) {
	common, err := CommonSorted(first, second, false)
	if err != nil {
		return nil, err
	}
	if first.stopped() || second.stopped() {
		return nil, ErrCancelled
	}
	remove := func(vector *IntVector) {
		write, ci := 0, 0
		for read := 0; read < vector.size; read++ {
			if ci < len(common) && vector.data[read] == common[ci] {
				ci++
				continue
			}
			vector.data[write] = vector.data[read]
			write++
		}
		vector.size = write
	}
	remove(first)
	remove(second)
	return common, nil
}

func RemoveSorted(first, second *IntVector) error {
	if err := requireSorted(first, false); err != nil {
		return err
	}
	if err := requireSorted(second, false); err != nil {
		return err
	}
	if first.stopped() {
		return ErrCancelled
	}
	write, i, j := 0, 0, 0
	for i < first.size && j < second.size {
		if first.data[i] == second.data[j] {
			i++
			j++
		} else if first.data[i] < second.data[j] {
			first.data[write] = first.data[i]
			write++
			i++
		} else {
			j++
		}
	}
	for i < first.size {
		first.data[write] = first.data[i]
		write++
		i++
	}
	first.size = write
	return nil
}

func MergeSorted(first, second *IntVector) (*IntVector, error) {
	if err := requireSorted(first, false); err != nil {
		return nil, err
	}
	if err := requireSorted(second, false); err != nil {
		return nil, err
	}
	if second.size > math.MaxInt-first.size {
		return nil, ErrLimit
	}
	// Vec_IntTwoMerge reserves the sum of both input sizes before removing
	// cross-input duplicates. Preserve that capacity and its limit behavior.
	result, err := NewIntVector(first.size+second.size, first.limits)
	if err != nil {
		return nil, err
	}
	for i, j := 0, 0; i < first.size || j < second.size; {
		if i == first.size {
			copy(result.data[result.size:], second.data[j:second.size])
			result.size += second.size - j
			break
		}
		if j == second.size {
			copy(result.data[result.size:], first.data[i:first.size])
			result.size += first.size - i
			break
		}
		if first.data[i] == second.data[j] {
			result.data[result.size] = first.data[i]
			result.size++
			i++
			j++
		} else if first.data[i] < second.data[j] {
			result.data[result.size] = first.data[i]
			result.size++
			i++
		} else {
			result.data[result.size] = second.data[j]
			result.size++
			j++
		}
	}
	return result, nil
}

func SplitSorted(first, second *IntVector) (common, firstOnly, secondOnly []int, err error) {
	if err = requireSorted(first, false); err != nil {
		return
	}
	if err = requireSorted(second, false); err != nil {
		return
	}
	for i, j := 0, 0; i < first.size || j < second.size; {
		if i == first.size {
			secondOnly = append(secondOnly, second.data[j:]...)
			break
		}
		if j == second.size {
			firstOnly = append(firstOnly, first.data[i:]...)
			break
		}
		if first.data[i] == second.data[j] {
			common = append(common, first.data[i])
			i++
			j++
		} else if first.data[i] < second.data[j] {
			firstOnly = append(firstOnly, first.data[i])
			i++
		} else {
			secondOnly = append(secondOnly, second.data[j])
			j++
		}
	}
	return
}

func SelectSort(values []int, reverse bool) {
	for i := 0; i+1 < len(values); i++ {
		best := i
		for j := i + 1; j < len(values); j++ {
			if reverse && values[j] > values[best] || !reverse && values[j] < values[best] {
				best = j
			}
		}
		values[i], values[best] = values[best], values[i]
	}
}

func SelectSortByIndexedCost(values []int, costs *IntVector, reverse bool) error {
	if err := costs.check(); err != nil {
		return err
	}
	for _, value := range values {
		if value < 0 || value >= costs.size {
			return ErrDimension
		}
	}
	for i := 0; i+1 < len(values); i++ {
		best := i
		for j := i + 1; j < len(values); j++ {
			a, b := costs.data[values[j]], costs.data[values[best]]
			if reverse && a > b || !reverse && a < b {
				best = j
			}
		}
		values[i], values[best] = values[best], values[i]
	}
	return nil
}

func SelectSortParallel(values, costs []int, reverse bool) error {
	if len(values) != len(costs) {
		return ErrDimension
	}
	for i := 0; i+1 < len(values); i++ {
		best := i
		for j := i + 1; j < len(values); j++ {
			if reverse && costs[j] > costs[best] || !reverse && costs[j] < costs[best] {
				best = j
			}
		}
		values[i], values[best] = values[best], values[i]
		costs[i], costs[best] = costs[best], costs[i]
	}
	return nil
}

func (vector *IntVector) Format() (string, error) {
	if err := vector.check(); err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.WriteString("Vector has ")
	builder.WriteString(strconv.Itoa(vector.size))
	builder.WriteString(" entries: {")
	for _, value := range vector.data[:vector.size] {
		builder.WriteByte(' ')
		builder.WriteString(strconv.Itoa(value))
	}
	builder.WriteString(" }\n")
	return builder.String(), nil
}

func (vector *IntVector) BinaryText() (string, error) {
	if err := vector.check(); err != nil {
		return "", err
	}
	var builder strings.Builder
	builder.Grow(vector.size)
	for _, value := range vector.data[:vector.size] {
		if value == 0 {
			builder.WriteByte('0')
		} else {
			builder.WriteByte('1')
		}
	}
	return builder.String(), nil
}

func CompareIntVectors(first, second *IntVector) (int, error) {
	if first == nil || second == nil {
		if first == second {
			return 0, nil
		}
		if first == nil {
			return -1, nil
		}
		return 1, nil
	}
	if err := first.check(); err != nil {
		return 0, err
	}
	if err := second.check(); err != nil {
		return 0, err
	}
	if first.size != second.size {
		return first.size - second.size, nil
	}
	// Vec_IntCompareVec uses memcmp. ABC's integer-vector byte-oriented
	// routines assume four-byte little-endian source integers, so compare the
	// same byte stream instead of applying numeric integer ordering.
	for i := 0; i < first.size; i++ {
		a, b := uint32(first.data[i]), uint32(second.data[i])
		for shift := 0; shift < 32; shift += 8 {
			ab, bb := byte(a>>shift), byte(b>>shift)
			if ab < bb {
				return -1, nil
			}
			if ab > bb {
				return 1, nil
			}
		}
	}
	return 0, nil
}

func (vector *IntVector) Append(other *IntVector) error {
	values, err := other.Entries()
	if err != nil {
		return err
	}
	return vector.Push(values...)
}

func (vector *IntVector) AppendSkip(other *IntVector, skip int) error {
	values, err := other.Entries()
	if err != nil {
		return err
	}
	if skip < 0 || skip >= len(values) {
		return vector.Push(values...)
	}
	values = append(values[:skip], values[skip+1:]...)
	return vector.Push(values...)
}

func (vector *IntVector) ReplaceWithSigned(other *IntVector, negate bool) error {
	if err := vector.check(); err != nil {
		return err
	}
	values, err := other.Entries()
	if err != nil {
		return err
	}
	if negate {
		for i := range values {
			if values[i] == math.MinInt {
				return ErrLimit
			}
			values[i] = -values[i]
		}
	}
	if len(values) > vector.limits.MaxEntries {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(max(len(vector.data), len(values)))
	copy(vector.data, values)
	vector.size = len(values)
	return nil
}

func RemapIntVector(oldToNew, old *IntVector, newSize int, limits IntVectorLimits) (*IntVector, error) {
	if err := oldToNew.check(); err != nil {
		return nil, err
	}
	if err := old.check(); err != nil {
		return nil, err
	}
	if newSize < 0 {
		return nil, ErrDimension
	}
	result, err := NewZeroIntVector(newSize, limits)
	if err != nil {
		return nil, err
	}
	for oldID, newID := range oldToNew.data[:oldToNew.size] {
		if newID > 0 && newID < newSize && oldID < old.size && old.data[oldID] != 0 {
			result.data[newID] = old.data[oldID]
		}
	}
	return result, nil
}

func (vector *IntVector) Clone() (*IntVector, error) {
	values, err := vector.Entries()
	if err != nil {
		return nil, err
	}
	return NewIntVectorFrom(values, vector.limits)
}

func (vector *IntVector) Move() (*IntVector, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if vector.stopped() {
		return nil, ErrCancelled
	}
	result := &IntVector{data: vector.data, size: vector.size, limits: vector.limits}
	vector.data = nil
	vector.size = 0
	return result, nil
}

func (vector *IntVector) Release() ([]int, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if vector.stopped() {
		return nil, ErrCancelled
	}
	values := vector.data
	vector.data = nil
	vector.size = 0
	return values, nil
}

func (vector *IntVector) ReleaseLengthPrefixed() ([]int, error) {
	values, err := vector.Entries()
	if err != nil {
		return nil, err
	}
	result := make([]int, len(values)+1)
	result[0] = len(result)
	copy(result[1:], values)
	return result, nil
}

func (vector *IntVector) Erase() error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.data = nil
	vector.size = 0
	return nil
}

func (vector *IntVector) Close() error {
	if vector == nil || vector.closed {
		return nil
	}
	vector.data = nil
	vector.size = 0
	vector.closed = true
	return nil
}

type IntVectorIndexedEntry struct{ Index, Value int }
type IntVectorPair struct{ Index, First, Second int }
type IntVectorTriple struct{ Index, First, Second, Third int }

func (vector *IntVector) EntriesRange(start, stop int, reverse bool) ([]IntVectorIndexedEntry, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if start < 0 || stop < start || stop > vector.size {
		return nil, ErrDimension
	}
	result := make([]IntVectorIndexedEntry, 0, stop-start)
	if reverse {
		for i := stop - 1; i >= start; i-- {
			result = append(result, IntVectorIndexedEntry{i, vector.data[i]})
		}
	} else {
		for i := start; i < stop; i++ {
			result = append(result, IntVectorIndexedEntry{i, vector.data[i]})
		}
	}
	return result, nil
}

func ZipIntVectors(first, second *IntVector, start int) ([]IntVectorPair, error) {
	if err := first.check(); err != nil {
		return nil, err
	}
	if err := second.check(); err != nil {
		return nil, err
	}
	if start < 0 || start > first.size || second.size < first.size {
		return nil, ErrDimension
	}
	result := make([]IntVectorPair, 0, first.size-start)
	for i := start; i < first.size; i++ {
		result = append(result, IntVectorPair{i, first.data[i], second.data[i]})
	}
	return result, nil
}

func (vector *IntVector) Groups(width, start int) ([]IntVectorTriple, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if width != 2 && width != 3 || start < 0 || start > vector.size {
		return nil, ErrDimension
	}
	result := make([]IntVectorTriple, 0, (vector.size-start)/width)
	for i := start; i+width <= vector.size; i += width {
		entry := IntVectorTriple{Index: i, First: vector.data[i], Second: vector.data[i+1]}
		if width == 3 {
			entry.Third = vector.data[i+2]
		}
		result = append(result, entry)
	}
	return result, nil
}

// ThisNext reproduces the unusual source macro, which advances by two.
func (vector *IntVector) ThisNext() ([]IntVectorPair, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	result := make([]IntVectorPair, 0, vector.size/2)
	this := -1
	if vector.size != 0 {
		this = vector.data[0]
	}
	for i := 0; i+1 < vector.size; i += 2 {
		next := vector.data[i+1]
		result = append(result, IntVectorPair{i, this, next})
		this = next
	}
	return result, nil
}

func EntriesByIndex(values, indexes *IntVector) ([]int, error) {
	if err := values.check(); err != nil {
		return nil, err
	}
	if err := indexes.check(); err != nil {
		return nil, err
	}
	result := make([]int, indexes.size)
	for i, index := range indexes.data[:indexes.size] {
		if index < 0 || index >= values.size {
			return nil, ErrDimension
		}
		result[i] = values.data[index]
	}
	return result, nil
}
