package abc

// Line-by-line port of src/opt/sfm/sfmTim.c (timing manager of the mfs3
// decomposition engine, sfmDec.c). It lives here because it works on
// Abc_Ntk_t networks.

import (
	"fmt"
	"sort"

	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/scl"
	"github.com/glycerine/abcgo/misc/util"
)

const sfmAbcInfinity = 1000000000 // ABC_INFINITY

type Sfm_Tim_t struct {
	// external
	pLib      *mio.Mio_Library_t // library
	pExt      *scl.Scl_Con_t     // external timing
	pNtk      *Network           // mapped network
	Delay     int32              // the largest delay
	DeltaCrit int32              // critical delay delta
	// timing info
	vTimArrs []int32 // arrivals (rise/fall)
	vTimReqs []int32 // required (rise/fall)
	// incremental timing
	vLevels [][]int32 // levels
	// critical path
	vPath     []int32  // critical path
	vSortData []uint64 // node priority order
}

func Sfm_TimArrId(p *Sfm_Tim_t, Id int32) []int32 { return p.vTimArrs[2*Id : 2*Id+2] }
func Sfm_TimReqId(p *Sfm_Tim_t, Id int32) []int32 { return p.vTimReqs[2*Id : 2*Id+2] }

func Sfm_TimArr(p *Sfm_Tim_t, pNode *Object) []int32 { return Sfm_TimArrId(p, Abc_ObjId(pNode)) }
func Sfm_TimReq(p *Sfm_Tim_t, pNode *Object) []int32 { return Sfm_TimReqId(p, Abc_ObjId(pNode)) }

func Sfm_TimArrMaxId(p *Sfm_Tim_t, Id int32) int32 {
	a := Sfm_TimArrId(p, Id)
	return max(a[0], a[1])
}
func Sfm_TimArrMax(p *Sfm_Tim_t, pNode *Object) int32 {
	a := Sfm_TimArr(p, pNode)
	return max(a[0], a[1])
}
func Sfm_TimSetReq(p *Sfm_Tim_t, pNode *Object, t int32) {
	r := Sfm_TimReq(p, pNode)
	r[0], r[1] = t, t
}
func Sfm_TimSlack(p *Sfm_Tim_t, pNode *Object) int32 {
	r, a := Sfm_TimReq(p, pNode), Sfm_TimArr(p, pNode)
	return min(r[0]-a[0], r[1]-a[1])
}

func Sfm_TimEdgeArrival(p *Sfm_Tim_t, pPin *mio.Mio_Pin_t, pTimeIn []int32, pTimeOut []int32) {
	PinPhase := mio.Mio_PinReadPhase(pPin)
	tDelayBlockRise := scl.Scl_Flt2Int(float32(mio.Mio_PinReadDelayBlockRise(pPin)))
	tDelayBlockFall := scl.Scl_Flt2Int(float32(mio.Mio_PinReadDelayBlockFall(pPin)))
	if PinPhase != mio.MIO_PHASE_INV { // NONINV phase is present
		pTimeOut[0] = max(pTimeOut[0], pTimeIn[0]+tDelayBlockRise)
		pTimeOut[1] = max(pTimeOut[1], pTimeIn[1]+tDelayBlockFall)
	}
	if PinPhase != mio.MIO_PHASE_NONINV { // INV phase is present
		pTimeOut[0] = max(pTimeOut[0], pTimeIn[1]+tDelayBlockRise)
		pTimeOut[1] = max(pTimeOut[1], pTimeIn[0]+tDelayBlockFall)
	}
}
func Sfm_TimGateArrival(p *Sfm_Tim_t, pGate *mio.Mio_Gate_t, pTimesIn [][]int32, pTimeOut []int32) {
	i := 0
	pTimeOut[0], pTimeOut[1] = 0, 0
	for pPin := mio.Mio_GateReadPins(pGate); pPin != nil; pPin = mio.Mio_PinReadNext(pPin) {
		Sfm_TimEdgeArrival(p, pPin, pTimesIn[i], pTimeOut)
		i++
	}
	assert(int32(i) == mio.Mio_GateReadPinNum(pGate), "i == Mio_GateReadPinNum(pGate)")
}
func Sfm_TimNodeArrival(p *Sfm_Tim_t, pNode *Object) {
	var pTimesIn [6][]int32
	pTimeOut := Sfm_TimArr(p, pNode)
	assert(Abc_ObjFaninNum(pNode) <= 6, "Abc_ObjFaninNum(pNode) <= 6")
	for i, e := range pNode.fanins {
		pTimesIn[i] = Sfm_TimArrId(p, int32(e.Object.ID))
	}
	Sfm_TimGateArrival(p, pNode.Data.(*mio.Mio_Gate_t), pTimesIn[:], pTimeOut)
}

