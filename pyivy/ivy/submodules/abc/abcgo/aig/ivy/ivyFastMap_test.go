package ivy

import (
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func ivyCheckMappedSupports(t *testing.T, p *Ivy_Man_t, limit int32) {
	t.Helper()
	for _, n := range ivyFastMapOrder(p) {
		support := Ivy_ObjSupp(p, n)
		if len(support.pArray) < 2 || len(support.pArray) > int(limit) || !slices.IsSorted(support.pArray) {
			t.Fatalf("invalid mapped support at %d: %v", n.Id, support.pArray)
		}
		cut := map[int32]bool{}
		for _, id := range support.pArray {
			if cut[id] {
				t.Fatal("duplicate LUT input")
			}
			cut[id] = true
		}
		seen := map[*Ivy_Obj_t]bool{}
		var covered func(*Ivy_Obj_t) bool
		covered = func(current *Ivy_Obj_t) bool {
			if cut[current.Id] || seen[current] {
				return true
			}
			if Ivy_ObjIsCi(current) != 0 || Ivy_ObjIsConst1(current) != 0 {
				return false
			}
			seen[current] = true
			return covered(Ivy_ObjFanin0(current)) && covered(Ivy_ObjFanin1(current))
		}
		if !covered(n) {
			t.Fatalf("LUT support does not separate node %d", n.Id)
		}
		if got := Ivy_FastMapNodeDelay(p, n); got != support.Delay {
			t.Fatalf("node %d delay %d want %d", n.Id, support.Delay, got)
		}
		if got := Ivy_TableLookup(p, n); got != n {
			t.Fatal("mapping changed canonical lookup")
		}
	}
}
func TestFastMapTreeDepthAreaAndRecovery(t *testing.T) {
	for limit := int32(2); limit <= 8; limit++ {
		p := Ivy_ManStart()
		var inputs []*Ivy_Obj_t
		for i := 0; i < 8; i++ {
			inputs = append(inputs, Ivy_ObjCreatePi(p))
		}
		root := inputs[0]
		for _, input := range inputs[1:] {
			root = Ivy_And(p, root, input)
		}
		Ivy_ObjCreatePo(p, root)
		Ivy_ObjCreatePo(p, Ivy_Not(root))
		Ivy_FastMapPerform(p, limit, 0, 0)
		beforeDelay, beforeArea := Ivy_FastMapDelay(p), Ivy_FastMapArea(p)
		ivyCheckMappedSupports(t, p, limit)
		if limit == 8 && (beforeDelay != 1 || beforeArea != 1) {
			t.Fatalf("8-LUT skew AND maps to delay=%d area=%d", beforeDelay, beforeArea)
		}
		Ivy_FastMapPerform(p, limit, 1, 0)
		ivyCheckMappedSupports(t, p, limit)
		if Ivy_FastMapDelay(p) > beforeDelay || Ivy_FastMapArea(p) > beforeArea {
			t.Fatal("recovery worsened delay or area")
		}
		if Ivy_ManCheck(p) == 0 {
			t.Fatal("mapping changed AIG invariants")
		}
		Ivy_FastMapStop(p)
		Ivy_FastMapStop(p)
		if p.pData != nil {
			t.Fatal("mapping memory not released")
		}
	}
}
func TestFastMapRandomGraphReferencesAndCanonicalLookup(t *testing.T) {
	rng := rand.New(rand.NewPCG(2, 37))
	for trial := 0; trial < 100; trial++ {
		p := Ivy_ManStart()
		var pool []*Ivy_Obj_t
		for i := 0; i < 6; i++ {
			pool = append(pool, Ivy_ObjCreatePi(p))
		}
		for i := 0; i < 50; i++ {
			a, b := pool[rng.IntN(len(pool))], pool[rng.IntN(len(pool))]
			a = Ivy_NotCond(a, int32(rng.IntN(2)))
			b = Ivy_NotCond(b, int32(rng.IntN(2)))
			pool = append(pool, Ivy_And(p, a, b))
		}
		for i := 0; i < 5; i++ {
			Ivy_ObjCreatePo(p, pool[len(pool)-1-i])
		}
		Ivy_ManCleanup(p)
		limit := int32(2 + rng.IntN(5))
		Ivy_FastMapPerform(p, limit, 1, 0)
		ivyCheckMappedSupports(t, p, limit)
		delay := Ivy_FastMapDelay(p)
		Ivy_FastMapRequired(p, delay, 1)
		manager := ivySupportManager(p)
		expected := make([]int32, len(manager.supports))
		for _, out := range p.vPos {
			expected[Ivy_ObjFaninId0(out)]++
		}
		for _, level := range manager.vLuts {
			for _, n := range level {
				for _, id := range Ivy_ObjSupp(p, n).pArray {
					expected[id]++
				}
			}
		}
		for id, reference := range expected {
			if manager.supports[id].nRefs != reference {
				t.Fatalf("trial %d object %d refs %d want %d", trial, id, manager.supports[id].nRefs, reference)
			}
		}
		for _, level := range manager.vLuts {
			for _, n := range level {
				before := make([]int32, len(manager.supports))
				for id := range before {
					before[id] = manager.supports[id].nRefs
				}
				Ivy_FastMapNodeAreaRefed(p, n)
				for id, reference := range before {
					if manager.supports[id].nRefs != reference {
						t.Fatal("area query changed references")
					}
				}
			}
		}
	}
}
func TestFastMapMergeRandomUnionAndFailedAliasing(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 11))
	for trial := 0; trial < 10000; trial++ {
		a, b := Ivy_Supp_t{}, Ivy_Supp_t{}
		set := map[int32]bool{}
		for id := int32(0); id < 20; id++ {
			if rng.IntN(4) == 0 {
				a.pArray = append(a.pArray, id)
				set[id] = true
			}
			if rng.IntN(4) == 0 {
				b.pArray = append(b.pArray, id)
				set[id] = true
			}
		}
		var union []int32
		for id := range set {
			union = append(union, id)
		}
		slices.Sort(union)
		limit := int32(2 + rng.IntN(15))
		before := slices.Clone(a.pArray)
		ok := Ivy_FastMapMerge(&a, &b, &a, limit)
		if len(union) > int(limit) {
			if ok != 0 || !slices.Equal(a.pArray, before) {
				t.Fatal("failed aliased merge changed original support")
			}
		} else if ok == 0 || !slices.Equal(a.pArray, union) {
			t.Fatal("support union differs from reference set")
		}
	}
}
func TestFastMapUsesTopologyAfterObjectReplacement(t *testing.T) {
	p := Ivy_ManStart()
	a, b, c, d := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	old := Ivy_And(p, a, b)
	consumer := Ivy_And(p, old, d)
	Ivy_ObjCreatePo(p, consumer)
	newer := Ivy_And(p, Ivy_And(p, a, c), b)
	Ivy_ObjReplace(p, old, newer, 1, 0, 0)
	Ivy_ManResetLevels(p)
	Ivy_FastMapPerform(p, 3, 1, 0)
	ivyCheckMappedSupports(t, p, 3)
}

