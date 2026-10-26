package abc

// Stand-alone Go tests for selected src/opt/fret invariants, the generic cut
// and lag models, and the separately ported AIG forward-frontier retimer.

import (
	"errors"
	"slices"
	"testing"
)

func flowRetimeTestAIG(t testing.TB) *AigManager {
	t.Helper()
	graph := NewAigManager()
	primary := make([]AigLit, 2)
	state := make([]AigLit, 3)
	var err error
	for index := range primary {
		if primary[index], err = graph.CreateCI(); err != nil {
			t.Fatal(err)
		}
	}
	for index := range state {
		if state[index], err = graph.CreateCI(); err != nil {
			t.Fatal(err)
		}
	}
	first, err := graph.And(state[0], state[1])
	if err != nil {
		t.Fatal(err)
	}
	second, err := graph.And(first, state[2])
	if err != nil {
		t.Fatal(err)
	}
	third, err := graph.And(primary[0], primary[1])
	if err != nil {
		t.Fatal(err)
	}
	for _, output := range []AigLit{first, second, primary[0], primary[1], third} {
		if _, err = graph.CreateCO(output); err != nil {
			t.Fatal(err)
		}
	}
	if err = graph.SetRegisterCount(3); err != nil {
		t.Fatal(err)
	}
	return graph
}

func flowRetimeTraceSet(t testing.TB, graph *AigManager, frames int) map[uint64]bool {
	t.Helper()
	pis, registers, outputs := graph.PrimaryInputCount(), graph.RegisterCount(), graph.PrimaryOutputCount()
	result := make(map[uint64]bool)
	for inputCode := 0; inputCode < 1<<(pis*(frames+1)); inputCode++ {
		for stateCode := 0; stateCode < 1<<registers; stateCode++ {
			state := make([]bool, registers)
			for index := range state {
				state[index] = stateCode&(1<<index) != 0
			}
			var trace uint64
			for frame := 0; frame <= frames; frame++ {
				input := make([]bool, pis)
				for index := range input {
					input[index] = inputCode&(1<<(frame*pis+index)) != 0
				}
				values, next, err := graph.Step(input, state)
				if err != nil {
					t.Fatal(err)
				}
				if frame != 0 {
					for index := 0; index < outputs; index++ {
						if values[index] {
							trace |= 1 << uint((frame-1)*outputs+index)
						}
					}
				}
				state = next
			}
			result[uint64(inputCode)<<32|trace] = true
		}
	}
	return result
}

func TestFlowRetimeMarksAndForwardAIG(t *testing.T) {
	if FlowRetimeVisited != FlowRetimeVisitedEdge|FlowRetimeVisitedReverse || FlowRetimeInitialCare != FlowRetimeInitialZero|FlowRetimeInitialOne || FlowRetimeBlockOrConserv != FlowRetimeBlock|FlowRetimeConservative {
		t.Fatal("flag layout")
	}
	graph := flowRetimeTestAIG(t)
	before := flowRetimeTraceSet(t, graph, 2)
	result, err := FlowRetimeAIG(graph, FlowRetimeAIGOptions{Direction: FlowRetimeForward, MaxMoves: 2})
	if err != nil || result.Graph == nil || result.Moves != 2 || result.Graph.Check() != nil {
		t.Fatal(result.Moves, err)
	}
	if after := flowRetimeTraceSet(t, result.Graph, 2); !equalFlowRetimeSet(before, after) {
		t.Fatal("trace language changed")
	}
	if graph.RegisterCount() != 3 {
		t.Fatal("source mutated")
	}
	for _, options := range []FlowRetimeAIGOptions{
		{Direction: FlowRetimeBackward},
		{Direction: FlowRetimeBoth},
		{Direction: FlowRetimeForward, ComputeInitialState: true},
		{Direction: FlowRetimeForward, GuaranteeInitialState: true},
		{Direction: FlowRetimeForward, MaxDelay: 1},
		{Direction: FlowRetimeForward, BlockConstants: true},
		{Direction: FlowRetimeForward, ConservativeTiming: true},
	} {
		if got, gotErr := FlowRetimeAIG(graph, options); got.Graph != nil || !errors.Is(gotErr, ErrUnsupported) {
			t.Fatal(options, got, gotErr)
		}
	}
	if got, gotErr := FlowRetimeAIG(nil, FlowRetimeAIGOptions{}); got.Graph != nil || !errors.Is(gotErr, ErrDimension) {
		t.Fatal(got, gotErr)
	}
	if got, gotErr := FlowRetimeAIG(graph, FlowRetimeAIGOptions{Stop: func() bool { return true }}); got.Graph != nil || !errors.Is(gotErr, ErrCancelled) {
		t.Fatal(got, gotErr)
	}
}

