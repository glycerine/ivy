package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaEdgeSelectNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_edge_select.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		DelayMax, EdgeCount, UseTwo int
		Delays, Packed              []int
		Edges                       [][2]GiaID
		Objects                     [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 24 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		m := NewGiaManager()
		for _, obj := range row.Objects[1:] {
			switch obj[0] {
			case 1:
				_, err = m.AppendCI()
			case 2:
				_, err = m.AppendCO(GiaLit(obj[1]))
			case 3:
				_, err = m.AppendAnd(GiaLit(obj[1]), GiaLit(obj[2]))
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		mapping, err := GiaLUTMappingFromPacked(m.ObjectCount(), row.Packed, 0)
		if err != nil {
			t.Fatal(err)
		}

		before := slices.Clone(m.objects)
		edges, r, err := mapping.SelectEdges(m, row.UseTwo != 0, GiaLUTAnalysisOptions{})
		if err != nil || r.Maximum != row.DelayMax || r.Edges != row.EdgeCount || !slices.Equal(r.Delays, row.Delays) || !slices.Equal(edges.Rows(), row.Edges) {
			t.Fatal(i, r, row, err)
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("source changed")
		}
	}
}

func TestGiaEdgeSelectBoundsAndChain(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	root, _ := m.AppendAnd(a, b)
	records := []GiaLUTRecord{{Root: root.ID(), Fanins: []GiaID{a.ID(), b.ID()}}}
	for i := 0; i < 10000; i++ {
		next, _ := m.AppendAnd(root, a)
		records = append(records, GiaLUTRecord{Root: next.ID(), Fanins: []GiaID{root.ID(), a.ID()}})
		root = next
	}
	m.AppendCO(root)
	mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, useTwo := range []bool{false, true} {
		edges, r, err := mapping.SelectEdges(m, useTwo, GiaLUTAnalysisOptions{MaxVisits: 20002})
		if err != nil || r.Maximum != 20010 || r.Edges != 10000 {
			t.Fatal(r.Maximum, r.Edges, err)
		}
		evaluated, err := edges.EvaluateDelay(m, mapping, GiaLUTAnalysisOptions{})
		if err != nil || !slices.Equal(evaluated.Delays, r.Delays) {
			t.Fatal("reevaluation", err)
		}
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxVisits: 20001}, {MaxObjects: 1}, {MaxObjects: -1}, {MaxVisits: -1}, {Stop: func() bool { return true }}} {
		if edges, r, err := mapping.SelectEdges(m, false, options); edges != nil || r.Delays != nil || err == nil {
			t.Fatal(edges, r, err)
		}
	}
	calls := 0
	if edges, r, err := mapping.SelectEdges(m, false, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= 7 }}); edges != nil || r.Delays != nil || err == nil {
		t.Fatal(edges, r, err)
	}
}
func FuzzGiaEdgeSelect(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, true)
	f.Add([]byte{255, 0, 17, 42}, false)
	f.Fuzz(func(t *testing.T, data []byte, useTwo bool) {
		if len(data) > 25 {
			data = data[:25]
		}
		m := buildGiaCutEnumerationTestGraph(t, data)
		records := make([]GiaLUTRecord, 0)
		for id, obj := range m.objects {
			if !obj.IsAnd() {
				continue
			}
			a, _ := obj.Fanin0(GiaID(id))
			b, _ := obj.Fanin1(GiaID(id))
			leaves := []GiaID{a.ID(), b.ID()}
			if id%3 == 0 {
				leaves = append(leaves, a.ID())
			}
			records = append(records, GiaLUTRecord{Root: GiaID(id), Fanins: leaves})
		}
		mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
		if err != nil {
			t.Fatal(err)
		}
		before := slices.Clone(m.objects)
		edges, r, err := mapping.SelectEdges(m, useTwo, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		evaluated, err := edges.EvaluateDelay(m, mapping, GiaLUTAnalysisOptions{})
		if err != nil || r.Maximum != evaluated.Maximum || r.Edges != evaluated.Edges || !slices.Equal(r.Delays, evaluated.Delays) {
			t.Fatal("selected timing differs from evaluation", r, evaluated, err)
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("source changed")
		}
	})
}

func TestGiaEdgeSelectTiesAndCapacity(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	d, _ := m.AppendCI()
	left, _ := m.AppendAnd(a, b)
	right, _ := m.AppendAnd(c, d)
	root, _ := m.AppendAnd(left, right)
	m.AppendCO(root)
	records := []GiaLUTRecord{{Root: left.ID(), Fanins: []GiaID{a.ID(), b.ID()}}, {Root: right.ID(), Fanins: []GiaID{c.ID(), d.ID()}}, {Root: root.ID(), Fanins: []GiaID{left.ID(), right.ID()}}}
	mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, useTwo := range []bool{false, true} {
		_, timing, err := mapping.SelectEdges(m, useTwo, GiaLUTAnalysisOptions{})
		want := 20
		if useTwo {
			want = 12
		}
		if err != nil || timing.Maximum != want {
			t.Fatal(useTwo, timing, err)
		}
	}
	records[2].Fanins = nil
	empty, _ := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if _, _, err := empty.SelectEdges(m, true, GiaLUTAnalysisOptions{}); err == nil {
		t.Fatal("empty LUT accepted")
	}
	records[2].Fanins = []GiaID{left.ID(), right.ID(), left.ID(), right.ID(), left.ID()}
	wide, _ := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if _, _, err := wide.SelectEdges(m, true, GiaLUTAnalysisOptions{}); err == nil {
		t.Fatal("wide LUT accepted")
	}
	star := NewGiaManager()
	x, _ := star.AppendCI()
	y, _ := star.AppendCI()
	base, _ := star.AppendAnd(x, y)
	rs := []GiaLUTRecord{{Root: base.ID(), Fanins: []GiaID{x.ID(), y.ID()}}}
	ids := make([]GiaID, 3)
	for i := range ids {
		gate, _ := star.AppendAnd(base, x)
		ids[i] = gate.ID()
		rs = append(rs, GiaLUTRecord{Root: gate.ID(), Fanins: []GiaID{base.ID(), x.ID()}})
		star.AppendCO(gate)
	}
	sm, _ := NewGiaLUTMapping(star.ObjectCount(), rs, 0)
	_, timing, err := sm.SelectEdges(star, true, GiaLUTAnalysisOptions{})
	if err != nil || timing.Delays[ids[0]] != 12 || timing.Delays[ids[1]] != 12 || timing.Delays[ids[2]] != 20 {
		t.Fatal(timing, err)
	}
}
