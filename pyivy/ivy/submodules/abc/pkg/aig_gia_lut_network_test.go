package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func equalGiaIDLevels(a, b [][]GiaID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !slices.Equal(a[i], b[i]) {
			return false
		}
	}
	return true
}
func TestGiaLUTNetworkNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_lut_network.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Maximum, ReverseMax      int
		Forward, Reverse, Packed []int
		Edges                    [][2]GiaID
		Fanouts, Critical        [][]GiaID
		Objects                  [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 12 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		mapping, err := GiaLUTMappingFromPacked(len(row.Objects), row.Packed, 0)
		if err != nil {
			t.Fatal(err)
		}
		network, err := mapping.Network(GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for id := range row.Objects {
			fanouts, err := network.Fanouts(GiaID(id))
			if err != nil || !slices.Equal(fanouts, row.Fanouts[id]) {
				t.Fatal(index, id, fanouts, err)
			}
		}
		e := &GiaEdgeAssignment{rows: slices.Clone(row.Edges)}
		r, err := network.EvaluateEdges(e, GiaLUTAnalysisOptions{})
		if err != nil || r.Maximum != row.Maximum || r.ReverseMaximum != row.ReverseMax || !slices.Equal(r.Forward, row.Forward) || !slices.Equal(r.Reverse, row.Reverse) {
			t.Fatal(index, r, err)
		}
		critical, err := network.CriticalLevels(r)
		if err != nil || !equalGiaIDLevels(critical, row.Critical) {
			t.Fatal(index, critical, row.Critical, err)
		}
	}
}
func TestGiaLUTNetworkLimitsAndOwnership(t *testing.T) {
	mapping, err := NewGiaLUTMapping(5, []GiaLUTRecord{{Root: 3, Fanins: []GiaID{1, 2, 1}}, {Root: 4, Fanins: []GiaID{3, 2}}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxObjects: 4}, {MaxVisits: 4}, {MaxObjects: -1}, {MaxVisits: -1}, {Stop: func() bool { return true }}} {
		if n, err := mapping.Network(options); n != nil || err == nil {
			t.Fatal(n, err)
		}
	}
	n, err := mapping.Network(GiaLUTAnalysisOptions{MaxVisits: 5})
	if err != nil {
		t.Fatal(err)
	}
	fanins, _ := n.Fanins(3)
	fanins[0] = 0
	fanouts, _ := n.Fanouts(1)
	fanouts[0] = 0
	fresh, _ := n.Fanouts(1)
	if !slices.Equal(fresh, []GiaID{3, 3}) {
		t.Fatal("fanout ownership", fresh)
	}
	e, err := GiaEdgesFromPairs(5, nil, GiaEdgeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	// Forward visits five slots; reverse visits one mapped fanout slot.
	r, err := n.EvaluateEdges(e, GiaLUTAnalysisOptions{MaxVisits: 6})
	if err != nil || r.Maximum != 2 || r.ReverseMaximum != 1 {
		t.Fatal(r, err)
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxVisits: 5}, {MaxObjects: 4}, {MaxVisits: -1}, {Stop: func() bool { return true }}} {
		if r, err := n.EvaluateEdges(e, options); r.Forward != nil || err == nil {
			t.Fatal(r, err)
		}
	}
	calls := 0
	if r, err := n.EvaluateEdges(e, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= 7 }}); r.Forward != nil || err == nil {
		t.Fatal(r, err)
	}
	bad := r
	bad.Maximum = 100
	if _, err := n.CriticalLevels(bad); err == nil {
		t.Fatal("invalid max")
	}
	bad = r
	bad.Forward = slices.Clone(r.Forward)
	bad.Forward[3] = 100
	if _, err := n.CriticalLevels(bad); err == nil {
		t.Fatal("invalid timing")
	}
	wide, _ := NewGiaLUTMapping(5, []GiaLUTRecord{{Root: 4, Fanins: []GiaID{1, 2, 1, 2, 1}}}, 0)
	if _, err := wide.Network(GiaLUTAnalysisOptions{}); err == nil {
		t.Fatal("wide LUT")
	}
}
func FuzzGiaLUTNetwork(f *testing.F) {
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
			records = append(records, GiaLUTRecord{Root: GiaID(id), Fanins: []GiaID{a.ID(), b.ID(), a.ID()}})
			if id%2 == 0 {
				pairs = append(pairs, [2]GiaID{a.ID(), GiaID(id)})
			}
		}
		mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
		if err != nil {
			t.Fatal(err)
		}
		n, err := mapping.Network(GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		e, err := GiaEdgesFromPairs(m.ObjectCount(), pairs, GiaEdgeOptions{})
		if err != nil {
			t.Fatal(err)
		}
		actual, err := n.EvaluateEdges(e, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for _, reverse := range []bool{false, true} {
			memo := make([]int, m.ObjectCount())
			seen := make([]bool, m.ObjectCount())
			var value func(GiaID) int
			value = func(id GiaID) int {
				if seen[id] {
					return memo[id]
				}
				seen[id] = true
				record, ok := mapping.Record(id)
				if !ok || len(record.Fanins) == 0 {
					return 0
				}
				neighbors := record.Fanins
				if reverse {
					neighbors = nil
					for _, other := range records {
						for _, fanin := range other.Fanins {
							if fanin == id {
								neighbors = append(neighbors, other.Root)
							}
						}
					}
				}
				for _, neighbor := range neighbors {
					step := 1
					yes, _ := e.HasEdge(id, neighbor)
					if yes {
						step = 0
					}
					memo[id] = max(memo[id], value(neighbor)+step)
				}
				return memo[id]
			}
			maximum := 0
			for _, id := range n.Roots() {
				maximum = max(maximum, value(id))
			}
			expected := actual.Forward
			wantMax := actual.Maximum
			if reverse {
				expected = actual.Reverse
				wantMax = actual.ReverseMaximum
			}
			if maximum != wantMax || !slices.Equal(memo, expected) {
				t.Fatal("recursive model", reverse, memo, actual)
			}
		}
	})
}
