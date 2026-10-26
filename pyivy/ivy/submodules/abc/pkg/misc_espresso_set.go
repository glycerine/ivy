package abc

// Packed sets and set families from src/misc/espresso/set.c.

import (
	"bytes"
	"math/bits"
	"slices"
	"strconv"
	"strings"
)

type EspressoSetFlags uint32

const (
	EspressoSetPrime EspressoSetFlags = 0x8000 >> iota
	EspressoSetNonessential
	EspressoSetActive
	EspressoSetRedundant
	EspressoSetCovered
	EspressoSetRelEssential
)

const espressoSetMaximumBits = 1023 * 32

type EspressoSet struct {
	bits  int
	words []uint32
	Rank  int
	Flags EspressoSetFlags
}

func espressoSetWordCount(bitCount int) int { return max(1, (bitCount+31)/32) }

func NewEspressoSet(bitCount int) (*EspressoSet, error) {
	if bitCount < 0 {
		return nil, ErrDimension
	}
	if bitCount > espressoSetMaximumBits {
		return nil, ErrLimit
	}
	return &EspressoSet{bits: bitCount, words: make([]uint32, espressoSetWordCount(bitCount))}, nil
}

func NewFullEspressoSet(bitCount int) (*EspressoSet, error) {
	result, err := NewEspressoSet(bitCount)
	if err != nil {
		return nil, err
	}
	result.Fill()
	return result, nil
}

func (set *EspressoSet) check() error {
	if set == nil {
		return ErrDimension
	}
	if set.bits < 0 || set.bits > espressoSetMaximumBits || len(set.words) != espressoSetWordCount(set.bits) {
		return ErrInvariant
	}
	if set.Rank < 0 || set.Rank > 0xffff || set.Flags&^EspressoSetFlags(0xfc00) != 0 {
		return ErrInvariant
	}
	if set.bits&31 != 0 && set.words[len(set.words)-1]>>uint(set.bits&31) != 0 {
		return ErrInvariant
	}
	if set.bits == 0 && set.words[0] != 0 {
		return ErrInvariant
	}
	return nil
}

func (set *EspressoSet) BitCount() int {
	if set == nil {
		return 0
	}
	return set.bits
}

func (set *EspressoSet) Words() ([]uint32, error) {
	if err := set.check(); err != nil {
		return nil, err
	}
	return slices.Clone(set.words), nil
}

func (set *EspressoSet) Clone() (*EspressoSet, error) {
	if err := set.check(); err != nil {
		return nil, err
	}
	return &EspressoSet{bits: set.bits, words: slices.Clone(set.words), Rank: set.Rank, Flags: set.Flags}, nil
}

func (set *EspressoSet) Clear() error {
	if err := set.check(); err != nil {
		return err
	}
	clear(set.words)
	return nil
}

func (set *EspressoSet) Fill() error {
	if err := set.check(); err != nil {
		return err
	}
	for index := range set.words {
		set.words[index] = ^uint32(0)
	}
	if set.bits == 0 {
		set.words[0] = 0
	} else if set.bits&31 != 0 {
		set.words[len(set.words)-1] &= uint32(1)<<uint(set.bits&31) - 1
	}
	return nil
}

func (set *EspressoSet) Has(element int) (bool, error) {
	if err := set.check(); err != nil {
		return false, err
	}
	if element < 0 || element >= set.bits {
		return false, ErrDimension
	}
	return set.words[element>>5]&(uint32(1)<<uint(element&31)) != 0, nil
}

func (set *EspressoSet) Set(element int, present bool) error {
	if err := set.check(); err != nil {
		return err
	}
	if element < 0 || element >= set.bits {
		return ErrDimension
	}
	mask := uint32(1) << uint(element&31)
	if present {
		set.words[element>>5] |= mask
	} else {
		set.words[element>>5] &^= mask
	}
	return nil
}

func (set *EspressoSet) Cardinality() (int, error) {
	if err := set.check(); err != nil {
		return 0, err
	}
	count := 0
	for _, word := range set.words {
		count += bits.OnesCount32(word)
	}
	return count, nil
}

func (set *EspressoSet) Elements() ([]int, error) {
	if err := set.check(); err != nil {
		return nil, err
	}
	result := make([]int, 0)
	for wordIndex, word := range set.words {
		for word != 0 {
			bit := bits.TrailingZeros32(word)
			result = append(result, wordIndex*32+bit)
			word &^= uint32(1) << uint(bit)
		}
	}
	return result, nil
}

