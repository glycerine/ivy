package int2

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSourceJustificationTreatsHighBitPriorityAsUnsigned(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x, y, state := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, gate)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManSetRegNum(g, 1)
	a, b, node := gia.Gia_ManObj(g, x>>1), gia.Gia_ManObj(g, y>>1), gia.Gia_ManObj(g, gate>>1)
	gia.Gia_ObjSetMark0(a, 0)
	gia.Gia_ObjSetMark0(b, 1)
	gia.Gia_ObjSetValue(a, -2147483648) // C Value stores 0x80000000 as unsigned.
	gia.Gia_ObjSetValue(b, 2)
	gia.Gia_ObjSetValue(node, 2)
	Int2_ManJustify_rec(g, node, vec.Vec_IntAlloc(0))
	if gia.Gia_ObjMark1(a) != 0 {
		t.Fatal("C unsigned priority is above ABC_INFINITY; input was visited")
	}
}
