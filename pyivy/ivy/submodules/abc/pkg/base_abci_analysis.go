package abc

// Truth-table analysis from src/base/abci/abcMeasure.c, with exact bounded
// quantification and unateness semantics from abcQuant.c and abcUnate.c.

import (
	"math/bits"
	"strconv"
)

const abciIntBytes = strconv.IntSize / 8

// AbciTruthTable is an immutable, owned, least-significant-minterm-first truth
// table. Its zero value is invalid; construct it with NewAbciTruthTable.
type AbciTruthTable struct {
	nVars int
	words []uint64
}

type AbciAnalysisOptions struct {
	MaxEvaluations int
	MaxBytes       int
	Stop           func() bool
}

type AbciTruthMeasures struct {
	Ones                int
	OnSetOffSetPairs    uint64
	VariableDifferences []int
	PairDifferences     [][]int
	Evaluations         int
}

type AbciBoundDecomposition struct {
	Size        int
	Mask        uint32
	Evaluations int
}

type AbciUnateKind uint8

const (
	AbciIndependent AbciUnateKind = iota
	AbciPositiveUnate
	AbciNegativeUnate
	AbciBinate
)

func NewAbciTruthTable(nVars int, words []uint64) (AbciTruthTable, error) {
	if nVars < 0 || nVars > 16 {
		return AbciTruthTable{}, ErrDimension
	}
	nBits := 1 << uint(nVars)
	nWords := (nBits + 63) / 64
	if len(words) != nWords {
		return AbciTruthTable{}, ErrDimension
	}
	owned := append([]uint64(nil), words...)
	if rem := nBits & 63; rem != 0 && owned[nWords-1]>>uint(rem) != 0 {
		return AbciTruthTable{}, ErrDimension
	}
	return AbciTruthTable{nVars: nVars, words: owned}, nil
}

func (t AbciTruthTable) Vars() int {
	return t.nVars
}

// Words returns an owned snapshot. Callers cannot mutate a table after
// construction, so concurrent read-only operations on one table are safe.
func (t AbciTruthTable) Words() []uint64 {
	return append([]uint64(nil), t.words...)
}

func (t AbciTruthTable) value(minterm int) bool {
	return t.words[minterm>>6]&(uint64(1)<<uint(minterm&63)) != 0
}

func (t AbciTruthTable) validate() error {
	if t.nVars < 0 || t.nVars > 16 {
		return ErrDimension
	}
	nBits := 1 << uint(t.nVars)
	nWords := (nBits + 63) / 64
	if len(t.words) != nWords {
		return ErrDimension
	}
	if rem := nBits & 63; rem != 0 && t.words[nWords-1]>>uint(rem) != 0 {
		return ErrDimension
	}
	return nil
}

type abciBudget struct {
	options AbciAnalysisOptions
	used    int
	bytes   int
}

func (b *abciBudget) take(n int) error {
	if n < 0 || b.used > int(^uint(0)>>1)-n {
		return ErrLimit
	}
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	b.used += n
	if b.options.MaxEvaluations < 0 {
		return ErrDimension
	}
	if b.options.MaxEvaluations != 0 && b.used > b.options.MaxEvaluations {
		return ErrLimit
	}
	return nil
}

func (b *abciBudget) takeBytes(n int) error {
	if n < 0 || b.bytes > int(^uint(0)>>1)-n {
		return ErrLimit
	}
	if b.options.Stop != nil && b.options.Stop() {
		return ErrCancelled
	}
	b.bytes += n
	if b.options.MaxBytes < 0 {
		return ErrDimension
	}
	if b.options.MaxBytes != 0 && b.bytes > b.options.MaxBytes {
		return ErrLimit
	}
	return nil
}

func newAbciBudget(t AbciTruthTable, options AbciAnalysisOptions) (abciBudget, error) {
	b := abciBudget{options: options}
	if err := t.validate(); err != nil || options.MaxEvaluations < 0 || options.MaxBytes < 0 {
		return b, ErrDimension
	}
	if err := b.take(0); err != nil {
		return b, err
	}
	return b, nil
}

func abciBytesForInts(n int) (int, error) {
	if n < 0 || n > int(^uint(0)>>1)/abciIntBytes {
		return 0, ErrLimit
	}
	return n * abciIntBytes, nil
}

