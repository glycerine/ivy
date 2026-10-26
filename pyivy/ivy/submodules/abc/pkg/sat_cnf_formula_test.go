package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func formulaFlat(f *CNFFormula) []SatLit {
	var out []SatLit
	for _, c := range f.clauses {
		out = append(out, c...)
	}
	return out
}
func TestCNFFlipPositionsNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_flip_positions.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variable             SatVar
		Base                 SatVar
		Mode                 int
		Collected, Positions []int
		Flat                 []SatLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 84 {
		t.Fatal(len(rows))
	}
	source := [][]SatLit{{0, 3, 4}, {}, {7, 8, 1, 2}, {5, 6, 9}}
	f, e := NewCNFFormula(5, source, CNFTransformOptions{})
	if e != nil {
		t.Fatal(e)
	}
	original := formulaFlat(f)
	for _, row := range rows {
		got, e := f.CollectFlipPositions(row.Variable, CNFTransformOptions{})
		if e != nil || !slices.Equal(got, row.Collected) {
			t.Fatal(row, got, e)
		}
		out, e := f.LiftAndFlipPositions(row.Base, row.Positions, CNFTransformOptions{})
		if e != nil || !slices.Equal(formulaFlat(out), row.Flat) {
			t.Fatal(row, out, e)
		}
		span := 5
		for _, l := range row.Flat {
			span = max(span, int(l.Var())+1)
		}
		if out.variables != span {
			t.Fatal(span, out.variables)
		}
		for i := range source {
			if len(out.clauses[i]) != len(source[i]) {
				t.Fatal("clause boundary changed")
			}
		}
		if !slices.Equal(original, formulaFlat(f)) {
			t.Fatal("source changed")
		}
	}
}
func TestCNFFormulaOwnershipAndBounds(t *testing.T) {
	source := [][]SatLit{{0, 3}, {4}}
	f, e := NewCNFFormula(3, source, CNFTransformOptions{})
	if e != nil {
		t.Fatal(e)
	}
	source[0][0] = 9
	copy := f.Clauses()
	copy[0][0] = 9
	if f.clauses[0][0] != 0 {
		t.Fatal("alias")
	}
	for _, positions := range [][]int{{-1}, {3}, {0, 3}} {
		if out, e := f.LiftAndFlipPositions(0, positions, CNFTransformOptions{}); out != nil || e != ErrDimension {
			t.Fatal(out, e)
		}
	}
	for _, o := range []CNFTransformOptions{{MaxVariables: 3}, {MaxLiterals: 2}, {Stop: func() bool { return true }}} {
		out, e := f.LiftAndFlipPositions(3, []int{0}, o)
		if out != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(out, e)
		}
	}
	empty, e := NewCNFFormula(0, nil, CNFTransformOptions{})
	if e != nil {
		t.Fatal(e)
	}
	ok, e := empty.LoadInto(NewSatSolver(), 0)
	if e != nil || !ok {
		t.Fatal(ok, e)
	}
	unsat, e := NewCNFFormula(0, [][]SatLit{{}}, CNFTransformOptions{})
	if e != nil {
		t.Fatal(e)
	}
	ok, e = unsat.LoadInto(NewSatSolver(), 0)
	if e != nil || ok {
		t.Fatal(ok, e)
	}
	if _, e = NewCNFFormula(0, [][]SatLit{{0}}, CNFTransformOptions{}); e != ErrInvalidLiteral {
		t.Fatal(e)
	}
	out, e := f.LiftAndFlipPositions(0, []int{0, 0}, CNFTransformOptions{})
	if e != nil || !equalSatClauses(out.clauses, f.clauses) {
		t.Fatal(out, e)
	}
}
func FuzzCNFFlipPositions(f *testing.F) {
	f.Add([]byte{0, 3, 4, 7, 8, 1, 2, 5, 6, 9}, uint64(12345), uint8(2))
	f.Fuzz(func(t *testing.T, data []byte, seed uint64, base uint8) {
		if len(data) > 80 {
			data = data[:80]
		}
		flat := make([]SatLit, len(data))
		for i, b := range data {
			flat[i] = SatLit(b % 10)
		}
		middle := len(flat) / 2
		clauses := [][]SatLit{flat[:middle], {}, flat[middle:]}
		source, e := NewCNFFormula(5, clauses, CNFTransformOptions{})
		if e != nil {
			t.Fatal(e)
		}
		var positions []int
		if len(flat) > 0 {
			for i := 0; i < 8; i++ {
				positions = append(positions, int(seed%uint64(len(flat))))
				seed = seed*1664525 + 1013904223
			}
		}
		before := formulaFlat(source)
		want := slices.Clone(before)
		for _, p := range positions {
			want[p] = (want[p] ^ 1) + SatLit(base%4)*2
		}
		out, e := source.LiftAndFlipPositions(SatVar(base%4), positions, CNFTransformOptions{})
		if e != nil || !slices.Equal(formulaFlat(out), want) || !slices.Equal(formulaFlat(source), before) {
			t.Fatal(out, e)
		}
	})
}

func TestCNFFormulaMidOperationCancellation(t *testing.T) {
	source, e := NewCNFFormula(3, [][]SatLit{{0, 3}, {4}}, CNFTransformOptions{})
	if e != nil {
		t.Fatal(e)
	}
	before := formulaFlat(source)
	calls := 0
	out, e := source.LiftAndFlipPositions(1, []int{0, 0}, CNFTransformOptions{Stop: func() bool { calls++; return calls > 6 }})
	if out != nil || e != ErrCancelled || !slices.Equal(before, formulaFlat(source)) {
		t.Fatal(out, e, calls)
	}
	if out, e = source.LiftAndFlipPositions(0, []int{0, 0, 0, 0}, CNFTransformOptions{MaxLiterals: 3}); out != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(out, e)
	}
}
