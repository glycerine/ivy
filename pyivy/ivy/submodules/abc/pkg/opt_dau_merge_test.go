package abc

// Go-only unit and semantic fuzz checks for complete owned DSD merging.

import (
	"errors"
	"testing"
)

func identityDauPermutation(variables int) []int {
	result := make([]int, variables)
	for i := range result {
		result[i] = i
	}
	return result
}

func TestDauDSDMergeBasicAndConstants(t *testing.T) {
	tests := []struct {
		first, second    string
		complementFirst  bool
		complementSecond bool
		variables        int
		want             string
	}{
		{"a", "b", false, false, 2, "(ab)"},
		{"a", "a", false, false, 1, "a"},
		{"(ab)", "(ac)", false, false, 3, "(abc)"},
		{"0", "a", false, false, 1, "0"},
		{"0", "a", true, false, 1, "a"},
		{"1", "a", true, false, 1, "0"},
		{"a", "1", true, false, 1, "!a"},
	}
	for _, test := range tests {
		permutation := identityDauPermutation(test.variables)
		got, err := DauDSDMerge(test.first, permutation, test.second, permutation, test.complementFirst, test.complementSecond, test.variables, DauMergeOptions{})
		if err != nil || got != test.want {
			t.Fatalf("DauDSDMerge(%q,%q) = %q, %v; want %q", test.first, test.second, got, err, test.want)
		}
	}
	got, err := DauDSDMerge("a", []int{1, 0}, "a", []int{0, 1}, false, false, 2, DauMergeOptions{})
	if err != nil || got != "(ab)" {
		t.Fatal(got, err)
	}
}

func TestDauDSDMergeGeneralDecompositionAndErrors(t *testing.T) {
	primeA := "69969669966969969669966969969669{abcdefg}"
	primeB := "E8E8E8E817171717E8E8E8E817171717{abcdefg}"
	permutation := identityDauPermutation(7)
	result, err := DauDSDMerge(primeA, permutation, primeB, permutation, false, false, 7, DauMergeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseDauDSD(result, DauDSDOptions{})
	if err != nil || parsed.VariableCount() != 7 {
		t.Fatal(result, err)
	}
	if got, err := DauDSDMerge(primeA, permutation, primeB, permutation, false, false, 7, DauMergeOptions{MaxWords: 1}); got != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(got, err)
	}
	badPermutations := [][]int{nil, {0, 0}, {0, 2}}
	for _, bad := range badPermutations {
		if got, err := DauDSDMerge("a", bad, "b", []int{0, 1}, false, false, 2, DauMergeOptions{}); got != "" || !errors.Is(err, ErrDimension) {
			t.Fatal(bad, got, err)
		}
	}
	if got, err := DauDSDMerge("a", []int{0}, "!a", []int{0}, false, false, 1, DauMergeOptions{MaxBytes: 1}); got != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(got, err)
	}
	if got, err := DauDSDMerge("a", []int{0}, "a", []int{0}, false, false, 1, DauMergeOptions{Stop: func() bool { return true }}); got != "" || !errors.Is(err, ErrCancelled) {
		t.Fatal(got, err)
	}
	if got, err := DauDSDMerge("(ab)", []int{0, 1}, "a", []int{0, 1}, false, false, 2, DauMergeOptions{MaxNodes: 1}); got != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(got, err)
	}
	if got, err := DauDSDMerge("(ab)", []int{0, 1}, "a", []int{0, 1}, false, false, 2, DauMergeOptions{MaxDepth: 1}); got != "" || !errors.Is(err, ErrLimit) {
		t.Fatal(got, err)
	}
	if got, err := DauDSDMerge("a", []int{0, 1}, "b", []int{0, 1}, false, false, 2, DauMergeOptions{MaxVariables: 1}); got != "" || !errors.Is(err, ErrDimension) {
		t.Fatal(got, err)
	}
	if got, err := DauDSDMerge("(a)", []int{0}, "a", []int{0}, false, false, 1, DauMergeOptions{}); got != "" || err == nil {
		t.Fatal(got, err)
	}
}

func FuzzDauDSDMerge(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3}, []byte{4, 5, 6}, uint8(1), uint8(2), false, true)
	f.Add([]byte{9, 8, 7, 6, 5}, []byte{1, 3, 5, 7}, uint8(0), uint8(0), true, false)
	f.Fuzz(func(t *testing.T, firstData, secondData []byte, firstShift, secondShift uint8, complementFirst, complementSecond bool) {
		if len(firstData) > 64 {
			firstData = firstData[:64]
		}
		if len(secondData) > 64 {
			secondData = secondData[:64]
		}
		first, err := ParseDauDSD(fuzzDauFormula(firstData), DauDSDOptions{})
		if err != nil {
			return
		}
		second, err := ParseDauDSD(fuzzDauFormula(secondData), DauDSDOptions{})
		if err != nil {
			return
		}
		variables := max(first.VariableCount(), second.VariableCount())
		if variables == 0 || variables > 8 {
			return
		}
		firstPermutation := make([]int, variables)
		secondPermutation := make([]int, variables)
		for i := 0; i < variables; i++ {
			firstPermutation[i] = (i + int(firstShift)%variables) % variables
			secondPermutation[i] = (i + int(secondShift)%variables) % variables
		}
		result, err := DauDSDMerge(first.String(), firstPermutation, second.String(), secondPermutation, complementFirst, complementSecond, variables, DauMergeOptions{MaxBytes: 2000, MaxNodes: 2000, MaxVariables: 12, MaxWords: 64})
		if err != nil {
			t.Fatal(first.String(), second.String(), err)
		}
		merged, err := ParseDauDSD(result, DauDSDOptions{})
		if err != nil {
			t.Fatal(result, err)
		}
		for assignment := 0; assignment < 1<<uint(variables); assignment++ {
			firstAssignment, secondAssignment := uint64(0), uint64(0)
			for source := 0; source < variables; source++ {
				if assignment&(1<<uint(firstPermutation[source])) != 0 {
					firstAssignment |= 1 << uint(source)
				}
				if assignment&(1<<uint(secondPermutation[source])) != 0 {
					secondAssignment |= 1 << uint(source)
				}
			}
			left, err := first.eval(first.root, firstAssignment, nil)
			if err != nil {
				t.Fatal(err)
			}
			right, err := second.eval(second.root, secondAssignment, nil)
			if err != nil {
				t.Fatal(err)
			}
			if complementFirst {
				left = !left
			}
			if complementSecond {
				right = !right
			}
			got, err := merged.eval(merged.root, uint64(assignment), nil)
			if err != nil || got != left && right {
				t.Fatal("merge mismatch", first.String(), second.String(), result, assignment, left, right, got, err)
			}
		}
	})
}
