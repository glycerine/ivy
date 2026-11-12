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

func TestTruthCountsOriginalC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF function section removal required")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional C oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	source := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	dir := t.TempDir()
	main, binary := filepath.Join(dir, "main.c"), filepath.Join(dir, "oracle")
	code := `#include "opt/dau/dauCanon.c"
#include "opt/dau/dauCount.c"
int main(void){int n;word t[1024];unsigned long long x;while(scanf("%d",&n)==1){int w=Abc_TtWordNum(n),a[16]={0},b[16]={0},s[16]={0};for(int i=0;i<w;i++){if(scanf("%llx",&x)!=1)return 2;t[i]=x;}int na=Abc_TtCountOnesInCofsSimple(t,n,a);int nb=Abc_TtCountOnesInCofsQuick(t,n,b);int ns=Abc_TtScc(t,n);Abc_TtSccInCofs(t,n,s);printf("%d %d %d",na,nb,ns);for(int i=0;i<n;i++)printf(" %d %d %d",a[i],b[i],s[i]);puts("");}return 0;}`
	if err := os.WriteFile(main, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", source, main, "-o", binary)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatal(err, string(b))
	}
	type sample struct {
		n int32
		t []uint64
	}
	var samples []sample
	var input bytes.Buffer
	rng := rand.New(rand.NewSource(52))
	for n := int32(0); n <= 16; n++ {
		for trial := 0; trial < 8; trial++ {
			table := make([]uint64, dauCountWords(n))
			for i := range table {
				table[i] = rng.Uint64()
				if trial == 0 {
					table[i] = ^uint64(0)
				}
				if trial == 1 {
					table[i] = 0
				}
			}
			Abc_TtNormalizeSmallTruth(table, n)
			samples = append(samples, sample{n, table})
			fmt.Fprint(&input, n)
			for _, w := range table {
				fmt.Fprintf(&input, " %x", w)
			}
			fmt.Fprintln(&input)
		}
	}
	cmd = exec.Command(binary)
	cmd.Stdin = &input
	data, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(data))
	}
	scan := bufio.NewScanner(bytes.NewReader(data))
	sawOverflow, sawOverwrite := false, false
	for index, s := range samples {
		if !scan.Scan() {
			t.Fatal("missing source sample", index)
		}
		f := strings.Fields(scan.Text())
		if len(f) != 3+3*int(s.n) {
			t.Fatal("source diagnostic", scan.Text())
		}
		get := func(i int) int32 {
			v, err := strconv.ParseInt(f[i], 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			return int32(v)
		}
		a, b, c := make([]int32, s.n), make([]int32, s.n), make([]int32, s.n)
		want := Abc_TtCountOnesInCofsSimple(s.t, s.n, a)
		quick := Abc_TtCountOnesInCofsQuick(s.t, s.n, b)
		scc := Abc_TtScc(s.t, s.n)
		Abc_TtSccInCofs(s.t, s.n, c)
		if get(0) != want || get(2) != scc {
			t.Fatalf("n%d source totals differ", s.n)
		}
		if s.n == 8 && get(1) != quick {
			sawOverflow = true
			if get(1) != 0 || quick != 256 {
				t.Fatal("unexpected source packed overflow")
			}
		} else if s.n <= 7 && get(1) != quick {
			t.Fatal("source quick total mismatch")
		}
		for v := range a {
			if get(3+3*v) != a[v] || get(5+3*v) != c[v] {
				t.Fatalf("n%d v%d source simple/SCC mismatch", s.n, v)
			}
			if s.n > 9 && get(4+3*v) != b[v] {
				sawOverwrite = true
			}
			if s.n > 0 && a[v] != b[v] {
				t.Fatalf("native quick cofactor differs at n%d v%d", s.n, v)
			}
		}
	}
	if !sawOverflow || !sawOverwrite {
		t.Fatal("original-C defect regressions were not exercised")
	}
}
