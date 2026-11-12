package fret

import (
	"testing"

	abc "github.com/glycerine/abcgo/base/abc"
	core "github.com/glycerine/abcgo/bdd/cudd/core"
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

func TestInitToOrigLookupReturnsMappedObjectAndLag(t *testing.T) {
	original := abc.NewNetwork("original")
	originalObj := abc.Abc_NtkCreatePi(original)
	initNetwork := abc.NewNetwork("init")
	initObj := abc.Abc_NtkCreatePi(initNetwork)
	pManMR = &MinRegMan_t{pNtk: original}
	t.Cleanup(func() { pManMR = nil })
	ensureInitToOrig(abc.Abc_ObjId(initObj))
	pManMR.pInitToOrig[abc.Abc_ObjId(initObj)] = NodeLag_t{id: abc.Abc_ObjId(originalObj), lag: 7}

	var gotObj *Abc_Obj_t
	var gotLag int32
	Abc_FlowRetime_GetInitToOrig(initObj, &gotObj, &gotLag)
	if gotObj != originalObj || gotLag != 7 {
		t.Fatalf("mapping = (%p, %d), want (%p, 7)", gotObj, gotLag, originalObj)
	}
}

func TestClearInitToOrigMarksEntryUnmapped(t *testing.T) {
	initNetwork := abc.NewNetwork("clear-init-map")
	initObj := abc.Abc_NtkCreatePi(initNetwork)
	pManMR = &MinRegMan_t{}
	t.Cleanup(func() { pManMR = nil })
	ensureInitToOrig(abc.Abc_ObjId(initObj))
	pManMR.pInitToOrig[abc.Abc_ObjId(initObj)] = NodeLag_t{id: 3, lag: 8}

	Abc_FlowRetime_ClearInitToOrig(initObj)
	if got := pManMR.pInitToOrig[abc.Abc_ObjId(initObj)]; got != (NodeLag_t{id: -1}) {
		t.Fatalf("cleared entry = %+v, want id=-1 and lag=0", got)
	}
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

func TestExactForwardTimingFindsBoundaryAfterLatch(t *testing.T) {
	network := abc.NewNetwork("exact-forward-latch-boundary")
	pi := abc.Abc_NtkCreatePi(network)
	upstream := abc.Abc_NtkCreateNode(network)
	bi := abc.Abc_NtkCreateBi(network)
	latch := abc.Abc_NtkCreateLatch(network)
	bo := abc.Abc_NtkCreateBo(network)
	downstream1 := abc.Abc_NtkCreateNode(network)
	downstream2 := abc.Abc_NtkCreateNode(network)
	for _, edge := range [][2]*abc.Abc_Obj_t{
		{upstream, pi},
		{bi, upstream},
		{latch, bi},
		{bo, latch},
		{downstream1, bo},
		{downstream2, downstream1},
	} {
		abc.Abc_ObjAddFanin(edge[0], edge[1])
	}
	pManMR = &MinRegMan_t{
		fIsForward: 1,
		maxDelay:   2,
		pNtk:       network,
		pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNumMax(network)+1),
		vNodes:     vec.Vec_PtrAlloc(8),
	}
	t.Cleanup(func() { pManMR = nil })
	Abc_FlowRetime_InitTiming(network)

	Abc_FlowRetime_ConstrainExact(downstream2)
	var got []*abc.Abc_Obj_t
	for i := 0; i < vec.Vec_PtrSize(FTIMEEDGES(downstream2)); i++ {
		obj, ok := vec.Vec_PtrEntry(FTIMEEDGES(downstream2), i).(*abc.Abc_Obj_t)
		if !ok {
			t.Fatalf("timing edge %d has unexpected type", i)
		}
		got = append(got, obj)
	}
	if len(got) != 1 || got[0] != upstream {
		t.Fatalf("exact forward boundary edges = %v, want [%p]", got, upstream)
	}
	if pManMR.nExactConstraints != 1 {
		t.Fatalf("exact constraint count = %d, want 1", pManMR.nExactConstraints)
	}
	for _, obj := range network.Objects() {
		if obj.Level != 0 || obj.MarkA || obj.MarkB {
			t.Fatalf("timing traversal leaked temporary state on object %d: level=%d marks=(%v,%v)", obj.ID, obj.Level, obj.MarkA, obj.MarkB)
		}
	}
}

func TestExactBackwardTimingFindsBoundaryAfterLatch(t *testing.T) {
	network := abc.NewNetwork("exact-backward-latch-boundary")
	pi := abc.Abc_NtkCreatePi(network)
	upstream := abc.Abc_NtkCreateNode(network)
	bi := abc.Abc_NtkCreateBi(network)
	latch := abc.Abc_NtkCreateLatch(network)
	bo := abc.Abc_NtkCreateBo(network)
	downstream := abc.Abc_NtkCreateNode(network)
	po := abc.Abc_NtkCreatePo(network)
	for _, edge := range [][2]*abc.Abc_Obj_t{
		{upstream, pi},
		{bi, upstream},
		{latch, bi},
		{bo, latch},
		{downstream, bo},
		{po, downstream},
	} {
		abc.Abc_ObjAddFanin(edge[0], edge[1])
	}
	pManMR = &MinRegMan_t{
		maxDelay:   1,
		pNtk:       network,
		pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNumMax(network)+1),
		vNodes:     vec.Vec_PtrAlloc(8),
	}
	t.Cleanup(func() { pManMR = nil })
	Abc_FlowRetime_InitTiming(network)

	Abc_FlowRetime_ConstrainExact(upstream)
	if got := vec.Vec_PtrSize(FTIMEEDGES(upstream)); got != 1 {
		t.Fatalf("exact backward edge count = %d, want 1", got)
	}
	if got := vec.Vec_PtrEntry(FTIMEEDGES(upstream), 0); got != downstream {
		t.Fatalf("exact backward edge = %v, want downstream node %d", got, downstream.ID)
	}
}

