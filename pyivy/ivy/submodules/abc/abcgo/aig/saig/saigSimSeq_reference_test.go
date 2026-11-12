package saig

import (
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRandomSimulationAgainstOriginalC(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	source, err := os.ReadFile("../../../src/aig/saig/saigSimSeq.c")
	if err != nil {
		t.Skip("original sequential source unavailable")
	}
	random, err := os.ReadFile("../../../src/aig/aig/aigUtil.c")
	if err != nil {
		t.Fatal(err)
	}
	harness, err := os.ReadFile("test_vectors/seq_reference.c")
	if err != nil {
		t.Fatal(err)
	}
	var program strings.Builder
	program.WriteString(`#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#define ABC_REALLOC(T,P,N) ((T*)realloc((P),sizeof(T)*(N)))
#define NUMBER1 3716960521u
#define NUMBER2 2174103536u
static int Aig_WordFindFirstBit(unsigned x){return __builtin_ctz(x);}
typedef struct {int nSize;int *pArray;} Vec_Int_t;
static int Vec_IntEntry(Vec_Int_t *v,int i){return v->pArray[i];}
static int Vec_IntSize(Vec_Int_t *v){return v->nSize;}
#define Vec_IntForEachEntry(V,E,I) for((I)=0;(I)<(V)->nSize&&(((E)=(V)->pArray[I]),1);(I)++)
typedef struct {void *pAig;int nWords,nPis,nPos,nCis,nCos,nNodes,nObjs;int *pFans0,*pFans1;Vec_Int_t *vCis2Ids,*vLos,*vLis;int *pRefs;unsigned *pSims,*pMems;int nWordsAlloc,nMems,nMemsMax,MemFree;} Raig_Man_t;
`)
	program.WriteString(extractSimulationC(t, string(random), "Aig_ManRandom"))
	for _, name := range []string{"Raig_Lit2Var", "Raig_LitIsCompl", "Raig_ManSimRef", "Raig_ManSimDeref", "Raig_ManSimulateRound"} {
		program.WriteString(extractSimulationC(t, string(source), name))
	}
	program.Write(harness)
	dir := t.TempDir()
	file, binary := filepath.Join(dir, "seq.c"), filepath.Join(dir, "seq")
	if err = os.WriteFile(file, []byte(program.String()), 0600); err != nil {
		t.Fatal(err)
	}
	if log, err := exec.Command(cc, "-std=c99", "-O2", file, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile original simulator: %v\n%s", err, log)
	}
	for _, miter := range []bool{false, true} {
		graph := aig.NewGraph()
		x, state := graph.NewInput(), graph.NewInput()
		output := graph.Const0()
		if miter {
			output = graph.AND(x, state)
		}
		graph.AddOutput(output)
		graph.AddOutput(state.Not())
		aig.Aig_ManSetRegNum(graph, 1)
		p := Raig_ManCreate(graph)
		p.nWords = 2
		var input bytes.Buffer
		fmt.Fprintf(&input, "%d %d %d %d %d %d %d 8 %d\n", p.nWords, len(p.objects), 1, 1, vec.Vec_IntSize(p.vCis2Ids), len(p.outputs), aig.Aig_ManNodeNum(graph), len(p.los))
		for _, obj := range p.objects {
			fmt.Fprintf(&input, "%d %d %d\n", obj.fanins[0], obj.fanins[1], obj.refs)
		}
		for _, id := range vec.Vec_IntArray(p.vCis2Ids) {
			fmt.Fprintf(&input, "%d ", id)
		}
		fmt.Fprintln(&input)
		for i, id := range p.los {
			fmt.Fprintf(&input, "%d %d ", id, p.lis[i])
		}
		command := exec.Command(binary)
		command.Stdin = &input
		outputC, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("original simulation: %v %s", err, outputC)
		}
		var native strings.Builder
		aig.Aig_ManRandom(1)
		for frame := 0; frame < 8; frame++ {
			pattern := int32(-1)
			first := int32(0)
			if frame == 0 {
				first = 1
			}
			failed := Raig_ManSimulateRound(p, 1, first, &pattern)
			fmt.Fprintf(&native, "%d %d ", failed, pattern)
			if failed != 0 {
				break
			}
			for _, id := range p.lis {
				for _, word := range p.blocks[p.slots[id]][1:] {
					fmt.Fprintf(&native, "%d ", word)
				}
			}
		}
		if strings.TrimSpace(native.String()) != strings.TrimSpace(string(outputC)) {
			t.Fatalf("miter=%v native=%s original=%s", miter, native.String(), outputC)
		}
	}
}
