package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaLUTMappingNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_lut_mapping.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Stats          [3]int
		Maximum        int
		Packed, Levels []int
		Truths         [][2]uint64
		Objects        [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 20 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		graph := NewGiaManager()
		for _, obj := range row.Objects[1:] {
			switch obj[0] {
			case 1:
				_, err = graph.AppendCI()
			case 2:
				_, err = graph.AppendCO(GiaLit(obj[1]))
			case 3:
				_, err = graph.AppendAnd(GiaLit(obj[1]), GiaLit(obj[2]))
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		mapping, err := GiaLUTMappingFromPacked(graph.ObjectCount(), row.Packed, 0)
		if err != nil {
			t.Fatal(index, err)
		}
		stats := mapping.Statistics()
		if stats != (GiaLUTStatistics{row.Stats[0], row.Stats[1], row.Stats[2]}) {
			t.Fatal(stats, row.Stats)
		}
		levels, maximum, err := mapping.Levels(graph)
		if err != nil || maximum != row.Maximum || !slices.Equal(levels, row.Levels) {
			t.Fatal(index, levels, maximum, err)
		}
		for _, expected := range row.Truths {
			word, err := mapping.Truth6(graph, GiaID(expected[0]), GiaCutTruthOptions{})
			if err != nil || word != expected[1] {
				t.Fatal(index, expected, word, err)
			}
		}
		roots := mapping.Roots()
		if !slices.IsSorted(roots) {
			t.Fatal(roots)
		}
		records := make([]GiaLUTRecord, 0)
		// The immutable compatibility record uses reverse object order, then the optional constant.
		for i := len(roots) - 1; i >= 0; i-- {
			record, ok := mapping.Record(roots[i])
			if !ok {
				t.Fatal("missing record")
			}
			records = append(records, record)
		}
		if record, ok := mapping.Record(0); ok {
			records = append(records, record)
		}
		rebuilt, err := NewGiaLUTMapping(graph.ObjectCount(), records, 0)
		if err != nil || !slices.Equal(rebuilt.Packed(), row.Packed) {
			t.Fatal(index, "packed roundtrip", err)
		}
	}
}
func TestGiaLUTMappingValidationAndOwnership(t *testing.T) {
	records := []GiaLUTRecord{{Root: 3, Fanins: []GiaID{1, 2, 1}, MuxID: -3}}
	mapping, err := NewGiaLUTMapping(4, records, 9)
	if err != nil {
		t.Fatal(err)
	}
	records[0].Fanins[0] = 0
	packed := mapping.Packed()
	packed[3] = 0
	record, ok := mapping.Record(3)
	if !ok || record.Fanins[0] != 1 || record.MuxID != -3 {
		t.Fatal(record, ok)
	}
	record.Fanins[0] = 0
	fresh, _ := mapping.Record(3)
	if fresh.Fanins[0] != 1 {
		t.Fatal("aliased fanins")
	}
	for _, data := range [][]int{{0, 0, 0, 1}, {0, 0, 0, 99}, {0, 0, 0, 4, -1, 0}, {0, 0, 0, 4, 10, 0}, {0, 0, 0, 4, 1, 3, 3}, {0, 0, 0, 4, 1, -1, 3}, {0, 0, 0, 4, 0, 4}, {0, 4, 4, 0, 0, 0}} {
		if _, err := GiaLUTMappingFromPacked(4, data, 0); err == nil {
			t.Fatal("invalid packed accepted", data)
		}
	}
	if _, err := NewGiaLUTMapping(4, []GiaLUTRecord{{Root: 3}, {Root: 3}}, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewGiaLUTMapping(4, []GiaLUTRecord{{Root: 3, Fanins: []GiaID{3}}}, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewGiaLUTMapping(4, []GiaLUTRecord{{Root: 3, Fanins: []GiaID{1, 2, 1}}}, 8); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := GiaLUTMappingFromPacked(4, mapping.Packed(), 8); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	m := NewGiaManager()
	m.AppendCI()
	m.AppendCI()
	m.AppendAnd(2, 4)
	if _, _, err := mapping.Levels(m); err != nil {
		t.Fatal(err)
	}
	if _, err := mapping.Truth6(m, 2, GiaCutTruthOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	m.AppendCI()
	if _, _, err := mapping.Levels(m); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	var missing *GiaLUTMapping
	if missing.Statistics().MaxSize != -1 || missing.Packed() != nil || missing.Roots() != nil {
		t.Fatal("nil accessors")
	}
	if _, ok := missing.Record(0); ok {
		t.Fatal("nil record")
	}
	if _, _, err := missing.Levels(m); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}
func FuzzGiaLUTMappingPacked(f *testing.F) {
	f.Add([]byte{0, 0, 0, 4, 2, 1, 2, 3})
	f.Add([]byte{0, 0, 0, 4, 0, 3})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 128 {
			data = data[:128]
		}
		packed := make([]int, len(data))
		for i, v := range data {
			packed[i] = int(int8(v))
		}
		mapping, err := GiaLUTMappingFromPacked(4, packed, 128)
		if err != nil {
			return
		}
		copy := mapping.Packed()
		if !slices.Equal(copy, packed) {
			t.Fatal("roundtrip")
		}
		if len(packed) > 0 {
			packed[0] = 99
		}
		if !slices.Equal(copy, mapping.Packed()) {
			t.Fatal("input alias")
		}
		for _, root := range mapping.Roots() {
			record, ok := mapping.Record(root)
			if !ok {
				t.Fatal(root)
			}
			for _, id := range record.Fanins {
				if id >= root {
					t.Fatal(record)
				}
			}
		}
	})
}
