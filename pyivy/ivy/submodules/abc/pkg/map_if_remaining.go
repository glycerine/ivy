package abc

// Source-shaped bounded cut operations from src/map/if/ifCut.c, ifTruth.c,
// ifLibLut.c, and ifTime.c.

import (
	"math"
	"strconv"
	"strings"
)

const (
	ifMapMaximumLeaves       = 16
	ifMapCutHeaderWords      = 11
	ifMapPermutationHdrWords = 13
)

type IfMapLimits struct {
	MaxLeaves int
	MaxWork   int
	MaxBytes  int
	Stop      func() bool
}

type IfCutCost struct {
	Delay float64
	Area  float64
	Edge  float64
	Power float64
}

// IfMappingCut owns a normalized truth table. The stored table has minterm
// zero clear and complemented records whether the represented function is its
// complement, matching If_CutComputeTruth's truth-ID convention.
type IfMappingCut struct {
	leaves       []int
	truth        []uint64
	complemented bool
	cost         IfCutCost
}

type IfPinPermutation struct {
	Leaves       []int
	Delays       []float64
	Truth        []uint64
	OldPins      []int
	Complemented bool
}

type IfLUTEntry struct {
	Area   float64
	Delays []float64
}

type IfLUTLibrary struct {
	entries  []IfLUTEntry
	variable bool
}

type ifMapBudget struct {
	limits IfMapLimits
	work   int
}

func normalizedIfLimits(limits IfMapLimits) (IfMapLimits, error) {
	if limits.MaxLeaves == 0 {
		limits.MaxLeaves = ifMapMaximumLeaves
	}
	if limits.MaxLeaves < 0 || limits.MaxLeaves > ifMapMaximumLeaves || limits.MaxWork < 0 || limits.MaxBytes < 0 {
		return IfMapLimits{}, ErrDimension
	}
	return limits, nil
}

func (b *ifMapBudget) take(n int) error {
	if n < 0 || b.work > int(^uint(0)>>1)-n {
		return ErrLimit
	}
	if b.limits.Stop != nil && b.limits.Stop() {
		return ErrCancelled
	}
	b.work += n
	if b.limits.MaxWork != 0 && b.work > b.limits.MaxWork {
		return ErrLimit
	}
	return nil
}

func ifMapBytesOK(limits IfMapLimits, counts ...int) bool {
	if limits.MaxBytes == 0 {
		return true
	}
	total := 0
	for _, count := range counts {
		if count < 0 || count > (int(^uint(0)>>1)-total)/8 {
			return false
		}
		total += count * 8
	}
	return total <= limits.MaxBytes
}

func ifTruthWords(leaves int) int {
	bitsCount := 1 << uint(leaves)
	return (bitsCount + 63) >> 6
}

func ifTruthMask(leaves int) uint64 {
	rem := (1 << uint(leaves)) & 63
	if rem == 0 {
		return ^uint64(0)
	}
	return uint64(1)<<uint(rem) - 1
}

func NewIfMappingCut(leaves []int, truth []uint64, cost IfCutCost, limits IfMapLimits) (*IfMappingCut, error) {
	limits, err := normalizedIfLimits(limits)
	if err != nil || len(leaves) > limits.MaxLeaves || len(leaves) > ifMapMaximumLeaves {
		return nil, ErrDimension
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	if len(truth) != ifTruthWords(len(leaves)) {
		return nil, ErrDimension
	}
	if !ifMapBytesOK(limits, ifMapCutHeaderWords, len(leaves), len(truth)) {
		return nil, ErrLimit
	}
	for i, leaf := range leaves {
		if leaf < 0 || i > 0 && leaves[i-1] >= leaf {
			return nil, ErrDimension
		}
	}
	for _, value := range []float64{cost.Delay, cost.Area, cost.Edge, cost.Power} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, ErrDimension
		}
	}
	mask := ifTruthMask(len(leaves))
	if truth[len(truth)-1]&^mask != 0 {
		return nil, ErrDimension
	}
	cut := &IfMappingCut{leaves: append([]int(nil), leaves...), truth: append([]uint64(nil), truth...), cost: cost}
	cut.complemented = cut.truth[0]&1 != 0
	if cut.complemented {
		for i := range cut.truth {
			cut.truth[i] = ^cut.truth[i]
		}
		cut.truth[len(cut.truth)-1] &= mask
	}
	return cut, nil
}

