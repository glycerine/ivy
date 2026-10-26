package abc

// Stand-alone tests for the paired mesh embedders in bmcMesh.c and bmcMesh2.c.

import (
	"errors"
	"reflect"
	"slices"
	"testing"
)

func giaMeshInputGraph(t testing.TB) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	input, _ := graph.AppendCI()
	graph.AppendCO(input)
	return graph
}

func giaMeshAndGraph(t testing.TB, chain bool) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	input0, _ := graph.AppendCI()
	input1, _ := graph.AppendCI()
	first, err := graph.AppendAnd(input0, input1.Not())
	if err != nil {
		t.Fatal(err)
	}
	root := first
	if chain {
		root, err = graph.AppendAnd(first, input0.Not())
		if err != nil {
			t.Fatal(err)
		}
	}
	graph.AppendCO(root)
	return graph
}

func giaMeshVerifyPlacement(t testing.TB, graph *GiaManager, placement GiaMeshPlacement) {
	t.Helper()
	if placement.Status != SatSatisfiable || len(placement.Cells) != placement.X*placement.Y || placement.TargetGraph != placement.GraphCount-1 {
		t.Fatal("incomplete placement", placement.Status, len(placement.Cells), placement.TargetGraph)
	}
	inputs, fanins, err := giaMeshGraph(graph)
	if err != nil || inputs != placement.InputCount || inputs+len(fanins) != placement.GraphCount {
		t.Fatal(inputs, fanins, err, placement)
	}
	cellAt := func(x, y int) GiaMeshCell { return placement.Cells[y*placement.X+x] }
	directionX := [4]int{-1, 0, 1, 0}
	directionY := [4]int{0, -1, 0, 1}
	configurations := [12][2]int{{0, 1}, {0, 2}, {0, 3}, {1, 2}, {1, 3}, {2, 3}, {1, 0}, {2, 0}, {3, 0}, {2, 1}, {3, 1}, {3, 2}}
	bufferCount, nodeCount := 0, 0
	for y := 0; y < placement.Y; y++ {
		for x := 0; x < placement.X; x++ {
			cell := cellAt(x, y)
			boundary := x == 0 || x == placement.X-1 || y == 0 || y == placement.Y-1
			if cell.X != x || cell.Y != y || cell.Boundary != boundary {
				t.Fatal("cell order", cell, x, y)
			}
			if boundary {
				if cell.Time != 0 || cell.Graph >= inputs {
					t.Fatal("bad boundary", cell)
				}
				continue
			}
			if cell.Configuration >= 0 {
				if cell.Configuration < 4 {
					bufferCount++
				} else {
					nodeCount++
				}
			}
			if cell.Graph < 0 {
				continue
			}
			if !cell.Used || cell.Time < 1 || cell.Time >= placement.Latency || cell.Configuration < 0 || cell.Configuration >= giaMeshConfigurations {
				t.Fatal("incomplete active cell", cell)
			}
			if cell.Configuration < 4 {
				direction := cell.Configuration
				neighbor := cellAt(x+directionX[direction], y+directionY[direction])
				if neighbor.Time != cell.Time-1 || neighbor.Graph != cell.Graph {
					t.Fatal("bad buffer", cell, neighbor)
				}
				continue
			}
			if cell.Graph < inputs {
				t.Fatal("PI selected with node configuration", cell)
			}
			fanin := fanins[cell.Graph-inputs]
			directions := configurations[cell.Configuration-4]
			for input := 0; input < 2; input++ {
				direction := directions[input]
				neighbor := cellAt(x+directionX[direction], y+directionY[direction])
				if neighbor.Time != cell.Time-1 || neighbor.Graph != fanin[input] {
					t.Fatal("bad node", cell, input, neighbor, fanin)
				}
			}
		}
	}
	if center := cellAt(1, 1); center.Graph != placement.TargetGraph {
		t.Fatal("missing final graph", center, placement.TargetGraph)
	}
	if placement.Buffers != bufferCount || placement.Nodes != nodeCount {
		t.Fatal("wrong configuration counts", placement.Buffers, bufferCount, placement.Nodes, nodeCount)
	}
}

