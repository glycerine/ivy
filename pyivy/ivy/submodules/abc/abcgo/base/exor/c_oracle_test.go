package exor

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestOriginalCExorLinkAndCostOracle(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section collection required")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional original C oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	dir := t.TempDir()
	main := filepath.Join(dir, "main.c")
	binary := filepath.Join(dir, "oracle")
	code := `#include "base/exor/exor.h"
extern cinfo g_CoverInfo;extern int GetDistance(Cube*,Cube*);extern int GetQCost(int,int);extern int ExorLinkCubeIteratorStart(Cube**,Cube*,Cube*,cubedist);extern int ExorLinkCubeIteratorNext(Cube**);extern void ExorLinkCubeIteratorCleanUp(int);
int main(void){int n;unsigned ai,bi,ao,bo;PrepareBitSetModule();while(scanf("%d%x%x%x%x",&n,&ai,&bi,&ao,&bo)==5){memset(&g_CoverInfo,0,sizeof(g_CoverInfo));g_CoverInfo.nVarsIn=n;g_CoverInfo.nVarsOut=2;g_CoverInfo.nWordsIn=g_CoverInfo.nWordsOut=1;g_CoverInfo.nCubesAlloc=128;g_CoverInfo.cIDs=1;AllocateCover(128,1,1);Cube*a=GetFreeCube(),*b=GetFreeCube();a->pCubeDataIn[0]=ai;b->pCubeDataIn[0]=bi;a->pCubeDataOut[0]=ao;b->pCubeDataOut[0]=bo;int la=0,lb=0,na=0;for(int i=0;i<n;i++){int v=(ai>>(2*i))&3;la+=v!=3;na+=v==1;lb+=((bi>>(2*i))&3)!=3;}a->a=la;b->a=lb;a->z=__builtin_popcount(ao);b->z=__builtin_popcount(bo);a->q=ComputeQCostBits(a);b->q=ComputeQCostBits(b);int d=GetDistance(a,b);printf("%d %d",d,GetQCost(la,na));if(d>=2&&d<=4){Cube*group[4];if(ExorLinkCubeIteratorStart(group,a,b,(cubedist)(d-2))){do{printf(" |");for(int i=0;i<d;i++)printf(" %x:%x",group[i]->pCubeDataIn[0],group[i]->pCubeDataOut[0]);}while(ExorLinkCubeIteratorNext(group));ExorLinkCubeIteratorCleanUp(0);}}puts("");DelocateCover();}return 0;}`
	if err = os.WriteFile(main, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	args := []string{"-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", src, main}
	for _, f := range []string{"exor.c", "exorBits.c", "exorCubes.c", "exorLink.c"} {
		args = append(args, filepath.Join(src, "base/exor", f))
	}
	args = append(args, "-o", binary)
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("build C oracle: %v\n%s", err, out)
	}
	rng := rand.New(rand.NewSource(389))
	var input bytes.Buffer
	type expected struct {
		distance, cost int
		groups         []string
	}
	var expectedResults []expected
	for trial := 0; trial < 300; trial++ {
		n := 2 + rng.Intn(6)
		valuesA, valuesB := make([]VarValue, n), make([]VarValue, n)
		for i := range valuesA {
			valuesA[i] = VarValue(1 + rng.Intn(3))
			valuesB[i] = VarValue(1 + rng.Intn(3))
		}
		oa, ob := 1+rng.Intn(3), 1+rng.Intn(3)
		a, b := NewCube(valuesA, []bool{oa&1 != 0, oa&2 != 0}), NewCube(valuesB, []bool{ob&1 != 0, ob&2 != 0})
		fmt.Fprintf(&input, "%d %x %x %x %x\n", n, a.Inputs[0], b.Inputs[0], oa, ob)
		e := expected{distance: GetDistance(a, b), cost: ComputeQCostBits(a)}
		if e.distance >= 2 && e.distance <= 4 {
			for _, group := range ExorLink(a, b, 4) {
				var terms []string
				for _, c := range group {
					terms = append(terms, fmt.Sprintf("%x:%x", c.Inputs[0], c.Outputs[0]))
				}
				sort.Strings(terms)
				e.groups = append(e.groups, strings.Join(terms, " "))
			}
			sort.Strings(e.groups)
		}
		expectedResults = append(expectedResults, e)
	}
	cmd := exec.Command(binary)
	cmd.Stdin = &input
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run oracle: %v\n%s", err, out)
	}
	scan := bufio.NewScanner(bytes.NewReader(out))
	index := 0
	for scan.Scan() {
		if index >= len(expectedResults) {
			t.Fatal("extra C output")
		}
		parts := strings.Split(scan.Text(), "|")
		header := strings.Fields(parts[0])
		d, _ := strconv.Atoi(header[0])
		cost, _ := strconv.Atoi(header[1])
		want := expectedResults[index]
		if d != want.distance || cost != want.cost {
			t.Fatalf("case %d distance/cost %d/%d want %d/%d", index, d, cost, want.distance, want.cost)
		}
		var groups []string
		for _, g := range parts[1:] {
			terms := strings.Fields(g)
			sort.Strings(terms)
			groups = append(groups, strings.Join(terms, " "))
		}
		sort.Strings(groups)
		if strings.Join(groups, "|") != strings.Join(want.groups, "|") {
			t.Fatalf("case %d group mismatch\nC %v\nGo %v", index, groups, want.groups)
		}
		index++
	}
	if scan.Err() != nil || index != len(expectedResults) {
		t.Fatal("truncated oracle output", index, scan.Err())
	}
}
