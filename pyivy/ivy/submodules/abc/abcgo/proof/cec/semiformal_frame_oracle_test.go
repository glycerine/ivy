package cec

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Characterize source behavior for nonpositive frame limits.
func TestSemiformalZeroFrameAgainstABC(t *testing.T) {
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
int main(void) { int limit,i,a,r,f; Gia_Man_t *g=Gia_ManStart(10),*q; Abc_Cex_t *init;
 a=Gia_ManAppendCi(g);r=Gia_ManAppendCi(g);Gia_ManAppendCo(g,r);Gia_ManAppendCo(g,a);Gia_ManSetRegNum(g,1);
 g->pReprs=ABC_CALLOC(Gia_Rpr_t,Gia_ManObjNum(g));for(i=0;i<Gia_ManObjNum(g);i++)Gia_ObjSetRepr(g,i,GIA_VOID);
 Gia_ObjSetRepr(g,r>>1,0);init=Abc_CexAlloc(1,0,0);
 for(limit=-1;limit<=0;limit++){f=-1;q=Gia_ManSpecReduceInitFrames(g,init,limit,&f,0,1);printf("limit:%d frames:%d outputs:%d\n",limit,f,Gia_ManPoNum(q));Gia_ManStop(q);}
 Abc_CexFree(init);Gia_ManStop(g);return 0;}`
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
	if got, want := strings.TrimSpace(string(out)), "limit:-1 frames:1 outputs:1\nlimit:0 frames:1 outputs:1"; got != want {
		t.Fatalf("ABC SAT frontier: %q; want %q", got, want)
	}
}
