package rwr

import (
	"slices"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/opt/cut"
)

type rewriteFanoutObserver struct {
	rewriteTestNetwork
	counted []int32
}

func (n *rewriteFanoutObserver) FanoutCount(v any) int {
	n.counted = append(n.counted, v.(int32))
	return 2
}

func TestDeletedCutLeafSkipsFanoutReads(t *testing.T) {
	p := Rwr_ManStart(0)
	defer Rwr_ManStop(p)
	const truth = 0x8000
	perm := p.pPerms4[p.pPerms[truth]]
	for missing := range 4 {
		leaves := []int32{10, 11, 12, 13}
		leaves[perm[missing]] = -1
		net := new(rewriteFanoutObserver)
		gain := Rwr_RewriteCuts(p, net, nil, []Rwr_Cut{{truth, leaves}}, 12, 0, 0)
		if gain != -1 || len(net.counted) != 0 || len(net.marks) != 0 || net.calls != 0 {
			t.Errorf("missing canonical leaf %d: gain %d, fanouts %v, boundary %v, graph calls %d", missing, gain, net.counted, net.marks, net.calls)
		}
	}
	if p.nCutsBad != 4 || p.nCutsGood != 0 {
		t.Errorf("bad/good cuts %d/%d", p.nCutsBad, p.nCutsGood)
	}
}

func balancedRewriteNode(t testing.TB, pinLeaves bool) (*abc.Object, *cut.Cut_Man_t) {
	t.Helper()
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	aig := n.FunctionManager.(*abc.Abc_Aig_t)
	level := make([]abc.Edge, 16)
	params := cut.DefaultParams(4)
	params.NKeepMax, params.FTruth, params.FFilter = 250, 1, 1
	cuts := cut.Cut_ManStart(params)
	t.Cleanup(func() { cut.Cut_ManStop(cuts) })
	for i := range level {
		pi := n.NewObject(abc.ObjectPI)
		level[i] = pi.Signal()
		cut.Cut_NodeSetTriv(cuts, int32(pi.ID))
	}
	for len(level) > 1 {
		for i, edge := range level {
			if pinLeaves && len(level) == 4 {
				if err := n.AddFanin(n.NewObject(abc.ObjectPO), edge); err != nil {
					t.Fatal(err)
				}
			}
			if i%2 == 1 {
				level[i/2] = abc.Abc_AigAnd(aig, level[i-1], edge)
			}
		}
		level = level[:len(level)/2]
	}
	if err := n.AddFanin(n.NewObject(abc.ObjectPO), level[0]); err != nil {
		t.Fatal(err)
	}
	rewriteNodeCuts(cuts, level[0].Object)
	return level[0].Object, cuts
}

func TestRewriteNodeMatchesOrderedSlice(t *testing.T) {
	for _, pinned := range []bool{false, true} {
		root, cuts := balancedRewriteNode(t, pinned)
		var ordered []Rwr_Cut
		for c := cut.Cut_NodeReadCutsNew(cuts, int32(root.ID)); c != nil; c = c.Next {
			if cut.Cut_CutReadLeaveNum(c) == 4 {
				ordered = append(ordered, Rwr_Cut{uint16(cut.Cut_CutReadTruth(c)[0]), cut.Cut_CutReadLeaves(c)})
			}
		}
		if len(ordered) != 5 {
			t.Fatalf("got %d four-leaf cuts, native tree has five", len(ordered))
		}
		p := Rwr_ManStart(0)
		q := Rwr_ManStart(0)
		net := &abcRewriteNetwork{abc.DecNetwork{Network: root.Network()}}
		want := Rwr_RewriteCuts(q, net, root.Signal().DecValue(), ordered, 1<<30-1, 1, 0)
		got := Rwr_NodeRewrite(p, cuts, root, 0, 1, 0)
		if got != want || p.nCutsGood != int32(len(ordered)) || p.nSubgraphs != q.nSubgraphs || p.fCompl != q.fCompl || !slices.Equal(p.vFanins, q.vFanins) || p.nScores != q.nScores {
			t.Fatalf("pinned=%t gains %d/%d, evaluated %d/%d, cuts %d/%d", pinned, got, want, p.nSubgraphs, q.nSubgraphs, p.nCutsGood, len(ordered))
		}
		if pinned {
			if got != 0 || p.nSubgraphs != 15 || p.fCompl != 0 || len(p.vFanins) != 4 {
				t.Fatalf("pinned native winner: gain %d, candidates %d, phase %d, fanins %v", got, p.nSubgraphs, p.fCompl, p.vFanins)
			}
			for i, value := range p.vFanins {
				edge := value.(abc.Edge)
				if edge.Object.ID != 25+i || !edge.Complement {
					t.Errorf("native winner leaf %d: got %d/%t", i, edge.Object.ID, edge.Complement)
				}
			}
		} else if got != -1 || p.nSubgraphs != 0 {
			t.Fatalf("unpinned native winner: gain %d, candidates %d", got, p.nSubgraphs)
		}
		Rwr_ManStop(p)
		Rwr_ManStop(q)
	}
}