func Sfm_TimEdgeRequired(p *Sfm_Tim_t, pPin *mio.Mio_Pin_t, pTimeIn []int32, pTimeOut []int32) {
	PinPhase := mio.Mio_PinReadPhase(pPin)
	tDelayBlockRise := scl.Scl_Flt2Int(float32(mio.Mio_PinReadDelayBlockRise(pPin)))
	tDelayBlockFall := scl.Scl_Flt2Int(float32(mio.Mio_PinReadDelayBlockFall(pPin)))
	if PinPhase != mio.MIO_PHASE_INV { // NONINV phase is present
		pTimeIn[0] = min(pTimeIn[0], pTimeOut[0]-tDelayBlockRise)
		pTimeIn[1] = min(pTimeIn[1], pTimeOut[1]-tDelayBlockFall)
	}
	if PinPhase != mio.MIO_PHASE_NONINV { // INV phase is present
		pTimeIn[0] = min(pTimeIn[0], pTimeOut[1]-tDelayBlockRise)
		pTimeIn[1] = min(pTimeIn[1], pTimeOut[0]-tDelayBlockFall)
	}
}
func Sfm_TimGateRequired(p *Sfm_Tim_t, pGate *mio.Mio_Gate_t, pTimesIn [][]int32, pTimeOut []int32) {
	i := 0
	for pPin := mio.Mio_GateReadPins(pGate); pPin != nil; pPin = mio.Mio_PinReadNext(pPin) {
		Sfm_TimEdgeRequired(p, pPin, pTimesIn[i], pTimeOut)
		i++
	}
	assert(int32(i) == mio.Mio_GateReadPinNum(pGate), "i == Mio_GateReadPinNum(pGate)")
}
func Sfm_TimNodeRequired(p *Sfm_Tim_t, pNode *Object) {
	var pTimesIn [6][]int32
	pTimeOut := Sfm_TimReq(p, pNode)
	assert(Abc_ObjFaninNum(pNode) <= 6, "Abc_ObjFaninNum(pNode) <= 6")
	for i, e := range pNode.fanins {
		pTimesIn[i] = Sfm_TimReqId(p, int32(e.Object.ID))
	}
	Sfm_TimGateRequired(p, pNode.Data.(*mio.Mio_Gate_t), pTimesIn[:], pTimeOut)
}

func Sfm_TimCriticalPath_int(p *Sfm_Tim_t, pObj *Object, vPath *[]int32, SlackMax int32) {
	if Abc_NodeIsTravIdCurrent(pObj) != 0 {
		return
	}
	Abc_NodeSetTravIdCurrent(pObj)
	assert(Abc_ObjIsNode(pObj) != 0, "Abc_ObjIsNode(pObj)")
	for i := int32(0); i < Abc_ObjFaninNum(pObj); i++ {
		pNext := Abc_ObjFanin(pObj, i)
		if Abc_ObjIsCi(pNext) != 0 || Abc_ObjFaninNum(pNext) == 0 {
			continue
		}
		assert(Abc_ObjIsNode(pNext) != 0, "Abc_ObjIsNode(pNext)")
		if Sfm_TimSlack(p, pNext) <= SlackMax {
			Sfm_TimCriticalPath_int(p, pNext, vPath, SlackMax)
		}
	}
	if Abc_ObjFaninNum(pObj) > 0 {
		*vPath = append(*vPath, Abc_ObjId(pObj))
	}
}
func Sfm_TimCriticalPath(p *Sfm_Tim_t, Window int32) int32 {
	SlackMax := p.Delay * Window / 100
	p.vPath = p.vPath[:0]
	Abc_NtkIncrementTravId(p.pNtk)
	for i := int32(0); i < Abc_NtkCoNum(p.pNtk); i++ {
		pObj := Abc_NtkCo(p.pNtk, i)
		pNext := Abc_ObjFanin0(pObj)
		if Abc_ObjIsCi(pNext) != 0 || Abc_ObjFaninNum(pNext) == 0 {
			continue
		}
		assert(Abc_ObjIsNode(pNext) != 0, "Abc_ObjIsNode(pNext)")
		if Sfm_TimSlack(p, pNext) <= SlackMax {
			Sfm_TimCriticalPath_int(p, pNext, &p.vPath, SlackMax)
		}
	}
	return int32(len(p.vPath))
}

