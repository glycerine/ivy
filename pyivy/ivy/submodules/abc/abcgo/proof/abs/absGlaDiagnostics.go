package abs

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Ga2_ManRefinePrint(p *Ga2_Man_t, core *vec.Vec_Int_t) {
	fmt.Print("\n         Unsat core: \n")
	printObject := func(o *gia.Gia_Obj_t) {
		if gia.Gia_ObjIsRo(p.pGia, o) != 0 {
			fmt.Print("ff ")
		} else {
			fmt.Print("   ")
		}
		if Ga2_ObjIsAbs0(p, o) {
			fmt.Print("a ")
		} else if Ga2_ObjIsLeaf0(p, o) {
			fmt.Print("l ")
		} else {
			fmt.Print("  ")
		}
	}
	for i, id := range vec.Vec_IntArray(core) {
		o := gia.Gia_ManObj(p.pGia, id)
		fmt.Printf("%12d : Obj =%6d ", i, id)
		printObject(o)
		fmt.Print("Fanins: ")
		for _, leaf := range vec.Vec_IntArray(Ga2_ObjLeaves(p.pGia, o)) {
			fmt.Printf("%6d ", leaf)
			printObject(gia.Gia_ManObj(p.pGia, leaf))
		}
		fmt.Println()
	}
}
func Ga2_ManRefinePrintPPis(p *Ga2_Man_t) {
	values := vec.Vec_IntAlloc(100)
	for i, id := range vec.Vec_IntArray(p.vValues) {
		if i == 0 {
			continue
		}
		o := gia.Gia_ManObj(p.pGia, id)
		if Ga2_ObjIsAbs(p, o) {
			continue
		}
		if gia.Gia_ObjPhase(o) == 0 || !Ga2_ObjIsLeaf(p, o) || (gia.Gia_ObjIsAnd(o) == 0 && gia.Gia_ObjIsCi(o) == 0) {
			panic("invalid GLA pseudo-input")
		}
		vec.Vec_IntPush(values, id)
	}
	fmt.Printf("        Current PPIs (%d): ", vec.Vec_IntSize(values))
	vec.Vec_IntSort(values, 1)
	for _, id := range vec.Vec_IntArray(values) {
		fmt.Printf("%d ", id)
	}
	fmt.Println()
}
