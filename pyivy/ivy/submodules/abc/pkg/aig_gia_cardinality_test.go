package abc

import (
	"encoding/json"
	"math/bits"
	"os"
	"slices"
	"testing"
)

func TestGiaCardinalityNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_cardinality.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		LogN, Variables int
		Outputs         []SatVar
		Clauses         [][]SatLit
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 8 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		r, e := BuildGiaCardinalityCNF(row.LogN, GiaWindowCNFOptions{})
		if e != nil || r.Variables != row.Variables || !slices.Equal(r.Outputs, row.Outputs) || !equalSatClauses(r.Clauses, row.Clauses) {
			t.Fatal(row.LogN, e)
		}
	}
}
func checkGiaCardinality(t *testing.T, logN int, mask uint64, k int) {
	t.Helper()
	r, e := BuildGiaCardinalityCNF(logN, GiaWindowCNFOptions{})
	if e != nil {
		t.Fatal(e)
	}
	s := NewSatSolver()
	if e = s.SetVariableCount(r.Variables); e != nil {
		t.Fatal(e)
	}
	for _, c := range r.Clauses {
		if _, e = s.AddClause(c); e != nil {
			t.Fatal(e)
		}
	}
	a := make([]SatLit, 0, r.Inputs+1)
	for i := 0; i < r.Inputs; i++ {
		l, _ := MakeSatLit(SatVar(i), mask>>uint(i)&1 == 0)
		a = append(a, l)
	}
	l, _ := MakeSatLit(r.Outputs[k], true)
	a = append(a, l)
	answer, e := s.Solve(a, SatLimits{})
	if e != nil {
		t.Fatal(e)
	}
	count := bits.OnesCount64(mask & ((uint64(1) << uint(r.Inputs)) - 1))
	expected := SatUnsatisfiable
	if count <= k {
		expected = SatSatisfiable
	}
	if answer.Status != expected {
		t.Fatal(logN, mask, k, count, answer.Status, expected)
	}
}
func TestGiaCardinalitySemantics(t *testing.T) {
	for logN := 0; logN <= 3; logN++ {
		n := 1 << uint(logN)
		for mask := uint64(0); mask < 1<<uint(n); mask++ {
			for k := 0; k < n; k++ {
				checkGiaCardinality(t, logN, mask, k)
			}
		}
	}
}
func TestGiaCardinalityLimits(t *testing.T) {
	for _, o := range []GiaWindowCNFOptions{{MaxVariables: 2}, {MaxClauses: 1}, {MaxLiterals: 1}, {MaxVariables: -1}, {MaxClauses: -1}, {MaxLiterals: -1}, {Stop: func() bool { return true }}} {
		r, e := BuildGiaCardinalityCNF(2, o)
		if e == nil || r.Outputs != nil || r.Clauses != nil {
			t.Fatal(r, e)
		}
	}
	for _, n := range []int{-1, 21} {
		if _, e := BuildGiaCardinalityCNF(n, GiaWindowCNFOptions{}); e == nil {
			t.Fatal(n)
		}
	}
}
func FuzzGiaCardinality(f *testing.F) {
	f.Add(uint8(3), uint64(37), uint8(2))
	f.Fuzz(func(t *testing.T, l uint8, m uint64, k uint8) {
		logN := int(l % 5)
		checkGiaCardinality(t, logN, m, int(k)%(1<<uint(logN)))
	})
}
