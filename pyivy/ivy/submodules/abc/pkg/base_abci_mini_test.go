package abc

// Standalone semantic, frame, concurrency and fuzz tests for abcMini.c.

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func abcMiniSequentialFixture(t testing.TB) *MiniAIG {
	t.Helper()
	mini, err := NewMiniAIG(MiniNetworkLimits{MaxNodes: 1000})
	if err != nil {
		t.Fatal(err)
	}
	input, _ := mini.CreatePI()
	state, _ := mini.CreatePI()
	output, _ := mini.AND(input, state.Not())
	next, _ := mini.XOR(input, state)
	mini.CreatePO(output)
	mini.CreatePO(next)
	if err = mini.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err = mini.Check(); err != nil {
		t.Fatal(err)
	}
	return mini
}

func abcMiniCompareAll(t testing.TB, first, second *MiniAIG) {
	t.Helper()
	firstPIs, firstPOs, _ := first.Counts()
	secondPIs, secondPOs, _ := second.Counts()
	if firstPIs != secondPIs || firstPOs != secondPOs || first.RegisterCount() != second.RegisterCount() {
		t.Fatalf("shape first=%d/%d/%d second=%d/%d/%d", firstPIs, firstPOs, first.RegisterCount(), secondPIs, secondPOs, second.RegisterCount())
	}
	for assignment := 0; assignment < 1<<uint(firstPIs); assignment++ {
		inputs := make([]bool, firstPIs)
		for i := range inputs {
			inputs[i] = assignment&(1<<uint(i)) != 0
		}
		want, firstErr := first.Evaluate(inputs)
		got, secondErr := second.Evaluate(inputs)
		if firstErr != nil || secondErr != nil || !slices.Equal(got, want) {
			t.Fatalf("assignment=%d got=%v want=%v firstErr=%v secondErr=%v", assignment, got, want, firstErr, secondErr)
		}
	}
}

