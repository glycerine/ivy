package abc

// Go-only assignment-level checks for NPN enumerator pruning analysis.

import (
	"errors"
	"slices"
	"testing"
)

func TestAnalyzeDauNPNKnownFunctions(t *testing.T) {
	tests := []struct {
		name        string
		table       *TruthTable
		pairs       int
		second      []bool
		complements []bool
	}{
		{
			name:  "and3",
			table: dauTruthFromPredicate(t, 3, func(a uint64) bool { return a&7 == 7 }),
			pairs: 3, second: []bool{false, true, true}, complements: []bool{false, false, false},
		},
		{
			name:  "xor3",
			table: dauTruthFromPredicate(t, 3, func(a uint64) bool { return a&1 != 0 != (a&2 != 0) != (a&4 != 0) }),
			pairs: 3, second: []bool{false, true, true}, complements: []bool{true, true, true},
		},
		{
			name:  "variable",
			table: dauTruthFromPredicate(t, 3, func(a uint64) bool { return a&1 != 0 }),
			pairs: 1, second: []bool{false, false, true}, complements: []bool{true, true, true},
		},
	}
	for _, test := range tests {
		result, err := AnalyzeDauNPN(test.table, DauNPNAnalysisOptions{})
		if err != nil || result.SymmetricPairs != test.pairs || !slices.Equal(result.SymmetricSecondVariables, test.second) || !slices.Equal(result.ComplementEquivalent, test.complements) {
			t.Fatal(test.name, result, err)
		}
		mask, err := result.SymmetricSecondMask()
		if err != nil {
			t.Fatal(err)
		}
		wantMask := uint64(0)
		for variable, yes := range test.second {
			if yes {
				wantMask |= 1 << uint(variable)
			}
		}
		if mask != wantMask || result.ComplementCount() != countTrue(test.complements) {
			t.Fatal(test.name, mask, wantMask, result.ComplementCount())
		}
	}
}

func countTrue(values []bool) int {
	result := 0
	for _, value := range values {
		if value {
			result++
		}
	}
	return result
}

func TestAnalyzeDauNPNLimitsAndOwnership(t *testing.T) {
	table := dauTruthFromPredicate(t, 4, func(a uint64) bool { return a&15 == 15 })
	if result, err := AnalyzeDauNPN(table, DauNPNAnalysisOptions{MaxComparisons: 1}); result.SymmetricSecondVariables != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := AnalyzeDauNPN(table, DauNPNAnalysisOptions{Stop: func() bool { return true }}); result.SymmetricSecondVariables != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	if result, err := AnalyzeDauNPN(nil, DauNPNAnalysisOptions{}); result.SymmetricSecondVariables != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	result, err := AnalyzeDauNPN(table, DauNPNAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result.SymmetricSecondVariables[1] = false
	again, err := AnalyzeDauNPN(table, DauNPNAnalysisOptions{})
	if err != nil || !again.SymmetricSecondVariables[1] {
		t.Fatal(again, err)
	}
	wide := DauNPNAnalysis{SymmetricSecondVariables: make([]bool, 65)}
	if _, err := wide.SymmetricSecondMask(); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func FuzzAnalyzeDauNPN(f *testing.F) {
	f.Add(uint8(5), []byte{0x96, 0x69})
	f.Add(uint8(7), []byte{0x31, 0x72, 0x55, 0xaa})
	f.Fuzz(func(t *testing.T, width uint8, data []byte) {
		variables := 1 + int(width%8)
		if len(data) == 0 {
			return
		}
		values := make([]bool, 1<<uint(variables))
		for assignment := range values {
			values[assignment] = data[assignment%len(data)]&(1<<uint(assignment&7)) != 0
		}
		table := dauTruthFromPredicate(t, variables, func(a uint64) bool { return values[a] })
		result, err := AnalyzeDauNPN(table, DauNPNAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		pairs := 0
		secondMask := make([]bool, variables)
		for first := 0; first < variables; first++ {
			for second := first + 1; second < variables; second++ {
				symmetric := true
				for assignment := range values {
					swapped := assignment
					if ((assignment>>uint(first))^(assignment>>uint(second)))&1 != 0 {
						swapped ^= (1 << uint(first)) | (1 << uint(second))
					}
					if values[assignment] != values[swapped] {
						symmetric = false
						break
					}
				}
				if symmetric {
					pairs++
					secondMask[second] = true
				}
			}
		}
		complementEquivalent := make([]bool, variables)
		for variable := 0; variable < variables; variable++ {
			outputComplement := true
			for assignment := range values {
				if values[assignment^(1<<uint(variable))] == values[assignment] {
					outputComplement = false
					break
				}
			}
			if outputComplement {
				complementEquivalent[variable] = true
				continue
			}
			for other := 0; other < variables; other++ {
				if other == variable {
					continue
				}
				equal := true
				for assignment := range values {
					if values[assignment^(1<<uint(variable))] != values[assignment^(1<<uint(other))] {
						equal = false
						break
					}
				}
				if equal {
					complementEquivalent[variable] = true
					break
				}
			}
		}
		if result.SymmetricPairs != pairs || !slices.Equal(result.SymmetricSecondVariables, secondMask) || !slices.Equal(result.ComplementEquivalent, complementEquivalent) {
			t.Fatal(result, pairs, secondMask, complementEquivalent)
		}
	})
}
