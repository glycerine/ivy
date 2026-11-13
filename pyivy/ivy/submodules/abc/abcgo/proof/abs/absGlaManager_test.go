package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGlaObjectFrameLiterals(t *testing.T) {
	g := gia.Gia_ManStart(5)
	x := gia.Gia_ManAppendCi(g)
	o := gia.Gia_ManObj(g, x>>1)
	ids := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(g)))
	values := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(values, x>>1)
	frame := vec.Vec_IntStartFull(1)
	p := &Ga2_Man_t{pGia: g, vIds: ids, vValues: values, vId2Lit: []*vec.Vec_Int_t{frame}, nSatVars: 4, vCnfs: []*vec.Vec_Int_t{nil, nil}}
	Ga2_ObjSetId(p, o, 0)
	if !Ga2_ObjIsLeaf(p, o) || Ga2_ObjIsAbs(p, o) || Ga2_ObjFindLit(p, o, 0) != -1 {
		t.Fatal("incorrect initial object state")
	}
	if got := Ga2_ObjFindOrAddLit(p, o, 0); got != 8 || p.nSatVars != 5 {
		t.Fatalf("literal %d vars %d", got, p.nSatVars)
	}
	if got := Ga2_ObjFindOrAddLit(p, o, 0); got != 8 || p.nSatVars != 5 {
		t.Fatal("existing literal not reused")
	}
	p.vCnfs[0] = vec.Vec_IntAlloc(0)
	if !Ga2_ObjIsAbs(p, o) || Ga2_ObjIsLeaf(p, o) {
		t.Fatal("non-nil empty CNF should mark abstracted object")
	}
}
