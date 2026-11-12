package dau

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

func TestHierarchicalOriginalC(t *testing.T) {
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
	code := `#include "opt/dau/dauCanon.c"
int main(void){Abc_TtHieMan_t *m[17][6][2]={0};int n,l,e;word t[1024];char p[16];unsigned long long x;while(scanf("%d%d%d",&n,&l,&e)==3){if(!m[n][l][e])m[n][l][e]=Abc_TtHieManStart(n,l);int w=Abc_TtWordNum(n);for(int i=0;i<w;i++){if(scanf("%llx",&x)!=1)return 2;t[i]=x;}for(int i=0;i<16;i++)p[i]=0;unsigned phase=Abc_TtCanonicizeHie(m[n][l][e],t,n,p,e);printf("%u",phase);for(int i=0;i<n;i++)printf(" %d",p[i]);for(int i=0;i<w;i++)printf(" %llx",(unsigned long long)t[i]);puts("");}for(int n=0;n<17;n++)for(int l=0;l<6;l++)for(int e=0;e<2;e++)if(m[n][l][e])Abc_TtHieManStop(m[n][l][e]);return 0;}`
	if err := os.WriteFile(main, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	args := []string{"-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", src, main, filepath.Join(src, "bool/lucky/luckySimple.c"), filepath.Join(src, "bool/lucky/luckySwap.c"), filepath.Join(src, "bool/lucky/lucky.c"), "-o", bin}
	cmd := exec.Command(cc, args...)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(b))
	}
	type sample struct {
		n, l, e int32
		truth   []uint64
	}
	var samples []sample
	for f := uint64(0); f < 256; f++ {
		for e := int32(0); e < 2; e++ {
			samples = append(samples, sample{3, 5, e, []uint64{f * 0x0101010101010101}})
		}
	}
	rng := rand.New(rand.NewSource(17))
	for n := int32(0); n <= 10; n++ {
		for l := int32(1); l <= 5; l++ {
			for e := int32(0); e < 2; e++ {
				if e == 1 && n > 6 {
					continue
				}
				for trial := 0; trial < 8; trial++ {
					truth := make([]uint64, dauCountWords(n))
					for i := range truth {
						truth[i] = rng.Uint64()
					}
					Abc_TtNormalizeSmallTruth(truth, n)
					samples = append(samples, sample{n, l, e, truth})
					samples = append(samples, sample{n, l, e, append([]uint64(nil), truth...)})
				}
			}
		}
	}
	var input bytes.Buffer
	for _, s := range samples {
		fmt.Fprint(&input, s.n, " ", s.l, " ", s.e)
		for _, w := range s.truth {
			fmt.Fprintf(&input, " %x", w)
		}
		fmt.Fprintln(&input)
	}
	cmd = exec.Command(bin)
	cmd.Stdin = &input
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(b))
	}
	scan := bufio.NewScanner(bytes.NewReader(b))
	scan.Buffer(make([]byte, 4096), 1<<20)
	managers := map[[3]int32]*Abc_TtHieMan_t{}
	for index, s := range samples {
		if !scan.Scan() {
			t.Fatal("missing source sample", index)
		}
		f := strings.Fields(scan.Text())
		if len(f) != 1+int(s.n)+len(s.truth) {
			t.Fatal("source diagnostic", scan.Text())
		}
		key := [3]int32{s.n, s.l, s.e}
		p := managers[key]
		if p == nil {
			p = Abc_TtHieManStart(s.n, s.l)
			managers[key] = p
		}
		truth := append([]uint64(nil), s.truth...)
		perm := make([]byte, s.n)
		phase := Abc_TtCanonicizeHie(p, truth, s.n, perm, s.e)
		v, _ := strconv.ParseUint(f[0], 10, 32)
		if uint32(v) != phase {
			t.Fatal(index, "phase")
		}
		for i, x := range perm {
			v, _ := strconv.ParseUint(f[1+i], 10, 8)
			if byte(v) != x {
				t.Fatalf("sample%d n%d l%d e%d perm%d Go%d C%d", index, s.n, s.l, s.e, i, x, v)
			}
		}
		for i, x := range truth {
			v, _ := strconv.ParseUint(f[1+int(s.n)+i], 16, 64)
			if v != x {
				t.Fatalf("sample%d n%d l%d e%d word%d Go%x C%x", index, s.n, s.l, s.e, i, x, v)
			}
		}
	}
}

func TestBalancedWrapPreservesComplementPhase(t *testing.T) {
	truth := []uint64{0xaaaaaaaaaaaaaaaa}
	perm := []byte{0}
	identity := func(_ *Abc_TtHieMan_t, _ []uint64, n int32, perm []byte, _ int32) uint32 {
		for i := int32(0); i < n; i++ {
			perm[i] = byte(i)
		}
		return 0
	}
	phase := Abc_TtCanonicizeWrap(identity, nil, truth, 1, perm, 0)
	if truth[0] != 0x5555555555555555 || phase != 2 {
		t.Fatalf("wrapped complement lost output phase: truth%x phase%x", truth[0], phase)
	}
}

func TestHierarchicalExactWideConstant(t *testing.T) {
	p := Abc_TtHieManStart(16, 5)
	truth := make([]uint64, 1024)
	for i := range truth {
		truth[i] = ^uint64(0)
	}
	perm := make([]byte, 16)
	Abc_TtCanonicizeHie(p, truth, 16, perm, 1)
	for _, w := range truth {
		if w != 0 {
			t.Fatal("constant did not canonicalize to zero")
		}
	}
}
