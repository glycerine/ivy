package abc

import (
	"fmt"

	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Link the actual ABC objects: this exercises the complete original cut manager,
// allocation, enumeration, truth manipulation, hash lookup, and cleanup.
func TestNetworkOriginalC(t *testing.T) {
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
	if _, err := os.Stat(filepath.Join(src, "base/abc/abcNtk.o")); err != nil {
		t.Skip("original ABC object files unavailable")
	}
	dir := t.TempDir()
	source, object, exe := filepath.Join(dir, "oracle.c"), filepath.Join(dir, "oracle.o"), filepath.Join(dir, "oracle")
	code := `#include "base/abc/abc.h"
#include "base/main/main.h"
extern Abc_Ntk_t*Abc_NtkDouble(Abc_Ntk_t*);
extern Abc_Ntk_t*Abc_NtkBottom(Abc_Ntk_t*,int);
extern Abc_Ntk_t*Abc_NtkCreateFromNode(Abc_Ntk_t*,Abc_Obj_t*);
extern void Abc_NtkMakeSeq(Abc_Ntk_t*,int);
Abc_Ntk_t*fixture(void){Vec_Ptr_t*v=Vec_PtrAlloc(2);Abc_Ntk_t*n;Vec_PtrPush(v,"11 1\n");Vec_PtrPush(v,"1- 1\n-1 1\n");n=Abc_NtkCreateFromSops("sample",v);Vec_PtrFree(v);return n;}
void dump(Abc_Ntk_t*n){Abc_Obj_t*o;Vec_Ptr_t*v=Abc_NtkDfs(n,0);int x,i,j;unsigned long long truth;printf("RESULT %d %d %d",Abc_NtkCiNum(n),Abc_NtkCoNum(n),Abc_NtkNodeNum(n));Abc_NtkForEachCo(n,o,i){truth=0;for(x=0;x<(1<<Abc_NtkCiNum(n));x++){Abc_NtkForEachCi(n,o,j)o->iTemp=(x>>j)&1;Vec_PtrForEachEntry(Abc_Obj_t*,v,o,j){char*s=(char*)o->pData;int vars=Abc_SopGetVarNum(s),k,t=0;for(;*s;s+=vars+3){int a=1;for(k=0;k<vars;k++)if(s[k]!='-')a&=Abc_ObjFanin(o,k)->iTemp==(s[k]=='1');t|=a;}if(!Abc_SopGetPhase((char*)o->pData))t^=1;o->iTemp=t;}o=Abc_NtkCo(n,i);truth|=(unsigned long long)Abc_ObjFanin0(o)->iTemp<<x;}printf(" %llx",truth);}puts("");Vec_PtrFree(v);}
int main(void){Abc_Ntk_t*n,*p;Abc_Obj_t*node;Abc_Start();n=fixture();p=Abc_NtkDup(n);dump(p);Abc_NtkDelete(p);p=Abc_NtkDupDfs(n);dump(p);Abc_NtkDelete(p);node=Abc_ObjFanin0(Abc_NtkPo(n,0));p=Abc_NtkCreateCone(n,node,"and",0);dump(p);Abc_NtkDelete(p);p=Abc_NtkCreateMffc(n,node,"and");dump(p);Abc_NtkDelete(p);p=Abc_NtkDouble(n);dump(p);Abc_NtkDelete(p);p=Abc_NtkBottom(n,1);dump(p);Abc_NtkDelete(p);p=Abc_NtkCreateFromNode(n,node);dump(p);Abc_NtkDelete(p);p=Abc_NtkCreateWithNode("11 1\n");dump(p);Abc_NtkDelete(p);Abc_NtkMakeSeq(n,1);Abc_NtkMakeComb(n,0);dump(n);Abc_NtkDelete(n);Abc_Stop();return 0;}
`
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

	command := exec.Command(exe)
	data, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("oracle %v\n%s", err, data)
	}
	n := ntkFixture()
	node := n.kindOrder[ObjectNode][0]
	graphs := []*Network{Abc_NtkDup(n), Abc_NtkDupDfs(n), Abc_NtkCreateCone(n, node, "and", 0), Abc_NtkCreateMffc(n, node, "and"), Abc_NtkDouble(n), Abc_NtkBottom(n, 1), Abc_NtkCreateFromNode(n, node), Abc_NtkCreateWithNode("11 1\n")}
	Abc_NtkMakeSeq(n, 1)
	Abc_NtkMakeComb(n, 0)
	graphs = append(graphs, n)
	sample := 0
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "RESULT ") {
			continue
		}
		if sample >= len(graphs) {
			t.Fatal("extra oracle output")
		}
		g := graphs[sample]
		reader := strings.NewReader(strings.TrimPrefix(line, "RESULT "))
		var inputs, outputs, nodes int
		if _, err := fmt.Fscan(reader, &inputs, &outputs, &nodes); err != nil {
			t.Fatal(err)
		}
		if inputs != len(g.cis) || outputs != len(g.cos) || nodes != len(g.kindOrder[ObjectNode]) {
			t.Fatalf("sample%d shape Go%d/%d/%d C%d/%d/%d", sample, len(g.cis), len(g.cos), len(g.kindOrder[ObjectNode]), inputs, outputs, nodes)
		}
		for i := 0; i < outputs; i++ {
			var want uint64
			if _, err := fmt.Fscanf(reader, "%x", &want); err != nil {
				t.Fatal(err)
			}
			var got uint64
			for x := 0; x < 1<<inputs; x++ {
				values := make(map[*Object]bool)
				for j, pi := range g.cis {
					values[pi] = x&(1<<j) != 0
				}
				value, err := g.Evaluate(g.cos[i].Signal(), values)
				if err != nil {
					t.Fatal(err)
				}
				if value {
					got |= 1 << x
				}
			}
			if got != want {
				t.Fatalf("sample%d output%d Go%x C%x", sample, i, got, want)
			}
		}
		sample++
	}
	if sample != len(graphs) {
		t.Fatalf("missing oracle samples %d/%d: %s", sample, len(graphs), data)
	}
}
