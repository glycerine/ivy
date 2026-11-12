package amap

import (
	"strings"
)

func Amap_OuputStrsav(_ *Aig_MmFlex_t, value string) string { return strings.Clone(value) }

func Amap_OutputStructAlloc(_ *Aig_MmFlex_t, gate *Amap_Gat_t) *Amap_Out_t {
	nFans := 1
	var name string
	if gate != nil {
		nFans = int(gate.nPins)
		name = strings.Clone(gate.pName)
	}
	fans := make([]int32, nFans)
	for i := range fans {
		fans[i] = -1
	}
	return &Amap_Out_t{pName: name, nFans: int16(nFans), pFans: fans}
}

// Amap_ManProduceMapped serializes only objects selected by the mapper. The
// last vector entry remains the flexible-manager marker consumed by the ABC
// network adapter; Go-owned output records and slices do not depend on it.
func Amap_ManProduceMapped(p *Amap_Man_t) *Vec_Ptr_t {
	if p == nil || p.pLib == nil {
		return nil
	}
	mem := Aig_MmFlexStart()
	vNodes := Vec_PtrAlloc(int32(Amap_ManObjNum(p) + 1))
	for i := int32(0); i < Vec_PtrSize(p.vObjs); i++ {
		obj, ok := Vec_PtrEntry(p.vObjs, i).(*Amap_Obj_t)
		if !ok || obj == nil {
			continue
		}
		switch {
		case Amap_ObjIsPi(obj) != 0:
			assert(obj.fPolar == 0, "PI polarity is positive")
			res := Amap_OutputStructAlloc(mem, nil)
			res.Type, res.nFans = -1, 0
			obj.iData = int32(Vec_PtrSize(vNodes))
			Vec_PtrPush(vNodes, res)
			if obj.nFouts[1] != 0 {
				assert(p.pLib.pGateInv != nil, "negative PI use requires an inverter gate")
				inv := Amap_OutputStructAlloc(mem, p.pLib.pGateInv)
				inv.pFans[0] = obj.iData
				Vec_PtrPush(vNodes, inv)
			}

		case Amap_ObjIsNode(obj) != 0:
			if Amap_ObjRefsTotal(obj) == 0 {
				continue
			}
			gate := Amap_LibGate(p.pLib, int32(obj.Best.pSet.iGate))
			assert(gate != nil && int(gate.nPins) == len(obj.Best.pCut.Fans), "gate inputs match cut fanins")
			res := Amap_OutputStructAlloc(mem, gate)
			cut, set := obj.Best.pCut, obj.Best.pSet
			for k := 0; k < len(cut.Fans); k++ {
				setLit := int32(set.Ins[k])
				cutFan := cut.Fans[Abc_Lit2Var(setLit)]
				fanin := Amap_ManObj(p, Abc_Lit2Var(cutFan))
				assert(Amap_ObjRefsTotal(fanin) != 0, "matched fanin has references")
				complement := Abc_LitIsCompl(setLit) ^ Abc_LitIsCompl(cutFan)
				index := fanin.iData
				if int32(fanin.fPolar) != complement {
					index++
				}
				res.pFans[k] = index
			}
			obj.iData = int32(Vec_PtrSize(vNodes))
			Vec_PtrPush(vNodes, res)
			opposite := 1 - int(obj.fPolar)
			if obj.nFouts[opposite] != 0 {
				assert(p.pLib.pGateInv != nil, "opposite-polarity use requires an inverter gate")
				inv := Amap_OutputStructAlloc(mem, p.pLib.pGateInv)
				inv.pFans[0] = obj.iData
				Vec_PtrPush(vNodes, inv)
			}

		case Amap_ObjIsPo(obj) != 0:
			assert(obj.fPolar == 0, "PO polarity is positive")
			fanin := Amap_ObjFanin0(p, obj)
			assert(Amap_ObjRefsTotal(fanin) != 0, "PO driver has references")
			faninIndex := int32(0)
			if Amap_ObjIsConst1(fanin) != 0 {
				gate := p.pLib.pGate1
				if Amap_ObjFaninC0(obj) != 0 {
					gate = p.pLib.pGate0
				}
				assert(gate != nil, "constant output requires a constant gate")
				constant := Amap_OutputStructAlloc(mem, gate)
				faninIndex = int32(Vec_PtrSize(vNodes))
				Vec_PtrPush(vNodes, constant)
			} else if int32(fanin.fPolar) == Amap_ObjFaninC0(obj) {
				faninIndex = fanin.iData
			} else {
				faninIndex = fanin.iData + 1
			}
			res := Amap_OutputStructAlloc(mem, nil)
			res.Type, res.pFans[0] = 1, faninIndex
			Vec_PtrPush(vNodes, res)
		}
	}
	Vec_PtrPush(vNodes, mem)
	return vNodes
}
