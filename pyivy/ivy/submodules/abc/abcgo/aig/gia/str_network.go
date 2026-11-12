package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

const (
	STR_SUPER = 100
	MAX_TREE  = 10000
)
const (
	STR_NONE = iota
	STR_CONST0
	STR_PI
	STR_AND
	STR_XOR
	STR_MUX
	STR_BUF
	STR_PO
	STR_UNUSED
)

type Str_Obj_t struct {
	Type, nFanins        uint32
	iOffset, iTop, iCopy int32
	id                   int32
}
type Str_Ntk_t struct {
	nObjs, nObjsAlloc          int32
	pObjs                      []Str_Obj_t
	vFanins                    Vec_Int_t
	nObjCount                  [STR_UNUSED]int32
	nTrees, nGroups, DelayGain int32
}
type Str_Man_t struct {
	pOld              *Gia_Man_t
	nLutSize, fCutMin int32
	pNtk              *Str_Ntk_t
	pNew              *Gia_Man_t
	vDelays           *Vec_Int_t
}

func Str_NtkObj(p *Str_Ntk_t, id int32) *Str_Obj_t {
	if id < 0 || id >= p.nObjs {
		panic("gia: invalid structured object")
	}
	return &p.pObjs[id]
}
func Str_ObjFaninId(p *Str_Ntk_t, obj *Str_Obj_t, i int32) int32 {
	return vec.Vec_IntEntry(&p.vFanins, int(obj.iOffset+i)) >> 1
}
func Str_ObjFanin(p *Str_Ntk_t, obj *Str_Obj_t, i int32) *Str_Obj_t {
	return Str_NtkObj(p, Str_ObjFaninId(p, obj, i))
}
func Str_ObjFaninC(p *Str_Ntk_t, obj *Str_Obj_t, i int32) int32 {
	return vec.Vec_IntEntry(&p.vFanins, int(obj.iOffset+i)) & 1
}
func Str_ObjFaninCopy(p *Str_Ntk_t, obj *Str_Obj_t, i int32) int32 {
	return Str_ObjFanin(p, obj, i).iCopy ^ Str_ObjFaninC(p, obj, i)
}
func Str_ObjId(p *Str_Ntk_t, obj *Str_Obj_t) int32 { return obj.id }
func Str_ObjCreate(p *Str_Ntk_t, kind int32, fanins []int32) int32 {
	if p.nObjs >= p.nObjsAlloc || kind < 0 || kind >= STR_UNUSED {
		panic("gia: invalid structured object allocation")
	}
	id := p.nObjs
	obj := &p.pObjs[id]
	*obj = Str_Obj_t{Type: uint32(kind), nFanins: uint32(len(fanins)), iOffset: int32(vec.Vec_IntSize(&p.vFanins)), iTop: -1, iCopy: -1, id: id}
	for _, lit := range fanins {
		if lit < 0 {
			panic("gia: invalid structured fanin")
		}
		vec.Vec_IntPush(&p.vFanins, lit)
	}
	p.nObjCount[kind]++
	p.nObjs++
	return 2 * id
}
func Str_NtkCreate(objects, fanins int32) *Str_Ntk_t {
	if objects <= 0 || fanins < 0 {
		panic("gia: invalid structured network capacity")
	}
	p := &Str_Ntk_t{nObjsAlloc: objects, pObjs: make([]Str_Obj_t, objects)}
	Str_ObjCreate(p, STR_CONST0, nil)
	vec.Vec_IntGrow(&p.vFanins, int(fanins))
	return p
}
func Str_NtkDelete(p *Str_Ntk_t) { p.pObjs = nil; p.vFanins = Vec_Int_t{} }
func Str_NtkPs(p *Str_Ntk_t, start time.Time) {
	fmt.Printf("Network contains %d ands, %d xors, %d muxes (%d trees in %d groups). Time: %s\n", p.nObjCount[STR_AND], p.nObjCount[STR_XOR], p.nObjCount[STR_MUX], p.nTrees, p.nGroups, time.Since(start))
}
func Str_ObjReadGroup(p *Str_Ntk_t, obj *Str_Obj_t, groups, muxes *int32) {
	*groups, *muxes = 0, 0
	if obj.iTop == 0 {
		return
	}
	top := Str_NtkObj(p, obj.iTop)
	end := Str_NtkObj(p, top.iTop)
	*muxes = top.id - obj.id + 1
	*groups = (end.id - obj.id + 1) / *muxes
}
func Str_NtkPrintGroups(p *Str_Ntk_t) {
	for i := int32(0); i < p.nObjs; i++ {
		obj := Str_NtkObj(p, i)
		if obj.Type == STR_MUX && obj.iTop > 0 {
			var groups, muxes int32
			Str_ObjReadGroup(p, obj, &groups, &muxes)
			i += groups*muxes - 1
			fmt.Printf("%d x %d  ", groups, muxes)
		}
	}
	fmt.Println()
}
func Str_NtkToGia(original *Gia_Man_t, p *Str_Ntk_t) *Gia_Man_t {
	if original.pMuxes != nil {
		panic("gia: expected expanded muxes")
	}
	dst := Gia_ManStart(3 * original.nObjs / 2)
	dst.pName, dst.pSpec = original.pName, original.pSpec
	Gia_ManHashStart(dst)
	for id := int32(0); id < p.nObjs; id++ {
		obj := Str_NtkObj(p, id)
		switch obj.Type {
		case STR_CONST0:
			obj.iCopy = 0
		case STR_PI:
			obj.iCopy = Gia_ManAppendCi(dst)
		case STR_AND:
			obj.iCopy = 1
			for k := int32(0); k < int32(obj.nFanins); k++ {
				obj.iCopy = Gia_ManHashAnd(dst, obj.iCopy, Str_ObjFaninCopy(p, obj, k))
			}
		case STR_XOR:
			obj.iCopy = 0
			for k := int32(0); k < int32(obj.nFanins); k++ {
				obj.iCopy = Gia_ManHashXor(dst, obj.iCopy, Str_ObjFaninCopy(p, obj, k))
			}
		case STR_MUX:
			obj.iCopy = Gia_ManHashMux(dst, Str_ObjFaninCopy(p, obj, 2), Str_ObjFaninCopy(p, obj, 1), Str_ObjFaninCopy(p, obj, 0))
		case STR_PO:
			obj.iCopy = Gia_ManAppendCo(dst, Str_ObjFaninCopy(p, obj, 0))
		default:
			panic("gia: invalid structured object type")
		}
	}
	Gia_ManHashStop(dst)
	Gia_ManSetRegNum(dst, Gia_ManRegNum(original))
	result := Gia_ManCleanup(dst)
	Gia_ManStop(dst)
	return result
}
