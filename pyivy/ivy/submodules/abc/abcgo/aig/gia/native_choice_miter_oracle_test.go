package gia

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNativeChoiceMiterAgainstCompiledABC(t *testing.T) {
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
	fixture := filepath.Join(filepath.Dir(file), "test_vectors/native_choice_miter_oracle.c")
	if out, e := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, fixture, lib, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", exe).CombinedOutput(); e != nil {
		t.Fatalf("link: %v %s", e, out)
	}
	source, e := exec.Command(exe).CombinedOutput()
	if e != nil {
		t.Fatalf("run: %v %s", e, source)
	}
	snapshots := make([]*Gia_Man_t, 3)
	// C receives its own AND expansions; Go receives the equivalent real
	// XOR/MUX representation. Compare the resulting miter's size and truth.
	for i := range snapshots {
		g := Gia_ManStart(24)
		a, b, c, d := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		g.EnableMuxes()
		x := Gia_ManAppendXorReal(g, a, b^1)
		m := Gia_ManAppendMuxReal(g, c, b, d^1)
		Gia_ManAppendCo(g, x^int32(i&1))
		Gia_ManAppendCo(g, m^int32((i>>1)&1))
		snapshots[i] = g
	}
	out := Gia_ManChoiceMiter(snapshots)
	var got strings.Builder
	fmt.Fprintf(&got, "nodes:%d,outputs:%d\n", Gia_ManObjNum(out), Gia_ManCoNum(out))
	for mask := 0; mask < 16; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
		values, e := out.Evaluate(in)
		if e != nil {
			t.Fatal(e)
		}
		fmt.Fprintf(&got, "case:%d", mask)
		for _, v := range values {
			if v {
				got.WriteString(",1")
			} else {
				got.WriteString(",0")
			}
		}
		got.WriteByte('\n')
	}
	if string(source) != got.String() {
		t.Fatalf("C:\n%sGo:\n%s", source, got.String())
	}
}
