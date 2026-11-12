package gia

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenDumpFormats(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b))
	base := filepath.Join(t.TempDir(), "sample")
	Gia_ManDumpPlaFiles(p, 1, 1, 0, base)
	rows, err := giaReadPatterns(base + ".train.pla")
	if err != nil || len(rows) != 64 {
		t.Fatalf("rows=%d err=%v", len(rows), err)
	}
	for _, row := range rows {
		if row[2] != row[0]&row[1] {
			t.Fatalf("incorrect PLA row %v", row)
		}
	}
	Gia_ManDumpFiles(p, 1, 1, 0, base)
	x, err := os.ReadFile(base + "_x.train.data")
	if err != nil {
		t.Fatal(err)
	}
	y, err := os.ReadFile(base + "_y.train.data")
	if err != nil {
		t.Fatal(err)
	}
	if len(x) != 28 || len(y) != 20 || binary.NativeEndian.Uint32(x[:4]) != 2 || binary.NativeEndian.Uint32(x[4:8]) != 64 || binary.NativeEndian.Uint32(x[8:12]) != 2 || binary.NativeEndian.Uint32(y[8:12]) != 1 {
		t.Fatalf("invalid data headers")
	}
	bit := func(b []byte, i int) byte { return b[12+i/8] >> uint(i%8) & 1 }
	for i := 0; i < 64; i++ {
		if bit(y, i) != bit(x, 2*i)&bit(x, 2*i+1) {
			t.Fatalf("wrong binary pattern %d", i)
		}
	}
	list, err := os.ReadFile(base + ".flist")
	if err != nil || len(strings.Fields(string(list))) != 4 {
		t.Fatalf("file list %q %v", list, err)
	}
}
