package gia

// Line-by-line port of src/aig/gia/giaEdge.c (edge-related procedures).
// The C asserts are compiled out in release builds; the port follows the
// release behaviour (for example, LUTs with more than four fanins are
// evaluated over all their fanins).

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ObjEdgeCount(iObj int32, vEdge1, vEdge2 *Vec_Int_t) int32 {
	return giaBool(vec.Vec_IntEntry(vEdge1, int(iObj)) > 0) + giaBool(vec.Vec_IntEntry(vEdge2, int(iObj)) > 0)
}
func Gia_ObjEdgeAdd(iObj, iNext int32, vEdge1, vEdge2 *Vec_Int_t) int32 {
	RetValue := int32(0)
	if vec.Vec_IntEntry(vEdge1, int(iObj)) == 0 {
		vec.Vec_IntWriteEntry(vEdge1, int(iObj), iNext)
	} else if vec.Vec_IntEntry(vEdge2, int(iObj)) == 0 {
		vec.Vec_IntWriteEntry(vEdge2, int(iObj), iNext)
	} else {
		RetValue = 1
	}
	return RetValue
}
func Gia_ObjEdgeRemove(iObj, iNext int32, vEdge1, vEdge2 *Vec_Int_t) {
	if vec.Vec_IntEntry(vEdge1, int(iObj)) == iNext {
		vec.Vec_IntWriteEntry(vEdge1, int(iObj), vec.Vec_IntEntry(vEdge2, int(iObj)))
	}
	vec.Vec_IntWriteEntry(vEdge2, int(iObj), 0)
}
func Gia_ObjEdgeClean(iObj int32, vEdge1, vEdge2 *Vec_Int_t) {
	vec.Vec_IntWriteEntry(vEdge1, int(iObj), 0)
	vec.Vec_IntWriteEntry(vEdge2, int(iObj), 0)
}

// Gia_ManEdgeFromArray transforms edge assignment.
func Gia_ManEdgeFromArray(p *Gia_Man_t, vArray *Vec_Int_t) {
	Count := int32(0)
	p.VEdge1 = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	p.VEdge2 = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	data := vec.Vec_IntArray(vArray)
	for i := 0; i+1 < len(data); i += 2 {
		iObj1, iObj2 := data[i], data[i+1]
		Count += Gia_ObjEdgeAdd(iObj1, iObj2, p.VEdge1, p.VEdge2)
		Count += Gia_ObjEdgeAdd(iObj2, iObj1, p.VEdge1, p.VEdge2)
	}
	if Count != 0 {
		fmt.Printf("Found %d violations during edge conversion.\n", Count)
	}
}
func Gia_ManEdgeToArray(p *Gia_Man_t) *Vec_Int_t {
	vArray := vec.Vec_IntAlloc(1000)
	for iObj := int32(0); iObj < Gia_ManObjNum(p); iObj++ {
		iFanin := vec.Vec_IntEntry(p.VEdge1, int(iObj))
		if iFanin != 0 && iFanin < iObj {
			vec.Vec_IntPushTwo(vArray, iFanin, iObj)
		}
		iFanin = vec.Vec_IntEntry(p.VEdge2, int(iObj))
		if iFanin != 0 && iFanin < iObj {
			vec.Vec_IntPushTwo(vArray, iFanin, iObj)
		}
	}
	return vArray
}

