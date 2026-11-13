package acec

// Arithmetic box lifecycle and overlap filtering from acecTree.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Acec_BoxFree(box *Acec_Box_t) {
	if box == nil {
		return
	}
	for _, v := range []**Vec_Wec_t{&box.vAdds, &box.vLeafLits, &box.vRootLits, &box.vUnique, &box.vShared} {
		vec.Vec_WecFreeP(v)
	}
}
func Acec_BoxFreeP(box **Acec_Box_t) {
	if box != nil {
		Acec_BoxFree(*box)
		*box = nil
	}
}
func Acec_VerifyBoxLeaves(box *Acec_Box_t, ignore *Vec_Bit_t) {
	if ignore == nil {
		return
	}
	count := 0
	for i := 0; i < vec.Vec_WecSize(box.vLeafLits); i++ {
		for _, lit := range vec.Vec_IntArray(vec.Vec_WecEntry(box.vLeafLits, i)) {
			id := lit >> 1
			if gia.Gia_ObjIsAnd(gia.Gia_ManObj(box.pGia, id)) != 0 && vec.Vec_BitEntry(ignore, int(id)) == 0 {
				fmt.Printf("Internal node %d of rank %d is not part of PPG.\n", id, i)
				count++
			}
		}
	}
	fmt.Printf("Detected %d suspicious leaves.\n", count)
}
func acecFilterMarked(adds, tree *Vec_Int_t, marked *Vec_Bit_t) {
	k := 0
	for i := 0; i < vec.Vec_IntSize(tree); i += 2 {
		box, rank := vec.Vec_IntEntry(tree, i), vec.Vec_IntEntry(tree, i+1)
		a, b := vec.Vec_IntEntry(adds, int(6*box)+3), vec.Vec_IntEntry(adds, int(6*box)+4)
		if vec.Vec_BitEntry(marked, int(a)) != 0 || vec.Vec_BitEntry(marked, int(b)) != 0 {
			fmt.Printf("Removing box %d=(%d,%d) of rank %d.\n", box, a, b, rank)
			continue
		}
		vec.Vec_IntWriteEntry(tree, k, box)
		vec.Vec_IntWriteEntry(tree, k+1, rank)
		k += 2
	}
	vec.Vec_IntShrink(tree, k)
}
func Acec_TreeFilterOne(p *Gia_Man_t, adds, tree *Vec_Int_t) {
	roots, marked := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p))), vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; i < vec.Vec_IntSize(tree); i += 2 {
		box := vec.Vec_IntEntry(tree, i)
		for k := 3; k < 5; k++ {
			vec.Vec_BitWriteEntry(roots, int(vec.Vec_IntEntry(adds, int(6*box)+k)), 1)
		}
	}
	for i := 0; i < vec.Vec_IntSize(tree); i += 2 {
		box := vec.Vec_IntEntry(tree, i)
		for k := 0; k < 3; k++ {
			vec.Vec_BitWriteEntry(roots, int(vec.Vec_IntEntry(adds, int(6*box)+k)), 0)
		}
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 {
			continue
		}
		a, b := gia.Gia_ObjFaninId0(o, id), gia.Gia_ObjFaninId1(o, id)
		if vec.Vec_BitEntry(roots, int(a)) != 0 || vec.Vec_BitEntry(roots, int(b)) != 0 || vec.Vec_BitEntry(marked, int(a)) != 0 || vec.Vec_BitEntry(marked, int(b)) != 0 {
			vec.Vec_BitWriteEntry(marked, int(id), 1)
		}
	}
	acecFilterMarked(adds, tree, marked)
}
func Acec_TreeFilterTrees(p *Gia_Man_t, adds *Vec_Int_t, trees *Vec_Wec_t) {
	for i := 0; i < vec.Vec_WecSize(trees); i++ {
		Acec_TreeFilterOne(p, adds, vec.Vec_WecEntry(trees, i))
	}
}
func Acec_TreeMarkTFI_rec(p *Gia_Man_t, id int32, marked *Vec_Bit_t) {
	if vec.Vec_BitEntry(marked, int(id)) != 0 {
		return
	}
	vec.Vec_BitWriteEntry(marked, int(id), 1)
	o := gia.Gia_ManObj(p, id)
	if gia.Gia_ObjIsAnd(o) == 0 {
		return
	}
	Acec_TreeMarkTFI_rec(p, gia.Gia_ObjFaninId0(o, id), marked)
	Acec_TreeMarkTFI_rec(p, gia.Gia_ObjFaninId1(o, id), marked)
}
func Acec_TreeFilterOne2(p *Gia_Man_t, adds, tree *Vec_Int_t) {
	leaves, marked := vec.Vec_BitStart(int(gia.Gia_ManObjNum(p))), vec.Vec_BitStart(int(gia.Gia_ManObjNum(p)))
	for i := 0; i < vec.Vec_IntSize(tree); i += 2 {
		box := vec.Vec_IntEntry(tree, i)
		for k := 0; k < 3; k++ {
			vec.Vec_BitWriteEntry(leaves, int(vec.Vec_IntEntry(adds, int(6*box)+k)), 1)
		}
	}
	for i := 0; i < vec.Vec_IntSize(tree); i += 2 {
		box := vec.Vec_IntEntry(tree, i)
		for k := 3; k < 5; k++ {
			vec.Vec_BitWriteEntry(leaves, int(vec.Vec_IntEntry(adds, int(6*box)+k)), 0)
		}
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		if gia.Gia_ObjIsAnd(gia.Gia_ManObj(p, id)) != 0 && vec.Vec_BitEntry(leaves, int(id)) != 0 {
			Acec_TreeMarkTFI_rec(p, id, marked)
		}
	}
	acecFilterMarked(adds, tree, marked)
}
func Acec_TreeFilterTrees2(p *Gia_Man_t, adds *Vec_Int_t, trees *Vec_Wec_t) {
	for i := 0; i < vec.Vec_WecSize(trees); i++ {
		Acec_TreeFilterOne2(p, adds, vec.Vec_WecEntry(trees, i))
	}
}