func (cut *IfMappingCut) Leaves() []int {
	if cut == nil {
		return nil
	}
	return append([]int(nil), cut.leaves...)
}

func (cut *IfMappingCut) Cost() IfCutCost {
	if cut == nil {
		return IfCutCost{}
	}
	return cut.cost
}

func (cut *IfMappingCut) Complemented() bool { return cut != nil && cut.complemented }

func (cut *IfMappingCut) value(assignment int) bool {
	value := cut.truth[assignment>>6]&(uint64(1)<<uint(assignment&63)) != 0
	return value != cut.complemented
}

func (cut *IfMappingCut) Truth() []uint64 {
	if cut == nil {
		return nil
	}
	result := append([]uint64(nil), cut.truth...)
	if cut.complemented {
		for i := range result {
			result[i] = ^result[i]
		}
		result[len(result)-1] &= ifTruthMask(len(cut.leaves))
	}
	return result
}

func validateIfCut(cut *IfMappingCut, limits IfMapLimits) error {
	if cut == nil || len(cut.leaves) > limits.MaxLeaves || len(cut.truth) != ifTruthWords(len(cut.leaves)) {
		return ErrDimension
	}
	for i, leaf := range cut.leaves {
		if leaf < 0 || i > 0 && cut.leaves[i-1] >= leaf {
			return ErrInvariant
		}
	}
	if cut.truth[0]&1 != 0 || cut.truth[len(cut.truth)-1]&^ifTruthMask(len(cut.leaves)) != 0 {
		return ErrInvariant
	}
	return nil
}

// MergeIfMappingCuts is the ordered cut union and expanded truth-table AND of
// If_CutMergeOrdered and If_CutComputeTruth.
func MergeIfMappingCuts(left, right *IfMappingCut, complementLeft, complementRight bool, cost IfCutCost, limits IfMapLimits) (*IfMappingCut, error) {
	limits, err := normalizedIfLimits(limits)
	if err != nil {
		return nil, err
	}
	if err = validateIfCut(left, limits); err != nil {
		return nil, err
	}
	if err = validateIfCut(right, limits); err != nil {
		return nil, err
	}
	for _, value := range []float64{cost.Delay, cost.Area, cost.Edge, cost.Power} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, ErrDimension
		}
	}
	budget := ifMapBudget{limits: limits}
	if err = budget.take(len(left.leaves) + len(right.leaves) + len(left.truth) + len(right.truth)); err != nil {
		return nil, err
	}
	if !ifMapBytesOK(limits, len(left.leaves)+len(right.leaves)) {
		return nil, ErrLimit
	}
	union := make([]int, 0, len(left.leaves)+len(right.leaves))
	i, j := 0, 0
	for i < len(left.leaves) || j < len(right.leaves) {
		if err = budget.take(1); err != nil {
			return nil, err
		}
		var leaf int
		switch {
		case j == len(right.leaves) || i < len(left.leaves) && left.leaves[i] < right.leaves[j]:
			leaf, i = left.leaves[i], i+1
		case i == len(left.leaves) || right.leaves[j] < left.leaves[i]:
			leaf, j = right.leaves[j], j+1
		default:
			leaf, i, j = left.leaves[i], i+1, j+1
		}
		union = append(union, leaf)
		if len(union) > limits.MaxLeaves {
			return nil, ErrLimit
		}
	}
	words := ifTruthWords(len(union))
	if !ifMapBytesOK(limits, cap(union), words) {
		return nil, ErrLimit
	}
	leftPos, rightPos := make([]int, len(left.leaves)), make([]int, len(right.leaves))
	// Peak storage includes the merge temporaries and the result copies made by
	// NewIfMappingCut before those temporaries become unreachable.
	if !ifMapBytesOK(limits, cap(union), len(leftPos), len(rightPos), 2*words, len(union), ifMapCutHeaderWords) {
		return nil, ErrLimit
	}
	for u, leaf := range union {
		if err = budget.take(1 + len(left.leaves) + len(right.leaves)); err != nil {
			return nil, err
		}
		if k := indexIfLeaf(left.leaves, leaf); k >= 0 {
			leftPos[k] = u
		}
		if k := indexIfLeaf(right.leaves, leaf); k >= 0 {
			rightPos[k] = u
		}
	}
	truth := make([]uint64, words)
	for assignment := 0; assignment < 1<<uint(len(union)); assignment++ {
		if err = budget.take(1 + len(leftPos) + len(rightPos)); err != nil {
			return nil, err
		}
		leftAssignment, rightAssignment := 0, 0
		for k, position := range leftPos {
			leftAssignment |= ((assignment >> uint(position)) & 1) << uint(k)
		}
		for k, position := range rightPos {
			rightAssignment |= ((assignment >> uint(position)) & 1) << uint(k)
		}
		lv := left.value(leftAssignment) != complementLeft
		rv := right.value(rightAssignment) != complementRight
		if lv && rv {
			truth[assignment>>6] |= uint64(1) << uint(assignment&63)
		}
	}
	return NewIfMappingCut(union, truth, cost, limits)
}

