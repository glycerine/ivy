package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaEdgeEditNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_edge_edit.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Before, After [2]GiaID
		Next          GiaID
		Kind          GiaEdgeEditKind
		Skipped       int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 156 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		source := &GiaEdgeAssignment{rows: make([][2]GiaID, 4)}
		source.rows[1] = row.Before
		r, err := source.WithEdits([]GiaEdgeEdit{{Object: 1, Neighbor: row.Next, Kind: row.Kind}}, GiaEdgeOptions{})
		if err != nil || r.rows[1] != row.After || r.Skipped() != row.Skipped {
			t.Fatal(row, r, err)
		}
		if source.rows[1] != row.Before {
			t.Fatal("source changed")
		}
	}
}
func FuzzGiaEdgeEdits(f *testing.F) {
	f.Add([]byte{0, 1, 2, 0, 1, 3, 1, 1, 2})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 99 {
			data = data[:99]
		}
		source, _ := GiaEdgesFromPairs(4, nil, GiaEdgeOptions{})
		edits := make([]GiaEdgeEdit, 0)
		expected := make([][2]GiaID, 4)
		skipped := 0
		valid := true
		for i := 0; i+2 < len(data); i += 3 {
			edit := GiaEdgeEdit{Object: GiaID(data[i+1] % 4), Neighbor: GiaID(data[i+2] % 4), Kind: GiaEdgeEditKind(data[i] % 3)}
			edits = append(edits, edit)
			row := &expected[edit.Object]
			switch edit.Kind {
			case GiaEdgeInsert:
				if row[0] == 0 {
					row[0] = edit.Neighbor
				} else if row[1] == 0 {
					row[1] = edit.Neighbor
				} else {
					skipped++
				}
			case GiaEdgeRemove:
				if row[0] != edit.Neighbor && row[1] != edit.Neighbor {
					valid = false
				}
				if row[0] == edit.Neighbor {
					row[0] = row[1]
				}
				row[1] = 0
			case GiaEdgeClear:
				*row = [2]GiaID{}
			}
		}
		result, err := source.WithEdits(edits, GiaEdgeOptions{})
		if valid {
			if err != nil || !slices.Equal(result.Rows(), expected) || result.Skipped() != skipped {
				t.Fatal(result, err)
			}
		} else if err == nil || result != nil {
			t.Fatal("missing-neighbor assertion")
		}
		for _, row := range source.Rows() {
			if row != [2]GiaID{} {
				t.Fatal("source changed")
			}
		}
	})
}
func TestGiaEdgeEditsRejectAtomically(t *testing.T) {
	source, _ := GiaEdgesFromPairs(4, nil, GiaEdgeOptions{})
	good := GiaEdgeEdit{Object: 1, Neighbor: 2, Kind: GiaEdgeInsert}
	for _, bad := range []GiaEdgeEdit{{Object: 4}, {Object: 1, Neighbor: 4}, {Object: 1, Kind: 99}, {Object: 1, Neighbor: 3, Kind: GiaEdgeRemove}} {
		if r, err := source.WithEdits([]GiaEdgeEdit{good, bad}, GiaEdgeOptions{}); r != nil || err == nil {
			t.Fatal(r, err)
		}
	}
	for _, options := range []GiaEdgeOptions{{MaxObjects: 1}, {MaxPairs: 1}, {MaxPairs: -1}, {Stop: func() bool { return true }}} {
		if r, err := source.WithEdits([]GiaEdgeEdit{good, good}, options); r != nil || err == nil {
			t.Fatal(r, err)
		}
	}
	calls := 0
	if r, err := source.WithEdits([]GiaEdgeEdit{good, good}, GiaEdgeOptions{Stop: func() bool { calls++; return calls >= 3 }}); r != nil || err == nil {
		t.Fatal(r, err)
	}
}
