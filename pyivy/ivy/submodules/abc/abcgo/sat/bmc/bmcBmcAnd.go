package bmc

// Line-by-line port of src/sat/bmc/bmcBmcAnd.c: the memory-efficient BMC
// engine of &bmc (Gia_ManBmcPerform). Ternary simulation finds the frames
// and cones that matter, Gia_ManBmcUnroll builds them into one GIA, the Mf
// mapper derives its CNF, and one incremental sat_solver loads each frame's
// cones on demand and checks its outputs.
//
// Not ported: Gia_ManBmcDupCone, Gia_ManBmcAssignVarIds, Gia_ManBmcAddCnf,
// Gia_ManBmcAddCone, Gia_ManBmcPerform_Unr and Gia_ManBmcPerform_old_cnf,
// which nothing in ABC calls.

import (
	"fmt"
	"slices"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// bmcMnaCnf holds the fields of Cnf_Dat_t that Gia_ManBmcAddCnfNew reads:
// the clauses and each object's clause range. Mf (gia.Cnf_Dat_t) and
// Cnf_DeriveOther (cnf.Cnf_Dat_t) produce different Go types.
type bmcMnaCnf struct {
	Clauses               [][]int32
	Obj2Clause, Obj2Count []int32
}

type Bmc_Mna_t struct {
	pFrames  *gia.Gia_Man_t // time frames
	pCnf     *bmcMnaCnf     // CNF derived for the timeframes
	vPiMap   *vec.Vec_Int_t // maps unrolled GIA PIs into user GIA PIs
	vId2Var  *vec.Vec_Int_t // maps GIA IDs into SAT vars
	vInputs  *vec.Vec_Int_t // inputs of the cone
	vOutputs *vec.Vec_Int_t // outputs of the cone
	vNodes   *vec.Vec_Int_t // internal nodes of the cone
	pSat     *bsat.Solver   // SAT solver
	nSatVars int32          // the counter of SAT variables
	clkStart int64          // starting time
}

func bmcMnaTerSimInfoGet(pInfo []uint32, i int32) int32 {
	return int32(3 & (pInfo[i>>4] >> uint((i&15)<<1)))
}
func bmcMnaTerSimInfoSet(pInfo []uint32, i int32, Value int32) {
	if !(Value >= gia.GIA_ZER && Value <= gia.GIA_UND) {
		panic("Gia_ManTerSimInfoSet: Value >= GIA_ZER && Value <= GIA_UND")
	}
	Value ^= bmcMnaTerSimInfoGet(pInfo, i)
	pInfo[i>>4] ^= uint32(Value) << uint((i&15)<<1)
}

func bmcMnaStateSeen(vStates [][]uint32, pState []uint32) bool {
	for _, pPrev := range vStates {
		if slices.Equal(pPrev, pState) {
			return true
		}
	}
	return false
}

func bmcBool(b bool) int32 {
	if b {
		return 1
	}
	return 0
}

func bmcObjValue(pObj *gia.Gia_Obj_t) int32 { return int32(pObj.Value) }

// Bmc_MnaTernary performs ternary simulation of the manager.
func Bmc_MnaTernary(p *gia.Gia_Man_t, nFrames, nFramesAdd, fVerbose int32, iFirst *int32) [][]uint32 {
	var vStates [][]uint32
	var pState []uint32
	nStateWords := util.Abc_BitWordNum(2 * gia.Gia_ManCoNum(p))
	var Count [4]int32
	clk := util.Abc_Clock()
	gia.Gia_ManConst0(p).Value = gia.GIA_ZER
	for i := int32(0); i < gia.Gia_ManPiNum(p); i++ {
		gia.Gia_ManPi(p, i).Value = gia.GIA_UND
	}
	for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
		gia.Gia_ManRi(p, i).Value = gia.GIA_ZER
	}
	*iFirst = -1
	vStates = make([][]uint32, 0, 100)
	for f := int32(0); ; f++ {
		// if frames are given, break at frames
		if nFrames != 0 && f == nFrames {
			break
		}
		// if frames are not given, break after nFramesAdd from the first x-valued
		// (">=": with nFramesAdd == 0, f is already past *iFirst; CBUGS_FIXED.md #41)
		if nFrames == 0 && *iFirst >= 0 && f >= *iFirst+nFramesAdd {
			break
		}
		// aassign CI values
		for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
			gia.Gia_ManRo(p, i).Value = gia.Gia_ManRi(p, i).Value
		}
		for i := int32(0); i < gia.Gia_ManObjNum(p); i++ {
			pObj := gia.Gia_ManObj(p, i)
			if gia.Gia_ObjIsAnd(pObj) == 0 {
				continue
			}
			pObj.Value = uint32(gia.Gia_XsimAndCond(bmcObjValue(gia.Gia_ObjFanin0(pObj)), gia.Gia_ObjFaninC0(pObj), bmcObjValue(gia.Gia_ObjFanin1(pObj)), gia.Gia_ObjFaninC1(pObj)))
		}
		// compute and save CO values
		pState = make([]uint32, nStateWords)
		for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
			pObj := gia.Gia_ManCo(p, i)
			pObj.Value = uint32(gia.Gia_XsimNotCond(bmcObjValue(gia.Gia_ObjFanin0(pObj)), gia.Gia_ObjFaninC0(pObj)))
			bmcMnaTerSimInfoSet(pState, i, bmcObjValue(pObj))
			if *iFirst == -1 && i < gia.Gia_ManPoNum(p) && pObj.Value == gia.GIA_UND {
				*iFirst = f
			}
		}
		// if frames are not given and no PO is x-valued yet, a repeated state
		// means that no PO ever will be: stop (CBUGS_FIXED.md #41)
		if nFrames == 0 && *iFirst == -1 && bmcMnaStateSeen(vStates, pState) {
			break
		}
		vStates = append(vStates, pState)
		// print statistics
		if fVerbose == 0 {
			continue
		}
		Count[0], Count[1], Count[2], Count[3] = 0, 0, 0, 0
		for i := int32(0); i < gia.Gia_ManRegNum(p); i++ {
			Count[gia.Gia_ManRi(p, i).Value]++
		}
		out := "0"
		if gia.Gia_ManPo(p, 0).Value == gia.GIA_UND {
			out = "x"
		}
		fmt.Printf("%5d : 0 =%7d    1 =%7d    x =%7d    all =%7d   out = %s\n",
			f, Count[gia.GIA_ZER], Count[gia.GIA_ONE], Count[gia.GIA_UND], gia.Gia_ManRegNum(p), out)
	}
	if fVerbose != 0 {
		fmt.Printf("Finished %d frames. First x-valued PO is in frame %d.  ", nFrames, *iFirst)
	}
	if fVerbose != 0 {
		util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	}
	return vStates
}