func espressoSetPair(first, second *EspressoSet) error {
	if err := first.check(); err != nil {
		return err
	}
	if err := second.check(); err != nil {
		return err
	}
	if first.bits != second.bits {
		return ErrDimension
	}
	return nil
}

func espressoSetBinary(first, second *EspressoSet, operation byte) (*EspressoSet, error) {
	if err := espressoSetPair(first, second); err != nil {
		return nil, err
	}
	result, _ := NewEspressoSet(first.bits)
	for index := range result.words {
		switch operation {
		case '&':
			result.words[index] = first.words[index] & second.words[index]
		case '|':
			result.words[index] = first.words[index] | second.words[index]
		case '-':
			result.words[index] = first.words[index] &^ second.words[index]
		case '^':
			result.words[index] = first.words[index] ^ second.words[index]
		}
	}
	return result, nil
}

func (set *EspressoSet) Intersection(other *EspressoSet) (*EspressoSet, error) {
	return espressoSetBinary(set, other, '&')
}

func (set *EspressoSet) IntersectionNonempty(other *EspressoSet) (*EspressoSet, bool, error) {
	result, err := set.Intersection(other)
	if err != nil {
		return nil, false, err
	}
	empty, _ := result.Empty()
	return result, !empty, nil
}

func (set *EspressoSet) Union(other *EspressoSet) (*EspressoSet, error) {
	return espressoSetBinary(set, other, '|')
}

func (set *EspressoSet) UnionNonempty(other *EspressoSet) (*EspressoSet, bool, error) {
	result, err := set.Union(other)
	if err != nil {
		return nil, false, err
	}
	empty, _ := result.Empty()
	return result, !empty, nil
}

func (set *EspressoSet) Difference(other *EspressoSet) (*EspressoSet, error) {
	return espressoSetBinary(set, other, '-')
}

func (set *EspressoSet) SymmetricDifference(other *EspressoSet) (*EspressoSet, error) {
	return espressoSetBinary(set, other, '^')
}

func (set *EspressoSet) Merge(other, mask *EspressoSet) (*EspressoSet, error) {
	if err := espressoSetPair(set, other); err != nil {
		return nil, err
	}
	if err := espressoSetPair(set, mask); err != nil {
		return nil, err
	}
	result, _ := NewEspressoSet(set.bits)
	for index := range result.words {
		result.words[index] = set.words[index]&mask.words[index] | other.words[index]&^mask.words[index]
	}
	return result, nil
}

func (set *EspressoSet) CommonCount(other *EspressoSet) (int, error) {
	if err := espressoSetPair(set, other); err != nil {
		return 0, err
	}
	count := 0
	for index, word := range set.words {
		count += bits.OnesCount32(word & other.words[index])
	}
	return count, nil
}

func (set *EspressoSet) Empty() (bool, error) {
	count, err := set.Cardinality()
	return count == 0, err
}

func (set *EspressoSet) Full() (bool, error) {
	count, err := set.Cardinality()
	return count == set.bits, err
}

func (set *EspressoSet) Equal(other *EspressoSet) (bool, error) {
	if err := espressoSetPair(set, other); err != nil {
		return false, err
	}
	return slices.Equal(set.words, other.words), nil
}

func (set *EspressoSet) Disjoint(other *EspressoSet) (bool, error) {
	if err := espressoSetPair(set, other); err != nil {
		return false, err
	}
	for index, word := range set.words {
		if word&other.words[index] != 0 {
			return false, nil
		}
	}
	return true, nil
}

func (set *EspressoSet) Implies(other *EspressoSet) (bool, error) {
	if err := espressoSetPair(set, other); err != nil {
		return false, err
	}
	for index, word := range set.words {
		if word&^other.words[index] != 0 {
			return false, nil
		}
	}
	return true, nil
}

func (set *EspressoSet) String() string {
	elements, err := set.Elements()
	if err != nil {
		return "[]"
	}
	var output strings.Builder
	output.WriteByte('[')
	for index, element := range elements {
		if index != 0 {
			output.WriteByte(',')
		}
		output.WriteString(strconv.Itoa(element))
	}
	output.WriteByte(']')
	return output.String()
}

func (set *EspressoSet) BitString() string {
	if set.check() != nil {
		return ""
	}
	result := make([]byte, set.bits)
	for bit := range result {
		if set.words[bit>>5]&(uint32(1)<<uint(bit&31)) != 0 {
			result[bit] = '1'
		} else {
			result[bit] = '0'
		}
	}
	return string(result)
}

