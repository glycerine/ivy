package cut

import (
	"math/rand"
	"testing"
)

type cutTestNetwork []NetworkNode

func (n cutTestNetwork) CutNodes() []NetworkNode { return n }
func testCut(p *Cut_Man_t, leaves ...int32) *Cut_Cut_t {
	if len(leaves) == 1 {
		return Cut_CutCreateTriv(p, leaves[0])
	}
	c := Cut_CutAlloc(p)
	c.leaves = append([]int32(nil), leaves...)
	for _, leaf := range leaves {
		c.USign |= Cut_NodeSign(leaf)
	}
	return c
}
func TestGlobalRejectedCutCannotDeleteEligibleCut(t *testing.T) {
	params := DefaultParams(4)
	params.FGlobal = 1
	p := Cut_ManStart(params)
	Cut_ManSetNodeAttrs(p, []int32{0, 1, 0, 0})
	old := testCut(p, 1, 2, 3)
	var list Cut_List_t
	Cut_ListStart(&list)
	Cut_ListAdd(&list, old)
	a, b := testCut(p, 2), testCut(p, 3)
	Cut_NodeDoComputeCuts(p, &list, 4, 0, 0, a, b, 0, 0)
	if got := Cut_ListFinish(&list); got != old {
		t.Fatal("ineligible local candidate removed eligible global cut")
	}
}
func TestCombinationalCutTruthAgainstIndependentCones(t *testing.T) {
	rng := rand.New(rand.NewSource(71))
	for trial := 0; trial < 12; trial++ {
		network := cutTestNetwork{{ID: 0, Input: true}, {ID: 1, Input: true}, {ID: 2, Input: true}}
		for id := int32(3); id < 12; id++ {
			a, b := int32(rng.Intn(int(id))), int32(rng.Intn(int(id)))
			network = append(network, NetworkNode{ID: id, Fanins: [2]int32{a, b}, Compl: [2]bool{rng.Intn(2) != 0, rng.Intn(2) != 0}, XOR: trial%3 == 0})
		}
		network = append(network, NetworkNode{ID: 12, Output: true, Fanins: [2]int32{11, 0}})
		params := DefaultParams(4)
		params.FTruth = 1
		p := Abc_NtkCuts(network, params)
		for id := int32(3); id < 12; id++ {
			cuts := p.newCuts[id]
			if Cut_CutListVerify(cuts) == 0 {
				t.Fatal("invalid cut list", id)
			}
			for c := cuts; c != nil; c = c.Next {
				for assignment := 0; assignment < 1<<uint(len(c.leaves)); assignment++ {
					var eval func(int32) bool
					eval = func(node int32) bool {
						for index, leaf := range c.leaves {
							if node == leaf {
								return assignment&(1<<uint(index)) != 0
							}
						}
						n := network[node]
						if n.Input {
							t.Fatalf("cut does not block input: node%d leaves%v", id, c.leaves)
						}
						a, b := eval(n.Fanins[0]) != n.Compl[0], eval(n.Fanins[1]) != n.Compl[1]
						if n.XOR {
							return a != b
						}
						return a && b
					}
					want := eval(id)
					got := c.truth[assignment>>5]&(1<<uint(assignment&31)) != 0
					if got != want {
						t.Fatalf("trial%d node%d leaves%v assignment%d got%v want%v", trial, id, c.leaves, assignment, got, want)
					}
				}
			}
		}
		if p.stats.Current != p.stats.Allocated-p.stats.Recycled {
			t.Fatal(p.stats)
		}
		Cut_ManStop(p)
	}
}
func TestRecordedOracleReplaysTruth(t *testing.T) {
	params := DefaultParams(4)
	params.FTruth = 1
	params.FRecord = 1
	p := Cut_ManStart(params)
	Cut_NodeSetTriv(p, 0)
	Cut_NodeSetTriv(p, 1)
	Cut_NodeSetTriv(p, 2)
	Cut_NodeComputeCuts(p, 3, 0, 1, 0, 1, 1, 0)
	Cut_NodeComputeCutsXor(p, 4, 3, 2, 0, 0, 1, 0)
	oracle := Cut_OracleStart(p)
	for id := int32(0); id < 3; id++ {
		Cut_OracleNodeSetTriv(oracle, id)
	}
	Cut_OracleComputeCuts(oracle, 3, 0, 1, 0, 1)
	Cut_OracleComputeCuts(oracle, 4, 3, 2, 0, 0)
	for id := int32(3); id < 5; id++ {
		a, b := p.newCuts[id], oracle.manager.newCuts[id]
		for a != nil && b != nil {
			if Cut_CutCompare(a, b) != 0 {
				t.Fatal("replay support")
			}
			for w := range a.truth {
				if a.truth[w] != b.truth[w] {
					t.Fatal("replay truth", id, a.truth, b.truth)
				}
			}
			a, b = a.Next, b.Next
		}
		if a != nil || b != nil {
			t.Fatal("replay lengths")
		}
	}
}
func TestTruthComposition(t *testing.T) {
	p := Cut_ManStart(&Cut_Params_t{NVarsMax: 4, NKeepMax: 100, FTruth: 1})
	a, b, c := testCut(p, 1), testCut(p, 2), testCut(p, 3)
	tree := Cut_CutMergeTwo(p, a, b)
	Cut_TruthCompute(p, tree, a, b, 0, 0)
	node := testCut(p, 10)
	factor := Cut_CutMergeTwo(p, node, c)
	Cut_TruthCompute(p, factor, node, c, 0, 1)
	out := testCut(p, 1, 2, 3)
	Cut_TruthCompose(factor, 10, tree, out)
	for assignment := 0; assignment < 32; assignment++ {
		want := assignment&1 != 0 && assignment&2 != 0 && assignment&4 == 0
		if (out.truth[0]>>uint(assignment)&1 != 0) != want {
			t.Fatal(assignment, out.truth)
		}
	}
}
func TestSequentialLeafShiftAndFrontiers(t *testing.T) {
	params := DefaultParams(3)
	params.FSeq = 1
	params.FTruth = 1
	p := Cut_ManStart(params)
	Cut_NodeSetTriv(p, 0)
	Cut_NodeSetTriv(p, 1)
	Cut_NodeComputeCutsSeq(p, 2, 0, 1, 0, 0, 1, 2, 1, -1)
	list := p.newCuts[2]
	if list == nil || list.Next == nil {
		t.Fatal("missing sequential cuts")
	}
	if got := list.Next.leaves; len(got) != 2 || got[0] != 1 || got[1] != 258 {
		t.Fatal(got)
	}
	if list.Next.truth[0] != 0x88888888 {
		t.Fatal(list.Next.truth)
	}
	if p.newCuts[0].leaves[0] != 0 || p.newCuts[1].leaves[0] != 256 {
		t.Fatal("mutated source frontiers")
	}
	Cut_NodeNewMergeWithOld(p, 2)
	if p.newCuts[2] != nil || p.oldCuts[2] == nil {
		t.Fatal("merge")
	}
	Cut_NodeOldTransferToNew(p, 2)
	if p.oldCuts[2] != nil || p.newCuts[2] == nil {
		t.Fatal("transfer")
	}
}