// Bmc_MnaCollect_rec collects AIG nodes for the group of POs.
func Bmc_MnaCollect_rec(p *gia.Gia_Man_t, pObj *gia.Gia_Obj_t, vNodes *vec.Vec_Int_t, pState []uint32) {
	if gia.Gia_ObjPhase(pObj) != 0 {
		return
	}
	gia.Gia_ObjSetPhaseValue(pObj, 1)
	if gia.Gia_ObjIsAnd(pObj) != 0 {
		Bmc_MnaCollect_rec(p, gia.Gia_ObjFanin0(pObj), vNodes, pState)
		Bmc_MnaCollect_rec(p, gia.Gia_ObjFanin1(pObj), vNodes, pState)
		pObj.Value = uint32(gia.Gia_XsimAndCond(bmcObjValue(gia.Gia_ObjFanin0(pObj)), gia.Gia_ObjFaninC0(pObj), bmcObjValue(gia.Gia_ObjFanin1(pObj)), gia.Gia_ObjFaninC1(pObj)))
	} else if gia.Gia_ObjIsRo(p, pObj) != 0 {
		if pState != nil {
			pObj.Value = uint32(bmcMnaTerSimInfoGet(pState, gia.Gia_ObjCioId(gia.Gia_ObjRoToRi(p, pObj))))
		} else {
			pObj.Value = gia.GIA_ZER
		}
	} else if gia.Gia_ObjIsPi(p, pObj) != 0 {
		pObj.Value = gia.GIA_UND
	} else {
		panic("Bmc_MnaCollect_rec: unexpected object")
	}
	vec.Vec_IntPush(vNodes, gia.Gia_ObjId(p, pObj))
}
func Bmc_MnaCollect(p *gia.Gia_Man_t, vCos, vNodes *vec.Vec_Int_t, pState []uint32) {
	vec.Vec_IntClear(vNodes)
	gia.Gia_ObjSetPhaseValue(gia.Gia_ManConst0(p), 1)
	gia.Gia_ManConst0(p).Value = gia.GIA_ZER
	for i := 0; i < vec.Vec_IntSize(vCos); i++ {
		pObj := gia.Gia_ManObj(p, vec.Vec_IntEntry(vCos, i))
		if gia.Gia_ObjIsCo(pObj) == 0 {
			panic("Bmc_MnaCollect: Gia_ObjIsCo(pObj)")
		}
		Bmc_MnaCollect_rec(p, gia.Gia_ObjFanin0(pObj), vNodes, pState)
		pObj.Value = uint32(gia.Gia_XsimNotCond(bmcObjValue(gia.Gia_ObjFanin0(pObj)), gia.Gia_ObjFaninC0(pObj)))
		if pObj.Value != gia.GIA_UND {
			panic("Bmc_MnaCollect: pObj->Value == GIA_UND")
		}
	}
}

