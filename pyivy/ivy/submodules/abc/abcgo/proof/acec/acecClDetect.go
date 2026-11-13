package acec

// XOR-root and support detection from src/proof/acec/acecCl.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Acec_CollectXorTops(p *Gia_Man_t) *Vec_Int_t {
	roots := vec.Vec_IntAlloc(int(gia.Gia_ManCoNum(p)))
	first := false
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		o := gia.Gia_ObjFanin0(gia.Gia_ManCo(p, i))
		var a, b *Gia_Obj_t
		if gia.Gia_ObjRecognizeExor(o, &a, &b) == 0 {
			if first {
				fmt.Println("XORs do not form a continuous sequence")
				roots = nil
				break
			}
			continue
		}
		first = true
		var a0, a1, b0, b1 *Gia_Obj_t
		xa := gia.Gia_ObjRecognizeExor(gia.Gia_Regular(a), &a0, &a1)
		xb := gia.Gia_ObjRecognizeExor(gia.Gia_Regular(b), &b0, &b1)
		if xa == xb {
			fmt.Println("Both inputs of top level XOR have XOR/non-XOR")
			roots = nil
			break
		}
		vec.Vec_IntPush(roots, gia.Gia_ObjId(p, o))
		if xb != 0 {
			for _, x := range []*Gia_Obj_t{a, b0, b1} {
				vec.Vec_IntPush(roots, gia.Gia_ObjId(p, gia.Gia_Regular(x)))
			}
		} else {
			for _, x := range []*Gia_Obj_t{b, a0, a1} {
				vec.Vec_IntPush(roots, gia.Gia_ObjId(p, gia.Gia_Regular(x)))
			}
		}
	}
	for i := 0; 4*i < vec.Vec_IntSize(roots); i++ {
		fmt.Printf("%2d : %4d <- %4d %4d %4d \n", i, vec.Vec_IntEntry(roots, 4*i), vec.Vec_IntEntry(roots, 4*i+1), vec.Vec_IntEntry(roots, 4*i+2), vec.Vec_IntEntry(roots, 4*i+3))
	}
	return roots
}
func Acec_DetectComputeSuppOne_rec(p *Gia_Man_t, o *Gia_Obj_t, supp, nodes *Vec_Int_t) {
	if gia.Gia_ObjIsTravIdCurrent(p, o) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrent(p, o)
	if gia.Gia_ObjMark0(o) != 0 {
		vec.Vec_IntPush(supp, gia.Gia_ObjId(p, o))
		return
	}
	if gia.Gia_ObjIsAnd(o) == 0 {
		panic("acec: unmarked primary input in support")
	}
	Acec_DetectComputeSuppOne_rec(p, gia.Gia_ObjFanin0(o), supp, nodes)
	Acec_DetectComputeSuppOne_rec(p, gia.Gia_ObjFanin1(o), supp, nodes)
	vec.Vec_IntPush(nodes, gia.Gia_ObjId(p, o))
}
func Acec_DetectComputeSupports(p *Gia_Man_t, roots *Vec_Int_t) {
	nodes, pols, supp := vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100), vec.Vec_IntAlloc(100)
	for i := 0; 4*i < vec.Vec_IntSize(roots); i++ {
		for j := 1; j < 4; j++ {
			gia.Gia_ObjSetMark0(gia.Gia_ManObj(p, vec.Vec_IntEntry(roots, 4*i+j)), 1)
		}
	}
	for i := 1; 4*i < vec.Vec_IntSize(roots); i++ {
		vec.Vec_IntClear(supp)
		gia.Gia_ManIncrementTravId(p)
		id := vec.Vec_IntEntry(roots, 4*i+1)
		o := gia.Gia_ManObj(p, id)
		gia.Gia_ObjSetMark0(o, 0)
		Acec_DetectComputeSuppOne_rec(p, o, supp, nodes)
		gia.Gia_ObjSetMark0(o, 1)
		vec.Vec_IntSort(supp, 0)
		fmt.Printf("Out %4d : %4d  \n", i, id)
		vec.Vec_IntPrint(supp)
		fmt.Println("Cone:")
		for _, n := range vec.Vec_IntArray(nodes) {
			gia.Gia_ObjPrint(p, gia.Gia_ManObj(p, n))
		}
		vec.Vec_IntClear(pols)
		for _, n := range vec.Vec_IntArray(supp) {
			vec.Vec_IntPush(pols, Acec_DetectLitPolarity(p, id, n))
		}
		for k, n := range vec.Vec_IntArray(supp) {
			fmt.Printf("%d(%d)  ", n, vec.Vec_IntEntry(pols, k)&1)
		}
		fmt.Println()
		vec.Vec_IntPrint(supp)
	}
	for i := 0; 4*i < vec.Vec_IntSize(roots); i++ {
		for j := 1; j < 4; j++ {
			gia.Gia_ObjSetMark0(gia.Gia_ManObj(p, vec.Vec_IntEntry(roots, 4*i+j)), 0)
		}
	}
}