func TestABCMiniAIGDefinitionsConversionsAndNames(t *testing.T) {
	if len(ABCMiniAIGDefinitions) != 10 || ABCMiniAIGDefinitions[0] != "Abc_NodeFanin0Copy" || ABCMiniAIGDefinitions[9] != "Abc_NtkToMiniAig" {
		t.Fatalf("definitions=%v", ABCMiniAIGDefinitions)
	}
	source := abcMiniSequentialFixture(t)
	sourceNodes := slices.Clone(source.nodes)
	network, err := ABCNetworkFromMiniAIG(source, ABCMiniAIGOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if network.Name != "MiniAig" || len(network.pis) != 1 || len(network.pos) != 1 || len(network.boxes) != 1 || network.objects[network.pis[0]].Name != "pi0" || network.objects[network.pos[0]].Name != "po0" {
		t.Fatalf("network name/shape=%q %d/%d/%d", network.Name, len(network.pis), len(network.pos), len(network.boxes))
	}
	if !slices.Equal(source.nodes, sourceNodes) {
		t.Fatal("source MiniAIG mutated")
	}
	round, err := ABCNetworkToMiniAIG(network, ABCMiniAIGOptions{})
	if err != nil {
		t.Fatal(err)
	}
	abcMiniCompareAll(t, source, round)
	converted, err := ABCNetworkToAIG(network, ABCFunctionalOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for primary := 0; primary < 2; primary++ {
		for state := 0; state < 2; state++ {
			outputs, next, stepErr := converted.Graph.Step([]bool{primary != 0}, []bool{state != 0})
			miniOutputs, evalErr := source.Evaluate([]bool{primary != 0, state != 0})
			if stepErr != nil || evalErr != nil || !slices.Equal(append(outputs, next...), miniOutputs) {
				t.Fatalf("primary=%d state=%d step=%v/%v mini=%v errors=%v/%v", primary, state, outputs, next, miniOutputs, stepErr, evalErr)
			}
		}
	}
}

func TestABCMiniAIGRoundTripFrameAndFlopNumber(t *testing.T) {
	source := abcMiniSequentialFixture(t)
	network, _ := ABCNetworkFromMiniAIG(source, ABCMiniAIGOptions{})
	roundNetwork, data, err := ABCMiniAIGRoundTrip(network, ABCMiniAIGOptions{})
	if err != nil || len(data) == 0 {
		t.Fatalf("round data=%d err=%v", len(data), err)
	}
	roundMini, err := ABCNetworkToMiniAIG(roundNetwork, ABCMiniAIGOptions{})
	if err != nil {
		t.Fatal(err)
	}
	abcMiniCompareAll(t, source, roundMini)

	frame := NewABCMiniAIGFrame()
	if _, err = frame.Output(ABCMiniAIGOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatalf("empty output err=%v", err)
	}
	if err = frame.Input(source, ABCMiniAIGOptions{}); err != nil {
		t.Fatal(err)
	}
	frameOutput, err := frame.Output(ABCMiniAIGOptions{})
	if err != nil {
		t.Fatal(err)
	}
	abcMiniCompareAll(t, source, frameOutput)
	snapshot, err := frame.Network()
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Name = "changed"
	again, _ := frame.Network()
	if again.Name == "changed" {
		t.Fatal("frame exposed network ownership")
	}

	combinational := source.Clone()
	if err = combinational.SetRegisterCount(0); err != nil {
		t.Fatal(err)
	}
	combNetwork, _ := ABCNetworkFromMiniAIG(combinational, ABCMiniAIGOptions{})
	sequential, err := ABCSetFlopNumber(combNetwork, 1, ABCMiniAIGOptions{})
	if err != nil || len(sequential.boxes) != 1 || len(sequential.pis) != 1 || len(sequential.pos) != 1 {
		t.Fatalf("set flops shape=%v/%v/%v err=%v", len(sequential.boxes), len(sequential.pis), len(sequential.pos), err)
	}
	combFrame := NewABCMiniAIGFrame()
	if err = combFrame.Input(combinational, ABCMiniAIGOptions{}); err != nil {
		t.Fatal(err)
	}
	if err = combFrame.SetFlopNumber(1, ABCMiniAIGOptions{}); err != nil {
		t.Fatal(err)
	}
	frameNetwork, _ := combFrame.Network()
	if len(frameNetwork.boxes) != 1 {
		t.Fatalf("frame flop count=%d", len(frameNetwork.boxes))
	}
}

func TestABCMiniAIGLimitsErrorsAndConcurrency(t *testing.T) {
	source := abcMiniSequentialFixture(t)
	network, _ := ABCNetworkFromMiniAIG(source, ABCMiniAIGOptions{})
	badRegisters := source.Clone()
	badRegisters.registers = 99
	for _, test := range []struct {
		mini    *MiniAIG
		options ABCMiniAIGOptions
		want    error
	}{
		{nil, ABCMiniAIGOptions{}, ErrDimension},
		{badRegisters, ABCMiniAIGOptions{}, ErrDimension},
		{source, ABCMiniAIGOptions{MaxObjects: 1}, ErrLimit},
		{source, ABCMiniAIGOptions{MaxMappingEntries: 1}, ErrLimit},
		{source, ABCMiniAIGOptions{MaxWork: 1}, ErrLimit},
		{source, ABCMiniAIGOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, err := ABCNetworkFromMiniAIG(test.mini, test.options); result != nil || !errors.Is(err, test.want) {
			t.Fatalf("mini=%p options=%+v result=%v err=%v want=%v", test.mini, test.options, result, err, test.want)
		}
	}
	logic, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	for _, test := range []struct {
		network *ABCNetwork
		options ABCMiniAIGOptions
		want    error
	}{
		{nil, ABCMiniAIGOptions{}, ErrDimension},
		{logic, ABCMiniAIGOptions{}, ErrDimension},
		{network, ABCMiniAIGOptions{MaxObjects: 1}, ErrLimit},
		{network, ABCMiniAIGOptions{MaxMappingEntries: 1}, ErrLimit},
		{network, ABCMiniAIGOptions{MaxWork: 1}, ErrLimit},
		{network, ABCMiniAIGOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		if result, err := ABCNetworkToMiniAIG(test.network, test.options); result != nil || !errors.Is(err, test.want) {
			t.Fatalf("network=%p options=%+v result=%v err=%v want=%v", test.network, test.options, result, err, test.want)
		}
	}
	if result, err := ABCSetFlopNumber(network, 99, ABCMiniAIGOptions{}); result != nil || !errors.Is(err, ErrDimension) {
		t.Fatalf("flop limit result=%v err=%v", result, err)
	}
	if err := (*ABCMiniAIGFrame)(nil).Input(source, ABCMiniAIGOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("nil frame input err=%v", err)
	}
	frame := NewABCMiniAIGFrame()
	if err := frame.Input(source, ABCMiniAIGOptions{}); err != nil {
		t.Fatal(err)
	}
	var wait sync.WaitGroup
	var failures atomic.Int32
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			mini, runErr := ABCNetworkToMiniAIG(network, ABCMiniAIGOptions{})
			framed, frameErr := frame.Output(ABCMiniAIGOptions{})
			if runErr != nil || mini.RegisterCount() != 1 || frameErr != nil || framed.RegisterCount() != 1 {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func FuzzABCMiniAIGConversions(f *testing.F) {
	f.Add([]byte{2, 0, 1, 3, 5, 7}, uint8(1))
	f.Add([]byte{}, uint8(0))
	f.Fuzz(func(t *testing.T, data []byte, registerByte uint8) {
		mini, err := NewMiniAIG(MiniNetworkLimits{MaxNodes: 128})
		if err != nil {
			t.Fatal(err)
		}
		piCount := 1
		if len(data) > 0 {
			piCount += int(data[0] % 4)
		}
		signals := make([]AIGLiteral, piCount, piCount+16)
		for i := range signals {
			signals[i], _ = mini.CreatePI()
		}
		for offset := 1; offset+2 < len(data) && len(signals) < piCount+16; offset += 3 {
			first := signals[int(data[offset])%len(signals)].NotIf(data[offset+2]&1 != 0)
			second := signals[int(data[offset+1])%len(signals)].NotIf(data[offset+2]&2 != 0)
			node, nodeErr := mini.AND(first, second)
			if nodeErr != nil {
				t.Fatal(nodeErr)
			}
			signals = append(signals, node)
		}
		poCount := max(1, piCount)
		for i := 0; i < poCount; i++ {
			edge := signals[(i+len(data))%len(signals)].NotIf(i&1 != 0)
			if _, err = mini.CreatePO(edge); err != nil {
				t.Fatal(err)
			}
		}
		registers := int(registerByte) % (min(piCount, poCount) + 1)
		if err = mini.SetRegisterCount(registers); err != nil {
			t.Fatal(err)
		}
		network, err := ABCNetworkFromMiniAIG(mini, ABCMiniAIGOptions{MaxObjects: 256, MaxMappingEntries: 256, MaxWork: 10000})
		if err != nil {
			t.Fatal(err)
		}
		round, err := ABCNetworkToMiniAIG(network, ABCMiniAIGOptions{MaxObjects: 256, MaxMappingEntries: 256, MaxWork: 10000})
		if err != nil {
			t.Fatal(err)
		}
		abcMiniCompareAll(t, mini, round)
	})
}
