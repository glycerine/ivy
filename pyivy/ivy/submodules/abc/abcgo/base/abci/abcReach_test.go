package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func makeSelfLoopReachabilityNetwork(t *testing.T, initial ...abc.Abc_InitType_t) *abc.Network {
	t.Helper()
	init := abc.ABC_INIT_ZERO
	if len(initial) != 0 {
		init = initial[0]
	}
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	pi := n.NewObject(abc.ObjectPI)
	pi.SetName("input")
	bo := abc.Abc_NtkAddLatch(n, nil, init)
	bo.SetName("state")
	bi := n.ObjectsOfKind(abc.ObjectBI)[0]
	if err := n.AddFanin(bi, abc.Edge{Object: bo}); err != nil {
		t.Fatal(err)
	}
	po := n.NewObject(abc.ObjectPO)
	po.SetName("miter")
	if err := n.AddFanin(po, abc.Edge{Object: bo}); err != nil {
		t.Fatal(err)
	}
	abc.Abc_NtkOrderCisCos(n)
	return n
}

func TestInitStateOneAndDontCare(t *testing.T) {
	for _, tc := range []struct {
		name string
		init abc.Abc_InitType_t
		want bool
	}{
		{name: "one", init: abc.ABC_INIT_ONE, want: true},
		{name: "dont-care defaults to zero", init: abc.ABC_INIT_DC, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			n := makeSelfLoopReachabilityNetwork(t, tc.init)
			dd := Abc_NtkBuildGlobalBdds(n, 1000, 1, 0, 0, 0)
			if dd == nil {
				t.Fatal("global BDD construction failed")
			}
			defer Abc_NtkFreeGlobalBdds(n, 1)
			initial := Abc_NtkInitStateVarMap(dd, n, 0)
			if initial == nil {
				t.Fatal("initial state construction failed")
			}
			assignment := make([]bool, len(dd.Vars))
			assignment[len(n.ObjectsOfKind(abc.ObjectPI))] = tc.want
			got, err := dd.Evaluate(initial, assignment)
			if err != nil {
				t.Fatal(err)
			}
			if !got {
				t.Errorf("initial state excludes expected value %v", tc.want)
			}
			assignment[len(n.ObjectsOfKind(abc.ObjectPI))] = !tc.want
			got, err = dd.Evaluate(initial, assignment)
			if err != nil {
				t.Fatal(err)
			}
			if got {
				t.Errorf("initial state includes unexpected value %v", !tc.want)
			}
		})
	}
}

func TestSequentialReachTransitionAndUnreachableDcs(t *testing.T) {
	n := makeSelfLoopReachabilityNetwork(t)
	dd := Abc_NtkBuildGlobalBdds(n, 1000, 1, 0, 0, 0)
	if dd == nil {
		t.Fatal("global BDD construction failed")
	}
	defer Abc_NtkFreeGlobalBdds(n, 1)
	transition := Abc_NtkTransitionRelation(dd, n, 0)
	if transition == nil {
		t.Fatal("transition relation construction failed")
	}
	core.Cudd_Ref(transition)
	defer core.Cudd_RecursiveDeref(dd, transition)
	initial := Abc_NtkInitStateAndVarMap(dd, n, 0)
	if initial == nil {
		t.Fatal("initial-state construction failed")
	}
	if got := Abc_ObjGlobalBdd(n.ObjectsOfKind(abc.ObjectBI)[0]); got != nil {
		t.Fatal("transition relation should release per-object global BDDs")
	}
	piCount := len(n.ObjectsOfKind(abc.ObjectPI))
	current, next := piCount, len(n.Inputs())
	for state := 0; state < 2; state++ {
		for successor := 0; successor < 2; successor++ {
			assignment := make([]bool, len(dd.Vars))
			assignment[current], assignment[next] = state != 0, successor != 0
			got, err := dd.Evaluate(transition, assignment)
			if err != nil {
				t.Fatal(err)
			}
			if got != (state == successor) {
				t.Errorf("transition %d -> %d = %v", state, successor, got)
			}
		}
	}
	unreachable := Abc_NtkComputeUnreachable(dd, n, transition, initial, 0)
	if unreachable == nil {
		t.Fatal("fixed-point computation failed")
	}
	for state := 0; state < 2; state++ {
		assignment := make([]bool, len(dd.Vars))
		assignment[current] = state != 0
		got, err := dd.Evaluate(unreachable, assignment)
		if err != nil {
			t.Fatal(err)
		}
		if got != (state == 1) {
			t.Errorf("unreachable(%d) = %v", state, got)
		}
	}
	core.Cudd_Ref(unreachable)
	exdc := Abc_NtkConstructExdc(dd, n, unreachable)
	core.Cudd_RecursiveDeref(dd, unreachable)
	if exdc == nil {
		t.Fatal("EXDC network construction failed")
	}
	defer abc.Abc_NtkDelete(exdc)
	if len(exdc.Inputs()) != 1 || len(exdc.Outputs()) != 1 {
		t.Fatalf("EXDC dimensions = %d inputs, %d outputs", len(exdc.Inputs()), len(exdc.Outputs()))
	}
	for state := 0; state < 2; state++ {
		values := map[*abc.Object]bool{exdc.Inputs()[0]: state != 0}
		got, err := exdc.Evaluate(exdc.Outputs()[0].Fanins()[0], values)
		if err != nil {
			t.Fatal(err)
		}
		if got != (state == 1) {
			t.Errorf("EXDC(%d) = %v", state, got)
		}
	}
}

