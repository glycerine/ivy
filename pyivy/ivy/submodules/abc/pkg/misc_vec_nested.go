package abc

// Typed nested vectors from src/misc/vec/vecVec.h.

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"unsafe"
)

type NestedVectorLimits struct {
	MaxLevels         int
	MaxEntries        int
	MaxStorageEntries int
	Stop              func() bool
}

func normalizeNestedVectorLimits(limits NestedVectorLimits) (NestedVectorLimits, error) {
	if limits.MaxLevels < 0 || limits.MaxEntries < 0 || limits.MaxStorageEntries < 0 {
		return limits, ErrDimension
	}
	if limits.MaxLevels == 0 {
		limits.MaxLevels = 1 << 20
	}
	if limits.MaxEntries == 0 {
		limits.MaxEntries = 1 << 26
	}
	if limits.MaxStorageEntries == 0 {
		limits.MaxStorageEntries = 1 << 27
	}
	return limits, nil
}

// NestedVector owns every level and returns owned snapshots to callers. The
// comparable constraint is compile-time only and permits Vec_VecPushUnique's
// pointer-or-integer identity test without runtime interface dispatch.
type NestedVector[T comparable] struct {
	levels  [][]T
	total   int
	storage int
	limits  NestedVectorLimits
	closed  bool
}

func NewNestedVector[T comparable](capacity int, limits NestedVectorLimits) (*NestedVector[T], error) {
	limits, err := normalizeNestedVectorLimits(limits)
	if err != nil || capacity < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if capacity > 0 && capacity < 8 {
		capacity = 8
	}
	if capacity > limits.MaxLevels {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &NestedVector[T]{levels: make([][]T, 0, capacity), limits: limits}, nil
}

func NewNestedVectorWithLevels[T comparable](levels int, limits NestedVectorLimits) (*NestedVector[T], error) {
	vector, err := NewNestedVector[T](levels, limits)
	if err != nil || levels < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	vector.levels = vector.levels[:levels]
	return vector, nil
}

func (vector *NestedVector[T]) check() error {
	if vector == nil {
		return ErrDimension
	}
	if vector.closed || len(vector.levels) > cap(vector.levels) || cap(vector.levels) > vector.limits.MaxLevels || vector.total < 0 || vector.total > vector.limits.MaxEntries || vector.storage < 0 || vector.storage > vector.limits.MaxStorageEntries {
		return ErrInvariant
	}
	total, storage := 0, 0
	for _, level := range vector.levels {
		if len(level) > cap(level) || total > vector.limits.MaxEntries-len(level) || storage > vector.limits.MaxStorageEntries-cap(level) {
			return ErrInvariant
		}
		total += len(level)
		storage += cap(level)
	}
	if total != vector.total || storage != vector.storage {
		return ErrInvariant
	}
	return nil
}

func (vector *NestedVector[T]) stopped() bool {
	return vector.limits.Stop != nil && vector.limits.Stop()
}

func (vector *NestedVector[T]) Expand(level int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if level < 0 {
		return ErrDimension
	}
	if level < len(vector.levels) {
		return nil
	}
	needed := level + 1
	if needed <= 0 || needed > vector.limits.MaxLevels {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	if needed > cap(vector.levels) {
		levels := make([][]T, len(vector.levels), needed)
		copy(levels, vector.levels)
		vector.levels = levels
	}
	vector.levels = vector.levels[:needed]
	return nil
}

func (vector *NestedVector[T]) LevelCount() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return len(vector.levels), nil
}

func (vector *NestedVector[T]) Capacity() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return cap(vector.levels), nil
}

func (vector *NestedVector[T]) LevelSize(level int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if level < 0 || level >= len(vector.levels) {
		return 0, ErrDimension
	}
	return len(vector.levels[level]), nil
}

func (vector *NestedVector[T]) LevelCapacity(level int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if level < 0 || level >= len(vector.levels) {
		return 0, ErrDimension
	}
	return cap(vector.levels[level]), nil
}

func (vector *NestedVector[T]) TotalEntries() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return vector.total, nil
}

func (vector *NestedVector[T]) Level(level int) ([]T, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if level < 0 || level >= len(vector.levels) {
		return nil, ErrDimension
	}
	return append([]T(nil), vector.levels[level]...), nil
}