func TestTrivialCutSimulationMatchesCurrentNode(t *testing.T) {
	p := Cut_ManStart(DefaultParams(3))
	Cut_NodeSetTriv(p, 0)
	Cut_NodeSetTriv(p, 1)
	first := Cut_NodeComputeCuts(p, 2, 0, 1, 1, 1, 1, 0)
	if !first.FSimul {
		t.Fatal("trivial cut inherited previous node's simulation instead of current NAND inputs")
	}
	second := Cut_NodeComputeCuts(p, 3, 0, 1, 0, 0, 1, 0)
	if second.FSimul {
		t.Fatal("stale true simulation leaked into following AND")
	}
}
func TestComplementedChoiceCutTruth(t *testing.T) {
	params := DefaultParams(3)
	params.FTruth = 1
	p := Cut_ManStart(params)
	a, b := testCut(p, 0), testCut(p, 1)
	a.FCompl = true
	c := Cut_CutMergeTwo(p, a, b)
	Cut_TruthCompute(p, c, a, b, 0, 0)
	if c.truth[0] != 0x44444444 {
		t.Fatalf("choice phase ignored: got %08x want44444444", c.truth[0])
	}
}
func TestBufferOracleMaintainsTrivialFirst(t *testing.T) {
	params := DefaultParams(3)
	params.FTruth = 1
	params.FRecord = 1
	p := Cut_ManStart(params)
	Cut_NodeSetTriv(p, 0)
	original := Cut_NodeComputeBuffer(p, 1, 0, 1)
	oracle := Cut_OracleStart(p)
	Cut_OracleNodeSetTriv(oracle, 0)
	replay := Cut_OracleComputeCuts(oracle, 1, 0, 0, 1, 0)
	for a, b := original, replay; a != nil && b != nil; a, b = a.Next, b.Next {
		if Cut_CutCompare(a, b) != 0 || a.truth[0] != b.truth[0] {
			t.Fatal("buffer cut order/record mismatch")
		}
	}
}

