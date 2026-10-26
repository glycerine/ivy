package abc

// Go-only tests for src/misc/extra/extraUtilPath.c behavior.

import (
	"errors"
	"math/big"
	"slices"
	"strings"
	"testing"
)

func extraGridBrutePath(g ExtraGridGraph, selected []bool) bool {
	degree := make([]int, g.Nodes)
	adjacency := make([][]int, g.Nodes)
	for index, edge := range g.Edges {
		if !selected[index] {
			continue
		}
		degree[edge.From]++
		degree[edge.To]++
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
		adjacency[edge.To] = append(adjacency[edge.To], edge.From)
	}
	for node, value := range degree {
		if node == 0 || node == g.Nodes-1 {
			if value != 1 {
				return false
			}
		} else if value != 0 && value != 2 {
			return false
		}
	}
	seen := make([]bool, g.Nodes)
	stack := []int{0}
	seen[0] = true
	for len(stack) != 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, next := range adjacency[node] {
			if !seen[next] {
				seen[next] = true
				stack = append(stack, next)
			}
		}
	}
	if !seen[g.Nodes-1] {
		return false
	}
	for node, value := range degree {
		if value != 0 && !seen[node] {
			return false
		}
	}
	return true
}

func extraMaskBits(count int, mask uint64) []bool {
	bits := make([]bool, count)
	for index := range bits {
		bits[index] = mask>>uint(index)&1 != 0
	}
	return bits
}

func TestExtraGridGraphOrderLifeAndFrontiers(t *testing.T) {
	g, err := BuildExtraGridGraph(2)
	if err != nil {
		t.Fatal(err)
	}
	wantEdges := []ExtraGridEdge{{0, 1}, {0, 2}, {1, 3}, {2, 3}}
	if !slices.Equal(g.Edges, wantEdges) {
		t.Fatalf("edges = %v", g.Edges)
	}
	if !slices.Equal(g.Life, []ExtraGridLife{{0, 1}, {0, 2}, {1, 3}, {2, 3}}) {
		t.Fatalf("life = %v", g.Life)
	}
	wantFrontiers := [][]int{nil, {0, 1}, {1, 2}, {2, 3}, nil}
	for index := range wantFrontiers {
		if !slices.Equal(g.Frontiers[index], wantFrontiers[index]) {
			t.Fatalf("frontier %d = %v", index, g.Frontiers[index])
		}
	}
	if _, err = BuildExtraGridGraph(1); !errors.Is(err, ErrDimension) {
		t.Fatalf("size error = %v", err)
	}
}

