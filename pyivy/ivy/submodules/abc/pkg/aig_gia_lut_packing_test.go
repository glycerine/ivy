package abc

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func TestGiaLUTPackingNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_lut_packing.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		SourceObjects, TargetObjects                 int
		SourceMapping, TargetMapping, Source, Target []int
		Stats                                        [6]int
		Correspondence                               []GiaLit
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 32 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		source, err := GiaLUTMappingFromPacked(row.SourceObjects, row.SourceMapping, 0)
		if err != nil {
			t.Fatal(err)
		}
		target, err := GiaLUTMappingFromPacked(row.TargetObjects, row.TargetMapping, 0)
		if err != nil {
			t.Fatal(err)
		}
		packing, err := GiaLUTPackingFromPacked(source, row.Source, 0)
		if err != nil {
			t.Fatal(err)
		}
		stats := packing.Statistics()
		if [6]int{stats.Groups, stats.LUTs, stats.MaxSize, stats.BySize[1], stats.BySize[2], stats.BySize[3]} != row.Stats {
			t.Fatal(index, stats)
		}
		rebuilt, err := NewGiaLUTPacking(source, packing.Groups(), 0)
		if err != nil || !slices.Equal(rebuilt.Packed(), row.Source) {
			t.Fatal(index, err)
		}
		result, err := packing.Transfer(target, row.Correspondence, GiaLUTTransferOptions{})
		if err != nil || !slices.Equal(result.Packed(), row.Target) {
			t.Fatal(index, result, err)
		}
		if !slices.Equal(packing.Packed(), row.Source) {
			t.Fatal("source changed")
		}
	}
}
func testGiaPackingMapping(t *testing.T) *GiaLUTMapping {
	t.Helper()
	m, err := NewGiaLUTMapping(6, []GiaLUTRecord{{Root: 3}, {Root: 4}, {Root: 5}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func TestGiaLUTPackingValidationAndOwnership(t *testing.T) {
	m := testGiaPackingMapping(t)
	groups := [][]GiaID{{3}, {4, 5}, {3, 4, 5}}
	p, err := NewGiaLUTPacking(m, groups, 10)
	if err != nil {
		t.Fatal(err)
	}
	groups[0][0] = 5
	copy := p.Groups()
	copy[0][0] = 4
	packed := p.Packed()
	packed[0] = 99
	if p.Groups()[0][0] != 3 {
		t.Fatal("group alias")
	}
	for _, data := range [][]int{nil, {-1}, {1}, {1, 0}, {1, 2, 3}, {1, 1, 2}, {0, 1, 3}, {1, 1, -1}, {1, 1, 6}, {2, 1, 3}} {
		if _, err := GiaLUTPackingFromPacked(m, data, 0); err == nil {
			t.Fatal("bad packing", data)
		}
	}
	for _, groups := range [][][]GiaID{{{}}, {{2}}} {
		if _, err := NewGiaLUTPacking(m, groups, 0); err == nil {
			t.Fatal(groups)
		}
	}
	wide, err := NewGiaLUTPacking(m, [][]GiaID{{3, 4, 5, 3}}, 0)
	if err != nil || wide.Statistics().MaxSize != 4 || wide.Statistics().BySize[4] != 1 {
		t.Fatal(wide, err)
	}
	if _, err = wide.Edges(GiaEdgeOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if _, err := NewGiaLUTPacking(m, p.Groups(), 9); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := GiaLUTPackingFromPacked(m, p.Packed(), 9); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	correspondence := []GiaLit{0, 2, 4, 6, 8, 10}
	for _, corr := range [][]GiaLit{nil, {0, 2, 4, InvalidGiaLit, 8, 10}, {0, 2, 4, 4, 8, 10}, {0, 2, 4, 12, 8, 10}} {
		if result, err := p.Transfer(m, corr, GiaLUTTransferOptions{}); result != nil || err == nil {
			t.Fatal(result, err)
		}
	}
	for _, options := range []GiaLUTTransferOptions{{MaxEntries: -1}, {MaxEntries: 9}, {Stop: func() bool { return true }}} {
		if result, err := p.Transfer(m, correspondence, options); result != nil || err == nil {
			t.Fatal(result, err)
		}
	}
	for _, threshold := range []int{2, 4, 8} {
		calls := 0
		if result, err := p.Transfer(m, correspondence, GiaLUTTransferOptions{Stop: func() bool { calls++; return calls >= threshold }}); result != nil || !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, result, err)
		}
	}
	var missing *GiaLUTPacking
	if missing.Packed() != nil || missing.Groups() != nil || missing.Statistics().MaxSize != -1 {
		t.Fatal("nil accessors")
	}
	if _, err := missing.Transfer(m, correspondence, GiaLUTTransferOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}
func FuzzGiaLUTPacking(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4})
	f.Add([]byte{0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 60 {
			data = data[:60]
		}
		m := testGiaPackingMapping(t)
		packed := make([]int, len(data))
		for i, v := range data {
			packed[i] = int(int8(v))
		}
		if p, err := GiaLUTPackingFromPacked(m, packed, 128); err == nil {
			if !slices.Equal(p.Packed(), packed) {
				t.Fatal("import roundtrip")
			}
		}
		groups := make([][]GiaID, 0)
		for i, v := range data {
			group := make([]GiaID, 1+int(v%3))
			for j := range group {
				group[j] = GiaID(3 + (i+j)%3)
			}
			groups = append(groups, group)
		}
		p, err := NewGiaLUTPacking(m, groups, 256)
		if err != nil {
			t.Fatal(err)
		}
		corr := []GiaLit{0, 2, 4, 11, 8, 7}
		target, err := p.Transfer(m, corr, GiaLUTTransferOptions{MaxEntries: 256})
		if err != nil {
			t.Fatal(err)
		}
		expected := make([][]GiaID, len(groups))
		for i, g := range groups {
			expected[i] = make([]GiaID, len(g))
			for j, id := range g {
				expected[i][j] = corr[id].ID()
			}
		}
		if !reflect.DeepEqual(target.Groups(), expected) {
			t.Fatal("transfer model", target.Groups(), expected)
		}
	})
}
