package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestAigGla1GateClauses(t *testing.T) {
	solver := bsat.NewSolver()
	a, b, z := solver.NewVariable(), solver.NewVariable(), solver.NewVariable()
	if Aig_Gla1AddNode(solver, z, a, b, 0, 0) == 0 {
		t.Fatal("failed to add AND encoding")
	}
	if solver.Solve(bsat.Positive(z)) != bsat.Satisfiable {
		t.Fatal("AND output should be satisfiable")
	}
	if value, ok := solver.Value(a); !ok || !value {
		t.Fatalf("a = %v, ok=%v; want true", value, ok)
	}
	if value, ok := solver.Value(b); !ok || !value {
		t.Fatalf("b = %v, ok=%v; want true", value, ok)
	}
}

func TestAigGla1BoundedCheckFindsReachableFailure(t *testing.T) {
	g := aig.NewGraph()
	input := g.NewInput()
	g.AddOutput(input)
	aig.Aig_ManSetRegNum(g, 0)
	var frame int32 = -1
	classes := Aig_Gla1ManPerform(g, nil, 0, 3, 0, 0, 1, 0, &frame)
	if classes != nil || frame != 0 {
		t.Fatalf("result=%v frame=%d, want SAT at frame 0", classes, frame)
	}
}

func TestAigGla1BoundedCheckProvesConstantSafeBound(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 0)
	var frame int32 = -1
	classes := Aig_Gla1ManPerform(g, nil, 0, 3, 0, 0, 1, 0, &frame)
	if classes == nil || frame != 3 {
		t.Fatalf("result=%v frame=%d, want full abstraction after 3 frames", classes, frame)
	}
	if got := vec.Vec_IntSize(classes); got != len(g.Nodes()) {
		t.Fatalf("class bitmap length = %d, want %d", got, len(g.Nodes()))
	}
	if got := vec.Vec_IntEntry(classes, 0); got != 1 {
		t.Fatalf("constant inclusion = %d, want 1", got)
	}
	for id := 1; id < len(g.Nodes()); id++ {
		if got := vec.Vec_IntEntry(classes, id); got != 0 {
			t.Fatalf("safe constant property unexpectedly included object %d", id)
		}
	}
}

func TestAigGla1RefinesAbstractLatchBeforeProvingSafe(t *testing.T) {
	g := aig.NewGraph()
	q := g.NewInput()
	g.AddOutput(q)          // property: latch is one
	g.AddOutput(g.Const0()) // latch transition: next state is zero
	aig.Aig_ManSetRegNum(g, 1)

	var frame int32 = -1
	classes := Aig_Gla1ManPerform(g, nil, 0, 3, 0, 0, 1, 0, &frame)
	if classes == nil {
		t.Fatal("expected abstraction refinement followed by a safe result")
	}
	if frame != 3 {
		t.Fatalf("frame = %d, want 3 after all three bounded frames", frame)
	}
	if got := vec.Vec_IntEntry(classes, int(q.NodeID())); got != 1 {
		t.Fatalf("latch inclusion bitmap[%d] = %d, want 1 after it refines the spurious frame-0 CEX", q.NodeID(), got)
	}
	for id := 0; id < len(g.Nodes()); id++ {
		want := int32(0)
		if id == 0 || id == int(q.NodeID()) {
			want = 1
		}
		if got := vec.Vec_IntEntry(classes, id); got != want {
			t.Fatalf("inclusion bitmap[%d] = %d, want %d", id, got, want)
		}
	}
}

func TestSaigManCbaFilterInputsKeepsReasonTraversalOrder(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	left := g.AND(a, b)
	g.AddOutput(g.AND(c, left))
	aig.Aig_ManSetRegNum(g, 0)
	cex := util.Abc_CexAlloc(0, 3, 1)
	cex.IPo, cex.IFrame = 0, 0
	for i := int32(0); i < 3; i++ {
		util.Abc_InfoSetBit(cex.PData, i)
	}

	got := Saig_ManCbaFilterInputs(g, 0, cex, 0)
	want := []int32{2, 0, 1}
	if vec.Vec_IntSize(got) != len(want) {
		t.Fatalf("reason inputs = %v, want %v", got, want)
	}
	for i, input := range want {
		if actual := vec.Vec_IntEntry(got, i); actual != input {
			t.Fatalf("reason input[%d] = %d, want %d (order matters for source behavior)", i, actual, input)
		}
	}
}
