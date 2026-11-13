package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/cec/core"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func sourcePairFixture(t *testing.T) (*gia.Gia_Man_t, []int32) {
	t.Helper()
	g := gia.Gia_ManStart(100)
	inputs := make([]int32, 13)
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(g)
	}
	x := inputs[0]
	for _, lit := range inputs[1:] {
		x = gia.Gia_ManAppendAnd(g, x, lit)
	}
	y := gia.Gia_ManAppendAnd(g, x, inputs[0])
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, y)
	p := Cec_ParFra_t{nWords: 1, nRounds: 1, nItersMax: 10, nBTLimit: 1000, nDepthMax: 0, fUseOrigIds: 1}
	if out := Cec_ManSatSweeping(g, &p, 1); out == nil || p.Error != nil {
		t.Fatalf("source sweep: %v", p.Error)
	}
	return g, append([]int32(nil), vec.Vec_IntArray(gia.Gia_ManOrigEquivPairs(g))...)
}

func TestSourceSweepPublishesSpeculativePairs(t *testing.T) {
	g, pairs := sourcePairFixture(t)
	found := false
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i] == 22 && pairs[i+1] == 26 {
			found = true
		}
	}
	if !found {
		t.Fatalf("source pair (22,26) absent: %v", pairs)
	}
	model, err := core.NewModel(g)
	if err != nil {
		t.Fatal(err)
	}
	zero, err := model.Values(make([]bool, 13))
	if err != nil {
		t.Fatal(err)
	}
	input := make([]bool, 13)
	for i := range input {
		input[i] = (1023>>i)&1 != 0
	}
	values, err := model.Values(input)
	if err != nil {
		t.Fatal(err)
	}
	if (zero[22] != zero[26]) == (values[22] != values[26]) {
		t.Fatal("fixture no longer demonstrates the source's false phase-normalized pair")
	}
}

