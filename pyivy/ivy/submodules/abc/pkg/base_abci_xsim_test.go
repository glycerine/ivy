package abc

// Standalone unit, concurrency and fuzz tests for src/base/abci/abcXsim.c.

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func abcXSimFixture(t testing.TB, init ABCLatchInit) *ABCNetwork {
	t.Helper()
	network, err := NewABCNetwork(ABCNetworkStrash, ABCFunctionAIG, ABCNetworkLimits{MaxObjects: 1000, MaxEdges: 4000})
	if err != nil {
		t.Fatal(err)
	}
	first, _ := network.CreatePI()
	second, _ := network.CreatePI()
	bi, _, bo, err := network.CreateLatchBox(init)
	if err != nil {
		t.Fatal(err)
	}
	node, err := network.CreateLogicNode("11 1\n", ABCObjectEdge{ID: first}, ABCObjectEdge{ID: bo})
	if err != nil {
		t.Fatal(err)
	}
	po, _ := network.CreatePO()
	if err = network.AddFanin(po, ABCObjectEdge{ID: node}); err != nil {
		t.Fatal(err)
	}
	next, err := network.CreateLogicNode("11 1\n", ABCObjectEdge{ID: first, Complemented: true}, ABCObjectEdge{ID: second})
	if err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(bi, ABCObjectEdge{ID: next}); err != nil {
		t.Fatal(err)
	}
	if err = CheckABCNetwork(network, ABCAnalysisOptions{}); err != nil {
		t.Fatal(err)
	}
	return network
}

func abcXSimObjectsEqual(first, second []ABCNetworkObject) bool {
	if len(first) != len(second) {
		return false
	}
	for i := range first {
		a, b := first[i], second[i]
		if a.ID != b.ID || a.Type != b.Type || a.Name != b.Name || a.Function != b.Function || a.Init != b.Init || a.Level != b.Level || a.Phase != b.Phase || a.MarkA != b.MarkA || a.MarkB != b.MarkB || a.MarkC != b.MarkC || a.Alive != b.Alive || !slices.Equal(a.Fanins, b.Fanins) || !slices.Equal(a.Fanouts, b.Fanouts) {
			return false
		}
	}
	return true
}

func TestABCXSimDefinitionsTextAndRandomHelpers(t *testing.T) {
	if len(ABCXSimDefinitions) != 11 || ABCXSimDefinitions[0] != "Abc_NtkCycleInitState" || ABCXSimDefinitions[10] != "Abc_XsimRand3" {
		t.Fatalf("definitions=%v", ABCXSimDefinitions)
	}
	text, err := ABCXSimText([]AigTernary{AigTernaryZero, AigTernaryOne, AigTernaryUnknown})
	if err != nil || text != "01x" {
		t.Fatalf("text=%q err=%v", text, err)
	}
	if _, err = ABCXSimText([]AigTernary{0}); !errors.Is(err, ErrDimension) {
		t.Fatalf("invalid text err=%v", err)
	}
	random := newABCXSimRandom()
	want := []AigTernary{AigTernaryOne, AigTernaryZero, AigTernaryOne, AigTernaryZero, AigTernaryZero}
	got := make([]AigTernary, len(want))
	for i := range got {
		got[i] = abcXSimRandom2(&random)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("native random bits=%v want=%v", got, want)
	}
	random = newABCXSimRandom()
	for i := 0; i < 100; i++ {
		if value := abcXSimRandom3(&random); !abcXSimValid(value) {
			t.Fatalf("random ternary=%v", value)
		}
	}
}

func TestABCXValueSimulationTraceStateAndOwnership(t *testing.T) {
	network := abcXSimFixture(t, ABCInitZero)
	before := network.Clone()
	trace, err := ABCXValueSimulate(network, ABCXSimOptions{Frames: 3, UnknownInputs: true})
	if err != nil {
		t.Fatal(err)
	}
	wantStates := []AigTernary{AigTernaryZero, AigTernaryUnknown, AigTernaryUnknown}
	wantOutputs := []AigTernary{AigTernaryZero, AigTernaryUnknown, AigTernaryUnknown}
	if len(trace.Frames) != 3 {
		t.Fatalf("frames=%d", len(trace.Frames))
	}
	for frame, row := range trace.Frames {
		if row.Frame != frame || !slices.Equal(row.Inputs, []AigTernary{AigTernaryUnknown, AigTernaryUnknown}) || !slices.Equal(row.Latches, []AigTernary{wantStates[frame]}) || !slices.Equal(row.Outputs, []AigTernary{wantOutputs[frame]}) {
			t.Fatalf("frame %d: %+v", frame, row)
		}
	}
	if !abcXSimObjectsEqual(network.objects, before.objects) || !slices.Equal(network.cis, before.cis) || !slices.Equal(network.cos, before.cos) {
		t.Fatal("source network mutated")
	}
	trace.Frames[0].Inputs[0] = AigTernaryZero
	again, err := ABCXValueSimulate(network, ABCXSimOptions{Frames: 1, UnknownInputs: true})
	if err != nil || again.Frames[0].Inputs[0] != AigTernaryUnknown {
		t.Fatalf("trace aliases state: %+v err=%v", again, err)
	}
	unknown, err := ABCXValueSimulate(network, ABCXSimOptions{Frames: 1, UnknownInputs: true, UnknownState: true})
	if err != nil || unknown.Frames[0].Latches[0] != AigTernaryUnknown || unknown.Frames[0].Outputs[0] != AigTernaryUnknown {
		t.Fatalf("unknown state trace=%+v err=%v", unknown, err)
	}
}

