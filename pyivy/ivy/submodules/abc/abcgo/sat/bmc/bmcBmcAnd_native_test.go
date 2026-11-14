package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestGiaManBmcPerformAdapterStoresCounterexample(t *testing.T) {
	graph := aig.NewGraph()
	pi, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(state)
	graph.AddOutput(pi)
	aig.Aig_ManSetRegNum(graph, 1)
	g := gia.Gia_ManFromAig(graph)
	params := &Bmc_AndPar_t{NFramesMax: 2}
	if result := Gia_ManBmcPerform(g, params); result != 0 {
		t.Fatalf("result=%d, want counterexample (0)", result)
	}
	cex := g.SequentialCounterexample()
	if cex == nil || cex.IFrame != 1 || cex.IPo != 0 {
		t.Fatalf("counterexample=%+v, want frame 1 property 0", cex)
	}
}

func TestProjectAIGPropertiesRetainsRegisterInterface(t *testing.T) {
	graph := aig.NewGraph()
	pi, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(state.Not())
	graph.AddOutput(state)
	graph.AddOutput(pi)
	aig.Aig_ManSetRegNum(graph, 1)
	projected, err := projectAIGProperties(graph, []int{0})
	if err != nil {
		t.Fatal(err)
	}
	if projected.InputCount() != 2 || projected.OutputCount() != 2 || projected.RegisterCount() != 1 {
		t.Fatalf("projected terminals = (%d,%d,%d), want (2 inputs, 1 property + 1 register output, 1 register)", projected.InputCount(), projected.OutputCount(), projected.RegisterCount())
	}
	if got := projected.Node(projected.Outputs()[0]).Fanins[0]; !got.Complemented() {
		t.Fatalf("projected property driver=%v, want complement of state", got)
	}
}

func TestGiaManMultiProveAigStoresPerOutputWitness(t *testing.T) {
	graph := aig.NewGraph()
	pi, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(state)
	graph.AddOutput(graph.Const0())
	graph.AddOutput(pi)
	aig.Aig_ManSetRegNum(graph, 1)
	params := &Bmc_MulPar_t{}
	results := Gia_ManMultiProveAig(graph, params)
	if results == nil || vec.Vec_PtrSize(results) != 2 {
		t.Fatalf("result slots=%v, want 2 properties", results)
	}
	if model, ok := vec.Vec_PtrEntry(results, 0).(*util.Abc_Cex_t); !ok || model == nil || model.IPo != 0 {
		t.Fatalf("output 0 witness=%v, want counterexample", vec.Vec_PtrEntry(results, 0))
	}
	if model := vec.Vec_PtrEntry(results, 1); model != nil {
		t.Fatalf("output 1 witness=%v, want nil for invariant", model)
	}
}

func TestBmcChainTestChecksEachPropertyAndKeepsOutputIndex(t *testing.T) {
	graph := aig.NewGraph()
	pi, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(state)
	graph.AddOutput(graph.Const0())
	graph.AddOutput(pi)
	aig.Aig_ManSetRegNum(graph, 1)
	manager := gia.Gia_ManFromAig(graph)
	var witnesses *vec.Vec_Ptr_t
	if status := Bmc_ChainTest(manager, 2, 0, 0, 0, &witnesses); status != 0 {
		t.Fatalf("status=%d, want 0", status)
	}
	if vec.Vec_PtrSize(witnesses) != 2 {
		t.Fatalf("witness slots=%d, want 2", vec.Vec_PtrSize(witnesses))
	}
	model, ok := vec.Vec_PtrEntry(witnesses, 0).(*util.Abc_Cex_t)
	if !ok || model == nil || model.IPo != 0 {
		t.Fatalf("output 0 witness=%v, want counterexample for output 0", vec.Vec_PtrEntry(witnesses, 0))
	}
	if model := vec.Vec_PtrEntry(witnesses, 1); model != nil {
		t.Fatalf("output 1 witness=%v, want nil for invariant", model)
	}
}

func TestGiaManMultiProveStoresOutputIndexedWitnesses(t *testing.T) {
	graph := aig.NewGraph()
	pi, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(state)
	graph.AddOutput(graph.Const0())
	graph.AddOutput(pi)
	aig.Aig_ManSetRegNum(graph, 1)
	manager := gia.Gia_ManFromAig(graph)
	if status := Gia_ManMultiProve(manager, &Bmc_MulPar_t{}); status != 0 {
		t.Fatalf("status=%d, want counterexample status 0", status)
	}
	witnesses := manager.SequentialCounterexamples()
	if len(witnesses) != 2 || witnesses[0] == nil || witnesses[0].IPo != 0 || witnesses[1] != nil {
		t.Fatalf("witnesses=%+v, want [output-0 counterexample, nil]", witnesses)
	}
}

func TestGiaManBmcPerformAdapterProvesZeroStateInvariant(t *testing.T) {
	graph := aig.NewGraph()
	_, state := graph.NewInput(), graph.NewInput()
	graph.AddOutput(graph.Const0())
	graph.AddOutput(state)
	aig.Aig_ManSetRegNum(graph, 1)
	g := gia.Gia_ManFromAig(graph)
	params := &Bmc_AndPar_t{NFramesMax: 2}
	if result := Gia_ManBmcPerform(g, params); result != 1 {
		t.Fatalf("result=%d, want bounded-safe (1)", result)
	}
	if cex := g.SequentialCounterexample(); cex != nil {
		t.Fatalf("unexpected counterexample: %+v", cex)
	}
}
