package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthPhaseNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_phase.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode     int
		Phase               uint64
		Source, Result      []uint64
		Compare, CompareRev []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 60 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.NormalizePhases()
		if e != nil {
			t.Fatal(e)
		}
		phase := uint64(0)
		if r.OutputPhase {
			phase |= uint64(1) << uint(row.Variables)
		}
		for i, flip := range r.InputPhases {
			if flip {
				phase |= uint64(1) << uint(i)
			}
		}
		if phase != row.Phase || !slices.Equal(r.Table.Words(), row.Result) {
			t.Fatal(row.Variables, row.Mode, phase, row.Phase, r.Table.Words(), row.Result)
		}
		applied, e := source.TransformNPN(nil, r.InputPhases, r.OutputPhase)
		if e != nil || !applied.Equal(r.Table) {
			t.Fatal("phase reconstruction", e)
		}
		for i := 0; i < row.Variables; i++ {
			forward, e := source.CompareCofactors(i, false)
			if e != nil {
				t.Fatal(e)
			}
			reverse, e := source.CompareCofactors(i, true)
			if e != nil || forward != row.Compare[i] || reverse != row.CompareRev[i] {
				t.Fatal(row.Variables, i, forward, reverse, e)
			}
		}
	}
}
func TestTruthPhaseOwnershipAndDimensions(t *testing.T) {
	source, _ := NewVariableTruthTable(8, 7, 0)
	before := source.Words()
	r, e := source.NormalizePhases()
	if e != nil {
		t.Fatal(e)
	}
	r.Table.Set(0, true)
	if !slices.Equal(before, source.Words()) {
		t.Fatal("alias")
	}
	for _, v := range []int{-1, 8} {
		if _, e := source.CompareCofactors(v, true); !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	if r, e := (*TruthTable)(nil).NormalizePhases(); r.Table != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzTruthPhase(f *testing.F) {
	f.Add(uint8(7), []byte{1, 2, 3, 4, 5})
	f.Add(uint8(0), []byte{255})
	f.Fuzz(func(t *testing.T, flags uint8, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 96 {
			data = data[:96]
		}
		variables := int(flags % 10)
		n := 1 << uint(variables)
		model := make([]bool, n)
		for i := range model {
			model[i] = data[(i/8)%len(data)]&(1<<uint(i%8)) != 0
		}
		count, _ := TruthTableWordCount(variables)
		words := make([]uint64, count)
		for i := 0; i < count*64; i++ {
			if model[i%n] {
				words[i/64] |= uint64(1) << uint(i%64)
			}
		}
		source, e := TruthTableFromWords(variables, words, 0)
		if e != nil {
			t.Fatal(e)
		}
		expected := slices.Clone(model)
		outputPhase := expected[n-1]
		if outputPhase {
			for i, v := range expected {
				expected[i] = !v
			}
		}
		phases := make([]bool, variables)
		for variable := variables - 1; variable >= 0; variable-- {
			flip := false
			for assignment := n - 1; assignment >= 0; assignment-- {
				candidate, current := expected[assignment^(1<<uint(variable))], expected[assignment]
				if candidate != current {
					flip = !candidate && current
					break
				}
			}
			if flip {
				next := make([]bool, n)
				for assignment := range next {
					next[assignment] = expected[assignment^(1<<uint(variable))]
				}
				expected = next
				phases[variable] = true
			}
		}
		r, e := source.NormalizePhases()
		if e != nil || r.OutputPhase != outputPhase || !slices.Equal(r.InputPhases, phases) {
			t.Fatal("phase metadata", r.InputPhases, phases, e)
		}
		for assignment := 0; assignment < n; assignment++ {
			got, e := r.Table.Get(uint64(assignment))
			if e != nil || got != expected[assignment] {
				t.Fatal("greedy result", variables, assignment)
			}
			original := assignment
			for v, flip := range r.InputPhases {
				if flip {
					original ^= 1 << uint(v)
				}
			}
			if got != (model[original] != r.OutputPhase) {
				t.Fatal("configuration")
			}
		}
		for variable := 0; variable < variables; variable++ {
			zero, _ := source.Cofactor(variable, false)
			one, _ := source.Cofactor(variable, true)
			for _, reverse := range []bool{false, true} {
				want, _ := zero.Compare(one, reverse)
				got, e := source.CompareCofactors(variable, reverse)
				if e != nil || got != want {
					t.Fatal("cofactor comparison", variable, reverse, got, want)
				}
			}
		}
		if !slices.Equal(words, source.Words()) {
			t.Fatal("source mutation")
		}
	})
}
