package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaLUTWindowNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_lut_window.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Two, Maximum   int
		Nodes          []GiaID
		Window         [][]GiaID
		Delays, Packed []int
		Edges          [][2]GiaID
		Objects        [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 12 {
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
		network, err := mapping.Network(GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		before := slices.Clone(m.objects)
		e, r, err := network.EvaluateWindow(m, row.Nodes, row.Window, row.Two != 0, GiaLUTAnalysisOptions{})
		if err != nil || r.Maximum != row.Maximum || !slices.Equal(r.Delays, row.Delays) || !slices.Equal(e.Rows(), row.Edges) {
			t.Fatal(index, r, err)
		}
		if !slices.Equal(before, m.objects) || !slices.Equal(mapping.Packed(), row.Packed) {
			t.Fatal("source changed")
		}
		for _, id := range mapping.Roots() {
			record, _ := mapping.Record(id)
			actual, _ := network.Fanins(id)
			if !slices.Equal(actual, record.Fanins) {
				t.Fatal("original network changed")
			}
		}
	}
}
func TestGiaLUTWindowValidationAndOwnership(t *testing.T) {
	mapping, _ := NewGiaLUTMapping(5, []GiaLUTRecord{{Root: 3, Fanins: []GiaID{1, 2}}, {Root: 4, Fanins: []GiaID{3, 1}}}, 0)
	n, err := mapping.Network(GiaLUTAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	window := [][]GiaID{{2}}
	proposal, err := n.WithWindow([]GiaID{4}, window, GiaLUTAnalysisOptions{MaxVisits: 3})
	if err != nil {
		t.Fatal(err)
	}
	window[0][0] = 1
	actual, _ := proposal.Fanins(4)
	if !slices.Equal(actual, []GiaID{2}) {
		t.Fatal("window alias")
	}
	fanouts, _ := proposal.Fanouts(3)
	if len(fanouts) != 0 {
		t.Fatal("stale fanout")
	}
	fanouts, _ = proposal.Fanouts(2)
	if !slices.Equal(fanouts, []GiaID{3, 4}) {
		t.Fatal(fanouts)
	}
	for _, nodes := range [][]GiaID{{0}, {5}, {3, 3}} {
		win := make([][]GiaID, len(nodes))
		if _, err := n.WithWindow(nodes, win, GiaLUTAnalysisOptions{}); err == nil {
			t.Fatal(nodes)
		}
	}
	for _, win := range [][][]GiaID{nil, {{4}}, {{1, 2, 1, 2, 1, 2, 1}}} {
		if _, err := n.WithWindow([]GiaID{4}, win, GiaLUTAnalysisOptions{}); err == nil {
			t.Fatal(win)
		}
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxVisits: 2}, {MaxObjects: 4}, {MaxVisits: -1}, {MaxObjects: -1}, {Stop: func() bool { return true }}} {
		if p, err := n.WithWindow([]GiaID{4}, [][]GiaID{{2}}, options); p != nil || err == nil {
			t.Fatal(p, err)
		}
	}
	calls := 0
	if p, err := n.WithWindow([]GiaID{4}, [][]GiaID{{2}}, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= 4 }}); p != nil || err == nil {
		t.Fatal(p, err)
	}
	if _, _, err := n.EvaluateWindow(nil, nil, nil, false, GiaLUTAnalysisOptions{}); err == nil {
		t.Fatal("nil graph")
	}
}
func FuzzGiaLUTWindow(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, uint8(3))
	f.Add([]byte{255, 0, 17, 42}, uint8(0))
	f.Fuzz(func(t *testing.T, data []byte, flags uint8) {
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
		n, err := mapping.Network(GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		root := records[len(records)-1].Root
		replacement := make([]GiaID, int(flags%5))
		for i := range replacement {
			replacement[i] = m.cis[(i+int(flags))%len(m.cis)]
		}
		proposal, err := n.WithWindow([]GiaID{root}, [][]GiaID{replacement}, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		expectedRecords := append([]GiaLUTRecord(nil), records[:len(records)-1]...)
		if len(replacement) > 0 {
			expectedRecords = append(expectedRecords, GiaLUTRecord{Root: root, Fanins: replacement})
		}
		expected, err := NewGiaLUTMapping(m.ObjectCount(), expectedRecords, 0)
		if err != nil {
			t.Fatal(err)
		}
		for id := range m.objects {
			want := make([]GiaID, 0)
			for _, r := range expectedRecords {
				for _, fanin := range r.Fanins {
					if int(fanin) == id {
						want = append(want, r.Root)
					}
				}
			}
			actual, err := proposal.Fanouts(GiaID(id))
			if err != nil || !slices.Equal(actual, want) {
				t.Fatal("fanout model", id, actual, want, err)
			}
		}
		edges, timing, err := n.EvaluateWindow(m, []GiaID{root}, [][]GiaID{replacement}, flags&128 != 0, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		wantEdges, wantTiming, err := expected.SelectEdges(m, flags&128 != 0, GiaLUTAnalysisOptions{})
		if err != nil || !slices.Equal(edges.Rows(), wantEdges.Rows()) || !slices.Equal(timing.Delays, wantTiming.Delays) {
			t.Fatal("temporary evaluation model", err)
		}
	})
}
