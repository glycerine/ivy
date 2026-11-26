package gia

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestNativeMuxPosSkipsCoDrivers links test_vectors/muxpos_oracle.c against
// the native library named by ABC_LIBABC (with giaUtil.c from ../src) and
// checks that Gia_ManDupWithMuxPos adds no output for a node that already
// drives one (CBUGS_FIXED.md #6).
func TestNativeMuxPosSkipsCoDrivers(t *testing.T) {
	lib := os.Getenv("ABC_LIBABC")
	if lib == "" {
		t.Skip("set ABC_LIBABC to a native libabc.a built from ../src")
	}
	lib, _ = filepath.Abs(lib)
	src, _ := filepath.Abs("../../../src")
	dir := t.TempDir()
	bin := filepath.Join(dir, "muxpos")
	for _, c := range [][]string{
		{"gcc", "-O1", "-DABC_USE_STDINT_H", "-I", src, "-c", "test_vectors/muxpos_oracle.c", "-o", filepath.Join(dir, "main.o")},
		{"gcc", "-O1", "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "aig/gia/giaUtil.c"), "-o", filepath.Join(dir, "giaUtil.o")},
		{"g++", filepath.Join(dir, "main.o"), filepath.Join(dir, "giaUtil.o"), lib, "-lreadline", "-lpthread", "-lm", "-ldl", "-o", bin},
	} {
		if out, err := exec.Command(c[0], c[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c[:3], err, out)
		}
	}
	out, err := exec.Command(bin, "../../base/abci/test_vectors/bench/mul8.aig").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "dup=0") {
		t.Fatalf("native Gia_ManDupWithMuxPos: %v %s", err, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}

// The port adds the same outputs as the fixed ABC (174 on mul8, none for a
// node that already drives an output), and like release ABC it also runs on
// sequential GIAs.
func TestMuxPosSkipsCoDrivers(t *testing.T) {
	p, err := Gia_AigerRead("../../base/abci/test_vectors/bench/mul8.aig", 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	old := Gia_ManCoNum(p)
	q := Gia_ManDupWithMuxPos(p)
	dup := 0
	for i := old; i < Gia_ManCoNum(q); i++ {
		for k := int32(0); k < old; k++ {
			if Gia_ObjFaninId0p(q, Gia_ManCo(q, i)) == Gia_ObjFaninId0p(q, Gia_ManCo(q, k)) {
				dup++
				break
			}
		}
	}
	if added := Gia_ManCoNum(q) - old; added != 174 || dup != 0 {
		t.Fatalf("added %d outputs, %d duplicating a CO driver", added, dup)
	}
}
