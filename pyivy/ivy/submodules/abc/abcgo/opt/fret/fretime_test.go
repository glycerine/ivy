package fret

import (
	"testing"

	abc "github.com/glycerine/abcgo/base/abc"
	vec "github.com/glycerine/abcgo/misc/vec"
)

func TestFlowMarkHelpers(t *testing.T) {
	network := abc.NewNetwork("flow-marks")
	obj := abc.Abc_NtkCreatePi(network)
	pManMR = &MinRegMan_t{pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNum(network))}
	t.Cleanup(func() { pManMR = nil })

	FSET(obj, BLOCK|CONSERVATIVE)
	if got := FTEST(obj, BLOCK_OR_CONS); got != BLOCK_OR_CONS {
		t.Fatalf("combined mark = %#x, want %#x", got, BLOCK_OR_CONS)
	}
	FUNSET(obj, BLOCK)
	if got := FTEST(obj, BLOCK_OR_CONS); got != CONSERVATIVE {
		t.Fatalf("mark after clearing block = %#x, want %#x", got, CONSERVATIVE)
	}
}

func TestFlowPredecessorCannotOverwriteLatchInitMapping(t *testing.T) {
	network := abc.NewNetwork("flow-latch-pred")
	latch := abc.Abc_NtkCreateLatch(network)
	pManMR = &MinRegMan_t{pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNum(network))}
	t.Cleanup(func() { pManMR = nil })

	defer func() {
		if recover() == nil {
			t.Fatal("FSETPRED accepted a latch whose state field is reserved")
		}
	}()
	FSETPRED(latch, nil)
}

func TestFlowPredecessorRoundTrip(t *testing.T) {
	network := abc.NewNetwork("flow-pred-roundtrip")
	obj := abc.Abc_NtkCreatePi(network)
	pred := abc.Abc_NtkCreatePi(network)
	pManMR = &MinRegMan_t{pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNum(network))}
	t.Cleanup(func() { pManMR = nil })

	FSETPRED(obj, pred)
	if got := FGETPRED(obj); got != pred {
		t.Fatalf("flow predecessor = %p, want %p", got, pred)
	}
}

func TestInitToOrigSliceGrowthPreservesMappings(t *testing.T) {
	pManMR = &MinRegMan_t{}
	t.Cleanup(func() { pManMR = nil })

	ensureInitToOrig(3)
	pManMR.pInitToOrig[3] = NodeLag_t{id: 12, lag: -2}
	ensureInitToOrig(90)
	if got := pManMR.pInitToOrig[3]; got != (NodeLag_t{id: 12, lag: -2}) {
		t.Fatalf("mapping changed during growth: %+v", got)
	}
	if got := pManMR.pInitToOrig[90]; got.id != -1 {
		t.Fatalf("new mapping sentinel = %d, want -1", got.id)
	}
}

func TestInitToOrigRejectsNegativeID(t *testing.T) {
	pManMR = &MinRegMan_t{}
	t.Cleanup(func() { pManMR = nil })
	defer func() {
		if recover() == nil {
			t.Fatal("negative object ID was accepted")
		}
	}()
	ensureInitToOrig(-1)
}

func TestTimingEdgesUsePerObjectStorage(t *testing.T) {
	network := abc.NewNetwork("flow-time-edges")
	obj := abc.Abc_NtkCreatePi(network)
	pManMR = &MinRegMan_t{
		pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNum(network)),
		vTimeEdges: make([]vec.Vec_Ptr_t, abc.Abc_NtkObjNumMax(network)+1),
	}
	t.Cleanup(func() { pManMR = nil })

	edges := FTIMEEDGES(obj)
	vec.Vec_PtrPush(edges, obj)
	if got := vec.Vec_PtrSize(FTIMEEDGES(obj)); got != 1 {
		t.Fatalf("timing edge count = %d, want 1", got)
	}
}

func TestInitTimingAllocatesAnEdgeVectorPerObjectID(t *testing.T) {
	network := abc.NewNetwork("flow-timing-init")
	obj := abc.Abc_NtkCreatePi(network)
	pManMR = &MinRegMan_t{
		nExactConstraints:   4,
		nConservConstraints: 3,
		pDataArray:          make([]Flow_Data_t, abc.Abc_NtkObjNum(network)),
	}
	t.Cleanup(func() { pManMR = nil })

	Abc_FlowRetime_InitTiming(network)
	if pManMR.nExactConstraints != 0 || pManMR.nConservConstraints != 0 {
		t.Fatalf("constraint counts = (%d, %d), want (0, 0)", pManMR.nExactConstraints, pManMR.nConservConstraints)
	}
	if got, want := len(pManMR.vTimeEdges), int(abc.Abc_NtkObjNumMax(network)+1); got != want {
		t.Fatalf("timing edge slots = %d, want %d", got, want)
	}
	if got := vec.Vec_PtrSize(pManMR.vExactNodes); got != 0 {
		t.Fatalf("exact-node list starts with %d entries, want empty", got)
	}
}

func TestClearFlowsPreservesNonFlowMarksUntilFullReset(t *testing.T) {
	network := abc.NewNetwork("clear-flow-state")
	obj := abc.Abc_NtkCreatePi(network)
	pManMR = &MinRegMan_t{
		nNodes:     1,
		pDataArray: []Flow_Data_t{{mark: BLOCK | INIT_1 | VISITED_E | FLOW, e_dist: 4, r_dist: 5, pred: obj}},
	}
	t.Cleanup(func() { pManMR = nil })

	Abc_FlowRetime_ClearFlows(0)
	data := pManMR.pDataArray[0]
	if data.mark != BLOCK|INIT_1 || data.e_dist != 0 || data.r_dist != 0 || data.pred != nil {
		t.Fatalf("partial flow reset left unexpected state: %+v", data)
	}
	Abc_FlowRetime_ClearFlows(1)
	if data = pManMR.pDataArray[0]; data != (Flow_Data_t{}) {
		t.Fatalf("full flow reset left unexpected state: %+v", data)
	}
}