func TestGiaMeshInputAndNodePlacements(t *testing.T) {
	for _, test := range []struct {
		graph   *GiaManager
		buffers int
		nodes   int
	}{
		{giaMeshInputGraph(t), 1, 0},
		{giaMeshAndGraph(t, false), 0, 1},
	} {
		before := test.graph.Clone()
		placement, err := test.graph.EmbedGiaMesh(GiaMeshOptions{X: 3, Y: 3, Latency: 2})
		if err != nil || placement.Status != SatSatisfiable || placement.Buffers != test.buffers || placement.Nodes != test.nodes || placement.Iterations < 1 || placement.Variables == 0 || placement.Clauses == 0 || placement.Literals == 0 {
			t.Fatal(placement, err)
		}
		giaMeshVerifyPlacement(t, test.graph, placement)
		if !reflect.DeepEqual(test.graph, before) {
			t.Fatal("source graph mutated")
		}
	}

	chain := giaMeshAndGraph(t, true)
	placement, err := chain.EmbedGiaMesh(GiaMeshOptions{X: 3, Y: 3, Latency: 2})
	if err != nil || placement.Status != SatUnsatisfiable || placement.Cells != nil {
		t.Fatal(placement, err)
	}
	placement, err = chain.EmbedGiaMesh(GiaMeshOptions{X: 4, Y: 4, Latency: 3})
	if err != nil || placement.Status != SatSatisfiable || placement.Nodes < 2 || placement.Buffers < 1 {
		t.Fatal(placement, err)
	}
	giaMeshVerifyPlacement(t, chain, placement)
}

func TestGiaMeshTargetsLastGraphItemNotOutputDriver(t *testing.T) {
	graph := NewGiaManager()
	input0, _ := graph.AppendCI()
	input1, _ := graph.AppendCI()
	first, _ := graph.AppendAnd(input0, input1)
	second, _ := graph.AppendAnd(input0.Not(), input1)
	graph.AppendCO(first)
	placement, err := graph.EmbedGiaMesh(GiaMeshOptions{X: 3, Y: 3, Latency: 2})
	if err != nil || placement.Status != SatSatisfiable || placement.TargetGraph != int(second.ID())-1 || placement.TargetGraph == int(first.ID())-1 {
		t.Fatal(placement, err)
	}
	giaMeshVerifyPlacement(t, graph, placement)
}

