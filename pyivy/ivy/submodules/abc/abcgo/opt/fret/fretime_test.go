package fret

import (
	"fmt"
	"testing"

	hop "github.com/glycerine/abcgo/aig/hop"
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

	var gotObj *abc.Abc_Obj_t
	var gotLag int32
	Abc_FlowRetime_GetInitToOrig(initObj, &gotObj, &gotLag)
	if gotObj != originalObj || gotLag != 7 {
		t.Fatalf("mapping = (%p, %d), want (%p, 7)", gotObj, gotLag, originalObj)
	}
}

func TestInitToOrigCanMapConstantConflictNode(t *testing.T) {
	original := abc.NewNetwork("original-constant")
	initNetwork := abc.NewNetwork("init-constant")
	pManMR = &MinRegMan_t{pNtk: original, vLags: vec.Vec_IntStart(int(abc.Abc_NtkObjNumMax(original)) + 1)}
	t.Cleanup(func() { pManMR = nil })
	initConstant := initNetwork.ObjectsOfKind(abc.ABC_OBJ_CONST1)[0]
	originalConstant := original.ObjectsOfKind(abc.ABC_OBJ_CONST1)[0]

	Abc_FlowRetime_SetInitToOrig(initConstant, originalConstant)
	var got *abc.Abc_Obj_t
	var lag int32
	Abc_FlowRetime_GetInitToOrig(initConstant, &got, &lag)
	if got != originalConstant || lag != 0 {
		t.Fatalf("constant mapping = (%p, %d), want (%p, 0)", got, lag, originalConstant)
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
	abc.Abc_NtkCreatePi(network)
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

func TestForwardConservativeTimingMarksOverlongLatchCone(t *testing.T) {
	network := abc.NewNetwork("conservative-forward-latch-cone")
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
		maxDelay:   1,
		pNtk:       network,
		pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNumMax(network)+1),
		vNodes:     vec.Vec_PtrAlloc(8),
	}
	t.Cleanup(func() { pManMR = nil })
	Abc_FlowRetime_InitTiming(network)

	Abc_FlowRetime_ConstrainConserv_forw(network)

	if got := FTEST(downstream2, BLOCK); got == 0 {
		t.Fatal("overlong latch cone node was not blocked by hard delay")
	}
	if got := FTEST(downstream2, CONSERVATIVE); got == 0 {
		t.Fatal("overlong latch cone node was not marked conservative")
	}
	if pManMR.nConservConstraints == 0 {
		t.Fatal("conservative constraint count = 0, want a positive count")
	}
	if got := vec.Vec_PtrSize(pManMR.vNodes); got != 0 {
		t.Fatalf("scratch node count after timing pass = %d, want 0", got)
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
	// C records every object at the limiting level, including the zero-delay
	// latch and BI terminals, before the upstream logic increments the level.
	if len(got) != 3 || got[0] != latch || got[1] != bi || got[2] != upstream {
		t.Fatalf("exact forward boundary edges = %v, want latch, BI, upstream", got)
	}
	if pManMR.nExactConstraints != 3 {
		t.Fatalf("exact constraint count = %d, want 1", pManMR.nExactConstraints)
	}
	for _, obj := range network.Objects() {
		if obj.Level != 0 || obj.MarkA || obj.MarkB {
			t.Fatalf("timing traversal leaked temporary state on object %d: level=%d marks=(%v,%v)", obj.ID, obj.Level, obj.MarkA, obj.MarkB)
		}
	}
}

// The source recursion crosses the first latch and stops at the second.
func TestExactTimingTraversalCrossesOnlyOneLatch(t *testing.T) {
	ntk := abc.NewNetwork("two-latch-path")
	pi := abc.Abc_NtkCreatePi(ntk)
	bi1, latch1, bo1 := abc.Abc_NtkCreateBi(ntk), abc.Abc_NtkCreateLatch(ntk), abc.Abc_NtkCreateBo(ntk)
	middle := abc.Abc_NtkCreateNode(ntk)
	bi2, latch2, bo2 := abc.Abc_NtkCreateBi(ntk), abc.Abc_NtkCreateLatch(ntk), abc.Abc_NtkCreateBo(ntk)
	end := abc.Abc_NtkCreateNode(ntk)
	for _, edge := range [][2]*abc.Abc_Obj_t{{bi1, pi}, {latch1, bi1}, {bo1, latch1}, {middle, bo1}, {bi2, middle}, {latch2, bi2}, {bo2, latch2}, {end, bo2}} {
		abc.Abc_ObjAddFanin(edge[0], edge[1])
	}
	for _, tc := range []struct {
		root, firstLatch, beyond, secondLatch *abc.Abc_Obj_t
		forward                               bool
	}{
		{end, latch2, middle, latch1, true},
		{pi, latch1, middle, latch2, false},
	} {
		paths := collectTimingPaths(tc.root, tc.forward)
		seen := make(map[*abc.Abc_Obj_t]bool)
		for _, path := range paths {
			seen[path.obj] = true
			path.obj.MarkA, path.obj.MarkB = false, false
		}
		if !seen[tc.firstLatch] || !seen[tc.beyond] || seen[tc.secondLatch] {
			t.Fatalf("forward=%v: first=%v beyond=%v second=%v", tc.forward, seen[tc.firstLatch], seen[tc.beyond], seen[tc.secondLatch])
		}
	}
}

// A PI appears on both an unregistered path of length three and a path
// crossing one latch with length one. C clears the first-pass level before
// processing the latter, so the PI is not an exact-delay boundary.
func TestExactTimingClearsLevelsBetweenLatchPasses(t *testing.T) {
	ntk := abc.NewNetwork("reconvergent-timing")
	pi := abc.Abc_NtkCreatePi(ntk)
	bi, latch, bo := abc.Abc_NtkCreateBi(ntk), abc.Abc_NtkCreateLatch(ntk), abc.Abc_NtkCreateBo(ntk)
	first, second, root := abc.Abc_NtkCreateNode(ntk), abc.Abc_NtkCreateNode(ntk), abc.Abc_NtkCreateNode(ntk)
	for _, edge := range [][2]*abc.Abc_Obj_t{{bi, pi}, {latch, bi}, {bo, latch}, {first, pi}, {second, first}, {root, second}, {root, bo}} {
		abc.Abc_ObjAddFanin(edge[0], edge[1])
	}
	pManMR = &MinRegMan_t{fIsForward: 1, maxDelay: 3, pNtk: ntk, pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNumMax(ntk)+1), vNodes: vec.Vec_PtrAlloc(8)}
	t.Cleanup(func() { pManMR = nil })
	Abc_FlowRetime_InitTiming(ntk)
	Abc_FlowRetime_ConstrainExact(root)
	if got := vec.Vec_PtrSize(FTIMEEDGES(root)); got != 0 {
		t.Fatalf("found %d spurious timing edges across the latch", got)
	}
}