// Gia_ManConvertPackingToEdges derives edges from the packing. Only blocks
// of two or three LUTs produce edges; C stores at most four nodes of a block
// (nNodes[4], see CBUGS_FIXED.md #30).
func Gia_ManConvertPackingToEdges(p *Gia_Man_t) {
	var nNodes [4]int32
	Count := int32(0)
	if p.VPacking == nil {
		return
	}
	p.VEdge1 = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	p.VEdge2 = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	// iterate through structures
	for i := 1; i < vec.Vec_IntSize(p.VPacking); i++ {
		Entry := vec.Vec_IntEntry(p.VPacking, i)
		i++
		for k := int32(0); k < Entry; k, i = k+1, i+1 {
			if k < 4 {
				nNodes[k] = vec.Vec_IntEntry(p.VPacking, i)
			}
		}
		i--
		// create edges
		if Entry == 2 {
			Count += Gia_ObjEdgeAdd(nNodes[0], nNodes[1], p.VEdge1, p.VEdge2)
			Count += Gia_ObjEdgeAdd(nNodes[1], nNodes[0], p.VEdge1, p.VEdge2)
		} else if Entry == 3 {
			Count += Gia_ObjEdgeAdd(nNodes[0], nNodes[2], p.VEdge1, p.VEdge2)
			Count += Gia_ObjEdgeAdd(nNodes[2], nNodes[0], p.VEdge1, p.VEdge2)
			Count += Gia_ObjEdgeAdd(nNodes[1], nNodes[2], p.VEdge1, p.VEdge2)
			Count += Gia_ObjEdgeAdd(nNodes[2], nNodes[1], p.VEdge1, p.VEdge2)
		}
	}
	if Count != 0 {
		fmt.Printf("Skipped %d illegal edges.\n", Count)
	}
}

// Gia_ObjHaveEdge evaluates given edge assignment.
func Gia_ObjHaveEdge(p *Gia_Man_t, iObj, iNext int32) int32 {
	return giaBool(vec.Vec_IntEntry(p.VEdge1, int(iObj)) == iNext || vec.Vec_IntEntry(p.VEdge2, int(iObj)) == iNext)
}
func Gia_ObjCheckEdge(p *Gia_Man_t, iObj, iNext int32) int32 {
	return Gia_ObjHaveEdge(p, iObj, iNext)
}

// giaEdgeLutFanins is Gia_LutForEachFanin for a mapped LUT, else
// Gia_LutForEachFanin2 (the C branches of the edge-delay routines).
func giaEdgeLutFanins(p *Gia_Man_t, iObj int32) []int32 {
	if Gia_ManHasMapping(p) != 0 && Gia_ObjIsLut(p, iObj) != 0 {
		return Gia_ObjLutFanins(p, iObj)
	} else if Gia_ObjIsLut2(p, iObj) != 0 {
		return vec.Vec_IntArray(Gia_ObjLutFanins2(p, iObj))
	}
	return nil // assert( 0 )
}
func Gia_ObjEvalEdgeDelay(p *Gia_Man_t, iObj int32, vDelay *Vec_Int_t) int32 {
	nEdgeDelay := int32(2)
	DelayMax := int32(0)
	for _, iFan := range giaEdgeLutFanins(p, iObj) {
		Delay := vec.Vec_IntEntry(vDelay, int(iFan))
		if Gia_ObjHaveEdge(p, iObj, iFan) != 0 {
			Delay += nEdgeDelay
		} else {
			Delay += 10
		}
		DelayMax = max(DelayMax, Delay)
	}
	return DelayMax
}

func giaEdgeTimMan(p *Gia_Man_t) *tim.Tim_Man_t {
	if p.PManTime == nil {
		return nil
	}
	return p.PManTime.(*tim.Tim_Man_t)
}

// giaEdgeForEachBoxOrder runs the timing-box traversal shared by
// Gia_ManEvalEdgeDelay and Gia_ManComputeEdgeDelay.
func giaEdgeForEachBoxOrder(p *Gia_Man_t, isLut func(int32) bool, lut func(int32)) {
	pTim := giaEdgeTimMan(p)
	vNodes := Gia_ManOrderWithBoxes(p)
	tim.Tim_ManIncrementTravId(pTim)
	for _, iLut := range vec.Vec_IntArray(vNodes) {
		pObj := Gia_ManObj(p, iLut)
		if Gia_ObjIsAnd(pObj) != 0 {
			if isLut(iLut) {
				lut(iLut)
			}
		} else if Gia_ObjIsCi(pObj) != 0 {
			arrTime := tim.Tim_ManGetCiArrival(pTim, int(Gia_ObjCioId(pObj)))
			vec.Vec_IntWriteEntry(p.vEdgeDelay, int(iLut), int32(arrTime))
		} else if Gia_ObjIsCo(pObj) != 0 {
			arrTime := vec.Vec_IntEntry(p.vEdgeDelay, int(Gia_ObjFaninId0(pObj, iLut)))
			tim.Tim_ManSetCoArrival(pTim, int(Gia_ObjCioId(pObj)), float32(arrTime))
		}
	}
}