func TestExtractSequentialDcsAttachesExdc(t *testing.T) {
	n := makeSelfLoopReachabilityNetwork(t)
	if got := Abc_NtkExtractSequentialDcs(n, 0); got != 1 {
		t.Fatalf("Abc_NtkExtractSequentialDcs() = %d", got)
	}
	if n.Exdc == nil {
		t.Fatal("sequential unreachable states were not attached as EXDC")
	}
	defer abc.Abc_NtkDelete(n.Exdc)
	if got := len(n.Exdc.Inputs()); got != 1 {
		t.Fatalf("EXDC input count = %d, want 1", got)
	}
	if got := len(n.Exdc.Outputs()); got != 1 {
		t.Fatalf("EXDC output count = %d, want 1", got)
	}
	for state := 0; state < 2; state++ {
		values := map[*abc.Object]bool{n.Exdc.Inputs()[0]: state != 0}
		got, err := n.Exdc.Evaluate(n.Exdc.Outputs()[0].Fanins()[0], values)
		if err != nil {
			t.Fatal(err)
		}
		if got != (state == 1) {
			t.Errorf("EXDC(%d) = %v, want %v", state, got, state == 1)
		}
	}
}

func TestComputeReachablePartitionedAndMonolithic(t *testing.T) {
	for _, partitioned := range []int32{0, 1} {
		t.Run(map[int32]string{0: "monolithic", 1: "partitioned"}[partitioned], func(t *testing.T) {
			n := makeSelfLoopReachabilityNetwork(t)
			dd := Abc_NtkBuildGlobalBdds(n, 1000, 1, 0, 0, 0)
			if dd == nil {
				t.Fatal("global BDD construction failed")
			}
			defer Abc_NtkFreeGlobalBdds(n, 1)
			output := Abc_ObjGlobalBdd(n.ObjectsOfKind(abc.ObjectPO)[0])
			if output == nil {
				t.Fatal("output BDD missing")
			}
			core.Cudd_Ref(output)
			defer core.Cudd_RecursiveDeref(dd, output)
			parts := Abc_NtkCreatePartitions(dd, n, 0, 0)
			if parts == nil {
				t.Fatal("partition construction failed")
			}
			defer func() {
				for _, part := range parts {
					core.Cudd_RecursiveDeref(dd, part)
				}
			}()
			initial := Abc_NtkInitStateVarMap(dd, n, 0)
			if initial == nil {
				t.Fatal("initial state construction failed")
			}
			reached := Abc_NtkComputeReachable(dd, n, parts, initial, output, 1000, 10, partitioned, 0, 0)
			if reached == nil {
				t.Fatal("reachable-state computation failed")
			}
			core.Cudd_Ref(reached)
			defer core.Cudd_RecursiveDeref(dd, reached)
			for state := 0; state < 2; state++ {
				assignment := make([]bool, len(dd.Vars))
				assignment[len(n.ObjectsOfKind(abc.ObjectPI))] = state != 0
				got, err := dd.Evaluate(reached, assignment)
				if err != nil {
					t.Fatal(err)
				}
				if got != (state == 0) {
					t.Errorf("reached(%d) = %v, want %v", state, got, state == 0)
				}
			}
		})
	}
}
