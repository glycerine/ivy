package cec

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Compare source sequential synthesis option paths on a shared-transition domain.
func TestSequentialSynthesisOptionsAgainstABC(t *testing.T) {
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
int main(void){int mode;for(mode=0;mode<3;mode++){Gia_Man_t *g=Gia_ManStart(12);Cec_ParSeq_t p;Vec_Ptr_t *domains=Vec_PtrAlloc(1);Vec_Int_t *part=Vec_IntAlloc(2);int a=Gia_ManAppendCi(g),r0=Gia_ManAppendCi(g),r1=Gia_ManAppendCi(g);Gia_ManAppendCo(g,a);Gia_ManAppendCo(g,a);Gia_ManAppendCo(g,a);Gia_ManSetRegNum(g,2);Vec_IntPush(part,0);Vec_IntPush(part,1);Vec_PtrPush(domains,part);g->vClockDoms=domains;Cec_SeqSynthesisSetDefaultParams(&p);p.fUseScorr=mode==1;p.fUseLcorr=mode==2;int status=Cec_SequentialSynthesisPart(g,&p);printf("case:%d status:%d repr:%d expected:%d\n",mode,status,Gia_ObjRepr(g,r1>>1),r0>>1);Gia_ManStop(g);}return 0;}`
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
	for _, want := range []string{"case:0 status:1 repr:2 expected:2", "case:1 status:1 repr:2 expected:2", "case:2 status:1 repr:2 expected:2"} {
		if !strings.Contains(string(out), want) {
			t.Fatalf("ABC missing %q: %s", want, out)
		}
	}
}