func TestGiaMeshOneHotRefinement(t *testing.T) {
	model := []bool{true, true, false, true}
	var clauses [][]SatLit
	count, err := giaMeshAddOneHotness(model, 0, 4, func(clause []SatLit) error {
		clauses = append(clauses, slices.Clone(clause))
		return nil
	})
	want := [][]SatLit{{1, 3}, {1, 7}, {3, 7}}
	if err != nil || count != 3 || !reflect.DeepEqual(clauses, want) {
		t.Fatal(count, clauses, want, err)
	}
	for _, bounds := range [][2]int{{-1, 1}, {1, 1}, {0, 5}} {
		if count, err = giaMeshAddOneHotness(model, bounds[0], bounds[1], func([]SatLit) error { return nil }); count != 0 || !errors.Is(err, ErrDimension) {
			t.Fatal(bounds, count, err)
		}
	}
	large := make([]bool, 101)
	if _, err = giaMeshAddOneHotness(large, 0, len(large), func([]SatLit) error { return nil }); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestGiaMeshErrorsLimitsCancellationAndUnsupported(t *testing.T) {
	graph := giaMeshInputGraph(t)
	tests := []struct {
		graph   *GiaManager
		options GiaMeshOptions
		want    error
	}{
		{nil, GiaMeshOptions{X: 3, Y: 3, Latency: 2}, ErrDimension},
		{graph, GiaMeshOptions{X: 2, Y: 3, Latency: 2}, ErrDimension},
		{graph, GiaMeshOptions{X: 3, Y: 2, Latency: 2}, ErrDimension},
		{graph, GiaMeshOptions{X: 3, Y: 3, Latency: 1}, ErrDimension},
		{graph, GiaMeshOptions{X: 101, Y: 3, Latency: 2}, ErrDimension},
		{graph, GiaMeshOptions{X: 3, Y: 101, Latency: 2}, ErrDimension},
		{graph, GiaMeshOptions{X: 3, Y: 3, Latency: 101}, ErrDimension},
		{graph, GiaMeshOptions{X: 3, Y: 3, Latency: 2, MaxVariables: -1}, ErrDimension},
		{graph, GiaMeshOptions{X: 3, Y: 3, Latency: 2, MaxClauses: -1}, ErrDimension},
		{graph, GiaMeshOptions{X: 3, Y: 3, Latency: 2, MaxLiterals: -1}, ErrDimension},
		{graph, GiaMeshOptions{X: 3, Y: 3, Latency: 2, MaxIterations: -1}, ErrDimension},
		{graph, GiaMeshOptions{X: 3, Y: 3, Latency: 2, MaxVariables: 1}, ErrLimit},
		{graph, GiaMeshOptions{X: 3, Y: 3, Latency: 2, MaxClauses: 1}, ErrLimit},
		{graph, GiaMeshOptions{X: 3, Y: 3, Latency: 2, MaxLiterals: 1}, ErrLimit},
		{graph, GiaMeshOptions{X: 3, Y: 3, Latency: 2, Stop: func() bool { return true }}, ErrCancelled},
	}
	for index, test := range tests {
		placement, err := test.graph.EmbedGiaMesh(test.options)
		if !reflect.DeepEqual(placement, GiaMeshPlacement{}) || !errors.Is(err, test.want) {
			t.Fatal(index, placement, err, test.want)
		}
	}
	unknown, err := graph.EmbedGiaMesh(GiaMeshOptions{X: 3, Y: 3, Latency: 2, SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || unknown.Status != SatUnknown || unknown.Reason != SatCancelled || unknown.Cells != nil {
		t.Fatal(unknown, err)
	}

	noOutput := NewGiaManager()
	noOutput.AppendCI()
	if _, err = noOutput.EmbedGiaMesh(GiaMeshOptions{X: 3, Y: 3, Latency: 2}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	constant := NewGiaManager()
	input, _ := constant.AppendCI()
	gate, _ := constant.AppendAnd(GiaConst0, input)
	constant.AppendCO(gate)
	if _, err = constant.EmbedGiaMesh(GiaMeshOptions{X: 3, Y: 3, Latency: 2}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	buffered := NewGiaManager()
	input, _ = buffered.AppendCI()
	buffer, _ := buffered.AppendBuffer(input)
	buffered.AppendCO(buffer)
	if _, err = buffered.EmbedGiaMesh(GiaMeshOptions{X: 3, Y: 3, Latency: 2}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	gapped := NewGiaManager()
	input0, _ := gapped.AppendCI()
	input1, _ := gapped.AppendCI()
	gapped.AppendCO(input0)
	gapped.AppendAnd(input0, input1)
	if _, err = gapped.EmbedGiaMesh(GiaMeshOptions{X: 3, Y: 3, Latency: 2}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	wide := NewGiaManager()
	var last GiaLit
	for index := 0; index < 101; index++ {
		last, _ = wide.AppendCI()
	}
	wide.AppendCO(last)
	if _, err = wide.EmbedGiaMesh(GiaMeshOptions{X: 3, Y: 3, Latency: 2}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}

func FuzzGiaMeshThreeByThree(f *testing.F) {
	f.Add([]byte{1, 3, 5, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		graph := NewGiaManager()
		inputs := []GiaLit{}
		for index := 0; index < 2+int(data[0]&1); index++ {
			input, _ := graph.AppendCI()
			inputs = append(inputs, input)
		}
		signals := slices.Clone(inputs)
		lastFaninsAreInputs := true
		for index := 0; index < len(data) && index < 3; index++ {
			left := int(data[index]) % len(signals)
			right := (left + 1 + int(data[index]>>4)) % len(signals)
			if signals[left].ID() == signals[right].ID() {
				continue
			}
			lastFaninsAreInputs = left < len(inputs) && right < len(inputs)
			gate, err := graph.AppendAnd(signals[left].NotIf(data[index]&1 != 0), signals[right].NotIf(data[index]&2 != 0))
			if err == nil {
				signals = append(signals, gate)
			}
		}
		graph.AppendCO(signals[len(signals)-1])
		expectedSAT := len(signals) == len(inputs) || lastFaninsAreInputs
		placement, err := graph.EmbedGiaMesh(GiaMeshOptions{X: 3, Y: 3, Latency: 2, MaxVariables: 1000, MaxClauses: 10000, MaxLiterals: 40000, MaxIterations: 100})
		if err != nil || (placement.Status == SatSatisfiable) != expectedSAT {
			t.Fatal(expectedSAT, placement.Status, placement, err)
		}
		if placement.Status == SatSatisfiable {
			giaMeshVerifyPlacement(t, graph, placement)
		}
	})
}