// giaEdgeForEachLut is the shared body of Gia_ManEvalEdgeDelay and
// Gia_ManComputeEdgeDelay: the mapping (or the secondary mapping) in object
// order, or in timing-box order when boxes are present.
func giaEdgeForEachLut(p *Gia_Man_t, lut func(int32)) {
	pTim := giaEdgeTimMan(p)
	if Gia_ManHasMapping(p) != 0 {
		if pTim != nil && tim.Tim_ManBoxNum(pTim) != 0 {
			giaEdgeForEachBoxOrder(p, func(i int32) bool { return Gia_ObjIsLut(p, i) != 0 }, lut)
		} else {
			for iLut := int32(1); iLut < Gia_ManObjNum(p); iLut++ {
				if Gia_ObjIsLut(p, iLut) != 0 {
					lut(iLut)
				}
			}
		}
	} else if Gia_ManHasMapping2(p) != 0 {
		if pTim != nil && tim.Tim_ManBoxNum(pTim) != 0 {
			giaEdgeForEachBoxOrder(p, func(i int32) bool { return Gia_ObjIsLut2(p, i) != 0 }, lut)
		} else {
			for iLut := int32(1); iLut < Gia_ManObjNum(p); iLut++ {
				if Gia_ObjIsLut2(p, iLut) != 0 {
					lut(iLut)
				}
			}
		}
	}
}
func giaEdgeCoDriverDelayMax(p *Gia_Man_t) int32 {
	DelayMax := int32(0)
	for k := int32(0); k < Gia_ManCoNum(p); k++ {
		iLut := Gia_ObjFaninId0p(p, Gia_ManCo(p, k))
		DelayMax = max(DelayMax, vec.Vec_IntEntry(p.vEdgeDelay, int(iLut)))
	}
	return DelayMax
}
func Gia_ManEvalEdgeDelay(p *Gia_Man_t) int32 {
	p.vEdgeDelay = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	giaEdgeForEachLut(p, func(iLut int32) {
		vec.Vec_IntWriteEntry(p.vEdgeDelay, int(iLut), Gia_ObjEvalEdgeDelay(p, iLut, p.vEdgeDelay))
	})
	return giaEdgeCoDriverDelayMax(p)
}
func Gia_ManEvalEdgeCount(p *Gia_Man_t) int32 {
	return int32((vec.Vec_IntCountPositive(p.VEdge1) + vec.Vec_IntCountPositive(p.VEdge2)) / 2)
}

