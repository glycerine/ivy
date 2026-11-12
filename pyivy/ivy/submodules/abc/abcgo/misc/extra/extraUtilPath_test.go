package extra

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestPathOriginalFrontierTransitions(t *testing.T) {
	hash := sha256.New()
	wantCounts := []float64{2, 12, 184, 8512, 1262816}
	for n := int32(2); n <= 6; n++ {
		edges := Abc_GraphGrid(n)
		life := Abc_GraphNodeLife(edges, n)
		fronts := Abc_GraphFrontiers(edges, life)
		nodes, _ := graphBuildFrontier(edges, life, fronts, 0)
		for level := 0; level < vec.Vec_WecSize(nodes); level++ {
			data := vec.Vec_IntArray(vec.Vec_WecEntry(nodes, level))
			var b [4]byte
			binary.LittleEndian.PutUint32(b[:], uint32(len(data)))
			hash.Write(b[:])
			for _, v := range data {
				binary.LittleEndian.PutUint32(b[:], uint32(v))
				hash.Write(b[:])
			}
		}
		count := Abc_GraphCountPaths(nodes, edges)
		if count != wantCounts[n-2] {
			t.Fatalf("grid%d count%.0f", n, count)
		}
		var b [8]byte
		binary.LittleEndian.PutUint64(b[:], math.Float64bits(count))
		hash.Write(b[:])
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "a30dd81a6fd611a2ff8394fc8f07a69a03389b04fea768d80a44ba1fd0fe623a" {
		t.Fatalf("original C transition mismatch %s", got)
	}
}
func simpleGridPath(edges []int32, chosen []bool, n int) bool {
	adjacency := make([][]int, n*n)
	used := 0
	for e, take := range chosen {
		if !take {
			continue
		}
		a, b := int(edges[2*e]), int(edges[2*e+1])
		adjacency[a] = append(adjacency[a], b)
		adjacency[b] = append(adjacency[b], a)
		used++
	}
	if len(adjacency[0]) != 1 || len(adjacency[n*n-1]) != 1 {
		return false
	}
	for node, row := range adjacency {
		if node != 0 && node != n*n-1 && len(row) != 0 && len(row) != 2 {
			return false
		}
	}
	seen := make([]bool, n*n)
	previous, current, walked := -1, 0, 0
	for current != n*n-1 {
		if seen[current] {
			return false
		}
		seen[current] = true
		next := -1
		for _, v := range adjacency[current] {
			if v != previous {
				next = v
				break
			}
		}
		if next < 0 {
			return false
		}
		previous, current = current, next
		walked++
	}
	return walked == used
}
func TestPathAigAgainstBruteForce(t *testing.T) {
	for n := int32(2); n <= 3; n++ {
		edges := Abc_GraphGrid(n)
		life := Abc_GraphNodeLife(edges, n)
		fronts := Abc_GraphFrontiers(edges, life)
		nodes, _ := graphBuildFrontier(edges, life, fronts, 0)
		graph := Abc_GraphDeriveGia(nodes, edges)
		for assignment := 0; assignment < 1<<graph.InputCount(); assignment++ {
			inputs := make([]bool, graph.InputCount())
			for k := range inputs {
				inputs[k] = assignment>>k&1 != 0
			}
			out, err := graph.Evaluate(inputs)
			if err != nil {
				t.Fatal(err)
			}
			if out[0] != simpleGridPath(vec.Vec_IntArray(edges), inputs, int(n)) {
				t.Fatalf("grid%d assignment%x", n, assignment)
			}
		}
		if _, err := graph.AigerModel(); err != nil {
			t.Fatal(err)
		}
	}
}
func TestPathDirectedReachability(t *testing.T) {
	for size := int32(1); size <= 2; size++ {
		graph := Abc_EnumeratePaths(size)
		for assignment := 0; assignment < 1<<graph.InputCount(); assignment++ {
			inputs := make([]bool, graph.InputCount())
			for k := range inputs {
				inputs[k] = assignment>>k&1 != 0
			}
			reachable := make([]bool, (size+1)*(size+1))
			reachable[0] = true
			for y := int32(0); y <= size; y++ {
				for x := int32(0); x <= size; x++ {
					at := y*(size+1) + x
					if x > 0 {
						reachable[at] = reachable[at] || reachable[at-1] && inputs[Abc_NodeVarX(size, y, x)/2-1]
					}
					if y > 0 {
						reachable[at] = reachable[at] || reachable[at-size-1] && inputs[Abc_NodeVarY(size, y, x)/2-1]
					}
				}
			}
			out, err := graph.Evaluate(inputs)
			if err != nil || out[0] != reachable[len(reachable)-1] {
				t.Fatalf("directed size%d assignment%x", size, assignment)
			}
		}
	}
}
func TestPathLexicographicSamples(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c := graph.NewInput(), graph.NewInput(), graph.NewInput()
	graph.AddOutput(graph.OR(graph.AND(a, b), c))
	result := graphSolvePaths(graph, 1000)
	// Ordered assignments are 001,011,101,110,111.
	if result.Samples != 5 || result.First != 1 || result.Last != 7 || result.Average != 1.5 {
		t.Fatalf("lex samples %+v", result)
	}
	result = graphSolvePaths(graph, 3)
	if result.Samples != 3 || result.First != 1 || result.Last != 5 || result.Average != 2 {
		t.Fatalf("bounded samples %+v", result)
	}
	impossible := aig.NewGraph()
	impossible.NewInput()
	impossible.AddOutput(impossible.Const0())
	if result = graphSolvePaths(impossible, 10); result.Samples != 0 || result.Estimate != 0 {
		t.Fatalf("unsatisfiable sample %+v", result)
	}
	only := aig.NewGraph()
	x := only.NewInput()
	only.AddOutput(x)
	if result = graphSolvePaths(only, 10); result.Samples != 1 || result.First != 1 || result.Average != 0 {
		t.Fatalf("single sample %+v", result)
	}
}
