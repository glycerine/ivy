package cgt

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestClockGatingOriginalCandidateOracle(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section collection required")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("optional C oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	dir := t.TempDir()
	main := filepath.Join(dir, "main.c")
	binary := filepath.Join(dir, "oracle")
	code := `#include "opt/cgt/cgtAig.c"
int main(void){int n,root,level;while(scanf("%d%d%d",&n,&root,&level)==3){Aig_Obj_t*o=calloc(n,sizeof(Aig_Obj_t));int kind,a,b,l,use;Vec_Int_t*v=Vec_IntStart(n);for(int i=0;i<n;i++){scanf("%d%d%d%d%d",&kind,&a,&b,&l,&use);o[i].Id=i;o[i].Level=l;o[i].Type=kind==0?AIG_OBJ_CONST1:kind==1?AIG_OBJ_CI:kind==2?AIG_OBJ_AND:kind==3?AIG_OBJ_EXOR:kind==4?AIG_OBJ_BUF:AIG_OBJ_CO;if(kind>=2){o[i].pFanin0=Aig_NotCond(o+(a>>1),a&1);o[i].pFanin1=Aig_NotCond(o+(b>>1),b&1);}Vec_IntWriteEntry(v,i,use);}Aig_Man_t g={0};g.nTravIds=1;Vec_Ptr_t*result=Vec_PtrAlloc(n);if(Aig_ObjIsNode(o+root))Cgt_ManDetectCandidates_rec(&g,v,o+root,level,result);printf("%d",Vec_PtrSize(result));for(int i=0;i<Vec_PtrSize(result);i++)printf(" %d",((Aig_Obj_t*)Vec_PtrEntry(result,i))->Id);puts("");Vec_PtrFree(result);Vec_IntFree(v);free(o);}return 0;}`
	if err = os.WriteFile(main, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", src, main, "-o", binary)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build source oracle %v\n%s", err, out)
	}
	rng := rand.New(rand.NewSource(668))
	var input bytes.Buffer
	var expected [][]aig.Edge
	for trial := 0; trial < 100; trial++ {
		g := aig.NewGraph()
		pool := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 12; i++ {
			a, b := pool[rng.Intn(len(pool))], pool[rng.Intn(len(pool))]
			if rng.Intn(2) != 0 {
				b = b.Not()
			}
			if i%3 == 0 {
				pool = append(pool, g.NewXor(a, b))
			} else {
				pool = append(pool, g.AND(a, b))
			}
		}
		root := pool[len(pool)-1]
		limit := rng.Intn(8)
		nodes := g.Nodes()
		useful := make([]bool, len(nodes))
		fmt.Fprintf(&input, "%d %d %d\n", len(nodes), root.NodeID(), limit)
		for id, n := range nodes {
			useful[id] = rng.Intn(3) != 0
			kind := 0
			switch n.Kind {
			case aig.Input:
				kind = 1
			case aig.And:
				kind = 2
			case aig.ExclusiveOr:
				kind = 3
			case aig.Buffer:
				kind = 4
			case aig.Output:
				kind = 5
			}
			a, b := n.Fanins[0], n.Fanins[1]
			if kind < 2 {
				a, b = 0, 0
			}
			u := 0
			if useful[id] {
				u = 1
			}
			fmt.Fprintf(&input, "%d %d %d %d %d\n", kind, a, b, n.Level, u)
		}
		expected = append(expected, Cgt_ManDetectCandidates(g, useful, root, limit))
	}
	cmd = exec.Command(binary)
	cmd.Stdin = &input
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run source oracle %v\n%s", err, out)
	}
	scan := bufio.NewScanner(bytes.NewReader(out))
	line := 0
	for scan.Scan() {
		if line >= len(expected) {
			t.Fatal("extra source output")
		}
		fields := strings.Fields(scan.Text())
		count, err := strconv.Atoi(fields[0])
		if err != nil || count != len(expected[line]) {
			t.Fatal("candidate count mismatch", line, scan.Text(), expected[line])
		}
		for i, e := range expected[line] {
			id, err := strconv.Atoi(fields[i+1])
			if err != nil || id != e.NodeID() {
				t.Fatal("candidate traversal order mismatch", line, i, id, e)
			}
		}
		line++
	}
	if scan.Err() != nil || line != len(expected) {
		t.Fatal("truncated source oracle")
	}
}