// Bmc_MnaSelect_rec selects related logic cones for the COs.
func Bmc_MnaSelect_rec(p *gia.Gia_Man_t, pObj *gia.Gia_Obj_t, vLeaves *vec.Vec_Int_t) {
	if gia.Gia_ObjPhase(pObj) == 0 {
		return
	}
	gia.Gia_ObjSetPhaseValue(pObj, 0)
	if pObj.Value != gia.GIA_UND {
		panic("Bmc_MnaSelect_rec: pObj->Value == GIA_UND")
	}
	if gia.Gia_ObjIsAnd(pObj) != 0 {
		if gia.Gia_ObjFanin0(pObj).Value == gia.GIA_UND {
			Bmc_MnaSelect_rec(p, gia.Gia_ObjFanin0(pObj), vLeaves)
		} else if bmcObjValue(gia.Gia_ObjFanin0(pObj))+gia.Gia_ObjFaninC0(pObj) != gia.GIA_ONE {
			panic("Bmc_MnaSelect_rec: fanin0 is not 1")
		}
		if gia.Gia_ObjFanin1(pObj).Value == gia.GIA_UND {
			Bmc_MnaSelect_rec(p, gia.Gia_ObjFanin1(pObj), vLeaves)
		} else if bmcObjValue(gia.Gia_ObjFanin1(pObj))+gia.Gia_ObjFaninC1(pObj) != gia.GIA_ONE {
			panic("Bmc_MnaSelect_rec: fanin1 is not 1")
		}
	} else if gia.Gia_ObjIsRo(p, pObj) != 0 {
		vec.Vec_IntPush(vLeaves, gia.Gia_ObjId(p, gia.Gia_ObjRoToRi(p, pObj)))
	}
}
func Bmc_MnaSelect(p *gia.Gia_Man_t, vCos, vNodes, vLeaves *vec.Vec_Int_t) {
	vec.Vec_IntClear(vLeaves)
	for i := 0; i < vec.Vec_IntSize(vCos); i++ {
		pObj := gia.Gia_ManObj(p, vec.Vec_IntEntry(vCos, i))
		Bmc_MnaSelect_rec(p, gia.Gia_ObjFanin0(pObj), vLeaves)
	}
	gia.Gia_ObjSetPhaseValue(gia.Gia_ManConst0(p), 0)
	for i := 0; i < vec.Vec_IntSize(vNodes); i++ {
		gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(p, vec.Vec_IntEntry(vNodes, i)), 0)
	}
}