func EspressoBitIndex(word uint32) int {
	if word == 0 {
		return -1
	}
	return bits.TrailingZeros32(word)
}

type EspressoSetFamily struct {
	bits        int
	sets        []*EspressoSet
	activeCount int
	closed      bool
}

func NewEspressoSetFamily(capacity, bitCount int) (*EspressoSetFamily, error) {
	if capacity < 0 || bitCount < 0 {
		return nil, ErrDimension
	}
	if bitCount > espressoSetMaximumBits {
		return nil, ErrLimit
	}
	return &EspressoSetFamily{bits: bitCount, sets: make([]*EspressoSet, 0, capacity)}, nil
}

func (family *EspressoSetFamily) check() error {
	if family == nil {
		return ErrDimension
	}
	if family.closed || family.bits < 0 || family.bits > espressoSetMaximumBits || family.activeCount < 0 || family.activeCount > len(family.sets) {
		return ErrInvariant
	}
	active := 0
	for _, set := range family.sets {
		if err := set.check(); err != nil || set.bits != family.bits {
			return ErrInvariant
		}
		if set.Flags&EspressoSetActive != 0 {
			active++
		}
	}
	if active != family.activeCount {
		return ErrInvariant
	}
	return nil
}

func (family *EspressoSetFamily) BitCount() int {
	if family == nil {
		return 0
	}
	return family.bits
}

func (family *EspressoSetFamily) Counts() (sets, active int, err error) {
	if err = family.check(); err != nil {
		return 0, 0, err
	}
	return len(family.sets), family.activeCount, nil
}

func (family *EspressoSetFamily) Close() error {
	if err := family.check(); err != nil {
		return err
	}
	family.sets = nil
	family.activeCount = 0
	family.closed = true
	return nil
}

func (family *EspressoSetFamily) Clone() (*EspressoSetFamily, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	result, _ := NewEspressoSetFamily(len(family.sets), family.bits)
	for _, set := range family.sets {
		clone, _ := set.Clone()
		result.sets = append(result.sets, clone)
	}
	result.activeCount = family.activeCount
	return result, nil
}

func (family *EspressoSetFamily) Add(set *EspressoSet) error {
	if err := family.check(); err != nil {
		return err
	}
	if err := set.check(); err != nil {
		return err
	}
	if set.bits != family.bits {
		return ErrDimension
	}
	clone, _ := set.Clone()
	family.sets = append(family.sets, clone)
	if clone.Flags&EspressoSetActive != 0 {
		family.activeCount++
	}
	return nil
}

func (family *EspressoSetFamily) Set(index int) (*EspressoSet, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	if index < 0 || index >= len(family.sets) {
		return nil, ErrDimension
	}
	return family.sets[index].Clone()
}

func (family *EspressoSetFamily) Delete(index int) error {
	if err := family.check(); err != nil {
		return err
	}
	if index < 0 || index >= len(family.sets) {
		return ErrDimension
	}
	if family.sets[index].Flags&EspressoSetActive != 0 {
		family.activeCount--
	}
	last := len(family.sets) - 1
	if index != last {
		family.sets[index] = family.sets[last]
	}
	family.sets = family.sets[:last]
	return nil
}

func (family *EspressoSetFamily) ActivateAll() error {
	if err := family.check(); err != nil {
		return err
	}
	for _, set := range family.sets {
		set.Flags |= EspressoSetActive
	}
	family.activeCount = len(family.sets)
	return nil
}

func (family *EspressoSetFamily) SetActive(index int, active bool) error {
	if err := family.check(); err != nil {
		return err
	}
	if index < 0 || index >= len(family.sets) {
		return ErrDimension
	}
	was := family.sets[index].Flags&EspressoSetActive != 0
	if was == active {
		return nil
	}
	if active {
		family.sets[index].Flags |= EspressoSetActive
		family.activeCount++
	} else {
		family.sets[index].Flags &^= EspressoSetActive
		family.activeCount--
	}
	return nil
}

func (family *EspressoSetFamily) CompactActive() error {
	if err := family.check(); err != nil {
		return err
	}
	result := family.sets[:0]
	for _, set := range family.sets {
		if set.Flags&EspressoSetActive != 0 {
			result = append(result, set)
		}
	}
	family.sets = result
	family.activeCount = len(result)
	return nil
}

