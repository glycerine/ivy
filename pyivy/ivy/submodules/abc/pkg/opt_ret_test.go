package abc

// Stand-alone Go models and replay tests for src/opt/ret.

import (
	"errors"
	"slices"
	"testing"
)

func retimingANDNetwork(t testing.TB, first, second bool) *RetimingNetwork {
	t.Helper()
	network, err := NewRetimingNetwork(
		[]RetimingNode{
			{PrimaryInput: true},
			{PrimaryInput: true},
			{Delay: 1, Truth: []bool{false, false, false, true}},
			{Truth: []bool{false, true}},
		},
		[]RetimingEdge{
			{From: 0, To: 2, Initial: []bool{first}},
			{From: 1, To: 2, Initial: []bool{second}},
			{From: 2, To: 3},
		},
		[]int{3}, RetimingLimits{},
	)
	if err != nil {
		t.Fatal(err)
	}
	return network
}

func replayRetiming(t testing.TB, network *RetimingNetwork, inputs [][]bool) []bool {
	t.Helper()
	result := make([]bool, 0, len(inputs))
	state := network
	for _, input := range inputs {
		output, next, err := state.Step(input)
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, output[0])
		state = next
	}
	return result
}

func TestRetimingExactForwardBackwardReplay(t *testing.T) {
	original := retimingANDNetwork(t, true, false)
	if enabled, err := original.NodeEnabled(2, RetimingForward); err != nil || !enabled {
		t.Fatal(enabled, err)
	}
	forward, err := original.Move(2, RetimingForward)
	if err != nil {
		t.Fatal(err)
	}
	if got := []int{len(forward.edges[0].Initial), len(forward.edges[1].Initial), len(forward.edges[2].Initial)}; !slices.Equal(got, []int{0, 0, 1}) {
		t.Fatal(got)
	}
	inputs := [][]bool{{false, false}, {true, true}, {true, false}, {true, true}}
	if before, after := replayRetiming(t, original, inputs), replayRetiming(t, forward, inputs); !slices.Equal(before, after) {
		t.Fatal(before, after)
	}
	backward, err := forward.Move(2, RetimingBackward)
	if err != nil {
		t.Fatal(err)
	}
	if before, after := replayRetiming(t, original, inputs), replayRetiming(t, backward, inputs); !slices.Equal(before, after) {
		t.Fatal(before, after)
	}
	if original.RegisterCount() != 2 {
		t.Fatal("source mutated")
	}
}

func TestRetimingIncrementalTimingLagsAndOwnership(t *testing.T) {
	network := retimingANDNetwork(t, true, true)
	nodes, edges, outputs := network.Nodes(), network.Edges(), network.Outputs()
	nodes[2].Truth[3] = false
	edges[0].Initial[0] = false
	outputs[0] = 0
	if !network.nodes[2].Truth[3] || !network.edges[0].Initial[0] || network.outputs[0] != 3 {
		t.Fatal("borrowed snapshot")
	}
	delay, critical, err := network.Timing(RetimingForward)
	if err != nil || delay != 1 || !slices.Equal(critical, []int{2}) {
		t.Fatal(delay, critical, err)
	}
	run, err := network.Incremental(RetimingForward, true)
	if err != nil || run.Moves != 1 || run.Network.RegisterCount() != 1 {
		t.Fatal(run.Moves, err)
	}
	lagged, err := network.ApplyLags([]int{0, 0, -1, 0})
	if err != nil || lagged.Moves != 1 {
		t.Fatal(lagged.Moves, err)
	}
	if !equalRetimingEdges(run.Network.edges, lagged.Network.edges) {
		t.Fatal(run.Network.edges, lagged.Network.edges)
	}
	value, err := network.LValue()
	if err != nil || value.Period < 1 || len(value.Lags) != 4 {
		t.Fatal(value, err)
	}
	if value.Lags[0] != 0 || value.Lags[1] != 0 {
		t.Fatal("primary-input lags", value.Lags)
	}
	cut, err := network.MaxFlowCut(RetimingForward)
	if err != nil || cut.Flow != 1 || len(cut.Nodes) != 1 {
		t.Fatal(cut, err)
	}
}

