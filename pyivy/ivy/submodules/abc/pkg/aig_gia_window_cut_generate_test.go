package abc

import (
	"encoding/json"
	"math/bits"
	"os"
	"slices"
	"testing"
)

func TestGiaWindowCutGenerateNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_window_cut_generate.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Leaves, Width int
		Fanins        [][2]int
		Ranges        []GiaWindowCutRange
		Cuts          []GiaWindowCut
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 24 {
		t.Fatal(len(rows))
	}
	for ri, row := range rows {
		m := NewGiaManager()
		var all []GiaLit
		var leaves, nodes []GiaID
		for i := 0; i < row.Leaves; i++ {
			l, e := m.AppendCI()
			if e != nil {
				t.Fatal(e)
			}
			all = append(all, l)
			leaves = append(leaves, l.ID())
		}
		for _, pair := range row.Fanins {
			var l GiaLit
			if pair[0] == pair[1] {
				l, e = m.AppendBuffer(all[pair[0]])
			} else {
				l, e = m.AppendAnd(all[pair[0]], all[pair[1]])
			}
			if e != nil {
				t.Fatal(e)
			}
			all = append(all, l)
			nodes = append(nodes, l.ID())
		}
		r, e := m.GenerateWindowCuts(leaves, nodes, row.Width, GiaWindowCutOptions{})
		if e != nil || !slices.Equal(r.Cuts, row.Cuts) || !slices.Equal(r.Ranges, row.Ranges) {
			t.Fatal(ri, e, r.Ranges)
		}
		if len(nodes) <= 12 {
			log := 0
			for 1<<uint(log) < len(nodes) {
				log++
			}
			s, e := NewGiaWindowSolver(log, len(nodes), r.Cuts, r.Ranges, []int{len(nodes) - 1}, GiaWindowCNFOptions{})
			if e != nil {
				t.Fatal(e)
			}
			answer, e := s.Solve(len(nodes), SatLimits{})
			if e != nil || answer.SAT.Status != SatSatisfiable {
				t.Fatal(answer, e)
			}
		}
	}
}
func TestGiaWindowCutGenerateSafety(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	n, _ := m.AppendAnd(a, b)
	leaves := []GiaID{a.ID(), b.ID()}
	nodes := []GiaID{n.ID()}
	for _, o := range []GiaWindowCutOptions{{MaxCuts: 3}, {MaxVisits: 1}, {MaxCuts: -1}, {MaxVisits: -1}, {Stop: func() bool { return true }}} {
		r, e := m.GenerateWindowCuts(leaves, nodes, 4, o)
		if e == nil || r.Cuts != nil || r.Ranges != nil {
			t.Fatal(r, e)
		}
	}
	for _, bad := range [][]GiaID{{a.ID(), a.ID()}, {a.ID()}, {a.ID(), b.ID(), n.ID()}} {
		if _, e := m.GenerateWindowCuts(bad, nodes, 4, GiaWindowCutOptions{}); e == nil {
			t.Fatal(bad)
		}
	}
	r, e := m.GenerateWindowCuts(leaves, nodes, 1, GiaWindowCutOptions{MaxCuts: 4})
	if e != nil || len(r.Cuts) != 4 {
		t.Fatal(r, e)
	}
	for _, width := range []int{0, 7} {
		if _, e := m.GenerateWindowCuts(leaves, nodes, width, GiaWindowCutOptions{}); e == nil {
			t.Fatal(width)
		}
	}
	c, _ := m.AppendCI()
	mux, _ := m.AppendMuxReal(c, a, b)
	if _, e := m.GenerateWindowCuts([]GiaID{a.ID(), b.ID(), c.ID()}, []GiaID{mux.ID()}, 4, GiaWindowCutOptions{}); e == nil {
		t.Fatal("mux")
	}
}
func FuzzGiaWindowCutGenerate(f *testing.F) {
	f.Add([]byte{0, 1, 2, 0, 3, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 32 {
			data = data[:32]
		}
		m := NewGiaManager()
		var lits []GiaLit
		var leaves, nodes []GiaID
		for i := 0; i < 4; i++ {
			l, _ := m.AppendCI()
			lits = append(lits, l)
			leaves = append(leaves, l.ID())
		}
		for i := 0; i+1 < len(data); i += 2 {
			a, b := lits[int(data[i])%len(lits)], lits[int(data[i+1])%len(lits)]
			var l GiaLit
			if a.ID() == b.ID() {
				l, _ = m.AppendBuffer(a)
			} else {
				l, _ = m.AppendAnd(a, b)
			}
			lits = append(lits, l)
			nodes = append(nodes, l.ID())
		}
		r, e := m.GenerateWindowCuts(leaves, nodes, 4, GiaWindowCutOptions{})
		if e != nil {
			t.Fatal(e)
		}
		for i, span := range r.Ranges {
			for _, c := range r.Cuts[span.Start : span.Start+span.Count-1] {
				if c.Node != i || bits.OnesCount64(c.Leaves[0])+bits.OnesCount64(c.Nodes[0]) > 4 || c.Nodes[0]>>uint(i) != 0 {
					t.Fatal(c, i)
				}
				var boundary []GiaID
				for j, id := range leaves {
					if c.Leaves[0]>>uint(j)&1 != 0 {
						boundary = append(boundary, id)
					}
				}
				for j, id := range nodes {
					if c.Nodes[0]>>uint(j)&1 != 0 {
						boundary = append(boundary, id)
					}
				}
				if _, e := m.BoundaryTruth6(lits[len(leaves)+i], boundary, GiaCutTruthOptions{}); e != nil {
					t.Fatal("not a structural cut", c, e)
				}
			}
		}
	})
}
