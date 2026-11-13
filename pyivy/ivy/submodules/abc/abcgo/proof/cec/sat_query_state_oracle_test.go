package cec

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Compare the original CEC SAT frontier, status and model with the Go fixture.
func TestSourceSATPatternStateAgainstABC(t *testing.T) {
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
int main(void) {
 Gia_Man_t *g=Gia_ManStart(8); Cec_ParSat_t pars; Cec_ManSat_t *p;
 int a=Gia_ManAppendCi(g), s0,s1,s2; Gia_ManSetPhase(g);
 Cec_ManSatSetDefaultParams(&pars);p=Cec_ManSatCreate(g,&pars);
 s0=Cec_ManSatCheckNode(p,Gia_ManObj(g,a>>1));Vec_IntPush(p->vCex,17);
 s1=Cec_ManSatCheckNode(p,Gia_ManConst0(g));
 s2=Cec_ManSatCheckNode(p,Gia_Not(Gia_ManObj(g,a>>1)));
 printf("state:%d,%d,%d,%d,%d\n",s0,s1,s2,Vec_IntSize(p->vCex),Vec_IntEntry(p->vCex,0));
 Cec_ManSatStop(p);Gia_ManStop(g);return 0;
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
	if got, want := strings.TrimSpace(string(out)), "state:0,1,0,1,17"; got != want {
		t.Fatalf("ABC SAT frontier: %q; want %q", got, want)
	}
}
