package gia

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/util"
)

// Hcd_Man_t owns the combinational equivalence simulation workspace.
type Hcd_Man_t struct {
	pGia                            *Gia_Man_t
	nBTLimit, fVerbose              int32
	pSimInfo                        []uint32
	vSimInfo, vSimPres              *Vec_Ptr_t
	vClassOld, vClassNew, vRefinedC *Vec_Int_t
}

func Hcd_ObjSim(p *Hcd_Man_t, id int32) uint32                  { return p.pSimInfo[id] }
func Hcd_ObjSimP(p *Hcd_Man_t, id int32) []uint32               { return p.pSimInfo[id : id+1 : id+1] }
func Hcd_ObjSetSim(p *Hcd_Man_t, id int32, value uint32) uint32 { p.pSimInfo[id] = value; return value }

func Gia_ManEquivStart(g *Gia_Man_t, limit, verbose int32) *Hcd_Man_t {
	p := &Hcd_Man_t{pGia: g, nBTLimit: limit, fVerbose: verbose, pSimInfo: make([]uint32, g.nObjs), vClassOld: Vec_IntAlloc(100), vClassNew: Vec_IntAlloc(100), vRefinedC: Vec_IntAlloc(100), vSimInfo: Vec_PtrAlloc(int(Gia_ManCiNum(g)))}
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		Vec_PtrPush(p.vSimInfo, Hcd_ObjSimP(p, Gia_ManCi(g, i).id))
	}
	p.vSimPres = Vec_PtrAllocSimInfo(int(Gia_ManCiNum(g)), 1)
	return p
}
func Gia_ManEquivStop(p *Hcd_Man_t) {
	if p != nil {
		*p = Hcd_Man_t{}
	}
}
func Hcd_ManCompareEqual(a, b uint32) int32 { return giaBool(a == b || a == ^b) }
func Hcd_ManCompareConst(a uint32) int32    { return giaBool(a == 0 || a == ^uint32(0)) }

func Hcd_ManClassCreate(g *Gia_Man_t, class *Vec_Int_t) {
	if Vec_IntSize(class) == 0 {
		panic("gia: empty equivalence class")
	}
	ids := Vec_IntArray(class)
	Gia_ObjSetRepr(g, ids[0], GIA_VOID)
	for i := 1; i < len(ids); i++ {
		Gia_ObjSetRepr(g, ids[i], ids[0])
		Gia_ObjSetNext(g, ids[i-1], ids[i])
	}
	Gia_ObjSetNext(g, ids[len(ids)-1], 0)
}
func Hcd_ManClassClassRemoveOne(p *Hcd_Man_t, id int32) int32 {
	g := p.pGia
	if Gia_ObjIsConst(g, id) != 0 {
		Gia_ObjSetRepr(g, id, GIA_VOID)
		return 1
	}
	if Gia_ObjIsClass(g, id) == 0 {
		return 0
	}
	head := Gia_ObjRepr(g, id)
	if head == GIA_VOID {
		head = id
	}
	old, newClass := p.vClassOld, p.vClassNew
	Vec_IntClear(old)
	Vec_IntClear(newClass)
	for member := head; member != 0; member = Gia_ObjNext(g, member) {
		if member == id {
			Vec_IntPush(newClass, member)
		} else {
			Vec_IntPush(old, member)
		}
	}
	Hcd_ManClassCreate(g, old)
	Hcd_ManClassCreate(g, newClass)
	return 1
}
func Hcd_ManClassRefineOne(p *Hcd_Man_t, head int32) int32 {
	old, newClass := p.vClassOld, p.vClassNew
	Vec_IntClear(old)
	Vec_IntClear(newClass)
	Vec_IntPush(old, head)
	for id := Gia_ObjNext(p.pGia, head); id != 0; id = Gia_ObjNext(p.pGia, id) {
		if Hcd_ManCompareEqual(Hcd_ObjSim(p, head), Hcd_ObjSim(p, id)) != 0 {
			Vec_IntPush(old, id)
		} else {
			Vec_IntPush(newClass, id)
		}
	}
	if Vec_IntSize(newClass) == 0 {
		return 0
	}
	Hcd_ManClassCreate(p.pGia, old)
	Hcd_ManClassCreate(p.pGia, newClass)
	if Vec_IntSize(newClass) > 1 {
		return 1 + Hcd_ManClassRefineOne(p, Vec_IntEntry(newClass, 0))
	}
	return 1
}
func Hcd_ManHashKey(sim []uint32, words, tableSize int32) int32 {
	primes := [16]uint32{1291, 1699, 1999, 2357, 2953, 3313, 3907, 4177, 4831, 5147, 5647, 6343, 6899, 7103, 7873, 8147}
	if words <= 0 || int(words) > len(sim) || tableSize <= 0 {
		panic("gia: invalid simulation hash")
	}
	var hash uint32
	for i := int32(0); i < words; i++ {
		v := sim[i]
		if sim[0]&1 != 0 {
			v = ^v
		}
		hash ^= v * primes[i&15]
	}
	return int32(hash % uint32(tableSize))
}
func Hcd_ManClassesRehash(p *Hcd_Man_t, refined *Vec_Int_t) {
	g := p.pGia
	size := int32(util.Abc_PrimeCudd(uint32(100 + Vec_IntSize(refined)/5)))
	table := make([]int32, size)
	for _, id := range Vec_IntArray(refined) {
		key := Hcd_ManHashKey(Hcd_ObjSimP(p, id), 1, size)
		previous := table[key]
		if previous == 0 {
			Gia_ObjSetRepr(g, id, GIA_VOID)
		} else {
			Gia_ObjSetNext(g, previous, id)
			repr := Gia_ObjRepr(g, previous)
			if repr == GIA_VOID {
				repr = previous
			}
			Gia_ObjSetRepr(g, id, repr)
		}
		table[key] = id
	}
	for _, id := range Vec_IntArray(refined) {
		if Gia_ObjIsHead(g, id) != 0 {
			Hcd_ManClassRefineOne(p, id)
		}
	}
	Gia_ManEquivPrintClasses(g, 0, 0)
}
func Hcd_ManClassesRefine(p *Hcd_Man_t) {
	Vec_IntClear(p.vRefinedC)
	for id := int32(1); id < p.pGia.nObjs; id++ {
		if Gia_ObjIsAnd(p.pGia.pObjs[id]) == 0 {
			continue
		}
		if Gia_ObjIsTail(p.pGia, id) != 0 {
			Hcd_ManClassRefineOne(p, Gia_ObjRepr(p.pGia, id))
		} else if Gia_ObjIsConst(p.pGia, id) != 0 && Hcd_ManCompareConst(Hcd_ObjSim(p, id)) == 0 {
			Vec_IntPush(p.vRefinedC, id)
		}
	}
	Hcd_ManClassesRehash(p, p.vRefinedC)
}
func Hcd_ManClassesCreate(p *Hcd_Man_t) {
	g := p.pGia
	if g.PReprs != nil {
		panic("gia: equivalence classes already exist")
	}
	g.PReprs = make([]Gia_Rpr_t, g.nObjs)
	g.PNexts = make([]int32, g.nObjs)
	for id, obj := range g.pObjs {
		repr := int32(GIA_VOID)
		if Gia_ObjIsAnd(obj) != 0 {
			repr = 0
		}
		Gia_ObjSetRepr(g, int32(id), repr)
	}
}
func Hcd_ManSimulationInit(p *Hcd_Man_t) {
	for i := int32(0); i < Gia_ManCiNum(p.pGia); i++ {
		Hcd_ObjSetSim(p, Gia_ManCi(p.pGia, i).id, Gia_ManRandom(0)<<1)
	}
}