func indexIfLeaf(leaves []int, leaf int) int {
	lo, hi := 0, len(leaves)
	for lo < hi {
		mid := lo + (hi-lo)/2
		if leaves[mid] < leaf {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo < len(leaves) && leaves[lo] == leaf {
		return lo
	}
	return -1
}

func IfCutDominates(base, candidate *IfMappingCut) bool {
	if base == nil || candidate == nil || len(base.leaves) > len(candidate.leaves) {
		return false
	}
	for _, leaf := range base.leaves {
		if indexIfLeaf(candidate.leaves, leaf) < 0 {
			return false
		}
	}
	return true
}

type IfCutOrder uint8

const (
	IfCutDelayFirst IfCutOrder = iota
	IfCutAreaFirst
)

func CompareIfCutCost(left, right *IfMappingCut, order IfCutOrder, epsilon float64) (int, error) {
	if left == nil || right == nil || epsilon < 0 || math.IsNaN(epsilon) || math.IsInf(epsilon, 0) || order > IfCutAreaFirst {
		return 0, ErrDimension
	}
	compare := func(a, b float64) int {
		if a < b-epsilon {
			return -1
		}
		if a > b+epsilon {
			return 1
		}
		return 0
	}
	if order == IfCutDelayFirst {
		if result := compare(left.cost.Delay, right.cost.Delay); result != 0 {
			return result, nil
		}
		if len(left.leaves) < len(right.leaves) {
			return -1, nil
		}
		if len(left.leaves) > len(right.leaves) {
			return 1, nil
		}
		return compare(left.cost.Area, right.cost.Area), nil
	}
	if result := compare(left.cost.Area, right.cost.Area); result != 0 {
		return result, nil
	}
	if len(left.leaves) < len(right.leaves) {
		return -1, nil
	}
	if len(left.leaves) > len(right.leaves) {
		return 1, nil
	}
	return compare(left.cost.Delay, right.cost.Delay), nil
}

// PermuteIfCutPins implements If_CutTruthPermute's stable decreasing-delay
// order while returning owned data instead of mutating the cut. Truth remains
// normalized with minterm zero clear and Complemented carries the phase, just
// as the native truth ID and its complement bit remain separate.
func PermuteIfCutPins(cut *IfMappingCut, delays []float64, limits IfMapLimits) (IfPinPermutation, error) {
	limits, err := normalizedIfLimits(limits)
	if err != nil || cut == nil || len(delays) != len(cut.leaves) {
		return IfPinPermutation{}, ErrDimension
	}
	if err = validateIfCut(cut, limits); err != nil {
		return IfPinPermutation{}, err
	}
	words := ifTruthWords(len(cut.leaves))
	if !ifMapBytesOK(limits, ifMapPermutationHdrWords, len(cut.leaves)*3, words) {
		return IfPinPermutation{}, ErrLimit
	}
	for _, delay := range delays {
		if math.IsNaN(delay) || math.IsInf(delay, 0) {
			return IfPinPermutation{}, ErrDimension
		}
	}
	budget := ifMapBudget{limits: limits}
	if err = budget.take(len(cut.leaves) + len(cut.truth) + len(delays)); err != nil {
		return IfPinPermutation{}, err
	}
	result := IfPinPermutation{Leaves: append([]int(nil), cut.leaves...), Delays: append([]float64(nil), delays...), Truth: make([]uint64, words), OldPins: make([]int, len(cut.leaves)), Complemented: cut.complemented}
	for i := range result.Delays {
		result.OldPins[i] = i
	}
	for i := 1; i < len(result.OldPins); i++ {
		for j := i; j > 0; j-- {
			if err = budget.take(1); err != nil {
				return IfPinPermutation{}, err
			}
			if result.Delays[j] <= result.Delays[j-1] {
				break
			}
			result.Delays[j], result.Delays[j-1] = result.Delays[j-1], result.Delays[j]
			result.Leaves[j], result.Leaves[j-1] = result.Leaves[j-1], result.Leaves[j]
			result.OldPins[j], result.OldPins[j-1] = result.OldPins[j-1], result.OldPins[j]
		}
	}
	for assignment := 0; assignment < 1<<uint(len(cut.leaves)); assignment++ {
		if err = budget.take(1 + len(cut.leaves)); err != nil {
			return IfPinPermutation{}, err
		}
		oldAssignment := 0
		for newPin, oldPin := range result.OldPins {
			oldAssignment |= ((assignment >> uint(newPin)) & 1) << uint(oldPin)
		}
		if cut.truth[oldAssignment>>6]&(uint64(1)<<uint(oldAssignment&63)) != 0 {
			result.Truth[assignment>>6] |= uint64(1) << uint(assignment&63)
		}
	}
	return result, nil
}

func ParseIfLUTLibrary(text string, limits IfMapLimits) (*IfLUTLibrary, error) {
	limits, err := normalizedIfLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	if limits.MaxBytes != 0 && len(text) > limits.MaxBytes {
		return nil, ErrLimit
	}
	lineCount := 1 + strings.Count(text, "\n")
	// strings.Split retains two words per line descriptor. Account for that
	// scratch before creating it; per-line token descriptor scratch is checked
	// below before strings.Fields allocates it.
	if !ifMapBytesOK(limits, 2*lineCount) {
		return nil, ErrLimit
	}
	budget := ifMapBudget{limits: limits}
	var parsed [ifMapMaximumLeaves + 1]IfLUTEntry
	entryCount := 1
	mode := 0
	totalDelays := 0
	for _, line := range strings.Split(text, "\n") {
		if err = budget.take(1); err != nil {
			return nil, err
		}
		maxFields := (len(line) + 1) / 2
		if !ifMapBytesOK(limits, 2*lineCount, 4*entryCount, totalDelays, 2*maxFields) {
			return nil, ErrLimit
		}
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		if err = budget.take(len(fields)); err != nil {
			return nil, err
		}
		size, parseErr := strconv.Atoi(fields[0])
		if parseErr != nil || size != entryCount || size > limits.MaxLeaves || len(fields) < 3 {
			return nil, ErrFormat
		}
		area, parseErr := strconv.ParseFloat(fields[1], 64)
		if parseErr != nil || area < 0 || math.IsNaN(area) || math.IsInf(area, 0) {
			return nil, ErrFormat
		}
		delayCount := len(fields) - 2
		if delayCount != 1 && delayCount != size {
			return nil, ErrFormat
		}
		if size > 1 && mode == 0 {
			mode = delayCount
		} else if size > 1 && (mode == 1) != (delayCount == 1) {
			return nil, ErrFormat
		}
		if !ifMapBytesOK(limits, 2*lineCount, 4*(entryCount+1), totalDelays+delayCount, 2*len(fields)) {
			return nil, ErrLimit
		}
		entry := IfLUTEntry{Area: area, Delays: make([]float64, delayCount)}
		for i := range entry.Delays {
			entry.Delays[i], parseErr = strconv.ParseFloat(fields[i+2], 64)
			if parseErr != nil || entry.Delays[i] <= 0 || math.IsNaN(entry.Delays[i]) || math.IsInf(entry.Delays[i], 0) || i > 0 && entry.Delays[i-1] > entry.Delays[i] {
				return nil, ErrFormat
			}
		}
		parsed[entryCount] = entry
		entryCount++
		totalDelays += delayCount
	}
	if entryCount == 1 {
		return nil, ErrFormat
	}
	if !ifMapBytesOK(limits, 2*lineCount, 4*entryCount, totalDelays, 4) {
		return nil, ErrLimit
	}
	entries := make([]IfLUTEntry, entryCount)
	copy(entries, parsed[:entryCount])
	return &IfLUTLibrary{entries: entries, variable: mode > 1}, nil
}

func NewSimpleIfLUTLibrary(maxSize int) (*IfLUTLibrary, error) {
	if maxSize < 3 || maxSize > 10 {
		return nil, ErrDimension
	}
	entries := make([]IfLUTEntry, maxSize+1)
	for size := 1; size <= maxSize; size++ {
		entries[size] = IfLUTEntry{Area: 1, Delays: []float64{1}}
	}
	return &IfLUTLibrary{entries: entries}, nil
}

func (library *IfLUTLibrary) Entry(size int) (IfLUTEntry, error) {
	if library == nil || size <= 0 || size >= len(library.entries) {
		return IfLUTEntry{}, ErrDimension
	}
	entry := library.entries[size]
	entry.Delays = append([]float64(nil), entry.Delays...)
	return entry, nil
}

func (library *IfLUTLibrary) FastestPinDelay() float64 {
	if library == nil || len(library.entries) <= 1 {
		return 1
	}
	return library.entries[len(library.entries)-1].Delays[0]
}

func (library *IfLUTLibrary) SlowestPinDelay() float64 {
	if library == nil || len(library.entries) <= 1 {
		return 1
	}
	delays := library.entries[len(library.entries)-1].Delays
	return delays[len(delays)-1]
}

// IfCutDelay computes If_CutDelay's LUT-library branch.
func IfCutDelay(cut *IfMappingCut, arrivals []float64, library *IfLUTLibrary, limits IfMapLimits) (float64, error) {
	limits, err := normalizedIfLimits(limits)
	if err != nil || cut == nil || library == nil || len(arrivals) != len(cut.leaves) || len(cut.leaves) == 0 || len(cut.leaves) >= len(library.entries) {
		return 0, ErrDimension
	}
	if err = validateIfCut(cut, limits); err != nil {
		return 0, err
	}
	entry := library.entries[len(cut.leaves)]
	for _, arrival := range arrivals {
		if math.IsNaN(arrival) || math.IsInf(arrival, 0) {
			return 0, ErrDimension
		}
	}
	budget := ifMapBudget{limits: limits}
	if library.variable {
		if len(entry.Delays) != len(arrivals) || !ifMapBytesOK(limits, len(arrivals)) {
			return 0, ErrInvariant
		}
		sorted := append([]float64(nil), arrivals...)
		for i := 1; i < len(sorted); i++ {
			for j := i; j > 0; j-- {
				if err = budget.take(1); err != nil {
					return 0, err
				}
				if sorted[j] <= sorted[j-1] {
					break
				}
				sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
			}
		}
		result := math.Inf(-1)
		for i, arrival := range sorted {
			if err = budget.take(1); err != nil {
				return 0, err
			}
			result = math.Max(result, arrival+entry.Delays[i])
		}
		return result, nil
	}
	if len(entry.Delays) != 1 {
		return 0, ErrInvariant
	}
	result := math.Inf(-1)
	for _, arrival := range arrivals {
		if err = budget.take(1); err != nil {
			return 0, err
		}
		result = math.Max(result, arrival+entry.Delays[0])
	}
	return result, nil
}

// IfCompiledSources is the exact module.make inventory, returned as an owned
// snapshot so callers cannot mutate package state.
func IfCompiledSources() []string {
	return []string{"ifCom.c", "ifCache.c", "ifCore.c", "ifCut.c", "ifData2.c", "ifDec07.c", "ifDec08.c", "ifDec10.c", "ifDec16.c", "ifDec75.c", "ifDelay.c", "ifDsd.c", "ifLibBox.c", "ifLibLut.c", "ifMan.c", "ifMap.c", "ifMatch2.c", "ifReduce.c", "ifSat.c", "ifSelect.c", "ifSeq.c", "ifTest.c", "ifTime.c", "ifTruth.c", "ifTune.c", "ifUtil.c"}
}

func IfTruthSupportSize(cut *IfMappingCut) int {
	if cut == nil {
		return 0
	}
	support := 0
	for variable := range cut.leaves {
		bit := 1 << uint(variable)
		dependent := false
		for assignment := 0; assignment < 1<<uint(len(cut.leaves)); assignment++ {
			if assignment&bit == 0 && cut.value(assignment) != cut.value(assignment|bit) {
				dependent = true
				break
			}
		}
		if dependent {
			support++
		}
	}
	return support
}

func IfCutSignature(cut *IfMappingCut) uint32 {
	if cut == nil {
		return 0
	}
	var signature uint32
	for _, leaf := range cut.leaves {
		signature |= uint32(1) << uint(leaf%31)
	}
	return signature
}