func TestRetimingMinimumDelayCriticalIteration(t *testing.T) {
	network, err := NewRetimingNetwork(
		[]RetimingNode{
			{PrimaryInput: true},
			{Delay: 1, Truth: []bool{false, true}},
			{Delay: 1, Truth: []bool{false, true}},
			{Truth: []bool{false, true}},
		},
		[]RetimingEdge{
			{From: 0, To: 1, Initial: []bool{true}},
			{From: 1, To: 2},
			{From: 2, To: 3},
		}, []int{3}, RetimingLimits{MaxIterations: 8},
	)
	if err != nil {
		t.Fatal(err)
	}
	before, _, err := network.Timing(RetimingForward)
	if err != nil || before != 2 {
		t.Fatal(before, err)
	}
	result, err := network.MinDelay(RetimingForward, 1, false)
	if err != nil || result.Delay != 1 || result.BestRound != 1 || result.Network == nil {
		t.Fatal(result, err)
	}
	inputs := [][]bool{{false}, {true}, {false}, {true}}
	if first, second := replayRetiming(t, network, inputs), replayRetiming(t, result.Network, inputs); !slices.Equal(first, second) {
		t.Fatal(first, second)
	}
}

func equalRetimingEdges(first, second []RetimingEdge) bool {
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		if first[i].From != second[i].From || first[i].To != second[i].To || !slices.Equal(first[i].Initial, second[i].Initial) {
			return false
		}
	}
	return true
}

func retimingModelNext(registers [3][]bool, truth []bool, input [2]bool) (bool, [3][]bool) {
	args := input
	for i := range args {
		if len(registers[i]) != 0 {
			args[i] = registers[i][len(registers[i])-1]
		}
	}
	index := 0
	if args[0] {
		index |= 1
	}
	if args[1] {
		index |= 2
	}
	gate := truth[index]
	output := gate
	if len(registers[2]) != 0 {
		output = registers[2][len(registers[2])-1]
	}
	next := [3][]bool{}
	sources := [3]bool{input[0], input[1], gate}
	for i := range registers {
		next[i] = slices.Clone(registers[i])
		if len(next[i]) != 0 {
			copy(next[i][1:], registers[i][:len(registers[i])-1])
			next[i][0] = sources[i]
		}
	}
	return output, next
}

func TestRetimingIndependentSequentialModel5000(t *testing.T) {
	state := uint64(0x243f6a8885a308d3)
	for execution := 0; execution < 5000; execution++ {
		state ^= state << 7
		state ^= state >> 9
		truth := make([]bool, 4)
		for i := range truth {
			truth[i] = state&(1<<uint(i)) != 0
		}
		first, second := state&(1<<8) != 0, state&(1<<9) != 0
		network, err := NewRetimingNetwork(
			[]RetimingNode{{PrimaryInput: true}, {PrimaryInput: true}, {Delay: 1, Truth: truth}, {Truth: []bool{false, true}}},
			[]RetimingEdge{{From: 0, To: 2, Initial: []bool{first}}, {From: 1, To: 2, Initial: []bool{second}}, {From: 2, To: 3}},
			[]int{3}, RetimingLimits{},
		)
		if err != nil {
			t.Fatal(err)
		}
		moved, err := network.Move(2, RetimingForward)
		if err != nil {
			t.Fatal(execution, err)
		}
		registers := [3][]bool{{first}, {second}, nil}
		modelMoved := [3][]bool{nil, nil, {truth[retimingBoolIndex(first, second)]}}
		before, after := network, moved
		for frame := 0; frame < 5; frame++ {
			state = state*6364136223846793005 + 1
			input := [2]bool{state&1 != 0, state&2 != 0}
			want, nextRegisters := retimingModelNext(registers, truth, input)
			wantMoved, nextModelMoved := retimingModelNext(modelMoved, truth, input)
			outBefore, nextBefore, beforeErr := before.Step(input[:])
			outAfter, nextAfter, afterErr := after.Step(input[:])
			if beforeErr != nil || afterErr != nil || outBefore[0] != want || outAfter[0] != wantMoved || want != wantMoved {
				t.Fatalf("execution=%d frame=%d got=%v/%v want=%v/%v errors=%v/%v", execution, frame, outBefore, outAfter, want, wantMoved, beforeErr, afterErr)
			}
			registers, modelMoved, before, after = nextRegisters, nextModelMoved, nextBefore, nextAfter
		}
	}
}

