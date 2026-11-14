package abci

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mio"
)

func darTestOldSequentialNetwork() *abc.Network {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	pi := abc.Abc_NtkCreatePi(n)
	po := abc.Abc_NtkCreatePo(n)
	latch := abc.Abc_NtkCreateLatch(n)
	bi, bo := abc.Abc_NtkCreateBi(n), abc.Abc_NtkCreateBo(n)
	abc.Abc_ObjAddFanin(latch, bi)
	abc.Abc_ObjAddFanin(bo, latch)
	abc.Abc_LatchSetInit0(latch)
	abc.Abc_ObjAddFanin(po, pi)
	abc.Abc_ObjAssignName(pi, "input", "")
	abc.Abc_ObjAssignName(po, "output", "")
	abc.Abc_ObjAssignName(bo, "state", "")
	abc.Abc_ObjAssignName(bi, "next", "")
	abc.Abc_NtkOrderCisCos(n)
	return n
}

func darTestOneRegisterAIG() *aig.Graph {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	and := aig.Aig_And(g, pi, lo.Not())
	g.AddOutput(and)
	g.AddOutput(and.Not())
	aig.Aig_ManSetRegNum(g, 1)
	g.InputNumbers = []int32{0, 1}
	return g
}

func TestAbcNtkFromDarSeqPreservesRegisterTerminalsAndPhases(t *testing.T) {
	old := darTestOldSequentialNetwork()
	g := darTestOneRegisterAIG()
	n := abcNtkFromDarSeq(old, g)
	if n == nil || abc.Abc_NtkLatchNum(n) != 1 || abc.Abc_NtkPiNum(n) != 1 || abc.Abc_NtkPoNum(n) != 1 {
		t.Fatalf("sequential conversion shape mismatch: %#v", n)
	}
	if abc.Abc_NtkCheck(n) == 0 {
		t.Fatal("converted sequential network failed ABC checks")
	}
	if got := abc.Abc_ObjName(abc.Abc_NtkPi(n, 0)); got != "input" {
		t.Fatalf("PI name = %q, want input", got)
	}
	po0, po1 := abc.Abc_NtkCo(n, 0), abc.Abc_NtkCo(n, 1)
	if abc.Abc_ObjFanin0(po0) != abc.Abc_ObjFanin0(po1) || abc.Abc_ObjFaninC0(po0) != 0 || abc.Abc_ObjFaninC0(po1) != 1 {
		t.Fatal("complemented output edges were not preserved")
	}
}

func TestAbcNtkAfterTrimUsesOriginalCIMapAndLatchNames(t *testing.T) {
	old := darTestOldSequentialNetwork()
	g := darTestOneRegisterAIG()
	n := abcNtkAfterTrim(g, old)
	if n == nil || abc.Abc_NtkLatchNum(n) != 1 || abc.Abc_NtkPiNum(n) != 1 || abc.Abc_NtkPoNum(n) != 1 {
		t.Fatalf("trim conversion shape mismatch: %#v", n)
	}
	if got := abc.Abc_ObjName(abc.Abc_NtkPo(n, 0)); got != "output" {
		t.Fatalf("PO name = %q, want output", got)
	}
	boxes := n.ObjectsOfKind(abc.ObjectBO)
	if len(boxes) != 1 || abc.Abc_ObjName(boxes[0]) != "state" {
		t.Fatalf("latch output names = %v, want state", boxes)
	}
	if abc.Abc_NtkCheck(n) == 0 {
		t.Fatal("trimmed network failed ABC checks")
	}
}

func TestAbcNtkFromDarChoicesLinksEquivalentNodes(t *testing.T) {
	old := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	pi0, pi1, pi2 := abc.Abc_NtkCreatePi(old), abc.Abc_NtkCreatePi(old), abc.Abc_NtkCreatePi(old)
	po := abc.Abc_NtkCreatePo(old)
	abc.Abc_ObjAssignName(pi0, "a", "")
	abc.Abc_ObjAssignName(pi1, "b", "")
	abc.Abc_ObjAssignName(pi2, "c", "")
	abc.Abc_ObjAssignName(po, "y", "")
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	main := aig.Aig_And(g, aig.Aig_And(g, a, b), c)
	alt := aig.Aig_And(g, a, aig.Aig_And(g, b, c))
	aig.Aig_ObjSetEquiv(g, main, alt)
	g.AddOutput(main)
	n := abcNtkFromDarChoices(old, g)
	if n == nil {
		t.Fatal("choice conversion returned nil")
	}
	var choices int
	for _, node := range n.ObjectsOfKind(abc.ObjectNode) {
		if node.Choice != nil {
			choices++
		}
	}
	if choices != 1 {
		t.Fatalf("choice links = %d, want one", choices)
	}
	if abc.Abc_NtkCheck(n) == 0 {
		t.Fatalf("choice network failed ABC checks: %v", n.Check())
	}
}

func TestAbcNtkFromCellMappedGiaPreservesMappedFunction(t *testing.T) {
	const genlib = `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 1 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0
GATE and2 2 O=a*b; PIN * NONINV 1 999 2 0 2 0
.end
`
	lib := mio.Mio_LibraryReadBuffer(genlib, 0, nil, 0)
	if lib == nil {
		t.Fatal("failed to parse mapper library")
	}
	oldLib := abcmain.Abc_FrameReadLibGen()
	abcmain.Abc_FrameSetLibGen(lib)
	defer abcmain.Abc_FrameSetLibGen(oldLib)

	g := gia.Gia_ManStart(10)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	out := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, out)
	var pars gia.Jf_Par_t
	gia.Nf_ManSetDefaultPars(&pars)
	if gia.Nf_ManPerformMapping(g, &pars) == nil || gia.Gia_ManHasCellMapping(g) == 0 {
		t.Fatal("failed to create source cell mapping")
	}
	n := abcNtkFromCellMappedGia(g, 1)
	if n == nil || abc.Abc_NtkPiNum(n) != 2 || abc.Abc_NtkPoNum(n) != 1 {
		t.Fatalf("mapped conversion shape mismatch: %#v", n)
	}
	if n.FunctionManager != lib {
		t.Fatal("converted network did not retain the active cell library")
	}
	if abc.Abc_NtkCheck(n) == 0 {
		t.Fatal("mapped network failed ABC checks")
	}
}