func (family *EspressoSetFamily) UnionAll() (*EspressoSet, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	result, _ := NewEspressoSet(family.bits)
	for _, set := range family.sets {
		for index, word := range set.words {
			result.words[index] |= word
		}
	}
	return result, nil
}

func (family *EspressoSetFamily) IntersectionAll() (*EspressoSet, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	result, _ := NewFullEspressoSet(family.bits)
	for _, set := range family.sets {
		for index, word := range set.words {
			result.words[index] &= word
		}
	}
	return result, nil
}

func JoinEspressoSetFamilies(first, second *EspressoSetFamily) (*EspressoSetFamily, error) {
	if err := first.check(); err != nil {
		return nil, err
	}
	if err := second.check(); err != nil {
		return nil, err
	}
	if first.bits != second.bits {
		return nil, ErrDimension
	}
	result, _ := NewEspressoSetFamily(len(first.sets)+len(second.sets), first.bits)
	for _, source := range [][]*EspressoSet{first.sets, second.sets} {
		for _, set := range source {
			clone, _ := set.Clone()
			result.sets = append(result.sets, clone)
			if clone.Flags&EspressoSetActive != 0 {
				result.activeCount++
			}
		}
	}
	return result, nil
}

func (family *EspressoSetFamily) AppendAndClose(other *EspressoSetFamily) error {
	if err := family.check(); err != nil {
		return err
	}
	if err := other.check(); err != nil {
		return err
	}
	if family == other || family.bits != other.bits {
		return ErrDimension
	}
	for _, set := range other.sets {
		clone, _ := set.Clone()
		family.sets = append(family.sets, clone)
		if clone.Flags&EspressoSetActive != 0 {
			family.activeCount++
		}
	}
	return other.Close()
}

func (family *EspressoSetFamily) ColumnCounts() ([]int, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	result := make([]int, family.bits)
	for _, set := range family.sets {
		for wordIndex, word := range set.words {
			for word != 0 {
				bit := bits.TrailingZeros32(word)
				result[wordIndex*32+bit]++
				word &^= uint32(1) << uint(bit)
			}
		}
	}
	return result, nil
}

func (family *EspressoSetFamily) RestrictedColumnCounts(restrict *EspressoSet) ([]int, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	if err := espressoSetPair(restrict, &EspressoSet{bits: family.bits, words: make([]uint32, espressoSetWordCount(family.bits))}); err != nil {
		return nil, err
	}
	result := make([]int, family.bits)
	for _, set := range family.sets {
		cardinality, _ := set.Cardinality()
		if cardinality <= 1 {
			return nil, ErrInvariant
		}
		weight := 1024 / (cardinality - 1)
		for wordIndex, word := range set.words {
			word &= restrict.words[wordIndex]
			for word != 0 {
				bit := bits.TrailingZeros32(word)
				result[wordIndex*32+bit] += weight
				word &^= uint32(1) << uint(bit)
			}
		}
	}
	return result, nil
}

func (set *EspressoSet) AdjustCounts(counts []int, weight int) error {
	if err := set.check(); err != nil {
		return err
	}
	if len(counts) != set.bits {
		return ErrDimension
	}
	elements, _ := set.Elements()
	for _, element := range elements {
		counts[element] += weight
	}
	return nil
}

func (family *EspressoSetFamily) transformColumns(first, delta int) (*EspressoSetFamily, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	if first < 0 || first > family.bits || delta > 0 && first+delta > family.bits || family.bits-delta < 0 || family.bits-delta > espressoSetMaximumBits {
		return nil, ErrDimension
	}
	result, err := NewEspressoSetFamily(len(family.sets), family.bits-delta)
	if err != nil {
		return nil, err
	}
	for _, source := range family.sets {
		destination, _ := NewEspressoSet(result.bits)
		for _, element := range mustEspressoElements(source) {
			switch {
			case element < first:
				_ = destination.Set(element, true)
			case delta > 0 && element < first+delta:
				continue
			default:
				_ = destination.Set(element-delta, true)
			}
		}
		result.sets = append(result.sets, destination)
	}
	return result, nil
}

func mustEspressoElements(set *EspressoSet) []int {
	result, _ := set.Elements()
	return result
}

func (family *EspressoSetFamily) DeleteColumns(first, count int) (*EspressoSetFamily, error) {
	if count < 0 {
		return nil, ErrDimension
	}
	return family.transformColumns(first, count)
}