func TestOraclePreservesChoiceUnionFlow(t *testing.T) {
	params := DefaultParams(4)
	params.FRecord = 1
	params.FTruth = 1
	p := Cut_ManStart(params)
	for id := int32(0); id < 3; id++ {
		Cut_NodeSetTriv(p, id)
	}
	Cut_NodeComputeCuts(p, 3, 0, 1, 0, 0, 1, 0)
	Cut_NodeComputeCuts(p, 4, 1, 2, 0, 0, 1, 0)
	Cut_NodeComputeCuts(p, 5, 3, 2, 0, 0, 1, 0)
	Cut_NodeComputeCuts(p, 6, 0, 4, 0, 0, 1, 0)
	original := Cut_NodeUnionCuts(p, []int32{5, 6})
	oracle := Cut_OracleStart(p)
	for id := int32(0); id < 3; id++ {
		Cut_OracleNodeSetTriv(oracle, id)
	}
	Cut_OracleComputeCuts(oracle, 3, 0, 1, 0, 0)
	Cut_OracleComputeCuts(oracle, 4, 1, 2, 0, 0)
	replay := Cut_OracleComputeCuts(oracle, 5, 3, 2, 0, 0)
	a, b := original, replay
	for a != nil && b != nil {
		if Cut_CutCompare(a, b) != 0 {
			t.Fatalf("choice oracle supports differ: %v / %v", a.leaves, b.leaves)
		}
		a, b = a.Next, b.Next
	}
	if a != nil || b != nil {
		t.Fatal("choice oracle list length mismatch")
	}
}

func TestOracleChoiceKeepsAlternativeInputPhases(t *testing.T) {
	params := DefaultParams(4)
	params.FRecord = 1
	params.FTruth = 1
	p := Cut_ManStart(params)
	for id := int32(0); id < 3; id++ {
		Cut_NodeSetTriv(p, id)
	}
	Cut_NodeComputeCuts(p, 3, 0, 1, 1, 0, 1, 0)
	Cut_NodeComputeCuts(p, 4, 1, 2, 0, 0, 1, 0)
	Cut_NodeComputeCuts(p, 5, 3, 2, 0, 0, 1, 0)
	Cut_NodeComputeCuts(p, 6, 0, 4, 1, 0, 1, 0)
	original := Cut_NodeUnionCuts(p, []int32{5, 6})
	oracle := Cut_OracleStart(p)
	for id := int32(0); id < 3; id++ {
		Cut_OracleNodeSetTriv(oracle, id)
	}
	Cut_OracleComputeCuts(oracle, 3, 0, 1, 1, 0)
	Cut_OracleComputeCuts(oracle, 4, 1, 2, 0, 0)
	replay := Cut_OracleComputeCuts(oracle, 5, 3, 2, 0, 0)
	for a, b := original, replay; a != nil && b != nil; a, b = a.Next, b.Next {
		if Cut_CutCompare(a, b) != 0 || a.truth[0] != b.truth[0] {
			t.Fatalf("alternative choice phase lost: leaves %v truths%08x/%08x", a.leaves, a.truth[0], b.truth[0])
		}
	}
}

func TestNetworkConstantAndBufferOracle(t *testing.T) {
	for _, value := range []bool{false, true} {
		n := cutTestNetwork{{ID: 0, Constant: true, Value: value}, {ID: 1, Input: true}, {ID: 2, Buffer: true, Fanins: [2]int32{0, 0}}, {ID: 3, Fanins: [2]int32{1, 2}}}
		params := DefaultParams(4)
		params.FTruth = 1
		params.FRecord = 1
		p := Abc_NtkCuts(n, params)
		c := p.newCuts[0]
		want := uint32(0)
		if value {
			want = ^uint32(0)
		}
		if c.truth[0] != want || c.FSimul != value {
			t.Fatalf("constant %v truth=%x simulation=%v", value, c.truth[0], c.FSimul)
		}
		oracle := Cut_OracleStart(p)
		Abc_NtkCutsOracle(n, oracle)
		for id, list := range p.newCuts {
			a, b := cutSlice(list), cutSlice(oracle.manager.newCuts[id])
			if len(a) != len(b) {
				t.Fatalf("node %d lengths", id)
			}
			for i := range a {
				if a[i].truth[0] != b[i].truth[0] {
					t.Fatalf("node %d cut %d truth mismatch", id, i)
				}
			}
		}
	}
}

func TestSequentialNetworkXorAndUnary(t *testing.T) {
	n := cutTestNetwork{{ID: 0, Input: true}, {ID: 1, Input: true}, {ID: 2, XOR: true, Fanins: [2]int32{0, 1}}, {ID: 3, Buffer: true, Fanins: [2]int32{2, 0}, Compl: [2]bool{true, false}}}
	params := DefaultParams(4)
	params.FTruth = 1
	p := Abc_NtkSeqCuts(n, params)
	for _, check := range []struct {
		id    int32
		truth uint32
	}{{2, 0x66666666}, {3, 0x99999999}} {
		found := false
		for c := p.newCuts[check.id]; c != nil; c = c.Next {
			if len(c.leaves) == 2 && c.leaves[0] == 0 && c.leaves[1] == 256 {
				found = true
				if c.truth[0] != check.truth {
					t.Fatalf("node %d truth %x", check.id, c.truth[0])
				}
			}
		}
		if !found {
			t.Fatalf("missing sequential cut for %d", check.id)
		}
	}
}
