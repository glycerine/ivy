package acec

// XOR-cone utilities ported from src/proof/acec/acecUtil.c.
import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_PolynCollectXors_rec(p *gia.Gia_Man_t, id int32, xors *vec.Vec_Int_t) {
	obj := gia.Gia_ManObj(p, id)
	if gia.Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p, obj)
	if gia.Gia_ObjIsAnd(obj) == 0 || gia.Gia_ObjIsXor(obj) == 0 || gia.Gia_ObjRefNum(p, obj) > 1 {
		return
	}
	Gia_PolynCollectXors_rec(p, gia.Gia_ObjFaninId0(obj, id), xors)
	Gia_PolynCollectXors_rec(p, gia.Gia_ObjFaninId1(obj, id), xors)
	vec.Vec_IntPushUnique(xors, id)
}
func Gia_PolynCollectLastXor(p *gia.Gia_Man_t, verbose int32) *vec.Vec_Int_t {
	_ = verbose
	xors := vec.Vec_IntAlloc(100)
	obj := gia.Gia_ManCo(p, gia.Gia_ManCoNum(p)-1)
	p.Refs = nil
	gia.Gia_ManCreateRefs(p)
	gia.Gia_ManIncrementTravId(p)
	Gia_PolynCollectXors_rec(p, gia.Gia_ObjFaninId0p(p, obj), xors)
	vec.Vec_IntReverseOrder(xors)
	p.Refs = nil
	return xors
}
func Gia_PolynAnalyzeXors(p *gia.Gia_Man_t, verbose int32) {
	_ = verbose
	if !gia.Gia_ManHasMuxData(p) {
		fmt.Println("AIG does not have XORs extracted.")
		return
	}
	xors := vec.Vec_IntAlloc(100)
	count := 0
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		vec.Vec_IntClear(xors)
		gia.Gia_ManIncrementTravId(p)
		Gia_PolynCollectXors_rec(p, gia.Gia_ObjFaninId0p(p, gia.Gia_ManCo(p, i)), xors)
		fmt.Printf("%d=%d  ", i, vec.Vec_IntSize(xors))
		count += vec.Vec_IntSize(xors)
	}
	fmt.Printf("Total = %d.\n", count)
}
func Gia_ManDupTopMostRange(p *gia.Gia_Man_t) *gia.Gia_Man_t {
	tops := make([]int32, 0, 7)
	for i := int32(45); i < 52; i++ {
		tops = append(tops, gia.Gia_ObjId(p, gia.Gia_ObjFanin0(gia.Gia_ManCo(p, i))))
	}
	return gia.Gia_ManDupAndConesLimit(p, tops, int32(len(tops)), 100)
}