// Gia_ObjComputeEdgeDelay finds edge assignment.
func Gia_ObjComputeEdgeDelay(p *Gia_Man_t, iObj int32, vDelay, vEdge1, vEdge2 *Vec_Int_t, fUseTwo int32) int32 {
	var Status1, Status2 int32
	DelayMax, DelayMax2, nCountMax := int32(0), int32(0), int32(0)
	iFanMax1, iFanMax2 := int32(-1), int32(-1)
	vec.Vec_IntWriteEntry(vEdge1, int(iObj), 0)
	vec.Vec_IntWriteEntry(vEdge2, int(iObj), 0)
	for _, iFan := range giaEdgeLutFanins(p, iObj) {
		Delay := vec.Vec_IntEntry(vDelay, int(iFan)) + 10
		if DelayMax < Delay {
			DelayMax2 = DelayMax
			DelayMax = Delay
			iFanMax1 = iFan
			nCountMax = 1
		} else if DelayMax == Delay {
			iFanMax2 = iFan
			nCountMax++
			if fUseTwo == 0 {
				DelayMax2 = DelayMax
			}
		} else {
			DelayMax2 = max(DelayMax2, Delay)
		}
	}
	if DelayMax <= 10 {
		// skip first level
	} else if nCountMax == 1 {
		Status1 = Gia_ObjEdgeCount(iFanMax1, vEdge1, vEdge2)
		if Status1 <= 1 {
			Gia_ObjEdgeAdd(iFanMax1, iObj, vEdge1, vEdge2)
			Gia_ObjEdgeAdd(iObj, iFanMax1, vEdge1, vEdge2)
			DelayMax = max(DelayMax2, DelayMax-8)
			vec.Vec_IntWriteEntry(vDelay, int(iObj), DelayMax)
			return DelayMax
		}
	} else if fUseTwo != 0 && nCountMax == 2 {
		Status1 = Gia_ObjEdgeCount(iFanMax1, vEdge1, vEdge2)
		Status2 = Gia_ObjEdgeCount(iFanMax2, vEdge1, vEdge2)
		if Status1 <= 1 && Status2 <= 1 {
			Gia_ObjEdgeAdd(iFanMax1, iObj, vEdge1, vEdge2)
			Gia_ObjEdgeAdd(iFanMax2, iObj, vEdge1, vEdge2)
			Gia_ObjEdgeAdd(iObj, iFanMax1, vEdge1, vEdge2)
			Gia_ObjEdgeAdd(iObj, iFanMax2, vEdge1, vEdge2)
			DelayMax = max(DelayMax2, DelayMax-8)
			vec.Vec_IntWriteEntry(vDelay, int(iObj), DelayMax)
			return DelayMax
		}
	}
	vec.Vec_IntWriteEntry(vDelay, int(iObj), DelayMax)
	return DelayMax
}
func Gia_ManComputeEdgeDelay(p *Gia_Man_t, fUseTwo int32) int32 {
	p.VEdge1 = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	p.VEdge2 = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	p.vEdgeDelay = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	giaEdgeForEachLut(p, func(iLut int32) {
		Gia_ObjComputeEdgeDelay(p, iLut, p.vEdgeDelay, p.VEdge1, p.VEdge2, fUseTwo)
	})
	return giaEdgeCoDriverDelayMax(p)
}

