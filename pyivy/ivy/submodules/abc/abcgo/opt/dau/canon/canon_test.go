package canon

import (
	"bufio"
	"bytes"
	"fmt"
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

func TestCanonicalizeCOracle(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("C oracle uses ELF function-section garbage collection")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional source differential oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	source := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../src"))
	root := t.TempDir()
	main := filepath.Join(root, "main.c")
	binary := filepath.Join(root, "oracle")
	code := `#include "opt/dau/dauCanon.c"
int main(void){int n;unsigned long long x;word t[1024];char p[16];while(scanf("%d",&n)==1){int w=n<=6?1:1<<(n-6);for(int i=0;i<w;i++){if(scanf("%llx",&x)!=1)return 2;t[i]=x;}unsigned phase=Abc_TtCanonicize(t,n,p);printf("%u",phase);for(int i=0;i<n;i++)printf(" %d",p[i]);for(int i=0;i<w;i++)printf(" %llx",(unsigned long long)t[i]);puts("");}return 0;}`
	if err := os.WriteFile(main, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", source, main, "-o", binary)
	if data, err := command.CombinedOutput(); err != nil {
		t.Fatalf("compile actual dauCanon.c oracle: %v\n%s", err, data)
	}
	type sample struct {
		n     int32
		truth []uint64
	}
	var samples []sample
	for f := uint64(0); f < 256; f++ {
		samples = append(samples, sample{3, []uint64{f * 0x0101010101010101}})
	}
	rng := rand.New(rand.NewSource(44))
	for n := int32(2); n <= 16; n++ {
		words := 1
		if n > 6 {
			words = 1 << uint(n-6)
		}
		// Structured functions force count ties and adjacent refinements across
		// the 64-bit word boundary; random functions alone seldom exercise these.
		for kind := 0; kind < 5; kind++ {
			table := make([]uint64, words)
			for a := 0; a < words*64; a++ {
				x := a & ((1 << uint(n)) - 1)
				on := false
				switch kind {
				case 0:
					on = x == (1<<uint(n))-1
				case 1:
					on = bits.OnesCount(uint(x))&1 != 0
				case 2:
					on = bits.OnesCount(uint(x)) >= int(n)/2
				case 3:
					on = x&(1<<uint(n-1)) != 0
				case 4:
					for v := int32(0); v+1 < n; v += 2 {
						on = on != (x&(3<<uint(v)) == (3 << uint(v)))
					}
				}
				if on {
					table[a>>6] |= 1 << uint(a&63)
				}
			}
			samples = append(samples, sample{n, table})
		}
		for trial := 0; trial < 15; trial++ {
			table := make([]uint64, words)
			for i := range table {
				table[i] = rng.Uint64()
			}
			if n < 6 {
				size := uint(1) << uint(n)
				table[0] &= uint64(1)<<size - 1
				for shift := size; shift < 64; shift *= 2 {
					table[0] |= table[0] << shift
				}
			}
			samples = append(samples, sample{n, table})
		}
	}
	var input bytes.Buffer
	for _, s := range samples {
		fmt.Fprint(&input, s.n)
		for _, x := range s.truth {
			fmt.Fprintf(&input, " %x", x)
		}
		fmt.Fprintln(&input)
	}
	command = exec.Command(binary)
	command.Stdin = &input
	data, err := command.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(data))
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	for index, s := range samples {
		if !scanner.Scan() {
			t.Fatalf("oracle missing sample %d: %v", index, scanner.Err())
		}
		fields := strings.Fields(scanner.Text())
		if len(fields) != 1+int(s.n)+len(s.truth) {
			t.Fatalf("oracle diagnostic: %s", scanner.Text())
		}
		table := append([]uint64(nil), s.truth...)
		perm := make([]byte, s.n)
		phase := Canonicalize(table, s.n, perm)
		expected, err := strconv.ParseUint(fields[0], 10, 32)
		if err != nil || uint32(expected) != phase {
			t.Fatalf("sample=%d n=%d phase=%x expected=%x", index, s.n, phase, expected)
		}
		for i, p := range perm {
			expected, _ := strconv.Atoi(fields[i+1])
			if int(p) != expected {
				t.Fatalf("sample=%d permutation %v != %s", index, perm, scanner.Text())
			}
		}
		for i, x := range table {
			expected, _ := strconv.ParseUint(fields[1+int(s.n)+i], 16, 64)
			if x != expected {
				t.Fatalf("sample=%d n=%d word=%d got=%x expected=%x", index, s.n, i, x, expected)
			}
		}
	}
}

// This check remains active without a C compiler and verifies returned metadata.
func TestCanonicalizeTransformation(t *testing.T) {
	rng := rand.New(rand.NewSource(93))
	for n := int32(0); n <= 10; n++ {
		for trial := 0; trial < 20; trial++ {
			words := 1
			if n > 6 {
				words = 1 << uint(n-6)
			}
			original := make([]uint64, words)
			for i := range original {
				original[i] = rng.Uint64()
			}
			if n < 6 {
				size := uint(1) << uint(n)
				original[0] &= (uint64(1) << size) - 1
				for s := size; s < 64; s *= 2 {
					original[0] |= original[0] << s
				}
			}
			table := append([]uint64(nil), original...)
			perm := make([]byte, n)
			phase := Canonicalize(table, n, perm)
			seen := map[byte]bool{}
			for _, v := range perm {
				if int(v) >= int(n) || seen[v] {
					t.Fatal("invalid permutation", perm)
				}
				seen[v] = true
			}
			for a := 0; a < 1<<uint(n); a++ {
				old := 0
				for j, v := range perm {
					if (a>>uint(j))&1 != int(phase>>uint(j))&1 {
						old |= 1 << v
					}
				}
				want := (original[old>>6]>>uint(old&63))&1 ^ uint64((phase>>uint(n))&1)
				if got := (table[a>>6] >> uint(a&63)) & 1; got != want {
					t.Fatalf("n=%d assignment=%d metadata does not reconstruct transformation", n, a)
				}
			}
		}
	}
}
