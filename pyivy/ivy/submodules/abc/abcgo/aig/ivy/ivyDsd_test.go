package ivy

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestDsdExhaustiveFourVariableTruth(t *testing.T) {
	var tree []int32
	recognized := 0
	for function := uint32(0); function < 65536; function++ {
		truth := function | function<<16
		if Ivy_TruthDsd(truth, &tree) != 0 {
			recognized++
			if got := Ivy_TruthDsdCompute(tree); got != truth {
				t.Fatalf("function %08x computed %08x", truth, got)
			}
		}
	}
	if recognized < 1000 || recognized >= 65536 {
		t.Fatalf("unexpected decomposable count %d", recognized)
	}
}
func TestDsdConstructEveryRecognizedThreeVariableFunction(t *testing.T) {
	var tree []int32
	for function := uint32(0); function < 256; function++ {
		truth := function * 0x01010101
		if Ivy_TruthDsd(truth, &tree) == 0 {
			continue
		}
		p := Ivy_ManStart()
		front := make([]int32, 5)
		for i := range front {
			front[i] = Ivy_LeafCreate(Ivy_ObjCreatePi(p).Id, 0)
		}
		root := Ivy_ManDsdConstruct(p, front, tree)
		Ivy_ObjCreatePo(p, root)
		for assignment := 0; assignment < 32; assignment++ {
			if ivyBalanceEvaluate(p, assignment)[0] != (truth&(1<<uint(assignment)) != 0) {
				t.Fatalf("constructed truth differs: %08x assignment %d", truth, assignment)
			}
		}
		if Ivy_ManCheck(p) == 0 {
			t.Fatal("constructed graph invalid")
		}
		Ivy_ManStop(p)
	}
}
func TestDsdPackingAndInvalidTree(t *testing.T) {
	n := Ivy_Dec_t{Type: IVY_DEC_AND, fCompl: 1, nFans: 6, fans: [6]uint32{0, 2, 4, 6, 8, 14}}
	if got := Ivy_IntToDec(Ivy_DecToInt(n)); got != n {
		t.Fatal("packed record roundtrip failed")
	}
	tree := []int32{0, 0, 0, 0, 0}
	bad := Ivy_Dec_t{Type: IVY_DEC_AND, nFans: 2, fans: [6]uint32{10, 0}}
	tree = append(tree, Ivy_DecToInt(bad))
	defer func() {
		if recover() == nil {
			t.Error("cyclic packed tree accepted")
		}
	}()
	Ivy_TruthDsdCompute(tree)
}
func TestDsdPrintingAndFiveVariableRandom(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 9))
	var tree []int32
	for i := 0; i < 10000; i++ {
		truth := rng.Uint32()
		if Ivy_TruthDsd(truth, &tree) != 0 && Ivy_TruthDsdCompute(tree) != truth {
			t.Fatal("five-variable truth mismatch")
		}
	}
	Ivy_TruthDsd(0xe8e8e8e8, &tree)
	var out bytes.Buffer
	Ivy_TruthDsdPrint(&out, tree)
	if got := out.String(); got != "F = MAJ(a,b,c)\n" {
		t.Fatalf("majority print %q", got)
	}
}

// Compile the actual C decomposition implementation with a tiny vector shim;
// compare both success decisions and the exact original packed tree records.
func TestDsdSourceCDifferential(t *testing.T) {
	compiler, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	source, err := os.ReadFile(filepath.Join("..", "..", "..", "src", "aig", "ivy", "ivyDsd.c"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	begin := strings.Index(text, "// decomposition types")
	end := strings.Index(text, "void Ivy_TruthDsdPrint_rec")
	if begin < 0 || end < begin {
		t.Fatal("cannot locate source decomposition section")
	}
	preamble := `#include <stdio.h>
#include <stdlib.h>
#include <assert.h>
typedef struct {int nSize,nCap;int *pArray;} Vec_Int_t;
static void Vec_IntClear(Vec_Int_t *v){v->nSize=0;}
static void Vec_IntPush(Vec_Int_t *v,int x){if(v->nSize==v->nCap){v->nCap=v->nCap?2*v->nCap:16;v->pArray=realloc(v->pArray,4*v->nCap);}v->pArray[v->nSize++]=x;}
static int Vec_IntSize(Vec_Int_t *v){return v->nSize;}
static int Vec_IntEntry(Vec_Int_t *v,int i){return v->pArray[i];}
static int Vec_IntPop(Vec_Int_t *v){return v->pArray[--v->nSize];}
unsigned Ivy_TruthDsdCompute(Vec_Int_t *);
`
	harness := `int main(void){unsigned truth;Vec_Int_t tree={0};while(scanf("%x",&truth)==1){int ok=Ivy_TruthDsd(truth,&tree);printf("%d",ok);if(ok){printf(" %d",tree.nSize);for(int i=0;i<tree.nSize;i++)printf(" %08x",(unsigned)tree.pArray[i]);}puts("");}free(tree.pArray);return 0;}`
	dir := t.TempDir()
	cfile := filepath.Join(dir, "reference.c")
	binary := filepath.Join(dir, "reference")
	if err := os.WriteFile(cfile, []byte(preamble+text[begin:end]+harness), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(compiler, "-std=c99", "-O2", cfile, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile reference: %v\n%s", err, out)
	}
	truths := make([]uint32, 0, 70000)
	for f := uint32(0); f < 65536; f++ {
		truths = append(truths, f|f<<16)
	}
	rng := rand.New(rand.NewPCG(23, 71))
	for i := 0; i < 4096; i++ {
		truths = append(truths, rng.Uint32())
	}
	var input strings.Builder
	for _, truth := range truths {
		fmt.Fprintf(&input, "%08x\n", truth)
	}
	cmd := exec.Command(binary)
	cmd.Stdin = strings.NewReader(input.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run reference: %v\n%s", err, output)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) != len(truths) {
		t.Fatalf("reference lines %d want %d", len(lines), len(truths))
	}
	var tree []int32
	for i, truth := range truths {
		fields := strings.Fields(lines[i])
		success := Ivy_TruthDsd(truth, &tree)
		if len(fields) == 0 || fields[0] != strconv.Itoa(int(success)) {
			t.Fatalf("decision differs at %08x: %q Go=%d", truth, lines[i], success)
		}
		if success == 0 {
			continue
		}
		if len(fields) != len(tree)+2 {
			t.Fatalf("tree size differs at %08x: C %s Go %v", truth, lines[i], tree)
		}
		for k, word := range tree {
			value, err := strconv.ParseUint(fields[k+2], 16, 32)
			if err != nil || uint32(word) != uint32(value) {
				t.Fatalf("packed tree differs at %08x node %d: Go=%08x C=%s", truth, k, uint32(word), fields[k+2])
			}
		}
	}
}
