package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// ABC cecSweep.c applies the level limit to the representative's reduced-copy
// ID in the original graph. Hashing can make that ID name a different object.
func TestSpecReductionLevelLimitMatchesSourceCopyIndex(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	first := gia.Gia_ManAppendAnd(g, a, b)
	_ = gia.Gia_ManAppendAnd(g, a, b)     // Collapses under structural hashing.
	_ = gia.Gia_ManAppendAnd(g, first, c) // Level two occupies the representative's reduced ID.
	repr := gia.Gia_ManAppendAnd(g, a, c)
	candidate := gia.Gia_ManAppendAnd(g, b, c)
	gia.Gia_ManAppendCo(g, candidate)
	g.StartRepresentatives()
	gia.Gia_ObjSetRepr(g, candidate>>1, repr>>1)
	p := Cec_ManFraStart(g, &Cec_ParFra_t{NLevelMax: 1, NDepthMax: 1})
	queries := Cec_ManFraSpecReduction(p)
	if gia.Gia_ManCoNum(queries) != 0 || vec.Vec_IntSize(p.vXorNodes) != 0 {
		t.Fatalf("source level check should suppress this pair; got %d queries and pairs %v", gia.Gia_ManCoNum(queries), vec.Vec_IntArray(p.vXorNodes))
	}
}
