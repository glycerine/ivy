package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaLUTTransferNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_lut_transfer.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		SourceObjects, TargetObjects int
		Correspondence               []GiaLit
		Source, Target               []int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 32 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		mapping, err := GiaLUTMappingFromPacked(row.SourceObjects, row.Source, 0)
		if err != nil {
			t.Fatal(err)
		}
		before := slices.Clone(row.Correspondence)
		result, err := mapping.Transfer(row.TargetObjects, row.Correspondence, GiaLUTTransferOptions{})
		if err != nil || !slices.Equal(result.Packed(), row.Target) {
			t.Fatal(i, result, err)
		}
		if !slices.Equal(mapping.Packed(), row.Source) || !slices.Equal(before, row.Correspondence) {
			t.Fatal("source changed")
		}
		if _, err := GiaLUTMappingFromPacked(row.TargetObjects, result.Packed(), 0); err != nil {
			t.Fatal("invalid target", err)
		}
	}
}
func TestGiaLUTTransferErrors(t *testing.T) {
	mapping, err := NewGiaLUTMapping(4, []GiaLUTRecord{{Root: 3, Fanins: []GiaID{1, 2}, MuxID: 3}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	valid := []GiaLit{0, 2, 4, 6}
	for _, correspondence := range [][]GiaLit{{0}, {0, 2, 4, 7}, {0, 2, 4, 8}, {0, 6, 4, 6}} {
		if result, err := mapping.Transfer(4, correspondence, GiaLUTTransferOptions{}); result != nil || err == nil {
			t.Fatal(result, err)
		}
	}
	footerMap, err := NewGiaLUTMapping(4, []GiaLUTRecord{{Root: 3, Fanins: []GiaID{1, 2}, MuxID: -1}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := footerMap.Transfer(4, []GiaLit{0, InvalidGiaLit, 4, 6}, GiaLUTTransferOptions{}); !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(err)
	}
	for _, options := range []GiaLUTTransferOptions{{MaxEntries: -1}, {MaxEntries: 7}, {Stop: func() bool { return true }}} {
		if result, err := mapping.Transfer(4, valid, options); result != nil || err == nil {
			t.Fatal(result, err)
		}
	}
	for _, threshold := range []int{2, 3, 4} {
		calls := 0
		if result, err := mapping.Transfer(4, valid, GiaLUTTransferOptions{Stop: func() bool { calls++; return calls >= threshold }}); result != nil || !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, result, err)
		}
	}
	var missing *GiaLUTMapping
	if _, err := missing.Transfer(4, valid, GiaLUTTransferOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}
func FuzzGiaLUTTransfer(f *testing.F) {
	f.Add(uint8(0))
	f.Add(uint8(255))
	f.Fuzz(func(t *testing.T, flags uint8) {
		records := []GiaLUTRecord{{Root: 3, Fanins: []GiaID{1, 2, 1}, MuxID: -3}, {Root: 4, Fanins: []GiaID{1, 2}, MuxID: 4}}
		mapping, err := NewGiaLUTMapping(5, records, 0)
		if err != nil {
			t.Fatal(err)
		}
		correspondence := []GiaLit{0, 2, 4, 6, 8}
		if flags&1 != 0 {
			correspondence[1] = InvalidGiaLit
		}
		if flags&2 != 0 {
			correspondence[2] ^= 1
		}
		if flags&4 != 0 {
			correspondence[3] = InvalidGiaLit
		}
		if flags&8 != 0 {
			correspondence[3] = 8
		}
		if flags&16 != 0 {
			correspondence[4] = InvalidGiaLit
		}
		result, err := mapping.Transfer(5, correspondence, GiaLUTTransferOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := GiaLUTMappingFromPacked(5, result.Packed(), 0); err != nil {
			t.Fatal(err)
		}
		expected := make([]GiaLUTRecord, 5)
		present := make([]bool, 5)
		for _, r := range records {
			lit := correspondence[r.Root]
			if lit == InvalidGiaLit {
				continue
			}
			r.Root = lit.ID()
			r.MuxID = int(lit.ID())
			if r.Root == correspondence[3].ID() && r.MuxID == int(correspondence[3].ID()) && len(r.Fanins) == 3 {
				r.MuxID = -r.MuxID
			}
			leaves := make([]GiaID, 0)
			for _, id := range r.Fanins {
				if value := correspondence[id]; value != InvalidGiaLit {
					leaves = append(leaves, value.ID())
				}
			}
			r.Fanins = leaves
			expected[r.Root] = r
			present[r.Root] = true
		}
		for id := 1; id < 5; id++ {
			r, ok := result.Record(GiaID(id))
			if ok != present[id] || ok && (!slices.Equal(r.Fanins, expected[id].Fanins) || r.MuxID != expected[id].MuxID) {
				t.Fatal(id, r, expected[id])
			}
		}
	})
}