func TestRetimingIndependentDeepSequentialModel5000(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	next := func() bool {
		state ^= state << 13
		state ^= state >> 7
		state ^= state << 17
		return state&1 != 0
	}
	for execution := 0; execution < 5000; execution++ {
		truth := make([]bool, 4)
		for i := range truth {
			truth[i] = next()
		}
		depths := [3]int{1 + int(state&3), 1 + int(state>>2&3), int(state >> 4 & 3)}
		registers := [3][]bool{}
		for edge, depth := range depths {
			registers[edge] = make([]bool, depth)
			for bit := range registers[edge] {
				registers[edge][bit] = next()
			}
		}
		network, err := NewRetimingNetwork(
			[]RetimingNode{{PrimaryInput: true}, {PrimaryInput: true}, {Delay: 1, Truth: truth}, {Truth: []bool{false, true}}},
			[]RetimingEdge{{From: 0, To: 2, Initial: registers[0]}, {From: 1, To: 2, Initial: registers[1]}, {From: 2, To: 3, Initial: registers[2]}},
			[]int{3}, RetimingLimits{},
		)
		if err != nil {
			t.Fatal(execution, err)
		}
		moved, err := network.Move(2, RetimingForward)
		if err != nil {
			t.Fatal(execution, err)
		}
		modelMoved := [3][]bool{
			slices.Clone(registers[0][:len(registers[0])-1]),
			slices.Clone(registers[1][:len(registers[1])-1]),
			append([]bool{truth[retimingBoolIndex(registers[0][len(registers[0])-1], registers[1][len(registers[1])-1])]}, registers[2]...),
		}
		before, after := network, moved
		for frame := 0; frame < 8; frame++ {
			input := [2]bool{next(), next()}
			want, nextRegisters := retimingModelNext(registers, truth, input)
			wantMoved, nextMovedRegisters := retimingModelNext(modelMoved, truth, input)
			gotBefore, nextBefore, beforeErr := before.Step(input[:])
			gotAfter, nextAfter, afterErr := after.Step(input[:])
			if beforeErr != nil || afterErr != nil || gotBefore[0] != want || gotAfter[0] != wantMoved || want != wantMoved {
				t.Fatalf("execution=%d frame=%d got=%v/%v want=%v/%v errors=%v/%v", execution, frame, gotBefore, gotAfter, want, wantMoved, beforeErr, afterErr)
			}
			registers, modelMoved, before, after = nextRegisters, nextMovedRegisters, nextBefore, nextAfter
		}
	}
}

func retimingBoolIndex(first, second bool) int {
	index := 0
	if first {
		index++
	}
	if second {
		index += 2
	}
	return index
}

