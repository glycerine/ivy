package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManPrintRepr(p *Gia_Man_t) {
	if len(p.pReprsOld) != int(p.nObjs) || len(p.PReprs) != int(p.nObjs) {
		panic("gia: invalid representatives")
	}
	for i, repr := range p.pReprsOld {
		if repr >= 0 {
			fmt.Printf("%d->%d ", i, p.PReprs[i].iRepr)
		}
	}
	fmt.Println()
}
func Gia_ManDupDfsClasses(src *Gia_Man_t) *Gia_Man_t {
	if len(src.pReprsOld) != int(src.nObjs) {
		panic("gia: missing old representatives")
	}
	Gia_ManFillValue(src)
	out := Gia_ManStart(src.nObjs)
	out.CopyIdentityFrom(src)
	src.pObjs[0].Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(out))
	}
	Gia_ManHashAlloc(out)
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		Gia_ManDupDfs2_rec(out, src, Gia_ManCo(src, i))
	}
	Gia_ManHashStop(out)
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	return giaCleanComposed(out)
}
func Gia_ManDupTopAnd_iter(src *Gia_Man_t, verbose int32) *Gia_Man_t {
	front := make([]int32, 0)
	leaves := make([]int32, 0)
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		obj := Gia_ManCo(src, i)
		lit := Gia_ObjFaninLit0(obj, obj.id)
		if lit>>1 == 0 {
			continue
		}
		if lit&1 != 0 {
			leaves = append(leaves, lit)
		} else {
			front = append(front, lit>>1)
		}
	}
	if len(front) == 0 {
		if verbose != 0 {
			fmt.Println("The AIG cannot be decomposed using AND-decomposition.")
		}
		return Gia_ManDupNormalize(src, 0)
	}
	for i := 0; i < len(front); i++ {
		id := front[i]
		obj := src.pObjs[id]
		if Gia_ObjIsCi(obj) != 0 || id == 0 {
			leaves = append(leaves, id<<1)
			continue
		}
		a, b, _, kind := src.Gate(id)
		if kind != 2 {
			leaves = append(leaves, id<<1)
			continue
		}
		for _, lit := range []int32{a, b} {
			if lit&1 != 0 {
				leaves = append(leaves, lit)
			} else {
				front = append(front, lit>>1)
			}
		}
	}
	ciMap := make([]int32, Gia_ManCiNum(src))
	for i := range ciMap {
		ciMap[i] = -1
	}
	phases := make([]int8, src.nObjs)
	for i := range phases {
		phases[i] = -1
	}
	ciLeaves := 0
	for _, lit := range leaves {
		id := lit >> 1
		phase := int8(lit & 1)
		obj := src.pObjs[id]
		if Gia_ObjIsCi(obj) != 0 {
			ciMap[Gia_ObjCioId(obj)] = int32(phase ^ 1)
			ciLeaves++
		}
		if phases[id] >= 0 && phases[id] != phase {
			fmt.Println("Problem is trivially UNSAT.")
			return Gia_ManDupNormalize(src, 0)
		}
		phases[id] = phase
	}
	selected := vec.Vec_IntAlloc(len(leaves))
	for id, phase := range phases {
		if phase >= 0 && Gia_ObjIsCi(src.pObjs[id]) == 0 {
			vec.Vec_IntPush(selected, int32(id<<1)|int32(phase))
		}
	}
	if verbose != 0 {
		fmt.Printf("Detected %6d AND leaves and %6d CI leaves.\n", vec.Vec_IntSize(selected), ciLeaves)
	}
	if ciLeaves == 0 {
		return Gia_ManDupDfsLitArray(src, selected)
	}
	return Gia_ManDupDfsCiMap(src, ciMap, selected)
}
func Gia_ManDupTopAnd(src *Gia_Man_t, verbose int32) *Gia_Man_t {
	current := Gia_ManDupNormalize(src, 0)
	for iteration := 0; ; iteration++ {
		next := Gia_ManDupTopAnd_iter(current, verbose)
		done := Gia_ManCoNum(next) == Gia_ManCoNum(current) && Gia_ManAndNum(next) == Gia_ManAndNum(current)
		Gia_ManStop(current)
		current = next
		if verbose != 0 {
			fmt.Printf("Iter %2d : outputs=%d ands=%d\n", iteration+1, Gia_ManCoNum(current), Gia_ManAndNum(current))
		}
		if done {
			return current
		}
	}
}
