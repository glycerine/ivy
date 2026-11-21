package rwr

import (
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
