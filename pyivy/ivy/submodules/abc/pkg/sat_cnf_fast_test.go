package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func checkFastCNFProjection(t *testing.T, g *AigManager, c *AigMappedCNF) {
	t.Helper()
	s := NewSatSolver()
	s.SetVariableCount(c.variables)
	for _, clause := range c.clauses {
		satTestAdd(t, s, clause)
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
		for outputs := 0; outputs < 1<<uint(len(g.cos)); outputs++ {
			assumptions := slices.Clone(base)
			expected := true
			for i, id := range g.cos {
				driver := g.objects[id].Fanin0
				value := values[driver.ID()] != driver.Complemented()
				if c.objectVariables[id] < 0 {
					expected = expected && value
				} else {
					requested := outputs>>uint(i)&1 != 0
					expected = expected && (value == requested)
					l := c.outputs[i]
					if !requested {
						l = l.Not()
					}
					assumptions = append(assumptions, l)
				}
			}
			answer, e := s.Solve(assumptions, SatLimits{})
			want := SatUnsatisfiable
			if expected {
				want = SatSatisfiable
			}
			if e != nil || answer.Status != want {
				t.Fatal(assignment, outputs, answer, want, e)
			}
		}
	}
}
func TestFastCNFNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/cnf_fast_mark.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Ops                          [][4]uint32
		Outputs                      []uint32
		CountVariables, CountClauses int
		Fast                         []struct {
			Mode, Variables int
			ObjectVariables []int
			Clauses         [][]SatLit
		}
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	count := 0
	for index, row := range rows {
		countGraph := fastMarkGraph(t, row.Ops, row.Outputs)
		beforeCount := slices.Clone(countGraph.objects)
		size, e := CountAigFastCNFNodes(countGraph, FastCNFOptions{})
		if e != nil || size.MarkedObjects != row.CountVariables || size.Clauses != row.CountClauses {
			t.Fatal(index, size, row.CountVariables, row.CountClauses, e)
		}
		if !slices.Equal(beforeCount, countGraph.objects) {
			t.Fatal("count changed source")
		}
		for _, native := range row.Fast {
			count++
			g := fastMarkGraph(t, row.Ops, row.Outputs)
			exposed := 0
			if native.Mode == 1 {
				exposed = 2
			}
			if native.Mode == 2 {
				g.SetRegisterCount(1)
				exposed = 1
			}
			before := slices.Clone(g.objects)
			c, e := DeriveAigFastCNF(g, exposed, FastCNFOptions{CNF: MappedCNFOptions{ISOP: TruthISOPOptions{TryBoth: true}}})
			if e != nil || c.variables != native.Variables || !slices.Equal(c.objectVariables, native.ObjectVariables) || !equalSatClauses(c.clauses, native.Clauses) {
				t.Fatal(index, native.Mode, c, native, e)
			}
			if !slices.Equal(before, g.objects) {
				t.Fatal("source changed")
			}
			checkFastCNFProjection(t, g, c)
		}
	}
	if count != 96 {
		t.Fatal(count)
	}
}
func TestFastCNFLimitsAndWideAND(t *testing.T) {
	ops := [][4]uint32{{0, 2, 4, 0}, {1, 6, 8, 0}, {2, 10, 14, 16}}
	g := fastMarkGraph(t, ops, []uint32{18, 2})
	for _, o := range []FastCNFOptions{{Cone: FastCNFConeOptions{MaxVisits: 1}}, {CNF: MappedCNFOptions{MaxObjects: 1}}, {CNF: MappedCNFOptions{MaxVariables: 1}}, {CNF: MappedCNFOptions{MaxClauses: 1}}, {CNF: MappedCNFOptions{MaxLiterals: 1}}, {CNF: MappedCNFOptions{ISOP: TruthISOPOptions{MaxElements: 1}}}, {CNF: MappedCNFOptions{Stop: func() bool { return true }}}} {
		c, e := DeriveAigFastCNF(g, 2, o)
		if c != nil || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(c, e)
		}
	}
	for _, o := range []FastCNFOptions{{Cone: FastCNFConeOptions{MaxVisits: 1}}, {CNF: MappedCNFOptions{MaxObjects: 1}}, {CNF: MappedCNFOptions{MaxVariables: 1}}, {CNF: MappedCNFOptions{MaxClauses: 1}}, {CNF: MappedCNFOptions{ISOP: TruthISOPOptions{MaxElements: 1}}}, {CNF: MappedCNFOptions{Stop: func() bool { return true }}}} {
		size, e := CountAigFastCNFNodes(g, o)
		if size != (FastCNFNodeSize{}) || !(errors.Is(e, ErrLimit) || errors.Is(e, ErrCancelled)) {
			t.Fatal(size, e)
		}
	}
	wide, _ := wideCNFGraph(t, 1000, 1, 0)
	c, e := DeriveAigFastCNF(wide, 2, FastCNFOptions{})
	if e != nil {
		t.Fatal(e)
	}
	s := NewSatSolver()
	if ok, e := c.LoadInto(s, 0); e != nil || !ok {
		t.Fatal(ok, e)
	}
	assumptions := slices.Clone(c.inputs)
	assumptions = append(assumptions, c.outputs[0])
	answer, e := s.Solve(assumptions, SatLimits{})
	if e != nil || answer.Status != SatSatisfiable {
		t.Fatal(answer, e)
	}
	assumptions[0] ^= 1
	answer, e = s.Solve(assumptions, SatLimits{})
	if e != nil || answer.Status != SatUnsatisfiable {
		t.Fatal(answer, e)
	}
	empty, e := DeriveAigFastCNF(NewAigManager(), 0, FastCNFOptions{})
	if e != nil || empty.variables != 2 || !equalSatClauses(empty.clauses, [][]SatLit{{2}}) {
		t.Fatal(empty, e)
	}
	size, e := CountAigFastCNFNodes(NewAigManager(), FastCNFOptions{})
	if e != nil || size != (FastCNFNodeSize{}) {
		t.Fatal(size, e)
	}
}
func TestFastCNFCountContradictorySupergate(t *testing.T) {
	// Native counting regularizes the two boundary polarities and sees false,
	// while clause emission retains both signed leaves of the AND supergate.
	ops := [][4]uint32{{48, 48, 48, 48}, {48, 50, 55, 48}, {48, 88, 50, 48}}
	g := fastMarkGraph(t, ops, []uint32{18, 3})
	size, e := CountAigFastCNFNodes(g, FastCNFOptions{})
	if e != nil || size != (FastCNFNodeSize{MarkedObjects: 7, Clauses: 1}) {
		t.Fatal(size, e)
	}
	c, e := DeriveAigFastCNF(g, 2, FastCNFOptions{})
	if e != nil {
		t.Fatal(e)
	}
	emittedNodeClauses := len(c.clauses) - len(g.cos) - 2 - 1
	if emittedNodeClauses != 4 || size.Clauses >= emittedNodeClauses {
		t.Fatal(size, emittedNodeClauses)
	}
	checkFastCNFProjection(t, g, c)
}
func FuzzFastCNF(f *testing.F) {
	f.Add([]byte{0, 2, 4, 0, 1, 6, 8, 0, 2, 10, 14, 16}, uint8(1))
	f.Fuzz(func(t *testing.T, data []byte, mode uint8) {
		if len(data) > 80 {
			data = data[:80]
		}
		var ops [][4]uint32
		for i := 0; i+3 < len(data); i += 4 {
			ops = append(ops, [4]uint32{uint32(data[i]), uint32(data[i+1]), uint32(data[i+2]), uint32(data[i+3])})
		}
		g := fastMarkGraph(t, ops, []uint32{uint32(2 * (6 + len(ops))), 3})
		exposed := 0
		if mode%3 == 1 {
			exposed = 2
		}
		if mode%3 == 2 {
			g.SetRegisterCount(1)
			exposed = 1
		}
		c, e := DeriveAigFastCNF(g, exposed, FastCNFOptions{})
		if e != nil {
			t.Fatal(e)
		}
		size, e := CountAigFastCNFNodes(g, FastCNFOptions{})
		emittedNodeClauses := len(c.clauses) - len(g.cos) - exposed - 1
		if e != nil || size.Clauses > emittedNodeClauses {
			t.Fatal(size, len(c.clauses), e)
		}
		checkFastCNFProjection(t, g, c)
	})
}
