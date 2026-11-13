package cec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/cnf"
)

// The split source derives CNF by converting GIA to AIG and calling
// Cnf_Derive(aig, 0); compare its clauses and source-object variable map.
func TestSplitCNFAgainstCompiledABC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ABC object oracle requires Linux")
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
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	src := filepath.Join(root, "src")
	if _, e = os.Stat(filepath.Join(src, "proof/cec/cecCore.o")); e != nil {
		t.Skip("prebuilt ABC object files unavailable")
	}
	dir := t.TempDir()
	var objs []string
	e = filepath.WalkDir(src, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "misc/st/st.o") && path != filepath.Join(src, "proof/cec/cecSplit.o") {
			objs = append(objs, path)
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	st := filepath.Join(dir, "st.o")
	if out, e := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", st).CombinedOutput(); e != nil {
		t.Fatalf("st: %v %s", e, out)
	}
	objs = append(objs, st)
	lib := filepath.Join(dir, "abc.a")
	args := append([]string{"qcs", lib}, objs...)
	if out, e := exec.Command(ar, args...).CombinedOutput(); e != nil {
		t.Fatalf("archive: %v %s", e, out)
	}
	const program = `#include "aig/gia/gia.h"
#include "aig/gia/giaAig.h"
#include "sat/cnf/cnf.h"
#include <stdio.h>
int Cec_GiaSplitTest(Gia_Man_t *p,int a,int b,int c,int d,int e,int f,int g){return -1;}
int main(void) { Gia_Man_t *g=Gia_ManStart(16);int x=Gia_ManAppendCi(g),y=Gia_ManAppendCi(g),z=Gia_ManAppendAnd(g,x,y),i,j;Aig_Man_t *a;Cnf_Dat_t *c;
 Gia_ManAppendCo(g,z);a=Gia_ManToAigSimple(g);a->nRegs=0;c=Cnf_Derive(a,0);
 printf("vars %d\n",c->nVars);for(i=0;i<Gia_ManObjNum(g);i++)printf("map %d %d\n",i,c->pVarNums[i]);
 for(i=0;i<c->nClauses;i++){printf("clause");for(j=0;c->pClauses[i]+j<c->pClauses[i+1];j++)printf(" %d",c->pClauses[i][j]);puts("");}
 Cnf_DataFree(c);Aig_ManStop(a);Gia_ManStop(g);return 0;}`
	cfile := filepath.Join(dir, "oracle.c")
	exe := filepath.Join(dir, "oracle")
	if e = os.WriteFile(cfile, []byte(program), 0600); e != nil {
		t.Fatal(e)
	}
	args = []string{"-DABC_USE_STDINT_H", "-I", src, cfile, lib, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", exe}
	if out, e := exec.Command(cc, args...).CombinedOutput(); e != nil {
		t.Fatalf("link: %v %s", e, out)
	}
	out, e := exec.Command(exe).CombinedOutput()
	if e != nil {
		t.Fatalf("run: %v %s", e, out)
	}
	var vars int
	var maps []int32
	var clauses [][]int32
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "vars":
			fmt.Sscan(fields[1], &vars)
		case "map":
			var id, v int
			fmt.Sscan(fields[1], &id)
			fmt.Sscan(fields[2], &v)
			maps = append(maps, int32(v))
		case "clause":
			var c []int32
			for _, f := range fields[1:] {
				var v int32
				fmt.Sscan(f, &v)
				c = append(c, v)
			}
			clauses = append(clauses, c)
		}
	}
	g := gia.Gia_ManStart(16)
	x, y := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	z := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, z)
	data := cnf.Cnf_Derive(gia.Gia_ManToAigSimple(g), 0)
	if data == nil || data.Variables != vars || !reflect.DeepEqual(data.Clauses, clauses) {
		t.Fatalf("C vars=%d clauses=%v; Go %+v", vars, clauses, data)
	}
	if len(data.VarNums) < len(maps) || !reflect.DeepEqual(data.VarNums[:len(maps)], maps) {
		t.Fatalf("C var map %v Go %v", maps, data.VarNums)
	}
	// Force only the static SAT result in a temporary copy of cecSplit.c. Its
	// serial and pthread scheduling/cofactor/model code remains source text.
	sourceBytes, err := os.ReadFile(filepath.Join(src, "proof/cec/cecSplit.c"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(sourceBytes)
	marker := "static inline int Cnf_GiaSolveOne("
	start := strings.Index(source, marker)
	if start < 0 {
		t.Fatal("source solver function not found")
	}
	brace := strings.Index(source[start:], "{") + start
	depth, end := 0, -1
	for i := brace; i < len(source); i++ {
		if source[i] == '{' {
			depth++
		}
		if source[i] == '}' {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}
	if end < 0 {
		t.Fatal("source solver body unterminated")
	}
	const fake = `{
  int i,lit,zero; *pnVars=0; *pnConfs=0;
  if(!p->vCofVars || Vec_IntSize(p->vCofVars)==0) return -1;
  lit=Vec_IntEntry(p->vCofVars,Vec_IntSize(p->vCofVars)-1);
  zero=!Gia_ObjFaninC0(Gia_ManCo(p,0));
  printf("branch:%d,lit:%d\n",!zero,lit);fflush(stdout);
  if(zero) return 1;
  p->pCexComb=Abc_CexAlloc(0,Gia_ManPiNum(p),1);
  for(i=0;i<Vec_IntSize(p->vCofVars);i++){int v=Vec_IntEntry(p->vCofVars,i);if(!Abc_LitIsCompl(v)) Abc_InfoSetBit(p->pCexComb->pData,Abc_Lit2Var(v));}
  return 0;
 }`
	source = source[:brace] + fake + source[end:]
	splitFile := filepath.Join(dir, "split_instrumented.c")
	if err = os.WriteFile(splitFile, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	const splitMain = `#include "aig/gia/gia.h"
#include "proof/cec/cec.h"
#include "aig/gia/giaAig.h"
#include "sat/cnf/cnf.h"
#include "sat/bsat/satSolver.h"
#include <stdio.h>
extern Abc_Cex_t * Cec_SplitDeriveModel(Gia_Man_t *, Cnf_Dat_t *, sat_solver *);
extern int Cec_GiaSplitTest2(Gia_Man_t*,int,int,int,int,int,int,int);
extern int Cec_GiaSplitTestInt(Gia_Man_t*,int,int,int,int,int,int,int);
static Gia_Man_t *make(void){Gia_Man_t *g=Gia_ManStart(8);int a=Gia_ManAppendCi(g);Gia_ManAppendCo(g,a);return g;}
int main(void){Gia_Man_t *g=make();int s=Cec_GiaSplitTest2(g,1,0,0,1,0,0,1);printf("serial:%d,%u\n",s,g->pCexComb->pData[0]&1);Gia_ManStop(g);
 g=make();s=Cec_GiaSplitTestInt(g,3,0,0,1,0,0,1);printf("parallel:%d,%u\n",s,g->pCexComb->pData[0]&1);Gia_ManStop(g);
 g=make();s=Cec_GiaSplitTestInt(g,2,0,0,1,0,0,1);printf("oneworker:%d,%u\n",s,g->pCexComb->pData[0]&1);Gia_ManStop(g);
 {Aig_Man_t *aig;Cnf_Dat_t *cnf;sat_solver *sat;Abc_Cex_t *cex;int x,r;
 g=Gia_ManStart(8);x=Gia_ManAppendCi(g);r=Gia_ManAppendCi(g);Gia_ManAppendCo(g,x);Gia_ManAppendCo(g,r);Gia_ManSetRegNum(g,1);
 aig=Gia_ManToAigSimple(g);aig->nRegs=0;cnf=Cnf_Derive(aig,0);sat=(sat_solver*)Cnf_DataWriteIntoSolver(cnf,1,0);
 sat_solver_solve(sat,NULL,NULL,0,0,0,0);cex=Cec_SplitDeriveModel(g,cnf,sat);printf("layout:%d,%d\n",cex->nRegs,cex->nPis);
 Abc_CexFree(cex);sat_solver_delete(sat);Cnf_DataFree(cnf);Aig_ManStop(aig);Gia_ManStop(g);}
 return 0;}`
	splitMainFile := filepath.Join(dir, "split_main.c")
	if err = os.WriteFile(splitMainFile, []byte(splitMain), 0600); err != nil {
		t.Fatal(err)
	}
	splitExe := filepath.Join(dir, "split_oracle")
	args = []string{"-DABC_USE_STDINT_H", "-DABC_USE_PTHREADS", "-I", src, "-I", filepath.Join(src, "proof/cec"), splitMainFile, splitFile, lib, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", splitExe}
	if out, err = exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("link split oracle: %v %s", err, out)
	}
	out, err = exec.Command(splitExe).CombinedOutput()
	if err != nil {
		t.Fatalf("run split oracle: %v %s", err, out)
	}
	if !strings.Contains(string(out), "layout:0,1") || !strings.Contains(string(out), "branch:0,lit:1\nbranch:1,lit:0\nserial:1,1") || !strings.Contains(string(out), "branch:1,lit:1") || !strings.Contains(string(out), "parallel:0,0") || !strings.Contains(string(out), "oneworker:0,0") {
		t.Fatalf("unexpected source scheduling/cofactor witness: %s", out)
	}

	for _, workers := range []int32{1, 2} {
		g := gia.Gia_ManStart(8)
		a := gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, a)
		var mu sync.Mutex
		var branches []int32
		solve := func(part splitPart, _ int32) splitResult {
			if len(part.fixes) == 0 {
				return splitResult{status: -1}
			}
			value := part.graph.OutputLiterals()[0]
			mu.Lock()
			branches = append(branches, value)
			mu.Unlock()
			if value == 0 {
				return splitResult{status: 1}
			}
			return splitResult{status: 0, model: []bool{part.fixes[len(part.fixes)-1].value}}
		}
		status, model := splitRun(g, workers, 0, 0, 1, 0, 0, 1, solve)
		if workers == 1 {
			if status != 1 || !reflect.DeepEqual(model, []bool{true}) || !reflect.DeepEqual(branches, []int32{0, 1}) {
				t.Fatalf("Go serial status=%d model=%v branches=%v", status, model, branches)
			}
		} else if status != 0 || !reflect.DeepEqual(model, []bool{false}) {
			t.Fatalf("Go parallel status=%d model=%v branches=%v", status, model, branches)
		}
	}

	g2 := gia.Gia_ManStart(8)
	a2 := gia.Gia_ManAppendCi(g2)
	gia.Gia_ManAppendCo(g2, a2)
	fakeSolve := func(part splitPart, _ int32) splitResult {
		if len(part.fixes) == 0 {
			return splitResult{status: -1}
		}
		if part.graph.OutputLiterals()[0] == 0 {
			return splitResult{status: 1}
		}
		return splitResult{status: 0, model: []bool{part.fixes[0].value}}
	}
	status, model := splitRunMode(g2, 1, 0, 0, 1, 0, 0, 1, true, fakeSolve)
	if status != 0 || !reflect.DeepEqual(model, []bool{false}) {
		t.Fatalf("Go one-worker status %d model %v", status, model)
	}

}