// Gia_ObjComputeEdgeDelay2 finds edge assignment (reverse order).
func Gia_ObjComputeEdgeDelay2(p *Gia_Man_t, iObj int32, vDelay, vEdge1, vEdge2, vFanMax1, vFanMax2, vCountMax *Vec_Int_t) int32 {
	var Status1, Status2, iFanMax1, iFanMax2 int32
	vec.Vec_IntWriteEntry(vEdge1, int(iObj), 0)
	vec.Vec_IntWriteEntry(vEdge2, int(iObj), 0)
	// analyze this node
	DelayMax := vec.Vec_IntEntry(vDelay, int(iObj))
	nCountMax := vec.Vec_IntEntry(vCountMax, int(iObj))
	if DelayMax == 0 {
	} else if nCountMax == 1 {
		iFanMax1 = vec.Vec_IntEntry(vFanMax1, int(iObj))
		Status1 = Gia_ObjEdgeCount(iFanMax1, vEdge1, vEdge2)
		if Status1 <= 1 {
			Gia_ObjEdgeAdd(iFanMax1, iObj, vEdge1, vEdge2)
			Gia_ObjEdgeAdd(iObj, iFanMax1, vEdge1, vEdge2)
			DelayMax--
		}
	} else if nCountMax == 2 {
		iFanMax1 = vec.Vec_IntEntry(vFanMax1, int(iObj))
		iFanMax2 = vec.Vec_IntEntry(vFanMax2, int(iObj))
		Status1 = Gia_ObjEdgeCount(iFanMax1, vEdge1, vEdge2)
		Status2 = Gia_ObjEdgeCount(iFanMax2, vEdge1, vEdge2)
		if Status1 <= 1 && Status2 <= 1 {
			Gia_ObjEdgeAdd(iFanMax1, iObj, vEdge1, vEdge2)
			Gia_ObjEdgeAdd(iFanMax2, iObj, vEdge1, vEdge2)
			Gia_ObjEdgeAdd(iObj, iFanMax1, vEdge1, vEdge2)
			Gia_ObjEdgeAdd(iObj, iFanMax2, vEdge1, vEdge2)
			DelayMax--
		}
	}
	vec.Vec_IntWriteEntry(vDelay, int(iObj), DelayMax)
	// computed DelayMax at this point
	for _, iFan := range giaEdgeLutFanins(p, iObj) {
		DelayFanin := vec.Vec_IntEntry(vDelay, int(iFan))
		if DelayFanin < DelayMax+1 {
			vec.Vec_IntWriteEntry(vDelay, int(iFan), DelayMax+1)
			vec.Vec_IntWriteEntry(vFanMax1, int(iFan), iObj)
			vec.Vec_IntWriteEntry(vCountMax, int(iFan), 1)
		} else if DelayFanin == DelayMax+1 {
			vec.Vec_IntWriteEntry(vFanMax2, int(iFan), iObj)
			vec.Vec_IntAddToEntry(vCountMax, int(iFan), 1)
		}
	}
	return DelayMax
}
func Gia_ManComputeEdgeDelay2(p *Gia_Man_t) int32 {
	DelayMax := int32(0)
	vFanMax1 := vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	vFanMax2 := vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	vCountMax := vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	p.vEdgeDelay = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	p.VEdge1 = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	p.VEdge2 = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	if Gia_ManHasMapping(p) != 0 {
		for iLut := Gia_ManObjNum(p) - 1; iLut > 0; iLut-- {
			if Gia_ObjIsLut(p, iLut) != 0 {
				Gia_ObjComputeEdgeDelay2(p, iLut, p.vEdgeDelay, p.VEdge1, p.VEdge2, vFanMax1, vFanMax2, vCountMax)
			}
		}
	} else if Gia_ManHasMapping2(p) != 0 {
		for iLut := Gia_ManObjNum(p) - 1; iLut > 0; iLut-- {
			if Gia_ObjIsLut2(p, iLut) != 0 {
				Gia_ObjComputeEdgeDelay2(p, iLut, p.vEdgeDelay, p.VEdge1, p.VEdge2, vFanMax1, vFanMax2, vCountMax)
			}
		}
	}
	for k := int32(0); k < Gia_ManCiNum(p); k++ {
		DelayMax = max(DelayMax, vec.Vec_IntEntry(p.vEdgeDelay, int(Gia_ObjId(p, Gia_ManCi(p, k)))))
	}
	return DelayMax
}

func Gia_ManUpdateMapping(p *Gia_Man_t, vNodes *Vec_Int_t, vWin *Vec_Wec_t) {
	for i, iNode := range vec.Vec_IntArray(vNodes) {
		a, b := vec.Vec_WecEntry(p.vMapping2, int(iNode)), vec.Vec_WecEntry(vWin, i)
		*a, *b = *b, *a
	}
}
func Gia_ManEvalWindowInc(p *Gia_Man_t, vLeaves, vNodes *Vec_Int_t, vWin *Vec_Wec_t, vTemp *Vec_Int_t, fUseTwo int32) int32 {
	DelayMax := int32(0)
	Gia_ManUpdateMapping(p, vNodes, vWin)
	Gia_ManCollectTfo(p, vLeaves, vTemp)
	vec.Vec_IntReverseOrder(vTemp)
	for _, iLut := range vec.Vec_IntArray(vTemp) {
		if Gia_ObjIsLut(p, iLut) == 0 {
			continue
		}
		Delay := Gia_ObjComputeEdgeDelay(p, iLut, p.vEdgeDelay, p.VEdge1, p.VEdge2, fUseTwo)
		DelayMax = max(DelayMax, Delay)
	}
	Gia_ManUpdateMapping(p, vNodes, vWin)
	return DelayMax
}
func Gia_ManEvalWindow(p *Gia_Man_t, vLeaves, vNodes *Vec_Int_t, vWin *Vec_Wec_t, vTemp *Vec_Int_t, fUseTwo int32) int32 {
	Gia_ManUpdateMapping(p, vNodes, vWin)
	DelayMax := Gia_ManComputeEdgeDelay(p, fUseTwo)
	Gia_ManUpdateMapping(p, vNodes, vWin)
	return DelayMax
}

