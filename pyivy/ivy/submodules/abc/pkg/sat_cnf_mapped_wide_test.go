package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func wideCNFGraph(t *testing.T, n, output, phase int) (*AigManager, AigID) {
	t.Helper()
	g := NewAigManager()
	var ci []AigLit
	for i := 0; i < n; i++ {
		l, _ := g.CreateCI()
		ci = append(ci, l)
	}
	root, _ := g.And(ci[0], ci[1])
	for i := 2; i < n; i++ {
		root, _ = g.And(root, ci[i])
	}
	g.CreateCO(root.NotIf(phase != 0))
	g.CreateCO(ci[0].NotIf(phase == 0))
	if output == 2 {
		g.SetRegisterCount(1)
	}
	return g, root.ID()
}
func checkWideProjection(t *testing.T, g *AigManager, c *AigMappedCNF, truth *TruthTable, phase int) {
	t.Helper()
	s := NewSatSolver()
	s.SetVariableCount(c.VariableCount())
	for _, clause := range c.Clauses() {
		satTestAdd(t, s, clause)
	}
	count := 1 << uint(truth.variables)
	if count > 256 {
		count = 66
	}
	for sample := 0; sample < count; sample++ {
		a := sample
		if truth.variables > 8 {
			a = (sample*1103515245 + 12345) & ((1 << uint(truth.variables)) - 1)
			if sample == 0 {
				a = 0
			}
			if sample == 1 {
				a = (1 << uint(truth.variables)) - 1
			}
		}
		var base []SatLit
		for i, l := range c.inputs {
			if a>>uint(i)&1 == 0 {
				l = l.Not()
			}
			base = append(base, l)
		}
		root := truth.words[a/64]>>uint(a%64)&1 != 0
		expected := []bool{root != (phase != 0), (a&1 != 0) != (phase == 0)}
		for outputs := 0; outputs < 4; outputs++ {
			assumptions := slices.Clone(base)
			ok := true
			for i, id := range g.cos {
				if c.objectVariables[id] < 0 {
					ok = ok && expected[i]
				} else {
					requested := outputs>>uint(i)&1 != 0
					l := c.outputs[i]
					if !requested {
						l = l.Not()
					}
					assumptions = append(assumptions, l)
					ok = ok && (requested == expected[i])
				}
			}
			r, e := s.Solve(assumptions, SatLimits{})
			want := SatUnsatisfiable
			if ok {
				want = SatSatisfiable
			}
			if e != nil || r.Status != want {
				t.Fatal(a, outputs, r, want, e)
			}
		}
	}
}
func TestMappedCNFWideNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/cnf_mapped_wide.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Inputs, Mode, OutputMode, Phase, Root, Exposed, Variables int
		Truth                                                     []uint64
		ObjectVariables                                           []int
		Clauses                                                   [][]SatLit
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 156 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		g, root := wideCNFGraph(t, row.Inputs, row.OutputMode, row.Phase)
		if int(root) != row.Root {
			t.Fatal(i, root)
		}
		truth, e := TruthTableFromWords(row.Inputs, row.Truth, 0)
		if e != nil {
			t.Fatal(e)
		}
		cuts := []AigCNFCut{{Root: root, Leaves: g.Inputs(), Truth: truth}}
		c, e := WriteAigMappedCNF(g, cuts, row.Exposed, MappedCNFOptions{ISOP: TruthISOPOptions{TryBoth: true}})
		if e != nil || c.VariableCount() != row.Variables || !slices.Equal(c.ObjectVariables(), row.ObjectVariables) || !equalSatClauses(c.Clauses(), row.Clauses) {
			t.Fatal(i, c, e)
		}
		checkWideProjection(t, g, c, truth, row.Phase)
	}
}
func TestMappedCNFMixedAndLimits(t *testing.T) {
	g, root := wideCNFGraph(t, 6, 1, 0)
	small, _ := TruthTableFromWords(2, []uint64{0x8888888888888888}, 0)
	large, _ := TruthTableFromWords(5, []uint64{0x8000000080000000}, 0)
	cuts := []AigCNFCut{{Root: root, Leaves: []AigID{7, 3, 4, 5, 6}, Truth: large}, {Root: 7, Leaves: []AigID{1, 2}, Truth: small}}
	c, e := WriteAigMappedCNF(g, cuts, 2, MappedCNFOptions{})
	if e != nil {
		t.Fatal(e)
	}
	full, _ := TruthTableFromWords(6, []uint64{1 << 63}, 0)
	checkWideProjection(t, g, c, full, 0)
	for _, o := range []MappedCNFOptions{{ISOP: TruthISOPOptions{MaxElements: 1}}, {ISOP: TruthISOPOptions{MaxVisits: 1}}, {ISOP: TruthISOPOptions{Stop: func() bool { return true }}}, {MaxClauses: 1}} {
		c, e := WriteAigMappedCNF(g, cuts, 2, o)
		if e == nil || c != nil {
			t.Fatal(c, e)
		}
	}
	if c, e := WriteAigMappedCNF(g, []AigCNFCut{{Root: root, Leaves: g.Inputs()}}, 2, MappedCNFOptions{}); e == nil || c != nil {
		t.Fatal(c, e)
	}
}
func FuzzMappedCNFWide(f *testing.F) {
	f.Add(uint8(6), uint64(0x6996966996696996), uint8(1))
	f.Fuzz(func(t *testing.T, size uint8, word uint64, mode uint8) {
		n := 5 + int(size%2)
		if n == 5 {
			word = uint64(uint32(word))
			word |= word << 32
		}
		g, root := wideCNFGraph(t, n, int(mode%3), 0)
		truth, e := TruthTableFromWords(n, []uint64{word}, 0)
		if e != nil {
			t.Fatal(e)
		}
		exposed := 0
		if mode%3 == 1 {
			exposed = 2
		} else if mode%3 == 2 {
			exposed = 1
		}
		c, e := WriteAigMappedCNF(g, []AigCNFCut{{Root: root, Leaves: g.Inputs(), Truth: truth}}, exposed, MappedCNFOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkWideProjection(t, g, c, truth, 0)
	})
}