// The original simulator assumes ordinary ANDs. Read logical fanins so real
// gates and buffers retain their functions in both simulation entry points.
func giarfSimulateGate(g *Gia_Man_t, obj *Gia_Obj_t, read func(int32) uint32) uint32 {
	a, b, c, kind := g.Gate(obj.id)
	lit := func(literal int32) uint32 {
		value := read(literal >> 1)
		if literal&1 != 0 {
			value = ^value
		}
		return value
	}
	switch kind {
	case 2:
		return lit(a) & lit(b)
	case 3:
		return lit(a) ^ lit(b)
	case 4:
		control := lit(a)
		return control&lit(b) | ^control&lit(c)
	case 5:
		return lit(a)
	default:
		panic("gia: invalid internal simulation node")
	}
}

func Hcd_ManSimulateSimple(p *Hcd_Man_t) {
	read := func(id int32) uint32 { return Hcd_ObjSim(p, id) }
	for id := int32(1); id < p.pGia.nObjs; id++ {
		obj := p.pGia.pObjs[id]
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		Hcd_ObjSetSim(p, id, giarfSimulateGate(p.pGia, obj, read))
	}
}
func Gia_Resimulate_rec(p *Hcd_Man_t, id int32) uint32 {
	g := p.pGia
	if Gia_ObjIsTravIdCurrentId(g, id) != 0 {
		return Hcd_ObjSim(p, id)
	}
	Gia_ObjSetTravIdCurrentId(g, id)
	obj := g.pObjs[id]
	if Gia_ObjIsCi(obj) != 0 || id == 0 {
		return Hcd_ObjSim(p, id)
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: invalid resimulation node")
	}
	value := giarfSimulateGate(g, obj, func(fanin int32) uint32 { return Gia_Resimulate_rec(p, fanin) })
	return Hcd_ObjSetSim(p, id, value)
}
func Gia_ResimulateAndRefine(p *Hcd_Man_t, head int32) {
	Gia_ManIncrementTravId(p.pGia)
	for id := head; id != 0; id = Gia_ObjNext(p.pGia, id) {
		Gia_Resimulate_rec(p, id)
	}
	if Hcd_ManClassRefineOne(p, head) == 0 {
		fmt.Println("!!! no refinement !!!")
	}
}
