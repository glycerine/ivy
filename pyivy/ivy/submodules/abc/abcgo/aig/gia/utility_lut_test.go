package gia

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLutBinaryAndDistribution(t *testing.T) {
	p := Gia_ManStart(10)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x^1)
	p.pName = "sample"
	p.vMapping = Vec_IntStart(int(p.nObjs))
	Vec_IntWriteEntry(p.vMapping, int(x>>1), int32(Vec_IntSize(p.vMapping)))
	Vec_IntPush(p.vMapping, 2)
	Vec_IntPush(p.vMapping, a>>1)
	Vec_IntPush(p.vMapping, b>>1)
	Vec_IntPush(p.vMapping, x>>1)
	dir := t.TempDir()
	path := filepath.Join(dir, "luts.bin")
	Gia_AigerWriteLut(p, path)
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var rows [5]Gia_MapLut_t
	if err := binary.Read(f, binary.NativeEndian, &rows); err != nil {
		t.Fatal(err)
	}
	if rows[0].Type != 3 || rows[0].PTruth[0] != ^uint32(0) || rows[3].Type != 3 || rows[3].NFans != 2 || rows[3].PFans[0] != 2 || rows[3].PFans[1] != 4 || rows[3].PTruth[0] != 0x88888888 || rows[4].PTruth[0] != 0x55555555 {
		t.Fatalf("records %+v", rows)
	}
	distribution := filepath.Join(dir, "sizes.txt")
	Gia_DumpLutSizeDistrib(p, distribution)
	contents, err := os.ReadFile(distribution)
	if err != nil || !strings.HasPrefix(string(contents), "sample 0 1 0 ") {
		t.Fatalf("distribution %q %v", contents, err)
	}
}
