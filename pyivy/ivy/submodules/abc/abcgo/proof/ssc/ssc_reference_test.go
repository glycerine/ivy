package ssc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// The oracle links unchanged original SSC/GIA/SAT object files. Optimization
// outside the care set and native scheduling are deliberately unconstrained.
func TestOriginalCConstrainedSweeps(t *testing.T) {
	if os.Getenv("SSC_BUILD_ORACLE") != "1" {
		t.Skip("set SSC_BUILD_ORACLE=1 to build the original C oracle from source")
	}
	cc, e := exec.LookPath("cc")
	if e != nil {
		t.Skip("C compiler unavailable")
	}
	cxx, e := exec.LookPath("g++")
	if e != nil {
		t.Skip("C++ linker unavailable")
	}
	if _, e = exec.LookPath("make"); e != nil {
		t.Skip("make unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	source := os.Getenv("SSC_SOURCE_ROOT")
	if source == "" {
		source = filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	}
	d := t.TempDir()
	// Build in an isolated copy: object files in the checkout may be stale or
	// concurrently rebuilt with a different set of ABC compile-time options.
	root := filepath.Dir(source)
	build := filepath.Join(d, "abc")
	if e = os.MkdirAll(build, 0755); e != nil {
		t.Fatal(e)
	}
	copyFile := func(from, to string) error {
		in, err := os.Open(from)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(to)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	}
	for _, name := range []string{"Makefile", "arch_flags.c", "depends.sh"} {
		if e = copyFile(filepath.Join(root, name), filepath.Join(build, name)); e != nil {
			t.Fatal(e)
		}
	}
	if e = os.Chmod(filepath.Join(build, "depends.sh"), 0755); e != nil {
		t.Fatal(e)
	}
	e = filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(build, "src", rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if !entry.Type().IsRegular() || strings.HasSuffix(path, ".o") || strings.HasSuffix(path, ".d") {
			return nil
		}
		return copyFile(path, target)
	})
	if e != nil {
		t.Fatal(e)
	}
	makeCmd := exec.Command("make", "-j4", "ABC_USE_STDINT_H=1", "ABC_USE_NO_READLINE=1", "ABC_USE_NO_PTHREADS=1")
	makeCmd.Dir = build
	if out, err := makeCmd.CombinedOutput(); err != nil {
		t.Fatalf("build original ABC: %v\n%s", err, out)
	}
	source = filepath.Join(build, "src")
	c, o, exe := filepath.Join(d, "oracle.c"), filepath.Join(d, "oracle.o"), filepath.Join(d, "oracle")
	code := `#include "aig/gia/gia.h"
#include "proof/ssc/ssc.h"
int main(void){int n,a,b,j,i,q,k;Ssc_Pars_t pars;while(scanf("%d",&n)==1){Gia_Man_t*p=Gia_ManStart(500),*c=Gia_ManStart(50),*r;int es[100],ci[5];Gia_Obj_t*x;unsigned t;es[0]=0;Gia_ManHashStart(p);for(i=0;i<5;i++)es[i+1]=Gia_ManAppendCi(p);for(i=0;i<n;i++){scanf("%d%d",&a,&b);es[i+6]=Gia_ManHashAnd(p,es[a>>1]^(a&1),es[b>>1]^(b&1));}for(i=0;i<4;i++){scanf("%d",&a);Gia_ManAppendCo(p,es[a>>1]^(a&1));}for(i=0;i<5;i++)ci[i]=Gia_ManAppendCi(c);Gia_ManHashStart(c);Gia_ManAppendCo(c,Gia_ManHashXor(c,ci[0],ci[1]));Ssc_ManSetDefaultParams(&pars);r=Ssc_PerformSweeping(p,c,&pars);Gia_ManForEachCo(r,x,j){t=0;for(q=0;q<32;q++){Gia_ManConst0(r)->fMark0=0;Gia_ManForEachCi(r,x,i)x->fMark0=(q>>i)&1;Gia_ManForEachAnd(r,x,i)x->fMark0=(Gia_ObjFanin0(x)->fMark0^Gia_ObjFaninC0(x))&(Gia_ObjFanin1(x)->fMark0^Gia_ObjFaninC1(x));x=Gia_ManCo(r,j);k=Gia_ObjFanin0(x)->fMark0^Gia_ObjFaninC0(x);t|=(unsigned)k<<q;}printf("%x ",t);}puts("");Gia_ManStop(r);Gia_ManStop(p);Gia_ManStop(c);}return 0;}`
	if e = os.WriteFile(c, []byte(code), 0600); e != nil {
		t.Fatal(e)
	}
	if out, e := exec.Command(cc, "-std=gnu99", "-DABC_USE_STDINT_H=1", "-DABC_USE_CUDD=1", "-I"+source, "-c", c, "-o", o).CombinedOutput(); e != nil {
		t.Fatalf("compile %v %s", e, out)
	}
	args := []string{o}
	e = filepath.WalkDir(source, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && strings.HasSuffix(p, ".o") && p != filepath.Join(source, "base/main/main.o") {
			args = append(args, p)
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	args = append(args, "-lreadline", "-lpthread", "-lm", "-ldl", "-lz", "-o", exe)
	if out, e := exec.Command(cxx, args...).CombinedOutput(); e != nil {
		t.Fatalf("link %v %s", e, out)
	}
	rng := rand.New(rand.NewSource(917))
	var input strings.Builder
	var results []*gia.Gia_Man_t
	for trial := 0; trial < 40; trial++ {
		b := newBuilder(5, 100)
		es := append([]int32{0}, b.graph.InputLiterals()...)
		fmt.Fprintln(&input, 30)
		for i := 0; i < 30; i++ {
			a, c := rng.Intn(len(es))*2+rng.Intn(2), rng.Intn(len(es))*2+rng.Intn(2)
			fmt.Fprintln(&input, a, c)
			es = append(es, b.and(es[a>>1]^int32(a&1), es[c>>1]^int32(c&1)))
		}
		for i := 0; i < 4; i++ {
			a := rng.Intn(len(es))*2 + rng.Intn(2)
			fmt.Fprintln(&input, a)
			gia.Gia_ManAppendCo(b.graph, es[a>>1]^int32(a&1))
		}
		care := newBuilder(5, 30)
		ci := care.graph.InputLiterals()
		gia.Gia_ManAppendCo(care.graph, care.xor(ci[0], ci[1]))
		results = append(results, Ssc_PerformSweeping(b.graph, care.graph, nil))
	}
	command := exec.Command(exe)
	command.Stdin = strings.NewReader(input.String())
	out, e := command.CombinedOutput()
	if e != nil {
		t.Fatalf("oracle %v %s", e, out)
	}
	fields := strings.Fields(string(out))
	if len(fields) != len(results)*4 {
		t.Fatalf("unexpected oracle output %s", out)
	}
	for sample, r := range results {
		for j := 0; j < 4; j++ {
			parsed, e := strconv.ParseUint(fields[sample*4+j], 16, 32)
			if e != nil {
				t.Fatal(e)
			}
			truth := uint32(parsed)
			for mask := 0; mask < 32; mask++ {
				if (mask & 1) != ((mask >> 1) & 1) {
					continue
				}
				in := make([]bool, 5)
				for i := range in {
					in[i] = mask&(1<<i) != 0
				}
				got, e := r.Evaluate(in)
				if e != nil {
					t.Fatal(e)
				}
				if got[j] != (truth&(1<<uint(mask)) != 0) {
					t.Fatalf("sample%d output%d mask%d differs from original C", sample, j, mask)
				}
			}
		}
	}
}