func TestTimingDFSOrderChecksKeepAllMembershipMarksUntilDone(t *testing.T) {
	network := abc.NewNetwork("timing-dfs-order")
	pi := abc.Abc_NtkCreatePi(network)
	node := abc.Abc_NtkCreateNode(network)
	po := abc.Abc_NtkCreatePo(network)
	abc.Abc_ObjAddFanin(node, pi)
	abc.Abc_ObjAddFanin(po, node)
	nodes := vec.Vec_PtrAlloc(4)

	abc.Abc_NtkIncrementTravId(network)
	Abc_FlowRetime_Dfs_forw(pi, nodes)
	checkForwardOrder(nodes)
	if pi.MarkB || node.MarkB || po.MarkB {
		t.Fatal("forward order check left temporary marks set")
	}
	vec.Vec_PtrClear(nodes)

	abc.Abc_NtkIncrementTravId(network)
	Abc_FlowRetime_Dfs_back(po, nodes)
	checkBackwardOrder(nodes)
	if pi.MarkB || node.MarkB || po.MarkB {
		t.Fatal("backward order check left temporary marks set")
	}
}

func TestFixLatchBoxesPreservesComplementOnInsertedBI(t *testing.T) {
	network := abc.NewNetwork("fix-latch-box-complement")
	pi := abc.Abc_NtkCreatePi(network)
	node := abc.Abc_NtkCreateNode(network)
	abc.Abc_ObjAddFanin(node, pi)
	latch := abc.Abc_NtkCreateLatch(network)
	if err := network.AddFanin(latch, abc.Edge{Object: node, Complement: true}); err != nil {
		t.Fatal(err)
	}
	freeBI := abc.Abc_NtkCreateBi(network)
	freeBO := abc.Abc_NtkCreateBo(network)
	abc.Abc_ObjAddFanin(freeBI, pi)
	abc.Abc_ObjAddFanin(freeBO, freeBI)
	boxInputs := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(boxInputs, freeBI)

	Abc_FlowRetime_FixLatchBoxes(network, boxInputs)
	bi := abc.Abc_ObjFanin0(latch)
	if abc.Abc_ObjIsBi(bi) == 0 {
		t.Fatal("inserted latch input is not a BI")
	}
	fanin := bi.Fanins()[0]
	if fanin.Object != node || !fanin.Complement {
		t.Fatalf("inserted BI fanin = %+v, want complemented edge from node %d", fanin, node.ID)
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

func TestSinkDistanceHistogramMovesOneEntryAndDetectsEmptyBucket(t *testing.T) {
	histogram := vec.Vec_IntStart(4)
	vec.Vec_IntWriteEntry(histogram, 1, 1)
	pManMR = &MinRegMan_t{vSinkDistHist: histogram}
	t.Cleanup(func() { pManMR = nil })

	updateSinkDistanceHistogram(1, 2)
	if got := vec.Vec_IntEntry(histogram, 1); got != 0 {
		t.Fatalf("old bucket count = %d, want 0", got)
	}
	if got := vec.Vec_IntEntry(histogram, 2); got != 1 {
		t.Fatalf("new bucket count = %d, want 1", got)
	}
	if pManMR.fSinkDistTerminate == 0 {
		t.Fatal("empty old bucket did not terminate sink-distance search")
	}
}

func TestSopInitialStateSimulationAndDontCare(t *testing.T) {
	for _, tc := range []struct {
		name      string
		mark      int32
		wantValue int32
		wantCare  int32
	}{
		{name: "known one through inverter", mark: INIT_1, wantValue: 0, wantCare: INIT_0},
		{name: "unknown remains unknown", mark: 0, wantValue: 0, wantCare: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			network := abc.NewNetwork("sop-init-" + tc.name)
			pi := abc.Abc_NtkCreatePi(network)
			node := abc.Abc_NtkCreateNode(network)
			abc.Abc_ObjAddFanin(node, pi)
			abc.Abc_ObjSetData(node, "0 1\n")
			pManMR = &MinRegMan_t{pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNumMax(network)+1)}
			t.Cleanup(func() { pManMR = nil })
			FSET(pi, tc.mark)

			Abc_FlowRetime_SimulateSop(node, "0 1\n")
			if got := FTEST(node, INIT_CARE); got != tc.wantCare {
				t.Fatalf("care marks = %#x, want %#x", got, tc.wantCare)
			}
			if tc.wantCare != 0 && FTEST(node, INIT_1) != tc.wantValue {
				t.Fatalf("value marks = %#x, want %#x", FTEST(node, INIT_1), tc.wantValue)
			}
		})
	}
}

