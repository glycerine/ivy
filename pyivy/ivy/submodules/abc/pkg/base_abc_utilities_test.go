package abc

// Tests for src/base/abc/{abcBarBuf.c,abcLib.c,abcUtil.c,abcShow.c,module.make}.

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func utilityNetwork(t *testing.T) (*ABCNetwork, int, int, int, int, int) {
	t.Helper()
	network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	network.Name = "utility/net"
	a, _ := network.CreatePI()
	b, _ := network.CreatePI()
	if err = network.SetObjectName(a, "a"); err != nil {
		t.Fatal(err)
	}
	if err = network.SetObjectName(b, "b"); err != nil {
		t.Fatal(err)
	}
	and, err := network.CreateLogicNode("11 1\n", ABCObjectEdge{ID: a}, ABCObjectEdge{ID: b, Complemented: true})
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := network.CreateBuffer(ABCObjectEdge{ID: and})
	if err != nil {
		t.Fatal(err)
	}
	po, _ := network.CreatePO()
	if err = network.SetObjectName(po, "y"); err != nil {
		t.Fatal(err)
	}
	if err = network.AddFanin(po, ABCObjectEdge{ID: buffer}); err != nil {
		t.Fatal(err)
	}
	if err = network.Validate(); err != nil {
		t.Fatal(err)
	}
	return network, a, b, and, buffer, po
}

func evalUtilityPO(t *testing.T, network *ABCNetwork, assignment uint64, po int) bool {
	t.Helper()
	values := make([]bool, len(network.objects))
	done := make([]bool, len(network.objects))
	piIndex := make(map[int]int, len(network.pis))
	for i, id := range network.pis {
		piIndex[id] = i
	}
	var evaluate func(int) bool
	evaluate = func(id int) bool {
		if done[id] {
			return values[id]
		}
		object := network.objects[id]
		switch object.Type {
		case ABCObjectConst1:
			values[object.ID] = true
		case ABCObjectPI:
			values[object.ID] = assignment&(uint64(1)<<uint(piIndex[object.ID])) != 0
		case ABCObjectNode:
			sop, err := ParseABCSOP(object.Function, ABCMetadataOptions{})
			if err != nil {
				t.Fatal(err)
			}
			local := uint64(0)
			for i, edge := range object.Fanins {
				value := evaluate(edge.ID) != edge.Complemented
				if value {
					local |= uint64(1) << uint(i)
				}
			}
			values[object.ID], err = sop.Evaluate(local)
			if err != nil {
				t.Fatal(err)
			}
		default:
			if len(object.Fanins) == 1 {
				edge := object.Fanins[0]
				values[object.ID] = evaluate(edge.ID) != edge.Complemented
			}
		}
		done[id] = true
		return values[id]
	}
	return evaluate(po)
}

