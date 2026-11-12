package canon

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

func TestPermutationAndPhaseOriginalC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section collection required")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional C oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	source := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../src"))
	dir := t.TempDir()
	main, bin := filepath.Join(dir, "main.c"), filepath.Join(dir, "oracle")
	code := `#include "opt/dau/dauCanon.c"
int main(void){int n,mode;word t[1024];char p[16];unsigned long long x;while(scanf("%d%d",&n,&mode)==2){int w=Abc_TtWordNum(n);for(int i=0;i<w;i++){if(scanf("%llx",&x)!=1)return 2;t[i]=x;}for(int i=0;i<n;i++)p[i]=i;unsigned phase=mode?Abc_TtCanonicizePhase(t,n):Abc_TtCanonicizePerm(t,n,p);printf("%u",phase);for(int i=0;i<n;i++)printf(" %d",p[i]);for(int i=0;i<w;i++)printf(" %llx",(unsigned long long)t[i]);puts("");}return 0;}`
	if err := os.WriteFile(main, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", source, main, "-o", bin)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(b))
	}
	type sample struct {
		n     int32
		mode  int
		truth []uint64
	}
	var samples []sample
	for f := uint64(0); f < 256; f++ {
		for mode := 0; mode < 2; mode++ {
			samples = append(samples, sample{3, mode, []uint64{f * 0x0101010101010101}})
		}
	}
	rng := rand.New(rand.NewSource(143))
	for n := int32(0); n <= 16; n++ {
		words := 1
		if n > 6 {
			words = 1 << uint(n-6)
		}
		for trial := 0; trial < 20; trial++ {
			table := make([]uint64, words)
			for i := range table {
				table[i] = rng.Uint64()
				if trial == 0 {
					table[i] = ^uint64(0)
				}
			}
			if n < 6 {
				size := uint(1) << uint(n)
				table[0] &= (uint64(1) << size) - 1
				for s := size; s < 64; s *= 2 {
					table[0] |= table[0] << s
				}
			}
			for mode := 0; mode < 2; mode++ {
				samples = append(samples, sample{n, mode, append([]uint64(nil), table...)})
			}
		}
	}
	var input bytes.Buffer
	for _, s := range samples {
		fmt.Fprint(&input, s.n, " ", s.mode)
		for _, w := range s.truth {
			fmt.Fprintf(&input, " %x", w)
		}
		fmt.Fprintln(&input)
	}
	cmd = exec.Command(bin)
	cmd.Stdin = &input
	data, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(data))
	}
	scan := bufio.NewScanner(bytes.NewReader(data))
	scan.Buffer(make([]byte, 4096), 1<<20)
	for index, s := range samples {
		if !scan.Scan() {
			t.Fatal("missing source sample", index)
		}
		f := strings.Fields(scan.Text())
		if len(f) != 1+int(s.n)+len(s.truth) {
			t.Fatal("source diagnostic", scan.Text())
		}
		table := append([]uint64(nil), s.truth...)
		perm := make([]byte, s.n)
		for i := range perm {
			perm[i] = byte(i)
		}
		var phase uint32
		if s.mode == 0 {
			phase = CanonicalizePermutation(table, s.n, perm)
		} else {
			phase = CanonicalizePhase(table, s.n)
		}
		p, err := strconv.ParseUint(f[0], 10, 32)
		if err != nil || phase != uint32(p) {
			t.Fatalf("sample%d n%d mode%d phase=%x C=%s", index, s.n, s.mode, phase, f[0])
		}
		for i, v := range perm {
			p, _ := strconv.ParseUint(f[1+i], 10, 8)
			if byte(p) != v {
				t.Fatalf("sample%d permutation mismatch", index)
			}
		}
		for i, v := range table {
			p, _ := strconv.ParseUint(f[1+int(s.n)+i], 16, 64)
			if p != v {
				t.Fatalf("sample%d n%d mode%d word%d=%x C=%x", index, s.n, s.mode, i, v, p)
			}
		}
	}
}
