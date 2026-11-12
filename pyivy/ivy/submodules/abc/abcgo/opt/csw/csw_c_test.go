package csw

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Link the actual ABC objects: this exercises the complete original cut manager,
// allocation, enumeration, truth manipulation, hash lookup, and cleanup.
func TestSweepOriginalC(t *testing.T) {
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
	if _, err := os.Stat(filepath.Join(src, "opt/csw/cswCore.o")); err != nil {
		t.Skip("original ABC object files unavailable")
	}
	dir := t.TempDir()
	source, object, exe := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle.o"), filepath.Join(dir, "oracle")
	code := `#include "opt/csw/cswInt.h"
int main(void){int n,c,k,i,a,b,j,q;while(scanf("%d%d%d",&n,&c,&k)==3){Aig_Man_t*p=Aig_ManStart(1000),*r;Aig_Obj_t*e[1000],*o;unsigned long long t; e[0]=Aig_ManConst1(p);for(i=1;i<=5;i++)e[i]=Aig_ObjCreateCi(p);for(i=6;i<6+n;i++){scanf("%d%d",&a,&b);e[i]=Aig_And(p,Aig_NotCond(e[a>>1],a&1),Aig_NotCond(e[b>>1],b&1));}for(i=6;i<6+n;i++)Aig_ObjCreateCo(p,e[i]);r=Csw_Sweep(p,c,k,0);printf("%d",Aig_ManNodeNum(r));Aig_ManForEachCo(r,o,i){t=0;for(q=0;q<32;q++){Aig_ManConst1(r)->iData=1;Aig_ManForEachCi(r,o,j)o->iData=(q>>j)&1;Aig_ManForEachNode(r,o,j)o->iData=(Aig_Regular(o->pFanin0)->iData^Aig_IsComplement(o->pFanin0))&(Aig_Regular(o->pFanin1)->iData^Aig_IsComplement(o->pFanin1));o=Aig_ManCo(r,i);t|=(unsigned long long)(Aig_Regular(o->pFanin0)->iData^Aig_IsComplement(o->pFanin0))<<q;}printf(" %llx",t);}puts("");Aig_ManStop(r);Aig_ManStop(p);}return 0;}`
	if err := os.WriteFile(source, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, "-std=gnu99", "-DLIN64", "-I"+src, "-c", source, "-o", object).CombinedOutput(); err != nil {
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
	r := rand.New(rand.NewSource(73))
	var input strings.Builder
	var graphs []*aig.Graph
	var configs [][2]int32
	for sample := 0; sample < 60; sample++ {
		g := aig.NewGraph()
		edges := []aig.Edge{g.Const1()}
		for i := 0; i < 5; i++ {
			edges = append(edges, g.NewInput())
		}
		k, c := int32(2+sample%5), int32(2+sample%7)
		fmt.Fprintf(&input, "50 %d %d\n", c, k)
		for i := 0; i < 50; i++ {
			a, b := r.Intn(len(edges)*2), r.Intn(len(edges)*2)
			fmt.Fprintf(&input, "%d %d\n", a, b)
			x, y := edges[a>>1], edges[b>>1]
			if a&1 != 0 {
				x = x.Not()
			}
			if b&1 != 0 {
				y = y.Not()
			}
			edges = append(edges, g.AND(x, y))
		}
		for _, e := range edges[6:] {
			g.AddOutput(e)
		}
		graphs = append(graphs, g)
		configs = append(configs, [2]int32{c, k})
	}
	cmd := exec.Command(exe)
	cmd.Stdin = strings.NewReader(input.String())
	data, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("oracle: %v\n%s", err, data)
	}
	reader := strings.NewReader(string(data))
	for sample, g := range graphs {
		out := Csw_Sweep(g, configs[sample][0], configs[sample][1], 0)
		var nodes int32
		if _, err := fmt.Fscan(reader, &nodes); err != nil {
			t.Fatal(err)
		}
		if nodes != aig.Aig_ManNodeNum(out) {
			t.Fatalf("sample%d nodes Go%d C%d", sample, aig.Aig_ManNodeNum(out), nodes)
		}
		for i := 0; i < 50; i++ {
			var want uint64
			if _, err := fmt.Fscanf(reader, "%x", &want); err != nil {
				t.Fatal(err)
			}
			var got uint64
			for q := 0; q < 32; q++ {
				in := make([]bool, 5)
				for j := range in {
					in[j] = q&(1<<j) != 0
				}
				values, err := out.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				if values[i] {
					got |= 1 << q
				}
			}
			if got != want {
				t.Fatalf("sample%d output%d truth %x/%x", sample, i, got, want)
			}
		}
	}
}