// Bmc_MnaBuild_rec builds the AIG for the selected cones.
func Bmc_MnaBuild_rec(p *gia.Gia_Man_t, pObj *gia.Gia_Obj_t, pNew *gia.Gia_Man_t, vMap, vPiMap *vec.Vec_Int_t) {
	if gia.Gia_ObjPhase(pObj) == 0 {
		return
	}
	gia.Gia_ObjSetPhaseValue(pObj, 0)
	if pObj.Value != gia.GIA_UND {
		panic("Bmc_MnaBuild_rec: pObj->Value == GIA_UND")
	}
	if gia.Gia_ObjIsAnd(pObj) != 0 {
		iLit0, iLit1 := int32(1), int32(1)
		if gia.Gia_ObjFanin0(pObj).Value == gia.GIA_UND {
			Bmc_MnaBuild_rec(p, gia.Gia_ObjFanin0(pObj), pNew, vMap, vPiMap)
		}
		if gia.Gia_ObjFanin1(pObj).Value == gia.GIA_UND {
			Bmc_MnaBuild_rec(p, gia.Gia_ObjFanin1(pObj), pNew, vMap, vPiMap)
		}
		if gia.Gia_ObjFanin0(pObj).Value == gia.GIA_UND {
			iLit0 = util.Abc_LitNotCond(vec.Vec_IntEntry(vMap, int(gia.Gia_ObjFaninId0p(p, pObj))), gia.Gia_ObjFaninC0(pObj))
		}
		if gia.Gia_ObjFanin1(pObj).Value == gia.GIA_UND {
			iLit1 = util.Abc_LitNotCond(vec.Vec_IntEntry(vMap, int(gia.Gia_ObjFaninId1p(p, pObj))), gia.Gia_ObjFaninC1(pObj))
		}
		vec.Vec_IntWriteEntry(vMap, int(gia.Gia_ObjId(p, pObj)), gia.Gia_ManHashAnd(pNew, iLit0, iLit1))
	} else if gia.Gia_ObjIsRo(p, pObj) != 0 {
		if vec.Vec_IntEntry(vMap, int(gia.Gia_ObjId(p, pObj))) == -1 {
			panic("Bmc_MnaBuild_rec: unmapped register output")
		}
	} else if gia.Gia_ObjIsPi(p, pObj) != 0 {
		vec.Vec_IntPush(vPiMap, gia.Gia_ObjCioId(pObj))
		vec.Vec_IntWriteEntry(vMap, int(gia.Gia_ObjId(p, pObj)), gia.Gia_ManAppendCi(pNew))
	} else {
		panic("Bmc_MnaBuild_rec: unexpected object")
	}
}
func Bmc_MnaBuild(p *gia.Gia_Man_t, vCos, vNodes *vec.Vec_Int_t, pNew *gia.Gia_Man_t, vMap, vPiMap *vec.Vec_Int_t) {
	for i := 0; i < vec.Vec_IntSize(vCos); i++ {
		pObj := gia.Gia_ManObj(p, vec.Vec_IntEntry(vCos, i))
		if gia.Gia_ObjIsCo(pObj) == 0 {
			panic("Bmc_MnaBuild: Gia_ObjIsCo(pObj)")
		}
		Bmc_MnaBuild_rec(p, gia.Gia_ObjFanin0(pObj), pNew, vMap, vPiMap)
		iLit := util.Abc_LitNotCond(vec.Vec_IntEntry(vMap, int(gia.Gia_ObjFaninId0p(p, pObj))), gia.Gia_ObjFaninC0(pObj))
		vec.Vec_IntWriteEntry(vMap, int(gia.Gia_ObjId(p, pObj)), iLit)
	}
	gia.Gia_ObjSetPhaseValue(gia.Gia_ManConst0(p), 0)
	for i := 0; i < vec.Vec_IntSize(vNodes); i++ {
		gia.Gia_ObjSetPhaseValue(gia.Gia_ManObj(p, vec.Vec_IntEntry(vNodes, i)), 0)
	}
}

