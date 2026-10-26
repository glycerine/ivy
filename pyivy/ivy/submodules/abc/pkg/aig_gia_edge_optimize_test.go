package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaEdgeOptimizeNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_edge_optimize.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Limit, Maximum, ReverseMax int
		Forward, Reverse, Packed   []int
		Edges                      [][2]GiaID
		Objects                    [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 24 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
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
		result, err := mapping.OptimizeEdges(m, GiaEdgeOptimizationOptions{EdgeLimit: row.Limit})
		if err != nil || result.Timing.Maximum != row.Maximum || result.Timing.ReverseMaximum != row.ReverseMax || !slices.Equal(result.Timing.Forward, row.Forward) || !slices.Equal(result.Timing.Reverse, row.Reverse) || !slices.Equal(result.Edges.Rows(), row.Edges) {
			t.Fatal(index, result, err)
		}
		if !slices.Equal(before, m.objects) || !slices.Equal(mapping.Packed(), row.Packed) {
			t.Fatal("source changed")
		}
	}
}
func TestGiaEdgeOptimizeBounds(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	root, _ := m.AppendAnd(a, b)
	records := []GiaLUTRecord{{Root: root.ID(), Fanins: []GiaID{a.ID(), b.ID()}}}
	for i := 0; i < 4; i++ {
		next, _ := m.AppendAnd(root, a)
		records = append(records, GiaLUTRecord{Root: next.ID(), Fanins: []GiaID{root.ID(), a.ID()}})
		root = next
	}
	m.AppendCO(root)
	mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	result, err := mapping.OptimizeEdges(m, GiaEdgeOptimizationOptions{})
	if err != nil || result.Timing.Maximum != 1 {
		t.Fatal(result, err)
	}
	for _, options := range []GiaEdgeOptimizationOptions{{EdgeLimit: 3}, {EdgeLimit: -1}, {MaxObjects: -1}, {MaxVisits: -1}, {MaxIterations: -1}, {MaxObjects: 1}, {MaxVisits: 1}, {MaxIterations: 1}, {Stop: func() bool { return true }}} {
		if r, err := mapping.OptimizeEdges(m, options); r.Edges != nil || r.Timing.Forward != nil || err == nil {
			t.Fatal(options, r, err)
		}
	}
	for _, threshold := range []int{2, 15, 35, 80} {
		calls := 0
		if r, err := mapping.OptimizeEdges(m, GiaEdgeOptimizationOptions{Stop: func() bool { calls++; return calls >= threshold }}); r.Edges != nil || err == nil {
			t.Fatal(threshold, r, err)
		}
	}
	empty := NewGiaManager()
	em, _ := NewGiaLUTMapping(1, nil, 0)
	if _, err := em.OptimizeEdges(empty, GiaEdgeOptimizationOptions{}); err == nil {
		t.Fatal("native empty assertion precondition")
	}
}
func FuzzGiaEdgeOptimize(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, true)
	f.Add([]byte{255, 0, 17, 42}, false)
	f.Fuzz(func(t *testing.T, data []byte, two bool) {
		if len(data) == 0 {
			return
		}
		if len(data) > 20 {
			data = data[:20]
		}
		m := buildGiaCutEnumerationTestGraph(t, data)
		records := make([]GiaLUTRecord, 0)
		for id, obj := range m.objects {
			if !obj.IsAnd() {
				continue
			}
			a, _ := obj.Fanin0(GiaID(id))
			b, _ := obj.Fanin1(GiaID(id))
			records = append(records, GiaLUTRecord{Root: GiaID(id), Fanins: []GiaID{a.ID(), b.ID()}})
		}
		mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
		if err != nil {
			t.Fatal(err)
		}
		limit := 1
		if two {
			limit = 2
		}
		result, err := mapping.OptimizeEdges(m, GiaEdgeOptimizationOptions{EdgeLimit: limit})
		if err != nil {
			t.Fatal(err)
		}
		n, err := mapping.Network(GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		timing, err := n.EvaluateEdges(result.Edges, GiaLUTAnalysisOptions{})
		if err != nil || !slices.Equal(timing.Forward, result.Timing.Forward) || !slices.Equal(timing.Reverse, result.Timing.Reverse) {
			t.Fatal("result timing", err)
		}
		if _, err := n.CriticalLevels(timing); err != nil {
			t.Fatal(err)
		}
		for id := range m.objects {
			count, err := result.Edges.Count(GiaID(id))
			if err != nil || count > limit {
				t.Fatal("capacity", id, count, err)
			}
		}
		blank, _ := GiaEdgesFromPairs(m.ObjectCount(), nil, GiaEdgeOptions{})
		initial, err := n.EvaluateEdges(blank, GiaLUTAnalysisOptions{})
		if err != nil || result.Timing.Maximum > initial.Maximum {
			t.Fatal("regressed delay", err)
		}
	})
}