// AnalyzeAbciTruth computes the logical activity and pair-crossing measures of
// Abc_NtkPrintMeasures, but returns deterministic data rather than printing it.
// Unlike ABC's 32-bit representation, tables below five variables are not
// replicated to fill a machine word. The source routine's uninitialized buffer
// in its total-pair calculation is intentionally replaced by the input table.
func AnalyzeAbciTruth(t AbciTruthTable, options AbciAnalysisOptions) (AbciTruthMeasures, error) {
	var result AbciTruthMeasures
	budget, err := newAbciBudget(t, options)
	if err != nil {
		return result, err
	}
	n := 1 << uint(t.nVars)
	intCount := t.nVars + t.nVars*t.nVars
	bytes, err := abciBytesForInts(intCount)
	if err != nil {
		return result, err
	}
	// Include the outer slice headers in the operation's retained allocation.
	if t.nVars > (int(^uint(0)>>1)-bytes)/(3*abciIntBytes) {
		return result, ErrLimit
	}
	if err := budget.takeBytes(bytes + t.nVars*3*abciIntBytes); err != nil {
		return result, err
	}
	result.VariableDifferences = make([]int, t.nVars)
	result.PairDifferences = make([][]int, t.nVars)
	for i := range result.PairDifferences {
		result.PairDifferences[i] = make([]int, t.nVars)
	}
	for _, word := range t.words {
		if err := budget.take(1); err != nil {
			return AbciTruthMeasures{}, err
		}
		result.Ones += bits.OnesCount64(word)
	}
	result.OnSetOffSetPairs = uint64(result.Ones) * uint64(n-result.Ones)
	for v := 0; v < t.nVars; v++ {
		step := 1 << uint(v)
		for base := 0; base < n; base += 2 * step {
			for k := 0; k < step; k++ {
				if err := budget.take(1); err != nil {
					return AbciTruthMeasures{}, err
				}
				if t.value(base+k) != t.value(base+k+step) {
					result.VariableDifferences[v] += 2
				}
			}
		}
	}
	for i := 0; i < t.nVars; i++ {
		for k := 0; k < t.nVars; k++ {
			if i == k {
				continue
			}
			mask := (1 << uint(i)) | (1 << uint(k))
			for assignment := 0; assignment < n; assignment++ {
				if assignment&mask != 0 {
					continue
				}
				if err := budget.take(2); err != nil {
					return AbciTruthMeasures{}, err
				}
				f00 := t.value(assignment)
				f10 := t.value(assignment | (1 << uint(i)))
				f01 := t.value(assignment | (1 << uint(k)))
				f11 := t.value(assignment | mask)
				if f00 != f11 {
					result.PairDifferences[i][k] += 4
				}
				if f01 != f10 {
					result.PairDifferences[i][k] += 4
				}
			}
		}
	}
	result.Evaluations = budget.used
	return result, nil
}

// AbciVariableActivityOrder returns least-active variables first, preserving
// source variable order for equal activity.
func AbciVariableActivityOrder(t AbciTruthTable, options AbciAnalysisOptions) ([]int, error) {
	if options.MaxBytes > 0 {
		bytes, err := abciBytesForInts(t.nVars)
		if err != nil || options.MaxBytes <= bytes {
			if err != nil {
				return nil, err
			}
			return nil, ErrLimit
		}
		options.MaxBytes -= bytes
	}
	measures, err := AnalyzeAbciTruth(t, options)
	if err != nil {
		return nil, err
	}
	order := make([]int, t.nVars)
	for i := range order {
		order[i] = i
	}
	for i := 1; i < len(order); i++ {
		value := order[i]
		j := i
		for j > 0 && measures.VariableDifferences[value] < measures.VariableDifferences[order[j-1]] {
			order[j] = order[j-1]
			j--
		}
		order[j] = value
	}
	return order, nil
}

