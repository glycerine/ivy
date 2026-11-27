package sbd

// Port of src/opt/sbd/sbdPath.c: critical path of a LUT mapping.

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func Sbc_ManAddInternalToPath_rec(p *Gia_Man_t, iObj int32, vPath *Vec_Bit_t) int32 {
	Value := int32(0)
	if Gia_ObjIsTravIdCurrentId(p, iObj) {
		return int32(vec.Vec_BitEntry(vPath, int(iObj)))
	}
	Gia_ObjSetTravIdCurrentId(p, iObj)
	pObj := Gia_ManObj(p, iObj)
	if Gia_ObjIsCi(pObj) {
		return int32(vec.Vec_BitEntry(vPath, int(iObj)))
	}
	assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
	for _, iFan := range gia.Gia_ObjLutFanins(p, iObj) {
		Value |= Sbc_ManAddInternalToPath_rec(p, iFan, vPath)
	}
	if Value != 0 {
		vec.Vec_BitWriteEntry(vPath, int(iObj), 1)
	}
	return Value
}
func Sbc_ManAddInternalToPath(p *Gia_Man_t, vPath *Vec_Bit_t) {
	for iObj := int32(1); iObj < Gia_ManObjNum(p); iObj++ {
		if gia.Gia_ObjIsLut(p, iObj) == 0 {
			continue
		}
		if vec.Vec_BitEntry(vPath, int(iObj)) == 0 {
			continue
		}
		Gia_ManIncrementTravId(p)
		for _, iFan := range gia.Gia_ObjLutFanins(p, iObj) {
			Gia_ObjSetTravIdCurrentId(p, iFan)
		}
		Sbc_ManAddInternalToPath_rec(p, iObj, vPath)
	}
}

func Sbc_ManCriticalPath_rec(p *Gia_Man_t, pLevels []int32, iObj, LevelFan int32, vPath *Vec_Bit_t, Slack int32) {
	if Gia_ObjIsTravIdCurrentId(p, iObj) {
		return
	}
	Gia_ObjSetTravIdCurrentId(p, iObj)
	pObj := Gia_ManObj(p, iObj)
	vec.Vec_BitWriteEntry(vPath, int(iObj), 1)
	if Gia_ObjIsCi(pObj) {
		pManTime, _ := p.BoxTiming()
		iBox := -1
		if pManTime != nil {
			iBox = tim.Tim_ManBoxForCi(pManTime, int(Gia_ObjCioId(pObj)))
		}
		if iBox >= 0 {
			curCo := int32(tim.Tim_ManBoxInputFirst(pManTime, iBox))
			nBoxInputs := int32(tim.Tim_ManBoxInputNum(pManTime, iBox))
			for k := int32(0); k < nBoxInputs; k++ {
				pCo := Gia_ManCo(p, curCo+k)
				iDriver := Gia_ObjFaninId0p(p, pCo)
				if (pLevels[iDriver]+Slack >= LevelFan-1) && iDriver != 0 {
					Sbc_ManCriticalPath_rec(p, pLevels, iDriver, pLevels[iDriver], vPath, max(0, pLevels[iDriver]+Slack-(LevelFan-1)))
				}
			}
		}
		return
	}
	assert(Gia_ObjIsAnd(pObj), "Gia_ObjIsAnd(pObj)")
	for _, iFan := range gia.Gia_ObjLutFanins(p, iObj) {
		if pLevels[iFan]+Slack >= LevelFan-1 {
			Sbc_ManCriticalPath_rec(p, pLevels, iFan, pLevels[iFan], vPath, max(0, pLevels[iFan]+Slack-(LevelFan-1)))
		}
	}
}
func Sbc_ManCriticalPath(p *Gia_Man_t) *Vec_Bit_t {
	var pLevels []int32
	Slack := int32(1)
	pManTime, _ := p.BoxTiming()
	var nLevels int32
	if pManTime != nil {
		nLevels = gia.Gia_ManLutLevelWithBoxes(p)
	} else {
		nLevels = gia.Gia_ManLutLevel(p, &pLevels)
	}
	vPath := vec.Vec_BitStart(int(Gia_ManObjNum(p)))
	if pManTime != nil {
		pLevels = make([]int32, Gia_ManObjNum(p))
		for i := range pLevels {
			pLevels[i] = Gia_ObjLevelId(p, int32(i))
		}
	}
	Gia_ManIncrementTravId(p)
	for _, iDriver := range giaCoDriverIds(p) {
		if (pLevels[iDriver] == nLevels) && iDriver != 0 {
			Sbc_ManCriticalPath_rec(p, pLevels, iDriver, pLevels[iDriver], vPath, Slack)
		}
	}
	Sbc_ManAddInternalToPath(p, vPath)
	return vPath
}

func Sbc_ManDelayTrace(p *Gia_Man_t) {
	vPath := vec.Vec_BitStart(int(Gia_ManObjNum(p)))
	var pLevels []int32
	nLuts, nNodes, nEdges, nEdgesAll := 0, 0, 0, 0
	if gia.Gia_ManHasMapping(p) == 0 {
		fmt.Printf("No mapping is available.\n")
		return
	}
	// set critical CO drivers
	nLevels := gia.Gia_ManLutLevel(p, &pLevels)
	for _, iFan := range giaCoDriverIds(p) {
		if pLevels[iFan] == nLevels {
			vec.Vec_BitWriteEntry(vPath, int(iFan), 1)
		}
	}
	// set critical internal nodes
	for i := Gia_ManObjNum(p) - 1; i > 0; i-- {
		if gia.Gia_ObjIsLut(p, i) == 0 {
			continue
		}
		nLuts++
		if vec.Vec_BitEntry(vPath, int(i)) == 0 {
			continue
		}
		nNodes++
		for _, iFan := range gia.Gia_ObjLutFanins(p, i) {
			if pLevels[iFan]+1 < pLevels[i] {
				continue
			}
			assert(pLevels[iFan]+1 == pLevels[i], "pLevels[iFan] + 1 == pLevels[i]")
			vec.Vec_BitWriteEntry(vPath, int(iFan), 1)
			nEdges++
		}
	}
	for i := int32(1); i < Gia_ManObjNum(p); i++ {
		if gia.Gia_ObjIsLut(p, i) == 0 {
			continue
		}
		for _, iFan := range gia.Gia_ObjLutFanins(p, i) {
			if vec.Vec_BitEntry(vPath, int(i)) != 0 && vec.Vec_BitEntry(vPath, int(iFan)) != 0 {
				nEdgesAll++
			}
		}
	}
	fmt.Printf("AIG = %d. LUT = %d. Lev = %d.   Path nodes = %d.  Path edges = %d. (%d.)\n",
		gia.Gia_ManAndNum(p), nLuts, nLevels, nNodes, nEdges, nEdgesAll)
}