func Sfm_TimTrace(p *Sfm_Tim_t) int32 {
	Delay := int32(0)
	vNodes := Abc_NtkDfs(p.pNtk, 1)
	for _, pObj := range vNodes {
		Sfm_TimNodeArrival(p, pObj)
	}
	for i := int32(0); i < Abc_NtkCoNum(p.pNtk); i++ {
		pObj := Abc_NtkCo(p.pNtk, i)
		Delay = max(Delay, Sfm_TimArrMax(p, Abc_ObjFanin0(pObj)))
	}
	p.vTimReqs = sfmIntFill(p.vTimReqs, 2*int(Abc_NtkObjNumMax(p.pNtk)), sfmAbcInfinity)
	for i := int32(0); i < Abc_NtkCoNum(p.pNtk); i++ {
		pObj := Abc_NtkCo(p.pNtk, i)
		Sfm_TimSetReq(p, Abc_ObjFanin0(pObj), Delay)
	}
	for i := len(vNodes) - 1; i >= 0; i-- {
		Sfm_TimNodeRequired(p, vNodes[i])
	}
	return Delay
}

// sfmIntFill is Vec_IntFill: the vector gets nSize entries equal to Fill.
func sfmIntFill(v []int32, nSize int, Fill int32) []int32 {
	if cap(v) < nSize {
		v = make([]int32, nSize)
	}
	v = v[:nSize]
	for i := range v {
		v[i] = Fill
	}
	return v
}

// sfmIntFillExtra is Vec_IntFillExtra: grows the vector to nSize entries,
// filling the new ones.
func sfmIntFillExtra(v []int32, nSize int, Fill int32) []int32 {
	for len(v) < nSize {
		v = append(v, Fill)
	}
	return v
}

func Sfm_TimStart(pLib *mio.Mio_Library_t, pExt *scl.Scl_Con_t, pNtk *Network, DeltaCrit int32) *Sfm_Tim_t {
	p := new(Sfm_Tim_t)
	p.pLib = pLib
	p.pExt = pExt
	p.pNtk = pNtk
	p.vTimArrs = sfmIntFill(nil, 3*int(Abc_NtkObjNumMax(pNtk)), 0)
	p.vTimReqs = sfmIntFill(nil, 3*int(Abc_NtkObjNumMax(pNtk)), 0)
	p.Delay = Sfm_TimTrace(p)
	assert(DeltaCrit > 0 && DeltaCrit < scl.Scl_Flt2Int(1000.0), "DeltaCrit > 0 && DeltaCrit < Scl_Flt2Int(1000.0)")
	p.DeltaCrit = DeltaCrit
	return p
}
func Sfm_TimStop(p *Sfm_Tim_t) {
	*p = Sfm_Tim_t{}
}
func Sfm_TimReadNtkDelay(p *Sfm_Tim_t) int32 {
	return p.Delay
}
func Sfm_TimReadObjDelay(p *Sfm_Tim_t, iObj int32) int32 {
	return Sfm_TimArrMaxId(p, iObj)
}

func Sfm_TimTest(pNtk *Network) {
	pLib := pNtk.FunctionManager.(*mio.Mio_Library_t)
	p := Sfm_TimStart(pLib, nil, pNtk, 100)
	fmt.Printf("Max delay = %.2f.  Path = %d (%d).\n", scl.Scl_Int2Flt(p.Delay), Sfm_TimCriticalPath(p, 1), Abc_NtkNodeNum(p.pNtk))
	Sfm_TimStop(p)
}

func Sfm_TimUpdateClean(p *Sfm_Tim_t) {
	for i := range p.vLevels {
		for _, id := range p.vLevels[i] {
			pObj := Abc_NtkObj(p.pNtk, id)
			if pObj == nil { // Abc_NtkForEachObjVec skips deleted objects
				continue
			}
			assert(pObj.MarkC, "pObj->fMarkC == 1")
			pObj.MarkC = false
		}
		p.vLevels[i] = p.vLevels[i][:0]
	}
}

func Sfm_TimUpdateTiming(p *Sfm_Tim_t, vTimeNodes []int32) {
	assert(len(vTimeNodes) > 0 && len(vTimeNodes) <= 2, "Vec_IntSize(vTimeNodes) > 0 && Vec_IntSize(vTimeNodes) <= 2")
	p.vTimArrs = sfmIntFillExtra(p.vTimArrs, 2*int(Abc_NtkObjNumMax(p.pNtk)), 0)
	p.vTimReqs = sfmIntFillExtra(p.vTimReqs, 2*int(Abc_NtkObjNumMax(p.pNtk)), 0)
	p.Delay = Sfm_TimTrace(p)
}

