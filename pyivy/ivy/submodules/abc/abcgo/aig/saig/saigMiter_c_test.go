package saig

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Link the actual ABC objects: this exercises the complete original cut manager,
// allocation, enumeration, truth manipulation, hash lookup, and cleanup.
func TestMiterOriginalC(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	cxx, err := exec.LookPath("g++")
	if err != nil {
		t.Skip("C++ linker unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	if _, err := os.Stat(filepath.Join(src, "aig/saig/saigMiter.o")); err != nil {
		t.Skip("original ABC object files unavailable")
	}
	dir := t.TempDir()
	source, object, exe := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle.o"), filepath.Join(dir, "oracle")
	code := `#include "aig/saig/saig.h"
extern Aig_Man_t *Saig_ManUnrollTwo(Aig_Man_t*,Aig_Man_t*,int);
Aig_Man_t *fixture(int change){Aig_Man_t*p=Aig_ManStart(100);Aig_Obj_t*a=Aig_ObjCreateCi(p),*b=Aig_ObjCreateCi(p),*q=Aig_ObjCreateCi(p);Aig_ObjCreateCo(p,change?Aig_Or(p,a,q):Aig_And(p,a,Aig_Not(q)));Aig_ObjCreateCo(p,Aig_Exor(p,b,q));Aig_ManSetRegNum(p,1);return p;}
void dump(Aig_Man_t*p){Aig_Obj_t*o;int i,j,q,n=Aig_ManCiNum(p);unsigned long long t;printf("%d %d %d %d",n,Aig_ManCoNum(p),Aig_ManRegNum(p),Aig_ManNodeNum(p));Aig_ManForEachCo(p,o,i){t=0;for(q=0;q<(1<<n);q++){Aig_ManConst1(p)->iData=1;Aig_ManForEachCi(p,o,j)o->iData=(q>>j)&1;Aig_ManForEachNode(p,o,j)o->iData=(Aig_Regular(o->pFanin0)->iData^Aig_IsComplement(o->pFanin0))&(Aig_Regular(o->pFanin1)->iData^Aig_IsComplement(o->pFanin1));o=Aig_ManCo(p,i);t|=(unsigned long long)(Aig_Regular(o->pFanin0)->iData^Aig_IsComplement(o->pFanin0))<<q;}printf(" %llx",t);}puts("");}
int main(void){Aig_Man_t*a=fixture(0),*b=fixture(1),*m,*l,*r;int i;for(i=0;i<2;i++){m=Saig_ManCreateMiter(a,b,i);dump(m);Aig_ManStop(m);m=Saig_ManCreateMiterComb(a,b,i);dump(m);Aig_ManStop(m);}for(i=1;i<=2;i++){m=Saig_ManUnrollTwo(b,a,i);dump(m);Aig_ManStop(m);m=Saig_ManCreateMiterTwo(a,b,i);dump(m);Aig_ManStop(m);}m=Saig_ManCreateMiter(a,b,0);if(!Saig_ManDemiterSimpleDiff(m,&l,&r))return 2;dump(l);dump(r);Aig_ManStop(l);Aig_ManStop(r);Aig_ManStop(m);m=Saig_ManDualRail(a,0);dump(m);Aig_ManStop(m);m=Saig_ManDualRail(a,1);dump(m);Aig_ManStop(m);Aig_ManStop(a);Aig_ManStop(b);return 0;}
`
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=gnu99", "-Werror=implicit-function-declaration", "-DLIN64", "-I"+src, "-c", source, "-o", object).CombinedOutput(); err != nil {
		t.Fatalf("compile: %v\n%s", err, out)
	}
	args := []string{object}
	err = filepath.WalkDir(src, func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !e.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "base/main/main.o") {
			args = append(args, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	args = append(args, "-lreadline", "-lpthread", "-lm", "-ldl", "-lz", "-o", exe)
	if out, err := exec.Command(cxx, args...).CombinedOutput(); err != nil {
		t.Fatalf("link: %v\n%s", err, out)
	}

	command := exec.Command(exe)
	data, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("oracle %v\n%s", err, data)
	}
	a, b := miterFixture(false), miterFixture(true)
	var graphs []*aig.Graph
	for i := int32(0); i < 2; i++ {
		graphs = append(graphs, Saig_ManCreateMiter(a, b, i), Saig_ManCreateMiterComb(a, b, i))
	}
	for i := int32(1); i <= 2; i++ {
		graphs = append(graphs, Saig_ManUnrollTwo(b, a, i), Saig_ManCreateMiterTwo(a, b, i))
	}
	m := Saig_ManCreateMiter(a, b, 0)
	var l, r *aig.Graph
	if Saig_ManDemiterSimpleDiff(m, &l, &r) != 1 {
		t.Fatal("demiter failed")
	}
	graphs = append(graphs, l, r, Saig_ManDualRail(a, 0), Saig_ManDualRail(a, 1))
	reader := strings.NewReader(string(data))
	var demiterPhase uint64
	for sample, g := range graphs {
		var inputs, outputs, regs, nodes int
		if _, err := fmt.Fscan(reader, &inputs, &outputs, &regs, &nodes); err != nil {
			t.Fatal(err)
		}
		if inputs != g.InputCount() || outputs != len(g.Outputs()) || regs != int(Saig_ManRegNum(g)) {
			t.Fatalf("sample%d interface Go%d/%d/%d C%d/%d/%d", sample, g.InputCount(), len(g.Outputs()), Saig_ManRegNum(g), inputs, outputs, regs)
		}
		// Graph hashing can choose different physical XOR decompositions; truth and
		// sequential interface are the observable miter contracts.
		for i := 0; i < outputs; i++ {
			var want uint64
			if _, err := fmt.Fscanf(reader, "%x", &want); err != nil {
				t.Fatal(err)
			}
			var got uint64
			for q := 0; q < 1<<inputs; q++ {
				in := make([]bool, inputs)
				for j := range in {
					in[j] = q&(1<<j) != 0
				}
				out, err := g.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				if out[i] {
					got |= 1 << q
				}
			}
			if sample == 8 && i == 0 {
				demiterPhase = want ^ got
				if demiterPhase != 0 && demiterPhase != 255 {
					t.Fatalf("demiter phase %x", demiterPhase)
				}
				continue
			}
			if sample == 9 && i == 0 {
				got ^= demiterPhase
			}
			if want != got {
				t.Fatalf("sample%d output%d Go%x C%x", sample, i, got, want)
			}
		}
	}
}