func (vector *NestedVector[T]) Entry(level, index int) (T, error) {
	var zero T
	if err := vector.check(); err != nil {
		return zero, err
	}
	if level < 0 || level >= len(vector.levels) || index < 0 || index >= len(vector.levels[level]) {
		return zero, ErrDimension
	}
	return vector.levels[level][index], nil
}

func nestedVectorNextCapacity(current int) (int, error) {
	if current < 16 {
		return 16, nil
	}
	if current > math.MaxInt/2 {
		return 0, ErrLimit
	}
	return 2 * current, nil
}

func (vector *NestedVector[T]) Push(level int, entry T) error {
	if err := vector.check(); err != nil {
		return err
	}
	if level < 0 {
		return ErrDimension
	}
	if vector.total == vector.limits.MaxEntries {
		return ErrLimit
	}
	if level >= vector.limits.MaxLevels {
		return ErrLimit
	}
	currentCapacity := 0
	if level < len(vector.levels) {
		currentCapacity = cap(vector.levels[level])
	}
	nextCapacity := currentCapacity
	if level >= len(vector.levels) || len(vector.levels[level]) == currentCapacity {
		var err error
		nextCapacity, err = nestedVectorNextCapacity(currentCapacity)
		if err != nil || nextCapacity-currentCapacity > vector.limits.MaxStorageEntries-vector.storage {
			return ErrLimit
		}
	}
	if vector.stopped() {
		return ErrCancelled
	}
	if err := vector.Expand(level); err != nil {
		return err
	}
	entries := vector.levels[level]
	if len(entries) == cap(entries) {
		grown := make([]T, len(entries), nextCapacity)
		copy(grown, entries)
		entries = grown
		vector.storage += nextCapacity - currentCapacity
	}
	entries = entries[:len(entries)+1]
	entries[len(entries)-1] = entry
	vector.levels[level] = entries
	vector.total++
	return nil
}

func (vector *NestedVector[T]) PushUnique(level int, entry T) (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	if level < 0 {
		return false, ErrDimension
	}
	if level < len(vector.levels) {
		for _, existing := range vector.levels[level] {
			if existing == entry {
				return false, nil
			}
		}
	}
	if err := vector.Push(level, entry); err != nil {
		return false, err
	}
	return true, nil
}

func (vector *NestedVector[T]) Clear() error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	for index := range vector.levels {
		clear(vector.levels[index])
		vector.levels[index] = vector.levels[index][:0]
	}
	vector.total = 0
	return nil
}

func (vector *NestedVector[T]) Clone() (*NestedVector[T], error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if vector.stopped() {
		return nil, ErrCancelled
	}
	clone := &NestedVector[T]{
		levels:  make([][]T, len(vector.levels), cap(vector.levels)),
		total:   vector.total,
		storage: vector.storage,
		limits:  vector.limits,
	}
	for index, level := range vector.levels {
		clone.levels[index] = make([]T, len(level), cap(level))
		copy(clone.levels[index], level)
	}
	return clone, nil
}

type NestedVectorLevel[T comparable] struct {
	Index  int
	Values []T
}

type NestedVectorEntry[T comparable] struct {
	Level int
	Index int
	Value T
}

// Levels returns owned snapshots for the half-open level range [start, stop).
func (vector *NestedVector[T]) Levels(start, stop int, reverse bool) ([]NestedVectorLevel[T], error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if start < 0 || stop < start || stop > len(vector.levels) {
		return nil, ErrDimension
	}
	result := make([]NestedVectorLevel[T], 0, stop-start)
	if reverse {
		for level := stop - 1; level >= start; level-- {
			result = append(result, NestedVectorLevel[T]{level, append([]T(nil), vector.levels[level]...)})
		}
		return result, nil
	}
	for level := start; level < stop; level++ {
		result = append(result, NestedVectorLevel[T]{level, append([]T(nil), vector.levels[level]...)})
	}
	return result, nil
}

