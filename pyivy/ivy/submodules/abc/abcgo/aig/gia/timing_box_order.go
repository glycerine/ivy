package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManOrderWithBoxes_rec(p *Gia_Man_t, obj *Gia_Obj_t, nodes *Vec_Int_t) int32 {
	if Gia_ObjIsTravIdCurrent(p, obj) != 0 {
		return 0
	}
	Gia_ObjSetTravIdCurrent(p, obj)
	if Gia_ObjIsCi(obj) != 0 {
		p.iData2 = Gia_ObjCioId(obj)
		return 1
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: invalid boxed timing cone")
	}
	if sib := Gia_ObjSibl(p, obj.id); sib != 0 {
		if Gia_ManOrderWithBoxes_rec(p, Gia_ManObj(p, sib), nodes) != 0 {
			return 1
		}
	}
	for _, id := range giaObjectFanins(p, obj) {
		if Gia_ManOrderWithBoxes_rec(p, Gia_ManObj(p, id), nodes) != 0 {
			return 1
		}
	}
	vec.Vec_IntPush(nodes, obj.id)
	return 0
}
func Gia_ManOrderWithBoxes(p *Gia_Man_t) *Vec_Int_t {
	manager, ok := p.PManTime.(*tim.Tim_Man_t)
	if !ok || manager == nil || Gia_ManIsNormalized(p) == 0 {
		panic("gia: invalid boxed timing manager")
	}
	Gia_ManIncrementTravId(p)
	nodes := vec.Vec_IntAlloc(int(p.nObjs))
	vec.Vec_IntPush(nodes, 0)
	Gia_ObjSetTravIdCurrent(p, Gia_ManConst0(p))
	for i := 0; i < tim.Tim_ManPiNum(manager); i++ {
		obj := Gia_ManCi(p, int32(i))
		vec.Vec_IntPush(nodes, obj.id)
		Gia_ObjSetTravIdCurrent(p, obj)
	}
	currentCI, currentCO := tim.Tim_ManPiNum(manager), 0
	for box := 0; box < tim.Tim_ManBoxNum(manager); box++ {
		for k := 0; k < tim.Tim_ManBoxInputNum(manager, box); k++ {
			obj := Gia_ManCo(p, int32(currentCO+k))
			if Gia_ManOrderWithBoxes_rec(p, Gia_ObjFanin0(obj), nodes) != 0 {
				fmt.Println("Boxes are not in topological order.")
				p.iData2 = 0
				return nil
			}
		}
		for k := 0; k < tim.Tim_ManBoxInputNum(manager, box); k++ {
			vec.Vec_IntPush(nodes, Gia_ManCo(p, int32(currentCO+k)).id)
		}
		currentCO += tim.Tim_ManBoxInputNum(manager, box)
		for k := 0; k < tim.Tim_ManBoxOutputNum(manager, box); k++ {
			obj := Gia_ManCi(p, int32(currentCI+k))
			vec.Vec_IntPush(nodes, obj.id)
			Gia_ObjSetTravIdCurrent(p, obj)
		}
		currentCI += tim.Tim_ManBoxOutputNum(manager, box)
	}
	start := tim.Tim_ManCoNum(manager) - tim.Tim_ManPoNum(manager)
	for i := start; i < tim.Tim_ManCoNum(manager); i++ {
		obj := Gia_ManCo(p, int32(i))
		if Gia_ManOrderWithBoxes_rec(p, Gia_ObjFanin0(obj), nodes) != 0 {
			return nil
		}
	}
	for i := start; i < tim.Tim_ManCoNum(manager); i++ {
		vec.Vec_IntPush(nodes, Gia_ManCo(p, int32(i)).id)
	}
	currentCO += tim.Tim_ManPoNum(manager)
	if currentCI != int(Gia_ManCiNum(p)) || currentCO != int(Gia_ManCoNum(p)) {
		panic("gia: inconsistent boxed timing interface")
	}
	return nodes
}

// Gia_ManDupUnnormalize interleaves box inputs, outputs and their logic so
// each box's outputs follow the logic feeding that box.
func Gia_ManDupUnnormalize(src *Gia_Man_t) *Gia_Man_t {
	if Gia_ManBufNum(src) != 0 {
		panic("gia: boxed unnormalization requires no buffers")
	}
	order := Gia_ManOrderWithBoxes(src)
	if order == nil {
		return nil
	}
	defer vec.Vec_IntFree(order)
	Gia_ManFillValue(src)
	dst := Gia_ManStart(Gia_ManObjNum(src))
	dst.CopyIdentityFrom(src)
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	if Gia_ManHasChoices(src) != 0 {
		dst.pSibls = make([]int32, Gia_ManObjNum(src))
	}
	for _, id := range vec.Vec_IntArray(order) {
		obj := Gia_ManObj(src, id)
		a, b, c, kind := src.Gate(id)
		var result int32
		switch kind {
		case 0:
			result = 0
		case 1:
			result = Gia_ManAppendCi(dst)
		case 2:
			result = Gia_ManAppendAnd(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
		case 3:
			result = Gia_ManAppendXorReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
		case 4:
			result = Gia_ManAppendMuxReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b), Gia_ObjLitCopy(src, c))
		case 5:
			result = Gia_ManAppendBuf(dst, Gia_ObjLitCopy(src, a))
		case 6:
			result = Gia_ManAppendCo(dst, Gia_ObjLitCopy(src, a))
		default:
			panic("gia: unsupported boxed object")
		}
		obj.Value = uint32(result)
		if sibling := Gia_ObjSibl(src, id); sibling != 0 {
			copied := Gia_ManObj(src, sibling).Value
			if copied == ^uint32(0) {
				panic("gia: choice sibling missing from boxed order")
			}
			dst.pSibls[result>>1] = int32(copied >> 1)
		}
	}
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	dst.nConstrs = src.nConstrs
	return dst
}

func Gia_ManUpdateTimMan2(p *Gia_Man_t, boxes *Vec_Int_t, diff int32) any {
	timing := p.PManTime.(*tim.Tim_Man_t)
	ids := make([]int, 0, vec.Vec_IntSize(boxes))
	for _, id := range vec.Vec_IntArray(boxes) {
		ids = append(ids, int(id))
	}
	return tim.Tim_ManReduce(timing, ids, int(diff))
}
func Gia_ManUpdateExtraAig2(pTime any, p *Gia_Man_t, boxes *Vec_Int_t) *Gia_Man_t {
	timing := pTime.(*tim.Tim_Man_t)
	first := tim.Tim_ManPiNum(timing)
	outputs := vec.Vec_IntAlloc(100)
	for _, id := range vec.Vec_IntArray(boxes) {
		start := tim.Tim_ManBoxOutputFirst(timing, int(id)) - first
		for i := 0; i < tim.Tim_ManBoxOutputNum(timing, int(id)); i++ {
			vec.Vec_IntPush(outputs, int32(start+i))
		}
	}
	return Gia_ManDupSelectedOutputs(p, outputs)
}
