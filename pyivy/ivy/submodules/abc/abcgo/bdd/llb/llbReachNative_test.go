package llb

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSymbolicImageRenamesNextStateToCurrent(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state)       // property output
	g.AddOutput(state.Not()) // next state toggles
	aig.Aig_ManSetRegNum(g, 1)
	dd := core.NewBDDManager(2)
	defer core.Cudd_Quit(dd)
	manager := &Llb_Man_t_{pAig: g, dd: dd}
	initial := core.Cudd_Not(core.Cudd_bddIthVar(dd, 0))
	image := Llb_ManComputeImage(manager, initial, 0)
	if image != core.Cudd_bddIthVar(dd, 0) {
		t.Fatal("toggle image of state=0 should be current-state variable state=1")
	}
}

func TestNonlinearStateHelpersUseOrderedLatchVariables(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state)       // property output
	g.AddOutput(state.Not()) // next state
	aig.Aig_ManSetRegNum(g, 1)
	order := vec.Vec_IntStartFull(int(aig.Aig_ManObjNumMax(g)))
	for id := 0; id < vec.Vec_IntSize(order); id++ {
		vec.Vec_IntWriteEntry(order, id, int32(id))
	}
	dd := core.NewBDDManager(int(aig.Aig_ManObjNumMax(g)))
	defer core.Cudd_Quit(dd)
	initial := Llb_Nonlin4ComputeInitState(dd, g, order, 0)
	for _, test := range []struct{ value, want bool }{{false, true}, {true, false}} {
		assignment := make([]bool, len(dd.Vars))
		assignment[state.NodeID()] = test.value
		got, err := dd.Evaluate(initial, assignment)
		if err != nil || got != test.want {
			t.Fatalf("initial(state=%t)=%t, %v; want %t", test.value, got, err, test.want)
		}
	}
	quantify := Llb_Nonlin4CreateVars2Q(dd, g, order, 0)
	nextID := int(g.Outputs()[1].NodeID())
	if vec.Vec_IntEntry(quantify, nextID) != 0 || vec.Vec_IntEntry(quantify, int(state.NodeID())) != 1 {
		t.Fatal("forward quantification must retain next-state variables")
	}
	values := make([]byte, len(dd.Vars))
	values[state.NodeID()] = 1
	cube := Llb_Nonlin4ComputeCube(dd, g, order, string(values), 0)
	assignment := make([]bool, len(dd.Vars))
	assignment[nextID] = true
	if got, err := dd.Evaluate(cube, assignment); err != nil || !got {
		t.Fatalf("next-state cube evaluation = %t, %v; want true", got, err)
	}
	var stateWords [1]uint32
	Llb_Nonlin4RecordState(g, order, stateWords[:], string(values), 0)
	if stateWords[0]&1 == 0 {
		t.Fatal("state recorder did not preserve the selected state bit")
	}
}

func TestReachabilityFindsBadStateOnSecondFrame(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state)       // bad when state is one
	g.AddOutput(state.Not()) // toggle from initial zero to one
	aig.Aig_ManSetRegNum(g, 1)
	params := &Gia_ParLlb_t{nIterMax: 4, nBddMax: 100, fSilent: 1}
	manager := &Llb_Man_t_{pAig: g, pAigGlo: g, pPars: params}
	if got := Llb_ManReachability(manager, nil, nil); got != 0 {
		t.Fatalf("reachability returned %d; want reachable (0)", got)
	}
	cex := manager.pAigGlo.SequentialCounterexample()
	if cex == nil || Llb4_Nonlin4VerifyCex(g, cex) == nil {
		t.Fatal("expected a replayable sequential counterexample")
	}
}

func TestReachabilityProvesConstantSafeState(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state) // bad when state is one
	g.AddOutput(state) // state remains zero
	aig.Aig_ManSetRegNum(g, 1)
	params := &Gia_ParLlb_t{nIterMax: 4, nBddMax: 100, fSilent: 1}
	manager := &Llb_Man_t_{pAig: g, pPars: params}
	if got := Llb_ManReachability(manager, nil, nil); got != 1 {
		t.Fatalf("safe fixed point returned %d; want safe (1)", got)
	}
}

func TestReachableStatesBuildsAIGForTheFixedPoint(t *testing.T) {
	g := aig.NewGraph()
	state := g.NewInput()
	g.AddOutput(state) // safe because the initial state is zero
	g.AddOutput(state) // self-loop
	aig.Aig_ManSetRegNum(g, 1)
	reachable := Llb_ReachableStates(g)
	if reachable == nil {
		t.Fatal("expected an AIG for the reachable state set")
	}
	defer aig.Aig_ManStop(reachable)
	for _, test := range []struct{ state, want bool }{{false, true}, {true, false}} {
		values, err := reachable.Evaluate([]bool{test.state})
		if err != nil || len(values) != 1 || values[0] != test.want {
			t.Fatalf("reachable(%t) = %v, %v; want %t", test.state, values, err, test.want)
		}
	}
}