func Edg_ManToMapping(p *Gia_Man_t) {
	p.vMapping2 = vec.Vec_WecStart(int(Gia_ManObjNum(p)))
	p.vFanouts2 = vec.Vec_WecStart(int(Gia_ManObjNum(p)))
	for iObj := int32(1); iObj < Gia_ManObjNum(p); iObj++ {
		if Gia_ObjIsLut(p, iObj) == 0 {
			continue
		}
		for _, iFanin := range Gia_ObjLutFanins(p, iObj) {
			vec.Vec_WecPush(p.vMapping2, int(iObj), iFanin)
			vec.Vec_WecPush(p.vFanouts2, int(iFanin), iObj)
		}
	}
}

// Edg_ObjEvalEdgeDelay computes delay for the given edge assignment.
func Edg_ObjEvalEdgeDelay(p *Gia_Man_t, iObj int32, vDelay *Vec_Int_t) int32 {
	DelayEdge := int32(0)
	DelayNoEdge := int32(1)
	DelayMax := int32(0)
	for _, iFan := range vec.Vec_IntArray(Gia_ObjLutFanins2(p, iObj)) {
		Delay := vec.Vec_IntEntry(vDelay, int(iFan))
		if Gia_ObjHaveEdge(p, iObj, iFan) != 0 {
			Delay += DelayEdge
		} else {
			Delay += DelayNoEdge
		}
		DelayMax = max(DelayMax, Delay)
	}
	return DelayMax
}
func Edg_ManEvalEdgeDelay(p *Gia_Man_t) int32 {
	DelayMax := int32(0)
	if p.vEdgeDelay == nil {
		p.vEdgeDelay = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	} else {
		vec.Vec_IntFill(p.vEdgeDelay, int(Gia_ManObjNum(p)), 0)
	}
	for iLut := int32(1); iLut < Gia_ManObjNum(p); iLut++ {
		if Gia_ObjIsLut2(p, iLut) == 0 {
			continue
		}
		Delay := Edg_ObjEvalEdgeDelay(p, iLut, p.vEdgeDelay)
		vec.Vec_IntWriteEntry(p.vEdgeDelay, int(iLut), Delay)
		DelayMax = max(DelayMax, Delay)
	}
	return DelayMax
}
func Edg_ObjEvalEdgeDelayR(p *Gia_Man_t, iObj int32, vDelay *Vec_Int_t) int32 {
	DelayEdge := int32(0)
	DelayNoEdge := int32(1)
	DelayMax := int32(0)
	for _, iFan := range vec.Vec_IntArray(vec.Vec_WecEntry(p.vFanouts2, int(iObj))) {
		Delay := vec.Vec_IntEntry(vDelay, int(iFan))
		if Gia_ObjHaveEdge(p, iObj, iFan) != 0 {
			Delay += DelayEdge
		} else {
			Delay += DelayNoEdge
		}
		DelayMax = max(DelayMax, Delay)
	}
	return DelayMax
}
func Edg_ManEvalEdgeDelayR(p *Gia_Man_t) int32 {
	DelayMax := int32(0)
	if p.vEdgeDelayR == nil {
		p.vEdgeDelayR = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	} else {
		vec.Vec_IntFill(p.vEdgeDelayR, int(Gia_ManObjNum(p)), 0)
	}
	for iLut := Gia_ManObjNum(p) - 1; iLut > 0; iLut-- {
		if Gia_ObjIsLut2(p, iLut) == 0 {
			continue
		}
		Delay := Edg_ObjEvalEdgeDelayR(p, iLut, p.vEdgeDelayR)
		vec.Vec_IntWriteEntry(p.vEdgeDelayR, int(iLut), Delay)
		DelayMax = max(DelayMax, Delay)
	}
	return DelayMax
}
func Edg_ManCollectCritEdges(p *Gia_Man_t, vEdges *Vec_Wec_t, DelayMax int32) {
	vec.Vec_WecClear(vEdges)
	vec.Vec_WecInit(vEdges, int(DelayMax+1))
	for iLut := int32(1); iLut < Gia_ManObjNum(p); iLut++ {
		if Gia_ObjIsLut2(p, iLut) == 0 {
			continue
		}
		Delay1 := vec.Vec_IntEntry(p.vEdgeDelay, int(iLut))
		Delay2 := vec.Vec_IntEntry(p.vEdgeDelayR, int(iLut))
		if Delay1+Delay2 == DelayMax {
			vec.Vec_WecPush(vEdges, int(Delay1), iLut)
		}
	}
}

