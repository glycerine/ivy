package dau

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestNonDsdOriginalC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section removal required")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional C oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	dir := t.TempDir()
	main, bin := filepath.Join(dir, "main.c"), filepath.Join(dir, "oracle")
	code := `#include "opt/dau/dauNonDsd.c"
int main(void){char mode;word p[1024],c[2048],d[2048];unsigned long long x;while(scanf(" %c",&mode)==1){if(mode=='R'){char s[100];scanf("%99s",s);printf("%x\n",Dau_DecReadSet(s));continue;}int n;scanf("%d",&n);int nf=0,shared=0,mask=0;unsigned set=0;if(mode=='T')scanf("%d%d%d",&nf,&shared,&mask);if(mode=='D')scanf("%x",&set);for(int i=0;i<Abc_TtWordNum(n);i++){if(scanf("%llx",&x)!=1)return 2;p[i]=x;}if(mode=='F'){Vec_Int_t *v=Dau_DecFindSets(p,n);printf("%x",Vec_IntSize(v));for(int i=0;i<Vec_IntSize(v);i++)printf(" %x",(unsigned)Vec_IntEntry(v,i));puts("");Vec_IntFree(v);}else if(mode=='T'){memset(c,0,sizeof(c));memset(d,0,sizeof(d));int *sched=Extra_GreyCodeSchedule(n-nf);int status=Dau_DecCheckSetTop(p,n,nf,n-nf,shared,mask,sched,d,c);int nc=Abc_TtWordNum(nf+shared+1),nd=Abc_TtWordNum(n-nf);printf("%x %x %x",status,nc,nd);for(int i=0;i<nc;i++)printf(" %llx",(unsigned long long)c[i]);for(int i=0;i<nd;i++)printf(" %llx",(unsigned long long)d[i]);puts("");free(sched);}else if(mode=='D'){word tc=0,td=0;int pc[16]={0},pd[16]={0},nc=-1,nd=-1,ns=-1;int status=Dau_DecDecomposeSet(p,n,set,&tc,&td,pc,pd,&nc,&nd,&ns);printf("%x %x %x %x %llx %llx",status,(unsigned)nc,(unsigned)nd,(unsigned)ns,(unsigned long long)tc,(unsigned long long)td);for(int i=0;i<16;i++)printf(" %x %x",(unsigned)pc[i],(unsigned)pd[i]);puts("");}else return 3;}return 0;}`
	if err := os.WriteFile(main, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", src, main, filepath.Join(src, "misc/extra/extraUtilMisc.c"), "-o", bin)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(b))
	}
	var input bytes.Buffer
	var expected [][]uint64
	addTruth := func(p []uint64) {
		for _, w := range p {
			fmt.Fprintf(&input, " %x", w)
		}
		fmt.Fprintln(&input)
	}
	addT := func(p []uint64, n, f, s, mask int32) {
		fmt.Fprint(&input, "T ", n, " ", f, " ", s, " ", mask)
		addTruth(p)
		c, d := make([]uint64, dauCountWords(f+s+1)), make([]uint64, dauCountWords(n-f))
		status := Dau_DecCheckSetTop(p, n, f, n-f, s, mask, dauDecSchedule(n-f), d, c)
		want := []uint64{uint64(status), uint64(len(c)), uint64(len(d))}
		want = append(want, c...)
		want = append(want, d...)
		expected = append(expected, want)
	}
	addF := func(p []uint64, n int32) {
		for v := int32(0); v < n; v++ {
			if !dauHasVariable(p, int(v)) {
				return
			}
		}
		fmt.Fprint(&input, "F ", n)
		addTruth(p)
		sets := Dau_DecFindSets(p, n)
		want := []uint64{uint64(vec.Vec_IntSize(sets))}
		for _, s := range vec.Vec_IntArray(sets) {
			want = append(want, uint64(uint32(s)))
		}
		expected = append(expected, want)
	}
	for f := uint64(0); f < 256; f++ {
		p := []uint64{f * 0x0101010101010101}
		addT(p, 3, 1, 0, 0)
		addF(p, 3)
	}
	rng := rand.New(rand.NewSource(671))
	for n := int32(4); n <= 16; n++ {
		for trial := 0; trial < 10; trial++ {
			f := int32(1 + rng.Intn(int(n-2)))
			bound := n - f
			s := int32(rng.Intn(int(min(3, bound-2)) + 1))
			mask := int32(0)
			for bits.OnesCount32(uint32(mask)) < int(s) {
				mask |= 1 << uint(rng.Intn(int(bound)))
			}
			p := make([]uint64, dauCountWords(n))
			for i := range p {
				p[i] = rng.Uint64()
			}
			Abc_TtNormalizeSmallTruth(p, n)
			if trial&1 != 0 {
				c, d := make([]uint64, dauCountWords(f+s+1)), make([]uint64, dauCountWords(bound))
				for i := range c {
					c[i] = rng.Uint64()
				}
				for i := range d {
					d[i] = rng.Uint64()
				}
				p = dauDecCompose(c, d, n, f, s, mask)
			}
			addT(p, n, f, s, mask)
			if n <= 7 {
				addF(p, n)
			}
		}
	}
	for trial := 0; trial < 220; trial++ {
		n := int32(3 + trial%4)
		p := []uint64{rng.Uint64()}
		Abc_TtNormalizeSmallTruth(p, n)
		set := uint32(0)
		for v := int32(0); v < n; v++ {
			kind := rng.Intn(3)
			if kind == 2 {
				kind = 3
			}
			set |= uint32(kind) << uint(2*v)
		}
		var u, s, f int32
		Dau_DecSortSet(set, n, &u, &s, &f)
		if u < 2 || f < 1 {
			continue
		}
		fmt.Fprintf(&input, "D %d %x", n, set)
		addTruth(p)
		c, d := []uint64{0}, []uint64{0}
		pc, pd := make([]int32, 16), make([]int32, 16)
		nc, nd, ns := int32(-1), int32(-1), int32(-1)
		status := Dau_DecDecomposeSet(p, n, set, c, d, pc, pd, &nc, &nd, &ns)
		want := []uint64{uint64(status), uint64(uint32(nc)), uint64(uint32(nd)), uint64(uint32(ns)), c[0], d[0]}
		for i := range pc {
			want = append(want, uint64(uint32(pc[i])), uint64(uint32(pd[i])))
		}
		expected = append(expected, want)
	}
	// Uppercase branches in the C source shift a negative value; those
	// undefined results are excluded from the oracle and checked against
	// the intended two-bit shared encoding in the standalone Go regression.
	for _, s := range []string{"abc", "bcd", "a", "p", "ap"} {
		fmt.Fprintln(&input, "R", s)
		expected = append(expected, []uint64{uint64(Dau_DecReadSet(s))})
	}
	cmd = exec.Command(bin)
	cmd.Stdin = &input
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(b))
	}
	scan := bufio.NewScanner(bytes.NewReader(b))
	scan.Buffer(make([]byte, 4096), 1<<20)
	for sample, want := range expected {
		if !scan.Scan() {
			t.Fatal("missing source result", sample)
		}
		f := strings.Fields(scan.Text())
		if len(f) != len(want) {
			t.Fatalf("sample%d source diagnostic: %s", sample, scan.Text())
		}
		for i, s := range f {
			v, err := strconv.ParseUint(s, 16, 64)
			if err != nil || v != want[i] {
				t.Fatalf("sample%d field%d Go%x C%s", sample, i, want[i], s)
			}
		}
	}
}
