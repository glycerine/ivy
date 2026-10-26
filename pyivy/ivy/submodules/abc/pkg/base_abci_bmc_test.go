package abc

// Standalone integration and fuzz tests for src/base/abci/abcBmc.c.

import (
	"errors"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
)

func abcIvyBMCFixture(t testing.TB, inputs int) *ABCNetwork {
	t.Helper()
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{MaxObjects: 1000, MaxEdges: 4000})
	if err != nil {
		t.Fatal(err)
	}
	pis := make([]int, inputs)
	for i := range pis {
		pis[i], _ = network.CreatePI()
	}
	bi, _, bo, err := network.CreateLatchBox(ABCInitZero)
	if err != nil {
		t.Fatal(err)
	}
	fanins := []ABCObjectEdge{{ID: bo}, {ID: pis[0]}}
	node, err := network.CreateLogicNode("11 1\n", fanins...)
	if err != nil {
		t.Fatal(err)
	}
	po, _ := network.CreatePO()
	if err = network.AddFanin(po, ABCObjectEdge{ID: node}); err != nil {
		t.Fatal(err)
	}
	next := pis[len(pis)-1]
	if err = network.AddFanin(bi, ABCObjectEdge{ID: next}); err != nil {
		t.Fatal(err)
	}
	if err = network.Validate(); err != nil {
		t.Fatal(err)
	}
	return network
}

func TestABCIvyBMCFramesFraigAndReport(t *testing.T) {
	network := abcIvyBMCFixture(t, 2)
	before := network.Clone()
	result, err := ABCIvyBMC(network, ABCIvyBMCOptions{Frames: 3, Initialize: false, MaxFrameObjects: 10000, MaxMappingEntries: 100000, Fraig: ivyFraigUnlimited(false)})
	if err != nil {
		t.Fatal(err)
	}
	if result.Frames == nil || result.Fraig.Graph == nil || len(result.PerFrame) != 3 || result.FrameNodeCount < 1 || result.FraigNodeCount < 1 {
		t.Fatalf("result=%+v", result)
	}
	frameGraph := result.Frames.Graph()
	if frameGraph.InputCount() != result.Fraig.Graph.Count(IvyPI) || frameGraph.OutputCount() != len(result.Fraig.Graph.Outputs()) {
		t.Fatalf("terminals frames=%d/%d fraig=%d/%d", frameGraph.InputCount(), frameGraph.OutputCount(), result.Fraig.Graph.Count(IvyPI), len(result.Fraig.Graph.Outputs()))
	}
	for assignment := 0; assignment < 1<<uint(frameGraph.InputCount()); assignment++ {
		values := make([]bool, frameGraph.InputCount())
		for i := range values {
			values[i] = assignment&(1<<uint(i)) != 0
		}
		want, frameErr := frameGraph.Evaluate(values)
		got, fraigErr := result.Fraig.Graph.Evaluate(values, nil)
		if frameErr != nil || fraigErr != nil || !slices.Equal(got, want) {
			t.Fatalf("assignment=%d got=%v want=%v frameErr=%v fraigErr=%v", assignment, got, want, frameErr, fraigErr)
		}
	}
	if !slices.Equal(network.cis, before.cis) || !slices.Equal(network.cos, before.cos) || network.edges != before.edges {
		t.Fatal("source network mutated")
	}
	for frame, row := range result.PerFrame {
		if row.Frame != frame || row.FraigNodes > row.StrashedNodes {
			t.Fatalf("frame report %+v", row)
		}
	}
}

func TestABCIvyBMCInitializationLimitsAndCancellation(t *testing.T) {
	network := abcIvyBMCFixture(t, 1)
	initialized, err := ABCIvyBMC(network, ABCIvyBMCOptions{Frames: 2, Initialize: true, Fraig: ivyFraigUnlimited(false)})
	if err != nil {
		t.Fatal(err)
	}
	if initialized.Frames.Graph().InputCount() != 2 || initialized.Frames.Graph().OutputCount() != 2 {
		t.Fatalf("initialized terminals %d/%d", initialized.Frames.Graph().InputCount(), initialized.Frames.Graph().OutputCount())
	}
	for _, options := range []ABCIvyBMCOptions{
		{}, {Frames: -1}, {Frames: 1, MaxFrameObjects: -1}, {Frames: 1, MaxMappingEntries: -1}, {Frames: 1, MaxReportEntries: -1},
	} {
		if result, gotErr := ABCIvyBMC(network, options); result.Frames != nil || !errors.Is(gotErr, ErrDimension) {
			t.Fatalf("options=%+v result=%+v err=%v", options, result, gotErr)
		}
	}
	for _, options := range []ABCIvyBMCOptions{
		{Frames: 2, MaxFrameObjects: 1}, {Frames: 2, MaxMappingEntries: 1}, {Frames: 2, MaxReportEntries: 1},
	} {
		if result, gotErr := ABCIvyBMC(network, options); result.Frames != nil || !errors.Is(gotErr, ErrLimit) {
			t.Fatalf("limited options=%+v result=%+v err=%v", options, result, gotErr)
		}
	}
	if result, gotErr := ABCIvyBMC(network, ABCIvyBMCOptions{Frames: 2, Stop: func() bool { return true }}); result.Frames != nil || !errors.Is(gotErr, ErrCancelled) {
		t.Fatalf("cancel result=%+v err=%v", result, gotErr)
	}
	before := network.Clone()
	var wait sync.WaitGroup
	var failures atomic.Int32
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, runErr := ABCIvyBMC(network, ABCIvyBMCOptions{Frames: 2, Fraig: ivyFraigUnlimited(false)})
			if runErr != nil || result.Frames == nil || result.Fraig.Graph == nil {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 || !slices.Equal(network.cis, before.cis) || !slices.Equal(network.cos, before.cos) || network.edges != before.edges {
		t.Fatalf("concurrent failures=%d or source mutation", failures.Load())
	}
}

func FuzzABCIvyBMC(f *testing.F) {
	f.Add(uint8(2), uint8(2), false)
	f.Add(uint8(1), uint8(1), true)
	f.Fuzz(func(t *testing.T, inputByte, frameByte uint8, initialize bool) {
		inputs := 1 + int(inputByte%2)
		frames := 1 + int(frameByte%2)
		network := abcIvyBMCFixture(t, inputs)
		result, err := ABCIvyBMC(network, ABCIvyBMCOptions{Frames: frames, Initialize: initialize, MaxFrameObjects: 1000, MaxMappingEntries: 10000, Fraig: ivyFraigUnlimited(false)})
		if err != nil {
			t.Fatal(err)
		}
		frameGraph := result.Frames.Graph()
		if frameGraph.InputCount() > 8 {
			t.Fatal("fuzz bound")
		}
		for assignment := 0; assignment < 1<<uint(frameGraph.InputCount()); assignment++ {
			values := make([]bool, frameGraph.InputCount())
			for i := range values {
				values[i] = assignment&(1<<uint(i)) != 0
			}
			want, frameErr := frameGraph.Evaluate(values)
			got, fraigErr := result.Fraig.Graph.Evaluate(values, nil)
			if frameErr != nil || fraigErr != nil || !slices.Equal(got, want) {
				t.Fatalf("assignment=%d got=%v want=%v frameErr=%v fraigErr=%v", assignment, got, want, frameErr, fraigErr)
			}
		}
	})
}