// Edg_ObjImprove updates one node. C keeps the critical fanins and fanouts
// in int pFanins[4], pFanouts[4]; Edg_ManToMapping intends at most four
// fanins (the caller, Edg_ManAssignEdgeNew, is not used by any command).
func Edg_ObjImprove(p *Gia_Man_t, iObj, nEdgeLimit, DelayMax, fVerbose int32) int32 {
	nFaninsC, nFanoutsC := int32(0), int32(0)     // critical
	nFaninsEC, nFanoutsEC := int32(0), int32(0)   // edge-critical
	nFaninsENC, nFanoutsENC := int32(0), int32(0) // edge-non-critial
	var pFanins, pFanouts [4]int32
	nEdges, Count := int32(0), int32(0)
	var i int32
	// count how many fanins have critical edge
	Delay1 := vec.Vec_IntEntry(p.vEdgeDelayR, int(iObj))
	for _, iNext := range vec.Vec_IntArray(Gia_ObjLutFanins2(p, iObj)) {
		if Gia_ObjIsAnd(Gia_ManObj(p, iNext)) == 0 {
			continue
		}
		Delay2 := vec.Vec_IntEntry(p.vEdgeDelay, int(iNext))
		if Gia_ObjHaveEdge(p, iObj, iNext) != 0 {
			nEdges++
			if Delay1+Delay2 == DelayMax {
				nFaninsEC++
			} else {
				nFaninsENC++
			}
		} else {
			if Delay1+Delay2+1 == DelayMax {
				pFanins[nFaninsC] = iNext
				nFaninsC++
			}
		}
	}
	// count how many fanouts have critical edge
	Delay1 = vec.Vec_IntEntry(p.vEdgeDelay, int(iObj))
	for _, iNext := range vec.Vec_IntArray(vec.Vec_WecEntry(p.vFanouts2, int(iObj))) {
		Delay2 := vec.Vec_IntEntry(p.vEdgeDelayR, int(iNext))
		if Gia_ObjHaveEdge(p, iObj, iNext) != 0 {
			nEdges++
			if Delay1+Delay2 == DelayMax {
				nFanoutsEC++
			} else {
				nFanoutsENC++
			}
		} else {
			if Delay1+Delay2+1 == DelayMax {
				if nFanoutsC < nEdgeLimit {
					pFanouts[nFanoutsC] = iNext
				}
				nFanoutsC++
			}
		}
	}
	if fVerbose != 0 {
		fmt.Printf("%8d : ", iObj)
		fmt.Printf("Edges = %d  ", nEdges)
		fmt.Printf("Fanins (all %d  EC %d  ENC %d  C %d)  ",
			Gia_ObjLutSize2(p, iObj), nFaninsEC, nFaninsENC, nFaninsC)
		fmt.Printf("Fanouts (all %d  EC %d  ENC %d  C %d)  ",
			Gia_ObjLutFanoutNum2(p, iObj), nFanoutsEC, nFanoutsENC, nFanoutsC)
	}
	// consider simple cases
	if nEdges == nEdgeLimit {
		if fVerbose != 0 {
			fmt.Printf("Full\n")
		}
		return 0
	}
	nEdgeDiff := nEdgeLimit - nEdges
	// check if fanins or fanouts could be improved
	if nFaninsEC == 0 && nFaninsC != 0 && nFaninsC <= nEdgeDiff {
		for i = 0; i < nFaninsC; i++ {
			if Gia_ObjEdgeCount(pFanins[i], p.VEdge1, p.VEdge2) == nEdgeLimit {
				break
			}
		}
		if i == nFaninsC {
			for i = 0; i < nFaninsC; i++ {
				Count += Gia_ObjEdgeAdd(iObj, pFanins[i], p.VEdge1, p.VEdge2)
				Count += Gia_ObjEdgeAdd(pFanins[i], iObj, p.VEdge1, p.VEdge2)
			}
			if Count != 0 {
				fmt.Printf("Wrong number of edges.\n")
			}
			if fVerbose != 0 {
				fmt.Printf("Fixed %d critical fanins\n", nFaninsC)
			}
			return 1
		}
	}
	if nFanoutsEC == 0 && nFanoutsC != 0 && nFanoutsC <= nEdgeDiff {
		for i = 0; i < nFanoutsC; i++ {
			if Gia_ObjEdgeCount(pFanouts[i], p.VEdge1, p.VEdge2) == nEdgeLimit {
				break
			}
		}
		if i == nFanoutsC {
			for i = 0; i < nFanoutsC; i++ {
				Count += Gia_ObjEdgeAdd(iObj, pFanouts[i], p.VEdge1, p.VEdge2)
				Count += Gia_ObjEdgeAdd(pFanouts[i], iObj, p.VEdge1, p.VEdge2)
			}
			if Count != 0 {
				fmt.Printf("Wrong number of edges.\n")
			}
			if fVerbose != 0 {
				fmt.Printf("Fixed %d critical fanouts\n", nFanoutsC)
			}
			return 1
		}
	}
	if fVerbose != 0 {
		fmt.Printf("Cannot fix\n")
	}
	return 0
}

