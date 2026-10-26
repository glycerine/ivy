package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaEdgeReverseNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_edge_reverse.json")
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

		before := slices.Clone(m.objects)
		edges, r, err := mapping.SelectEdgesReverse(m, GiaLUTAnalysisOptions{})
		if err != nil || r.Maximum != row.DelayMax || r.Edges != row.EdgeCount || !slices.Equal(r.Distances, row.Delays) || !slices.Equal(edges.Rows(), row.Edges) {
			t.Fatal(i, r, row, err)
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("source changed")
		}
	}
}

func checkGiaReverseEdgeModel(t *testing.T, m *GiaManager, mapping *GiaLUTMapping, e *GiaEdgeAssignment, r GiaReverseEdgeTiming) {
	t.Helper()
	distances := make([]int, m.ObjectCount())
	roots := mapping.Roots()
	for i := len(roots) - 1; i >= 0; i-- {
		root := roots[i]
		record, _ := mapping.Record(root)
		for _, fanin := range record.Fanins {
			weight := 1
			yes, err := e.HasEdge(fanin, root)
			if err != nil {
				t.Fatal(err)
			}
			if yes {
				weight = 0
			}
			distances[fanin] = max(distances[fanin], distances[root]+weight)
		}
	}
	maximum := 0
	for _, id := range m.cis {
		maximum = max(maximum, distances[id])
	}
	if maximum != r.Maximum || !slices.Equal(distances, r.Distances) {
		t.Fatal("weighted reverse model", r, distances, maximum)
	}
}
func TestGiaEdgeReverseDepthAndLimits(t *testing.T) {
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
	// No CO: reverse timing still includes dangling mapped logic.
	mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	e, r, err := mapping.SelectEdgesReverse(m, GiaLUTAnalysisOptions{MaxVisits: 20002})
	if err != nil || r.Maximum != 1 {
		t.Fatal(r.Maximum, err)
	}
	checkGiaReverseEdgeModel(t, m, mapping, e, r)
	for _, options := range []GiaLUTAnalysisOptions{{MaxVisits: 20001}, {MaxObjects: 1}, {MaxVisits: -1}, {MaxObjects: -1}, {Stop: func() bool { return true }}} {
		if e, r, err := mapping.SelectEdgesReverse(m, options); e != nil || r.Distances != nil || err == nil {
			t.Fatal(e, r, err)
		}
	}
	calls := 0
	if e, r, err := mapping.SelectEdgesReverse(m, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= 9 }}); e != nil || r.Distances != nil || err == nil {
		t.Fatal(e, r, err)
	}
}
func FuzzGiaEdgeReverse(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{255, 0, 17, 42})
	f.Fuzz(func(t *testing.T, data []byte) {
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
		e, r, err := mapping.SelectEdgesReverse(m, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		checkGiaReverseEdgeModel(t, m, mapping, e, r)
		if !slices.Equal(before, m.objects) {
			t.Fatal("source changed")
		}
	})
}
