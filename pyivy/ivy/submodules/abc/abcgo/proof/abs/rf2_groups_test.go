package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestAlternateRefinementGrouping(t *testing.T) {
	g := gia.Gia_ManStart(80)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	mapping := vec.Vec_IntAlloc(40)
	vec.Vec_IntPush(mapping, a>>1)
	for i := 0; i < 35; i++ {
		n := gia.Gia_ManAppendAnd(g, a, b)
		vec.Vec_IntPush(mapping, n>>1)
	}
	gia.Gia_ManAppendCo(g, a)
	p := Rf2_ManStart(g)
	defer Rf2_ManStop(p, 0)
	p.vMap = mapping
	if groups := Rf2_ManAssignJustIds(p); groups != 18 {
		t.Fatalf("groups=%d", groups)
	}
	if n := vec.Vec_IntSize(vec.Vec_VecEntryInt(p.vGrp2Ppi, 17)); n != 1 {
		t.Fatalf("last group has %d entries", n)
	}
	p.vMap = vec.Vec_IntAlloc(0)
	if groups := Rf2_ManAssignJustIds(p); groups != 0 {
		t.Fatalf("empty groups=%d", groups)
	}
}