// Entries replaces all pointer and integer nested iteration macros with a
// checked, owned traversal in their exact level/entry orders.
func (vector *NestedVector[T]) Entries(start, stop int, reverseLevels, reverseEntries bool) ([]NestedVectorEntry[T], error) {
	levels, err := vector.Levels(start, stop, reverseLevels)
	if err != nil {
		return nil, err
	}
	result := make([]NestedVectorEntry[T], 0)
	for _, level := range levels {
		if reverseEntries {
			for index := len(level.Values) - 1; index >= 0; index-- {
				result = append(result, NestedVectorEntry[T]{level.Index, index, level.Values[index]})
			}
			continue
		}
		for index, value := range level.Values {
			result = append(result, NestedVectorEntry[T]{level.Index, index, value})
		}
	}
	return result, nil
}

type NestedVectorLevelPair[A, B comparable] struct {
	Index  int
	First  []A
	Second []B
}

func ZipNestedVectorLevels[A, B comparable](first *NestedVector[A], second *NestedVector[B]) ([]NestedVectorLevelPair[A, B], error) {
	if err := first.check(); err != nil {
		return nil, err
	}
	if err := second.check(); err != nil {
		return nil, err
	}
	if len(second.levels) < len(first.levels) {
		return nil, ErrDimension
	}
	result := make([]NestedVectorLevelPair[A, B], len(first.levels))
	for index := range first.levels {
		result[index] = NestedVectorLevelPair[A, B]{
			Index:  index,
			First:  append([]A(nil), first.levels[index]...),
			Second: append([]B(nil), second.levels[index]...),
		}
	}
	return result, nil
}

func (vector *NestedVector[T]) SortLevelsBySize(reverse bool) error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	sort.SliceStable(vector.levels, func(first, second int) bool {
		if reverse {
			return len(vector.levels[first]) > len(vector.levels[second])
		}
		return len(vector.levels[first]) < len(vector.levels[second])
	})
	return nil
}

func SortNestedIntsByFirst(vector *NestedVector[int], reverse bool) error {
	if err := vector.check(); err != nil {
		return err
	}
	for _, level := range vector.levels {
		if len(level) == 0 {
			return ErrDimension
		}
	}
	if vector.stopped() {
		return ErrCancelled
	}
	sort.SliceStable(vector.levels, func(first, second int) bool {
		if reverse {
			return vector.levels[first][0] > vector.levels[second][0]
		}
		return vector.levels[first][0] < vector.levels[second][0]
	})
	return nil
}

func FormatNestedInts(vector *NestedVector[int], skipSingles bool) (string, error) {
	if err := vector.check(); err != nil {
		return "", err
	}
	var output strings.Builder
	for level, entries := range vector.levels {
		if len(entries) == 0 || skipSingles && len(entries) == 1 {
			continue
		}
		fmt.Fprintf(&output, " %4d : {", level)
		for _, entry := range entries {
			fmt.Fprintf(&output, " %d", entry)
		}
		output.WriteString(" }\n")
	}
	return output.String(), nil
}

// NativeMemoryBytes reports Vec_VecMemory/Vec_VecMemoryInt accounting for an
// ABI whose level element is elementBytes wide. It counts logical C vector
// headers and capacities, not Go runtime allocation overhead.
func (vector *NestedVector[T]) NativeMemoryBytes(elementBytes int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	pointerBytes := int(unsafe.Sizeof(uintptr(0)))
	if elementBytes <= 0 || elementBytes > math.MaxInt/vector.limits.MaxStorageEntries {
		return 0, ErrDimension
	}
	headerBytes := 2*4 + pointerBytes
	if cap(vector.levels) > (math.MaxInt-headerBytes)/pointerBytes {
		return 0, ErrLimit
	}
	memory := headerBytes + cap(vector.levels)*pointerBytes
	for _, level := range vector.levels {
		if cap(level) > (math.MaxInt-memory-headerBytes)/elementBytes {
			return 0, ErrLimit
		}
		memory += headerBytes + cap(level)*elementBytes
	}
	return memory, nil
}

func (vector *NestedVector[T]) Validate() error {
	return vector.check()
}

func (vector *NestedVector[T]) Close() error {
	if err := vector.check(); err != nil {
		return err
	}
	for index := range vector.levels {
		clear(vector.levels[index])
	}
	vector.levels = nil
	vector.total = 0
	vector.storage = 0
	vector.closed = true
	return nil
}