// AbciBestBoundMask implements the bounded-set search used by
// Abc_NtkPrintOneDecomp for seven through eleven variables.
func AbciBestBoundMask(t AbciTruthTable, options AbciAnalysisOptions) (AbciBoundDecomposition, error) {
	var result AbciBoundDecomposition
	budget, err := newAbciBudget(t, options)
	if err != nil {
		return result, err
	}
	if t.nVars <= 6 || t.nVars > 11 {
		return result, ErrDimension
	}
	const bound = 6
	// ABC passes nVars to its recursive mask search even though there are only
	// six bound-set indices. Bits above five cannot partition its 64 cofactors
	// and can inflate the reported size. Search only the actual bound set.
	cofBits := 1 << uint(t.nVars-bound)
	if err := budget.takeBytes((1 << bound) * 4); err != nil {
		return result, err
	}
	cofactors := make([]uint32, 1<<bound)
	for i := range cofactors {
		var cof uint32
		for k := 0; k < cofBits; k++ {
			if err := budget.take(1); err != nil {
				return AbciBoundDecomposition{}, err
			}
			if t.value(i*cofBits + k) {
				cof |= 1 << uint(k)
			}
		}
		cofactors[i] = cof
	}
	result.Size = 1
	var search func(uint32, int) error
	search = func(mask uint32, last int) error {
		if bits.OnesCount32(mask) >= 4 {
			return nil
		}
		for v := last + 1; v < bound; v++ {
			next := mask | 1<<uint(v)
			var groups [1 << bound][2]uint32
			var counts [1 << bound]uint8
			valid := true
			for i, cof := range cofactors {
				if err := budget.take(1); err != nil {
					return err
				}
				group := i & int(next)
				count := counts[group]
				values := groups[group]
				if count == 0 {
					values[0], count = cof, 1
				} else if cof != values[0] && (count == 1 || cof != values[1]) {
					if count == 2 {
						valid = false
						break
					}
					values[1], count = cof, 2
				}
				groups[group], counts[group] = values, count
			}
			if !valid {
				continue
			}
			size := bits.OnesCount32(next)
			if size > result.Size {
				result.Size, result.Mask = size, next
			}
			if err := search(next, v); err != nil {
				return err
			}
		}
		return nil
	}
	for v := 0; v < bound; v++ {
		if err := search(1<<uint(v), v); err != nil {
			return AbciBoundDecomposition{}, err
		}
	}
	result.Evaluations = budget.used
	return result, nil
}

// AbciQuantifyVariable computes the two cofactors and combines them exactly as
// Abc_NtkQuantify: OR for existential and AND for universal quantification.
// The quantified input remains in the table shape but becomes irrelevant.
func AbciQuantifyVariable(t AbciTruthTable, variable int, universal bool, options AbciAnalysisOptions) (AbciTruthTable, error) {
	budget, err := newAbciBudget(t, options)
	if err != nil {
		return AbciTruthTable{}, err
	}
	if variable < 0 || variable >= t.nVars {
		return AbciTruthTable{}, ErrDimension
	}
	if err := budget.takeBytes(len(t.words) * 8); err != nil {
		return AbciTruthTable{}, err
	}
	n := 1 << uint(t.nVars)
	words := make([]uint64, len(t.words))
	bit := 1 << uint(variable)
	for base := 0; base < n; base++ {
		if base&bit != 0 {
			continue
		}
		if err := budget.take(1); err != nil {
			return AbciTruthTable{}, err
		}
		a, b := t.value(base), t.value(base|bit)
		value := a || b
		if universal {
			value = a && b
		}
		if value {
			words[base>>6] |= 1 << uint(base&63)
			other := base | bit
			words[other>>6] |= 1 << uint(other&63)
		}
	}
	return AbciTruthTable{nVars: t.nVars, words: words}, nil
}

// AbciTruthUnateness classifies every variable by direct cofactor comparison.
func AbciTruthUnateness(t AbciTruthTable, options AbciAnalysisOptions) ([]AbciUnateKind, error) {
	budget, err := newAbciBudget(t, options)
	if err != nil {
		return nil, err
	}
	if err := budget.takeBytes(t.nVars); err != nil {
		return nil, err
	}
	n := 1 << uint(t.nVars)
	result := make([]AbciUnateKind, t.nVars)
	for variable := 0; variable < t.nVars; variable++ {
		bit := 1 << uint(variable)
		rises, falls := false, false
		for base := 0; base < n; base++ {
			if base&bit != 0 {
				continue
			}
			if err := budget.take(1); err != nil {
				return nil, err
			}
			a, b := t.value(base), t.value(base|bit)
			rises = rises || !a && b
			falls = falls || a && !b
		}
		switch {
		case rises && falls:
			result[variable] = AbciBinate
		case rises:
			result[variable] = AbciPositiveUnate
		case falls:
			result[variable] = AbciNegativeUnate
		default:
			result[variable] = AbciIndependent
		}
	}
	return result, nil
}
