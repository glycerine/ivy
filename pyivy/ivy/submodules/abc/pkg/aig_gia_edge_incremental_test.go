package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaEdgeIncrementalNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_edge_incremental.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		UseTwo, DelayMax, EdgeCount int
		PriorRows, Edges            [][2]GiaID
		PriorDelays, Delays, Packed []int
		Roots, Order                []GiaID
		Objects                     [][3]int
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
		prior := &GiaEdgeAssignment{rows: slices.Clone(row.PriorRows)}
		timing := GiaEdgeTiming{Delays: slices.Clone(row.PriorDelays)}
		before := slices.Clone(m.objects)
		result, err := mapping.RecomputeEdges(m, row.Roots, prior, timing, row.UseTwo != 0, GiaLUTAnalysisOptions{})
		if err != nil || result.AffectedMaximum != row.DelayMax || result.Timing.Edges != row.EdgeCount || !slices.Equal(result.Timing.Delays, row.Delays) || !slices.Equal(result.Edges.Rows(), row.Edges) || !slices.Equal(result.Order, row.Order) {
			t.Fatal(index, result, err)
		}
		if !slices.Equal(prior.Rows(), row.PriorRows) || !slices.Equal(timing.Delays, row.PriorDelays) || !slices.Equal(before, m.objects) {
			t.Fatal("source changed")
		}
	}
}
func TestGiaEdgeIncrementalBounds(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(x, a)
	m.AppendCO(y)
	mapping, _ := NewGiaLUTMapping(m.ObjectCount(), []GiaLUTRecord{{Root: x.ID(), Fanins: []GiaID{a.ID(), b.ID()}}, {Root: y.ID(), Fanins: []GiaID{x.ID(), a.ID()}}}, 0)
	prior, timing, err := mapping.SelectEdges(m, true, GiaLUTAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := mapping.RecomputeEdges(m, nil, prior, timing, true, GiaLUTAnalysisOptions{})
	if err != nil || result.AffectedMaximum != 0 || !slices.Equal(result.Timing.Delays, timing.Delays) {
		t.Fatal(result, err)
	}
	result.Timing.Delays[0] = 99
	if timing.Delays[0] != 0 {
		t.Fatal("timing alias")
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxObjects: 1}, {MaxVisits: 1}, {MaxObjects: -1}, {MaxVisits: -1}, {Stop: func() bool { return true }}} {
		if r, err := mapping.RecomputeEdges(m, []GiaID{x.ID()}, prior, timing, true, options); r.Edges != nil || r.Timing.Delays != nil || err == nil {
			t.Fatal(r, err)
		}
	}
	if _, err := mapping.RecomputeEdges(m, []GiaID{GiaID(m.ObjectCount())}, prior, timing, true, GiaLUTAnalysisOptions{}); err == nil {
		t.Fatal("invalid root")
	}
	bad := timing
	bad.Delays = slices.Clone(timing.Delays)
	bad.Delays[0] = -1
	if _, err := mapping.RecomputeEdges(m, nil, prior, bad, true, GiaLUTAnalysisOptions{}); err == nil {
		t.Fatal("negative delay")
	}
	calls := 0
	if r, err := mapping.RecomputeEdges(m, []GiaID{x.ID()}, prior, timing, true, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= 4 }}); r.Edges != nil || err == nil {
		t.Fatal(r, err)
	}
}
func FuzzGiaEdgeIncremental(f *testing.F) {
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
		prior, timing, err := mapping.SelectEdges(m, two, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		beforeRows, beforeDelays := prior.Rows(), slices.Clone(timing.Delays)
		root := records[int(data[0])%len(records)].Root
		result, err := mapping.RecomputeEdges(m, []GiaID{root}, prior, timing, two, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		touched := make([]bool, m.ObjectCount())
		maximum := 0
		for _, id := range result.Order {
			touched[id] = true
			maximum = max(maximum, result.Timing.Delays[id])
		}
		if maximum != result.AffectedMaximum {
			t.Fatal("affected maximum")
		}
		for id, delay := range result.Timing.Delays {
			if !touched[id] && delay != timing.Delays[id] {
				t.Fatal("untouched delay changed", id)
			}
		}
		global := 0
		for _, id := range m.cos {
			lit, _ := m.objects[id].Fanin0(id)
			global = max(global, result.Timing.Delays[lit.ID()])
		}
		if global != result.Timing.Maximum {
			t.Fatal("CO maximum")
		}
		if !slices.Equal(prior.Rows(), beforeRows) || !slices.Equal(timing.Delays, beforeDelays) {
			t.Fatal("source changed")
		}
	})
}
