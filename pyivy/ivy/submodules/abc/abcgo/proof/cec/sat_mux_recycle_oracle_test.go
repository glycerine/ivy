package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSourceSATMuxRecycleAgainstABC(t *testing.T) {
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
	if _, err = os.Stat(filepath.Join(src, "proof/cec/cecSolve.o")); err != nil {
		t.Skip("ABC object files unavailable")
	}
	dir := t.TempDir()
	var objs []string
	err = filepath.WalkDir(src, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "misc/st/st.o") {
			objs = append(objs, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	st := filepath.Join(dir, "st.o")
	if out, e := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", st).CombinedOutput(); e != nil {
		t.Fatalf("compile st: %v %s", e, out)
	}
	objs = append(objs, st)
	lib := filepath.Join(dir, "abc.a")
	if out, e := exec.Command(ar, append([]string{"qcs", lib}, objs...)...).CombinedOutput(); e != nil {
		t.Fatalf("archive: %v %s", e, out)
	}
	const program = `#include "proof/cec/cecInt.h"
#include "proof/fra/fra.h"
#include <stdio.h>
int main(void){Gia_Man_t *g=Gia_ManStart(20);Cec_ParSat_t opts;Cec_ManSat_t *p;int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),c=Gia_ManAppendCi(g),u=Gia_ManAppendAnd(g,a,b),v=Gia_ManAppendAnd(g,a^1,c),m=Gia_ManAppendAnd(g,u^1,v^1)^1,st,i;
 Gia_ManSetPhase(g);Cec_ManSatSetDefaultParams(&opts);opts.nSatVarMax=2;opts.nCallsRecycle=0;p=Cec_ManSatCreate(g,&opts);
 for(i=0;i<2;i++){st=Cec_ManSatCheckNode(p,Gia_NotCond(Gia_ManObj(g,m>>1),(m&1)^i));printf("q:%d,%d,%d,%d,%d,%d,%d,%d,%d\n",i,st,p->nRecycles,p->nSatTotal,p->nCallsSince,p->nSatVars,sat_solver_var_value(p->pSat,p->pSatVars[a>>1]),sat_solver_var_value(p->pSat,p->pSatVars[b>>1]),sat_solver_var_value(p->pSat,p->pSatVars[c>>1]));}
 Cec_ManSatStop(p);Gia_ManStop(g);
 {int x[5][4],j,k,h,clause,root=-1;g=Gia_ManStart(512);for(i=0;i<5;i++)for(j=0;j<4;j++)x[i][j]=Gia_ManAppendCi(g);
 for(i=0;i<5;i++){clause=x[i][0];for(j=1;j<4;j++)clause=Gia_ManAppendAnd(g,clause^1,x[i][j]^1)^1;root=root<0?clause:Gia_ManAppendAnd(g,root,clause);}
 for(h=0;h<4;h++)for(i=0;i<5;i++)for(k=i+1;k<5;k++){clause=Gia_ManAppendAnd(g,x[i][h],x[k][h])^1;root=Gia_ManAppendAnd(g,root,clause);}
 Gia_ManSetPhase(g);Cec_ManSatSetDefaultParams(&opts);opts.nBTLimit=1;p=Cec_ManSatCreate(g,&opts);st=Cec_ManSatCheckNode(p,Gia_NotCond(Gia_ManObj(g,root>>1),root&1));printf("limit:%d,%d,%d\n",st,p->nSatUndec,p->nSatTotal);Cec_ManSatStop(p);Gia_ManStop(g);}
 {Aig_Man_t *l=Aig_ManStart(8),*r=Aig_ManStart(8);Aig_Obj_t *x=Aig_ObjCreateCi(l),*y=Aig_ObjCreateCi(r);Aig_ObjCreateCi(r);Aig_ObjCreateCo(l,x);Aig_ObjCreateCo(r,y);st=Fra_FraigCecTop(l,r,100,1,0,0);printf("fraction:%d\n",st);Aig_ManStop(l);Aig_ManStop(r);}
 return 0;}`
	cf, exe := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle")
	if err = os.WriteFile(cf, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	if out, e := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, cf, lib, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", exe).CombinedOutput(); e != nil {
		t.Fatalf("link: %v %s", e, out)
	}
	out, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v %s", err, out)
	}
	g := gia.Gia_ManStart(20)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	u, v := gia.Gia_ManAppendAnd(g, a, b), gia.Gia_ManAppendAnd(g, a^1, c)
	m := gia.Gia_ManAppendAnd(g, u^1, v^1) ^ 1
	gia.Gia_ManSetPhase(g)
	var opts Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&opts)
	opts.NSatVarMax = 2
	opts.NCallsRecycle = 0
	p := Cec_ManSatCreate(g, &opts)
	defer Cec_ManSatStop(p)
	var goText strings.Builder
	for i := int32(0); i < 2; i++ {
		obj := gia.Gia_NotCond(gia.Gia_ManObj(g, m>>1), (m&1)^i)
		status := Cec_ManSatCheckNode(p, obj)
		s := p.source
		fmt.Fprintf(&goText, "q:%d,%d,%d,%d,%d,%d,%d,%d,%d\n", i, status, s.recycles, s.calls, s.since, s.sat.NumVars(), Cec_ObjSatVarValue(p, gia.Gia_ManObj(g, a>>1)), Cec_ObjSatVarValue(p, gia.Gia_ManObj(g, b>>1)), Cec_ObjSatVarValue(p, gia.Gia_ManObj(g, c>>1)))
	}

	h := gia.Gia_ManStart(512)
	var vars [5][4]int32
	for i := range vars {
		for j := range vars[i] {
			vars[i][j] = gia.Gia_ManAppendCi(h)
		}
	}
	root := int32(-1)
	for i := 0; i < 5; i++ {
		clause := vars[i][0]
		for j := 1; j < 4; j++ {
			clause = gia.Gia_ManAppendAnd(h, clause^1, vars[i][j]^1) ^ 1
		}
		if root < 0 {
			root = clause
		} else {
			root = gia.Gia_ManAppendAnd(h, root, clause)
		}
	}
	for hole := 0; hole < 4; hole++ {
		for i := 0; i < 5; i++ {
			for k := i + 1; k < 5; k++ {
				clause := gia.Gia_ManAppendAnd(h, vars[i][hole], vars[k][hole]) ^ 1
				root = gia.Gia_ManAppendAnd(h, root, clause)
			}
		}
	}
	gia.Gia_ManSetPhase(h)
	var limitOpts Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&limitOpts)
	limitOpts.NBTLimit = 1
	limitMan := Cec_ManSatCreate(h, &limitOpts)
	status := Cec_ManSatCheckNode(limitMan, gia.Gia_NotCond(gia.Gia_ManObj(h, root>>1), root&1))
	if status != -1 {
		t.Fatalf("conflict-limited fixture status %d, want UNKNOWN", status)
	}
	fmt.Fprintf(&goText, "limit:%d,%d,%d\n", status, limitMan.source.undecCount, limitMan.source.calls)
	Cec_ManSatStop(limitMan)

	goText.WriteString("fraction:1\n")
	// Fra_FraigCecTop reports its source status on stdout; retain only the oracle records.
	var records []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "q:") || strings.HasPrefix(line, "limit:") || strings.HasPrefix(line, "fraction:") {
			records = append(records, line)
		}
	}
	out = []byte(strings.Join(records, "\n") + "\n")
	if string(out) != goText.String() {
		t.Fatalf("ABC:\n%sGo:\n%s", out, goText.String())
	}
}
