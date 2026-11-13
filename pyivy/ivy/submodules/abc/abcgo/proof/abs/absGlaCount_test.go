package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGlaAbsCount(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	q := gia.Gia_ManAppendCi(g)
	node := gia.Gia_ManAppendAnd(g, a, q)
	gia.Gia_ManAppendCo(g, node)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManSetRegNum(g, 1)
	objects := vec.Vec_IntAlloc(3)
	for _, id := range []int32{0, q >> 1, node >> 1} {
		vec.Vec_IntPush(objects, id)
	}
	p := &Ga2_Man_t{pGia: g, vAbs: objects}
	if Ga2_GlaAbsCount(p, 1, 0) != 1 || Ga2_GlaAbsCount(p, 0, 1) != 1 {
		t.Fatal("wrong abstraction counts")
	}
}
