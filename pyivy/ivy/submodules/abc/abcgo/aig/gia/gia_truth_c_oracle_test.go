package gia

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
	"testing"
)

func TestGiaTruthSourceOracle(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section collection required")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional C oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	source := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	dir := t.TempDir()
	main := filepath.Join(dir, "main.c")
	bin := filepath.Join(dir, "oracle")
	code := `#include "aig/gia/giaTruth.c"
int main(void) {
    int n, mode;
    while (scanf("%d%d", &n, &mode) == 2) {
        Gia_Man_t m = {0};
        m.nObjs = n;
        m.pObjs = calloc(n, sizeof(Gia_Obj_t));
        m.vCis = Vec_IntAlloc(16);
        for (int i = 0; i < n; i++) {
            unsigned a, b, c, d, e;
            if (scanf("%u%u%u%u%u", &a, &b, &c, &d, &e) != 5) return 2;
            m.pObjs[i].iDiff0 = a; m.pObjs[i].iDiff1 = b;
            m.pObjs[i].fCompl0 = c; m.pObjs[i].fCompl1 = d;
            m.pObjs[i].fTerm = e;
            if (e && a == GIA_NONE) Vec_IntPush(m.vCis, i);
        }
        if (mode) {
            word *t = Gia_ObjComputeTruthTable(&m, m.pObjs + n - 1);
            for (int i = 0; i < m.nTtWords; i++) printf("%llx\n", (unsigned long long)t[i]);
            Gia_ObjComputeTruthTableStop(&m);
        } else {
            for (int i = 0; i < n; i++) printf("%llx\n", (unsigned long long)Gia_LutComputeTruth6Simple_rec(&m, i));
        }
        Vec_IntFree(m.vCis);
        free(m.pObjs);
    }
    return 0;
}`
	if err := os.WriteFile(main, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", source, main, "-o", bin)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("C oracle build: %v\n%s", err, b)
	}
	rng := rand.New(rand.NewSource(271))
	var input bytes.Buffer
	var want []uint64
	for trial := 0; trial < 25; trial++ {
		p := Gia_ManStart(100)
		truthTestLeaves(p, 6)
		for j := 0; j < 35; j++ {
			a := int32(rng.Intn(int(p.nObjs) * 2))
			b := int32(rng.Intn(int(p.nObjs) * 2))
			lit := giaTruthAnd(p, a, b, false)
			if lit > 1 && rng.Intn(3) == 0 {
				o := p.pObjs[lit>>1]
				if Gia_ObjIsAnd(o) != 0 {
					o.iDiff0, o.iDiff1 = o.iDiff1, o.iDiff0
					o.fCompl0, o.fCompl1 = o.fCompl1, o.fCompl0
				}
			}
		}
		fmt.Fprintln(&input, p.nObjs, 0)
		for _, o := range p.pObjs {
			fmt.Fprintln(&input, o.iDiff0, o.iDiff1, o.fCompl0, o.fCompl1, o.fTerm)
			want = append(want, Gia_LutComputeTruth6Simple_rec(p, o.id))
		}
	}
	// The original multiword routine is a distinct algorithm; compare it too.
	for n := 7; n <= 12; n++ {
		for trial := 0; trial < 5; trial++ {
			p := Gia_ManStart(100)
			truthTestLeaves(p, n)
			for j := 0; j < 35; j++ {
				a := int32(1+rng.Intn(int(p.nObjs)-1))*2 + int32(rng.Intn(2))
				b := int32(1+rng.Intn(int(p.nObjs)-1))*2 + int32(rng.Intn(2))
				if a>>1 != b>>1 {
					Gia_ManAppendAnd(p, a, b)
				}
			}
			fmt.Fprintln(&input, p.nObjs, 1)
			for _, o := range p.pObjs {
				fmt.Fprintln(&input, o.iDiff0, o.iDiff1, o.fCompl0, o.fCompl1, o.fTerm)
			}
			want = append(want, Gia_ObjComputeTruthTable(p, p.pObjs[p.nObjs-1])...)
		}
	}
	cmd = exec.Command(bin)
	cmd.Stdin = &input
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err, string(b))
	}
	scan := bufio.NewScanner(bytes.NewReader(b))
	for i, w := range want {
		if !scan.Scan() {
			t.Fatalf("missing source result %d", i)
		}
		got, err := strconv.ParseUint(scan.Text(), 16, 64)
		if err != nil || got != w {
			t.Fatalf("source result %d = %s, Go = %x", i, scan.Text(), w)
		}
	}
}
