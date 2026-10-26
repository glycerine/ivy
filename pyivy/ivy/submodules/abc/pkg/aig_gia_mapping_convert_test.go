package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaMappingConvertNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_mapping_convert.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Input, Output []int
		Fanins        [][]GiaID
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 24 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		mapping, e := GiaLUTMappingFromPacked(16, row.Input, 0)
		if e != nil {
			t.Fatal(e)
		}
		before := mapping.Packed()
		n, e := mapping.AlternateNetwork(GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(i, e)
		}
		for id, f := range n.fanins {
			if !slices.Equal(f, row.Fanins[id]) {
				t.Fatal(i, id, f)
			}
		}
		r, e := n.PackedMapping(len(row.Output), GiaLUTAnalysisOptions{})
		if e != nil || !slices.Equal(r.Packed(), row.Output) {
			t.Fatal(i, r, e)
		}
		again, e := r.AlternateNetwork(GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		p, e := again.PackedMapping(0, GiaLUTAnalysisOptions{})
		if e != nil || !slices.Equal(p.Packed(), row.Output) {
			t.Fatal(i, p, e)
		}
		if !slices.Equal(before, mapping.Packed()) {
			t.Fatal("source changed")
		}
		n.fanins[4][0] = 3
		if !slices.Equal(r.Packed(), row.Output) {
			t.Fatal("output alias")
		}
	}
}
func TestGiaPackedUnitRemapPass(t *testing.T) {
	m, n := unitPassFixture(t, 0)
	mapping, e := n.PackedMapping(0, GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	before := mapping.Packed()
	r, e := mapping.RemapUnitPass(m, 0, GiaUnitRemapPassOptions{WindowLimit: 16})
	if e != nil || len(r.Mapping.Roots()) != 2 {
		t.Fatal(r, e)
	}
	checkGiaLUTComposition(t, m, r.Pass.Network, r.Pass.Functions)
	if !slices.Equal(before, mapping.Packed()) {
		t.Fatal("source changed")
	}
	back, e := r.Mapping.Network(GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	for id, f := range back.fanins {
		if !slices.Equal(f, r.Pass.Network.fanins[id]) {
			t.Fatal(id)
		}
	}
	for _, o := range []GiaUnitRemapPassOptions{{MaxPackedEntries: -1}, {MaxPackedEntries: 1}, {Search: GiaUnitWindowSearchOptions{Stop: func() bool { return true }}}} {
		r, e := mapping.RemapUnitPass(m, 0, o)
		if e == nil || r.Mapping != nil || r.Pass.Network != nil {
			t.Fatal(r, e)
		}
	}
}
func TestGiaMappingConvertLimits(t *testing.T) {
	m, n := unitPassFixture(t, 0)
	_ = m
	for _, o := range []GiaLUTAnalysisOptions{{MaxObjects: 1}, {MaxVisits: 1}, {MaxVisits: -1}, {Stop: func() bool { return true }}} {
		r, e := n.PackedMapping(0, o)
		if e == nil || r != nil {
			t.Fatal(r, e)
		}
	}
	for _, cap := range []int{-1, 1} {
		if r, e := n.PackedMapping(cap, GiaLUTAnalysisOptions{}); e == nil || r != nil {
			t.Fatal(r, e)
		}
	}
	bad := &GiaLUTNetwork{fanins: [][]GiaID{nil, {1}}}
	if _, e := bad.PackedMapping(0, GiaLUTAnalysisOptions{}); e == nil {
		t.Fatal("self fanin")
	}
}
func FuzzGiaMappingConvert(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 32 {
			data = data[:32]
		}
		var records []GiaLUTRecord
		for i, b := range data {
			root := GiaID(i + 1)
			var fanins []GiaID
			for j := 0; j < int(b%8); j++ {
				fanins = append(fanins, GiaID((int(b)+j)%int(root)))
			}
			records = append(records, GiaLUTRecord{Root: root, Fanins: fanins, MuxID: -int(root)})
		}
		mapping, e := NewGiaLUTMapping(len(data)+1, records, 0)
		if e != nil {
			t.Fatal(e)
		}
		n, e := mapping.AlternateNetwork(GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		packed, e := n.PackedMapping(0, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		next, e := packed.AlternateNetwork(GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		second, e := next.PackedMapping(0, GiaLUTAnalysisOptions{})
		if e != nil || !slices.Equal(packed.Packed(), second.Packed()) {
			t.Fatal(e)
		}
		for _, rec := range records {
			got, ok := packed.Record(rec.Root)
			if ok != (len(rec.Fanins) > 0) {
				t.Fatal(rec, ok)
			}
			if ok && (!slices.Equal(got.Fanins, rec.Fanins) || got.MuxID != int(rec.Root)) {
				t.Fatal(rec, got)
			}
		}
	})
}