func TestFastMapSourceCDelayMapping(t *testing.T) {
	compiler, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "src", "aig", "ivy", "ivyFastMap.c"))
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	extract := func(signature string) string {
		start := strings.Index(source, "\n"+signature)
		if start < 0 {
			t.Fatal("missing source function", signature)
		}
		end := strings.Index(source[start:], "\n}\n")
		if end < 0 {
			t.Fatal("missing source function end")
		}
		return source[start : start+end+3]
	}
	preamble := `#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
typedef struct {int nSize,Delay,fMark;int pArray[32];} Ivy_Supp_t;
typedef struct Ivy_Obj_t_ {int Id,nRefs;struct Ivy_Obj_t_ *a,*b;} Ivy_Obj_t;
typedef struct {Ivy_Obj_t nodes[2048];Ivy_Supp_t supp[2048];} Ivy_Man_t;
#define IVY_MAX(a,b) ((a)>(b)?(a):(b))
#define Ivy_ObjSupp(p,n) (&(p)->supp[(n)->Id])
#define Ivy_ObjFanin0(n) ((n)->a)
#define Ivy_ObjFanin1(n) ((n)->b)
#define Ivy_ObjFaninId0(n) ((n)->a->Id)
#define Ivy_ObjFaninId1(n) ((n)->b->Id)
#define Ivy_ObjId(n) ((n)->Id)
#define Ivy_ObjRefs(n) ((n)->nRefs)
#define Ivy_ManConst1(p) (&(p)->nodes[0])
#define Ivy_ObjIsNode(n) ((n)->a!=NULL)
#define Ivy_ObjIsNodeInt1(n) (Ivy_ObjIsNode(n)&&(n)->nRefs==1)
static void Vec_IntSelectSort(int *a,int size){for(int i=0;i<size;i++)for(int j=i+1;j<size;j++)if(a[i]>a[j]){int x=a[i];a[i]=a[j];a[j]=x;}}
static int Vec_IntRemoveDup(int *a,int size){int k=0;for(int i=0;i<size;i++)if(!k||a[i]!=a[k-1])a[k++]=a[i];return k;}
int Ivy_FastMapMerge(Ivy_Supp_t*,Ivy_Supp_t*,Ivy_Supp_t*,int);
`
	harness := `int main(void){int inputs,count,limit;while(scanf("%d%d%d",&inputs,&count,&limit)==3){Ivy_Man_t p={0};for(int i=0;i<=inputs+count;i++){p.nodes[i].Id=i;if(i<=inputs){p.supp[i].nSize=1;p.supp[i].pArray[0]=i;}}for(int i=inputs+1;i<=inputs+count;i++){int a,b;if(scanf("%d%d",&a,&b)!=2)return 2;p.nodes[i].a=&p.nodes[a];p.nodes[i].b=&p.nodes[b];Ivy_FastMapNode(&p,&p.nodes[i],limit);printf("%d %d",p.supp[i].Delay,p.supp[i].nSize);for(int k=0;k<p.supp[i].nSize;k++)printf(" %d",p.supp[i].pArray[k]);puts("");}}return 0;}`
	dir := t.TempDir()
	cfile, binary := filepath.Join(dir, "reference.c"), filepath.Join(dir, "reference")
	if err := os.WriteFile(cfile, []byte(preamble+extract("void Ivy_FastMapNode( Ivy_")+extract("int Ivy_FastMapMerge( Ivy_")+harness), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(compiler, "-std=c99", "-O2", cfile, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("reference compile: %v\n%s", err, output)
	}
	rng := rand.New(rand.NewPCG(5, 109))
	var input strings.Builder
	var expected [][]int32
	for trial := 0; trial < 200; trial++ {
		p := Ivy_ManStart()
		pool := make([]*Ivy_Obj_t, 6)
		for i := range pool {
			pool[i] = Ivy_ObjCreatePi(p)
		}
		for i := 0; i < 100; i++ {
			pool = append(pool, Ivy_And(p, pool[rng.IntN(len(pool))], pool[rng.IntN(len(pool))]))
		}
		limit := int32(2 + rng.IntN(7))
		fmt.Fprintf(&input, "6 %d %d\n", len(p.vObjs)-7, limit)
		for _, n := range p.vObjs[7:] {
			fmt.Fprintf(&input, "%d %d\n", Ivy_ObjFaninId0(n), Ivy_ObjFaninId1(n))
		}
		Ivy_FastMapPerform(p, limit, 0, 0)
		for _, n := range p.vObjs[7:] {
			s := Ivy_ObjSupp(p, n)
			row := []int32{s.Delay, int32(len(s.pArray))}
			expected = append(expected, append(row, s.pArray...))
		}
	}
	cmd := exec.Command(binary)
	cmd.Stdin = strings.NewReader(input.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("reference run: %v\n%s", err, output)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) != len(expected) {
		t.Fatalf("reference rows %d want %d", len(lines), len(expected))
	}
	for i, row := range expected {
		fields := strings.Fields(lines[i])
		if len(fields) != len(row) {
			t.Fatalf("row %d differs: C %s Go %v", i, lines[i], row)
		}
		for k, want := range row {
			got, err := strconv.ParseInt(fields[k], 10, 32)
			if err != nil || int32(got) != want {
				t.Fatalf("row %d differs: C %s Go %v", i, lines[i], row)
			}
		}
	}
}
