package fra

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestBmcDriverRefinesAcrossFrameWindows(t *testing.T) {
	makeManager := func(imp bool) *Fra_Man_t {
		g := aig.NewGraph()
		x, y, state := g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(state)
		g.AddOutput(g.AND(x, y))
		aig.Aig_ManSetRegNum(g, 1)
		p := &Fra_Man_t{PManAig: g, PPars: &Fra_Par_t{}, PCla: Fra_ClassesStart(g)}
		Fra_ClassesTest(p.PCla, int32(x.NodeID()), int32(state.NodeID()))
		if imp {
			p.PCla.VImps = vec.Vec_IntAlloc(2)
			vec.Vec_IntPush(p.PCla.VImps, Fra_ImpCreate(int32(x.NodeID()), int32(state.NodeID())))
		}
		return p
	}
	for _, imp := range []bool{false, true} {
		for prefix := int32(0); prefix < 2; prefix++ {
			for depth := int32(1); depth <= 3; depth++ {
				p := makeManager(imp)
				Fra_BmcPerform(p, prefix, depth)
				state := p.PManAig.Inputs()[2]
				if p.PCla.Reprs[state.NodeID()] >= 0 {
					t.Fatalf("imp=%v prefix=%d depth=%d: invalid equivalence retained", imp, prefix, depth)
				}
				if imp && vec.Vec_IntSize(p.PCla.VImps) != 0 {
					t.Fatalf("imp=%v prefix=%d depth=%d: invalid implication retained", imp, prefix, depth)
				}
				if p.PBmc != nil {
					t.Fatal("BMC state not cleared")
				}
			}
		}
	}
}

func TestFraBmcRefinementRespectsFramesAndFraig(t *testing.T) {
	g := aig.NewGraph()
	input, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(input.Not())
	aig.Aig_ManSetRegNum(g, 1)
	b := Fra_BmcStart(g, 0, 2)
	b.PAigFrames = Fra_BmcFrames(b, 0)
	b.PObjToFraig = make([]aig.Edge, len(b.PAigFrames.Nodes()))
	for i := range b.PObjToFraig {
		b.PObjToFraig[i] = aig.Edge(i << 1)
	}
	p := &Fra_Man_t{PManAig: g, PBmc: b}
	if Fra_BmcNodeIsConst(p, state) != 0 {
		t.Fatal("second frame state is not constant")
	}
	if got := Bmc_ObjFrames(b, state, 0); got != b.PAigFrames.Const0() {
		t.Fatalf("initial state: %v", got)
	}
	if got := Bmc_ObjFrames(b, state, 1); got != Bmc_ObjFrames(b, input, 0).Not() {
		t.Fatalf("next state is not complemented input: %v", got)
	}
	mapped := Bmc_ObjFrames(b, state, 1)
	b.PObjToFraig[mapped.NodeID()] = b.PAigFrames.Const1()
	if Fra_BmcNodeIsConst(p, state) != 1 {
		t.Fatal("fraig representative not considered")
	}
	b.PObjToFraig[Bmc_ObjFrames(b, input, 1).NodeID()] = b.PAigFrames.Const1()
	b.NPref = 1
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
	b.PAigFrames = frames
	Bmc_ObjSetFrames(b, x, 0, fx)
	Bmc_ObjSetFrames(b, y, 0, fy)
	b.VImps = vec.Vec_IntAlloc(4)
	imp := Fra_ImpCreate(int32(x.NodeID()), int32(y.NodeID()))
	vec.Vec_IntPush(b.VImps, imp)
	p := &Fra_Man_t{PManAig: frames, PManFraig: reduced, NFramesAll: 1, PMemFraig: make([]aig.Edge, len(frames.Nodes()))}
	Fra_ObjSetFraig(p, fx, 0, rep)
	Fra_ObjSetFraig(p, fy, 0, rep)
	Fra_BmcFilterImplications(p, b)
	if vec.Vec_IntSize(b.VImps) != 1 {
		t.Fatal("equivalent implication removed")
	}
	Fra_ObjSetFraig(p, fy, 0, rep.Not())
	Fra_BmcFilterImplications(p, b)
	if vec.Vec_IntSize(b.VImps) != 0 {
		t.Fatal("complementary implication retained")
	}
}

func TestFraBmcPerformRefinesInitializedClasses(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(x)
	aig.Aig_ManSetRegNum(g, 1)
	p := &Fra_Man_t{PManAig: g, PPars: &Fra_Par_t{}, PCla: Fra_ClassesStart(g)}
	Fra_ClassesTest(p.PCla, int32(x.NodeID()), int32(state.NodeID()))
	Fra_BmcPerform(p, 0, 1)
	if p.PBmc != nil {
		t.Fatal("BMC manager not cleared")
	}
	if p.PCla.Reprs[state.NodeID()] >= 0 {
		t.Fatalf("invalid equivalence survived: %v", p.PCla.Reprs)
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
	p := &Fra_Man_t{PManAig: g, PPars: &Fra_Par_t{}, PCla: Fra_ClassesStart(g)}
	p.PCla.VImps = vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(p.PCla.VImps, Fra_ImpCreate(int32(x.NodeID()), int32(state.NodeID())))
	Fra_BmcPerform(p, 0, 1)
	if vec.Vec_IntSize(p.PCla.VImps) != 0 {
		t.Fatal("invalid implication survived BMC")
	}
}