func TestSilentRestrashRetainsLatchAndComplement(t *testing.T) {
	for _, cleanup := range []int32{0, 1} {
		ntk := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
		input := abc.Abc_NtkCreatePi(ntk)
		bi, latch, bo := abc.Abc_NtkCreateBi(ntk), abc.Abc_NtkCreateLatch(ntk), abc.Abc_NtkCreateBo(ntk)
		output := abc.Abc_NtkCreatePo(ntk)
		for i, obj := range []*abc.Abc_Obj_t{input, bi, latch, bo, output} {
			obj.SetName(fmt.Sprintf("n%d", i))
		}
		abc.Abc_ObjAddFanin(latch, bi)
		abc.Abc_ObjAddFanin(bo, latch)
		abc.Abc_LatchSetInit1(latch)
		node := ntk.AndEdges(input.Signal().Not(), bo.Signal())
		abc.Abc_ObjAddFanin(bi, node)
		abc.Abc_ObjAddFanin(output, bo)
		abc.Abc_NtkOrderCisCos(ntk)
		result := Abc_FlowRetime_NtkSilentRestrash(ntk, cleanup)
		if result == nil || abc.Abc_NtkLatchNum(result) != 1 || abc.Abc_LatchIsInit1(result.ObjectsOfKind(abc.ABC_OBJ_LATCH)[0]) == 0 {
			t.Fatalf("cleanup=%d: latch lost", cleanup)
		}
		nodes := result.ObjectsOfKind(abc.ABC_OBJ_NODE)
		if len(nodes) != 1 || !nodes[0].Fanins()[0].Complement && !nodes[0].Fanins()[1].Complement {
			t.Fatalf("cleanup=%d: complemented input lost", cleanup)
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
	// Match the C pass: the latch and BO also lie at the limiting level.
	if got := vec.Vec_PtrSize(FTIMEEDGES(upstream)); got != 3 {
		t.Fatalf("exact backward edge count = %d, want 3", got)
	}
	for i, want := range []*abc.Abc_Obj_t{latch, bo, downstream} {
		if got := vec.Vec_PtrEntry(FTIMEEDGES(upstream), i); got != want {
			t.Fatalf("exact backward edge %d = %v, want object %d", i, got, want.ID)
		}
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

func TestHopInitialStateSimulationTracksCareAndComplement(t *testing.T) {
	manager := hop.Hop_ManStart()
	defer hop.Hop_ManStop(manager)
	a, b := hop.Hop_IthVar(manager, 0), hop.Hop_IthVar(manager, 1)
	and := hop.Hop_And(manager, a, b)
	for _, tc := range []struct {
		name string
		root *hop.Hop_Obj_t
		in   map[*hop.Hop_Obj_t][2]int32 // value, care
		want [2]int32                    // value, don't-care
	}{
		{"known_one", and, map[*hop.Hop_Obj_t][2]int32{a: {1, 1}, b: {1, 1}}, [2]int32{1, 0}},
		{"known_zero_dominates_unknown", and, map[*hop.Hop_Obj_t][2]int32{a: {0, 1}, b: {0, 0}}, [2]int32{0, 0}},
		{"unknown_and_one", and, map[*hop.Hop_Obj_t][2]int32{a: {0, 0}, b: {1, 1}}, [2]int32{0, 1}},
		{"complemented_known_one", hop.Hop_Not(and), map[*hop.Hop_Obj_t][2]int32{a: {1, 1}, b: {1, 1}}, [2]int32{0, 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, dc := evalHopInit(tc.root, tc.in)
			if got := [2]int32{value, dc}; got != tc.want {
				t.Fatalf("evalHopInit() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestBackwardInitStoresLatchInitNetworkMapping(t *testing.T) {
	network := abc.NewNetwork("backward-init-latch-mapping")
	pi := abc.Abc_NtkCreatePi(network)
	bi, latch, bo := abc.Abc_NtkCreateBi(network), abc.Abc_NtkCreateLatch(network), abc.Abc_NtkCreateBo(network)
	for _, edge := range [][2]*abc.Abc_Obj_t{{bi, pi}, {latch, bi}, {bo, latch}} {
		abc.Abc_ObjAddFanin(edge[0], edge[1])
	}
	initNetwork := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 1)
	initPI := abc.Abc_NtkCreatePi(initNetwork)
	pManMR = &MinRegMan_t{
		fComputeInitState: 1,
		pNtk:              network,
		pInitNtk:          initNetwork,
		pDataArray:        make([]Flow_Data_t, abc.Abc_NtkObjNumMax(network)+1),
	}
	pManMR.pDataArray[abc.Abc_ObjId(bo)].pInitObj = initPI
	t.Cleanup(func() { pManMR = nil })

	Abc_FlowRetime_UpdateBackwardInit(network)

	if got := flowData(latch).pInitObj; got != initPI {
		t.Fatalf("latch init-network mapping = %p, want %p", got, initPI)
	}
	if got := abc.Abc_ObjData(latch); got != initPI {
		t.Fatalf("latch object data mapping = %v, want %p", got, initPI)
	}
}

func TestConnectBiasNodeFollowsBoxAndLatchBoundaries(t *testing.T) {
	network := abc.NewNetwork("bias-box-boundaries")
	root := abc.Abc_NtkCreateNode(network)
	bi, latch, bo := abc.Abc_NtkCreateBi(network), abc.Abc_NtkCreateLatch(network), abc.Abc_NtkCreateBo(network)
	afterLatch := abc.Abc_NtkCreateNode(network)
	for _, edge := range [][2]*abc.Abc_Obj_t{{bi, root}, {latch, bi}, {bo, latch}, {afterLatch, bo}} {
		abc.Abc_ObjAddFanin(edge[0], edge[1])
	}
	pManMR = &MinRegMan_t{
		pNtk:       network,
		pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNumMax(network)+1),
		vLags:      vec.Vec_IntStart(int(abc.Abc_NtkObjNumMax(network))),
	}
	t.Cleanup(func() { pManMR = nil })

	Abc_FlowRetime_ConnectBiasNode(nil, root, 1)
	if FTEST(afterLatch, BLOCK) == 0 {
		t.Fatal("bias did not reach the node after the latch boundary")
	}
	if FTEST(root, BLOCK) != 0 {
		t.Fatal("bias for lag one unexpectedly marked the lag-zero root")
	}
	if abc.Abc_NodeIsTravIdCurrent(afterLatch) == 0 {
		t.Fatal("bias traversal did not mark the reachable node with the network traversal ID")
	}
}