func (family *EspressoSetFamily) AddColumns(first, count int) (*EspressoSetFamily, error) {
	if count < 0 {
		return nil, ErrDimension
	}
	if err := family.check(); err != nil {
		return nil, err
	}
	if first == family.bits && family.bits+count <= 32*espressoSetWordCount(family.bits) {
		result, _ := family.Clone()
		result.bits += count
		for _, set := range result.sets {
			set.bits += count
		}
		return result, nil
	}
	return family.transformColumns(first, -count)
}

func (family *EspressoSetFamily) CopyColumnFrom(destinationColumn int, source *EspressoSetFamily, sourceColumn int) error {
	if err := family.check(); err != nil {
		return err
	}
	if err := source.check(); err != nil {
		return err
	}
	if destinationColumn < 0 || destinationColumn >= family.bits || sourceColumn < 0 || sourceColumn >= source.bits || len(family.sets) != len(source.sets) {
		return ErrDimension
	}
	for index, sourceSet := range source.sets {
		present, _ := sourceSet.Has(sourceColumn)
		if present {
			_ = family.sets[index].Set(destinationColumn, true)
		}
	}
	return nil
}

func (family *EspressoSetFamily) Compress(columns *EspressoSet) (*EspressoSetFamily, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	if err := columns.check(); err != nil || columns.bits != family.bits {
		return nil, ErrDimension
	}
	permutation := mustEspressoElements(columns)
	return family.Permute(permutation)
}

func (family *EspressoSetFamily) Transpose() (*EspressoSetFamily, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	result, err := NewEspressoSetFamily(family.bits, len(family.sets))
	if err != nil {
		return nil, err
	}
	for column := 0; column < family.bits; column++ {
		set, _ := NewEspressoSet(len(family.sets))
		for row, source := range family.sets {
			present, _ := source.Has(column)
			if present {
				_ = set.Set(row, true)
			}
		}
		result.sets = append(result.sets, set)
	}
	return result, nil
}

func (family *EspressoSetFamily) Permute(permutation []int) (*EspressoSetFamily, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	if len(permutation) > espressoSetMaximumBits {
		return nil, ErrLimit
	}
	for _, sourceColumn := range permutation {
		if sourceColumn < 0 || sourceColumn >= family.bits {
			return nil, ErrDimension
		}
	}
	result, _ := NewEspressoSetFamily(len(family.sets), len(permutation))
	for _, source := range family.sets {
		destination, _ := NewEspressoSet(len(permutation))
		for destinationColumn, sourceColumn := range permutation {
			present, _ := source.Has(sourceColumn)
			if present {
				_ = destination.Set(destinationColumn, true)
			}
		}
		result.sets = append(result.sets, destination)
	}
	return result, nil
}

type EspressoSetFamilyLimits struct {
	MaxSets, MaxBits, MaxBytes int
}

func normalizeEspressoSetFamilyLimits(limits EspressoSetFamilyLimits) (EspressoSetFamilyLimits, error) {
	if limits.MaxSets < 0 || limits.MaxBits < 0 || limits.MaxBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxSets == 0 {
		limits.MaxSets = 1 << 20
	}
	if limits.MaxBits == 0 {
		limits.MaxBits = espressoSetMaximumBits
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = 64 << 20
	}
	return limits, nil
}

func espressoSetMetadata(set *EspressoSet) uint32 {
	return uint32(len(set.words)) | uint32(set.Flags) | uint32(set.Rank)<<16
}

func (family *EspressoSetFamily) MarshalText() ([]byte, error) {
	if err := family.check(); err != nil {
		return nil, err
	}
	var output strings.Builder
	output.WriteString(strconv.Itoa(len(family.sets)))
	output.WriteByte(' ')
	output.WriteString(strconv.Itoa(family.bits))
	output.WriteByte('\n')
	for _, set := range family.sets {
		output.WriteString(strconv.FormatUint(uint64(espressoSetMetadata(set)), 16))
		for _, word := range set.words {
			output.WriteByte(' ')
			output.WriteString(strconv.FormatUint(uint64(word), 16))
		}
		output.WriteByte('\n')
	}
	return []byte(output.String()), nil
}