func TestRetimingErrorsLimitsCancellationAtomicity(t *testing.T) {
	if _, err := NewRetimingNetwork([]RetimingNode{{PrimaryInput: true, Delay: -1}}, nil, nil, RetimingLimits{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err := NewRetimingNetwork([]RetimingNode{{PrimaryInput: true}}, nil, nil, RetimingLimits{MaxNodes: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	network := retimingANDNetwork(t, false, true)
	if result, err := network.Move(3, RetimingForward); result != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal(result, err)
	}
	if result, err := network.ApplyLags([]int{0}); result.Network != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(result, err)
	}
	cancelled, err := NewRetimingNetwork(network.Nodes(), network.Edges(), network.Outputs(), RetimingLimits{Stop: func() bool { return true }})
	if cancelled != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(cancelled, err)
	}
	if network.RegisterCount() != 2 {
		t.Fatal("failed operation mutated source")
	}
	constant, err := NewRetimingNetwork(
		[]RetimingNode{{Truth: []bool{true}}, {Truth: []bool{false, true}}},
		[]RetimingEdge{{From: 0, To: 1, Initial: []bool{true}}}, []int{1}, RetimingLimits{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if enabled, enabledErr := constant.NodeEnabled(0, RetimingBackward); enabledErr != nil || !enabled {
		t.Fatal("backward constant", enabled, enabledErr)
	}
	movedConstant, err := constant.Move(0, RetimingBackward)
	if err != nil || movedConstant.RegisterCount() != 0 {
		t.Fatal("backward constant move", err)
	}
	wideTruth := make([]bool, 1<<4)
	wideNodes := []RetimingNode{{PrimaryInput: true}, {PrimaryInput: true}, {PrimaryInput: true}, {PrimaryInput: true}, {Truth: wideTruth}, {Truth: []bool{false, true}}}
	wideEdges := []RetimingEdge{{From: 0, To: 4}, {From: 1, To: 4}, {From: 2, To: 4}, {From: 3, To: 4}, {From: 4, To: 5, Initial: []bool{false}}}
	wide, err := NewRetimingNetwork(wideNodes, wideEdges, []int{5}, RetimingLimits{MaxInitialAssignments: 8})
	if err != nil {
		t.Fatal(err)
	}
	if result, moveErr := wide.Move(4, RetimingBackward); result != nil || !errors.Is(moveErr, ErrLimit) {
		t.Fatal("initial assignment limit", result, moveErr)
	}
	moveLimited, err := NewRetimingNetwork(
		[]RetimingNode{{PrimaryInput: true}, {Truth: []bool{false, true}}, {Truth: []bool{false, true}}, {Truth: []bool{false, true}}},
		[]RetimingEdge{{From: 0, To: 1, Initial: []bool{false}}, {From: 1, To: 2}, {From: 2, To: 3}}, []int{3}, RetimingLimits{MaxMoves: 1},
	)
	if err != nil {
		t.Fatal(err)
	}
	if result, runErr := moveLimited.Incremental(RetimingForward, false); result.Network != nil || !errors.Is(runErr, ErrLimit) {
		t.Fatal("move limit", result, runErr)
	}
	if moveLimited.RegisterCount() != 1 {
		t.Fatal("limited run mutated source")
	}
}

func FuzzRetimingMoveReplay(f *testing.F) {
	f.Add(byte(0x8), true, false, []byte{0, 1, 3, 2})
	f.Add(byte(0x6), false, true, []byte{3, 3, 0})
	f.Fuzz(func(t *testing.T, table byte, first, second bool, raw []byte) {
		if len(raw) > 128 {
			return
		}
		truth := make([]bool, 4)
		for i := range truth {
			truth[i] = table&(1<<uint(i)) != 0
		}
		network, err := NewRetimingNetwork(
			[]RetimingNode{{PrimaryInput: true}, {PrimaryInput: true}, {Delay: 1, Truth: truth}, {Truth: []bool{false, true}}},
			[]RetimingEdge{{From: 0, To: 2, Initial: []bool{first}}, {From: 1, To: 2, Initial: []bool{second}}, {From: 2, To: 3}},
			[]int{3}, RetimingLimits{},
		)
		if err != nil {
			t.Fatal(err)
		}
		moved, err := network.Move(2, RetimingForward)
		if err != nil {
			t.Fatal(err)
		}
		inputs := make([][]bool, len(raw)+2)
		for i := range inputs {
			value := byte(i)
			if len(raw) != 0 {
				value = raw[i%len(raw)]
			}
			inputs[i] = []bool{value&1 != 0, value&2 != 0}
		}
		if before, after := replayRetiming(t, network, inputs), replayRetiming(t, moved, inputs); !slices.Equal(before, after) {
			t.Fatal(before, after)
		}
	})
}