func BenchmarkRewriteOrderedNodeCuts(b *testing.B) {
	root, cuts := balancedRewriteNode(b, false)
	p := Rwr_ManStart(0)
	defer Rwr_ManStop(p)
	b.ReportAllocs()
	for b.Loop() {
		if Rwr_NodeRewrite(p, cuts, root, 0, 1, 0) != -1 {
			b.Fatal("unexpected rewrite")
		}
	}
}

func TestWarmConcreteCutTraversalUsesNoTemporaryStorage(t *testing.T) {
	root, cuts := balancedRewriteNode(t, false)
	p := Rwr_ManStart(0)
	defer Rwr_ManStop(p)
	if allocs := testing.AllocsPerRun(100, func() { Rwr_NodeRewrite(p, cuts, root, 0, 1, 0) }); allocs != 0 {
		t.Fatalf("warm traversal allocated %.0f times", allocs)
	}
	if p.concreteNet.Network != nil {
		t.Fatal("rewrite lookup retains the network after returning")
	}
	for _, leaf := range p.concreteFanins {
		if leaf != nil {
			t.Fatal("rejected cut keeps the input network reachable through the saved boundary")
		}
	}
}

func BenchmarkRewritePinnedOrderedCuts(b *testing.B) {
	root, cuts := balancedRewriteNode(b, true)
	p := Rwr_ManStart(0)
	defer Rwr_ManStop(p)
	b.ReportAllocs()
	for b.Loop() {
		if Rwr_NodeRewrite(p, cuts, root, 0, 1, 0) != 0 {
			b.Fatal("unexpected gain for pinned cut")
		}
	}
}

func TestRewriteLookupMatchesNetworkAcrossPhasesAndCollisions(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	manager := n.FunctionManager.(*abc.Abc_Aig_t)
	edges := []abc.Edge{n.One()}
	for range 32 {
		edges = append(edges, n.NewObject(abc.ObjectPI).Signal())
	}
	for i := 1; i < len(edges); i++ {
		gate := abc.Abc_AigAnd(manager, edges[i-1].Not(), edges[i])
		edges[i].Object.Level = i
		if err := n.AddFanin(n.NewObject(abc.ObjectPO), gate); err != nil {
			t.Fatal(err)
		}
	}
	direct := abc.DecNetwork{Network: n}
	var cached cachedRewriteNetwork
	cached.beginNode(n)
	for i, a := range edges {
		for j, b := range edges {
			for phase := range 4 {
				invertA, invertB := phase&1 != 0, phase&2 != 0
				want, wantLevel := direct.LookupAndLevel(a, b, invertA, invertB)
				got, level := cached.LookupAndLevel(a, b, invertA, invertB)
				swapped, swappedLevel := cached.LookupAndLevel(b.Not(), a.Not(), !invertB, !invertA)
				if got != want || swapped != want || level != wantLevel || swappedLevel != wantLevel {
					t.Fatalf("inputs %d, %d phase %d: got %v/%d, swapped %v/%d, want %v/%d", i, j, phase, got, level, swapped, swappedLevel, want, wantLevel)
				}
			}
		}
	}
}

func TestRewriteLookupExpiresBetweenRootsAndNetworks(t *testing.T) {
	create := func() (*abc.Network, abc.Edge, abc.Edge) {
		n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
		return n, n.NewObject(abc.ObjectPI).Signal(), n.NewObject(abc.ObjectPI).Signal()
	}
	n, a, b := create()
	var cached cachedRewriteNetwork
	cached.beginNode(n)
	if got, _ := cached.LookupAndLevel(a, b, false, false); got != nil {
		t.Fatal("unexpected AND before insertion")
	}
	gate := abc.Abc_AigAnd(n.FunctionManager.(*abc.Abc_Aig_t), a, b)
	cached.generation = ^uint32(0)
	cached.beginNode(n)
	if got, _ := cached.LookupAndLevel(a, b, false, false); got != gate {
		t.Fatal("stale miss after generation wrap")
	}
	other, x, y := create()
	cached.beginNode(other)
	if got, _ := cached.LookupAndLevel(x, y, false, false); got != nil {
		t.Fatal("other network reused the first network's AND")
	}
}
