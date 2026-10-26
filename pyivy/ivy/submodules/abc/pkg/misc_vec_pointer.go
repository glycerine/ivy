package abc

// Typed pointer vectors from src/misc/vec/vecPtr.h.

import (
	"math"
	"slices"
	"strconv"
	"strings"
	"unsafe"
)

type PointerVectorLimits struct {
	MaxEntries int
	Stop       func() bool
}

func normalizePointerVectorLimits(limits PointerVectorLimits) (PointerVectorLimits, error) {
	if limits.MaxEntries < 0 {
		return limits, ErrDimension
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 1 << 26
	}
	return limits, nil
}

// PointerVector owns its slot array but not the pointed-to objects. Pointer
// identity supplies the source equality semantics without runtime interfaces.
type PointerVector[T any] struct {
	slots  []*T
	size   int
	limits PointerVectorLimits
	closed bool
}

func newPointerVector[T any](capacity int, exact bool, limits PointerVectorLimits) (*PointerVector[T], error) {
	limits, err := normalizePointerVectorLimits(limits)
	if err != nil || capacity < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if !exact && capacity > 0 && capacity < 8 {
		capacity = 8
	}
	if capacity > limits.MaxEntries {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &PointerVector[T]{slots: make([]*T, capacity), limits: limits}, nil
}

func NewPointerVector[T any](capacity int, limits PointerVectorLimits) (*PointerVector[T], error) {
	return newPointerVector[T](capacity, false, limits)
}

func NewExactPointerVector[T any](capacity int, limits PointerVectorLimits) (*PointerVector[T], error) {
	return newPointerVector[T](capacity, true, limits)
}

func NewPointerVectorWithSize[T any](size int, limits PointerVectorLimits) (*PointerVector[T], error) {
	vector, err := NewPointerVector[T](size, limits)
	if err != nil {
		return nil, err
	}
	vector.size = size
	return vector, nil
}

// NewPointerVectorFrom copies the caller's slots; the container never borrows
// a mutable slice even though Vec_PtrAllocArray could do so in C.
func NewPointerVectorFrom[T any](values []*T, limits PointerVectorLimits) (*PointerVector[T], error) {
	vector, err := NewExactPointerVector[T](len(values), limits)
	if err != nil {
		return nil, err
	}
	copy(vector.slots, values)
	vector.size = len(values)
	return vector, nil
}

func (vector *PointerVector[T]) check() error {
	if vector == nil {
		return ErrDimension
	}
	if vector.closed || vector.size < 0 || vector.size > len(vector.slots) || len(vector.slots) > vector.limits.MaxEntries {
		return ErrInvariant
	}
	return nil
}

func (vector *PointerVector[T]) stopped() bool {
	return vector.limits.Stop != nil && vector.limits.Stop()
}

func (vector *PointerVector[T]) growUnchecked(capacity int) {
	if capacity <= len(vector.slots) {
		return
	}
	slots := make([]*T, capacity)
	copy(slots, vector.slots[:vector.size])
	vector.slots = slots
}

func (vector *PointerVector[T]) Grow(capacity int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if capacity < 0 {
		return ErrDimension
	}
	if capacity <= len(vector.slots) {
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

func (vector *PointerVector[T]) Size() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return vector.size, nil
}

func (vector *PointerVector[T]) Capacity() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return len(vector.slots), nil
}

func (vector *PointerVector[T]) MemoryBytes() (int, error) {
	// Vec_PtrMemory explicitly accepts a null vector and reports zero.
	if vector == nil {
		return 0, nil
	}
	if err := vector.check(); err != nil {
		return 0, err
	}
	pointerBytes := int(unsafe.Sizeof(uintptr(0)))
	if len(vector.slots) > (math.MaxInt-(8+pointerBytes))/pointerBytes {
		return 0, ErrLimit
	}
	return 8 + pointerBytes + len(vector.slots)*pointerBytes, nil
}

func (vector *PointerVector[T]) CountNil() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	count := 0
	for _, value := range vector.slots[:vector.size] {
		if value == nil {
			count++
		}
	}
	return count, nil
}

func (vector *PointerVector[T]) Snapshot() ([]*T, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	return append([]*T(nil), vector.slots[:vector.size]...), nil
}

