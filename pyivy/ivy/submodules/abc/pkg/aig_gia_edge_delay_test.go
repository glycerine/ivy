package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaEdgeDelayNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_edge_delay.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		DelayMax, EdgeCount int
		Delays, Packed      []int
		Edges               [][2]GiaID
		Objects             [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 12 {
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
		edges := &GiaEdgeAssignment{rows: slices.Clone(row.Edges)}
		before := slices.Clone(m.objects)
		r, err := edges.EvaluateDelay(m, mapping, GiaLUTAnalysisOptions{})
		if err != nil || r.Maximum != row.DelayMax || r.Edges != row.EdgeCount || !slices.Equal(r.Delays, row.Delays) {
			t.Fatal(i, r, row, err)
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("source changed")
		}
	}
}
func TestGiaEdgeDelayLimits(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	root, _ := m.AppendAnd(a, b)
	m.AppendCO(root)
	mapping, err := NewGiaLUTMapping(m.ObjectCount(), []GiaLUTRecord{{Root: root.ID(), Fanins: []GiaID{a.ID(), b.ID()}}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	e, err := GiaEdgesFromPairs(m.ObjectCount(), [][2]GiaID{{a.ID(), root.ID()}}, GiaEdgeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	r, err := e.EvaluateDelay(m, mapping, GiaLUTAnalysisOptions{MaxVisits: 2})
	if err != nil || r.Maximum != 10 {
		t.Fatal(r, err)
	}
	r.Delays[root.ID()] = 99
	fresh, err := e.EvaluateDelay(m, mapping, GiaLUTAnalysisOptions{})
	if err != nil || fresh.Maximum != 10 {
		t.Fatal("alias", fresh, err)
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxVisits: 1}, {MaxObjects: 1}, {MaxObjects: -1}, {MaxVisits: -1}, {Stop: func() bool { return true }}} {
		if result, err := e.EvaluateDelay(m, mapping, options); result.Delays != nil || err == nil {
			t.Fatal(result, err)
		}
	}
	for _, threshold := range []int{2, 3, 5} {
		calls := 0
		if result, err := e.EvaluateDelay(m, mapping, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= threshold }}); result.Delays != nil || !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, result, err)
		}
	}
	wide, err := NewGiaLUTMapping(m.ObjectCount(), []GiaLUTRecord{{Root: root.ID(), Fanins: []GiaID{1, 2, 1, 2, 1}}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.EvaluateDelay(m, wide, GiaLUTAnalysisOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	var missing *GiaEdgeAssignment
	if _, err := missing.EvaluateDelay(m, mapping, GiaLUTAnalysisOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}
func FuzzGiaEdgeDelay(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{255, 0, 17, 42})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 20 {
			data = data[:20]
		}
		m := buildGiaCutEnumerationTestGraph(t, data)
		records := make([]GiaLUTRecord, 0)
		pairs := make([][2]GiaID, 0)
		for id, obj := range m.objects {
			if !obj.IsAnd() {
				continue
			}
			a, _ := obj.Fanin0(GiaID(id))
			b, _ := obj.Fanin1(GiaID(id))
			records = append(records, GiaLUTRecord{Root: GiaID(id), Fanins: []GiaID{a.ID(), b.ID()}})
			if id%2 == 0 {
				pairs = append(pairs, [2]GiaID{a.ID(), GiaID(id)})
			}
		}
		mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
		if err != nil {
			t.Fatal(err)
		}
		e, err := GiaEdgesFromPairs(m.ObjectCount(), pairs, GiaEdgeOptions{})
		if err != nil {
			t.Fatal(err)
		}
		actual, err := e.EvaluateDelay(m, mapping, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		values := make([]int, m.ObjectCount())
		seen := make([]bool, m.ObjectCount())
		var delay func(GiaID) int
		delay = func(id GiaID) int {
			if seen[id] {
				return values[id]
			}
			seen[id] = true
			record, ok := mapping.Record(id)
			if !ok {
				return 0
			}
			for _, fanin := range record.Fanins {
				step := 10
				marked, err := e.HasEdge(id, fanin)
				if err != nil {
					t.Fatal(err)
				}
				if marked {
					step = 2
				}
				values[id] = max(values[id], delay(fanin)+step)
			}
			return values[id]
		}
		for _, id := range mapping.Roots() {
			delay(id)
		}
		maximum := 0
		for _, id := range m.cos {
			driver, _ := m.objects[id].Fanin0(id)
			maximum = max(maximum, delay(driver.ID()))
		}
		if actual.Maximum != maximum || !slices.Equal(actual.Delays, values) {
			t.Fatal(actual, maximum, values)
		}
	})
}
