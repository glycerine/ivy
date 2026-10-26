package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaEdgesNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_edges.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Kind, Skipped int
		Input         []int
		Rows, Pairs   [][2]GiaID
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 32 {
		t.Fatal(len(rows))
	}
	records := make([]GiaLUTRecord, 7)
	for i := range records {
		records[i].Root = GiaID(i + 1)
	}
	mapping, err := NewGiaLUTMapping(8, records, 0)
	if err != nil {
		t.Fatal(err)
	}
	for index, row := range rows {
		var e *GiaEdgeAssignment
		if row.Kind == 0 {
			pairs := make([][2]GiaID, len(row.Input)/2)
			for i := range pairs {
				pairs[i] = [2]GiaID{GiaID(row.Input[2*i]), GiaID(row.Input[2*i+1])}
			}
			e, err = GiaEdgesFromPairs(8, pairs, GiaEdgeOptions{})
		} else {
			var p *GiaLUTPacking
			p, err = GiaLUTPackingFromPacked(mapping, row.Input, 0)
			if err != nil {
				t.Fatal(err)
			}
			e, err = p.Edges(GiaEdgeOptions{})
		}
		if err != nil || !slices.Equal(e.Rows(), row.Rows) || !slices.Equal(e.Pairs(), row.Pairs) || e.Skipped() != row.Skipped {
			t.Fatal(index, e, err)
		}
		for id, r := range row.Rows {
			want := 0
			for _, v := range r {
				if v > 0 {
					want++
				}
			}
			count, err := e.Count(GiaID(id))
			if err != nil || count != want {
				t.Fatal(count, err)
			}
			for next := 0; next < 8; next++ {
				yes, err := e.HasEdge(GiaID(id), GiaID(next))
				if err != nil || yes != (r[0] == GiaID(next) || r[1] == GiaID(next)) {
					t.Fatal("lookup")
				}
			}
		}
	}
}
func TestGiaEdgesLimits(t *testing.T) {
	pairs := [][2]GiaID{{1, 2}, {1, 3}}
	for _, options := range []GiaEdgeOptions{{MaxObjects: -1}, {MaxPairs: -1}, {MaxObjects: 3}, {MaxPairs: 1}, {Stop: func() bool { return true }}} {
		if e, err := GiaEdgesFromPairs(4, pairs, options); e != nil || err == nil {
			t.Fatal(e, err)
		}
	}
	for _, pairs := range [][][2]GiaID{{{2, 1}}, {{1, 1}}, {{1, 4}}} {
		if _, err := GiaEdgesFromPairs(4, pairs, GiaEdgeOptions{}); err == nil {
			t.Fatal(pairs)
		}
	}
	e, err := GiaEdgesFromPairs(4, pairs, GiaEdgeOptions{})
	if err != nil {
		t.Fatal(err)
	}
	rows := e.Rows()
	rows[1][0] = 0
	if e.Rows()[1][0] != 2 {
		t.Fatal("alias")
	}
	if _, err := e.Count(4); err == nil {
		t.Fatal("bad index")
	}
	if _, err := e.HasEdge(1, 4); err == nil {
		t.Fatal("bad neighbor")
	}
	var missing *GiaEdgeAssignment
	if missing.Rows() != nil || missing.Pairs() != nil || missing.Skipped() != 0 {
		t.Fatal("nil accessors")
	}
}
func FuzzGiaEdges(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6})
	f.Add([]byte{0, 7, 0, 7, 0, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 100 {
			data = data[:100]
		}
		pairs := make([][2]GiaID, 0)
		neighbors := make([][]GiaID, 8)
		skipped := 0
		for i := 0; i+1 < len(data); i += 2 {
			a, b := GiaID(data[i]%8), GiaID(data[i+1]%8)
			if a == b {
				continue
			}
			if a > b {
				a, b = b, a
			}
			pairs = append(pairs, [2]GiaID{a, b})
			for _, p := range [][2]GiaID{{a, b}, {b, a}} {
				if len(neighbors[p[0]]) == 2 {
					skipped++
				} else if p[1] != 0 {
					neighbors[p[0]] = append(neighbors[p[0]], p[1])
				}
			}
		}
		e, err := GiaEdgesFromPairs(8, pairs, GiaEdgeOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if e.Skipped() != skipped {
			t.Fatal(e.Skipped(), skipped)
		}
		for id, row := range e.Rows() {
			var expected [2]GiaID
			copy(expected[:], neighbors[id])
			if row != expected {
				t.Fatal(id, row, expected)
			}
		}
	})
}

func TestGiaPackingEdgesBoundsAndSelfEdges(t *testing.T) {
	mapping := testGiaPackingMapping(t)
	packing, err := NewGiaLUTPacking(mapping, [][]GiaID{{3, 4, 5}, {3, 4, 5}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	before := packing.Packed()
	if result, err := packing.Edges(GiaEdgeOptions{MaxPairs: 3}); result != nil || err == nil {
		t.Fatal(result, err)
	}
	if _, err := packing.Edges(GiaEdgeOptions{MaxPairs: 4}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	if result, err := packing.Edges(GiaEdgeOptions{Stop: func() bool { calls++; return calls >= 3 }}); result != nil || err == nil {
		t.Fatal(result, err)
	}
	if !slices.Equal(before, packing.Packed()) {
		t.Fatal("source packing changed")
	}
	self, err := NewGiaLUTPacking(mapping, [][]GiaID{{3, 3, 3}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	edges, err := self.Edges(GiaEdgeOptions{})
	if err != nil || edges.Rows()[3] != [2]GiaID{3, 3} || edges.Skipped() != 2 || len(edges.Pairs()) != 0 {
		t.Fatal(edges, err)
	}
}
