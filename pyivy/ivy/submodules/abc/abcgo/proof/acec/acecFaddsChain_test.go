package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestFullAdderChainDiscovery(t *testing.T) {
	p := gia.Gia_ManStart(128)
	var in [5]int32
	for i := range in {
		in[i] = gia.Gia_ManAppendCi(p)
	}
	var lo, hi [2]int32
	Acec_InsertFadd(p, in[:3], lo[:])
	Acec_InsertFadd(p, []int32{lo[1], in[3], in[4]}, hi[:])
	gia.Gia_ManAppendCo(p, lo[0])
	gia.Gia_ManAppendCo(p, hi[0])
	gia.Gia_ManAppendCo(p, hi[1])
	adds := Gia_ManDetectFullAdders(p, 0, nil)
	m := Gia_ManCreateMap(p, adds)
	Gia_ManFindChains(p, adds, m)
	chains := Gia_ManCollectTopmost(p, adds, m, 2)
	if vec.Vec_WecSize(chains) != 1 || vec.Vec_IntSize(vec.Vec_WecEntry(chains, 0)) != 2 {
		t.Fatalf("chains %d, adders %v", vec.Vec_WecSize(chains), vec.Vec_IntArray(adds))
	}
	mapping := Gia_ManFindMapping(p, adds, m, chains)
	if vec.Vec_IntEntry(mapping, int(lo[0]>>1)) != 0 || vec.Vec_IntEntry(mapping, int(hi[0]>>1)) != 0 || vec.Vec_IntEntry(mapping, int(hi[1]>>1)) != 0 {
		t.Fatal("incorrect output mapping")
	}
	truths := Gia_ManCollectTruthTables(p, adds)
	if vec.Vec_IntSize(truths) != 4 {
		t.Fatalf("truths %v", vec.Vec_IntArray(truths))
	}
}
