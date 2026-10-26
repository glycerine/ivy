package abc

// Six-variable truth-table resubstitution from src/bool/rsb/rsbDec6.c,
// rsb.h, and rsbInt.h.

import (
	"math/bits"
	"slices"
	"strconv"
	"strings"
)

// RsbResubOptions bounds an atomic decomposition search. A zero operation
// limit leaves the source-compatible search unbounded.
type RsbResubOptions struct {
	MaxOperations int
	Stop          func() bool
}

// RsbCounterexample is the first zero/one assignment pair that the selected
// divisors fail to distinguish.
type RsbCounterexample struct {
	ZeroAssignment int
	OneAssignment  int
}

// RsbDecompositionCheck is the result of Rsb_DecCheck. Pattern uses the
// source layout: the low 2^n bits record observed zero cofactors and the next
// 2^n bits record observed one cofactors.
type RsbDecompositionCheck struct {
	Pattern        uint32
	Counterexample *RsbCounterexample
}

// RsbResub6Result owns the first source-order decomposition found.
type RsbResub6Result struct {
	Found           bool
	Fanins          []int
	Pattern         uint32
	Offset          uint64
	Onset           uint64
	Counterexamples []RsbCounterexample
	Tries           [][]int
}

func (r RsbResub6Result) Clone() RsbResub6Result {
	r.Fanins = slices.Clone(r.Fanins)
	r.Counterexamples = slices.Clone(r.Counterexamples)
	r.Tries = make([][]int, len(r.Tries))
	for i := range r.Tries {
		r.Tries[i] = slices.Clone(r.Tries[i])
	}
	return r
}

// Format replaces the source diagnostic printers with deterministic owned
// text. It reports the two compatible cofactor functions separately.
func (r RsbResub6Result) Format() string {
	var b strings.Builder
	b.WriteString("fanins {")
	for _, fanin := range r.Fanins {
		b.WriteByte(' ')
		b.WriteString(strconv.Itoa(fanin))
	}
	b.WriteString(" } offset=")
	b.WriteString(strconv.FormatUint(r.Offset, 16))
	b.WriteString(" onset=")
	b.WriteString(strconv.FormatUint(r.Onset, 16))
	return b.String()
}

type rsbSearchBudget struct {
	remaining int
	limited   bool
	stop      func() bool
}

func newRsbSearchBudget(options RsbResubOptions) (*rsbSearchBudget, error) {
	if options.MaxOperations < 0 {
		return nil, ErrDimension
	}
	limited := options.MaxOperations != 0
	remaining := options.MaxOperations
	b := &rsbSearchBudget{remaining: remaining, limited: limited, stop: options.Stop}
	if b.stop != nil && b.stop() {
		return nil, ErrCancelled
	}
	return b, nil
}

func (b *rsbSearchBudget) step() error {
	if b.stop != nil && b.stop() {
		return ErrCancelled
	}
	if b.limited {
		if b.remaining == 0 {
			return ErrLimit
		}
		b.remaining--
	}
	return nil
}

func rsbLogicalAssignments(variables int) (int, error) {
	if variables < 0 || variables > 6 {
		return 0, ErrDimension
	}
	return 1 << uint(variables), nil
}

func rsbValidateTruth(word uint64, variables int) error {
	normalized, err := NormalizeSmallTruthWord(word, variables)
	if err != nil {
		return err
	}
	if normalized != word {
		return ErrInvariant
	}
	return nil
}

func rsbTruthBit(word uint64, assignment int) bool {
	return word&(uint64(1)<<uint(assignment&63)) != 0
}

// rsbDecTry implements the six Rsb_DecTryN inline helpers without relying on
// host byte order. Function zero is the least-significant pattern bit.
func rsbDecTry(c uint64, functions ...uint64) uint32 {
	var result uint32
	for assignment := 0; assignment < 64; assignment++ {
		if !rsbTruthBit(c, assignment) {
			continue
		}
		pattern := 0
		for index, function := range functions {
			if rsbTruthBit(function, assignment) {
				pattern |= 1 << uint(index)
			}
		}
		result |= 1 << uint(pattern)
	}
	return result
}