func TestABCUtilityStatsReportAndCollections(t *testing.T) {
	network, a, b, _, _, _ := utilityNetwork(t)
	stats, err := ABCComputeNetworkStats(network, ABCUtilityOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Objects != 5 || stats.PIs != 2 || stats.POs != 1 || stats.Nodes != 2 || stats.Edges != 4 || stats.Cubes != 2 || stats.Literals != 3 || stats.Buffers != 1 || stats.FaninMaximum != 2 || stats.FanoutMaximum != 1 || stats.TotalFanins != 3 || stats.CrossCutMaximum != 2 {
		t.Fatalf("stats=%+v", stats)
	}
	objects, err := ABCCollectObjects(network)
	if err != nil || len(objects) != 5 {
		t.Fatalf("objects=%v err=%v", objects, err)
	}
	counts, err := ABCFanoutCounts(network)
	if err != nil || counts[a] != 1 || counts[b] != 1 {
		t.Fatalf("fanouts=%v err=%v", counts, err)
	}
	cis, _ := ABCCIIDs(network)
	if !slices.Equal(cis, []int{a, b}) {
		t.Fatalf("cis=%v", cis)
	}
	report, err := ABCNetworkReport(network, ABCUtilityOptions{})
	if err != nil || !strings.Contains(report, "objects=5") || !strings.Contains(report, "cube=2") {
		t.Fatalf("report=%q err=%v", report, err)
	}
	if err = ABCOrderCIsCOs(network); err != nil {
		t.Fatal(err)
	}
}

func TestABCUtilityBarrierRoundTripAndAtomicity(t *testing.T) {
	network, _, _, and, _, po := utilityNetwork(t)
	before := []bool{evalUtilityPO(t, network, 0, po), evalUtilityPO(t, network, 1, po), evalUtilityPO(t, network, 2, po), evalUtilityPO(t, network, 3, po)}
	buffers, err := ABCInsertBarrierBuffers(network, []ABCBarrierRequest{{Fanin: network.objects[po].Fanins[0].ID, Fanout: po}, {Fanin: and, Fanout: network.objects[and].Fanouts[0]}}, ABCUtilityOptions{})
	if err != nil || len(buffers) != 2 {
		t.Fatalf("buffers=%v err=%v", buffers, err)
	}
	for assignment, want := range before {
		if got := evalUtilityPO(t, network, uint64(assignment), po); got != want {
			t.Fatalf("insert assignment=%d got=%v want=%v", assignment, got, want)
		}
	}
	if err = ABCRemoveBarrierBuffers(network, buffers, ABCUtilityOptions{}); err != nil {
		t.Fatal(err)
	}
	for assignment, want := range before {
		if got := evalUtilityPO(t, network, uint64(assignment), po); got != want {
			t.Fatalf("remove assignment=%d got=%v want=%v", assignment, got, want)
		}
	}
	state := network.Clone()
	calls := 0
	_, err = ABCInsertBarrierBuffers(network, []ABCBarrierRequest{{Fanin: and, Fanout: network.objects[and].Fanouts[0]}}, ABCUtilityOptions{Stop: func() bool { calls++; return calls == 1 }})
	if err != ErrCancelled || fmt.Sprint(network.objects) != fmt.Sprint(state.objects) {
		t.Fatalf("atomic err=%v", err)
	}
	calls = 0
	_, err = ABCInsertBarrierBuffers(network, []ABCBarrierRequest{{Fanin: and, Fanout: network.objects[and].Fanouts[0]}}, ABCUtilityOptions{Stop: func() bool { calls++; return calls == 2 }})
	if err != ErrCancelled || fmt.Sprint(network.objects) != fmt.Sprint(state.objects) {
		t.Fatalf("final cancellation atomic err=%v", err)
	}
}

func TestABCUtilityCrossCutSourceOrder(t *testing.T) {
	network, _, _, _, _, _ := utilityNetwork(t)
	got, err := ABCCrossCut(network, ABCUtilityOptions{})
	if err != nil || got != 2 {
		t.Fatalf("shared cone crosscut=%d err=%v", got, err)
	}

	disjoint, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		pi, createErr := disjoint.CreatePI()
		if createErr != nil {
			t.Fatal(createErr)
		}
		buffer, createErr := disjoint.CreateBuffer(ABCObjectEdge{ID: pi})
		if createErr != nil {
			t.Fatal(createErr)
		}
		po, createErr := disjoint.CreatePO()
		if createErr != nil {
			t.Fatal(createErr)
		}
		if createErr = disjoint.AddFanin(po, ABCObjectEdge{ID: buffer}); createErr != nil {
			t.Fatal(createErr)
		}
	}
	got, err = ABCCrossCut(disjoint, ABCUtilityOptions{})
	if err != nil || got != 2 {
		t.Fatalf("disjoint crosscut=%d err=%v", got, err)
	}
	if _, err = ABCCrossCut(disjoint, ABCUtilityOptions{MaxOperations: 1}); err != ErrLimit {
		t.Fatalf("crosscut limit=%v", err)
	}
}

func TestABCUtilityDesignOwnershipAndTopModels(t *testing.T) {
	leaf, _, _, _, _, _ := utilityNetwork(t)
	leaf.Name = "leaf"
	top, _, _, _, _, _ := utilityNetwork(t)
	top.Name = "top"
	design, err := NewABCDesign("demo")
	if err != nil {
		t.Fatal(err)
	}
	if err = design.AddModel(leaf, nil); err != nil {
		t.Fatal(err)
	}
	if err = design.AddModel(top, []string{"leaf"}); err != nil {
		t.Fatal(err)
	}
	tops, err := design.TopModels()
	if err != nil || !slices.Equal(tops, []string{"top"}) {
		t.Fatalf("tops=%v err=%v", tops, err)
	}
	found, ok := design.FindModel("leaf")
	if !ok || found.Name != "leaf" {
		t.Fatal("missing leaf")
	}
	found.Name = "changed"
	foundAgain, _ := design.FindModel("leaf")
	if foundAgain.Name != "leaf" {
		t.Fatal("model snapshot aliases")
	}
	summary, err := design.Summary(ABCUtilityOptions{})
	if err != nil || !strings.Contains(summary, "design demo models 2") {
		t.Fatalf("summary=%q err=%v", summary, err)
	}
	clone, err := design.Clone()
	if err != nil {
		t.Fatal(err)
	}
	design.Close()
	if _, ok = clone.FindModel("top"); !ok {
		t.Fatal("clone lost top")
	}
	one, _ := NewABCDesign("one")
	if err = one.AddModel(leaf, nil); err != nil {
		t.Fatal(err)
	}
	root, err := one.DeriveRoot()
	if err != nil || root.Name != "leaf" {
		t.Fatalf("root=%v err=%v", root, err)
	}
}

