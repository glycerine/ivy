package gia

import (
	"fmt"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

// Gia_ManMin_t holds cones whose conjunction is minimized with BDD support.
type Gia_ManMin_t struct {
	pGia        *Gia_Man_t
	iLits       [2]int32
	vCis, vObjs [2]*Vec_Int_t
	vCleared    *Vec_Int_t
	dd          *core.DdManager
}

func Gia_ManSuppStart(p *Gia_Man_t) *Gia_ManMin_t {
	m := &Gia_ManMin_t{pGia: p, vCleared: Vec_IntAlloc(512), dd: core.Cudd_Init(0, 0, 0, 0, 0)}
	for i := range m.vCis {
		m.vCis[i], m.vObjs[i] = Vec_IntAlloc(512), Vec_IntAlloc(512)
	}
	return m
}
func Gia_ManSuppStop(p *Gia_ManMin_t) { core.Cudd_Quit(p.dd); p.dd = nil }
func Gia_ManGatherSupp_rec(p *Gia_Man_t, id int32, cis, objs *Vec_Int_t) int32 {
	return Gia_Min2ManGatherSupp_rec(p, id, cis, objs)
}
func Gia_ManGatherSupp(p *Gia_ManMin_t) int32 {
	Gia_ManIncrementTravId(p.pGia)
	var overlap int32
	for i, lit := range p.iLits {
		Vec_IntClear(p.vCis[i])
		Vec_IntClear(p.vObjs[i])
		Gia_ManIncrementTravId(p.pGia)
		overlap = Gia_ManGatherSupp_rec(p.pGia, lit>>1, p.vCis[i], p.vObjs[i])
		if i == 0 && overlap != 0 {
			panic("gia: first support cone overlaps previous traversal")
		}
	}
	return overlap
}

func Gia_ManFindRemoved(p *Gia_ManMin_t) int32 {
	funcs := make([]*core.DdNode, Gia_ManObjNum(p.pGia))
	inputs := make([]int32, 0, Vec_IntSize(p.vCis[0])+Vec_IntSize(p.vCis[1]))
	for _, cis := range p.vCis {
		for _, id := range Vec_IntArray(cis) {
			funcs[id] = core.Cudd_bddIthVar(p.dd, len(inputs))
			inputs = append(inputs, id)
		}
	}
	for _, nodes := range p.vObjs {
		for _, id := range Vec_IntArray(nodes) {
			obj := Gia_ManObj(p.pGia, id)
			a := core.Cudd_NotCond(funcs[Gia_ObjFaninId0(obj, id)], Gia_ObjFaninC0(obj) != 0)
			b := core.Cudd_NotCond(funcs[Gia_ObjFaninId1(obj, id)], Gia_ObjFaninC1(obj) != 0)
			funcs[id] = core.Cudd_bddAndLimit(p.dd, a, b, 100000)
			if funcs[id] == nil {
				panic("gia: BDD node limit reached")
			}
		}
	}
	a := core.Cudd_NotCond(funcs[p.iLits[0]>>1], p.iLits[0]&1 != 0)
	b := core.Cudd_NotCond(funcs[p.iLits[1]>>1], p.iLits[1]&1 != 0)
	root := core.Cudd_bddAndLimit(p.dd, a, b, 100000)
	if root == nil {
		panic("gia: BDD node limit reached")
	}
	present := make([]bool, len(inputs))
	for _, i := range p.dd.Support(root) {
		present[i] = true
	}
	Vec_IntClear(p.vCleared)
	for i, id := range inputs {
		if !present[i] {
			Vec_IntPush(p.vCleared, id)
		}
	}
	return int32(Vec_IntSize(p.vCleared))
}

func Gia_ManRebuildOne(p *Gia_ManMin_t, n int32) int32 {
	copies := &p.pGia.vCopies
	Vec_IntFillExtra(copies, int(Gia_ManObjNum(p.pGia)), -1)
	for _, id := range Vec_IntArray(p.vCis[n]) {
		Vec_IntWriteEntry(copies, int(id), id<<1)
	}
	for _, id := range Vec_IntArray(p.vCleared) {
		Vec_IntWriteEntry(copies, int(id), 0)
	}
	if Vec_IntSize(p.vCleared) == 0 || p.iLits[n] < 2 {
		panic("gia: invalid support rebuild")
	}
	result := Vec_IntEntry(copies, int(p.iLits[n]>>1))
	for _, id := range Vec_IntArray(p.vObjs[n]) {
		obj := Gia_ManObj(p.pGia, id)
		a := Vec_IntEntry(copies, int(Gia_ObjFaninId0(obj, id))) ^ Gia_ObjFaninC0(obj)
		b := Vec_IntEntry(copies, int(Gia_ObjFaninId1(obj, id))) ^ Gia_ObjFaninC1(obj)
		result = Gia_ManHashAnd(p.pGia, a, b)
		Vec_IntWriteEntry(copies, int(id), result)
	}
	return result ^ (p.iLits[n] & 1)
}

func Gia_ManSupportAnd(p *Gia_ManMin_t, a, b int32) int32 {
	p.iLits = [2]int32{a, b}
	if a < 2 || b < 2 || Gia_ManGatherSupp(p) == 0 || Gia_ManFindRemoved(p) == 0 {
		return Gia_ManHashAnd(p.pGia, a, b)
	}
	return Gia_ManHashAnd(p.pGia, Gia_ManRebuildOne(p, 0), Gia_ManRebuildOne(p, 1))
}
func Gia_ManSupportAndTest(src *Gia_Man_t) *Gia_Man_t {
	Gia_ManFillValue(src)
	dst := Gia_ManStart(Gia_ManObjNum(src))
	dst.pName, dst.pSpec = src.pName, src.pSpec
	Gia_ManHashAlloc(dst)
	Gia_ManConst0(src).Value = 0
	m := Gia_ManSuppStart(dst)
	for i, obj := range src.pObjs[1:] {
		var lit int32
		switch {
		case Gia_ObjIsAnd(obj) != 0:
			lit = Gia_ManSupportAnd(m, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
		case Gia_ObjIsCi(obj) != 0:
			lit = Gia_ManAppendCi(dst)
		case Gia_ObjIsCo(obj) != 0:
			lit = Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj))
		default:
			panic("gia: unsupported support-minimization node")
		}
		obj.Value = uint32(lit)
		if (i+1)%10000 == 0 {
			fmt.Println(i + 1)
		}
	}
	Gia_ManSuppStop(m)
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	return Gia_ManCleanup(dst)
}