func rsbDecTry0(c uint64) uint32             { return rsbDecTry(c) }
func rsbDecTry1(c, f1 uint64) uint32         { return rsbDecTry(c, f1) }
func rsbDecTry2(c, f1, f2 uint64) uint32     { return rsbDecTry(c, f1, f2) }
func rsbDecTry3(c, f1, f2, f3 uint64) uint32 { return rsbDecTry(c, f1, f2, f3) }
func rsbDecTry4(c, f1, f2, f3, f4 uint64) uint32 {
	return rsbDecTry(c, f1, f2, f3, f4)
}
func rsbDecTry5(c, f1, f2, f3, f4, f5 uint64) uint32 {
	return rsbDecTry(c, f1, f2, f3, f4, f5)
}

func checkRsbDecomposition6(variables int, truth uint64, divisors []uint64, selected []int, budget *rsbSearchBudget) (RsbDecompositionCheck, error) {
	if _, err := rsbLogicalAssignments(variables); err != nil {
		return RsbDecompositionCheck{}, err
	}
	if len(selected) < 1 || len(selected) > 4 || len(divisors) >= RsbMaxDivisors {
		return RsbDecompositionCheck{}, ErrDimension
	}
	seen := make([]bool, len(divisors))
	for _, index := range selected {
		if index < 0 || index >= len(divisors) || seen[index] {
			return RsbDecompositionCheck{}, ErrDimension
		}
		seen[index] = true
	}
	firstZero := make([]int, 1<<uint(len(selected)))
	firstOne := make([]int, len(firstZero))
	for i := range firstZero {
		firstZero[i], firstOne[i] = -1, -1
	}
	var pattern uint32
	// The native routine always examines one complete packed word when the
	// variable count is at most six.  Repeated small functions therefore have
	// the same operation accounting and counterexample choice as six-input
	// functions.
	for assignment := 0; assignment < 64; assignment++ {
		if err := budget.step(); err != nil {
			return RsbDecompositionCheck{}, err
		}
		divisorPattern := 0
		for bit, index := range selected {
			if rsbTruthBit(divisors[index], assignment) {
				divisorPattern |= 1 << uint(bit)
			}
		}
		if rsbTruthBit(truth, assignment) {
			pattern |= 1 << uint(divisorPattern+len(firstZero))
			if firstOne[divisorPattern] < 0 {
				firstOne[divisorPattern] = assignment
			}
		} else {
			pattern |= 1 << uint(divisorPattern)
			if firstZero[divisorPattern] < 0 {
				firstZero[divisorPattern] = assignment
			}
		}
	}
	overlap := (pattern >> uint(len(firstZero))) & pattern & (uint32(1)<<uint(len(firstZero)) - 1)
	if overlap != 0 {
		divisorPattern := bits.TrailingZeros32(overlap)
		return RsbDecompositionCheck{Counterexample: &RsbCounterexample{
			ZeroAssignment: firstZero[divisorPattern], OneAssignment: firstOne[divisorPattern],
		}}, nil
	}
	return RsbDecompositionCheck{Pattern: pattern}, nil
}

// CheckRsbDecomposition6 is the checked owned form of Rsb_DecCheck.
func CheckRsbDecomposition6(variables int, truth uint64, divisors []uint64, selected []int, options RsbResubOptions) (RsbDecompositionCheck, error) {
	if err := rsbValidateTruth(truth, variables); err != nil {
		return RsbDecompositionCheck{}, err
	}
	for _, divisor := range divisors {
		if err := rsbValidateTruth(divisor, variables); err != nil {
			return RsbDecompositionCheck{}, err
		}
	}
	budget, err := newRsbSearchBudget(options)
	if err != nil {
		return RsbDecompositionCheck{}, err
	}
	return checkRsbDecomposition6(variables, truth, divisors, slices.Clone(selected), budget)
}

