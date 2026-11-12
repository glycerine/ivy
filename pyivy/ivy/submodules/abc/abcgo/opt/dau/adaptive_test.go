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

func TestAdaptiveOriginalC(t *testing.T) {
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
int main(void){Abc_TtHieMan_t *m[2000]={0};int id,n,l,mode,flag;word t[1024];char p[16];unsigned long long x;while(scanf("%d%d%d%d%d",&id,&n,&l,&mode,&flag)==5){Abc_TtHieMan_t *h=NULL;if(l){if(!m[id])m[id]=Abc_TtHieManStart(n,l);h=m[id];}int w=Abc_TtWordNum(n);for(int i=0;i<w;i++){if(scanf("%llx",&x)!=1)return 2;t[i]=x;}memset(p,0,16);unsigned phase=mode?Abc_TtCanonicizeCA(h,t,n,p,flag):Abc_TtCanonicizeAda(h,t,n,p,flag);printf("%u",phase);for(int i=0;i<n;i++)printf(" %d",p[i]);for(int i=0;i<w;i++)printf(" %llx",(unsigned long long)t[i]);puts("");}for(int i=0;i<2000;i++)if(m[i])Abc_TtHieManStop(m[i]);return 0;}`
	if err := os.WriteFile(main, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", src, main, "-o", bin)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(b))
	}
	type config struct{ mode, flag int32 }
	configs := []config{{0, 0}, {0, 99}, {0, 100}, {0, 199}, {0, 1000}, {0, 1099}, {0, 1199}, {1, 0}, {1, 1}}
	type sample struct {
		id, n, levels, mode, flag int32
		truth                     []uint64
	}
	var samples []sample
	add := func(n int32, truth []uint64) {
		for i, c := range configs {
			if n > 0 && (truth[0] == 0 || truth[0] == ^uint64(0)) && (c.mode == 1 || c.flag >= 1000) {
				continue
			}
			id := n*100 + int32(i)
			samples = append(samples, sample{id, n, 5, c.mode, c.flag, append([]uint64(nil), truth...)})
		}
	}
	for f := uint64(0); f < 256; f++ {
		add(3, []uint64{f * 0x0101010101010101})
	}
	rng := rand.New(rand.NewSource(93))
	for n := int32(0); n <= 10; n++ {
		for trial := 0; trial < 12; trial++ {
			truth := make([]uint64, dauCountWords(n))
			for i := range truth {
				truth[i] = rng.Uint64()
			}
			Abc_TtNormalizeSmallTruth(truth, n)
			add(n, truth)
			if trial%3 == 0 {
				add(n, truth)
			}
		}
	}
	var input bytes.Buffer
	for _, s := range samples {
		fmt.Fprint(&input, s.id, " ", s.n, " ", s.levels, " ", s.mode, " ", s.flag)
		for _, w := range s.truth {
			fmt.Fprintf(&input, " %x", w)
		}
		fmt.Fprintln(&input)
	}
	cmd = exec.Command(bin)
	cmd.Stdin = &input
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(out))
	}
	scan := bufio.NewScanner(bytes.NewReader(out))
	scan.Buffer(make([]byte, 4096), 1<<20)
	managers := map[int32]*Abc_TtHieMan_t{}
	for index, s := range samples {
		if !scan.Scan() {
			t.Fatal("missing source sample", index)
		}
		fields := strings.Fields(scan.Text())
		if len(fields) != 1+int(s.n)+len(s.truth) {
			t.Fatalf("sample%d n%d mode%d flag%d source diagnostic: %s", index, s.n, s.mode, s.flag, scan.Text())
		}
		p := managers[s.id]
		if p == nil {
			p = Abc_TtHieManStart(s.n, s.levels)
			managers[s.id] = p
		}
		truth := append([]uint64(nil), s.truth...)
		perm := make([]byte, s.n)
		phase := uint32(0)
		if s.mode == 0 {
			phase = Abc_TtCanonicizeAda(p, truth, s.n, perm, s.flag)
		} else {
			phase = Abc_TtCanonicizeCA(p, truth, s.n, perm, s.flag)
		}
		v, _ := strconv.ParseUint(fields[0], 10, 32)
		if uint32(v) != phase {
			t.Fatalf("sample%d n%d mode%d flag%d phase Go%x C%x", index, s.n, s.mode, s.flag, phase, v)
		}
		for i, x := range perm {
			v, _ := strconv.ParseUint(fields[1+i], 10, 8)
			if byte(v) != x {
				t.Fatalf("sample%d n%d mode%d flag%d perm%d Go%d C%d", index, s.n, s.mode, s.flag, i, x, v)
			}
		}
		for i, x := range truth {
			v, _ := strconv.ParseUint(fields[1+int(s.n)+i], 16, 64)
			if v != x {
				t.Fatalf("sample%d n%d mode%d flag%d word%d Go%x C%x", index, s.n, s.mode, s.flag, i, x, v)
			}
		}
	}
}
func TestAdaptiveMetadataAndConstants(t *testing.T) {
	rng := rand.New(rand.NewSource(71))
	for n := int32(1); n <= 10; n++ {
		for _, threshold := range []int32{0, 99, 100, 199, 1000, 1099, 1199} {
			for trial := 0; trial < 8; trial++ {
				truth := make([]uint64, dauCountWords(n))
				for i := range truth {
					truth[i] = rng.Uint64()
					if trial == 0 {
						truth[i] = 0
					}
					if trial == 1 {
						truth[i] = ^uint64(0)
					}
				}
				Abc_TtNormalizeSmallTruth(truth, n)
				original := append([]uint64(nil), truth...)
				p := Abc_TtHieManStart(n, 5)
				perm := make([]byte, n)
				phase := Abc_TtCanonicizeAda(p, truth, n, perm, threshold)
				var state Abc_TgMan_t
				Abc_TgInitMan(&state, truth, int(n), 0, nil)
				state.original = original
				state.uPhase = phase
				for i, v := range perm {
					state.pPermT[i] = int(v)
					state.pPermTRev[v] = i
				}
				if !Abc_TgCannonVerify(&state) {
					t.Fatalf("invalid adaptive transform n%d threshold%d trial%d", n, threshold, trial)
				}
			}
		}
	}
}

func TestCostAwareZeroSupportMetadata(t *testing.T) {
	for n := int32(1); n <= 10; n++ {
		for _, one := range []bool{false, true} {
			for _, mode := range []int32{0, 1} {
				truth := make([]uint64, dauCountWords(n))
				if one {
					for i := range truth {
						truth[i] = ^uint64(0)
					}
				}
				original := append([]uint64(nil), truth...)
				p := Abc_TtHieManStart(n, 5)
				perm := make([]byte, n)
				phase := Abc_TtCanonicizeCA(p, truth, n, perm, mode)
				var state Abc_TgMan_t
				Abc_TgInitMan(&state, truth, int(n), 0, nil)
				state.original = original
				state.uPhase = phase
				for i, v := range perm {
					state.pPermT[i] = int(v)
					state.pPermTRev[v] = i
				}
				if !Abc_TgCannonVerify(&state) {
					t.Fatal("zero-support cost-aware metadata", n, one, mode)
				}
			}
		}
	}
}
