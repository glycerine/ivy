package rwr

import (
	"sync"
	"testing"

	"github.com/glycerine/abcgo/bool/dec"
)

func BenchmarkRewriteManagerStart(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		p := Rwr_ManStart(0)
		if len(p.vForest) != 1785 || len(p.vClasses) != 222 {
			b.Fatal("incomplete rewrite library")
		}
		Rwr_ManStop(p)
	}
}

// The native perf_manager_start fixture hashes the same ordered fields.
func rewriteLibraryDigest(p *Rwr_Man_t) uint64 {
	var h uint64 = 14695981039346656037
	mix := func(v uint32) { h = (h ^ uint64(v)) * 1099511628211 }
	mix(uint32(p.nTravIds))
	for _, n := range p.vForest {
		mix(uint32(n.Id))
		mix(uint32(n.TravId))
		mix(n.Volume)
		mix(n.Level)
		mix(uint32(n.uTruth))
	}
	for class, members := range p.vClasses {
		mix(uint32(p.pMapInv[class]))
		mix(uint32(len(members)))
		for _, n := range members {
			g := n.graph
			leaves := dec.Dec_GraphLeaveNum(g)
			size := leaves + dec.Dec_GraphNodeNum(g)
			mix(uint32(n.Id))
			mix(uint32(dec.Dec_GraphIsConst(g)))
			mix(uint32(leaves))
			mix(uint32(size))
			mix(dec.Dec_EdgeToInt(dec.Dec_GraphRootEdge(g)))
			for i := leaves; i < size; i++ {
				node := dec.Dec_GraphNode(g, i)
				a, b := node.Edges()
				mix(dec.Dec_EdgeToInt(a))
				mix(dec.Dec_EdgeToInt(b))
				f := node.PrintingFlags()
				mix(f[0] | f[1]<<1 | f[2]<<2)
			}
		}
	}
	return h
}

func TestConcurrentStartsMatchNativeRewriteLibrary(t *testing.T) {
	const want = 0xddc262750c62bab1
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			p := Rwr_ManStart(0)
			defer Rwr_ManStop(p)
			if got := rewriteLibraryDigest(p); got != want {
				t.Errorf("rewrite library digest = %016x, native = %016x", got, uint64(want))
			}
		})
	}
	wg.Wait()
}

func TestRewriteManagersShareOnlyCanonicalTables(t *testing.T) {
	a, b := Rwr_ManStart(0), Rwr_ManStart(0)
	defer Rwr_ManStop(b)
	if &a.puCanons[0] != &b.puCanons[0] || &a.pPhases[0] != &b.pPhases[0] ||
		&a.pPerms[0] != &b.pPerms[0] || &a.pMap[0] != &b.pMap[0] {
		t.Error("managers rebuilt the native frame's shared canonical tables")
	}
	for class, candidates := range a.vClasses {
		if len(candidates) == 0 || dec.Dec_GraphLeaveNum(candidates[0].graph) != 4 {
			continue
		}
		left, right := candidates[0], b.vClasses[class][0]
		left.nScore++
		dec.Dec_GraphSetLeafFunc(left.graph, 0, "first manager")
		a.pPractical[0] = false
		if left == right || right.nScore != 0 || dec.Dec_GraphLeafFunc(right.graph, 0) != nil || !b.pPractical[0] {
			t.Fatal("rewriting state leaked between managers")
		}
		Rwr_ManStop(a)
		if b.pMap[right.uTruth] != byte(class) || uint16(dec.Dec_GraphDeriveTruth(right.graph)) != right.uTruth {
			t.Fatal("stopping one manager invalidated the other")
		}
		return
	}
	t.Fatal("no four-input candidate")
}