// Gia_ManBmcUnroll computes the first non-trivial timeframe and unrolls the
// cones that matter.
func Gia_ManBmcUnroll(pGia *gia.Gia_Man_t, nFramesMax, nFramesAdd, fVerbose int32, pvPiMap **vec.Vec_Int_t) *gia.Gia_Man_t {
	var vRoots *vec.Vec_Int_t
	var pStateF, pStateP []uint32
	var f, iFirst int32
	gia.Gia_ManCleanPhase(pGia)
	vCone := vec.Vec_IntAlloc(1000)
	vLeaves := vec.Vec_IntAlloc(1000)
	// perform ternary simulation
	vStates := Bmc_MnaTernary(pGia, nFramesMax, nFramesAdd, fVerbose, &iFirst)
	// go backward
	*pvPiMap = vec.Vec_IntAlloc(1000)
	vBegins := make([]*vec.Vec_Int_t, len(vStates))
	for f = int32(len(vStates)) - 1; f >= 0; f-- {
		// get ternary states
		pStateF = vStates[f]
		pStateP = nil
		if f != 0 {
			pStateP = vStates[f-1]
		}
		// collect roots of this frame
		vRoots = vec.Vec_IntAlloc(100)
		for i := int32(0); i < gia.Gia_ManPoNum(pGia); i++ {
			pObj := gia.Gia_ManPo(pGia, i)
			if bmcMnaTerSimInfoGet(pStateF, gia.Gia_ObjCioId(pObj)) == gia.GIA_UND {
				vec.Vec_IntPush(vRoots, gia.Gia_ObjId(pGia, pObj))
			}
		}
		// add leaves from the previous frame
		vec.Vec_IntAppend(vRoots, vLeaves)
		vBegins[f] = vRoots
		// find the cone
		Bmc_MnaCollect(pGia, vRoots, vCone, pStateP) // computes vCone
		Bmc_MnaSelect(pGia, vRoots, vCone, vLeaves)  // computes vLeaves
		if fVerbose != 0 {
			fmt.Printf("Frame %4d :  Roots = %6d  Leaves = %6d  Cone = %6d\n",
				f, vec.Vec_IntSize(vRoots), vec.Vec_IntSize(vLeaves), vec.Vec_IntSize(vCone))
		}
		// earlier frames may still have x-valued POs: do not stop when
		// vLeaves is empty (CBUGS_FIXED.md #39)
	}
	// go forward
	vMap := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(pGia)))
	pNew := gia.Gia_ManStart(10000)
	pNew.SetIdentity(gia.Gia_ManName(pGia), "")
	gia.Gia_ManHashStart(pNew)
	for f = 0; f < int32(len(vStates)); f++ {
		vRoots = vBegins[f]
		if vRoots == nil {
			for i := int32(0); i < gia.Gia_ManPoNum(pGia); i++ {
				gia.Gia_ManAppendCo(pNew, 0)
			}
			continue
		}
		// get ternary states
		pStateF = vStates[f]
		pStateP = nil
		if f != 0 {
			pStateP = vStates[f-1]
		}
		// clean POs (a PO that ternary simulation proves 1 is constant 1;
		// CBUGS_FIXED.md #40)
		for i := int32(0); i < gia.Gia_ManPoNum(pGia); i++ {
			pObj := gia.Gia_ManPo(pGia, i)
			vec.Vec_IntWriteEntry(vMap, int(gia.Gia_ObjId(pGia, pObj)), bmcBool(bmcMnaTerSimInfoGet(pStateF, gia.Gia_ObjCioId(pObj)) == gia.GIA_ONE))
		}
		// find the cone
		vec.Vec_IntPush(*pvPiMap, -f-1)
		Bmc_MnaCollect(pGia, vRoots, vCone, pStateP)            // computes vCone
		Bmc_MnaBuild(pGia, vRoots, vCone, pNew, vMap, *pvPiMap) // computes pNew
		if fVerbose != 0 {
			fmt.Printf("Frame %4d :  Roots = %6d  Leaves = %6d  Cone = %6d\n",
				f, vec.Vec_IntSize(vRoots), vec.Vec_IntSize(vLeaves), vec.Vec_IntSize(vCone))
		}
		// create POs
		for i := int32(0); i < gia.Gia_ManPoNum(pGia); i++ {
			gia.Gia_ManAppendCo(pNew, vec.Vec_IntEntry(vMap, int(gia.Gia_ObjId(pGia, gia.Gia_ManPo(pGia, i)))))
		}
		// set a new map
		for i := 0; i < vec.Vec_IntSize(vRoots); i++ {
			pObj := gia.Gia_ManObj(pGia, vec.Vec_IntEntry(vRoots, i))
			if gia.Gia_ObjIsRi(pGia, pObj) != 0 {
				vec.Vec_IntWriteEntry(vMap, int(gia.Gia_ObjId(pGia, gia.Gia_ObjRiToRo(pGia, pObj))), vec.Vec_IntEntry(vMap, int(gia.Gia_ObjId(pGia, pObj))))
			}
		}
	}
	gia.Gia_ManHashStop(pNew)
	// cleanup
	pTemp := pNew
	pNew = gia.Gia_ManCleanup(pTemp)
	gia.Gia_ManStop(pTemp)
	return pNew
}

// Bmc_MnaAlloc and Bmc_MnaFree manipulate the BMC manager.
func Bmc_MnaAlloc() *Bmc_Mna_t {
	p := &Bmc_Mna_t{}
	p.vId2Var = vec.Vec_IntAlloc(0)
	p.vInputs = vec.Vec_IntAlloc(1000)
	p.vOutputs = vec.Vec_IntAlloc(1000)
	p.vNodes = vec.Vec_IntAlloc(10000)
	p.pSat = bsat.NewSolver()
	p.nSatVars = 1
	p.clkStart = util.Abc_Clock()
	p.pSat.SetnvarsC(1000)
	return p
}
func Bmc_MnaFree(p *Bmc_Mna_t) {
	p.pCnf = nil
	p.vPiMap = nil
	p.vId2Var = nil
	p.vInputs = nil
	p.vOutputs = nil
	p.vNodes = nil
	p.pSat.Delete()
}