func TestABCCycleInitStateRandomUnknownAndConcurrency(t *testing.T) {
	network := abcXSimFixture(t, ABCInitZero)
	cycled, err := ABCCycleInitState(network, ABCCycleInitStateOptions{Frames: 2})
	if err != nil {
		t.Fatal(err)
	}
	values, _ := ABCLatchValues(cycled)
	if !slices.Equal(values, []ABCLatchInit{ABCInitZero}) {
		t.Fatalf("random cycled state=%v", values)
	}
	original, _ := ABCLatchValues(network)
	if !slices.Equal(original, []ABCLatchInit{ABCInitZero}) {
		t.Fatalf("source state changed=%v", original)
	}
	unknown, err := ABCCycleInitState(network, ABCCycleInitStateOptions{Frames: 1, UseUnknownInputs: true})
	if err != nil {
		t.Fatal(err)
	}
	values, _ = ABCLatchValues(unknown)
	if !slices.Equal(values, []ABCLatchInit{ABCInitDC}) {
		t.Fatalf("unknown cycled state=%v", values)
	}
	zero, err := ABCCycleInitState(network, ABCCycleInitStateOptions{Frames: 0})
	if err != nil {
		t.Fatal(err)
	}
	values, _ = ABCLatchValues(zero)
	if !slices.Equal(values, []ABCLatchInit{ABCInitZero}) {
		t.Fatalf("zero-frame state=%v", values)
	}
	var wait sync.WaitGroup
	var failures atomic.Int32
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			trace, runErr := ABCXValueSimulate(network, ABCXSimOptions{Frames: 4})
			if runErr != nil || len(trace.Frames) != 4 || trace.Frames[0].Inputs[0] != AigTernaryOne {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func TestABCXSimErrorsLimitsAndCancellation(t *testing.T) {
	network := abcXSimFixture(t, ABCInitZero)
	logic, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	for _, test := range []struct {
		network *ABCNetwork
		options ABCXSimOptions
		want    error
	}{
		{nil, ABCXSimOptions{}, ErrDimension},
		{logic, ABCXSimOptions{}, ErrDimension},
		{network, ABCXSimOptions{Frames: -1}, ErrDimension},
		{network, ABCXSimOptions{Frames: 2, MaxFrames: 1}, ErrLimit},
		{network, ABCXSimOptions{Frames: 1, MaxObjects: 1}, ErrLimit},
		{network, ABCXSimOptions{Frames: 2, MaxTraceEntries: 1}, ErrLimit},
		{network, ABCXSimOptions{Frames: 2, MaxWork: 1}, ErrLimit},
		{network, ABCXSimOptions{Frames: 2, Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, err := ABCXValueSimulate(test.network, test.options); len(result.Frames) != 0 || !errors.Is(err, test.want) {
			t.Fatalf("network=%p options=%+v result=%+v err=%v want=%v", test.network, test.options, result, err, test.want)
		}
	}
	bad := network.Clone()
	bad.objects[bad.boxes[0]].Init = ABCInitNone
	if result, err := ABCXValueSimulate(bad, ABCXSimOptions{Frames: 1}); len(result.Frames) != 0 || !errors.Is(err, ErrInvariant) {
		t.Fatalf("bad init result=%+v err=%v", result, err)
	}
	if result, err := ABCCycleInitState(network, ABCCycleInitStateOptions{Frames: -1}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("cycle dimension result=%v err=%v", result, err)
	}
}

func abcXMask(value AigTernary) uint8 {
	switch value {
	case AigTernaryZero:
		return 1
	case AigTernaryOne:
		return 2
	default:
		return 3
	}
}

func abcXMaskNot(value uint8) uint8 { return value&1<<1 | value&2>>1 }

func abcXMaskAnd(first, second uint8) uint8 {
	result := uint8(0)
	for a := 0; a < 2; a++ {
		for b := 0; b < 2; b++ {
			if first&(1<<uint(a)) != 0 && second&(1<<uint(b)) != 0 {
				result |= 1 << uint(a&b)
			}
		}
	}
	return result
}

func abcXMaskRead(values []uint8, edge ABCObjectEdge) uint8 {
	value := values[edge.ID]
	if edge.Complemented {
		value = abcXMaskNot(value)
	}
	return value
}

func abcXIndependent(network *ABCNetwork, frames int, unknownInputs, unknownState bool) [][]uint8 {
	values := make([]uint8, len(network.objects))
	values[0] = 2
	inputs, outputs, _ := abcLatchTerminals(network)
	state := make([]uint8, len(outputs))
	for i, latch := range network.boxes {
		state[i] = abcXMask(AigTernary(network.objects[latch].Init))
		if unknownState {
			state[i] = 3
		}
		values[outputs[i]] = state[i]
	}
	random := newABCXSimRandom()
	result := make([][]uint8, frames)
	for frame := 0; frame < frames; frame++ {
		for _, id := range network.pis {
			if unknownInputs {
				values[id] = 3
			} else {
				values[id] = abcXMask(abcXSimRandom2(&random))
			}
		}
		for id, object := range network.objects {
			if object.Type == ABCObjectNode {
				values[id] = abcXMaskAnd(abcXMaskRead(values, object.Fanins[0]), abcXMaskRead(values, object.Fanins[1]))
			}
		}
		for _, id := range network.cos {
			values[id] = abcXMaskRead(values, network.objects[id].Fanins[0])
		}
		row := append([]uint8(nil), state...)
		for _, id := range network.pos {
			row = append(row, values[id])
		}
		result[frame] = row
		for i, id := range inputs {
			state[i] = values[id]
			values[outputs[i]] = state[i]
		}
	}
	return result
}

func FuzzABCXValueSimulation(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, uint8(3), true, false)
	f.Add([]byte{}, uint8(1), false, true)
	f.Fuzz(func(t *testing.T, data []byte, frameByte uint8, unknownInputs, unknownState bool) {
		network, err := NewABCNetwork(ABCNetworkStrash, ABCFunctionAIG, ABCNetworkLimits{MaxObjects: 128, MaxEdges: 512})
		if err != nil {
			t.Fatal(err)
		}
		piCount := 1
		if len(data) > 0 {
			piCount += int(data[0] % 3)
		}
		signals := make([]int, 0, 24)
		for i := 0; i < piCount; i++ {
			id, _ := network.CreatePI()
			signals = append(signals, id)
		}
		init := ABCInitZero
		if len(data) > 1 {
			init = ABCLatchInit(1 + data[1]%3)
		}
		bi, _, bo, createErr := network.CreateLatchBox(init)
		if createErr != nil {
			t.Fatal(createErr)
		}
		signals = append(signals, bo)
		for offset := 2; offset+2 < len(data) && len(signals) < 20; offset += 3 {
			first := ABCObjectEdge{ID: signals[int(data[offset])%len(signals)], Complemented: data[offset+2]&1 != 0}
			second := ABCObjectEdge{ID: signals[int(data[offset+1])%len(signals)], Complemented: data[offset+2]&2 != 0}
			id, nodeErr := network.CreateLogicNode("11 1\n", first, second)
			if nodeErr != nil {
				t.Fatal(nodeErr)
			}
			signals = append(signals, id)
		}
		next := ABCObjectEdge{ID: signals[len(signals)-1], Complemented: len(data)&1 != 0}
		if err = network.AddFanin(bi, next); err != nil {
			t.Fatal(err)
		}
		po, _ := network.CreatePO()
		root := ABCObjectEdge{ID: signals[(len(data)+piCount)%len(signals)], Complemented: frameByte&0x80 != 0}
		if err = network.AddFanin(po, root); err != nil {
			t.Fatal(err)
		}
		frames := 1 + int(frameByte%4)
		trace, err := ABCXValueSimulate(network, ABCXSimOptions{Frames: frames, UnknownInputs: unknownInputs, UnknownState: unknownState, MaxObjects: 128, MaxTraceEntries: 128, MaxWork: 10000})
		if err != nil {
			t.Fatal(err)
		}
		want := abcXIndependent(network, frames, unknownInputs, unknownState)
		for frame, row := range trace.Frames {
			got := make([]uint8, 0, len(row.Latches)+len(row.Outputs))
			for _, value := range append(append([]AigTernary(nil), row.Latches...), row.Outputs...) {
				got = append(got, abcXMask(value))
			}
			if !slices.Equal(got, want[frame]) {
				t.Fatalf("frame=%d got=%v want=%v", frame, got, want[frame])
			}
		}
	})
}