func TestBddInitialStateSimulationUsesCofactors(t *testing.T) {
	for _, tc := range []struct {
		name      string
		mark      int32
		wantValue int32
		wantCare  int32
	}{
		{name: "known one", mark: INIT_1, wantValue: INIT_1, wantCare: INIT_1},
		{name: "unknown", mark: 0, wantValue: 0, wantCare: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			manager := core.NewBDDManager(1)
			variable := core.Cudd_bddIthVar(manager, 0)
			network := abc.NewNetwork("bdd-init-" + tc.name)
			network.Function = uint8(abc.ABC_FUNC_BDD)
			network.FunctionManager = manager
			pi := abc.Abc_NtkCreatePi(network)
			node := abc.Abc_NtkCreateNode(network)
			abc.Abc_ObjAddFanin(node, pi)
			abc.Abc_ObjSetData(node, variable)
			pManMR = &MinRegMan_t{pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNumMax(network)+1)}
			t.Cleanup(func() { pManMR = nil })
			FSET(pi, tc.mark)

			Abc_FlowRetime_SimulateNode(node)
			if got := FTEST(node, INIT_CARE); got != tc.wantCare {
				t.Fatalf("care marks = %#x, want %#x", got, tc.wantCare)
			}
			if got := FTEST(node, INIT_1); got != tc.wantValue {
				t.Fatalf("value marks = %#x, want %#x", got, tc.wantValue)
			}
		})
	}
}

func TestFastFlowPreorderDistancesOnSimpleAIGPath(t *testing.T) {
	network := abc.NewNetwork("flow-preorder-path")
	pi := abc.Abc_NtkCreatePi(network)
	node := abc.Abc_NtkCreateNode(network)
	po := abc.Abc_NtkCreatePo(network)
	abc.Abc_ObjAddFanin(node, pi)
	abc.Abc_ObjAddFanin(po, node)
	pManMR = &MinRegMan_t{
		fIsForward:    1,
		pNtk:          network,
		nNodes:        abc.Abc_NtkObjNum(network),
		pDataArray:    make([]Flow_Data_t, abc.Abc_NtkObjNum(network)),
		vSinkDistHist: vec.Vec_IntStart(20),
	}
	t.Cleanup(func() { pManMR = nil })

	dfsfast_preorder(network)
	want := map[*abc.Abc_Obj_t]Flow_Data_t{
		pi:   {e_dist: 1, r_dist: 2},
		node: {e_dist: 2, r_dist: 3},
		po:   {r_dist: 1},
	}
	for obj, expected := range want {
		got := pManMR.pDataArray[abc.Abc_ObjId(obj)]
		if got.e_dist != expected.e_dist || got.r_dist != expected.r_dist {
			t.Errorf("object %d distances=(%d,%d), want (%d,%d)", abc.Abc_ObjId(obj), got.e_dist, got.r_dist, expected.e_dist, expected.r_dist)
		}
	}
}