// Gia_ManBmcCheckOutputs returns 1 if outputs iStart..iStop-1 are constant 0.
func Gia_ManBmcCheckOutputs(pFrames *gia.Gia_Man_t, iStart, iStop int32) int32 {
	for i := iStart; i < iStop; i++ {
		pObj := gia.Gia_ManPo(pFrames, i)
		if gia.Gia_ObjFaninLit0p(pFrames, pObj) != 0 {
			return 0
		}
	}
	return 1
}
func Gia_ManBmcFindFirst(pFrames *gia.Gia_Man_t) int32 {
	for i := int32(0); i < gia.Gia_ManPoNum(pFrames); i++ {
		pObj := gia.Gia_ManPo(pFrames, i)
		if gia.Gia_ObjFaninLit0p(pFrames, pObj) != 0 {
			return i
		}
	}
	return -1
}

// Gia_ManBmcCexGen generates the counter-example.
func Gia_ManBmcCexGen(pMan *Bmc_Mna_t, p *gia.Gia_Man_t, iOut int32) *util.Abc_Cex_t {
	iFramePi, iFrame := int32(0), int32(-1)
	pCex := util.Abc_CexAlloc(gia.Gia_ManRegNum(p), gia.Gia_ManPiNum(p), iOut/gia.Gia_ManPoNum(p)+1)
	pCex.IFrame = iOut / gia.Gia_ManPoNum(p)
	pCex.IPo = iOut % gia.Gia_ManPoNum(p)
	// fill in the input values
	for i := 0; i < vec.Vec_IntSize(pMan.vPiMap); i++ {
		iOrigPi := vec.Vec_IntEntry(pMan.vPiMap, i)
		if iOrigPi < 0 {
			iFrame = -iOrigPi - 1
			continue
		}
		// iOrigPi in iFrame of pGia has PI index iFramePi in pMan->pFrames,
		iObjId := gia.Gia_ObjId(pMan.pFrames, gia.Gia_ManPi(pMan.pFrames, iFramePi))
		iSatVar := vec.Vec_IntEntry(pMan.vId2Var, int(iObjId))
		if pMan.pSat.VarValueC(iSatVar) != 0 {
			util.Abc_InfoSetBit(pCex.PData, gia.Gia_ManRegNum(p)+gia.Gia_ManPiNum(p)*iFrame+iOrigPi)
		}
		iFramePi++
	}
	if iFramePi != gia.Gia_ManPiNum(pMan.pFrames) {
		panic("Gia_ManBmcCexGen: iFramePi == Gia_ManPiNum(pMan->pFrames)")
	}
	return pCex
}

func Gia_ManBmcAddCnfNew_rec(p *Bmc_Mna_t, pObj *gia.Gia_Obj_t) {
	iObj := gia.Gia_ObjId(p.pFrames, pObj)
	if vec.Vec_IntEntry(p.vId2Var, int(iObj)) > 0 {
		return
	}
	vec.Vec_IntWriteEntry(p.vId2Var, int(iObj), 1)
	if gia.Gia_ObjIsAnd(pObj) != 0 && p.pCnf.Obj2Count[iObj] == -1 {
		Gia_ManBmcAddCnfNew_rec(p, gia.Gia_ObjFanin0(pObj))
		Gia_ManBmcAddCnfNew_rec(p, gia.Gia_ObjFanin1(pObj))
		return
	}
	vec.Vec_IntWriteEntry(p.vId2Var, int(iObj), p.nSatVars)
	p.nSatVars++
	if gia.Gia_ObjIsAnd(pObj) != 0 || gia.Gia_ObjIsPo(p.pFrames, pObj) != 0 {
		var i int32
		Gia_ManBmcAddCnfNew_rec(p, gia.Gia_ObjFanin0(pObj))
		if gia.Gia_ObjIsAnd(pObj) != 0 {
			Gia_ManBmcAddCnfNew_rec(p, gia.Gia_ObjFanin1(pObj))
		}
		// extend the SAT solver
		if p.nSatVars > p.pSat.NVarsC() {
			p.pSat.SetnvarsC(p.nSatVars)
		}
		// add clauses
		nClas := p.pCnf.Obj2Count[iObj]
		iCla := p.pCnf.Obj2Clause[iObj]
		for i = 0; i < nClas; i++ {
			pClauseThis := p.pCnf.Clauses[iCla+i]
			pLits := make([]bsat.Literal, len(pClauseThis))
			for nLits, lit := range pClauseThis {
				if lit < 2 {
					fmt.Printf("\n\n\nError in CNF generation:  Constant literal!\n\n\n")
				}
				pLits[nLits] = bsat.Literal(util.Abc_Lit2LitV(vec.Vec_IntArray(p.vId2Var), lit))
			}
			if len(pLits) > 9 {
				panic("Gia_ManBmcAddCnfNew_rec: nLits <= 9")
			}
			if p.pSat.AddClauseC(pLits) == 0 {
				break
			}
		}
		if i < nClas {
			fmt.Printf("SAT solver became UNSAT after adding clauses.\n")
		}
	} else if gia.Gia_ObjIsCi(pObj) == 0 {
		panic("Gia_ManBmcAddCnfNew_rec: Gia_ObjIsCi(pObj)")
	}
}
func Gia_ManBmcAddCnfNew(p *Bmc_Mna_t, iStart, iStop int32) {
	for i := iStart; i < iStop; i++ {
		pObj := gia.Gia_ManPo(p.pFrames, i)
		if gia.Gia_ObjFaninId0p(p.pFrames, pObj) == 0 {
			continue
		}
		Gia_ManBmcAddCnfNew_rec(p, pObj)
	}
}

