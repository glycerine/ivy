package saig

import (
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func simulationFixture() *aig.Graph {
	graph := aig.NewGraph()
	x, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(graph.AND(x, state))
	graph.AddOutput(state.Not())
	aig.Aig_ManSetRegNum(graph, 1)
	return graph
}
func TestFastSimulationProbabilities(t *testing.T) {
	graph := simulationFixture()
	probabilities, err := SwitchingProbabilities(graph, 101, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	lo := graph.Inputs()[1].NodeID()
	if probabilities[lo] < 0.99 || probabilities[lo] > 1 {
		t.Fatalf("toggle activity %g", probabilities[lo])
	}
	for _, p := range probabilities {
		if p < 0 || p > 1 || math.IsNaN(float64(p)) {
			t.Fatalf("invalid probability %g", p)
		}
	}
	one, err := SwitchingProbabilities(graph, 100, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	if one[lo] != 0.5 {
		t.Fatalf("toggle probability %g", one[lo])
	}
}
func TestFastSimulationSparseIDs(t *testing.T) {
	graph := aig.NewGraph()
	x, y := graph.NewInput(), graph.NewInput()
	dangling := graph.AND(x, y)
	aig.Aig_ObjDelete(graph, dangling)
	graph.AddOutput(graph.NewXor(x, y))
	p := Faig_ManCreate(graph)
	aig.Aig_ManRandom(1)
	counts := Faig_ManSimulateFrames(p, 16, 0, 0)
	if len(counts) != len(p.ids) || p.ids[len(p.ids)-1] != graph.Outputs()[0] {
		t.Fatal("sparse ID mapping lost output")
	}
	probabilities, err := SwitchingProbabilities(graph, 16, 0, true)
	if err != nil || len(probabilities) != int(aig.Aig_ManObjNumMax(graph)) {
		t.Fatalf("sparse graph: %v", err)
	}
	if probabilities[dangling.NodeID()] != 0 {
		t.Fatal("deleted object acquired probability")
	}
}
func extractSimulationC(t *testing.T, source, name string) string {
	t.Helper()
	pattern := regexp.MustCompile(`(?m)^(?:static\s+inline\s+)?(?:unsigned|int|float|void)\s*\*?\s*` + regexp.QuoteMeta(name) + `\s*\(`)
	loc := pattern.FindStringIndex(source)
	if loc == nil {
		t.Fatalf("missing original C function %s", name)
	}
	start := loc[0]
	open := strings.IndexByte(source[loc[1]:], '{') + loc[1]
	depth, end := 1, open+1
	for depth > 0 && end < len(source) {
		switch source[end] {
		case '{':
			depth++
		case '}':
			depth--
		}
		end++
	}
	return source[start:end] + "\n"
}
func TestFastSimulationAgainstOriginalC(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	source, err := os.ReadFile("../../../src/aig/saig/saigSimFast.c")
	if err != nil {
		t.Skip("original C source unavailable")
	}
	random, err := os.ReadFile("../../../src/aig/aig/aigUtil.c")
	if err != nil {
		t.Fatal(err)
	}
	harness, err := os.ReadFile("test_vectors/fast_reference.c")
	if err != nil {
		t.Fatal(err)
	}
	var program strings.Builder
	program.WriteString(`#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#define ABC_ALLOC(T,N) ((T*)malloc(sizeof(T)*(N)))
#define ABC_CALLOC(T,N) ((T*)calloc((N),sizeof(T)))
#define ABC_FREE(P) free(P)
#define NUMBER1 3716960521u
#define NUMBER2 2174103536u
static int Aig_WordCountOnes(unsigned x){return __builtin_popcount(x);}
typedef struct {int nPis,nPos,nCis,nCos,nFfs,nNos,nPis1,nCis1,nCisNos1,nCisNosPos1,nObjs,nWords;int pObjs[];} Faig_Man_t;
`)
	program.WriteString(extractSimulationC(t, string(random), "Aig_ManRandom"))
	for _, name := range []string{"Faig_CheckIdLo", "Faig_CheckIdNo", "Faig_CheckIdCo", "Faig_ObjIdToNumNo", "Faig_ObjIdToNumCo", "Faig_ObjLoToLi", "Faig_NodeChild0", "Faig_NodeChild1", "Faig_CoChild0", "Faig_ObjFaninC", "Faig_ObjFanin", "Faig_SimulateNode", "Faig_SimulateCo", "Faig_SimulateRandomShift", "Faig_SimulateTransferShift", "Faig_ManSimulateFrames"} {
		program.WriteString(extractSimulationC(t, string(source), name))
	}
	program.Write(harness)
	dir := t.TempDir()
	cfile, binary := filepath.Join(dir, "reference.c"), filepath.Join(dir, "reference")
	if err = os.WriteFile(cfile, []byte(program.String()), 0600); err != nil {
		t.Fatal(err)
	}
	if log, err := exec.Command(cc, "-std=c99", "-O2", cfile, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile original C: %v\n%s", err, log)
	}
	p := Faig_ManCreate(simulationFixture())
	for _, transitions := range []int32{0, 1} {
		var input bytes.Buffer
		fmt.Fprintf(&input, "%d %d %d %d 17 4 %d\n", p.nPis, p.nFfs, len(p.nodes), p.nPos, transitions)
		for _, node := range p.nodes {
			fmt.Fprintf(&input, "%d %d ", node.fanins[0], node.fanins[1])
		}
		for _, lit := range p.outputs {
			fmt.Fprintf(&input, "%d ", lit)
		}
		command := exec.Command(binary)
		command.Stdin = &input
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("original C: %v %s", err, output)
		}
		reference := make([]int32, 0)
		for _, word := range strings.Fields(string(output)) {
			value, err := strconv.ParseInt(word, 10, 32)
			if err != nil {
				t.Fatal(err)
			}
			reference = append(reference, int32(value))
		}
		aig.Aig_ManRandom(1)
		actual := Faig_ManSimulateFrames(p, 17, 4, transitions)
		if !reflect.DeepEqual(actual, reference) {
			t.Fatalf("transition=%d native=%v original=%v", transitions, actual, reference)
		}
	}
}
