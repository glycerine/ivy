package abc

// Unsigned word vectors from src/misc/vec/vecWrd.h.

import (
	"encoding/binary"
	"math"
	"os"
	"slices"
	"sort"
	"strings"
	"unsafe"
)

type WordVectorLimits struct {
	MaxWords int
	Stop     func() bool
}

func normalizeWordVectorLimits(limits WordVectorLimits) (WordVectorLimits, error) {
	if limits.MaxWords < 0 {
		return limits, ErrDimension
	}
	if limits.MaxWords == 0 {
		limits.MaxWords = 1 << 26
	}
	return limits, nil
}

type WordVector struct {
	values []uint64
	limits WordVectorLimits
	closed bool
}

func newWordVector(capacity int, exact bool, limits WordVectorLimits) (*WordVector, error) {
	limits, err := normalizeWordVectorLimits(limits)
	if err != nil || capacity < 0 {
		if err != nil {
			return nil, err
		}
		return nil, ErrDimension
	}
	if !exact && capacity > 0 && capacity < 16 {
		capacity = 16
	}
	if capacity > limits.MaxWords {
		return nil, ErrLimit
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	return &WordVector{values: make([]uint64, 0, capacity), limits: limits}, nil
}

func NewWordVector(capacity int, limits WordVectorLimits) (*WordVector, error) {
	return newWordVector(capacity, false, limits)
}

func NewExactWordVector(capacity int, limits WordVectorLimits) (*WordVector, error) {
	return newWordVector(capacity, true, limits)
}

func NewWordVectorFrom(values []uint64, limits WordVectorLimits) (*WordVector, error) {
	vector, err := NewExactWordVector(len(values), limits)
	if err != nil {
		return nil, err
	}
	vector.values = append(vector.values, values...)
	return vector, nil
}

func NewFilledWordVector(size int, value uint64, limits WordVectorLimits) (*WordVector, error) {
	vector, err := NewWordVector(size, limits)
	if err != nil {
		return nil, err
	}
	vector.values = vector.values[:size]
	for index := range vector.values {
		vector.values[index] = value
	}
	return vector, nil
}

func NewNaturalWordVector(size int, limits WordVectorLimits) (*WordVector, error) {
	vector, err := NewFilledWordVector(size, 0, limits)
	if err != nil {
		return nil, err
	}
	for index := range vector.values {
		vector.values[index] = uint64(index)
	}
	return vector, nil
}

func NewRandomWordVector(size int, random *ABCRandom, limits WordVectorLimits) (*WordVector, error) {
	if random == nil {
		return nil, ErrDimension
	}
	vector, err := NewFilledWordVector(size, 0, limits)
	if err != nil {
		return nil, err
	}
	for index := range vector.values {
		if vector.stopped() {
			return nil, ErrCancelled
		}
		vector.values[index] = random.Uint64(false)
	}
	return vector, nil
}

func NewElementaryTruthWordVector(variables int, limits WordVectorLimits) (*WordVector, error) {
	if variables < 0 || variables > 62 {
		return nil, ErrDimension
	}
	limits, err := normalizeWordVectorLimits(limits)
	if err != nil {
		return nil, err
	}
	words := 1
	if variables > 6 {
		wordCount := uint64(1) << uint(variables-6)
		if wordCount > uint64(limits.MaxWords/variables) {
			return nil, ErrLimit
		}
		words = int(wordCount)
	}
	if variables != 0 && words > limits.MaxWords/variables {
		return nil, ErrLimit
	}
	vector, err := NewFilledWordVector(words*variables, 0, limits)
	if err != nil {
		return nil, err
	}
	masks := [...]uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00, 0xffff0000}
	for variable := 0; variable < variables; variable++ {
		for word := 0; word < words; word++ {
			if variable < 5 {
				mask := uint64(masks[variable])
				vector.values[variable*words+word] = mask<<32 | mask
				continue
			}
			bit := 1 << uint(variable-5)
			var value uint64
			if 2*word&bit != 0 {
				value |= math.MaxUint32
			}
			if (2*word+1)&bit != 0 {
				value |= uint64(math.MaxUint32) << 32
			}
			vector.values[variable*words+word] = value
		}
	}
	return vector, nil
}

func (vector *WordVector) check() error {
	if vector == nil {
		return ErrDimension
	}
	if vector.closed || len(vector.values) > cap(vector.values) || cap(vector.values) > vector.limits.MaxWords {
		return ErrInvariant
	}
	return nil
}

func (vector *WordVector) stopped() bool { return vector.limits.Stop != nil && vector.limits.Stop() }

