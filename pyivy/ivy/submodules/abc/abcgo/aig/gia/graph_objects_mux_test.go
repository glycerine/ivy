package gia

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestObjectImportRejectsRealMuxMetadata(t *testing.T) {
	for signs := int32(0); signs < 8; signs++ {
		p := Gia_ManStart(1)
		control, then, otherwise := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		// Existing objects must still see a control table enabled later.
		Gia_ManAppendAnd(p, control, then)
		p.EnableMuxes()
		root := Gia_ManAppendMuxReal(p, control^(signs&1), then^((signs>>1)&1), otherwise^((signs>>2)&1))
		// Force both object and control-table growth after the MUX exists.
		for i := 0; i < 32; i++ {
			Gia_ManAppendBuf(p, root)
		}
		Gia_ManAppendCo(p, root)
		if g, err := GraphFromObjects(p.pObjs); err == nil {
			t.Errorf("signs %d: imported real MUX as an AND; assignment [true true false] gives %v", signs, g.Evaluate([]bool{true, true, false}))
		}
	}
}

func TestObjectImportAllowsAndWithMuxStorage(t *testing.T) {
	p := Gia_ManStart(1)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	p.EnableMuxes()
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b^1))
	g, err := GraphFromObjects(p.pObjs)
	if err != nil {
		t.Fatal("empty control table rejected ordinary graph:", err)
	}
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		if got := g.Evaluate(in)[0]; got != (in[0] && !in[1]) {
			t.Fatalf("assignment %d: %t", mask, got)
		}
	}
}

func TestRealMuxObjectEncodingMatchesSource(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("native fixture uses GNU section collection")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	src := filepath.Clean(filepath.Join(dir, "../../../src"))
	binary := filepath.Join(t.TempDir(), "mux-encoding")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	args := []string{"-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src, filepath.Join(dir, "test_vectors/object_mux_encoding.c"), "-Wl,--gc-sections", "-o", binary}
	if out, err := exec.CommandContext(ctx, cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile source encoding: %v: %s", err, out)
	}
	source, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run source encoding: %v: %s", err, source)
	}
	var got strings.Builder
	for signs := int32(0); signs < 8; signs++ {
		p := Gia_ManStart(8)
		control, yes, no := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		p.EnableMuxes()
		root := Gia_ManAppendMuxReal(p, control^(signs&1), yes^((signs>>1)&1), no^((signs>>2)&1))
		o := Gia_ManObj(p, root>>1)
		fmt.Fprintf(&got, "signs %d and %d xor %d mux %d pins %d %d control %d\n", signs, Gia_ObjIsAnd(o), Gia_ObjIsXor(o), Gia_ObjIsMux(p, o), Gia_ObjFaninLit0(o, root>>1), Gia_ObjFaninLit1(o, root>>1), p.PMuxes[root>>1])
		Gia_ManAppendCo(p, root)
		for mask := 0; mask < 8; mask++ {
			values, err := p.Evaluate([]bool{mask&1 != 0, mask&2 != 0, mask&4 != 0})
			if err != nil {
				t.Fatal(err)
			}
			value := 0
			if values[0] {
				value = 1
			}
			fmt.Fprintf(&got, "truth %d %d\n", mask, value)
		}
	}
	if got.String() != string(source) {
		t.Fatalf("Go MUX states:\n%sC MUX states:\n%s", got.String(), source)
	}
}
