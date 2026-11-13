package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4CandidatesDiscardClearedRepresentatives(t *testing.T) {
	g := gia.Gia_ManStart(10)
	for i := 0; i < 4; i++ {
		gia.Gia_ManAppendCi(g)
	}
	g.StartRepresentatives()
	for id := int32(1); id <= 4; id++ {
		gia.Gia_ObjSetRepr(g, id, 0)
	}
	c := newCec4Candidates(g)
	for range 4 {
		if c.next() == 0 {
			t.Fatal("candidate disappeared")
		}
	}
	for id := int32(1); id <= 4; id++ {
		gia.Gia_ObjUnsetRepr(g, id)
	}
	if c.next() != 0 || len(c.ids) != 0 {
		t.Fatal("cleared candidates retained")
	}
}