func (vector *WordVector) growUnchecked(capacity int) {
	if capacity <= cap(vector.values) {
		return
	}
	values := make([]uint64, len(vector.values), capacity)
	copy(values, vector.values)
	vector.values = values
}

func (vector *WordVector) Grow(capacity int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if capacity < 0 {
		return ErrDimension
	}
	if capacity <= cap(vector.values) {
		return nil
	}
	if capacity > vector.limits.MaxWords {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	return nil
}

func (vector *WordVector) Len() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return len(vector.values), nil
}
func (vector *WordVector) Capacity() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	return cap(vector.values), nil
}
func (vector *WordVector) Snapshot() ([]uint64, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	return append([]uint64(nil), vector.values...), nil
}

func (vector *WordVector) Entries(start, stop int, reverse bool) ([]uint64, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	if start < 0 || stop < start || stop > len(vector.values) {
		return nil, ErrDimension
	}
	result := make([]uint64, 0, stop-start)
	if reverse {
		for index := stop - 1; index >= start; index-- {
			result = append(result, vector.values[index])
		}
		return result, nil
	}
	return append(result, vector.values[start:stop]...), nil
}

func (vector *WordVector) Entry(index int) (uint64, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if index < 0 || index >= len(vector.values) {
		return 0, ErrDimension
	}
	return vector.values[index], nil
}

func (vector *WordVector) Last() (uint64, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if len(vector.values) == 0 {
		return 0, ErrDimension
	}
	return vector.values[len(vector.values)-1], nil
}

func (vector *WordVector) Write(index int, value uint64) error {
	if err := vector.check(); err != nil {
		return err
	}
	if index < 0 || index >= len(vector.values) {
		return ErrDimension
	}
	vector.values[index] = value
	return nil
}

func (vector *WordVector) Add(index int, addition uint64) (uint64, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if index < 0 || index >= len(vector.values) {
		return 0, ErrDimension
	}
	vector.values[index] += addition
	return vector.values[index], nil
}

func (vector *WordVector) nextPushCapacity(required int) (int, error) {
	capacity := cap(vector.values)
	if required <= capacity {
		return capacity, nil
	}
	if capacity < 16 {
		capacity = 16
	} else {
		if capacity > math.MaxInt/2 {
			return 0, ErrLimit
		}
		capacity *= 2
	}
	if capacity < required {
		capacity = required
	}
	if capacity > vector.limits.MaxWords {
		return 0, ErrLimit
	}
	return capacity, nil
}

func (vector *WordVector) Push(value uint64) error {
	if err := vector.check(); err != nil {
		return err
	}
	capacity, err := vector.nextPushCapacity(len(vector.values) + 1)
	if err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	vector.values = append(vector.values, value)
	return nil
}

func (vector *WordVector) PushFirst(value uint64) error {
	if err := vector.check(); err != nil {
		return err
	}
	capacity, err := vector.nextPushCapacity(len(vector.values) + 1)
	if err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	vector.values = append(vector.values, 0)
	copy(vector.values[1:], vector.values[:len(vector.values)-1])
	vector.values[0] = value
	return nil
}

func (vector *WordVector) PushOrdered(value uint64) error {
	if err := vector.check(); err != nil {
		return err
	}
	capacity, err := vector.nextPushCapacity(len(vector.values) + 1)
	if err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	index := sort.Search(len(vector.values), func(index int) bool { return vector.values[index] > value })
	vector.growUnchecked(capacity)
	vector.values = append(vector.values, 0)
	copy(vector.values[index+1:], vector.values[index:len(vector.values)-1])
	vector.values[index] = value
	return nil
}

func (vector *WordVector) PushUnique(value uint64) (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	if vector.Find(value) >= 0 {
		return false, nil
	}
	if err := vector.Push(value); err != nil {
		return false, err
	}
	return true, nil
}

func (vector *WordVector) PushUniqueOrdered(value uint64) (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	if vector.Find(value) >= 0 {
		return false, nil
	}
	if err := vector.PushOrdered(value); err != nil {
		return false, err
	}
	return true, nil
}

