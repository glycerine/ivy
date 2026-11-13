package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNativeExportsAgainstCompiledABC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ABC oracle requires Linux")
	}
	cc, e := exec.LookPath("gcc")
	if e != nil {
		t.Skip("gcc unavailable")
	}
	ar, e := exec.LookPath("ar")
	if e != nil {
		t.Skip("ar unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	if _, e = os.Stat(filepath.Join(src, "aig/gia/giaDup.o")); e != nil {
		t.Skip("ABC objects unavailable")
	}
	dir := t.TempDir()
	var objs []string
	if e = filepath.WalkDir(src, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "misc/st/st.o") {
			objs = append(objs, path)
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	st := filepath.Join(dir, "st.o")
	if out, e := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", st).CombinedOutput(); e != nil {
		t.Fatalf("st: %v %s", e, out)
	}
	objs = append(objs, st)
	lib := filepath.Join(dir, "abc.a")
	if out, e := exec.Command(ar, append([]string{"qcs", lib}, objs...)...).CombinedOutput(); e != nil {
		t.Fatalf("archive: %v %s", e, out)
	}
	exe := filepath.Join(dir, "oracle")
	fixture := filepath.Join(filepath.Dir(file), "test_vectors/native_export_oracle.c")
	if out, e := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, fixture, lib, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", exe).CombinedOutput(); e != nil {
		t.Fatalf("link: %v %s", e, out)
	}
	c, e := exec.Command(exe).CombinedOutput()
	if e != nil {
		t.Fatalf("run: %v %s", e, c)
	}
	g := Gia_ManStart(24)
	a, b, control, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	g.EnableMuxes()
	x := Gia_ManAppendXorReal(g, a, b^1)
	m := Gia_ManAppendMuxReal(g, control, b, d^1)
	Gia_ManAppendCo(g, x)
	Gia_ManAppendCo(g, m)
	funcs := []func(*Gia_Man_t) *aig.Graph{func(p *Gia_Man_t) *aig.Graph { return Gia_ManToAig(p, 0) }, func(p *Gia_Man_t) *aig.Graph { return Gia_ManToAigSkip(p, 1) }, Gia_ManToAigSimple}
	var result strings.Builder
	for i, fn := range funcs {
		out := fn(g)
		for mask := 0; mask < 16; mask++ {
			in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			v, err := out.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Fprintf(&result, "native %d %d %d %d\n", i, mask, boolInt(v[0]), boolInt(v[1]))
		}
	}
	if string(c) != result.String() {
		t.Fatalf("C:\n%sGo:\n%s", c, result.String())
	}
	orderExe := filepath.Join(dir, "order")
	orderFixture := filepath.Join(filepath.Dir(file), "test_vectors/simple_export_order_oracle.c")
	if out, e := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, orderFixture, lib, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", orderExe).CombinedOutput(); e != nil {
		t.Fatalf("order link: %v %s", e, out)
	}
	order, e := exec.Command(orderExe).CombinedOutput()
	if e != nil || strings.TrimSpace(string(order)) != "node=6 late=8 ci2=4" {
		t.Fatalf("order C: %v %s", e, order)
	}
	duplicate := exec.Command("sh", "-c", `ulimit -c 0; exec "$1" duplicate`, "sh", orderExe)
	out, e := duplicate.CombinedOutput()
	if e == nil || !strings.Contains(string(out), "Aig_ObjId(ppNodes[i]) == i") {
		t.Fatalf("duplicate source assertion: %v %s", e, out)
	}

}
