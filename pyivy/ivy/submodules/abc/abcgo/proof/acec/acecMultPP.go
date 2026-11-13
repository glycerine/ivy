package acec

// Booth partial-product recognition from src/proof/acec/acecMult.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

var acecPPTruths = [...]uint64{0xf335acc0f335acc0, 0x35c035c035c035c0, 0xd728d728d728d728, 0xfd80fd80fd80fd80, 0xacc0acc0acc0acc0, 0x7878787878787878, 0x2828282828282828, 0xd0d0d0d0d0d0d0d0, 0x8080808080808080, 0x8888888888888888, 0xaaaaaaaaaaaaaaaa, 0x5555555555555555, 0xd5a8d5a8d5a8d5a8, 0x2a572a572a572a57, 0xf3c0f3c0f3c0f3c0, 0x5858585858585858, 0xa7a7a7a7a7a7a7a7, 0x2727272727272727, 0xd8d8d8d8d8d8d8d8}

func acecPPMatches(p *Gia_Man_t, id int32, supp *Vec_Int_t, temp *vec.Vec_Wrd_t) bool {
	truth := gia.Gia_ObjComputeTruth6Cis(p, id<<1, supp, temp)
	if vec.Vec_IntSize(supp) > 6 {
		return false
	}
	n := util.Abc_Tt6MinBase(&truth, vec.Vec_IntArray(supp), int32(vec.Vec_IntSize(supp)))
	vec.Vec_IntShrink(supp, int(n))
	if n > 5 {
		return false
	}
	for _, saved := range acecPPTruths {
		if truth == saved || truth == ^saved {
			return true
		}
	}
	return false
}
func Acec_MultMarkPPs(p *Gia_Man_t) *Vec_Bit_t {
	marked := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	temp := vec.Vec_WrdStart(int(gia.Gia_ManObjNum(p)))
	supp := vec.Vec_IntAlloc(100)
	n := 0
	gia.Gia_ManCleanMark0(p)
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		if gia.Gia_ObjIsAnd(gia.Gia_ManObj(p, id)) != 0 && acecPPMatches(p, id, supp, temp) {
			vec.Vec_BitWriteEntry(marked, int(id), 1)
			n++
		}
	}
	gia.Gia_ManCleanMark0(p)
	fmt.Printf("Collected %d pps.\n", n)
	return marked
}
func Acec_MultFindPPs_rec(p *Gia_Man_t, id int32, bold *Vec_Int_t) {
	o := gia.Gia_ManObj(p, id)
	if gia.Gia_ObjMark0(o) != 0 {
		return
	}
	gia.Gia_ObjSetMark0(o, 1)
	if gia.Gia_ObjIsAnd(o) == 0 {
		return
	}
	Acec_MultFindPPs_rec(p, gia.Gia_ObjFaninId0(o, id), bold)
	Acec_MultFindPPs_rec(p, gia.Gia_ObjFaninId1(o, id), bold)
	vec.Vec_IntPush(bold, id)
}
func Acec_MultFindPPs(p *Gia_Man_t) *Vec_Int_t {
	bold := vec.Vec_IntAlloc(100)
	supp := vec.Vec_IntAlloc(100)
	temp := vec.Vec_WrdStart(int(gia.Gia_ManObjNum(p)))
	n := 0
	gia.Gia_ManCleanMark0(p)
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		if gia.Gia_ObjIsAnd(gia.Gia_ManObj(p, id)) != 0 && acecPPMatches(p, id, supp, temp) {
			Acec_MultFindPPs_rec(p, id, bold)
			n++
		}
	}
	gia.Gia_ManCleanMark0(p)
	fmt.Printf("Collected %d pps and %d nodes.\n", n, vec.Vec_IntSize(bold))
	return bold
}
func Acec_BoothFindPPG(p *Gia_Man_t) *Vec_Bit_t {
	ignore := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for _, id := range vec.Vec_IntArray(Acec_MultFindPPs(p)) {
		vec.Vec_BitWriteEntry(ignore, int(id), 1)
	}
	return ignore
}
func Acec_MultFindPPsTest(p *Gia_Man_t) { gia.Gia_ManShow(p, Acec_MultFindPPs(p), 1, 0, 0) }