func TestABCUtilityDOTFilenameFlopAndInventory(t *testing.T) {
	network, _, _, _, _, _ := utilityNetwork(t)
	dot, err := ABCDOT(network, true, ABCUtilityOptions{})
	if err != nil || !strings.HasPrefix(dot, "digraph G") || !strings.Contains(dot, "style=dashed") {
		t.Fatalf("dot=%q err=%v", dot, err)
	}
	name, err := ABCDOTFilename("a/b c")
	if err != nil || name != "a_b_c.dot" {
		t.Fatalf("name=%q err=%v", name, err)
	}
	seq, _ := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{})
	seq.Name = "seq"
	pi, _ := seq.CreatePI()
	bi, _, bo, err := seq.CreateLatchBox(ABCInitZero)
	if err != nil {
		t.Fatal(err)
	}
	if err = seq.AddFanin(bi, ABCObjectEdge{ID: pi}); err != nil {
		t.Fatal(err)
	}
	po, _ := seq.CreatePO()
	if err = seq.AddFanin(po, ABCObjectEdge{ID: bo}); err != nil {
		t.Fatal(err)
	}
	flopDOT, err := ABCFlopDependencyDOT(seq, ABCUtilityOptions{})
	if err != nil || !strings.Contains(flopDOT, "0 [label=\"0(1)\"]") {
		t.Fatalf("flop=%q err=%v", flopDOT, err)
	}
	if _, err = ABCFlopDependencyDOT(seq, ABCUtilityOptions{MaxTextBytes: 8}); err != ErrLimit {
		t.Fatalf("flop text limit=%v", err)
	}
	want := []string{"abcAig.c", "abcBarBuf.c", "abcBlifMv.c", "abcCheck.c", "abcDfs.c", "abcFanio.c", "abcFanOrder.c", "abcFunc.c", "abcHie.c", "abcHieCec.c", "abcHieGia.c", "abcHieNew.c", "abcLatch.c", "abcLib.c", "abcMinBase.c", "abcNames.c", "abcNetlist.c", "abcNtk.c", "abcObj.c", "abcRefs.c", "abcShow.c", "abcSop.c", "abcUtil.c"}
	if !slices.Equal(ABCBaseModuleSources(), want) {
		t.Fatal("module inventory mismatch")
	}
}

func TestABCUtilityIndependentModel5000(t *testing.T) {
	state := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 {
		state ^= state << 7
		state ^= state >> 9
		state ^= state << 8
		return state
	}
	for trial := 0; trial < 5000; trial++ {
		network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{MaxObjects: 64, MaxEdges: 128})
		if err != nil {
			t.Fatal(err)
		}
		network.Name = fmt.Sprintf("m%d", trial)
		piCount := 1 + int(next()%3)
		ids := make([]int, 0, 10)
		for i := 0; i < piCount; i++ {
			id, createErr := network.CreatePI()
			if createErr != nil {
				t.Fatal(createErr)
			}
			ids = append(ids, id)
		}
		nodeCount := int(next() % 6)
		edges, literals := 1, 0 // final PO edge
		for i := 0; i < nodeCount; i++ {
			first := ids[int(next()%uint64(len(ids)))]
			var id int
			if next()&1 == 0 {
				id, err = network.CreateBuffer(ABCObjectEdge{ID: first, Complemented: next()&1 != 0})
				edges++
				literals++
			} else {
				second := ids[int(next()%uint64(len(ids)))]
				id, err = network.CreateLogicNode("11 1\n", ABCObjectEdge{ID: first}, ABCObjectEdge{ID: second, Complemented: next()&1 != 0})
				edges += 2
				literals += 2
			}
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, id)
		}
		po, _ := network.CreatePO()
		if err = network.AddFanin(po, ABCObjectEdge{ID: ids[len(ids)-1]}); err != nil {
			t.Fatal(err)
		}
		stats, err := ABCComputeNetworkStats(network, ABCUtilityOptions{MaxOperations: 4096})
		if err != nil {
			t.Fatalf("trial=%d: %v", trial, err)
		}
		if stats.Objects != piCount+nodeCount+1 || stats.Nodes != nodeCount || stats.Edges != edges || stats.Cubes != nodeCount || stats.Literals != literals {
			t.Fatalf("trial=%d stats=%+v want objects=%d nodes=%d edges=%d cubes=%d literals=%d", trial, stats, piCount+nodeCount+1, nodeCount, edges, nodeCount, literals)
		}
		first, err := ABCNetworkReport(network, ABCUtilityOptions{})
		if err != nil {
			t.Fatal(err)
		}
		second, err := ABCNetworkReport(network, ABCUtilityOptions{})
		if err != nil || first != second {
			t.Fatalf("trial=%d nondeterministic report", trial)
		}
	}
}