// Cnf_DeriveGia derives the old (Cnf_DeriveOther) CNF of the frames (-c).
func Cnf_DeriveGia(p *gia.Gia_Man_t) *bmcMnaCnf {
	pAig := gia.Gia_ManToAigSimple(p)
	pCnf := cnf.Cnf_DeriveOther(pAig, 1)
	return &bmcMnaCnf{Clauses: pCnf.Clauses, Obj2Clause: pCnf.Obj2Clause, Obj2Count: pCnf.Obj2Count}
}

func Gia_ManBmcPerformInt(pGia *gia.Gia_Man_t, pPars *Bmc_AndPar_t) int32 {
	var nFramesMax, f, i, Lit, status int32
	RetValue := int32(-2)
	clk := util.Abc_Clock()
	p := Bmc_MnaAlloc()
	if pPars.NTimeOut != 0 {
		p.pSat.SetRuntimeDeadline(util.Abc_Clock, int64(pPars.NTimeOut)*util.CLOCKS_PER_SEC+util.Abc_Clock())
	} else {
		p.pSat.SetRuntimeDeadline(util.Abc_Clock, 0)
	}
	p.pFrames = Gia_ManBmcUnroll(pGia, pPars.NFramesMax, pPars.NFramesAdd, pPars.FVeryVerbose, &p.vPiMap)
	nFramesMax = gia.Gia_ManPoNum(p.pFrames) / gia.Gia_ManPoNum(pGia)
	if pPars.FVerbose != 0 {
		fmt.Printf("Unfolding for %d frames with first non-trivial PO %d.  ", nFramesMax, Gia_ManBmcFindFirst(p.pFrames))
		util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	}
	if pPars.FUseSynth != 0 {
		pTemp := p.pFrames
		p.pFrames = gia.Gia_ManAigSyn2(pTemp, 1, 0, 0, 0, 0, pPars.FVerbose, 0)
		gia.Gia_ManStop(pTemp)
	} else if pPars.FVerbose != 0 {
		gia.Gia_ManPrintStats(p.pFrames, nil)
	}
	if pPars.FDumpFrames != 0 {
		if err := gia.Gia_AigerWrite(p.pFrames, "frames.aig", 0, 0, 0); err != nil {
			fmt.Printf("Gia_AigerWrite(): %v\n", err)
		}
		fmt.Printf("Dumped unfolded frames into file \"frames.aig\".\n")
	}
	if pPars.FUseOldCnf != 0 {
		p.pCnf = Cnf_DeriveGia(p.pFrames)
	} else {
		pCnf := gia.Mf_ManGenerateCnfMapper(p.pFrames, pPars.NLutSize, 1, 0, 0, pPars.FVerbose)
		p.pCnf = &bmcMnaCnf{Clauses: pCnf.Clauses, Obj2Clause: pCnf.Obj2Clause, Obj2Count: pCnf.Obj2Count}
	}
	vec.Vec_IntFillExtra(p.vId2Var, int(gia.Gia_ManObjNum(p.pFrames)), 0)
	for f = 0; f < nFramesMax; f++ {
		if Gia_ManBmcCheckOutputs(p.pFrames, f*gia.Gia_ManPoNum(pGia), (f+1)*gia.Gia_ManPoNum(pGia)) == 0 {
			// create another slice
			Gia_ManBmcAddCnfNew(p, f*gia.Gia_ManPoNum(pGia), (f+1)*gia.Gia_ManPoNum(pGia))
			// try solving the outputs
			for i = f * gia.Gia_ManPoNum(pGia); i < (f+1)*gia.Gia_ManPoNum(pGia); i++ {
				pObj := gia.Gia_ManPo(p.pFrames, i)
				if gia.Gia_ObjFaninLit0p(p.pFrames, pObj) == 0 {
					continue
				}
				if gia.Gia_ObjFaninLit0p(p.pFrames, pObj) == 1 {
					RetValue = 0 // trivially SAT: asserted for any inputs (CBUGS_FIXED.md #40)
					break
				}
				Lit = util.Abc_Var2Lit(vec.Vec_IntEntry(p.vId2Var, int(gia.Gia_ObjId(p.pFrames, pObj))), 0)
				status = p.pSat.SolveC([]bsat.Literal{bsat.Literal(Lit)}, int64(pPars.NConfLimit), 0, 0, 0)
				if status == l_False { // unsat
					continue
				}
				if status == l_True { // sat
					RetValue = 0
				}
				if status == l_Undef { // undecided
					RetValue = -1
				}
				break
			}
			// report statistics
			if pPars.FVerbose != 0 {
				fmt.Printf("%4d :  PI =%9d.  AIG =%9d.  Var =%8d.  In =%6d.  And =%9d.  Cla =%9d.  Conf =%9d.  Mem =%7.1f MB   ",
					f, gia.Gia_ManPiNum(p.pFrames), gia.Gia_ManAndNum(p.pFrames),
					p.nSatVars-1, vec.Vec_IntSize(p.vInputs), vec.Vec_IntSize(p.vNodes),
					p.pSat.NClauses(), p.pSat.NConflicts(), gia.Gia_ManMemory(p.pFrames)/(1<<20))
				util.Abc_PrintTime(1, "Time", util.Abc_Clock()-p.clkStart)
			}
		}
		if RetValue != -2 {
			if RetValue == -1 {
				fmt.Printf("SAT solver reached conflict/runtime limit in frame %d.\n", f)
			} else {
				pGia.ClearSequentialCounterexample()
				if gia.Gia_ObjFaninLit0p(p.pFrames, gia.Gia_ManPo(p.pFrames, i)) == 1 {
					// no SAT call: the all-zero input sequence is a CEX
					pCex := util.Abc_CexAlloc(gia.Gia_ManRegNum(pGia), gia.Gia_ManPiNum(pGia), f+1)
					pCex.IFrame = f
					pCex.IPo = i - f*gia.Gia_ManPoNum(pGia)
					pGia.TransferSequentialCounterexampleSource(pCex)
				} else {
					pGia.TransferSequentialCounterexampleSource(Gia_ManBmcCexGen(p, pGia, i))
				}
				fmt.Printf("Output %d of miter \"%s\" was asserted in frame %d.  ",
					i-f*gia.Gia_ManPoNum(pGia), gia.Gia_ManName(pGia), f)
				util.Abc_PrintTime(1, "Time", util.Abc_Clock()-p.clkStart)
			}
			break
		}
		pPars.IFrame = f
	}
	if RetValue == -2 {
		RetValue = -1
	}
	// cleanup
	gia.Gia_ManStop(p.pFrames)
	Bmc_MnaFree(p)
	return RetValue
}

