package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func darBuildGraph(t testing.TB, ops [][2]uint32, outputs []uint32) *AigManager {
	t.Helper()
	g := NewAigManager()
	pool := []AigLit{0}
	for i := 0; i < 6; i++ {
		v, _ := g.CreateCI()
		pool = append(pool, v)
	}
	for _, op := range ops {
		a, b := op[0]%uint32(2*len(pool)), op[1]%uint32(2*len(pool))
		v, e := g.And(pool[a>>1].NotIf(a&1 != 0), pool[b>>1].NotIf(b&1 != 0))
		if e != nil {
			t.Fatal(e)
		}
		pool = append(pool, v)
	}
	for _, out := range outputs {
		out %= uint32(2 * len(pool))
		if _, e := g.CreateCO(pool[out>>1].NotIf(out&1 != 0)); e != nil {
			t.Fatal(e)
		}
	}
	return g
}
func darCheckSemantics(t testing.TB, r *DARCuts4) {
	t.Helper()
	g := r.Graph
	for assignment := 0; assignment < 64; assignment++ {
		values := make([]bool, len(g.objects))
		values[0] = true
		for i, id := range g.cis {
			values[id] = assignment>>uint(i)&1 != 0
		}
		value := func(l AigLit) bool { return values[l.ID()] != l.Complemented() }
		for id, obj := range g.objects {
			if obj.Type == AigAND {
				values[id] = value(obj.Fanin0) && value(obj.Fanin1)
			}
		}
		for _, c := range r.Candidates {
			index := 0
			for i, leaf := range c.Leaves {
				if values[leaf] {
					index |= 1 << uint(i)
				}
			}
			if (c.Truth>>uint(index)&1 != 0) != values[c.Root] {
				t.Fatal(assignment, c)
			}
		}
	}
}
func TestDARCutsNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/dar_cuts.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []struct {
		Seed, Capacity, Skip, Objects int
		Ops                           [][2]uint32
		Outputs                       []uint32
		Cuts                          []AigCNFCut4
	}
	if e = json.Unmarshal(data, &cases); e != nil {
		t.Fatal(e)
	}
	if len(cases) != 192 {
		t.Fatal(len(cases))
	}
	for _, c := range cases {
		g := darBuildGraph(t, c.Ops, c.Outputs)
		before := slices.Clone(g.objects)
		r, e := ComputeDARCuts4(g, DARCutOptions{CutsPerNode: c.Capacity, SkipTruthMinimization: c.Skip != 0})
		if e != nil {
			t.Fatal(c.Seed, c.Capacity, c.Skip, e)
		}
		if len(r.Graph.objects) != c.Objects || len(r.Candidates) != len(c.Cuts) {
			t.Fatal(c.Seed, c.Capacity, c.Skip, "count", len(r.Graph.objects), c.Objects, len(r.Candidates), len(c.Cuts))
		}
		for i, w := range c.Cuts {
			v := r.Candidates[i]
			if v.Root != w.Root || v.Truth != w.Truth || !slices.Equal(v.Leaves, w.Leaves) {
				t.Fatal(c.Seed, c.Capacity, c.Skip, i, v, w)
			}
		}
		if !slices.Equal(before, g.objects) {
			t.Fatal("mutated graph")
		}
		darCheckSemantics(t, r)
	}
}
func TestDARCutsLimits(t *testing.T) {
	g := darBuildGraph(t, [][2]uint32{{2, 4}, {6, 8}, {14, 16}}, []uint32{18})
	for _, o := range []DARCutOptions{{MaxObjects: 1}, {MaxCutSlots: 1}, {MaxPairs: 1}, {Stop: func() bool { return true }}} {
		r, e := ComputeDARCuts4(g, o)
		if r != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(r, e)
		}
	}
	for _, o := range []DARCutOptions{{CutsPerNode: 1}, {CutsPerNode: 256}, {MaxPairs: -1}} {
		r, e := ComputeDARCuts4(g, o)
		if r != nil || e != ErrDimension {
			t.Fatal(r, e)
		}
	}
}
func FuzzDARCuts(f *testing.F) {
	f.Add([]byte{2, 4, 6, 8, 14, 16}, uint8(10), false)
	f.Fuzz(func(t *testing.T, data []byte, capacity uint8, skip bool) {
		if len(data) > 160 {
			data = data[:160]
		}
		var ops [][2]uint32
		for i := 0; i+1 < len(data); i += 2 {
			ops = append(ops, [2]uint32{uint32(data[i]), uint32(data[i+1])})
		}
		g := darBuildGraph(t, ops, []uint32{uint32(2 * (6 + len(ops)))})
		r, e := ComputeDARCuts4(g, DARCutOptions{CutsPerNode: 2 + int(capacity%14), SkipTruthMinimization: skip})
		if e != nil {
			t.Fatal(e)
		}
		darCheckSemantics(t, r)
	})
}

func TestDeriveAigCNFNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/dar_cuts.json")
	if e != nil {
		t.Fatal(e)
	}
	var cases []struct {
		Seed    int
		Ops     [][2]uint32
		Outputs []uint32
		CNF     *struct {
			Variables       int
			ObjectVariables []int
			Clauses         [][]SatLit
		}
	}
	if e = json.Unmarshal(data, &cases); e != nil {
		t.Fatal(e)
	}
	count := 0
	for _, row := range cases {
		if row.CNF == nil {
			continue
		}
		count++
		g := darBuildGraph(t, row.Ops, row.Outputs)
		c, e := DeriveAigCNF4(g, 2, AigCNFDeriveOptions{})
		if e != nil {
			t.Fatal(row.Seed, e)
		}
		if c.variables != row.CNF.Variables || !slices.Equal(c.objectVariables, row.CNF.ObjectVariables) || len(c.clauses) != len(row.CNF.Clauses) {
			t.Fatal(row.Seed, "size", c, row.CNF)
		}
		for i, clause := range c.clauses {
			if !slices.Equal(clause, row.CNF.Clauses[i]) {
				t.Fatal(row.Seed, i, clause, row.CNF.Clauses[i])
			}
		}
		checkDerivedCNFSemantics(t, g, c)
	}
	if count != 24 {
		t.Fatal(count)
	}
}
func checkDerivedCNFSemantics(t *testing.T, g *AigManager, c *AigMappedCNF) {
	t.Helper()
	solver := NewSatSolver()
	solver.SetVariableCount(c.variables)
	for _, clause := range c.clauses {
		satTestAdd(t, solver, clause)
	}
	for assignment := 0; assignment < 64; assignment++ {
		values := make([]bool, len(g.objects))
		values[0] = true
		var base []SatLit
		for i, id := range g.cis {
			values[id] = assignment>>uint(i)&1 != 0
			l := c.inputs[i]
			if !values[id] {
				l = l.Not()
			}
			base = append(base, l)
		}
		for id, obj := range g.objects {
			if obj.Type == AigAND {
				values[id] = (values[obj.Fanin0.ID()] != obj.Fanin0.Complemented()) && (values[obj.Fanin1.ID()] != obj.Fanin1.Complemented())
			}
		}
		for outputs := 0; outputs < 4; outputs++ {
			assumptions := slices.Clone(base)
			expected := true
			for i, co := range g.cos {
				driver := g.objects[co].Fanin0
				want := values[driver.ID()] != driver.Complemented()
				requested := outputs>>uint(i)&1 != 0
				expected = expected && (want == requested)
				l := c.outputs[i]
				if !requested {
					l = l.Not()
				}
				assumptions = append(assumptions, l)
			}
			answer, e := solver.Solve(assumptions, SatLimits{})
			want := SatUnsatisfiable
			if expected {
				want = SatSatisfiable
			}
			if e != nil || answer.Status != want {
				t.Fatal(assignment, outputs, answer, e)
			}
		}
	}
}

func FuzzDeriveAigCNF(f *testing.F) {
	f.Add([]byte{2, 4, 6, 8, 14, 16})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64 {
			data = data[:64]
		}
		var ops [][2]uint32
		for i := 0; i+1 < len(data); i += 2 {
			ops = append(ops, [2]uint32{uint32(data[i]), uint32(data[i+1])})
		}
		g := darBuildGraph(t, ops, []uint32{uint32(2 * (6 + len(ops))), 3})
		c, e := DeriveAigCNF4(g, 2, AigCNFDeriveOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkDerivedCNFSemantics(t, g, c)
	})
}
