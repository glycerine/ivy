package abc

// Embedded integer level vectors from src/misc/vec/vecWec.h.

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"unsafe"
)

type IntLevelVectorLimits struct {
	MaxLevels         int
	MaxEntries        int
	MaxStorageEntries int
	Stop              func() bool
}

func normalizeIntLevelVectorLimits(limits IntLevelVectorLimits) (IntLevelVectorLimits, error) {
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

type intLevelSlot struct {
	values []int
	marked bool
}

// IntLevelVector keeps one embedded level header per outer-capacity slot. Slots
// beyond size intentionally survive Init shrinkage and may become visible on a
// later Init, matching Vec_Wec's embedded-header behavior.
type IntLevelVector struct {
	slots   []intLevelSlot
	size    int
	entries int
	storage int
	limits  IntLevelVectorLimits
	closed  bool
}

func newIntLevelVector(capacity int, exact bool, limits IntLevelVectorLimits) (*IntLevelVector, error) {
	limits, err := normalizeIntLevelVectorLimits(limits)
	if err != nil || capacity < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if !exact && capacity > 0 && capacity < 8 {
		capacity = 8
	}
	if capacity > limits.MaxLevels {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &IntLevelVector{slots: make([]intLevelSlot, capacity), limits: limits}, nil
}

func NewIntLevelVector(capacity int, limits IntLevelVectorLimits) (*IntLevelVector, error) {
	return newIntLevelVector(capacity, false, limits)
}

func NewExactIntLevelVector(capacity int, limits IntLevelVectorLimits) (*IntLevelVector, error) {
	return newIntLevelVector(capacity, true, limits)
}

func NewIntLevelVectorWithLevels(levels int, limits IntLevelVectorLimits) (*IntLevelVector, error) {
	vector, err := NewIntLevelVector(levels, limits)
	if err != nil {
		return nil, err
	}
	vector.size = levels
	return vector, nil
}

func (vector *IntLevelVector) check() error {
	if vector == nil {
		return ErrDimension
	}
	if vector.closed || vector.size < 0 || vector.size > len(vector.slots) || len(vector.slots) > vector.limits.MaxLevels || vector.entries < 0 || vector.entries > vector.limits.MaxEntries || vector.storage < 0 || vector.storage > vector.limits.MaxStorageEntries {
		return ErrInvariant
	}
	entries, storage := 0, 0
	for _, slot := range vector.slots {
		if len(slot.values) > cap(slot.values) || entries > vector.limits.MaxEntries-len(slot.values) || storage > vector.limits.MaxStorageEntries-cap(slot.values) {
			return ErrInvariant
		}
		entries += len(slot.values)
		storage += cap(slot.values)
	}
	if entries != vector.entries || storage != vector.storage {
		return ErrInvariant
	}
	return nil
}

func (vector *IntLevelVector) stopped() bool {
	return vector.limits.Stop != nil && vector.limits.Stop()
}

func (vector *IntLevelVector) growUnchecked(capacity int) {
	if capacity <= len(vector.slots) {
		return
	}
	slots := make([]intLevelSlot, capacity)
	copy(slots, vector.slots)
	vector.slots = slots
}

func (vector *IntLevelVector) Grow(capacity int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if capacity < 0 {
		return ErrDimension
	}
	if capacity <= len(vector.slots) {
		return nil
	}
	if capacity > vector.limits.MaxLevels {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	return nil
}

func (vector *IntLevelVector) Init(size int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 {
		return ErrDimension
	}
	if size > vector.limits.MaxLevels {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(size)
	vector.size = size
	return nil
}

func (vector *IntLevelVector) Capacity() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return len(vector.slots), nil
}

func (vector *IntLevelVector) LevelCount() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return vector.size, nil
}

func (vector *IntLevelVector) LevelSize(level int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if level < 0 || level >= vector.size {
		return 0, ErrDimension
	}
	return len(vector.slots[level].values), nil
}

func (vector *IntLevelVector) LevelCapacity(level int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if level < 0 || level >= vector.size {
		return 0, ErrDimension
	}
	return cap(vector.slots[level].values), nil
}

func (vector *IntLevelVector) Level(level int) ([]int, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if level < 0 || level >= vector.size {
		return nil, ErrDimension
	}
	return append([]int(nil), vector.slots[level].values...), nil
}

// LevelID replaces pointer subtraction on an embedded Vec_Int header.
func (vector *IntLevelVector) LevelID(level int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if level < 0 || level >= vector.size {
		return 0, ErrDimension
	}
	return level, nil
}

func (vector *IntLevelVector) LastLevel() ([]int, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if vector.size == 0 {
		return nil, ErrDimension
	}
	return append([]int(nil), vector.slots[vector.size-1].values...), nil
}

func (vector *IntLevelVector) Entry(level, index int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if level < 0 || level >= vector.size || index < 0 || index >= len(vector.slots[level].values) {
		return 0, ErrDimension
	}
	return vector.slots[level].values[index], nil
}

func (vector *IntLevelVector) Snapshot() ([][]int, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	result := make([][]int, vector.size)
	for index := range result {
		result[index] = append([]int(nil), vector.slots[index].values...)
	}
	return result, nil
}

func (vector *IntLevelVector) TotalEntries() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	total := 0
	for index := 0; index < vector.size; index++ {
		total += len(vector.slots[index].values)
	}
	return total, nil
}

func (vector *IntLevelVector) UsedLevels() (int, error) {
	return vector.UsedLevelsRange(0, vector.size)
}

func (vector *IntLevelVector) UsedLevelsRange(start, stop int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if start < 0 || stop < start || stop > vector.size {
		return 0, ErrDimension
	}
	used := 0
	for level := start; level < stop; level++ {
		if len(vector.slots[level].values) != 0 {
			used++
		}
	}
	return used, nil
}

func intLevelOuterPushCapacity(current int) (int, error) {
	if current < 16 {
		return 16, nil
	}
	if current > math.MaxInt/2 {
		return 0, ErrLimit
	}
	return 2 * current, nil
}

func (vector *IntLevelVector) Push(level, entry int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if level < 0 {
		return ErrDimension
	}
	if level >= vector.limits.MaxLevels || vector.entries == vector.limits.MaxEntries {
		return ErrLimit
	}
	outerCapacity := len(vector.slots)
	if level >= outerCapacity {
		doubled := vector.size
		if doubled > math.MaxInt/2 {
			return ErrLimit
		}
		doubled *= 2
		outerCapacity = level + 1
		if doubled > outerCapacity {
			outerCapacity = doubled
		}
		if outerCapacity > vector.limits.MaxLevels {
			return ErrLimit
		}
	}
	currentInnerCapacity := 0
	currentInnerSize := 0
	if level < len(vector.slots) {
		currentInnerCapacity = cap(vector.slots[level].values)
		currentInnerSize = len(vector.slots[level].values)
	}
	nextInnerCapacity := currentInnerCapacity
	if currentInnerSize == currentInnerCapacity {
		var err error
		nextInnerCapacity, err = nestedVectorNextCapacity(currentInnerCapacity)
		if err != nil || nextInnerCapacity-currentInnerCapacity > vector.limits.MaxStorageEntries-vector.storage {
			return ErrLimit
		}
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(outerCapacity)
	if vector.size <= level {
		vector.size = level + 1
	}
	slot := &vector.slots[level]
	if len(slot.values) == cap(slot.values) {
		values := make([]int, len(slot.values), nextInnerCapacity)
		copy(values, slot.values)
		slot.values = values
		vector.storage += nextInnerCapacity - currentInnerCapacity
	}
	slot.values = slot.values[:len(slot.values)+1]
	slot.values[len(slot.values)-1] = entry
	vector.entries++
	return nil
}

func (vector *IntLevelVector) PushUnique(level, entry int) (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	if level < 0 {
		return false, ErrDimension
	}
	if level < vector.size {
		for _, existing := range vector.slots[level].values {
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

func (vector *IntLevelVector) PushLevel() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	capacity := len(vector.slots)
	if vector.size == capacity {
		var err error
		capacity, err = intLevelOuterPushCapacity(capacity)
		if err != nil || capacity > vector.limits.MaxLevels {
			return 0, ErrLimit
		}
	}
	if vector.size == vector.limits.MaxLevels {
		return 0, ErrLimit
	}
	if vector.stopped() {
		return 0, ErrCancelled
	}
	vector.growUnchecked(capacity)
	level := vector.size
	vector.size++
	return level, nil
}

func (vector *IntLevelVector) InsertLevel(index int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if index < 0 || index > vector.size {
		return 0, ErrDimension
	}
	capacity := len(vector.slots)
	if vector.size == capacity {
		var err error
		capacity, err = intLevelOuterPushCapacity(capacity)
		if err != nil || capacity > vector.limits.MaxLevels {
			return 0, ErrLimit
		}
	}
	if vector.size == vector.limits.MaxLevels {
		return 0, ErrLimit
	}
	if vector.stopped() {
		return 0, ErrCancelled
	}
	vector.growUnchecked(capacity)
	discarded := vector.slots[vector.size]
	vector.entries -= len(discarded.values)
	vector.storage -= cap(discarded.values)
	copy(vector.slots[index+1:vector.size+1], vector.slots[index:vector.size])
	vector.slots[index] = intLevelSlot{}
	vector.size++
	return index, nil
}

func (vector *IntLevelVector) Shrink(size int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 || size > vector.size {
		return ErrDimension
	}
	if vector.stopped() {
		return ErrCancelled
	}
	for level := size; level < vector.size; level++ {
		vector.entries -= len(vector.slots[level].values)
		clear(vector.slots[level].values)
		vector.slots[level].values = vector.slots[level].values[:0]
	}
	vector.size = size
	return nil
}

func (vector *IntLevelVector) Clear() error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	for level := 0; level < vector.size; level++ {
		vector.entries -= len(vector.slots[level].values)
		clear(vector.slots[level].values)
		vector.slots[level].values = vector.slots[level].values[:0]
	}
	vector.size = 0
	return nil
}

func (vector *IntLevelVector) Clone() (*IntLevelVector, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if vector.stopped() {
		return nil, ErrCancelled
	}
	clone, err := NewIntLevelVector(vector.size, vector.limits)
	if err != nil {
		return nil, err
	}
	for level := 0; level < vector.size; level++ {
		for _, entry := range vector.slots[level].values {
			if err = clone.Push(level, entry); err != nil {
				return nil, err
			}
		}
	}
	return clone, nil
}

type IntLevelRecord struct {
	Index  int
	Values []int
}

func (vector *IntLevelVector) Levels(start, stop int, reverse bool) ([]IntLevelRecord, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if start < 0 || stop < start || stop > vector.size {
		return nil, ErrDimension
	}
	result := make([]IntLevelRecord, 0, stop-start)
	if reverse {
		for level := stop - 1; level >= start; level-- {
			result = append(result, IntLevelRecord{level, append([]int(nil), vector.slots[level].values...)})
		}
		return result, nil
	}
	for level := start; level < stop; level++ {
		result = append(result, IntLevelRecord{level, append([]int(nil), vector.slots[level].values...)})
	}
	return result, nil
}

func (vector *IntLevelVector) LevelsByIndex(indices []int) ([]IntLevelRecord, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	result := make([]IntLevelRecord, len(indices))
	for index, level := range indices {
		if level < 0 || level >= vector.size {
			return nil, ErrDimension
		}
		result[index] = IntLevelRecord{level, append([]int(nil), vector.slots[level].values...)}
	}
	return result, nil
}

type IntLevelPair struct {
	FirstIndex  int
	SecondIndex int
	First       []int
	Second      []int
}

func (vector *IntLevelVector) LevelPairs() ([]IntLevelPair, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if vector.size&1 != 0 {
		return nil, ErrDimension
	}
	result := make([]IntLevelPair, 0, vector.size/2)
	for level := 0; level < vector.size; level += 2 {
		result = append(result, IntLevelPair{
			FirstIndex:  level,
			SecondIndex: level + 1,
			First:       append([]int(nil), vector.slots[level].values...),
			Second:      append([]int(nil), vector.slots[level+1].values...),
		})
	}
	return result, nil
}

type IntLevelZip struct {
	Index  int
	First  []int
	Second []int
}

func ZipIntLevelVectors(first, second *IntLevelVector) ([]IntLevelZip, error) {
	if err := first.check(); err != nil {
		return nil, err
	}
	if err := second.check(); err != nil {
		return nil, err
	}
	if second.size < first.size {
		return nil, ErrDimension
	}
	result := make([]IntLevelZip, first.size)
	for level := 0; level < first.size; level++ {
		result[level] = IntLevelZip{
			Index:  level,
			First:  append([]int(nil), first.slots[level].values...),
			Second: append([]int(nil), second.slots[level].values...),
		}
	}
	return result, nil
}

func (vector *IntLevelVector) SortBySize(reverse bool) error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	sort.SliceStable(vector.slots[:vector.size], func(first, second int) bool {
		if reverse {
			return len(vector.slots[first].values) > len(vector.slots[second].values)
		}
		return len(vector.slots[first].values) < len(vector.slots[second].values)
	})
	return nil
}

func (vector *IntLevelVector) sortByBoundary(first bool, reverse bool) error {
	if err := vector.check(); err != nil {
		return err
	}
	for level := 0; level < vector.size; level++ {
		if len(vector.slots[level].values) == 0 {
			return ErrDimension
		}
	}
	if vector.stopped() {
		return ErrCancelled
	}
	value := func(level int) int {
		if first {
			return vector.slots[level].values[0]
		}
		return vector.slots[level].values[len(vector.slots[level].values)-1]
	}
	sort.SliceStable(vector.slots[:vector.size], func(left, right int) bool {
		if reverse {
			return value(left) > value(right)
		}
		return value(left) < value(right)
	})
	return nil
}

func (vector *IntLevelVector) SortByFirst(reverse bool) error {
	return vector.sortByBoundary(true, reverse)
}

func (vector *IntLevelVector) SortByLast(reverse bool) error {
	return vector.sortByBoundary(false, reverse)
}

func (vector *IntLevelVector) Format(skipSingles bool) (string, error) {
	if err := vector.check(); err != nil {
		return "", err
	}
	var output strings.Builder
	for level := 0; level < vector.size; level++ {
		entries := vector.slots[level].values
		if skipSingles && len(entries) == 1 {
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

func (vector *IntLevelVector) FormatLiterals() (string, error) {
	if err := vector.check(); err != nil {
		return "", err
	}
	var output strings.Builder
	for level := 0; level < vector.size; level++ {
		entries := vector.slots[level].values
		fmt.Fprintf(&output, " %4d : %2d  {", level, len(entries))
		for _, literal := range entries {
			if literal < 0 {
				return "", ErrInvalidLiteral
			}
			sign := byte('+')
			if literal&1 != 0 {
				sign = '-'
			}
			fmt.Fprintf(&output, " %c%d", sign, literal>>1)
		}
		output.WriteString(" }\n")
	}
	return output.String(), nil
}

func NewIntEquivalenceClasses(mapping []int, limits IntLevelVectorLimits) (*IntLevelVector, error) {
	maximum := 0
	for _, class := range mapping {
		if class < 0 {
			return nil, ErrDimension
		}
		if class > maximum {
			maximum = class
		}
	}
	classes, err := NewIntLevelVectorWithLevels(maximum+1, limits)
	if err != nil {
		return nil, err
	}
	for index, class := range mapping {
		if err = classes.Push(class, index); err != nil {
			return nil, err
		}
	}
	return classes, nil
}

func (vector *IntLevelVector) NonTrivialCounts() (classes, entries int, err error) {
	if err = vector.check(); err != nil {
		return 0, 0, err
	}
	for level := 0; level < vector.size; level++ {
		size := len(vector.slots[level].values)
		if size >= 2 {
			classes++
			entries += size
		}
	}
	return classes, entries, nil
}

func (vector *IntLevelVector) FirstEntries() ([]int, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	firsts := make([]int, 0, vector.size)
	for level := 0; level < vector.size; level++ {
		if len(vector.slots[level].values) != 0 {
			firsts = append(firsts, vector.slots[level].values[0])
		}
	}
	return firsts, nil
}

func (vector *IntLevelVector) LevelMarked(level int) (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	if level < 0 || level >= vector.size {
		return false, ErrDimension
	}
	return vector.slots[level].marked, nil
}

func (vector *IntLevelVector) SetLevelMark(level int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if level < 0 || level >= vector.size {
		return ErrDimension
	}
	vector.slots[level].marked = true
	return nil
}

func (vector *IntLevelVector) ToggleLevelMark(level int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if level < 0 || level >= vector.size {
		return ErrDimension
	}
	vector.slots[level].marked = !vector.slots[level].marked
	return nil
}

func (vector *IntLevelVector) changeLevelMarks(levels []int, mark bool) error {
	if err := vector.check(); err != nil {
		return err
	}
	seen := make(map[int]struct{}, len(levels))
	for _, level := range levels {
		if level < 0 || level >= vector.size {
			return ErrDimension
		}
		if _, exists := seen[level]; exists || vector.slots[level].marked != !mark {
			return ErrInvariant
		}
		seen[level] = struct{}{}
	}
	if vector.stopped() {
		return ErrCancelled
	}
	for _, level := range levels {
		vector.slots[level].marked = mark
	}
	return nil
}

func (vector *IntLevelVector) MarkLevels(levels []int) error {
	return vector.changeLevelMarks(levels, true)
}

func (vector *IntLevelVector) UnmarkLevels(levels []int) error {
	return vector.changeLevelMarks(levels, false)
}

func (vector *IntLevelVector) RemoveEmpty() error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	write := 0
	oldSize := vector.size
	for read := 0; read < oldSize; read++ {
		slot := vector.slots[read]
		if len(slot.values) == 0 {
			vector.storage -= cap(slot.values)
			vector.slots[read] = intLevelSlot{}
			continue
		}
		if write != read {
			vector.slots[write] = slot
			vector.slots[read] = intLevelSlot{}
		}
		write++
	}
	for level := write; level < oldSize; level++ {
		vector.slots[level] = intLevelSlot{}
	}
	vector.size = write
	return nil
}

func (vector *IntLevelVector) NativeMemoryBytes() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	pointerBytes := int(unsafe.Sizeof(uintptr(0)))
	headerBytes := 2*4 + pointerBytes
	if len(vector.slots) > math.MaxInt/headerBytes {
		return 0, ErrLimit
	}
	memory := len(vector.slots) * headerBytes
	for level := 0; level < vector.size; level++ {
		capacity := cap(vector.slots[level].values)
		if capacity > (math.MaxInt-memory)/4 {
			return 0, ErrLimit
		}
		memory += capacity * 4
	}
	return memory, nil
}

func (vector *IntLevelVector) Validate() error {
	return vector.check()
}

func (vector *IntLevelVector) Close() error {
	if err := vector.check(); err != nil {
		return err
	}
	for index := range vector.slots {
		clear(vector.slots[index].values)
		vector.slots[index] = intLevelSlot{}
	}
	vector.slots = nil
	vector.size = 0
	vector.entries = 0
	vector.storage = 0
	vector.closed = true
	return nil
}
