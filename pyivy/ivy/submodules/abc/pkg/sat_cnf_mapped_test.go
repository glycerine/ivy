package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func mappedCNFFixture(t *testing.T, output, phase int) *AigManager {
	t.Helper()
	g := NewAigManager()
	var ci []AigLit
	for i := 0; i < 4; i++ {
		l, _ := g.CreateCI()
		ci = append(ci, l)
	}
	a, _ := g.And(ci[0], ci[1])
	b, _ := g.And(ci[2], ci[3])
	root, _ := g.And(a, b)
	g.CreateCO(root.NotIf(phase != 0))
	g.CreateCO(ci[0].NotIf(phase == 0))
	if output == 2 {
		g.SetRegisterCount(1)
	}
	return g
}
func checkMappedCNFProjection(t *testing.T, g *AigManager, cuts []AigCNFCut4, c *AigMappedCNF) {
	t.Helper()
	solver := NewSatSolver()
	solver.SetVariableCount(c.VariableCount())
	for _, clause := range c.Clauses() {
		satTestAdd(t, solver, clause)
	}
	byRoot := make([]*AigCNFCut4, len(g.objects))
	for i := range cuts {
		byRoot[cuts[i].Root] = &cuts[i]
	}
	for mask := 0; mask < 16; mask++ {
		values := make([]bool, len(g.objects))
		values[0] = true
		var base []SatLit
		for i, id := range g.cis {
			values[id] = mask>>uint(i)&1 != 0
			l := c.inputs[i]
			if !values[id] {
				l = l.Not()
			}
			base = append(base, l)
		}
		for id := 1; id < len(values); id++ {
			if cut := byRoot[id]; cut != nil {
				index := 0
				for i, leaf := range cut.Leaves {
					if values[leaf] {
						index |= 1 << uint(i)
					}
				}
				values[id] = cut.Truth>>uint(index)&1 != 0
			}
		}
		for outputs := 0; outputs < 4; outputs++ {
			a := slices.Clone(base)
			expected := true
			for i, id := range g.cos {
				driver := g.objects[id].Fanin0
				value := values[driver.ID()] != driver.Complemented()
				if c.objectVariables[id] < 0 {
					expected = expected && value
				} else {
					requested := outputs>>uint(i)&1 != 0
					l := c.outputs[i]
					if !requested {
						l = l.Not()
					}
					a = append(a, l)
					expected = expected && (requested == value)
				}
			}
			answer, e := solver.Solve(a, SatLimits{})
			want := SatUnsatisfiable
			if expected {
				want = SatSatisfiable
			}
			if e != nil || answer.Status != want {
				t.Fatal(mask, outputs, cuts, answer, want, e)
			}
		}
	}
}
func TestMappedCNF4Native(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/cnf_mapped4.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Pattern, OutputMode, Phase, Exposed, Variables int
		ObjectVariables                                []int
		Cuts                                           []AigCNFCut4
		Clauses                                        [][]SatLit
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 72 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		g := mappedCNFFixture(t, row.OutputMode, row.Phase)
		before := slices.Clone(g.objects)
		c, e := WriteAigMappedCNF4(g, row.Cuts, row.Exposed, MappedCNFOptions{})
		if e != nil || c.VariableCount() != row.Variables || !slices.Equal(c.ObjectVariables(), row.ObjectVariables) || !equalSatClauses(c.Clauses(), row.Clauses) {
			t.Fatal(i, c, e)
		}
		checkMappedCNFProjection(t, g, row.Cuts, c)
		if !slices.Equal(before, g.objects) {
			t.Fatal("graph changed")
		}
	}
}
func TestMappedCNF4Validation(t *testing.T) {
	g := mappedCNFFixture(t, 0, 0)
	cuts := []AigCNFCut4{{Root: 7, Leaves: []AigID{1, 2, 3, 4}, Truth: 0x8000}}
	for _, o := range []MappedCNFOptions{{MaxObjects: 1}, {MaxVariables: 1}, {MaxClauses: 1}, {MaxLiterals: 1}, {MaxLiterals: -1}, {Stop: func() bool { return true }}} {
		if c, e := WriteAigMappedCNF4(g, cuts, 0, o); e == nil || c != nil {
			t.Fatal(c, e)
		}
	}
	for _, bad := range [][]AigCNFCut4{{{Root: 7, Leaves: []AigID{1}, Truth: 1}}, {{Root: 7, Leaves: []AigID{5}, Truth: 0xaaaa}}, {{Root: 7, Leaves: []AigID{7}, Truth: 0xaaaa}}, {{Root: 7, Truth: 0}, {Root: 7, Truth: 0}}, {{Root: 1, Truth: 0}}} {
		if c, e := WriteAigMappedCNF4(g, bad, 0, MappedCNFOptions{}); e == nil || c != nil {
			t.Fatal(bad, c, e)
		}
	}
	if _, e := WriteAigMappedCNF4(g, cuts, 1, MappedCNFOptions{}); e == nil {
		t.Fatal("invalid output mode")
	}
	c, e := WriteAigMappedCNF4(g, cuts, 2, MappedCNFOptions{})
	if e != nil {
		t.Fatal(e)
	}
	v := c.ObjectVariables()
	v[0] = -1
	clauses := c.Clauses()
	clauses[0][0] ^= 1
	if c.ObjectVariables()[0] < 0 || equalSatClauses(c.Clauses(), clauses) {
		t.Fatal("aliased output")
	}
	empty, e := WriteAigMappedCNF4(NewAigManager(), nil, 0, MappedCNFOptions{})
	if e != nil || empty.VariableCount() != 2 || !equalSatClauses(empty.Clauses(), [][]SatLit{{2}}) {
		t.Fatal(empty, e)
	}
}
func FuzzMappedCNF4(f *testing.F) {
	f.Add(uint16(0x6996), uint8(1), true)
	f.Fuzz(func(t *testing.T, truth uint16, mode uint8, phase bool) {
		output := int(mode % 3)
		p := 0
		if phase {
			p = 1
		}
		g := mappedCNFFixture(t, output, p)
		exposed := 0
		if output == 1 {
			exposed = 2
		} else if output == 2 {
			exposed = 1
		}
		cuts := []AigCNFCut4{{Root: 7, Leaves: []AigID{1, 2, 3, 4}, Truth: truth}}
		c, e := WriteAigMappedCNF4(g, cuts, exposed, MappedCNFOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkMappedCNFProjection(t, g, cuts, c)
	})
}

func TestMappedCNF4AIGRelation(t *testing.T) {
	for _, mode := range []int{0, 1, 2} {
		g := mappedCNFFixture(t, mode, 1)
		var exposed int
		if mode == 1 {
			exposed = 2
		} else if mode == 2 {
			exposed = 1
		}
		cuts := []AigCNFCut4{{Root: 7, Leaves: []AigID{1, 2, 3, 4}, Truth: 0x8000}}
		c, e := WriteAigMappedCNF4(g, cuts, exposed, MappedCNFOptions{})
		if e != nil {
			t.Fatal(e)
		}
		solver := NewSatSolver()
		solver.SetVariableCount(c.VariableCount())
		for _, clause := range c.Clauses() {
			satTestAdd(t, solver, clause)
		}
		for mask := 0; mask < 16; mask++ {
			inputs := make([]bool, 4)
			var base []SatLit
			for i := range inputs {
				inputs[i] = mask>>uint(i)&1 != 0
				l := c.inputs[i]
				if !inputs[i] {
					l = l.Not()
				}
				base = append(base, l)
			}
			outputs, e := g.Evaluate(inputs)
			if e != nil {
				t.Fatal(e)
			}
			for assignment := 0; assignment < 4; assignment++ {
				a := slices.Clone(base)
				want := true
				for i, id := range g.cos {
					if c.objectVariables[id] < 0 {
						want = want && outputs[i]
					} else {
						requested := assignment>>uint(i)&1 != 0
						l := c.outputs[i]
						if !requested {
							l = l.Not()
						}
						a = append(a, l)
						want = want && (requested == outputs[i])
					}
				}
				r, e := solver.Solve(a, SatLimits{})
				status := SatUnsatisfiable
				if want {
					status = SatSatisfiable
				}
				if e != nil || r.Status != status {
					t.Fatal(mode, mask, assignment, r, outputs, e)
				}
			}
		}
	}
}
