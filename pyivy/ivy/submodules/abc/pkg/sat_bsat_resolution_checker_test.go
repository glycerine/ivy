package abc

// Stand-alone tests for src/sat/bsat/satChecker.c behavior.

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestResolveSatProofClauses(t *testing.T) {
	tests := []struct {
		first, second, want []int
	}{{[]int{1, 2, 3}, []int{-2, 4}, []int{1, 3, 4}}, {[]int{-7}, []int{7}, []int{}}, {[]int{18}, []int{-18}, []int{}}, {[]int{3, 1}, []int{-3, 1, 2}, []int{1, 2}}}
	for _, test := range tests {
		got, err := ResolveSatProofClauses(test.first, test.second)
		if err != nil || !reflect.DeepEqual(got, test.want) {
			t.Fatal(test, got, err)
		}
	}
	for _, test := range []struct{ first, second []int }{
		{[]int{1}, []int{2}},
		{[]int{1, 2}, []int{-1, -2}},
		{[]int{1, 1, 2}, []int{-2}},
		{[]int{0}, []int{1}},
		{[]int{mathMinInt()}, []int{1}},
	} {
		if got, err := ResolveSatProofClauses(test.first, test.second); got != nil || err == nil {
			t.Fatal(test, got, err)
		}
	}
	if got := FormatSatProofClause(7, []int{1, -2, 3}); got != "Clause 7:  { 1 -2 3 }\n" {
		t.Fatal(got)
	}
}

const satResolutionProofGood = "1 1 2 0 0\n" +
	"2 -1 0 0\n" +
	"3 * 1 2 0\n" +
	"4 -2 0 0\n" +
	"5 * 3 4 0\n"

func TestCheckSatResolutionProof(t *testing.T) {
	report, err := CheckSatResolutionProof([]byte(satResolutionProofGood), SatResolutionProofLimits{})
	if err != nil || report.Roots != 3 || report.ResolutionSteps != 2 || !report.DerivedEmpty || len(report.Clauses) != 6 {
		t.Fatal(report, err)
	}
	want := [][]int{nil, {1, 2}, {-1}, {2}, {-2}, {}}
	if !reflect.DeepEqual(report.Clauses, want) {
		t.Fatal(report.Clauses, want)
	}
	withoutNewline := strings.TrimSuffix(satResolutionProofGood, "\n")
	report2, err := CheckSatResolutionProof([]byte(withoutNewline), SatResolutionProofLimits{})
	if err != nil || !reflect.DeepEqual(report, report2) {
		t.Fatal(report2, err)
	}
	nonempty, err := CheckSatResolutionProof([]byte("1 1 0 0\n"), SatResolutionProofLimits{})
	if err != nil || nonempty.DerivedEmpty || nonempty.Roots != 1 {
		t.Fatal(nonempty, err)
	}
}

func TestCheckSatResolutionProofErrorsLimitsAndFile(t *testing.T) {
	cases := []string{
		"",
		"1 1 0\n",
		"2 1 0 0\n",
		"1 0 1 0 0\n",
		"1 x 0 0\n",
		"1 1 0 0\n2 * 1 3 0\n",
		"1 1 0 0\n2 -1 0 0\n3 * 1 2 1\n",
		"1 1 2 0 0\n2 -1 -2 0 0\n3 * 1 2 0\n",
		"1 1 0 0\n\n",
	}
	for _, input := range cases {
		if report, err := CheckSatResolutionProof([]byte(input), SatResolutionProofLimits{}); err == nil {
			t.Fatal("accepted", input, report)
		}
	}
	for _, limits := range []SatResolutionProofLimits{
		{MaxBytes: 4}, {MaxClauses: 2}, {MaxLiterals: 1}, {MaxLineBytes: 4}, {MaxBytes: -1},
	} {
		if _, err := CheckSatResolutionProof([]byte(satResolutionProofGood), limits); err == nil {
			t.Fatal(limits)
		}
	}
	directory := t.TempDir()
	name := filepath.Join(directory, "proof.txt")
	if err := os.WriteFile(name, []byte(satResolutionProofGood), 0o600); err != nil {
		t.Fatal(err)
	}
	report, err := CheckSatResolutionProofFile(name, SatResolutionProofLimits{})
	if err != nil || !report.DerivedEmpty {
		t.Fatal(report, err)
	}
	if _, err = CheckSatResolutionProofFile(name, SatResolutionProofLimits{MaxBytes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = CheckSatResolutionProofFile(filepath.Join(directory, "missing"), SatResolutionProofLimits{}); !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}

func FuzzResolveSatProofClauses(f *testing.F) {
	f.Add([]byte{1, 2, 3}, []byte{2, 4})
	f.Fuzz(func(t *testing.T, leftData, rightData []byte) {
		if len(leftData) > 32 {
			leftData = leftData[:32]
		}
		if len(rightData) > 32 {
			rightData = rightData[:32]
		}
		toClause := func(data []byte, negate bool) []int {
			result := make([]int, len(data))
			for i, value := range data {
				literal := int(value%31) + 1
				if negate && value&1 != 0 {
					literal = -literal
				}
				result[i] = literal
			}
			return result
		}
		left, right := toClause(leftData, true), toClause(rightData, true)
		got, err := ResolveSatProofClauses(left, right)
		pairs, pivot := 0, 0
		for _, a := range left {
			for _, b := range right {
				if a == -b {
					pairs, pivot = pairs+1, a
				}
			}
		}
		if pairs != 1 {
			if err == nil {
				t.Fatal("accepted pivot count", pairs, left, right, got)
			}
			return
		}
		seen := make(map[int]bool)
		var want []int
		for _, literal := range left {
			if literal != pivot && literal != -pivot && !seen[literal] {
				seen[literal] = true
				want = append(want, literal)
			}
		}
		firstUnique := len(want)
		for _, literal := range right {
			if literal != pivot && literal != -pivot && !seen[literal] {
				seen[literal] = true
				want = append(want, literal)
			}
		}
		validFirst := firstUnique+1 == len(left)
		if validFirst != (err == nil) || err == nil && !slices.Equal(got, want) {
			t.Fatal(left, right, got, want, validFirst, err)
		}
	})
}

func FuzzCheckSatResolutionProof(f *testing.F) {
	f.Add([]byte(satResolutionProofGood))
	f.Add([]byte("1 1 0 0\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 4096 {
			data = data[:4096]
		}
		report, err := CheckSatResolutionProof(data, SatResolutionProofLimits{MaxBytes: 4096, MaxClauses: 128, MaxLiterals: 1024, MaxLineBytes: 1024})
		if err == nil {
			if len(report.Clauses) != report.Roots+report.ResolutionSteps+1 || report.DerivedEmpty != (len(report.Clauses[len(report.Clauses)-1]) == 0) {
				t.Fatal(report)
			}
		}
	})
}
