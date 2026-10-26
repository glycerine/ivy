package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func checkGiaSupportReduction(t *testing.T, m *GiaManager, r GiaSupportReduction) {
	t.Helper()
	for assignment := 0; assignment < 1<<m.InputCount(); assignment++ {
		input := make([]bool, m.InputCount())
		for i := range input {
			input[i] = assignment&(1<<i) != 0
		}
		want, e := m.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		for output, support := range r.Supports {
			projected := make([]bool, r.Graph.InputCount())
			for i, ci := range support {
				projected[i] = input[ci]
			}
			got, e := r.Graph.Evaluate(projected)
			if e != nil || got[r.Outputs[output]] != want[output] {
				t.Fatal(assignment, output, got, want, e)
			}
		}
	}
}
func TestGiaSupportReductionNative(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(b, c)
	m.AppendCO(x)
	m.AppendCO(y)
	m.AppendCO(y.Not())
	m.AppendCO(c)
	m.AppendCO(GiaConst0)
	raw, e := os.ReadFile("test_vectors/oracle/gia_support_reduce.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects, Inputs, Outputs int
		Map                      []int
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 2 {
		t.Fatal(e)
	}
	for mode, row := range rows {
		r, e := m.ReduceBySupport(mode != 0, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if r.Graph.ObjectCount() != row.Objects || r.Graph.InputCount() != row.Inputs || r.Graph.OutputCount() != row.Outputs || !slices.Equal(r.Outputs, row.Map) {
			t.Fatal(mode, r, row)
		}
		checkGiaSupportReduction(t, m, r)
	}
}
func TestGiaSupportReductionExtendedAndLimits(t *testing.T) {
	m := NewGiaManager()
	m.Name = "reduce"
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	buf, _ := m.AppendBuffer(a.Not())
	x, _ := m.AppendXorReal(a, b)
	mux, _ := m.AppendMuxReal(c, x, buf)
	m.AppendCO(mux)
	m.AppendCO(buf)
	m.AppendCO(GiaConst1)
	before := slices.Clone(m.objects)
	r, e := m.ReduceBySupport(true, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	checkGiaSupportReduction(t, m, r)
	if !slices.Equal(before, m.objects) || r.Graph.Name != m.Name {
		t.Fatal("source/metadata")
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.ReduceBySupport(true, options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 12} {
		calls := 0
		r, e := m.ReduceBySupport(true, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	m.SetConstraintCount(1)
	if _, e := m.ReduceBySupport(true, GiaCofactorOptions{}); e == nil {
		t.Fatal("constraint")
	}
}