// TestSourceSweepPairsAgainstABC links the checked-out ABC object files when
// available. It records source behavior, including its inherited false pair.
func TestSourceSweepPairsAgainstABC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("optional ABC object oracle requires Linux")
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
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	src := filepath.Join(root, "src")
	if _, err = os.Stat(filepath.Join(src, "proof/cec/cecCore.o")); err != nil {
		t.Skip("prebuilt ABC object files unavailable")
	}
	var objects []string
	err = filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "misc/st/st.o") && path != filepath.Join(src, "proof/cec/cecSimBack.o") {
			objects = append(objects, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	// Some checkouts contain a stale st.o missing symbols in the current st.c.
	// Compile this one source into the temporary oracle archive.
	stObject := filepath.Join(dir, "abc_st.o")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", stObject).CombinedOutput(); err != nil {
		t.Fatalf("compile ABC st: %v %s", err, out)
	}
	objects = append(objects, stObject)
	simObject := filepath.Join(dir, "cecSimBack.o")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "proof/cec/cecSimBack.c"), "-o", simObject).CombinedOutput(); err != nil {
		t.Fatalf("compile ABC backward simulation: %v %s", err, out)
	}
	objects = append(objects, simObject)
	archive := filepath.Join(dir, "libabc.a")
	args := append([]string{"qcs", archive}, objects...)
	if out, err := exec.Command(ar, args...).CombinedOutput(); err != nil {
		t.Fatalf("archive ABC objects: %v %s", err, out)
	}
	cfile := filepath.Join(dir, "oracle.c")
	binary := filepath.Join(dir, "oracle")
	const program = `#include "aig/gia/gia.h"
#include "proof/cec/cec.h"
#include "proof/cec/cecInt.h"
#include <stdio.h>
extern int Cec_ManCheckSat_rec(Gia_Man_t *, Gia_Obj_t *, int);
extern word Cec_ManCheckSat2(Gia_Man_t *, Gia_Obj_t *, int, Vec_Wrd_t *);
extern word Cec_ManCheckSat2_rec(Gia_Man_t *, Gia_Obj_t *, word, Vec_Wrd_t *);
int main(void) {
 int in[13],i,x,y; Gia_Man_t *g=Gia_ManStart(100),*q; Cec_ParFra_t p;
 Cec_ManFraSetDefaultParams(&p); p.nWords=1;p.nRounds=1;p.nItersMax=10;p.nBTLimit=1000;p.nDepthMax=0;p.fUseOrigIds=1;
 for(i=0;i<13;i++)in[i]=Gia_ManAppendCi(g); x=in[0];for(i=1;i<13;i++)x=Gia_ManAppendAnd(g,x,in[i]);
 y=Gia_ManAppendAnd(g,x,in[0]);Gia_ManAppendCo(g,x);Gia_ManAppendCo(g,y);
 q=Cec_ManSatSweeping(g,&p,1);if(!q)return 2;printf("pairs:");for(i=0;i<Vec_IntSize(g->vIdsEquiv);i++)printf(" %d",Vec_IntEntry(g->vIdsEquiv,i));puts("");Gia_ManStop(q);Gia_ManStop(g);
 { Cec_ParCec_t c; int a,b; g=Gia_ManStart(16); x=Gia_ManAppendCi(g);
 Gia_ManAppendCo(g,0);Gia_ManAppendCo(g,0);Gia_ManAppendCo(g,x);Gia_ManAppendCo(g,x);
 a=Vec_IntEntry(g->vCos,1);b=Vec_IntEntry(g->vCos,2);Vec_IntWriteEntry(g->vCos,1,b);Vec_IntWriteEntry(g->vCos,2,a);
 Cec_ManCecSetDefaultParams(&c); printf("reordered:%d\n",Cec_ManVerifyNaive(g,&c)); Gia_ManStop(g); }
 { Gia_Man_t *l=Gia_ManStart(8),*r=Gia_ManStart(8); int la=Gia_ManAppendCi(l),ra=Gia_ManAppendCi(r), result;
 Gia_ManAppendCo(l,la);Gia_ManAppendCo(r,ra);l->pCexComb=Abc_CexAlloc(0,1,1);
 result=Cec_ManVerifyTwo(l,r,0); printf("verifytwo:%d,%d\n",result,l->pCexComb != NULL);
 Gia_ManStop(l);Gia_ManStop(r); }
 { Gia_Man_t *l=Gia_ManStart(16),*r=Gia_ManStart(16); int la=Gia_ManAppendCi(l),lr=Gia_ManAppendCi(l),ra=Gia_ManAppendCi(r),rr=Gia_ManAppendCi(r),result;
 Gia_ManAppendCo(l,Gia_ManAppendAnd(l,la,lr));Gia_ManAppendCo(l,la);Gia_ManSetRegNum(l,1);
 Gia_ManAppendCo(r,Gia_ManAppendAnd(r,ra,rr));Gia_ManAppendCo(r,ra^1);Gia_ManSetRegNum(r,1);
 result=Cec_ManVerifyTwo(l,r,0); printf("verifytworeg:%d,%d,%d,%d\n", result,l->pCexComb->nRegs,l->pCexComb->nPis,l->pCexComb->iPo);
 Gia_ManStop(l);Gia_ManStop(r); }
 { Gia_Man_t *g=Gia_ManStart(8); Cec_ParCec_t c; int a=Gia_ManAppendCi(g), b=Gia_ManAppendBuf(g,a);
 Gia_ManAppendCo(g,a);Gia_ManAppendCo(g,b);Cec_ManCecSetDefaultParams(&c);c.fSilent=1;
 printf("specialbuffer:%d\n",Cec_ManHandleSpecialCases(g,&c)); Gia_ManStop(g); }
 { Gia_Man_t *g=Gia_ManStart(8); int x=Gia_ManAppendCi(g), stub, fresh, conflict;
 Gia_ManIncrementTravId(g); stub=(int)Cec_ManCheckSat2(g,Gia_ManObj(g,x>>1),1,NULL);
 fresh=Cec_ManCheckSat_rec(g,Gia_ManConst0(g),1); conflict=Cec_ManCheckSat_rec(g,Gia_ManConst0(g),0);
 printf("backward:%d,%d,%d\n",stub,fresh,conflict);
 Gia_ManIncrementTravId(g); fresh=(int)Cec_ManCheckSat2_rec(g,Gia_ManObj(g,x>>1),2,NULL); conflict=(int)Cec_ManCheckSat2_rec(g,Gia_ManObj(g,x>>1),2,NULL);
 printf("backword:%d,%d\n",fresh,conflict); Gia_ManStop(g); }
 { Gia_Man_t *g=Gia_ManStart(16); Cec_ParSat_t opts; Cec_ManSat_t *sat; Vec_Int_t *ptr; int a=Gia_ManAppendCi(g),b=Gia_ManAppendCi(g),c=Gia_ManAppendCi(g),x=Gia_ManAppendAnd(g,a,b),y=Gia_ManAppendAnd(g,x,c),status;
 Gia_ManAppendCo(g,y);Gia_ManSetPhase(g);Gia_ManLevelNum(g);Gia_ManIncrementTravId(g);
 Cec_ManSatSetDefaultParams(&opts); sat=Cec_ManSatCreate(g,&opts);status=Cec_ManSatCheckNode(sat,Gia_ManObj(g,y>>1));ptr=sat->vCex;
 Cec_ManSavePattern(sat,Gia_ManObj(g,x>>1),Gia_ManObj(g,y>>1));
 printf("pattern:%d,%d,%d,%d,%d,%d,%d\n",status,ptr==sat->vCex,sat->nCexLits,Vec_IntEntry(ptr,0),Vec_IntEntry(ptr,1),Vec_IntEntry(ptr,2),Gia_ObjIsTravIdCurrent(g,Gia_ManObj(g,c>>1)));
 Cec_ManSavePattern(sat,Gia_ManObj(g,x>>1),NULL);printf("pattern2:%d,%d,%d,%d,%d\n",ptr==sat->vCex,sat->nCexLits,Vec_IntEntry(ptr,0),Vec_IntEntry(ptr,1),Gia_ObjIsTravIdCurrent(g,Gia_ManObj(g,c>>1)));
 Cec_ManSatStop(sat);Gia_ManStop(g); }
 return 0;
}`
	if err = os.WriteFile(cfile, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	args = []string{"-DABC_USE_STDINT_H", "-I", src, cfile, archive, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", binary}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("link ABC oracle: %v %s", err, out)
	}
	out, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run ABC oracle: %v %s", err, out)
	}
	if !strings.Contains(string(out), "reordered:1") {
		t.Fatalf("ABC reordered CO result: %s", out)
	}
	if !strings.Contains(string(out), "pattern:0,1,3,0,2,4,1") || !strings.Contains(string(out), "pattern2:1,5,0,2,0") || !strings.Contains(string(out), "backword:1,0") || !strings.Contains(string(out), "backward:0,1,0") || !strings.Contains(string(out), "specialbuffer:-1") || !strings.Contains(string(out), "verifytwo:1,0") || !strings.Contains(string(out), "verifytworeg:0,0,2,1") {
		t.Fatalf("ABC verify-two source behavior: %s", out)
	}
	var original []int32
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "pairs:") {
			continue
		}
		for _, field := range strings.Fields(strings.TrimPrefix(line, "pairs:")) {
			var id int32
			if _, err := fmt.Sscan(field, &id); err != nil {
				t.Fatal(err)
			}
			original = append(original, id)
		}
	}
	_, got := sourcePairFixture(t)
	if !reflect.DeepEqual(got, original) {
		t.Fatalf("Go pairs %v, ABC pairs %v (output %s)", got, original, out)
	}
}