func (vector *WordVector) PushSlice(values []uint64) error {
	if err := vector.check(); err != nil {
		return err
	}
	if len(values) > vector.limits.MaxWords-len(vector.values) {
		return ErrLimit
	}
	required, capacity := len(vector.values)+len(values), cap(vector.values)
	for capacity < required {
		if capacity < 16 {
			capacity = 16
		} else {
			if capacity > math.MaxInt/2 {
				return ErrLimit
			}
			capacity *= 2
		}
		if capacity > vector.limits.MaxWords {
			return ErrLimit
		}
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	vector.values = append(vector.values, values...)
	return nil
}

// Fetch reserves existing capacity and returns its starting index. The region
// is zeroed instead of exposing native uninitialized or stale words.
func (vector *WordVector) Fetch(count int) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if count < 0 {
		return 0, ErrDimension
	}
	if count == 0 {
		return -1, nil
	}
	if count > cap(vector.values)-len(vector.values) {
		return 0, ErrLimit
	}
	if vector.stopped() {
		return 0, ErrCancelled
	}
	start := len(vector.values)
	vector.values = vector.values[:start+count]
	clear(vector.values[start:])
	return start, nil
}

func (vector *WordVector) Pop() (uint64, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if len(vector.values) == 0 {
		return 0, ErrDimension
	}
	last := len(vector.values) - 1
	value := vector.values[last]
	vector.values[last] = 0
	vector.values = vector.values[:last]
	return value, nil
}

func (vector *WordVector) Find(value uint64) int {
	if vector == nil || vector.closed {
		return -1
	}
	for index, entry := range vector.values {
		if entry == value {
			return index
		}
	}
	return -1
}

func (vector *WordVector) Remove(value uint64) (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	index := vector.Find(value)
	if index < 0 {
		return false, nil
	}
	copy(vector.values[index:], vector.values[index+1:])
	vector.values[len(vector.values)-1] = 0
	vector.values = vector.values[:len(vector.values)-1]
	return true, nil
}

func (vector *WordVector) Insert(index int, value uint64) error {
	if err := vector.check(); err != nil {
		return err
	}
	if index < 0 || index >= len(vector.values) {
		return ErrDimension
	}
	capacity, err := vector.nextPushCapacity(len(vector.values) + 1)
	if err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(capacity)
	vector.values = append(vector.values, 0)
	copy(vector.values[index+1:], vector.values[index:len(vector.values)-1])
	vector.values[index] = value
	return nil
}

func (vector *WordVector) Fill(size int, value uint64) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 {
		return ErrDimension
	}
	if size > vector.limits.MaxWords {
		return ErrLimit
	}
	if vector.stopped() {
		return ErrCancelled
	}
	vector.growUnchecked(size)
	vector.values = vector.values[:size]
	for index := range vector.values {
		vector.values[index] = value
	}
	return nil
}

func (vector *WordVector) FillExtra(size int, value uint64) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 {
		return ErrDimension
	}
	if size <= len(vector.values) {
		return nil
	}
	if size > vector.limits.MaxWords {
		return ErrLimit
	}
	capacity := cap(vector.values)
	if capacity > math.MaxInt/2 || size > 2*capacity {
		capacity = size
	} else if size > capacity {
		capacity *= 2
	}
	if capacity > vector.limits.MaxWords {
		capacity = size
	}
	if vector.stopped() {
		return ErrCancelled
	}
	oldSize := len(vector.values)
	vector.growUnchecked(capacity)
	vector.values = vector.values[:size]
	for index := oldSize; index < size; index++ {
		vector.values[index] = value
	}
	return nil
}

func (vector *WordVector) Get(index int) (uint64, error) {
	if index < 0 {
		return 0, ErrDimension
	}
	if err := vector.FillExtra(index+1, 0); err != nil {
		return 0, err
	}
	return vector.values[index], nil
}

func (vector *WordVector) Set(index int, value uint64) error {
	if index < 0 {
		return ErrDimension
	}
	if err := vector.FillExtra(index+1, 0); err != nil {
		return err
	}
	vector.values[index] = value
	return nil
}

func (vector *WordVector) Shrink(size int) error {
	if err := vector.check(); err != nil {
		return err
	}
	if size < 0 || size > len(vector.values) {
		return ErrDimension
	}
	clear(vector.values[size:])
	vector.values = vector.values[:size]
	return nil
}

func (vector *WordVector) Clear() error { return vector.Shrink(0) }

func (vector *WordVector) Reverse() error {
	if err := vector.check(); err != nil {
		return err
	}
	for left, right := 0, len(vector.values)-1; left < right; left, right = left+1, right-1 {
		vector.values[left], vector.values[right] = vector.values[right], vector.values[left]
	}
	return nil
}

func (vector *WordVector) Maximum() (uint64, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if len(vector.values) == 0 {
		return 0, nil
	}
	maximum := vector.values[0]
	for _, value := range vector.values[1:] {
		if value > maximum {
			maximum = value
		}
	}
	return maximum, nil
}

