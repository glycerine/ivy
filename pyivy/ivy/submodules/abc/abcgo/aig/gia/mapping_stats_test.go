package gia

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMappingStatsAppend(t *testing.T) {
	p := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	gate := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, gate)
	p.SetIdentity("mapping_stats", "")
	p.VMapping = Vec_IntStart(int(p.nObjs))
	id := gate >> 1
	Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
	for _, x := range []int32{2, a >> 1, b >> 1, id} {
		Vec_IntPush(p.VMapping, x)
	}
	path := filepath.Join(t.TempDir(), "mapping.txt")
	Gia_ManPrintMappingStats(p, path)
	Gia_ManPrintMappingStats(p, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "\nmapping_stats  1 1 1 ") || !strings.Contains(string(data), "  1 1 ") {
		t.Fatalf("unexpected statistics table: %q", data)
	}
}

func TestMappingMuxDataFaninCount(t *testing.T) {
	p := Gia_ManStart(12)
	control, a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	mux := Gia_ManAppendMux(p, control, a, b)
	Gia_ManAppendCo(p, mux)
	id := mux >> 1
	p.VMapping = Vec_IntStart(int(p.nObjs))
	Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
	for _, x := range []int32{3, control >> 1, a >> 1, b >> 1, -id} {
		Vec_IntPush(p.VMapping, x)
	}
	var fanins [3]int32
	Gia_ManPrintGetMuxFanins(p, Gia_ManObj(p, id), fanins[:])
	if fanins[0] != control>>1 || (fanins[1] != a>>1 && fanins[1] != b>>1) || (fanins[2] != a>>1 && fanins[2] != b>>1) || fanins[1] == fanins[2] {
		t.Fatalf("unexpected mux fanins: %v", fanins)
	}
	if got := Gia_ManCountDupLut(p); got != 2 {
		t.Fatalf("data fanin count = %d, want 2", got)
	}
	for _, obj := range p.pObjs {
		if obj.fMark0 != 0 || obj.fMark1 != 0 {
			t.Fatal("marks were not cleared")
		}
	}
	Gia_ManPrintMappingStats(p, "")
}