// Sfm_TimSortArrayByArrival sorts an array of nodes using their max arrival
// times and returns the number of new divisor nodes.
func Sfm_TimSortArrayByArrival(p *Sfm_Tim_t, vNodes *[]int32, iPivot int32) int32 {
	nDivNew := int32(-1)
	MaxDelay := int32(sfmAbcInfinity/2) + Sfm_TimArrMaxId(p, iPivot)
	assert(p.DeltaCrit > 0, "p->DeltaCrit > 0")
	// collect nodes
	p.vSortData = p.vSortData[:0]
	for _, Id := range *vNodes {
		Time := Sfm_TimArrMaxId(p, Id)
		assert(-sfmAbcInfinity/2 < Time && Time < sfmAbcInfinity/2, "time range")
		p.vSortData = append(p.vSortData, uint64(Id)<<32|uint64(uint32(sfmAbcInfinity/2+Time)))
	}
	// sort nodes by delay
	util.Abc_QuickSort3(p.vSortData, int32(len(p.vSortData)), 0)
	// collect sorted nodes and find place where divisors end
	*vNodes = (*vNodes)[:0]
	for i, Entry := range p.vSortData {
		*vNodes = append(*vNodes, int32(Entry>>32))
		if nDivNew == -1 && int32(Entry)+p.DeltaCrit > MaxDelay {
			nDivNew = int32(i)
		}
	}
	return nDivNew
}

// Sfm_TimPriorityNodes collects the priority of nodes to try remapping for delay.
func Sfm_TimPriorityNodes(p *Sfm_Tim_t, vCands *[]int32, Window int32) bool {
	assert(Window >= 0 && Window <= 100, "Window >= 0 && Window <= 100")
	// collect critical path
	Sfm_TimCriticalPath(p, Window)
	// add nodes to the levelized structure
	Sfm_TimUpdateClean(p)
	for _, id := range p.vPath {
		pObj := Abc_NtkObj(p.pNtk, id)
		assert(!pObj.MarkC, "pObj->fMarkC == 0")
		pObj.MarkC = true
		Level := int(Abc_ObjLevel(pObj))
		for len(p.vLevels) < Level+1 { // Vec_WecPush
			p.vLevels = append(p.vLevels, nil)
		}
		p.vLevels[Level] = append(p.vLevels[Level], Abc_ObjId(pObj))
	}
	// prioritize nodes by expected gain (Vec_WecSort: glibc's qsort is a
	// stable merge sort here)
	sort.SliceStable(p.vLevels, func(i, j int) bool { return len(p.vLevels[i]) < len(p.vLevels[j]) })
	*vCands = (*vCands)[:0]
	for _, vLevel := range p.vLevels {
		for _, id := range vLevel {
			pObj := Abc_NtkObj(p.pNtk, id)
			if pObj != nil && !pObj.MarkA {
				*vCands = append(*vCands, Abc_ObjId(pObj))
			}
		}
	}
	return len(*vCands) > 0
}

// Sfm_TimNodeIsNonCritical returns 1 if node is relatively non-critical
// compared to the pivot.
func Sfm_TimNodeIsNonCritical(p *Sfm_Tim_t, pPivot, pNode *Object) bool {
	return Sfm_TimArrMax(p, pNode)+p.DeltaCrit <= Sfm_TimArrMax(p, pPivot)
}

func Sfm_TimEvalRemapping(p *Sfm_Tim_t, vFanins []int32, vMap []int32, pGate1 *mio.Mio_Gate_t, pFans1 []byte, pGate2 *mio.Mio_Gate_t, pFans2 []byte) int32 {
	var TimeOut [2][2]int32
	var pTimesIn1, pTimesIn2 [6][]int32
	// process the first gate
	nFanins1 := mio.Mio_GateReadPinNum(pGate1)
	for i := int32(0); i < nFanins1; i++ {
		pTimesIn1[i] = Sfm_TimArrId(p, vMap[vFanins[pFans1[i]]])
	}
	Sfm_TimGateArrival(p, pGate1, pTimesIn1[:], TimeOut[0][:])
	if pGate2 == nil {
		return max(TimeOut[0][0], TimeOut[0][1])
	}
	// process the second gate
	nFanins2 := mio.Mio_GateReadPinNum(pGate2)
	for i := int32(0); i < nFanins2; i++ {
		if pFans2[i] == 16 {
			pTimesIn2[i] = TimeOut[0][:]
		} else {
			pTimesIn2[i] = Sfm_TimArrId(p, vMap[vFanins[pFans2[i]]])
		}
	}
	Sfm_TimGateArrival(p, pGate2, pTimesIn2[:], TimeOut[1][:])
	return max(TimeOut[1][0], TimeOut[1][1])
}
