package cec

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

// C returns zero for its incomplete stub and fails to build independent choices.
func TestCec4IncompleteChoiceBuilderAgainstABC(t *testing.T) {
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
#include <string.h>
#include <stdio.h>
extern int Cec4_ManSatSolverAnd_rec(Gia_Man_t *,Gia_Man_t *,Gia_Man_t *,int);
extern Gia_Man_t * Cec4_ManSatSolverChoices(Gia_Man_t *,Gia_Man_t *);
int main(int argc,char **argv) { Gia_Man_t *p,*q,*dst;int a,b,c,x,y,i;
 if (argc>1 && !strcmp(argv[1],"stub")) {printf("stub:%d\n",Cec4_ManSatSolverAnd_rec(NULL,NULL,NULL,0));return 0;}
 p=Gia_ManStart(16);a=Gia_ManAppendCi(p);b=Gia_ManAppendCi(p);c=Gia_ManAppendCi(p);x=Gia_ManAppendAnd(p,Gia_ManAppendAnd(p,a,b),c);y=Gia_ManAppendAnd(p,a,Gia_ManAppendAnd(p,b,c));Gia_ManAppendCo(p,x);Gia_ManAppendCo(p,y);
 p->pReprs=ABC_CALLOC(Gia_Rpr_t,Gia_ManObjNum(p));for(i=0;i<Gia_ManObjNum(p);i++)Gia_ObjSetRepr(p,i,GIA_VOID);Gia_ObjSetRepr(p,y>>1,x>>1);p->pNexts=Gia_ManDeriveNexts(p);q=Gia_ManDup(p);dst=Cec4_ManSatSolverChoices(p,q);printf("result:%p\n",(void*)dst);return 0;}`
	cfile, binary := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle")
	if err := os.WriteFile(cfile, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, cfile, archive, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("link: %v\n%s", err, out)
	}
	if out, err := exec.Command(binary, "stub").CombinedOutput(); err != nil || strings.TrimSpace(string(out)) != "stub:0" {
		t.Fatalf("C stub: %v %s", err, out)
	}
	cmd := exec.Command(binary)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	failure, ok := err.(*exec.ExitError)
	if !ok || !failure.Sys().(syscall.WaitStatus).Signaled() {
		t.Fatalf("ABC independent builder did not fail: %v %s", err, out)
	}
}
