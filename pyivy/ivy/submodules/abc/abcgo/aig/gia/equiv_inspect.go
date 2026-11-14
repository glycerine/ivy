package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManCheckTopoOrder_rec(p *Gia_Man_t, obj *Gia_Obj_t) int32 {
	if obj.Value == 0 {
		return 1
	}
	obj.Value = 0
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: unexpected object in topology check")
	}
	if Gia_ManCheckTopoOrder_rec(p, Gia_ObjFanin0(obj)) == 0 || Gia_ManCheckTopoOrder_rec(p, Gia_ObjFanin1(obj)) == 0 {
		return 0
	}
	if p.PReprs != nil {
		repr := Gia_ObjReprObj(p, obj.id)
		if repr != nil && repr.Value != 0 {
			return 0
		}
	}
	return 1
}
func Gia_ManCheckTopoOrder(p *Gia_Man_t) int32 {
	Gia_ManFillValue(p)
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ManCi(p, i).Value = 0
	}
	result := int32(1)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		result &= Gia_ManCheckTopoOrder_rec(p, Gia_ObjFanin0(Gia_ManCo(p, i)))
	}
	return result
}
func Gia_ManEquivCountLitsAll(p *Gia_Man_t) int32 {
	var n int32
	for i := int32(0); i < p.nObjs; i++ {
		if Gia_ObjRepr(p, i) != GIA_VOID {
			n++
		}
	}
	return n
}
func Gia_ManEquivCountClasses(p *Gia_Man_t) int32 {
	if p.PReprs == nil {
		return 0
	}
	var n int32
	for i := int32(1); i < p.nObjs; i++ {
		n += Gia_ObjIsHead(p, i)
	}
	return n
}
func Gia_ManEquivCheckLits(p *Gia_Man_t, n int32) int32 {
	if actual := Gia_ManEquivCountLitsAll(p); actual != n {
		fmt.Printf("Detected a mismatch in counting equivalence classes (%d).\n", actual-n)
	}
	return 1
}
func giaEquivStats(p *Gia_Man_t) (heads, constants, unused, proved, lits int32) {
	for i := int32(1); i < p.nObjs; i++ {
		if Gia_ObjIsHead(p, i) != 0 {
			heads++
		} else if Gia_ObjIsConst(p, i) != 0 {
			constants++
		} else if Gia_ObjIsNone(p, i) != 0 {
			unused++
		}
		if Gia_ObjProved(p, i) != 0 {
			proved++
		}
	}
	unused -= Gia_ManCoNum(p)
	lits = Gia_ManCiNum(p) + Gia_ManAndNum(p) - heads - unused
	return
}
func Gia_ManPrintStatsClasses(p *Gia_Man_t) {
	heads, constants, _, _, lits := giaEquivStats(p)
	fmt.Printf("cst =%3d  cls =%6d  lit =%8d\n", constants, heads, lits)
}
func Gia_ManEquivCountLits(p *Gia_Man_t) int32 {
	if p.PReprs == nil || p.PNexts == nil {
		return 0
	}
	_, _, _, _, lits := giaEquivStats(p)
	return lits
}
func Gia_ManEquivCountOne(p *Gia_Man_t, id int32) int32 {
	n := int32(1)
	for next := p.PNexts[id]; next != 0; next = p.PNexts[next] {
		if Gia_ObjRepr(p, next) != id {
			panic("gia: inconsistent equivalence class")
		}
		n++
	}
	return n
}
func Gia_ManEquivPrintOne(p *Gia_Man_t, id, counter int32) {
	fmt.Printf("Class %4d :  Num = %2d  {", counter, Gia_ManEquivCountOne(p, id))
	for member := id; member != 0; member = p.PNexts[member] {
		fmt.Printf(" %d", member)
		repr := p.PReprs[member]
		if repr.fColorA != 0 || repr.fColorB != 0 {
			fmt.Printf(" <%d%d>", repr.fColorA, repr.fColorB)
		}
	}
	fmt.Println(" }")
}
func Gia_ManEquivPrintClasses(p *Gia_Man_t, verbose int32, mem float32) {
	heads, constants, unused, proved, lits := giaEquivStats(p)
	fmt.Printf("cst =%8d  cls =%7d  lit =%8d  unused =%8d  proof =%6d\n", constants, heads, lits, unused, proved)
	Gia_ManEquivCheckLits(p, lits)
	if verbose != 0 {
		fmt.Printf("Const0 (%d) = ", constants)
		for i := int32(1); i < p.nObjs; i++ {
			if Gia_ObjIsConst(p, i) != 0 {
				fmt.Printf("%d ", i)
			}
		}
		fmt.Println()
		var n int32
		for i := int32(1); i < p.nObjs; i++ {
			if Gia_ObjIsHead(p, i) != 0 {
				n++
				Gia_ManEquivPrintOne(p, i, n)
			}
		}
	}
}
func Gia_ManChoiceMinLevel_rec(p *Gia_Man_t, id, dive int32, mapping *Vec_Int_t) int32 {
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsCi(obj) != 0 || id == 0 {
		return 0
	}
	if level := Gia_ObjLevel(p, obj); level != 0 {
		return level
	}
	if dive != 0 && Gia_ObjIsClass(p, id) != 0 {
		head := Gia_ObjRepr(p, id)
		if head == GIA_VOID {
			head = id
		}
		best, level := int32(-1), int32(embInfinity)
		for member := head; member != 0; member = p.PNexts[member] {
			current := Gia_ManChoiceMinLevel_rec(p, member, 0, mapping)
			if current < level {
				best, level = member, current
			}
		}
		if level <= 0 {
			panic("gia: invalid choice level")
		}
		vec.Vec_IntWriteEntry(mapping, int(head), best)
		for member := head; member != 0; member = p.PNexts[member] {
			Gia_ObjSetLevelId(p, member, level)
		}
		return level
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: choice root is not an AND")
	}
	level := 1 + max(Gia_ManChoiceMinLevel_rec(p, Gia_ObjFaninId0(obj, id), 1, mapping), Gia_ManChoiceMinLevel_rec(p, Gia_ObjFaninId1(obj, id), 1, mapping))
	Gia_ObjSetLevel(p, obj, level)
	return level
}
func Gia_ManChoiceMinLevel(p *Gia_Man_t) *Vec_Int_t {
	mapping := vec.Vec_IntStartFull(int(p.nObjs))
	Gia_ManCleanLevels(p, p.nObjs)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		Gia_ManChoiceMinLevel_rec(p, Gia_ObjFaninId0(obj, obj.id), 1, mapping)
	}
	return mapping
}
