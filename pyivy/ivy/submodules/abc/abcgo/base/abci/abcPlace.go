package abci

import (
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
	physplace "github.com/glycerine/abcgo/phys/place"
)

var abstractCells []*physplace.AbstractCell
var cells []*physplace.ConcreteCell
var nets []*physplace.ConcreteNet
var nAllocSize int32

// Abc_PlaceCreateCell registers one network object as a placement cell.
func Abc_PlaceCreateCell(obj *Abc_Obj_t, isAnd int32) {
	if obj == nil || obj.ID < 0 || int(obj.ID) >= len(cells) || isAnd < 0 || isAnd >= int32(len(abstractCells)) {
		panic("Abc_PlaceCreateCell: invalid object or placement storage")
	}
	if cells[obj.ID] != nil {
		panic("Abc_PlaceCreateCell: object already has a cell")
	}
	cell := physplace.NewCell(int32(obj.ID), "", abstractCells[isAnd], 0, 0)
	cells[obj.ID] = cell
	physplace.AddCell(cell)
}

// Abc_PlaceUpdateNet replaces the placement net for an ABC object. Net
// terminals are the placement cells for each ABC fanout, matching abcPlace.c.
func Abc_PlaceUpdateNet(obj *Abc_Obj_t) {
	if obj == nil || obj.ID < 0 || int(obj.ID) >= len(nets) {
		panic("Abc_PlaceUpdateNet: invalid object or placement storage")
	}
	terms := make([]*physplace.ConcreteCell, int(baseabc.Abc_ObjFanoutNum(obj)))
	for k := range terms {
		fanout := baseabc.Abc_ObjFanout(obj, int32(k))
		if fanout == nil || fanout.ID < 0 || int(fanout.ID) >= len(cells) || cells[fanout.ID] == nil {
			panic("Abc_PlaceUpdateNet: fanout has no placement cell")
		}
		terms[k] = cells[fanout.ID]
	}
	net := physplace.NewNet(int32(obj.ID), terms, 1)
	nets[obj.ID] = net
	physplace.AddNet(net)
}

// Abc_PlaceEvaluateCut is the placeholder cost function in abcPlace.c.
func Abc_PlaceEvaluateCut(_ *Abc_Obj_t, fanins *Vec_Ptr_t) float32 {
	if fanins != nil {
		for range vec.Vec_PtrArray(fanins) {
			// The original placement cost body is intentionally empty.
		}
	}
	return 0
}

// Abc_PlaceUpdate refreshes newly added cells and modified nets. The
// incremental fastPlace call is commented out in the source routine.
func Abc_PlaceUpdate(addedCells, updatedNets *Vec_Ptr_t) {
	vCells := vec.Vec_PtrAlloc(16)
	vNets := vec.Vec_PtrAlloc(32)
	defer vec.Vec_PtrFree(vCells)
	defer vec.Vec_PtrFree(vNets)
	if addedCells != nil {
		for _, value := range vec.Vec_PtrArray(addedCells) {
			obj, ok := value.(*Abc_Obj_t)
			if !ok || obj == nil {
				panic("Abc_PlaceUpdate: added-cell entry is not an object")
			}
			Abc_PlaceCreateCell(obj, 1)
			Abc_PlaceUpdateNet(obj)
			vec.Vec_PtrPush(vCells, cells[obj.ID])
			for _, fanin := range obj.Fanins() {
				if fanin.Object == nil || fanin.Object.ID < 0 || int(fanin.Object.ID) >= len(nets) {
					panic("Abc_PlaceUpdate: fanin has no placement net slot")
				}
				vec.Vec_PtrPushUnique(vNets, nets[fanin.Object.ID])
			}
		}
	}
	if updatedNets != nil {
		for _, value := range vec.Vec_PtrArray(updatedNets) {
			obj, ok := value.(*Abc_Obj_t)
			if !ok || obj == nil {
				panic("Abc_PlaceUpdate: updated-net entry is not an object")
			}
			if obj.Kind == baseabc.ObjectNone || baseabc.Abc_ObjNtk(obj) == nil {
				continue
			}
			Abc_PlaceUpdateNet(obj)
		}
	}
}

// Abc_PlaceBegin derives the placement cells and nets from a strashed network
// and invokes the placer, as in the source's pre-write hook.
func Abc_PlaceBegin(n *Abc_Ntk_t) {
	if n == nil {
		panic("Abc_PlaceBegin: nil network")
	}
	nAllocSize = 5 * baseabc.Abc_NtkObjNumMax(n)
	cells = make([]*physplace.ConcreteCell, nAllocSize)
	nets = make([]*physplace.ConcreteNet, nAllocSize)
	if abstractCells == nil {
		abstractCells = make([]*physplace.AbstractCell, 2)
	}
	abstractCells[0] = physplace.NewCellType("pio", 1, 1, true)
	abstractCells[1] = physplace.NewCellType("and", 1, 1, false)

	for i := int32(0); i < baseabc.Abc_NtkCiNum(n); i++ {
		Abc_PlaceCreateCell(baseabc.Abc_NtkCi(n, i), 0)
	}
	for i := int32(0); i < baseabc.Abc_NtkCoNum(n); i++ {
		Abc_PlaceCreateCell(baseabc.Abc_NtkCo(n, i), 0)
	}
	for i := int32(0); i < baseabc.Abc_NtkObjNumMax(n); i++ {
		obj := baseabc.Abc_NtkObj(n, i)
		if baseabc.Abc_AigNodeIsAnd(obj) != 0 {
			Abc_PlaceCreateCell(obj, 1)
		}
	}
	for i := int32(0); i < baseabc.Abc_NtkObjNumMax(n); i++ {
		obj := baseabc.Abc_NtkObj(n, i)
		if obj == nil || (baseabc.Abc_ObjIsCi(obj) == 0 && baseabc.Abc_ObjIsNode(obj) == 0) {
			continue
		}
		Abc_PlaceUpdateNet(obj)
	}
	if err := physplace.Preplace(0.8); err != nil {
		panic(err)
	}
	if err := physplace.GlobalPlace(); err != nil {
		panic(err)
	}
}

// Abc_PlaceEnd releases the placement records installed by Abc_PlaceBegin.
func Abc_PlaceEnd(_ *Abc_Ntk_t) {
	for _, net := range nets {
		if net != nil {
			physplace.DeleteNet(net)
		}
	}
	for _, cell := range cells {
		if cell != nil {
			physplace.DeleteCell(cell)
		}
	}
	abstractCells = nil
	cells = nil
	nets = nil
}
