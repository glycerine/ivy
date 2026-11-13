package acec

// Arithmetic core collection and duplication from src/proof/acec/acecCo.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Gia_PolynCoreOrderArray(p *Gia_Man_t, adds, boxes *Vec_Int_t) *Vec_Wec_t {
	marks := Acec_ManPoolGetPointed(p, adds)
	m := Gia_PolynComputeMap(adds, gia.Gia_ManObjNum(p))
	out := vec.Vec_WecStart(vec.Vec_IntSize(boxes))
	roots := vec.Vec_IntAlloc(64)
	for i, box := range vec.Vec_IntArray(boxes) {
		Gia_PolyCollectRoots(adds, m, marks, box, roots)
		order := Gia_PolynCoreOrder_int(p, adds, m, roots, nil)
		vec.Vec_IntAppend(vec.Vec_WecEntry(out, i), order)
	}
	return out
}
func Gia_PolynCoreCollect_rec(p *Gia_Man_t, id int32, nodes *Vec_Int_t, visited *Vec_Bit_t) {
	if vec.Vec_BitEntry(visited, int(id)) != 0 {
		return
	}
	vec.Vec_BitWriteEntry(visited, int(id), 1)
	o := gia.Gia_ManObj(p, id)
	Gia_PolynCoreCollect_rec(p, gia.Gia_ObjFaninId0p(p, o), nodes, visited)
	Gia_PolynCoreCollect_rec(p, gia.Gia_ObjFaninId1p(p, o), nodes, visited)
	vec.Vec_IntPush(nodes, id)
}
func Gia_PolynCoreCollect(p *Gia_Man_t, adds, order *Vec_Int_t) *Vec_Int_t {
	nodes := vec.Vec_IntAlloc(1000)
	visited := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := vec.Vec_IntSize(order) - 1; i >= 0; i-- {
		box := int(vec.Vec_IntEntry(order, i) * 6)
		for k := 0; k < 3; k++ {
			vec.Vec_BitWriteEntry(visited, int(vec.Vec_IntEntry(adds, box+k)), 1)
		}
		Gia_PolynCoreCollect_rec(p, vec.Vec_IntEntry(adds, box+3), nodes, visited)
		Gia_PolynCoreCollect_rec(p, vec.Vec_IntEntry(adds, box+4), nodes, visited)
	}
	return nodes
}
func Gia_PolynCorePrintCones(p *Gia_Man_t, leaves *Vec_Int_t, verbose int32) {
	var maxSupp, maxCone int32
	for i, id := range vec.Vec_IntArray(leaves) {
		supp := gia.Gia_ManSuppSize(p, []int32{id}, 1)
		cone := gia.Gia_ManConeSize(p, []int32{id}, 1)
		if verbose != 0 {
			fmt.Printf("%4d : Supp = %3d  Cone = %3d\n", i, supp, cone)
		}
		if supp > maxSupp {
			maxSupp = supp
		}
		if cone > maxCone {
			maxCone = cone
		}
	}
	if verbose == 0 {
		fmt.Printf("Remaining cones:  Count = %d.  SuppMax = %d.  ConeMax = %d.\n", vec.Vec_IntSize(leaves), maxSupp, maxCone)
	}
}
func Gia_PolynCoreDupTreePlus_rec(dst, p *Gia_Man_t, o *Gia_Obj_t) int32 {
	if o.Value != ^uint32(0) {
		return int32(o.Value)
	}
	if gia.Gia_ObjIsAnd(o) == 0 {
		panic("acec: expected AND in duplicated cone")
	}
	Gia_PolynCoreDupTreePlus_rec(dst, p, gia.Gia_ObjFanin0(o))
	Gia_PolynCoreDupTreePlus_rec(dst, p, gia.Gia_ObjFanin1(o))
	o.Value = uint32(gia.Gia_ManAppendAnd(dst, gia.Gia_ObjFanin0Copy(o), gia.Gia_ObjFanin1Copy(o)))
	return int32(o.Value)
}
func Gia_PolynCoreDupTree(p *Gia_Man_t, extra, leaves, nodes *Vec_Int_t, cones int32) *Gia_Man_t {
	if gia.Gia_ManRegNum(p) != 0 {
		panic("acec: core duplication requires combinational graph")
	}
	gia.Gia_ManFillValue(p)
	dst := gia.Gia_ManStart(gia.Gia_ManObjNum(p))
	dst.CopyIdentityFrom(p)
	gia.Gia_ManConst0(p).Value = 0
	if cones != 0 {
		for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
			gia.Gia_ManCi(p, i).Value = uint32(gia.Gia_ManAppendCi(dst))
		}
		for _, id := range vec.Vec_IntArray(leaves) {
			o := gia.Gia_ManObj(p, id)
			o.Value = uint32(Gia_PolynCoreDupTreePlus_rec(dst, p, o))
		}
	} else {
		for _, id := range vec.Vec_IntArray(leaves) {
			gia.Gia_ManObj(p, id).Value = uint32(gia.Gia_ManAppendCi(dst))
		}
	}
	for _, id := range vec.Vec_IntArray(nodes) {
		o := gia.Gia_ManObj(p, id)
		o.Value = uint32(gia.Gia_ManAppendAnd(dst, gia.Gia_ObjFanin0Copy(o), gia.Gia_ObjFanin1Copy(o)))
	}
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		gia.Gia_ManAppendCo(dst, gia.Gia_ObjFanin0Copy(gia.Gia_ManCo(p, i)))
	}
	if extra != nil {
		for _, id := range vec.Vec_IntArray(extra) {
			gia.Gia_ManAppendCo(dst, int32(gia.Gia_ManObj(p, id).Value))
		}
	}
	return dst
}
func Gia_PolynCoreDetectTest_int(p *Gia_Man_t, extra *Vec_Int_t, cones, verbose int32) *Gia_Man_t {
	start := time.Now()
	adds := Ree_ManComputeCuts(p, nil, 1)
	var leaves, roots *Vec_Int_t
	order := Gia_PolynCoreOrder(p, adds, extra, &leaves, &roots)
	nodes := Gia_PolynCoreCollect(p, adds, order)
	fmt.Printf("Detected %d FAs/HAs. Roots = %d. Leaves = %d. Nodes = %d. Adds = %d. Time = %v\n", vec.Vec_IntSize(adds)/6, vec.Vec_IntSize(leaves), vec.Vec_IntSize(roots), vec.Vec_IntSize(nodes), vec.Vec_IntSize(order), time.Since(start))
	Gia_PolynCorePrintCones(p, leaves, verbose)
	return Gia_PolynCoreDupTree(p, extra, leaves, nodes, cones)
}
func Gia_PolynCoreDetectTest(p *Gia_Man_t, extra, cones, verbose int32) *Gia_Man_t {
	outs := vec.Vec_IntAlloc(0)
	if extra != 0 {
		outs = Gia_PolynAddHaRoots(p)
	}
	dst := Gia_PolynCoreDetectTest_int(p, outs, cones, verbose)
	fmt.Printf("On top of %d COs, created %d new adder outputs.\n", gia.Gia_ManCoNum(p), vec.Vec_IntSize(outs))
	return dst
}
