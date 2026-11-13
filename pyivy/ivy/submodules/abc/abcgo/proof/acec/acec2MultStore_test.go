package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestCutStorageAndNodeMerge(t *testing.T) {
	g := gia.Gia_ManStart(16)
	gia.Gia_ManHashStart(g)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	out := gia.Gia_ManHashAnd(g, a, b)
	id := out >> 1
	p := &Sdb_Sto_t{pGia: g, nCutSize: 6, nCutNum: 10, vCuts: vec.Vec_WecStart(int(gia.Gia_ManObjNum(g))), vRefs: vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))}
	vec.Vec_IntWriteEntry(p.vRefs, int(a>>1), 1)
	vec.Vec_IntWriteEntry(p.vRefs, int(b>>1), 1)
	Sdb_CutAddUnit(p, a>>1)
	Sdb_CutAddUnit(p, b>>1)
	Sdb_StoMergeCuts(p, id)
	got := vec.Vec_IntArray(vec.Vec_WecEntry(p.vCuts, int(id)))
	want := []int32{2, 2, a >> 1, b >> 1, -1, 1, id, 2}
	if len(got) != len(want) {
		t.Fatalf("stored cuts %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("stored cuts %v, want %v", got, want)
		}
	}
	if p.nCutsR != 1 || p.CutCount != [4]float64{1, 1, 1, 1} {
		t.Fatalf("counts %d %v", p.nCutsR, p.CutCount)
	}
	if n := Sdb_StoPrepareSet(p, id, 0); n != 2 || p.pCuts[0][0].nTreeLeaves != 2 {
		t.Fatalf("round trip count %d, cut %+v", n, p.pCuts[0][0])
	}
}