func (vector *PointerVector[T]) Entry(index int) (*T, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if index < 0 || index >= vector.size {
		return nil, ErrDimension
	}
	return vector.slots[index], nil
}

func (vector *PointerVector[T]) Last() (*T, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if vector.size == 0 {
		return nil, ErrDimension
	}
	return vector.slots[vector.size-1], nil
}

func (vector *PointerVector[T]) Write(index int, value *T) error {
	if err := vector.check(); err != nil {
		return err
	}
	if index < 0 || index >= vector.size {
		return ErrDimension
	}
	vector.slots[index] = value
	return nil
}

func (vector *PointerVector[T]) Fill(size int, value *T) error {
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
	for index := 0; index < size; index++ {
		vector.slots[index] = value
	}
	if size < vector.size {
		clear(vector.slots[size:vector.size])
	}
	vector.size = size
	return nil
}

func (vector *PointerVector[T]) FillAlternating(size int, even, odd *T) error {
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
	for index := 0; index < size; index++ {
		if index&1 == 0 {
			vector.slots[index] = even
		} else {
			vector.slots[index] = odd
		}
	}
	if size < vector.size {
		clear(vector.slots[size:vector.size])
	}
	vector.size = size
	return nil
}

func (vector *PointerVector[T]) FillExtra(size int, value *T) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 {
		return ErrDimension
	}
	if size <= vector.size {
		return nil
	}
	if size > vector.limits.MaxEntries {
		return ErrLimit
	}
	capacity := len(vector.slots)
	doubled := capacity
	if capacity <= math.MaxInt/2 {
		doubled = 2 * capacity
	} else {
		doubled = math.MaxInt
	}
	if size > doubled {
		capacity = size
	} else if size > capacity {
		capacity = doubled
	}
	if capacity < size {
		capacity = size
	}
	if capacity > vector.limits.MaxEntries {
		capacity = vector.limits.MaxEntries
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	for index := vector.size; index < size; index++ {
		vector.slots[index] = value
	}
	vector.size = size
	return nil
}

func (vector *PointerVector[T]) Get(index int) (*T, error) {
	if index < 0 || vector == nil {
		return nil, ErrDimension
	}
	if err := vector.check(); err != nil {
		return nil, err
	}
	if index >= vector.limits.MaxEntries {
		return nil, ErrLimit
	}
	if err := vector.FillExtra(index+1, nil); err != nil {
		return nil, err
	}
	return vector.slots[index], nil
}

func (vector *PointerVector[T]) Set(index int, value *T) error {
	if index < 0 || vector == nil {
		return ErrDimension
	}
	if err := vector.check(); err != nil {
		return err
	}
	if index >= vector.limits.MaxEntries {
		return ErrLimit
	}
	if err := vector.FillExtra(index+1, nil); err != nil {
		return err
	}
	vector.slots[index] = value
	return nil
}

func (vector *PointerVector[T]) Shrink(size int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 || size > vector.size {
		return ErrDimension
	}
	clear(vector.slots[size:vector.size])
	vector.size = size
	return nil
}

func (vector *PointerVector[T]) Clear() error {
	return vector.Shrink(0)
}

func (vector *PointerVector[T]) Clone() (*PointerVector[T], error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if vector.stopped() {
		return nil, ErrCancelled
	}
	clone := &PointerVector[T]{slots: make([]*T, len(vector.slots)), size: vector.size, limits: vector.limits}
	copy(clone.slots, vector.slots[:vector.size])
	return clone, nil
}

func (vector *PointerVector[T]) Move() (*PointerVector[T], error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	moved := &PointerVector[T]{slots: vector.slots, size: vector.size, limits: vector.limits}
	vector.slots = nil
	vector.size = 0
	return moved, nil
}

func (vector *PointerVector[T]) Release() ([]*T, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	result := append([]*T(nil), vector.slots[:vector.size]...)
	clear(vector.slots)
	vector.slots = nil
	vector.size = 0
	return result, nil
}

