package abc

// Pairwise DSD divisor statistics from opt/dau/dauDivs.c:Dau_DsdDivisors.

import (
	"fmt"
	"strings"
)

type DauDivisorOptions struct {
	MaxPairs, MaxDigits int // zero selects one million pairs / 16 million digits
	Stop                func() bool
}

// DauDivisorPair records counts for !a*!b, !a*b, a*!b, a*b and a XOR b,
// respectively. A function and its complement contribute to the same count.
type DauDivisorPair struct {
	First, Second int
	Counts        [5]int
}

// DauDSDDivisors returns the source routine's pairwise four-bit cofactor
// classification without writing stdout. Results and temporary storage are
// owned. Work limits and cancellation return no partial rows.
func DauDSDDivisors(t *TruthTable, o DauDivisorOptions) ([]DauDivisorPair, error) {
	if t == nil || o.MaxPairs < 0 || o.MaxDigits < 0 {
		return nil, ErrDimension
	}
	if e := t.Check(); e != nil {
		return nil, e
	}
	if o.MaxPairs == 0 {
		o.MaxPairs = 1 << 20
	}
	if o.MaxDigits == 0 {
		o.MaxDigits = 1 << 24
	}
	variables := t.variables
	pairs := 0
	if variables >= 2 {
		var e error
		pairs, e = checkedSize(0, variables, variables-1, 2*o.MaxPairs)
		if e != nil {
			return nil, e
		}
		pairs /= 2
	}
	if pairs > o.MaxPairs {
		return nil, ErrLimit
	}
	digits := 1
	if variables > 2 {
		digits = 1 << uint(variables-2)
	}
	if _, e := checkedSize(0, pairs, digits, o.MaxDigits); e != nil {
		return nil, e
	}
	result := make([]DauDivisorPair, 0, pairs)
	for first := 0; first < variables; first++ {
		for second := first + 1; second < variables; second++ {
			if o.Stop != nil && o.Stop() {
				return nil, ErrCancelled
			}
			copy := t.Clone()
			copy.swapVariables(0, first)
			copy.swapVariables(1, second)
			row := DauDivisorPair{First: first, Second: second}
			for digitIndex := 0; digitIndex < digits; digitIndex++ {
				if o.Stop != nil && o.Stop() {
					return nil, ErrCancelled
				}
				digit := int(copy.words[digitIndex>>4]>>uint((digitIndex<<2)&63)) & 15
				switch digit {
				case 1, 14:
					row.Counts[0]++
				case 2, 13:
					row.Counts[1]++
				case 4, 11:
					row.Counts[2]++
				case 8, 7:
					row.Counts[3]++
				case 6, 9:
					row.Counts[4]++
				}
			}
			result = append(result, row)
		}
	}
	return result, nil
}

// FormatDauDSDDivisors preserves the source diagnostic columns while returning
// an owned deterministic string.
func FormatDauDSDDivisors(rows []DauDivisorPair) string {
	var out strings.Builder
	out.WriteString("      !a *!b !a * b  a *!b  a * b  a + b\n")
	for _, row := range rows {
		fmt.Fprintf(&out, "%c %c  ", byte('a'+row.First), byte('a'+row.Second))
		for _, count := range row.Counts {
			fmt.Fprintf(&out, "%7d", count)
		}
		out.WriteByte('\n')
	}
	return out.String()
}

// DauDSDTest000 preserves the source demonstration entry point's observable
// behavior. Its active C statements compute discarded locals and return before
// invoking divisor analysis, so the concrete Go operation has no side effects.
func DauDSDTest000() {}