func (vector *WordVector) Minimum() (uint64, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if len(vector.values) == 0 {
		return 0, nil
	}
	minimum := vector.values[0]
	for _, value := range vector.values[1:] {
		if value < minimum {
			minimum = value
		}
	}
	return minimum, nil
}

func (vector *WordVector) Invert(fill uint64) (*WordVector, error) {
	maximum, err := vector.Maximum()
	if err != nil {
		return nil, err
	}
	if maximum >= uint64(vector.limits.MaxWords) {
		return nil, ErrLimit
	}
	result, err := NewFilledWordVector(int(maximum)+1, fill, vector.limits)
	if err != nil {
		return nil, err
	}
	for index, value := range vector.values {
		if value != fill {
			result.values[int(value)] = uint64(index)
		}
	}
	return result, nil
}

func (vector *WordVector) Sum() (uint64, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	var sum uint64
	for _, value := range vector.values {
		sum += value
	}
	return sum, nil
}
func (vector *WordVector) CountZero() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	count := 0
	for _, value := range vector.values {
		if value == 0 {
			count++
		}
	}
	return count, nil
}

func (vector *WordVector) Equal(other *WordVector) (bool, error) {
	if err := vector.check(); err != nil {
		return false, err
	}
	if err := other.check(); err != nil {
		return false, err
	}
	if len(vector.values) != len(other.values) {
		return false, nil
	}
	for index := range vector.values {
		if vector.values[index] != other.values[index] {
			return false, nil
		}
	}
	return true, nil
}

func (vector *WordVector) CountCommon(other *WordVector) (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	if err := other.check(); err != nil {
		return 0, err
	}
	first, second := vector.values, other.values
	if len(first) < len(second) {
		first, second = second, first
	}
	present := make(map[uint64]struct{}, len(second))
	for _, value := range second {
		if value != math.MaxUint64 {
			present[value] = struct{}{}
		}
	}
	count := 0
	for _, value := range first {
		if _, ok := present[value]; ok {
			count++
		}
	}
	return count, nil
}

func (vector *WordVector) Sort(reverse bool) error {
	if err := vector.check(); err != nil {
		return err
	}
	if vector.stopped() {
		return ErrCancelled
	}
	slices.Sort(vector.values)
	if reverse {
		slices.Reverse(vector.values)
	}
	return nil
}

func (vector *WordVector) Uniqify() error {
	if err := vector.Sort(false); err != nil {
		return err
	}
	if len(vector.values) < 2 {
		return nil
	}
	write := 1
	for read := 1; read < len(vector.values); read++ {
		if vector.values[read] != vector.values[read-1] {
			vector.values[write] = vector.values[read]
			write++
		}
	}
	clear(vector.values[write:])
	vector.values = vector.values[:write]
	return nil
}

func wordRecordKey(values []uint64) string {
	data := make([]byte, 8*len(values))
	for index, value := range values {
		binary.LittleEndian.PutUint64(data[8*index:], value)
	}
	return string(data)
}

func (vector *WordVector) UniqueRecords(width int) (int, []int, error) {
	if err := vector.check(); err != nil {
		return 0, nil, err
	}
	if width <= 0 || len(vector.values)%width != 0 {
		return 0, nil, ErrDimension
	}
	first := make(map[string]int, len(vector.values)/width)
	mapping := make([]int, len(vector.values)/width)
	unique := 0
	for record := range mapping {
		key := wordRecordKey(vector.values[record*width : (record+1)*width])
		index, found := first[key]
		if !found {
			index = record
			first[key] = record
			unique++
		}
		mapping[record] = index
	}
	return unique, mapping, nil
}

func (vector *WordVector) UniqifyRecords(width int) (*WordVector, error) {
	unique, mapping, err := vector.UniqueRecords(width)
	if err != nil {
		return nil, err
	}
	capacity := unique * width
	if capacity > 0 && capacity < 8 {
		capacity = 8
	}
	result, err := NewExactWordVector(capacity, vector.limits)
	if err != nil {
		return nil, err
	}
	for record, first := range mapping {
		if record == first {
			result.values = append(result.values, vector.values[record*width:(record+1)*width]...)
		}
	}
	return result, nil
}

func (vector *WordVector) Append(other *WordVector) error {
	if err := other.check(); err != nil {
		return err
	}
	return vector.PushSlice(other.values)
}
func (vector *WordVector) Clone() (*WordVector, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	return NewWordVectorFrom(vector.values, vector.limits)
}

