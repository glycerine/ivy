package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManName(p *Gia_Man_t) string        { return p.pName }
func Gia_ManCiNum(p *Gia_Man_t) int32        { return int32(vec.Vec_IntSize(p.vCis)) }
func Gia_ManCoNum(p *Gia_Man_t) int32        { return int32(vec.Vec_IntSize(p.vCos)) }
func Gia_ManPiNum(p *Gia_Man_t) int32        { return Gia_ManCiNum(p) - p.nRegs }
func Gia_ManPoNum(p *Gia_Man_t) int32        { return Gia_ManCoNum(p) - p.nRegs }
func Gia_ManRegNum(p *Gia_Man_t) int32       { return p.nRegs }
func Gia_ManObjNum(p *Gia_Man_t) int32       { return p.nObjs }
func Gia_ManAndNum(p *Gia_Man_t) int32       { return p.nObjs - Gia_ManCiNum(p) - Gia_ManCoNum(p) - 1 }
func Gia_ManXorNum(p *Gia_Man_t) int32       { return p.nXors }
func Gia_ManMuxNum(p *Gia_Man_t) int32       { return p.nMuxes }
func Gia_ManBufNum(p *Gia_Man_t) int32       { return p.nBufs }
func Gia_ManAndNotBufNum(p *Gia_Man_t) int32 { return Gia_ManAndNum(p) - Gia_ManBufNum(p) }
func Gia_ManCandNum(p *Gia_Man_t) int32      { return Gia_ManCiNum(p) + Gia_ManAndNum(p) }
func Gia_ManConstrNum(p *Gia_Man_t) int32    { return p.nConstrs }
func Gia_ManFlipVerbose(p *Gia_Man_t)        { p.fVerbose ^= 1 }
func Gia_ManHasChoices(p *Gia_Man_t) int32   { return giaBool(p.pSibls != nil) }
func Gia_ManChoiceNum(p *Gia_Man_t) int32 {
	count := int32(0)
	for _, sibling := range p.pSibls {
		if sibling > 0 {
			count++
		}
	}
	return count
}
func Gia_ManCi(p *Gia_Man_t, id int32) *Gia_Obj_t {
	return Gia_ManObj(p, vec.Vec_IntEntry(p.vCis, int(id)))
}
func Gia_ManCo(p *Gia_Man_t, id int32) *Gia_Obj_t {
	return Gia_ManObj(p, vec.Vec_IntEntry(p.vCos, int(id)))
}
func Gia_ManPi(p *Gia_Man_t, id int32) *Gia_Obj_t {
	if id < 0 || id >= Gia_ManPiNum(p) {
		panic("gia: invalid PI index")
	}
	return Gia_ManCi(p, id)
}
func Gia_ManPo(p *Gia_Man_t, id int32) *Gia_Obj_t {
	if id < 0 || id >= Gia_ManPoNum(p) {
		panic("gia: invalid PO index")
	}
	return Gia_ManCo(p, id)
}
func Gia_ManRo(p *Gia_Man_t, id int32) *Gia_Obj_t {
	if id < 0 || id >= p.nRegs {
		panic("gia: invalid RO index")
	}
	return Gia_ManCi(p, Gia_ManPiNum(p)+id)
}
func Gia_ManRi(p *Gia_Man_t, id int32) *Gia_Obj_t {
	if id < 0 || id >= p.nRegs {
		panic("gia: invalid RI index")
	}
	return Gia_ManCo(p, Gia_ManPoNum(p)+id)
}

func Gia_ManSetRegNum(p *Gia_Man_t, count int32) {
	if p.nRegs != 0 || count < 0 {
		panic("gia: invalid or already assigned register count")
	}
	p.nRegs = count
}

// Dependency-neutral traversal methods used by higher-level algorithms.
func (p *Gia_Man_t) ObjectCount() int { return int(p.nObjs) }
func (p *Gia_Man_t) InputObjectIDs() []int32 {
	return append([]int32(nil), vec.Vec_IntArray(p.vCis)...)
}
func (p *Gia_Man_t) AndFanins(id int32) (a, b int32, ok bool) {
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsAnd(obj) == 0 {
		return 0, 0, false
	}
	return Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id), true
}

// Gia_ManSetConstrNum sets the trailing ordinary-output constraint count.
func Gia_ManSetConstrNum(p *Gia_Man_t, count int32) {
	if count < 0 || count > Gia_ManPoNum(p) {
		panic("gia: invalid constraint count")
	}
	p.nConstrs = count
}
