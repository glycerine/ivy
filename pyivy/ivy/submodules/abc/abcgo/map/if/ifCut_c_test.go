package if_

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestCutNativeOriginalCOracle(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section collection required")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional source oracle needs cc")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	dir := t.TempDir()
	path := filepath.Join(dir, "main.c")
	binary := filepath.Join(dir, "oracle")
	code := `#include "map/if/ifCut.c"
#include "map/if/ifCount.h"
int main(void){char mode;while(scanf(" %c",&mode)==1){if(mode=='M'){int limit,na,nb;scanf("%d%d%d",&limit,&na,&nb);If_Cut_t*a=calloc(1,sizeof(If_Cut_t)+64),*b=calloc(1,sizeof(If_Cut_t)+64),*c=calloc(1,sizeof(If_Cut_t)+64);If_Man_t p={0};a->nLimit=b->nLimit=c->nLimit=limit;a->nLeaves=na;b->nLeaves=nb;for(int i=0;i<na;i++)scanf("%d",a->pLeaves+i);for(int i=0;i<nb;i++)scanf("%d",b->pLeaves+i);a->uSign=If_ObjCutSignCompute(a);b->uSign=If_ObjCutSignCompute(b);int ok=If_CutMergeOrdered_(&p,a,b,c);printf("%d",ok);if(ok){printf(" %d %u %u",(int)c->nLeaves,c->uSign,p.uSharedMask);for(int i=0;i<c->nLeaves;i++)printf(" %d",c->pLeaves[i]);for(int i=0;i<na;i++)printf(" %d",p.pPerm[0][i]);for(int i=0;i<nb;i++)printf(" %d",p.pPerm[1][i]);}puts("");free(a);free(b);free(c);}else if(mode=='L'){int n,x,xor,times[64],size=0,result=0;scanf("%d%d",&n,&xor);for(int i=0;i<n;i++){scanf("%d",&x);result=If_LogCounterAdd(times,&size,x,xor);}printf("%d %d",result,size);for(int i=0;i<size;i++)printf(" %d",times[i]);puts("");}else if(mode=='C'){int power,mode;If_Par_t pars={0};If_Man_t p={0};If_Cut_t a={0},b={0};scanf("%d%d",&power,&mode);pars.fPower=power;p.pPars=&pars;p.SortMode=mode;p.fEpsilon=0.001;int v[12];for(int i=0;i<12;i++)scanf("%d",v+i);a.Delay=v[0];a.Area=v[1];a.Edge=v[2];a.Power=v[3];a.nLeaves=v[4];a.fUseless=v[5];b.Delay=v[6];b.Area=v[7];b.Edge=v[8];b.Power=v[9];b.nLeaves=v[10];b.fUseless=v[11];printf("%d\n",If_ManSortCompare(&p,&a,&b));}else return 2;}return 0;}`
	if err = os.WriteFile(path, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", src, path, "-o", binary)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile original C: %v\n%s", err, out)
	}
	r := rand.New(rand.NewSource(726))
	var input bytes.Buffer
	var expected [][]int64
	for trial := 0; trial < 500; trial++ {
		limit := 1 + r.Intn(16)
		a, b := []int32{}, []int32{}
		for id := int32(0); id < 16; id++ {
			if len(a) < limit && r.Intn(4) == 0 {
				a = append(a, id)
			}
			if len(b) < limit && r.Intn(4) == 0 {
				b = append(b, id)
			}
		}
		if len(a)+len(b) == 0 {
			a = []int32{0}
		}
		fmt.Fprintf(&input, "M %d %d %d", limit, len(a), len(b))
		for _, leaf := range append(append([]int32(nil), a...), b...) {
			fmt.Fprintf(&input, " %d", leaf)
		}
		fmt.Fprintln(&input)
		p := &If_Man_t{}
		ca, cb, c := ifTestCut(uint32(limit), a...), ifTestCut(uint32(limit), b...), ifTestCut(uint32(limit))
		ok := If_CutMergeOrdered_(p, ca, cb, c)
		want := []int64{int64(ok)}
		if ok != 0 {
			want = append(want, int64(c.nLeaves), int64(c.uSign), int64(p.uSharedMask))
			for _, leaf := range ifLeaves(c) {
				want = append(want, int64(leaf))
			}
			for i := range a {
				want = append(want, int64(p.pPerm[0][i]))
			}
			for i := range b {
				want = append(want, int64(p.pPerm[1][i]))
			}
		}
		expected = append(expected, want)
		n, xor := 1+r.Intn(32), int32(r.Intn(2))
		fmt.Fprintf(&input, "L %d %d", n, xor)
		times := make([]int32, 64)
		size, result := int32(0), int32(0)
		for i := 0; i < n; i++ {
			arrival := int32(r.Intn(30))
			fmt.Fprintf(&input, " %d", arrival)
			result = If_LogCounterAdd(times, &size, arrival, xor)
		}
		fmt.Fprintln(&input)
		want = []int64{int64(result), int64(size)}
		for _, arrival := range times[:size] {
			want = append(want, int64(arrival))
		}
		expected = append(expected, want)
		power, mode := r.Intn(2), r.Intn(3)
		fields := make([]int, 12)
		for i := range fields {
			fields[i] = r.Intn(4)
		}
		fields[5] %= 2
		fields[11] %= 2
		fmt.Fprintf(&input, "C %d %d", power, mode)
		for _, v := range fields {
			fmt.Fprintf(&input, " %d", v)
		}
		fmt.Fprintln(&input)
		p = &If_Man_t{pPars: &If_Par_t{FPower: int32(power)}, SortMode: int32(mode), fEpsilon: 0.001}
		ca = &If_Cut_t{Delay: float32(fields[0]), Area: float32(fields[1]), Edge: float32(fields[2]), Power: float32(fields[3]), nLeaves: uint32(fields[4]), fUseless: uint32(fields[5])}
		cb = &If_Cut_t{Delay: float32(fields[6]), Area: float32(fields[7]), Edge: float32(fields[8]), Power: float32(fields[9]), nLeaves: uint32(fields[10]), fUseless: uint32(fields[11])}
		expected = append(expected, []int64{int64(If_ManSortCompare(p, ca, cb))})
	}
	cmd = exec.Command(binary)
	cmd.Stdin = &input
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run source oracle: %v\n%s", err, output)
	}
	scan := bufio.NewScanner(bytes.NewReader(output))
	line := 0
	for scan.Scan() {
		if line >= len(expected) {
			t.Fatal("extra source output")
		}
		actual := strings.Fields(scan.Text())
		want := expected[line]
		if len(actual) != len(want) {
			t.Fatalf("line %d length %d want %d", line, len(actual), len(want))
		}
		for i, text := range actual {
			v, err := strconv.ParseInt(text, 10, 64)
			if err != nil || v != want[i] {
				t.Fatalf("source differential line%d col%d got%s want%d", line, i, text, want[i])
			}
		}
		line++
	}
	if scan.Err() != nil || line != len(expected) {
		t.Fatal("truncated source oracle")
	}
}
