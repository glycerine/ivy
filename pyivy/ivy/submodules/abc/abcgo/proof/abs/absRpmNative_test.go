package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestAbstractionSupportAndDominatorsIncludeMuxControl(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	c := gia.Gia_ManAppendCi(g)
	d := gia.Gia_ManAppendCi(g)
	control := gia.Gia_ManAppendAnd(g, c, d)
	g.EnableMuxes()
	out := gia.Gia_ManAppendMuxReal(g, control, a, b)
	gia.Gia_ManAppendCo(g, out)
	supports := Gia_ManCreateSupps(g, 0)
	got := vec.Vec_IntArray(vec.Vec_WecEntry(supports, int(out>>1)))
	if len(got) != 4 {
		t.Fatalf("MUX support=%v, want all 4 inputs", got)
	}
	for i, id := range got {
		if id != int32(i) {
			t.Fatalf("MUX support=%v", got)
		}
	}
	Gia_ManComputeDoms(g)
	if got := Gia_ObjDom(g, gia.Gia_ManObj(g, control>>1)); got != out>>1 {
		t.Fatalf("control dominator=%d, want %d", got, out>>1)
	}
}