func TestExtraGridFrontierExactAgainstBruteForce(t *testing.T) {
	for _, row := range []struct {
		size  int
		count int64
	}{{2, 2}, {3, 12}, {4, 184}} {
		frontier, err := BuildExtraGridFrontier(row.size, ExtraGridFrontierOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if frontier.PathCount.Cmp(big.NewInt(row.count)) != 0 {
			t.Fatalf("size %d count = %s", row.size, frontier.PathCount)
		}
		if frontier.TotalStates == 0 || len(frontier.Transitions) != len(frontier.Graph.Edges) {
			t.Fatalf("incomplete result: %+v", frontier)
		}
		if row.size <= 3 {
			brute := int64(0)
			for mask := uint64(0); mask < uint64(1)<<uint(len(frontier.Graph.Edges)); mask++ {
				if extraGridBrutePath(frontier.Graph, extraMaskBits(len(frontier.Graph.Edges), mask)) {
					brute++
				}
			}
			if brute != row.count {
				t.Fatalf("size %d brute = %d", row.size, brute)
			}
		}
	}
}

func TestExtraGridDerivedGIAMatchesStandalonePredicate(t *testing.T) {
	for size := 2; size <= 3; size++ {
		frontier, err := BuildExtraGridFrontier(size, ExtraGridFrontierOptions{})
		if err != nil {
			t.Fatal(err)
		}
		graph, err := frontier.DeriveGIA(1 << 16)
		if err != nil {
			t.Fatal(err)
		}
		for mask := uint64(0); mask < uint64(1)<<uint(len(frontier.Graph.Edges)); mask++ {
			inputs := extraMaskBits(len(frontier.Graph.Edges), mask)
			outputs, evalErr := graph.Evaluate(inputs)
			if evalErr != nil {
				t.Fatal(evalErr)
			}
			if outputs[0] != extraGridBrutePath(frontier.Graph, inputs) {
				t.Fatalf("size %d mask %x = %v", size, mask, outputs[0])
			}
		}
	}
}

func extraMonotoneDirect(size int, enabled []bool) bool {
	nodes := make([]bool, size+1)
	nodes[0] = true
	for x := 1; x <= size; x++ {
		v, _ := ExtraMonotoneHorizontalVariable(size, 0, x)
		nodes[x] = nodes[x-1] && enabled[v]
	}
	for y := 1; y <= size; y++ {
		v, _ := ExtraMonotoneVerticalVariable(size, y, 0)
		nodes[0] = nodes[0] && enabled[v]
		for x := 1; x <= size; x++ {
			h, _ := ExtraMonotoneHorizontalVariable(size, y, x)
			v, _ := ExtraMonotoneVerticalVariable(size, y, x)
			nodes[x] = nodes[x-1] && enabled[h] || nodes[x] && enabled[v]
		}
	}
	return nodes[size]
}

func TestExtraMonotonePathGIAExhaustive(t *testing.T) {
	for size := 1; size <= 2; size++ {
		graph, err := BuildExtraMonotonePathsGIA(size, 1<<16)
		if err != nil {
			t.Fatal(err)
		}
		variables := 2 * size * (size + 1)
		for mask := uint64(0); mask < uint64(1)<<uint(variables); mask++ {
			inputs := extraMaskBits(variables, mask)
			outputs, evalErr := graph.Evaluate(inputs)
			if evalErr != nil {
				t.Fatal(evalErr)
			}
			if outputs[0] != extraMonotoneDirect(size, inputs) {
				t.Fatalf("size %d mask %x = %v", size, mask, outputs[0])
			}
		}
	}
	if h, err := ExtraMonotoneHorizontalVariable(2, 0, 1); err != nil || h != 0 {
		t.Fatalf("horizontal = %d, %v", h, err)
	}
	if v, err := ExtraMonotoneVerticalVariable(2, 1, 0); err != nil || v != 6 {
		t.Fatalf("vertical = %d, %v", v, err)
	}
}

func TestExtraGridFormatAndFilename(t *testing.T) {
	g, _ := BuildExtraGridGraph(2)
	text, err := g.FormatPath([]bool{true, false, true, false})
	if err != nil {
		t.Fatal(err)
	}
	if text != "*---*\n    |\n    |\n    |\n*   *\n" {
		t.Fatalf("format:\n%s", text)
	}
	if strings.Contains(text, "=") || ExtraGridAIGFilename(5, 40) != "grid_5x5_e040.aig" {
		t.Fatal("presentation mismatch")
	}
	if _, err = g.FormatPath(nil); !errors.Is(err, ErrDimension) {
		t.Fatalf("format error = %v", err)
	}
}

func TestExtraGraphSolveLexicographic(t *testing.T) {
	graph, err := BuildExtraMonotonePathsGIA(1, 1024)
	if err != nil {
		t.Fatal(err)
	}
	result, err := SolveExtraGraphPaths(graph, ExtraGraphSolveOptions{MaxModels: 20})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Complete || result.Iterations != 7 || result.First != 6 || result.Last != 15 {
		t.Fatalf("solve = %+v", result)
	}
	previous := uint64(0)
	for index, model := range result.Models {
		var number uint64
		for _, bit := range model {
			number <<= 1
			if bit {
				number++
			}
		}
		if index > 0 && number <= previous {
			t.Fatalf("models not ordered: %d after %d", number, previous)
		}
		previous = number
	}
	if ExtraWordToFloat64(^uint64(0)) != float64(^uint64(0)) {
		t.Fatal("word conversion")
	}
}

func TestExtraGridLimitsCancellationAndErrors(t *testing.T) {
	if _, err := BuildExtraGridFrontier(3, ExtraGridFrontierOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatalf("cancel = %v", err)
	}
	if _, err := BuildExtraGridFrontier(3, ExtraGridFrontierOptions{MaxStatesPerLevel: 1}); !errors.Is(err, ErrLimit) {
		t.Fatalf("state limit = %v", err)
	}
	if _, err := BuildExtraGridFrontier(257, ExtraGridFrontierOptions{}); !errors.Is(err, ErrLimit) {
		t.Fatalf("packed-state limit = %v", err)
	}
	if _, err := BuildExtraGridGraph(extraGridMaximumSize + 1); !errors.Is(err, ErrLimit) {
		t.Fatalf("graph limit = %v", err)
	}
	frontier, _ := BuildExtraGridFrontier(2, ExtraGridFrontierOptions{})
	if _, err := frontier.DeriveGIA(1); !errors.Is(err, ErrLimit) {
		t.Fatalf("object limit = %v", err)
	}
	if _, err := BuildExtraMonotonePathsGIA(0, 10); !errors.Is(err, ErrDimension) {
		t.Fatalf("monotone error = %v", err)
	}
	if _, err := BuildExtraMonotonePathsGIA(10, 10); !errors.Is(err, ErrLimit) {
		t.Fatalf("monotone limit = %v", err)
	}
	if _, err := SolveExtraGraphPaths(nil, ExtraGraphSolveOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("solve error = %v", err)
	}
}

func FuzzExtraGridGoOnlySemantics(f *testing.F) {
	for _, seed := range []struct {
		size uint8
		mask uint64
	}{{2, 0}, {2, 9}, {3, 0x6a5}, {3, 0xfff}} {
		f.Add(seed.size, seed.mask)
	}
	f.Fuzz(func(t *testing.T, rawSize uint8, mask uint64) {
		size := 2 + int(rawSize%2)
		frontier, err := BuildExtraGridFrontier(size, ExtraGridFrontierOptions{})
		if err != nil {
			t.Fatal(err)
		}
		graph, err := frontier.DeriveGIA(1 << 16)
		if err != nil {
			t.Fatal(err)
		}
		inputs := extraMaskBits(len(frontier.Graph.Edges), mask)
		outputs, err := graph.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		if outputs[0] != extraGridBrutePath(frontier.Graph, inputs) {
			t.Fatalf("size %d mask %x", size, mask)
		}
	})
}
