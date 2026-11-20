package cnf

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	dar "github.com/glycerine/abcgo/opt/dar/core"
	"testing"
)

func TestDarCutSlotsFollowSourceMergeOrder(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	root := g.AND(inner, c)
	g.AddOutput(root)
	slots, err := dar.Dar_ManComputeCutSlots(g, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	set := slots[root]
	want := [][]aig.Edge{{root}, {c, inner}, {a, b, c}}
	for i, leaves := range want {
		if !set[i].Used || len(set[i].Cut.Leaves) != len(leaves) {
			t.Fatalf("slot %d got %+v", i, set[i])
		}
		for j, e := range leaves {
			if set[i].Cut.Leaves[j] != e {
				t.Fatalf("slot %d leaves %v want %v", i, set[i].Cut.Leaves, leaves)
			}
		}
	}
}
func TestDarCutFindsConstantContradiction(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	root := g.AND(inner, a.Not())
	g.AddOutput(root)
	slots, err := dar.Dar_ManComputeCutSlots(g, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, slot := range slots[root] {
		if slot.Used && len(slot.Cut.Leaves) == 0 && slot.Cut.Truth == 0 {
			found = true
		}
	}
	if !found {
		t.Fatalf("constant cut not found: %+v", slots[root])
	}
}
func TestDarAreaMappingChoosesSourceCost(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	root := g.AND(inner, c)
	g.AddOutput(root)
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	if err := cnfApplyDarMapping(m, false); err != nil {
		t.Fatal(err)
	}
	cut := m.selectedCuts[root.NodeID()]
	if cut == nil || len(cut.pFanins) != 3 {
		t.Fatalf("wrong selected cut: %+v", cut)
	}
}

func TestAreaFlowAfterInPlaceAIGReplacement(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(a.Not(), b)
	g.AddOutput(root)
	later := g.AND(a, b)
	replacement := g.AND(later, c)
	aig.Aig_ObjReplace(g, root, replacement, 0)
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	if root.NodeID() >= later.NodeID() || g.Node(root).Kind != aig.And {
		t.Fatal("fixture must have an AND whose fanin was allocated later")
	}
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	Cnf_DeriveMapping(m)
	chosen := m.selectedCuts[root.NodeID()]
	// The three-input cut costs four clauses. Keeping the intermediate
	// gate costs three clauses at the root plus three at its sole fanin gate.
	if len(chosen.pFanins) != 3 || chosen.Cost != 4 {
		t.Fatalf("selected leaves %v at cost %d, want all three inputs at cost 4", chosen.pFanins, chosen.Cost)
	}
}

func TestAreaFlowTieWithHighFanoutInput(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	root := g.AND(inner, c)
	g.AddOutput(root)
	g.AddOutput(inner)
	g.AddOutput(inner)
	for i := 0; i < 2047; i++ {
		g.AddOutput(a)
	}
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	Cnf_DeriveMapping(m)
	chosen := m.selectedCuts[root.NodeID()]
	// Both cuts score 40. DAR's 11-bit Value field gives the two-input
	// cut a tie value of 4 and the three-input cut (2050 mod 2048) 2.
	if len(chosen.pFanins) != 2 || chosen.pFanins[0] != int32(c.NodeID()) || chosen.pFanins[1] != int32(inner.NodeID()) {
		t.Fatalf("selected leaves %v, want the two-input cut", chosen.pFanins)
	}
}

func TestCNFAreaFlowUsesSOPCostAndFanoutDivision(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	root := g.AND(inner, c)
	g.AddOutput(root)
	g.AddOutput(inner)
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	area := make([]int32, aig.Aig_ManObjNumMax(g))
	area[inner.NodeID()] = 83
	cuts := aig.Aig_ComputeCuts(g, 10, 4, 1, 0)
	defer aig.Aig_ManCutStop(cuts)
	found := false
	for _, cut := range cuts.Cuts(root) {
		leaves := aig.Aig_CutLeaves(cut)
		if len(leaves) == 2 && (leaves[0] == int32(inner.NodeID()) || leaves[1] == int32(inner.NodeID())) {
			found = true
			if got := Cnf_CutAssignAreaFlow(m, cut, area); got != 71 {
				t.Errorf("area for two-input AND plus shared internal leaf = %d, want 30+83/2", got)
			}
		}
	}
	if !found {
		t.Fatal("missing two-leaf AIG cut")
	}
	super := vec.Vec_PtrAlloc(2)
	if got := Cnf_CutSuperAreaFlow(super, area, g); got != 100 {
		t.Errorf("empty supergate area %d, want 100", got)
	}
	vec.Vec_PtrPush(super, c)
	vec.Vec_PtrPush(super, inner.Not())
	if got := Cnf_CutSuperAreaFlow(super, area, g); got != 341 {
		t.Errorf("area for two-input supergate with complemented leaf = %d, want 300+83/2", got)
	}
}

func TestPreparedManagerReturnsDarBestCut(t *testing.T) {
	Cnf_ManPrepare()
	defer Cnf_ManFree()
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(a, b.Not())
	g.AddOutput(gate)
	m := Cnf_ManRead()
	m.pManAig = g
	Cnf_DeriveMapping(m)
	cut := Dar_ObjBestCut(&gate)
	if cut == nil {
		t.Fatal("selected DAR cut missing")
	}
	if cut.nLeaves != 2 || cut.uTruth != 0x2222 || cut.pLeaves[0] != int32(a.NodeID()) || cut.pLeaves[1] != int32(b.NodeID()) {
		t.Fatalf("selected DAR cut %+v", cut)
	}
	if Dar_ObjBestCut(&a) != nil || Dar_ObjBestCut(nil) != nil {
		t.Fatal("unexpected terminal or nil cut")
	}
}

func BenchmarkDeriveMappedAndChain(b *testing.B) {
	g := aig.NewGraph()
	last := g.NewInput()
	for range 1024 {
		last = g.AND(last, g.NewInput())
	}
	g.AddOutput(last.Not())
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	want := Cnf_DeriveWithMan(m, g, 0)
	if want.nClauses != 1710 || want.nLiterals != 4442 {
		b.Fatalf("native chain has 1710 clauses and 4442 literals, got %d and %d", want.nClauses, want.nLiterals)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		result := Cnf_DeriveWithMan(m, g, 0)
		if result.nClauses != want.nClauses || result.nVars != want.nVars || result.nLiterals != want.nLiterals {
			b.Fatal("mapping changed")
		}
		Cnf_DataFree(result)
	}
	b.ReportMetric(float64(want.nClauses), "clauses/op")
	b.ReportMetric(float64(want.nLiterals), "literals/op")
}

func TestDarSelectionSurvivesCnfCutReplacement(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	root := g.AND(inner, c)
	g.AddOutput(root)
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	Cnf_DeriveMapping(m)
	replacement := Cnf_CutFromTruth([]int32{int32(inner.NodeID())}, []uint32{0xaaaaaaaa})
	replacement.Cost = 2
	Cnf_ObjSetBestCut(m, &root, replacement)
	got := Cnf_CutCreate(m, &root)
	if got == nil || got.nFanins != 3 || got.truth[0] != 0x80808080 || got.Cost != 4 {
		t.Errorf("created cut %+v, want original three-input AND", got)
	}
	Aig_ManScanMapping(m, 0)
	if m.aArea != 4 || g.Node(inner).Refs != 0 {
		t.Errorf("DAR scan area %d and absorbed gate refs %d, want 4 and 0", m.aArea, g.Node(inner).Refs)
	}
	Cnf_ManTransferCuts(m)
	got = Cnf_ObjBestCut(m, &root)
	if got == nil || got.nFanins != 3 || got.Cost != 4 || Cnf_ObjBestCut(m, &inner) != nil {
		t.Errorf("transferred root %+v, absorbed %+v", got, Cnf_ObjBestCut(m, &inner))
	}
}

func TestDarBestCutOwnedByMappingManager(t *testing.T) {
	type mapped struct {
		m    *Cnf_Man_t
		gate aig.Edge
	}
	makeMapped := func(xor bool) mapped {
		g := aig.NewGraph()
		a, b := g.NewInput(), g.NewInput()
		var gate aig.Edge
		if xor {
			gate = g.NewXor(a, b)
		} else {
			gate = g.AND(a, b)
		}
		g.AddOutput(gate)
		m := Cnf_ManStart()
		m.pManAig = g
		Cnf_DeriveMapping(m)
		return mapped{m, gate}
	}
	l, r := makeMapped(false), makeMapped(true)
	defer Cnf_ManStop(l.m)
	defer Cnf_ManStop(r.m)
	if l.gate != r.gate {
		t.Fatal("fixture requires overlapping IDs")
	}
	left, right := l.m.Dar_ObjBestCut(&l.gate), r.m.Dar_ObjBestCut(&r.gate)
	if left == nil || right == nil || left.uTruth != 0x8888 || right.uTruth != 0x6666 || left.uSign != 30 || right.uSign != 40 || left.Value != 2 || right.Value != 2 {
		t.Fatalf("AND %+v XOR %+v", left, right)
	}
	edge := l.gate.Not()
	if l.m.Dar_ObjBestCut(&edge) != left {
		t.Error("complemented literal lost its regular object's cut")
	}
	cnf := Cnf_ObjBestCut(l.m, &l.gate)
	cnf.pFanins[0] = -1
	cnf.truth[0] = 0
	if left.pLeaves[0] != 1 || left.uTruth != 0x8888 {
		t.Error("DAR cut aliased mutable CNF storage")
	}
	Cnf_ManFreeCuts(l.m)
	if l.m.Dar_ObjBestCut(&l.gate) != nil || r.m.Dar_ObjBestCut(&r.gate) != right {
		t.Error("free affected wrong mapping")
	}
}
func TestDarBestCutWithNoLeaves(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	gate := g.AND(g.AND(a, b), a.Not())
	g.AddOutput(gate)
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	Cnf_DeriveMapping(m)
	cut := m.Dar_ObjBestCut(&gate)
	if cut == nil || cut.nLeaves != 0 || cut.uTruth != 0 {
		t.Fatalf("contradiction cut %+v", cut)
	}
}