func equalFlowRetimeSet(first, second map[uint64]bool) bool {
	if len(first) != len(second) {
		return false
	}
	for value := range first {
		if !second[value] {
			return false
		}
	}
	return true
}

func TestFlowRetimeLagEquationTimingAndOwnership(t *testing.T) {
	edges := []FlowRetimeEdge{{From: 0, To: 1, Registers: 1, Delay: 2}, {From: 1, To: 2, Delay: 3}, {From: 0, To: 2, Registers: 2, Delay: 1}}
	network, err := NewFlowRetimeLagNetwork(3, edges, FlowRetimeLagLimits{})
	if err != nil {
		t.Fatal(err)
	}
	edges[0].Registers = 99
	if got := network.Edges()[0].Registers; got != 1 {
		t.Fatal("borrowed edges", got)
	}
	retimed, err := network.Retimed([]int{0, 1, 1})
	if err != nil || !slices.Equal(retimed.Edges(), []FlowRetimeEdge{{From: 0, To: 1, Registers: 2, Delay: 2}, {From: 1, To: 2, Delay: 3}, {From: 0, To: 2, Registers: 3, Delay: 1}}) {
		t.Fatal(retimed.Edges(), err)
	}
	if registers, _ := retimed.RegisterCount(); registers != 5 {
		t.Fatal(registers)
	}
	if path, _ := retimed.PathRegisters([]int{0, 1}); path != 2 {
		t.Fatal(path)
	}
	if delay, delayErr := retimed.MaximumCombinationalDelay(); delayErr != nil || delay != 3 {
		t.Fatal(delay, delayErr)
	}
	second, err := retimed.Retimed([]int{1, 0, 0})
	combined, combinedErr := network.Retimed([]int{1, 1, 1})
	if err != nil || combinedErr != nil || !slices.Equal(second.Edges(), combined.Edges()) {
		t.Fatal(second, combined, err, combinedErr)
	}
	if _, err = network.Retimed([]int{0, -2, 0}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = network.PathRegisters([]int{1, 0}); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
	cycle, err := NewFlowRetimeLagNetwork(2, []FlowRetimeEdge{{From: 0, To: 1}, {From: 1, To: 0}}, FlowRetimeLagLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = cycle.MaximumCombinationalDelay(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func flowRetimeReachable(nodes int, arcs []FlowRetimeArc, sources, sinks []int, removed uint64) bool {
	reachable := make([]bool, nodes)
	queue := make([]int, 0, nodes)
	for _, source := range sources {
		if removed&(1<<uint(source)) == 0 {
			reachable[source] = true
			queue = append(queue, source)
		}
	}
	for len(queue) != 0 {
		from := queue[0]
		queue = queue[1:]
		for _, arc := range arcs {
			if arc.From == from && removed&(1<<uint(arc.To)) == 0 && !reachable[arc.To] {
				reachable[arc.To] = true
				queue = append(queue, arc.To)
			}
		}
	}
	for _, sink := range sinks {
		if reachable[sink] {
			return true
		}
	}
	return false
}

func flowRetimeBruteCut(nodes int, arcs []FlowRetimeArc, sources, sinks []int, blocked []bool) int {
	for count := 0; count <= nodes; count++ {
		for mask := uint64(0); mask < 1<<uint(nodes); mask++ {
			if bitsOnes64(mask) != count {
				continue
			}
			valid := true
			for _, node := range append(slices.Clone(sources), sinks...) {
				if mask&(1<<uint(node)) != 0 {
					valid = false
				}
			}
			if !valid {
				continue
			}
			for node, yes := range blocked {
				if yes {
					mask |= 1 << uint(node)
				}
			}
			if !flowRetimeReachable(nodes, arcs, sources, sinks, mask) {
				return count
			}
		}
	}
	return 0
}

func bitsOnes64(value uint64) int {
	count := 0
	for value != 0 {
		value &= value - 1
		count++
	}
	return count
}

func TestMinimumFlowRetimeCutAndModelCampaign(t *testing.T) {
	arcs := []FlowRetimeArc{{0, 1}, {1, 3}, {0, 2}, {2, 3}}
	result, err := MinimumFlowRetimeCut(4, arcs, []int{0}, []int{3}, nil, FlowRetimeCutOptions{})
	if err != nil || result.Flow != 2 || !slices.Equal(result.Cut, []int{1, 2}) {
		t.Fatal(result, err)
	}
	result, err = MinimumFlowRetimeCut(4, arcs, []int{0}, []int{3}, []bool{false, true, false, false}, FlowRetimeCutOptions{})
	if err != nil || result.Flow != 1 || !slices.Equal(result.Cut, []int{2}) {
		t.Fatal(result, err)
	}
	state := uint64(0x9e3779b97f4a7c15)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		nodes := 3 + int(state%5)
		raw := state
		modelArcs := make([]FlowRetimeArc, 0)
		for from := 0; from < nodes; from++ {
			for to := from + 1; to < nodes; to++ {
				raw = raw*6364136223846793005 + 1
				if raw>>63 != 0 {
					modelArcs = append(modelArcs, FlowRetimeArc{from, to})
				}
			}
		}
		got, modelErr := MinimumFlowRetimeCut(nodes, modelArcs, []int{0}, []int{nodes - 1}, nil, FlowRetimeCutOptions{})
		direct := false
		for _, arc := range modelArcs {
			if arc.From == 0 && arc.To == nodes-1 {
				direct = true
			}
		}
		if direct {
			if !errors.Is(modelErr, ErrLimit) {
				t.Fatalf("execution %d direct path: %v", execution, modelErr)
			}
			continue
		}
		if modelErr != nil {
			t.Fatal(execution, modelErr)
		}
		want := flowRetimeBruteCut(nodes, modelArcs, []int{0}, []int{nodes - 1}, make([]bool, nodes))
		if got.Flow != want || len(got.Cut) != want {
			t.Fatalf("execution %d nodes=%d arcs=%v got=%+v want=%d", execution, nodes, modelArcs, got, want)
		}
	}
}

func TestFlowRetimeErrorsLimitsCancellationAndAtomicity(t *testing.T) {
	if _, err := NewFlowRetimeLagNetwork(-1, nil, FlowRetimeLagLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewFlowRetimeLagNetwork(2, []FlowRetimeEdge{{0, 2, 0, 0}}, FlowRetimeLagLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewFlowRetimeLagNetwork(2, nil, FlowRetimeLagLimits{MaxNodes: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err := NewFlowRetimeLagNetwork(2, nil, FlowRetimeLagLimits{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	if result, err := MinimumFlowRetimeCut(2, []FlowRetimeArc{{0, 1}}, []int{0}, []int{1}, nil, FlowRetimeCutOptions{MaxFlow: 1}); result.Cut != nil || !errors.Is(err, ErrLimit) {
		t.Fatal(result, err)
	}
	if result, err := MinimumFlowRetimeCut(2, []FlowRetimeArc{{0, 1}}, []int{0}, []int{1}, nil, FlowRetimeCutOptions{Stop: func() bool { return true }}); result.Cut != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
}

func FuzzFlowRetimeLagAndCut(f *testing.F) {
	f.Add([]byte{3, 1, 2, 4, 5, 6})
	f.Add([]byte{7, 0xff, 0, 1, 9})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 || len(data) > 128 {
			return
		}
		nodes := 2 + int(data[0]%7)
		lags := make([]int, nodes)
		for index := range lags {
			lags[index] = int(int8(data[(index+1)%len(data)]) % 4)
		}
		edges := make([]FlowRetimeEdge, 0)
		arcs := make([]FlowRetimeArc, 0)
		cursor := 1
		for from := 0; from < nodes; from++ {
			for to := from + 1; to < nodes; to++ {
				value := data[cursor%len(data)]
				cursor++
				if value&1 == 0 {
					continue
				}
				delta := lags[to] - lags[from]
				registers := int(value % 3)
				if delta < 0 {
					registers += -delta
				}
				edges = append(edges, FlowRetimeEdge{From: from, To: to, Registers: registers, Delay: int(value % 17)})
				arcs = append(arcs, FlowRetimeArc{from, to})
			}
		}
		network, err := NewFlowRetimeLagNetwork(nodes, edges, FlowRetimeLagLimits{})
		if err != nil {
			t.Fatal(err)
		}
		retimed, err := network.Retimed(lags)
		if err != nil {
			t.Fatal(err)
		}
		for index, edge := range edges {
			want := edge.Registers + lags[edge.To] - lags[edge.From]
			if retimed.Edges()[index].Registers != want {
				t.Fatal(index, retimed.Edges()[index], want)
			}
		}
		result, err := MinimumFlowRetimeCut(nodes, arcs, []int{0}, []int{nodes - 1}, nil, FlowRetimeCutOptions{})
		direct := false
		for _, arc := range arcs {
			if arc.From == 0 && arc.To == nodes-1 {
				direct = true
			}
		}
		if direct {
			if !errors.Is(err, ErrLimit) {
				t.Fatal(err)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		want := flowRetimeBruteCut(nodes, arcs, []int{0}, []int{nodes - 1}, make([]bool, nodes))
		if result.Flow != want {
			t.Fatal(result, want)
		}
	})
}