// Gia_ManBmcPerform runs the engine, doubling the number of additional
// frames until the global timeout.
func Gia_ManBmcPerform(pGia *gia.Gia_Man_t, pPars *Bmc_AndPar_t) int32 {
	var TimeToStop int64
	if pPars.NTimeOut != 0 {
		TimeToStop = int64(pPars.NTimeOut)*util.CLOCKS_PER_SEC + util.Abc_Clock()
	}
	if pPars.NFramesAdd == 0 {
		return Gia_ManBmcPerformInt(pGia, pPars)
	}
	// iterate over the engine until we read the global timeout
	if pPars.NTimeOut < 0 {
		panic("Gia_ManBmcPerform: pPars->nTimeOut >= 0")
	}
	for {
		if TimeToStop != 0 && TimeToStop < util.Abc_Clock() {
			return -1
		}
		if Gia_ManBmcPerformInt(pGia, pPars) == 0 {
			return 0
		}
		// set the new runtime limit
		if pPars.NTimeOut != 0 {
			pPars.NTimeOut = util.Abc_MinInt(pPars.NTimeOut-1, int32((TimeToStop-util.Abc_Clock())/util.CLOCKS_PER_SEC))
			if pPars.NTimeOut <= 0 {
				return -1
			}
		} else {
			return -1
		}
		// set the new frames limit
		pPars.NFramesAdd *= 2
	}
}