// Edg_ManAssignEdgeNew finds edge assignment.
func Edg_ManAssignEdgeNew(p *Gia_Man_t, nEdges, fVerbose int32) int32 {
	vEdges := vec.Vec_WecStart(0)
	var vEdge1, vEdge2 *Vec_Int_t
	DelayD, DelayPrev := int32(0), int32(1000000000) // ABC_INFINITY
	iLast := -1
	if fVerbose != 0 {
		fmt.Printf("Running edge assignment with E = %d.\n", nEdges)
	}
	// create fanouts
	Edg_ManToMapping(p)
	// create empty assignment
	p.VEdge1 = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	p.VEdge2 = vec.Vec_IntStart(int(Gia_ManObjNum(p)))
	// perform optimization
	for i := 0; i < 10000; i++ {
		// if there is no improvement after 10 iterations, quit
		if i > iLast+50 {
			break
		}
		// create delay information
		DelayD = Edg_ManEvalEdgeDelay(p)
		Edg_ManEvalEdgeDelayR(p)
		if DelayPrev > DelayD {
			vEdge1 = vec.Vec_IntDup(p.VEdge1)
			vEdge2 = vec.Vec_IntDup(p.VEdge2)
			DelayPrev = DelayD
			iLast = i
		}
		if fVerbose != 0 {
			fmt.Printf("\nIter %4d : Delay = %4d\n", i, DelayD)
		}
		// collect critical nodes (nodes with critical edges)
		Edg_ManCollectCritEdges(p, vEdges, DelayD)
		// sort levels according to the number of critical edges
		vec.Vec_WecSort(vEdges, 0)
		k := 0
		for ; k < vec.Vec_WecSize(vEdges); k++ {
			vLevel := vec.Vec_IntArray(vec.Vec_WecEntry(vEdges, k))
			j := 0
			for ; j < len(vLevel); j++ {
				if Edg_ObjImprove(p, vLevel[j], nEdges, DelayD, fVerbose) != 0 { // improved
					break
				}
			}
			if j < len(vLevel) {
				break
			}
		}
		if k == vec.Vec_WecSize(vEdges) { // if we could not improve anything, quit
			break
		}
	}
	// update to the saved version
	p.VEdge1 = vEdge1
	p.VEdge2 = vEdge2
	return DelayD
}
