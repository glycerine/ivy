package goivy

import (
	"fmt"
	"strconv"
	"strings"
)

// CeilLog2 returns the ceiling of log2(n), i.e., the minimum number
// of bits needed to represent n distinct values.
// CeilLog2(0) = 0, CeilLog2(1) = 0, CeilLog2(2) = 1, CeilLog2(3) = 2, etc.
func MCCeilLog2(n int) int {
	bits, vals := 0, 1
	for vals < n {
		bits++
		vals *= 2
	}
	return bits
}

// GetEncodingBits returns the number of bits needed to binary-encode values
// of the given sort, using the provided interpretation map.
func GetEncodingBits(sort Sort, interp map[string]interface{}) (int, error) {
	th := GetSortTheory(sort, interp)
	switch t := th.(type) {
	case *LogicEnumeratedSort:
		return MCCeilLog2(len(t.Extension)), nil
	case *RangeSort:
		ub, err := strconv.Atoi(t.UbString())
		if err != nil {
			return 0, fmt.Errorf("invalid range sort upper bound: %s", t.UbString())
		}
		return MCCeilLog2(ub + 1), nil
	case *BooleanSort:
		return 1, nil
	case *Theory:
		if t.Kind == BitVectorKind {
			return t.Bits(), nil
		}
		return 0, fmt.Errorf("model checking cannot handle theory %s", t.Name)
	default:
		msg := fmt.Sprintf("model checking cannot handle sort %s", sort)
		return 0, fmt.Errorf("%s", msg)
	}
}

// GetEncodingBitsSimple returns the number of bits for a sort using
// simplified lookup, working with concrete sort types only.
func GetEncodingBitsSimple(sort Sort) (int, error) {
	switch t := sort.(type) {
	case *LogicEnumeratedSort:
		return MCCeilLog2(len(t.Extension)), nil
	case *RangeSort:
		ub, err := strconv.Atoi(t.UbString())
		if err != nil {
			return 0, fmt.Errorf("invalid range sort upper bound: %s", t.UbString())
		}
		return MCCeilLog2(ub + 1), nil
	case *BooleanSort:
		return 1, nil
	default:
		return 0, fmt.Errorf("model checking cannot handle sort %s", sort)
	}
}

// IsFiniteSort returns true if the sort is finite (enumerated, range,
// bit-vector, or boolean).
func IsFiniteSort(sort Sort) bool {
	if _, ok := sort.(*LogicFunctionSort); ok {
		return false
	}
	switch sort.(type) {
	case *LogicEnumeratedSort:
		return true
	case *RangeSort:
		return true
	case *BooleanSort:
		return true
	}
	return false
}

// IsFiniteSortWithInterp checks whether a sort is finite, considering
// theory interpretation.
func IsFiniteSortWithInterp(sort Sort, interp map[string]interface{}) bool {
	if _, ok := sort.(*LogicFunctionSort); ok {
		return false
	}
	switch sort.(type) {
	case *LogicEnumeratedSort:
		return true
	case *RangeSort:
		return true
	case *BooleanSort:
		return true
	}
	th := GetSortTheory(sort, interp)
	switch th.(type) {
	case *LogicEnumeratedSort:
		return true
	case *RangeSort:
		return true
	case *BooleanSort:
		return true
	}
	if t, ok := th.(*Theory); ok {
		return t.Kind == BitVectorKind
	}
	return false
}

// SortValues returns a list of string values for a finite sort.
// For enumerated sorts, the extension. For range sorts, the values from lb to ub.
// For boolean sorts, ["false", "true"].
func SortValues(sort Sort) ([]string, error) {
	switch t := sort.(type) {
	case *LogicEnumeratedSort:
		return t.Extension, nil
	case *RangeSort:
		lb, err := strconv.Atoi(t.LbString())
		if err != nil {
			return nil, fmt.Errorf("invalid range sort lower bound: %s", t.LbString())
		}
		ub, err := strconv.Atoi(t.UbString())
		if err != nil {
			return nil, fmt.Errorf("invalid range sort upper bound: %s", t.UbString())
		}
		vals := make([]string, 0, ub-lb+1)
		for n := lb; n <= ub; n++ {
			vals = append(vals, strconv.Itoa(n))
		}
		return vals, nil
	case *BooleanSort:
		return []string{"false", "true"}, nil
	default:
		return nil, fmt.Errorf("cannot enumerate values of sort %s", sort)
	}
}

// TermOrd compares two string-represented terms for ordering.
// Returns -1, 0, or 1.
func TermOrd(x, y string) int {
	if x < y {
		return -1
	}
	if x > y {
		return 1
	}
	return 0
}

// Normalize normalizes an expression string by ordering equality operands.
// If the expression is of the form "X = Y" and X > Y, swap them.
// If X = X, return "true".
func Normalize(expr string) string {
	if idx := strings.Index(expr, " = "); idx >= 0 {
		lhs := expr[:idx]
		rhs := expr[idx+3:]
		if lhs == rhs {
			return "true"
		}
		if TermOrd(lhs, rhs) > 0 {
			return rhs + " = " + lhs
		}
	}
	return expr
}

// GetTruth extracts a truth value from a digit string.
// '0' -> false, '1' -> true, 'x' -> unknown (empty string).
func GetTruth(digits string, idx int) (string, error) {
	if idx >= len(digits) {
		return "", fmt.Errorf("index %d out of range for digit string of length %d", idx, len(digits))
	}
	d := digits[idx]
	switch d {
	case '0':
		return "false", nil
	case '1':
		return "true", nil
	case 'x':
		return "", nil
	default:
		return "", fmt.Errorf("bad witness digit: %c", d)
	}
}
