package cec

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ABC's ordinary duplication, used by both classic CEC entry points, turns
// real XOR/MUX objects into two-fanin ANDs; buffers remain buffers. This is a
// characterization of the C source, not a claim that the resulting graph
// preserves the original logic.
func TestNativeDuplicationAgainstABC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ABC object oracle requires Linux")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	ar, err := exec.LookPath("ar")
	if err != nil {
		t.Skip("ar unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	if _, err = os.Stat(filepath.Join(src, "aig/gia/giaDup.o")); err != nil {
		t.Skip("prebuilt ABC objects unavailable")
	}
	var objects []string
	if err = filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".o") && p != filepath.Join(src, "misc/st/st.o") {
			objects = append(objects, p)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st := filepath.Join(dir, "st.o")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", st).CombinedOutput(); err != nil {
		t.Fatalf("compile st: %v\n%s", err, out)
	}
	archive := filepath.Join(dir, "abc.a")
	if out, err := exec.Command(ar, append([]string{"qcs", archive}, append(objects, st)...)...).CombinedOutput(); err != nil {
		t.Fatalf("archive: %v\n%s", err, out)
	}
	const program = `#include "aig/gia/gia.h"
#include <stdio.h>
int main(void) {
 Gia_Man_t *p=Gia_ManStart(32), *q; int a=Gia_ManAppendCi(p), b=Gia_ManAppendCi(p), c=Gia_ManAppendCi(p);
 int x, m, z; p->pMuxes=ABC_CALLOC(unsigned, p->nObjsAlloc);
 x=Gia_ManAppendXorReal(p,a,b); m=Gia_ManAppendMuxReal(p,c,a,b); z=Gia_ManAppendBuf(p,a);
 Gia_ManAppendCo(p,x); Gia_ManAppendCo(p,m); Gia_ManAppendCo(p,z);
 q=Gia_ManDup(p);
 printf("xor:%d mux:%d buf:%d and:%d,%d muxdata:%d\n", Gia_ObjIsXor(Gia_ManObj(q,Abc_Lit2Var(x))), Gia_ObjIsMux(q,Gia_ManObj(q,Abc_Lit2Var(m))), Gia_ObjIsBuf(Gia_ManObj(q,Abc_Lit2Var(z))), Gia_ObjIsAndReal(q,Gia_ManObj(q,Abc_Lit2Var(x))), Gia_ObjIsAndReal(q,Gia_ManObj(q,Abc_Lit2Var(m))), q->pMuxes != NULL);
 Gia_ManStop(q); Gia_ManStop(p); return 0;
}`
	cfile, binary := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle")
	if err := os.WriteFile(cfile, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, cfile, archive, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("link: %v\n%s", err, out)
	}
	out, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	if got, want := strings.TrimSpace(string(out)), "xor:0 mux:0 buf:1 and:1,1 muxdata:0"; got != want {
		t.Fatalf("ABC duplicate: %q; want %q", got, want)
	}
}