func TestABCUtilityLimitsCancellationAndErrors(t *testing.T) {
	network, _, _, _, buffer, _ := utilityNetwork(t)
	if _, err := ABCComputeNetworkStats(network, ABCUtilityOptions{Stop: func() bool { return true }}); err != ErrCancelled {
		t.Fatalf("cancellation=%v", err)
	}
	if _, err := ABCDOT(network, true, ABCUtilityOptions{MaxTextBytes: 8}); err != ErrLimit {
		t.Fatalf("text limit=%v", err)
	}
	if _, err := ABCNetworkReport(network, ABCUtilityOptions{MaxTextBytes: 8}); err != ErrLimit {
		t.Fatalf("report text limit=%v", err)
	}
	empty, err := NewABCDesign("empty")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = empty.Summary(ABCUtilityOptions{MaxTextBytes: 8}); err != ErrLimit {
		t.Fatalf("empty summary text limit=%v", err)
	}
	before := network.Clone()
	if err := ABCRemoveBarrierBuffers(network, []int{buffer, buffer}, ABCUtilityOptions{}); err != ErrDimension {
		t.Fatalf("duplicate=%v", err)
	}
	if fmt.Sprint(network.objects) != fmt.Sprint(before.objects) {
		t.Fatal("failed removal mutated network")
	}
	design, _ := NewABCDesign("bad")
	model := network.Clone()
	model.Name = "model"
	if err := design.AddModel(model, []string{"missing"}); err != nil {
		t.Fatal(err)
	}
	if _, err := design.TopModels(); err != ErrDimension {
		t.Fatalf("unknown dependency=%v", err)
	}
	design.Close()
	if _, ok := design.FindModel("model"); ok {
		t.Fatal("closed design returned model")
	}
}

func FuzzABCUtilityGraph(f *testing.F) {
	f.Add([]byte{2, 1, 3, 0, 2, 4})
	f.Add([]byte{1, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 96 {
			data = data[:96]
		}
		network, err := NewABCNetwork(ABCNetworkLogic, ABCFunctionSOP, ABCNetworkLimits{MaxObjects: 256, MaxEdges: 512})
		if err != nil {
			t.Fatal(err)
		}
		network.Name = "fuzz"
		piCount := 1
		if len(data) > 0 {
			piCount += int(data[0] % 4)
		}
		ids := make([]int, 0, len(data)+piCount)
		for i := 0; i < piCount; i++ {
			id, createErr := network.CreatePI()
			if createErr != nil {
				t.Fatal(createErr)
			}
			ids = append(ids, id)
		}
		for i, value := range data {
			if i >= 24 || len(ids) == 0 {
				break
			}
			first := ids[int(value)%len(ids)]
			var id int
			if value&1 == 0 || len(ids) == 1 {
				id, err = network.CreateBuffer(ABCObjectEdge{ID: first, Complemented: value&4 != 0})
			} else {
				second := ids[int(value>>2)%len(ids)]
				id, err = network.CreateLogicNode("11 1\n", ABCObjectEdge{ID: first}, ABCObjectEdge{ID: second, Complemented: value&8 != 0})
			}
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, id)
		}
		po, _ := network.CreatePO()
		if err = network.AddFanin(po, ABCObjectEdge{ID: ids[len(ids)-1]}); err != nil {
			t.Fatal(err)
		}
		if _, err = ABCComputeNetworkStats(network, ABCUtilityOptions{MaxOperations: 1 << 16}); err != nil {
			t.Fatal(err)
		}
		if _, err = ABCDOT(network, false, ABCUtilityOptions{MaxOperations: 1 << 16}); err != nil {
			t.Fatal(err)
		}
		if _, err = ABCNetworkReport(network, ABCUtilityOptions{MaxOperations: 1 << 16}); err != nil {
			t.Fatal(err)
		}
		before := make([]bool, uint64(1)<<uint(piCount))
		for assignment := range before {
			before[assignment] = evalUtilityPO(t, network, uint64(assignment), po)
		}
		fanin := network.objects[po].Fanins[0].ID
		buffers, err := ABCInsertBarrierBuffers(network, []ABCBarrierRequest{{Fanin: fanin, Fanout: po}}, ABCUtilityOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if err = ABCRemoveBarrierBuffers(network, buffers, ABCUtilityOptions{}); err != nil {
			t.Fatal(err)
		}
		for assignment, want := range before {
			if got := evalUtilityPO(t, network, uint64(assignment), po); got != want {
				t.Fatalf("assignment=%d got=%v want=%v", assignment, got, want)
			}
		}
	})
}
