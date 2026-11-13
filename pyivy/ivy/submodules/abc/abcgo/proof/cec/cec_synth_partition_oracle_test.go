package cec

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Check partition traversal epochs against compiled original ABC.
func TestRegisterPartitionTraversalAgainstABC(t *testing.T) {
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
	const program = `#include "proof/cec/cecInt.h"
#include <stdio.h>
#include <string.h>
extern Gia_Man_t * Gia_ManRegCreatePart(Gia_Man_t *,Vec_Int_t *,int *,int *,int **);
int main(int argc,char **argv){Gia_Man_t *g=Gia_ManStart(10),*q;Vec_Int_t *v=Vec_IntAlloc(1);int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g);Gia_ManAppendCi(g);if(argc>1 && !strcmp(argv[1],"xor")){int x=Gia_ManAppendXorReal(g,a,b);Gia_ManAppendCo(g,x);Gia_ManSetRegNum(g,1);Vec_IntPush(v,0);q=Gia_ManRegCreatePart(g,v,NULL,NULL,NULL);printf("input-xor:%d output-xor:%d\n",Gia_ObjIsXor(Gia_ManObj(g,x>>1)),Gia_ObjIsXor(Gia_ObjFanin0(Gia_ManCo(q,1))));return 0;}Gia_ManAppendCo(g,a);Gia_ManSetRegNum(g,1);Gia_ManIncrementTravId(g);Gia_ObjSetTravIdCurrent(g,Gia_ManObj(g,b>>1));Vec_IntPush(v,0);q=Gia_ManRegCreatePart(g,v,NULL,NULL,NULL);printf("previous:%d\n",Gia_ObjIsTravIdPrevious(g,Gia_ManObj(g,b>>1)));Gia_ManStop(q);Gia_ManStop(g);Vec_IntFree(v);return 0;}`
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
	if got, want := strings.TrimSpace(string(out)), "previous:0"; got != want {
		t.Fatalf("ABC SAT frontier: %q; want %q", got, want)
	}
	out, err = exec.Command(binary, "xor").CombinedOutput()
	if err != nil || strings.TrimSpace(string(out)) != "input-xor:1 output-xor:0" {
		t.Fatalf("ABC native XOR partition: %v %s", err, out)
	}
}