func (vector *WordVector) Move() (*WordVector, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	moved := &WordVector{values: vector.values, limits: vector.limits}
	vector.values = nil
	return moved, nil
}

func (vector *WordVector) Release() ([]uint64, error) {
	if err := vector.check(); err != nil {
		return nil, err
	}
	values := vector.values
	vector.values = nil
	return values, nil
}

func (vector *WordVector) NativeMemoryBytes() (int, error) {
	if err := vector.check(); err != nil {
		return 0, err
	}
	header := 2*4 + int(unsafe.Sizeof(uintptr(0)))
	if cap(vector.values) > (math.MaxInt-header)/8 {
		return 0, ErrLimit
	}
	return header + 8*cap(vector.values), nil
}

func formatWordHexRecord(words []uint64) string {
	const digits = "0123456789ABCDEF"
	var output strings.Builder
	output.Grow(16*len(words) + 1)
	for _, word := range words {
		for nibble := 0; nibble < 16; nibble++ {
			output.WriteByte(digits[(word>>uint(4*nibble))&15])
		}
	}
	output.WriteByte('\n')
	return output.String()
}

func (vector *WordVector) FormatHex(wordsPerRecord int) (string, error) {
	if err := vector.check(); err != nil {
		return "", err
	}
	if wordsPerRecord <= 0 || len(vector.values)%wordsPerRecord != 0 {
		return "", ErrDimension
	}
	var output strings.Builder
	for start := 0; start < len(vector.values); start += wordsPerRecord {
		output.WriteString(formatWordHexRecord(vector.values[start : start+wordsPerRecord]))
	}
	return output.String(), nil
}

func (vector *WordVector) DumpHexFile(path string, wordsPerRecord int) (int, error) {
	text, err := vector.FormatHex(wordsPerRecord)
	if err != nil {
		return 0, err
	}
	if err = os.WriteFile(path, []byte(text), 0o666); err != nil {
		return 0, err
	}
	return len(vector.values) / wordsPerRecord, nil
}

func wordHexDigit(character byte) (uint64, bool) {
	switch {
	case character >= '0' && character <= '9':
		return uint64(character - '0'), true
	case character >= 'A' && character <= 'F':
		return uint64(character-'A') + 10, true
	case character >= 'a' && character <= 'f':
		return uint64(character-'a') + 10, true
	default:
		return 0, false
	}
}

func ParseWordVectorHex(data []byte, limits WordVectorLimits) (*WordVector, int, error) {
	limits, err := normalizeWordVectorLimits(limits)
	if err != nil {
		return nil, 0, err
	}
	initialCapacity := 1000
	if initialCapacity > limits.MaxWords {
		initialCapacity = limits.MaxWords
	}
	vector, err := NewExactWordVector(initialCapacity, limits)
	if err != nil {
		return nil, 0, err
	}
	wordsPerRecord := -1
	lineWords := make([]uint64, 0)
	var word uint64
	nibbles := 0
	finishWord := func() { lineWords = append(lineWords, word); word, nibbles = 0, 0 }
	finishLine := func() error {
		if nibbles != 0 {
			finishWord()
		}
		if len(lineWords) == 0 {
			return nil
		}
		if wordsPerRecord < 0 {
			wordsPerRecord = len(lineWords)
		} else if len(lineWords) != wordsPerRecord {
			return ErrDimension
		}
		if err := vector.PushSlice(lineWords); err != nil {
			return err
		}
		lineWords = lineWords[:0]
		return nil
	}
	for _, character := range data {
		if character == '\n' {
			if err := finishLine(); err != nil {
				return nil, 0, err
			}
			continue
		}
		if character == '\r' || character == '\t' || character == ' ' {
			continue
		}
		digit, ok := wordHexDigit(character)
		if !ok {
			return nil, 0, ErrDimension
		}
		word |= digit << uint(4*nibbles)
		nibbles++
		if nibbles == 16 {
			finishWord()
		}
	}
	if err := finishLine(); err != nil {
		return nil, 0, err
	}
	if wordsPerRecord <= 0 {
		return nil, 0, ErrDimension
	}
	return vector, wordsPerRecord, nil
}

func ReadWordVectorHexFile(path string, limits WordVectorLimits) (*WordVector, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	return ParseWordVectorHex(data, limits)
}

func (vector *WordVector) Validate() error { return vector.check() }

func (vector *WordVector) Close() error {
	if err := vector.check(); err != nil {
		return err
	}
	clear(vector.values)
	vector.values = nil
	vector.closed = true
	return nil
}
