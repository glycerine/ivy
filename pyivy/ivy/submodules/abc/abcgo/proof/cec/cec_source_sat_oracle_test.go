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
func TestSourceSATFrontierAgainstABC(t *testing.T) {
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
 Gia_Man_t *g=Gia_ManStart(16); Cec_ParSat_t pars; Cec_ManSat_t *p;
 int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),c=Gia_ManAppendCi(g);
 int mid=Gia_ManAppendAnd(g,a,b), root=Gia_ManAppendAnd(g,mid,c), status;
 Gia_ManSetPhase(g); Cec_ManSatSetDefaultParams(&pars); p=Cec_ManSatCreate(g,&pars);
 status=Cec_ManSatCheckNode(p,Gia_ManObj(g,root>>1));
 printf("status:%d ids:%d,%d,%d,%d,%d vars:%d model:%d,%d,%d\n",status,
 p->pSatVars[a>>1],p->pSatVars[b>>1],p->pSatVars[c>>1],p->pSatVars[mid>>1],p->pSatVars[root>>1],p->nSatVars,
 sat_solver_var_value(p->pSat,p->pSatVars[a>>1]),sat_solver_var_value(p->pSat,p->pSatVars[b>>1]),sat_solver_var_value(p->pSat,p->pSatVars[c>>1]));
 Cec_ManSatStop(p); Gia_ManStop(g); return 0;
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
	if got, want := strings.TrimSpace(string(out)), "status:0 ids:4,5,3,0,2 vars:6 model:1,1,1"; got != want {
		t.Fatalf("ABC SAT frontier: %q; want %q", got, want)
	}
}
