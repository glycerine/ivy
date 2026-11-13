package framgr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestFraBmcRefinementRespectsFramesAndFraig(t *testing.T) {
	g := aig.NewGraph()
	input, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(input.Not())
	aig.Aig_ManSetRegNum(g, 1)
	b := Fra_BmcStart(g, 0, 2)
	b.pAigFrames = Fra_BmcFrames(b, 0)
	b.pObjToFraig = make([]aig.Edge, len(b.pAigFrames.Nodes()))
	for i := range b.pObjToFraig {
		b.pObjToFraig[i] = aig.Edge(i << 1)
	}
	p := &Fra_Man_t{pManAig: g, pBmc: b}
	if Fra_BmcNodeIsConst(p, state) != 0 {
		t.Fatal("second frame state is not constant")
	}
	if got := Bmc_ObjFrames(b, state, 0); got != b.pAigFrames.Const0() {
		t.Fatalf("initial state: %v", got)
	}
	if got := Bmc_ObjFrames(b, state, 1); got != Bmc_ObjFrames(b, input, 0).Not() {
		t.Fatalf("next state is not complemented input: %v", got)
	}
	mapped := Bmc_ObjFrames(b, state, 1)
	b.pObjToFraig[mapped.NodeID()] = b.pAigFrames.Const1()
	if Fra_BmcNodeIsConst(p, state) != 1 {
		t.Fatal("fraig representative not considered")
	}
	b.pObjToFraig[Bmc_ObjFrames(b, input, 1).NodeID()] = b.pAigFrames.Const1()
	b.nPref = 1
	if Fra_BmcNodesAreEqual(p, state, input) != 1 {
		t.Fatal("expected common fraig representative")
	}
}

func TestFraBmcFramesRejectsUnmappedNode(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	b := Fra_BmcStart(g, 0, 1)
	defer func() {
		if recover() == nil {
			t.Fatal("uninitialized mapping accepted")
		}
	}()
	Bmc_ObjFrames(b, x, 0)
}

func TestFraBmcFilterImplicationsSameRepresentative(t *testing.T) {
	original := aig.NewGraph()
	x, y := original.NewInput(), original.NewInput()
	frames := aig.NewGraph()
	fx, fy := frames.NewInput(), frames.NewInput()
	reduced := aig.NewGraph()
	rep := reduced.NewInput()
	b := Fra_BmcStart(original, 0, 1)
	b.pAigFrames = frames
	Bmc_ObjSetFrames(b, x, 0, fx)
	Bmc_ObjSetFrames(b, y, 0, fy)
	b.vImps = vec.Vec_IntAlloc(4)
	imp := Fra_ImpCreate(int32(x.NodeID()), int32(y.NodeID()))
	vec.Vec_IntPush(b.vImps, imp)
	p := &Fra_Man_t{pManAig: frames, pManFraig: reduced, nFramesAll: 1, pMemFraig: make([]aig.Edge, len(frames.Nodes()))}
	Fra_ObjSetFraig(p, fx, 0, rep)
	Fra_ObjSetFraig(p, fy, 0, rep)
	Fra_BmcFilterImplications(p, b)
	if vec.Vec_IntSize(b.vImps) != 1 {
		t.Fatal("equivalent implication removed")
	}
	Fra_ObjSetFraig(p, fy, 0, rep.Not())
	Fra_BmcFilterImplications(p, b)
	if vec.Vec_IntSize(b.vImps) != 0 {
		t.Fatal("complementary implication retained")
	}
}

func TestFraBmcPerformRefinesInitializedClasses(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	p := &Fra_Man_t{pManAig: g, pPars: &Fra_Par_t{}, pCla: Fra_ClassesStart(g)}
	Fra_ClassesTest(p.pCla, int32(x.NodeID()), int32(state.NodeID()))
	Fra_BmcPerform(p, 0, 1)
	if p.pBmc != nil {
		t.Fatal("BMC manager not cleared")
	}
	if p.pCla.reprs[state.NodeID()] >= 0 {
		t.Fatalf("invalid equivalence survived: %v", p.pCla.reprs)
	}
}

func TestFraBmcPerformSimpleFindsSequentialWitness(t *testing.T) {
	g := aig.NewGraph()
	x, y, state := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(g.AND(x, y))
	aig.Aig_ManSetRegNum(g, 1)
	Fra_BmcPerformSimple(g, 2, 1000, 0, 0)
	cex := g.SequentialCounterexample()
	if cex == nil || cex.IPo != 0 || cex.IFrame != 1 {
		t.Fatalf("missing sequential witness: %+v", cex)
	}
}

func TestFraBmcSimpleFindsDirectInputWitness(t *testing.T) {
	g := aig.NewGraph()
	input, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(input)
	aig.Aig_ManSetRegNum(g, 1)
	Fra_BmcPerformSimple(g, 2, 1000, 0, 0)
	cex := g.SequentialCounterexample()
	if cex == nil || cex.IPo != 0 || cex.IFrame != 1 {
		t.Fatalf("missing direct-input witness: %+v", cex)
	}
}

func TestFraBmcFiltersImplicationsWithNoInternalNodes(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	p := &Fra_Man_t{pManAig: g, pPars: &Fra_Par_t{}, pCla: Fra_ClassesStart(g)}
	p.pCla.vImps = vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(p.pCla.vImps, Fra_ImpCreate(int32(x.NodeID()), int32(state.NodeID())))
	Fra_BmcPerform(p, 0, 1)
	if vec.Vec_IntSize(p.pCla.vImps) != 0 {
		t.Fatal("invalid implication survived BMC")
	}
}