func ParseEspressoSetFamily(data []byte, limits EspressoSetFamilyLimits) (*EspressoSetFamily, error) {
	limits, err := normalizeEspressoSetFamilyLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	fields := bytes.Fields(data)
	if len(fields) < 2 {
		return nil, ErrFormat
	}
	setCount, countErr := strconv.Atoi(string(fields[0]))
	bitCount, bitErr := strconv.Atoi(string(fields[1]))
	if countErr != nil || bitErr != nil || setCount < 0 || bitCount < 0 {
		return nil, ErrFormat
	}
	if setCount > limits.MaxSets || bitCount > limits.MaxBits || bitCount > espressoSetMaximumBits {
		return nil, ErrLimit
	}
	wordCount := espressoSetWordCount(bitCount)
	if len(fields) != 2+setCount*(1+wordCount) {
		return nil, ErrFormat
	}
	family, _ := NewEspressoSetFamily(setCount, bitCount)
	position := 2
	for range setCount {
		metadata, parseErr := strconv.ParseUint(string(fields[position]), 16, 32)
		position++
		if parseErr != nil || int(metadata&0x3ff) != wordCount {
			_ = family.Close()
			return nil, ErrFormat
		}
		set, _ := NewEspressoSet(bitCount)
		set.Rank = int(uint32(metadata) >> 16)
		set.Flags = EspressoSetFlags(uint32(metadata) & 0xfc00)
		for word := range set.words {
			value, wordErr := strconv.ParseUint(string(fields[position]), 16, 32)
			position++
			if wordErr != nil {
				_ = family.Close()
				return nil, ErrFormat
			}
			set.words[word] = uint32(value)
		}
		if set.check() != nil {
			_ = family.Close()
			return nil, ErrFormat
		}
		family.sets = append(family.sets, set)
		if set.Flags&EspressoSetActive != 0 {
			family.activeCount++
		}
	}
	return family, nil
}

func (family *EspressoSetFamily) FormatSets() (string, error) {
	if err := family.check(); err != nil {
		return "", err
	}
	var output strings.Builder
	for index, set := range family.sets {
		output.WriteString("A[")
		output.WriteString(strconv.Itoa(index))
		output.WriteString("] = ")
		output.WriteString(set.String())
		output.WriteByte('\n')
	}
	return output.String(), nil
}

func (family *EspressoSetFamily) FormatBitMatrix() (string, error) {
	if err := family.check(); err != nil {
		return "", err
	}
	var output strings.Builder
	for index, set := range family.sets {
		output.WriteByte('[')
		text := strconv.Itoa(index)
		output.WriteString(strings.Repeat(" ", max(0, 4-len(text))))
		output.WriteString(text)
		output.WriteString("] ")
		output.WriteString(set.BitString())
		output.WriteByte('\n')
	}
	return output.String(), nil
}

func (family *EspressoSetFamily) MarshalBitMatrix() ([]byte, error) {
	formatted, err := family.FormatBitMatrix()
	if err != nil {
		return nil, err
	}
	var output strings.Builder
	output.WriteString(strconv.Itoa(len(family.sets)))
	output.WriteByte(' ')
	output.WriteString(strconv.Itoa(family.bits))
	output.WriteByte('\n')
	for _, line := range strings.Split(strings.TrimSuffix(formatted, "\n"), "\n") {
		if line == "" {
			continue
		}
		position := strings.Index(line, "] ")
		output.WriteString(line[position+2:])
		output.WriteByte('\n')
	}
	return []byte(output.String()), nil
}

func ParseEspressoBitMatrix(data []byte, limits EspressoSetFamilyLimits) (*EspressoSetFamily, error) {
	limits, err := normalizeEspressoSetFamilyLimits(limits)
	if err != nil {
		return nil, err
	}
	if len(data) > limits.MaxBytes {
		return nil, ErrLimit
	}
	fields := bytes.Fields(data)
	if len(fields) < 2 {
		return nil, ErrFormat
	}
	setCount, countErr := strconv.Atoi(string(fields[0]))
	bitCount, bitErr := strconv.Atoi(string(fields[1]))
	if countErr != nil || bitErr != nil || setCount < 0 || bitCount < 0 || len(fields) != 2+setCount {
		return nil, ErrFormat
	}
	if setCount > limits.MaxSets || bitCount > limits.MaxBits || bitCount > espressoSetMaximumBits {
		return nil, ErrLimit
	}
	family, _ := NewEspressoSetFamily(setCount, bitCount)
	for row := 0; row < setCount; row++ {
		field := fields[2+row]
		if len(field) != bitCount {
			_ = family.Close()
			return nil, ErrFormat
		}
		set, _ := NewEspressoSet(bitCount)
		for column, value := range field {
			switch value {
			case '0':
			case '1':
				_ = set.Set(column, true)
			default:
				_ = family.Close()
				return nil, ErrFormat
			}
		}
		family.sets = append(family.sets, set)
	}
	return family, nil
}
