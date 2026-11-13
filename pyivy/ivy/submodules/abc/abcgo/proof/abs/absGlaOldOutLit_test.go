package abs

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestOldGlaOutputAssumptionPolarity(t *testing.T) {
	p := &Gla_Man_t{nObjs: 3, pObjs: make([]Gla_Obj_t, 3), nSatVars: 1, vAddedNew: vec.Vec_IntAlloc(0)}
	for i := range p.pObjs {
		p.pObjs[i].id = int32(i)
	}
	p.pObjs[1].fRo = 1
	p.pObjs[2].fPo = 1
	p.pObjs[2].Fanins[0] = 1
	p.pObjRoot = &p.pObjs[2]
	v := Gla_ManGetVar(p, 1, 0)
	if lit := Gla_ManGetOutLit(p, 0); lit != -1 {
		t.Fatalf("initial zero flop assumption=%d", lit)
	}
	p.pObjRoot.fCompl0 = 1
	if lit := Gla_ManGetOutLit(p, 0); lit != v*2+1 {
		t.Fatalf("inverted output assumption=%d", lit)
	}
	v1 := Gla_ManGetVar(p, 1, 1)
	if lit := Gla_ManGetOutLit(p, 1); lit != v1*2+1 {
		t.Fatalf("later output assumption=%d", lit)
	}
	p.pObjRoot.fCompl0 = 0
	if lit := Gla_ManGetOutLit(p, 1); lit != v1*2 {
		t.Fatalf("positive output assumption=%d", lit)
	}
}