func (vector *PointerVector[T]) CopyFrom(source *PointerVector[T]) error {
	if err := vector.check(); err != nil {
		return err
	}
	if err := source.check(); err != nil {
		return err
	}
	if vector == source {
		return nil
	}
	if source.size > vector.limits.MaxEntries {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(source.size)
	clear(vector.slots[:vector.size])
	copy(vector.slots, source.slots[:source.size])
	vector.size = source.size
	return nil
}

func pointerVectorPushCapacity(capacity int) (int, error) {
	if capacity < 16 {
		return 16, nil
	}
	if capacity > math.MaxInt/2 {
		return 0, ErrLimit
	}
	return 2 * capacity, nil
}

func (vector *PointerVector[T]) Push(value *T) error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.size == vector.limits.MaxEntries {
		return ErrLimit
	}
	capacity := len(vector.slots)
	if vector.size == capacity {
		var err error
		capacity, err = pointerVectorPushCapacity(capacity)
		if err != nil || capacity > vector.limits.MaxEntries {
			return ErrLimit
		}
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	vector.slots[vector.size] = value
	vector.size++
	return nil
}

func (vector *PointerVector[T]) PushTwo(first, second *T) error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.size > vector.limits.MaxEntries-2 {
		return ErrLimit
	}
	capacity, size := len(vector.slots), vector.size
	for range 2 {
		if size == capacity {
			var err error
			capacity, err = pointerVectorPushCapacity(capacity)
			if err != nil || capacity > vector.limits.MaxEntries {
				return ErrLimit
			}
		}
		size++
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	vector.slots[vector.size] = first
	vector.slots[vector.size+1] = second
	vector.size += 2
	return nil
}

func (vector *PointerVector[T]) Append(other *PointerVector[T]) error {
	if err := vector.check(); err != nil {
		return err
	}
	if err := other.check(); err != nil {
		return err
	}
	if other.size > vector.limits.MaxEntries-vector.size {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	needed := vector.size + other.size
	capacity := len(vector.slots)
	for capacity < needed {
		var err error
		capacity, err = pointerVectorPushCapacity(capacity)
		if err != nil || capacity > vector.limits.MaxEntries {
			return ErrLimit
		}
	}
	vector.growUnchecked(capacity)
	copy(vector.slots[vector.size:], other.slots[:other.size])
	vector.size = needed
	return nil
}

func (vector *PointerVector[T]) PushFirst(value *T) error {
	if err := vector.Push(nil); err != nil {
		return err
	}
	copy(vector.slots[1:vector.size], vector.slots[:vector.size-1])
	vector.slots[0] = value
	return nil
}

// PushUnique reports true when the pointer was already present, matching the
// source's inverted integer return convention.
func (vector *PointerVector[T]) PushUnique(value *T) (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	if vector.Find(value) >= 0 {
		return true, nil
	}
	return false, vector.Push(value)
}

func (vector *PointerVector[T]) Pop() (*T, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if vector.size == 0 {
		return nil, ErrDimension
	}
	vector.size--
	value := vector.slots[vector.size]
	vector.slots[vector.size] = nil
	return value, nil
}

func (vector *PointerVector[T]) Find(value *T) int {
	if vector == nil || vector.closed {
		return -1
	}
	for index, existing := range vector.slots[:vector.size] {
		if existing == value {
			return index
		}
	}
	return -1
}

func (vector *PointerVector[T]) Remove(value *T) error {
	if err := vector.check(); err != nil {
		return err
	}
	index := -1
	for candidate := vector.size - 1; candidate >= 0; candidate-- {
		if vector.slots[candidate] == value {
			index = candidate
			break
		}
	}
	if index < 0 {
		return ErrDimension
	}
	return vector.Drop(index)
}

func (vector *PointerVector[T]) Drop(index int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if index < 0 || index >= vector.size {
		return ErrDimension
	}
	copy(vector.slots[index:vector.size-1], vector.slots[index+1:vector.size])
	vector.size--
	vector.slots[vector.size] = nil
	return nil
}

func (vector *PointerVector[T]) Insert(index int, value *T) error {
	if err := vector.check(); err != nil {
		return err
	}
	if index < 0 || index >= vector.size {
		return ErrDimension
	}
	if err := vector.Push(nil); err != nil {
		return err
	}
	copy(vector.slots[index+1:vector.size], vector.slots[index:vector.size-1])
	vector.slots[index] = value
	return nil
}

func (vector *PointerVector[T]) Reorder(firstCount int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if firstCount < 0 || firstCount >= vector.size {
		return ErrDimension
	}
	if firstCount == 0 {
		return nil
	}
	prefix := append([]*T(nil), vector.slots[:firstCount]...)
	copy(vector.slots, vector.slots[firstCount:vector.size])
	copy(vector.slots[vector.size-firstCount:vector.size], prefix)
	return nil
}

func (vector *PointerVector[T]) Reverse() error {
	if err := vector.check(); err != nil {
		return err
	}
	for left, right := 0, vector.size-1; left < right; left, right = left+1, right-1 {
		vector.slots[left], vector.slots[right] = vector.slots[right], vector.slots[left]
	}
	return nil
}

func (vector *PointerVector[T]) Equal(other *PointerVector[T]) bool {
	if vector == nil || other == nil || vector.closed || other.closed || vector.size != other.size {
		return false
	}
	for index := 0; index < vector.size; index++ {
		if vector.slots[index] != other.slots[index] {
			return false
		}
	}
	return true
}

type PointerVectorRecord[T any] struct {
	Index int
	Value *T
}

func (vector *PointerVector[T]) Entries(start, stop int, reverse bool) ([]PointerVectorRecord[T], error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if start < 0 || stop < start || stop > vector.size {
		return nil, ErrDimension
	}
	result := make([]PointerVectorRecord[T], 0, stop-start)
	if reverse {
		for index := stop - 1; index >= start; index-- {
			result = append(result, PointerVectorRecord[T]{index, vector.slots[index]})
		}
		return result, nil
	}
	for index := start; index < stop; index++ {
		result = append(result, PointerVectorRecord[T]{index, vector.slots[index]})
	}
	return result, nil
}

type PointerVectorZip[A, B any] struct {
	Index  int
	First  *A
	Second *B
}

func ZipPointerVectors[A, B any](first *PointerVector[A], second *PointerVector[B]) ([]PointerVectorZip[A, B], error) {
	if err := first.check(); err != nil {
		return nil, err
	}
	if err := second.check(); err != nil {
		return nil, err
	}
	if second.size < first.size {
		return nil, ErrDimension
	}
	result := make([]PointerVectorZip[A, B], first.size)
	for index := range result {
		result[index] = PointerVectorZip[A, B]{index, first.slots[index], second.slots[index]}
	}
	return result, nil
}

type PointerVectorPair[A, B any] struct {
	FirstIndex  int
	SecondIndex int
	First       *A
	Second      *B
}

// PointerVectorPairs covers the source double-entry macro. The final odd entry
// is intentionally ignored, as in the macro condition i+1 < size.
func PointerVectorPairs[A, B any](vector *PointerVector[A], convert func(*A) *B) ([]PointerVectorPair[A, B], error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if convert == nil {
		return nil, ErrDimension
	}
	result := make([]PointerVectorPair[A, B], 0, vector.size/2)
	for index := 0; index+1 < vector.size; index += 2 {
		result = append(result, PointerVectorPair[A, B]{index, index + 1, vector.slots[index], convert(vector.slots[index+1])})
	}
	return result, nil
}

type PointerCompare[T any] func(first, second *T) int

func (vector *PointerVector[T]) Sort(compare PointerCompare[T]) error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.size < 2 {
		return nil
	}
	if vector.stopped() {
		return ErrCancelled
	}
	if compare != nil {
		slices.SortStableFunc(vector.slots[:vector.size], func(first, second *T) int {
			return compare(first, second)
		})
		return nil
	}
	// C's default pointer-address order is implementation-dependent. Assigning
	// each identity its first-encounter rank groups duplicates deterministically.
	ranks := make(map[*T]int, vector.size)
	for _, value := range vector.slots[:vector.size] {
		if _, exists := ranks[value]; !exists {
			ranks[value] = len(ranks)
		}
	}
	slices.SortStableFunc(vector.slots[:vector.size], func(first, second *T) int {
		return ranks[first] - ranks[second]
	})
	return nil
}

func (vector *PointerVector[T]) Uniqify(compare PointerCompare[T]) error {
	if err := vector.Sort(compare); err != nil {
		return err
	}
	if vector.size < 2 {
		return nil
	}
	write := 1
	for read := 1; read < vector.size; read++ {
		if vector.slots[read] != vector.slots[read-1] {
			vector.slots[write] = vector.slots[read]
			write++
		}
	}
	clear(vector.slots[write:vector.size])
	vector.size = write
	return nil
}

func (vector *PointerVector[T]) UniqifyBy(compare PointerCompare[T], freeDuplicate func(*T)) ([]int, error) {
	if compare == nil {
		return nil, ErrDimension
	}
	if err := vector.Sort(compare); err != nil {
		return nil, err
	}
	if vector.size == 0 {
		return []int{1}, nil
	}
	counts := []int{1}
	write := 1
	for read := 1; read < vector.size; read++ {
		if compare(vector.slots[read], vector.slots[write-1]) != 0 {
			vector.slots[write] = vector.slots[read]
			write++
			counts = append(counts, 1)
		} else {
			if freeDuplicate != nil {
				freeDuplicate(vector.slots[read])
			}
			counts[len(counts)-1]++
		}
	}
	clear(vector.slots[write:vector.size])
	vector.size = write
	return counts, nil
}

// FreeValues invokes a typed destructor for each nonnil value. Invalid integer
// sentinel pointers 1 and 2 are not representable by ordinary Go pointers.
func (vector *PointerVector[T]) FreeValues(freeValue func(*T)) error {
	if err := vector.check(); err != nil {
		return err
	}
	if freeValue == nil {
		return ErrDimension
	}
	for _, value := range vector.slots[:vector.size] {
		if value != nil {
			freeValue(value)
		}
	}
	return nil
}

func (vector *PointerVector[T]) Validate() error { return vector.check() }

func (vector *PointerVector[T]) Close() error {
	if err := vector.check(); err != nil {
		return err
	}
	clear(vector.slots)
	vector.slots = nil
	vector.size = 0
	vector.closed = true
	return nil
}

func ClonePointerStrings(vector *PointerVector[string]) (*PointerVector[string], error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	clone, err := NewExactPointerVector[string](len(vector.slots), vector.limits)
	if err != nil {
		return nil, err
	}
	clone.size = vector.size
	for index, value := range vector.slots[:vector.size] {
		if value != nil {
			copyValue := *value
			clone.slots[index] = &copyValue
		}
	}
	return clone, nil
}

func FindPointerString(vector *PointerVector[string], value string) (int, error) {
	if err := vector.check(); err != nil {
		return -1, err
	}
	for index, existing := range vector.slots[:vector.size] {
		if existing != nil && *existing == value {
			return index, nil
		}
	}
	return -1, nil
}

func FormatPointerNames(vector *PointerVector[string]) (string, error) {
	if err := vector.check(); err != nil {
		return "", err
	}
	var output strings.Builder
	output.WriteString("Vector has ")
	output.WriteString(strconv.Itoa(vector.size))
	output.WriteString(" entries: {")
	for _, name := range vector.slots[:vector.size] {
		if name == nil {
			output.WriteString("(null) ")
		} else {
			output.WriteString(*name)
			output.WriteByte(' ')
		}
	}
	output.WriteString(" }\n")
	return output.String(), nil
}

type PointerSimulationLimits struct {
	MaxEntries int
	MaxWords   int
	MaxCells   int
	Stop       func() bool
}

type PointerSimulationInfo struct {
	rows   [][]uint32
	words  int
	limits PointerSimulationLimits
	closed bool
}

func normalizePointerSimulationLimits(limits PointerSimulationLimits) (PointerSimulationLimits, error) {
	if limits.MaxEntries < 0 || limits.MaxWords < 0 || limits.MaxCells < 0 {
		return limits, ErrDimension
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 1 << 20
	}
	if limits.MaxWords == 0 {
		limits.MaxWords = 1 << 20
	}
	if limits.MaxCells == 0 {
		limits.MaxCells = 1 << 28
	}
	return limits, nil
}

func NewPointerSimulationInfo(entries, words int, limits PointerSimulationLimits) (*PointerSimulationInfo, error) {
	limits, err := normalizePointerSimulationLimits(limits)
	if err != nil || entries < 0 || words < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if entries > limits.MaxEntries || words > limits.MaxWords || entries != 0 && words > limits.MaxCells/entries {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	rows := make([][]uint32, entries)
	for index := range rows {
		rows[index] = make([]uint32, words)
	}
	return &PointerSimulationInfo{rows: rows, words: words, limits: limits}, nil
}

func (info *PointerSimulationInfo) check() error {
	if info == nil {
		return ErrDimension
	}
	if info.closed || info.words < 0 || len(info.rows) > info.limits.MaxEntries || info.words > info.limits.MaxWords || len(info.rows) != 0 && info.words > info.limits.MaxCells/len(info.rows) {
		return ErrInvariant
	}
	for _, row := range info.rows {
		if len(row) != info.words {
			return ErrInvariant
		}
	}
	return nil
}

func (info *PointerSimulationInfo) EntryCount() (int, error) {
	if err := info.check(); err != nil {
		return 0, err
	}
	return len(info.rows), nil
}

func (info *PointerSimulationInfo) WordCount() (int, error) {
	if err := info.check(); err != nil {
		return 0, err
	}
	return info.words, nil
}

func (info *PointerSimulationInfo) Row(index int) ([]uint32, error) {
	if err := info.check(); err != nil {
		return nil, err
	}
	if index < 0 || index >= len(info.rows) {
		return nil, ErrDimension
	}
	return append([]uint32(nil), info.rows[index]...), nil
}

func (info *PointerSimulationInfo) Write(index int, values []uint32) error {
	if err := info.check(); err != nil {
		return err
	}
	if index < 0 || index >= len(info.rows) || len(values) != info.words {
		return ErrDimension
	}
	copy(info.rows[index], values)
	return nil
}

func (info *PointerSimulationInfo) fillFrom(firstWord int, value uint32) error {
	if err := info.check(); err != nil {
		return err
	}
	if firstWord < 0 || firstWord > info.words {
		return ErrDimension
	}
	if info.limits.Stop != nil && info.limits.Stop() {
		return ErrCancelled
	}
	for _, row := range info.rows {
		for word := firstWord; word < info.words; word++ {
			row[word] = value
		}
	}
	return nil
}

func (info *PointerSimulationInfo) CleanFrom(firstWord int) error { return info.fillFrom(firstWord, 0) }
func (info *PointerSimulationInfo) FillFrom(firstWord int) error {
	return info.fillFrom(firstWord, math.MaxUint32)
}

func (info *PointerSimulationInfo) DoubleEntries() error {
	if err := info.check(); err != nil {
		return err
	}
	if len(info.rows) <= 1 {
		return ErrDimension
	}
	if len(info.rows) > info.limits.MaxEntries/2 || info.words != 0 && 2*len(info.rows) > info.limits.MaxCells/info.words {
		return ErrLimit
	}
	if info.limits.Stop != nil && info.limits.Stop() {
		return ErrCancelled
	}
	rows := make([][]uint32, 2*len(info.rows))
	for index := range rows {
		rows[index] = make([]uint32, info.words)
	}
	for index := range info.rows {
		copy(rows[index], info.rows[index])
	}
	info.rows = rows
	return nil
}

func (info *PointerSimulationInfo) DoubleWords() error {
	if err := info.check(); err != nil {
		return err
	}
	if len(info.rows) <= 1 {
		return ErrDimension
	}
	if info.words > info.limits.MaxWords/2 || len(info.rows) != 0 && 2*info.words > info.limits.MaxCells/len(info.rows) {
		return ErrLimit
	}
	if info.limits.Stop != nil && info.limits.Stop() {
		return ErrCancelled
	}
	for index, old := range info.rows {
		row := make([]uint32, 2*info.words)
		copy(row, old)
		info.rows[index] = row
	}
	info.words *= 2
	return nil
}

func NewPointerTruthTables(variables int, limits PointerSimulationLimits) (*PointerSimulationInfo, error) {
	if variables < 0 || variables > 30 {
		return nil, ErrDimension
	}
	words := 1
	if variables > 5 {
		words = 1 << uint(variables-5)
	}
	info, err := NewPointerSimulationInfo(variables, words, limits)
	if err != nil {
		return nil, err
	}
	masks := [...]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	for variable := 0; variable < variables; variable++ {
		for word := 0; word < words; word++ {
			if variable < 5 {
				info.rows[variable][word] = masks[variable]
			} else if word&(1<<uint(variable-5)) != 0 {
				info.rows[variable][word] = math.MaxUint32
			}
		}
	}
	return info, nil
}

func (info *PointerSimulationInfo) Validate() error { return info.check() }

func (info *PointerSimulationInfo) Close() error {
	if err := info.check(); err != nil {
		return err
	}
	for index := range info.rows {
		clear(info.rows[index])
		info.rows[index] = nil
	}
	info.rows = nil
	info.words = 0
	info.closed = true
	return nil
}