func rsbCombinationSkipped(masks []uint64, selected []int) bool {
	common := ^uint64(0)
	for _, index := range selected {
		common &= masks[index]
	}
	return common != 0
}

func rsbRecordCounterexample(masks []uint64, divisors []uint64, counterexample RsbCounterexample, number int) {
	bit := uint64(1) << uint(number)
	for index, divisor := range divisors {
		if rsbTruthBit(divisor, counterexample.ZeroAssignment) == rsbTruthBit(divisor, counterexample.OneAssignment) {
			masks[index] |= bit
		}
	}
}

func rsbStretchHalf(pattern uint32, variables int, onset bool) (uint64, error) {
	shift := 1 << uint(variables)
	word := uint64(pattern)
	if onset {
		word >>= uint(shift)
	}
	if shift < 32 {
		word &= uint64(1)<<uint(shift) - 1
	}
	return NormalizeSmallTruthWord(word, variables)
}

func (m *RsbManager) performResub6(variables int, truth uint64, divisors []uint64, options RsbResubOptions) (RsbResub6Result, error) {
	if m == nil || m.closed {
		return RsbResub6Result{}, ErrInvariant
	}
	if _, err := rsbLogicalAssignments(variables); err != nil || variables > m.config.LeafMaximum {
		if err != nil {
			return RsbResub6Result{}, err
		}
		return RsbResub6Result{}, ErrLimit
	}
	if len(divisors) >= RsbMaxDivisors || len(divisors) > m.config.DivisorMaximum {
		return RsbResub6Result{}, ErrLimit
	}
	if err := rsbValidateTruth(truth, variables); err != nil {
		return RsbResub6Result{}, err
	}
	for _, divisor := range divisors {
		if err := rsbValidateTruth(divisor, variables); err != nil {
			return RsbResub6Result{}, err
		}
	}
	budget, err := newRsbSearchBudget(options)
	if err != nil {
		return RsbResub6Result{}, err
	}
	maximum := m.config.DecompositionMaximum
	if maximum == 0 || maximum > 4 {
		maximum = 4
	}
	result := RsbResub6Result{}
	masks := make([]uint64, len(divisors))
	exhausted := false
	try := func(selected []int) (bool, error) {
		if rsbCombinationSkipped(masks, selected) {
			return false, nil
		}
		check, checkErr := checkRsbDecomposition6(variables, truth, divisors, selected, budget)
		if checkErr != nil {
			return false, checkErr
		}
		if check.Counterexample == nil {
			result.Found = true
			result.Fanins = slices.Clone(selected)
			result.Pattern = check.Pattern
			result.Offset, checkErr = rsbStretchHalf(check.Pattern, len(selected), false)
			if checkErr == nil {
				result.Onset, checkErr = rsbStretchHalf(check.Pattern, len(selected), true)
			}
			return true, checkErr
		}
		if len(result.Counterexamples) == 64 {
			exhausted = true
			return false, nil
		}
		result.Counterexamples = append(result.Counterexamples, *check.Counterexample)
		rsbRecordCounterexample(masks, divisors, *check.Counterexample, len(result.Counterexamples)-1)
		if m.config.Verbose {
			result.Tries = append(result.Tries, slices.Clone(selected))
		}
		return false, nil
	}
	for i := 0; i < len(divisors) && maximum >= 1; i++ {
		if found, searchErr := try([]int{i}); searchErr != nil {
			return RsbResub6Result{}, searchErr
		} else if found {
			goto commit
		} else if exhausted {
			goto commit
		}
	}
	for i := 1; i < len(divisors) && maximum >= 2; i++ {
		for k := 0; k < i; k++ {
			if found, searchErr := try([]int{i, k}); searchErr != nil {
				return RsbResub6Result{}, searchErr
			} else if found {
				goto commit
			} else if exhausted {
				goto commit
			}
		}
	}
	for i := 2; i < len(divisors) && maximum >= 3; i++ {
		for k := 1; k < i; k++ {
			for j := 0; j < k; j++ {
				if found, searchErr := try([]int{i, k, j}); searchErr != nil {
					return RsbResub6Result{}, searchErr
				} else if found {
					goto commit
				} else if exhausted {
					goto commit
				}
			}
		}
	}
	for i := 3; i < len(divisors) && maximum >= 4; i++ {
		for k := 2; k < i; k++ {
			for j := 1; j < k; j++ {
				for n := 0; n < j; n++ {
					if found, searchErr := try([]int{i, k, j, n}); searchErr != nil {
						return RsbResub6Result{}, searchErr
					} else if found {
						goto commit
					} else if exhausted {
						goto commit
					}
				}
			}
		}
	}

commit:
	// Scratch and returned storage commit together; cancellation and limits do
	// not expose a partial search.
	m.counterexamples = append(m.counterexamples[:0], masks...)
	m.patterns = append(m.patterns[:0], result.Pattern)
	m.fanins = slices.Clone(result.Fanins)
	m.tries = m.tries[:0]
	for _, attempted := range result.Tries {
		m.tries = append(m.tries, attempted...)
		m.tries = append(m.tries, -1)
	}
	return result.Clone(), nil
}

