package fracomb_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/fracomb"
	"github.com/glycerine/abcgo/proof/framgr"
)

// Compare the complete rewrite/FRAIG fallback with the original C driver.
// The first SAT query must exhaust its 600-conflict budget in both versions.
func TestCecFallbackAgainstCompiledABC(t *testing.T) {
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
	if _, err := os.Stat(filepath.Join(src, "proof/fra/fraCec.o")); err != nil {
		t.Skip("prebuilt ABC objects unavailable")
	}
	var objects []string
	if err := filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
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
	st, archive := filepath.Join(dir, "st.o"), filepath.Join(dir, "abc.a")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", st).CombinedOutput(); err != nil {
		t.Fatalf("compile st: %v\n%s", err, out)
	}
	if out, err := exec.Command(ar, append([]string{"qcs", archive}, append(objects, st)...)...).CombinedOutput(); err != nil {
		t.Fatalf("archive: %v\n%s", err, out)
	}
	const program = `#include "proof/fra/fra.h"
#include "base/main/main.h"
#include <stdio.h>
int main(void) {
  Aig_Man_t *g; Aig_Obj_t *v[8][7], *f, *c;
  int i,j,h,first,result;
  Abc_Start(); g=Aig_ManStart(1000); f=Aig_ManConst1(g);
  for(i=0;i<8;i++) for(h=0;h<7;h++) v[i][h]=Aig_ObjCreateCi(g);
  for(i=0;i<8;i++) { c=Aig_ManConst0(g); for(h=0;h<7;h++) c=Aig_Or(g,c,v[i][h]); f=Aig_And(g,f,c); }
  for(h=0;h<7;h++) for(i=0;i<8;i++) for(j=0;j<i;j++) f=Aig_And(g,f,Aig_Or(g,Aig_Not(v[i][h]),Aig_Not(v[j][h])));
  Aig_ObjCreateCo(g,f);
  first=Fra_FraigSat(g,600,0,0,0,0,1,0,0,0); result=Fra_FraigCec(&g,100000,0);
  printf("oracle:%d,%d\n",first,result); Aig_ManStop(g);
  g=Aig_ManStart(10); v[0][0]=Aig_ObjCreateCi(g); v[0][1]=Aig_ObjCreateCi(g);
  Aig_ObjCreateCo(g,Aig_And(g,v[0][0],v[0][1]));
  first=Fra_FraigMiterStatus(g); result=Fra_FraigCec(&g,100000,0);
  printf("direct:%d,%d,%d,%d\n",first,result,((int *)g->pData)[0],((int *)g->pData)[1]);
  Aig_ManStop(g); Abc_Stop(); return 0;
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
		t.Fatalf("ABC: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "oracle:-1,1") {
		t.Fatalf("unexpected ABC result: %s", out)
	}
	if !strings.Contains(string(out), "direct:-1,0,1,1") {
		t.Fatalf("unexpected ABC direct SAT result: %s", out)
	}
	direct := aig.NewGraph()
	x, y := direct.NewInput(), direct.NewInput()
	direct.AddOutput(direct.AND(x, y))
	if status := fracomb.Fra_FraigMiterStatus(direct); status != -1 {
		t.Fatalf("direct structural status = %d, want -1", status)
	}
	original := direct
	directCalls := 0
	if status := fracomb.Fra_FraigCec(&direct, 100000, 0, func(*aig.Graph, int32, int32) *aig.Graph { directCalls++; return nil }); status != 0 {
		t.Fatalf("direct SAT status = %d, want 0", status)
	}
	model, ok := direct.Data.([]int32)
	if direct != original || directCalls != 0 || !ok || len(model) != 2 || model[0] != 1 || model[1] != 1 {
		t.Fatalf("direct graph retained=%v, sweeps=%d, model=%v; ABC retained graph with [1 1]", direct == original, directCalls, direct.Data)
	}
	g := aig.NewGraph()
	var v [8][7]aig.Edge
	for i := range v {
		for h := range v[i] {
			v[i][h] = g.NewInput()
		}
	}
	f := g.Const1()
	for i := range v {
		c := g.Const0()
		for _, x := range v[i] {
			c = g.OR(c, x)
		}
		f = g.AND(f, c)
	}
	for h := 0; h < 7; h++ {
		for i := 0; i < 8; i++ {
			for j := 0; j < i; j++ {
				f = g.AND(f, g.OR(v[i][h].Not(), v[j][h].Not()))
			}
		}
	}
	g.AddOutput(f)
	if got := fracomb.Fra_FraigSat(g, 600, 0, 0, 0, 0, 1, 0, 0, 0); got != -1 {
		t.Fatalf("initial SAT = %d, want -1", got)
	}
	calls := 0
	result := fracomb.Fra_FraigCec(&g, 100000, 0, func(g *aig.Graph, n, m int32) *aig.Graph { calls++; return framgr.SweepForCEC(g, n, m) })
	if result != 1 || calls == 0 {
		t.Fatalf("CEC result=%d sweeps=%d, ABC result=1", result, calls)
	}
}