// PerformResub6 is the owned form of Rsb_ManPerformResub6.
func (m *RsbManager) PerformResub6(variables int, truth uint64, divisors []uint64, options RsbResubOptions) (RsbResub6Result, error) {
	return m.performResub6(variables, truth, slices.Clone(divisors), options)
}

// VerifyRsbDecomposition checks the source's compatible offset/onset relation
// for every logical assignment.
func VerifyRsbDecomposition(variables int, truth uint64, divisors []uint64, result RsbResub6Result) (bool, error) {
	if !result.Found {
		return false, ErrDimension
	}
	if err := rsbValidateTruth(truth, variables); err != nil {
		return false, err
	}
	if len(result.Fanins) < 1 || len(result.Fanins) > 4 {
		return false, ErrDimension
	}
	seen := make([]bool, len(divisors))
	for _, index := range result.Fanins {
		if index < 0 || index >= len(divisors) || seen[index] {
			return false, ErrDimension
		}
		seen[index] = true
	}
	for _, divisor := range divisors {
		if err := rsbValidateTruth(divisor, variables); err != nil {
			return false, err
		}
	}
	if err := rsbValidateTruth(result.Offset, len(result.Fanins)); err != nil {
		return false, err
	}
	if err := rsbValidateTruth(result.Onset, len(result.Fanins)); err != nil {
		return false, err
	}
	check, err := CheckRsbDecomposition6(variables, truth, divisors, result.Fanins, RsbResubOptions{})
	if err != nil {
		return false, err
	}
	if check.Counterexample != nil || check.Pattern != result.Pattern {
		return false, nil
	}
	wantOffset, err := rsbStretchHalf(result.Pattern, len(result.Fanins), false)
	if err != nil {
		return false, err
	}
	wantOnset, err := rsbStretchHalf(result.Pattern, len(result.Fanins), true)
	if err != nil {
		return false, err
	}
	if result.Offset != wantOffset || result.Onset != wantOnset {
		return false, nil
	}
	assignments, _ := rsbLogicalAssignments(variables)
	for assignment := 0; assignment < assignments; assignment++ {
		pattern := 0
		for bit, index := range result.Fanins {
			if index < 0 || index >= len(divisors) {
				return false, ErrDimension
			}
			if rsbTruthBit(divisors[index], assignment) {
				pattern |= 1 << uint(bit)
			}
		}
		if rsbTruthBit(truth, assignment) {
			if !rsbTruthBit(result.Onset, pattern) {
				return false, nil
			}
		} else if !rsbTruthBit(result.Offset, pattern) {
			return false, nil
		}
	}
	return true, nil
}
