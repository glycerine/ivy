package cec

// Verification entry points ported line by line from ABC's
// src/proof/cec/cecCec.c (Alan Mishchenko, UC Berkeley).

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/proof/fracomb"
	"github.com/glycerine/abcgo/proof/framgr"
)

// Cec_ManTransformPattern saves the input pattern with the given number.
// A nil vector means the all-zero pattern.
func Cec_ManTransformPattern(p *Gia_Man_t, iOut int32, pValues []int32) {
	if p.PCexComb != nil {
		panic("cec: Cec_ManTransformPattern: p->pCexComb == NULL")
	}
	p.PCexComb = util.Abc_CexAlloc(0, gia.Gia_ManCiNum(p), 1)
	p.PCexComb.IPo = iOut
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		if pValues != nil && pValues[i] != 0 {
			util.Abc_InfoSetBit(p.PCexComb.PData, i)
		}
	}
}

// Cec_ManVerifyOld is the interface to the old CEC engine.
func Cec_ManVerifyOld(pMiter *Gia_Man_t, fVerbose int32, piOutFail *int32, clkTotal abctime, fSilent int32) int32 {
	pTemp := gia.Gia_ManTransformMiter(pMiter)
	pMiterCec := gia.Gia_ManToAig(pTemp, 0)
	var iOut, nOuts int32
	if piOutFail != nil {
		*piOutFail = -1
	}
	gia.Gia_ManStop(pTemp)
	// run CEC on this miter
	RetValue := fracomb.Fra_FraigCec(&pMiterCec, 10000000, fVerbose, framgr.SweepForCEC)
	// report the miter
	if RetValue == 1 {
		if fSilent == 0 {
			util.Abc_Print(1, "Networks are equivalent.  ")
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clkTotal)
		}
	} else if RetValue == 0 {
		if fSilent == 0 {
			util.Abc_Print(1, "Networks are NOT EQUIVALENT.  ")
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clkTotal)
		}
		pData, _ := pMiterCec.Data.([]int32)
		if pData == nil {
			util.Abc_Print(1, "Counter-example is not available.\n")
		} else {
			iOut = saig.Ssw_SecCexResimulate(pMiterCec, pData, &nOuts)
			if iOut == -1 {
				util.Abc_Print(1, "Counter-example verification has failed.\n")
			} else {
				if fSilent == 0 {
					util.Abc_Print(1, "Primary output %d has failed", iOut)
					if nOuts-1 >= 0 {
						util.Abc_Print(1, ", along with other %d incorrect outputs", nOuts-1)
					}
					util.Abc_Print(1, ".\n")
				}
				if piOutFail != nil {
					*piOutFail = iOut
				}
			}
			Cec_ManTransformPattern(pMiter, iOut, pData)
		}
	} else if fSilent == 0 {
		util.Abc_Print(1, "Networks are UNDECIDED.  ")
		util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clkTotal)
	}
	aig.Aig_ManStop(pMiterCec)
	return RetValue
}

// Cec_ManHandleSpecialCases resolves the trivial pair cases:
// 1) (SAT) a pair of POs have different value under all-0 pattern
// 2) (SAT) a pair of POs has different PI/Const drivers
// 3) (UNSAT) 1-2 do not hold and there is no nodes
// It returns -1 when a nontrivial proof remains.
func Cec_ManHandleSpecialCases(p *Gia_Man_t, pPars *Cec_ParCec_t) int32 {
	clk := util.Abc_Clock()
	gia.Gia_ManSetPhase(p)
	for i := int32(0); i < gia.Gia_ManPoNum(p); i++ {
		pObj1 := gia.Gia_ManPo(p, i)
		i++
		pObj2 := gia.Gia_ManPo(p, i)
		// check if they different on all-0 pattern
		// (for example, when they have the same driver but complemented)
		if gia.Gia_ObjPhase(pObj1) != gia.Gia_ObjPhase(pObj2) {
			if pPars.FSilent == 0 {
				util.Abc_Print(1, "Networks are NOT EQUIVALENT. Output %d trivially differs (different phase).  ", i/2)
				util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
			}
			pPars.iOutFail = i / 2
			Cec_ManTransformPattern(p, i/2, nil)
			return 0
		}
		// get the drivers
		pDri1 := gia.Gia_ObjFanin0(pObj1)
		pDri2 := gia.Gia_ObjFanin0(pObj2)
		// drivers are different PIs
		if gia.Gia_ObjIsPi(p, pDri1) != 0 && gia.Gia_ObjIsPi(p, pDri2) != 0 && pDri1 != pDri2 {
			if pPars.FSilent == 0 {
				util.Abc_Print(1, "Networks are NOT EQUIVALENT. Output %d trivially differs (different PIs).  ", i/2)
				util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
			}
			pPars.iOutFail = i / 2
			Cec_ManTransformPattern(p, i/2, nil)
			// if their compl attributes are the same - one should be complemented
			if gia.Gia_ObjFaninC0(pObj1) != gia.Gia_ObjFaninC0(pObj2) {
				panic("cec: Cec_ManHandleSpecialCases: Gia_ObjFaninC0(pObj1) == Gia_ObjFaninC0(pObj2)")
			}
			util.Abc_InfoSetBit(p.PCexComb.PData, gia.Gia_ObjCioId(pDri1))
			return 0
		}
		// one of the drivers is a PI; another is a constant 0
		if (gia.Gia_ObjIsPi(p, pDri1) != 0 && gia.Gia_ObjIsConst0(pDri2) != 0) ||
			(gia.Gia_ObjIsPi(p, pDri2) != 0 && gia.Gia_ObjIsConst0(pDri1) != 0) {
			if pPars.FSilent == 0 {
				util.Abc_Print(1, "Networks are NOT EQUIVALENT. Output %d trivially differs (PI vs. constant).  ", i/2)
				util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
			}
			pPars.iOutFail = i / 2
			Cec_ManTransformPattern(p, i/2, nil)
			// the compl attributes are the same - the PI should be complemented
			if gia.Gia_ObjFaninC0(pObj1) != gia.Gia_ObjFaninC0(pObj2) {
				panic("cec: Cec_ManHandleSpecialCases: Gia_ObjFaninC0(pObj1) == Gia_ObjFaninC0(pObj2)")
			}
			if gia.Gia_ObjIsPi(p, pDri1) != 0 {
				util.Abc_InfoSetBit(p.PCexComb.PData, gia.Gia_ObjCioId(pDri1))
			} else {
				util.Abc_InfoSetBit(p.PCexComb.PData, gia.Gia_ObjCioId(pDri2))
			}
			return 0
		}
	}
	if gia.Gia_ManAndNum(p) == 0 {
		if pPars.FSilent == 0 {
			util.Abc_Print(1, "Networks are equivalent.  ")
			util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
		}
		return 1
	}
	return -1
}

// Cec_ManVerify is the new CEC engine.
func Cec_ManVerify(pInit *Gia_Man_t, pPars *Cec_ParCec_t) int32 {
	if pPars == nil {
		pPars = new(Cec_ParCec_t)
		Cec_ManCecSetDefaultParams(pPars)
	}
	pPars.Error = nil
	var ParsFra Cec_ParFra_t
	pParsFra := &ParsFra
	var p, pNew *Gia_Man_t
	var RetValue int32
	clk := util.Abc_Clock()
	clkTotal := util.Abc_Clock()
	// consider special cases:
	// 1) (SAT) a pair of POs have different value under all-0 pattern
	// 2) (SAT) a pair of POs has different PI/Const drivers
	// 3) (UNSAT) 1-2 do not hold and there is no nodes
	RetValue = Cec_ManHandleSpecialCases(pInit, pPars)
	if RetValue == 0 || RetValue == 1 {
		return RetValue
	}
	// preprocess
	p = gia.Gia_ManDup(pInit)
	gia.Gia_ManEquivFixOutputPairs(p)
	pNew = p
	p = gia.Gia_ManCleanup(pNew)
	gia.Gia_ManStop(pNew)
	if pPars.FNaive != 0 {
		RetValue = Cec_ManVerifyNaive(p, pPars)
		gia.Gia_ManStop(p)
		return RetValue
	}
	if pInit.VSimsPi != nil {
		p.VSimsPi = vec.Vec_WrdDup(pInit.VSimsPi)
		p.NSimWords = pInit.NSimWords
	}
	// sweep for equivalences
	Cec_ManFraSetDefaultParams(pParsFra)
	pParsFra.NItersMax = 1000
	pParsFra.NBTLimit = pPars.NBTLimit
	pParsFra.TimeLimit = pPars.TimeLimit
	pParsFra.FVerbose = pPars.FVerbose
	pParsFra.FVeryVerbose = pPars.FVeryVerbose
	pParsFra.FCheckMiter = 1
	pParsFra.FDualOut = 1
	pNew = Cec_ManSatSweeping(p, pParsFra, pPars.FSilent)
	if pParsFra.Error != nil {
		// Go-only: the sweep rejected its input (no C counterpart).
		pPars.Error = pParsFra.Error
		gia.Gia_ManStop(p)
		return -1
	}
	pPars.iOutFail = pParsFra.iOutFail
	// update
	pInit.PCexComb = p.PCexComb
	p.PCexComb = nil
	gia.Gia_ManStop(p)
	p = pInit
	// continue
	if pNew == nil {
		if p.PCexComb != nil {
			if p.PCexComb != nil && gia.Gia_ManVerifyCex(p, p.PCexComb, 1) == 0 {
				util.Abc_Print(1, "Counter-example simulation has failed.\n")
			}
			if pPars.FSilent == 0 {
				util.Abc_Print(1, "Networks are NOT EQUIVALENT.  ")
				util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
			}
			return 0
		}
		p = gia.Gia_ManDup(pInit)
		gia.Gia_ManEquivFixOutputPairs(p)
		pNew = p
		p = gia.Gia_ManCleanup(pNew)
		gia.Gia_ManStop(pNew)
		pNew = p
	}
	if pPars.FVerbose != 0 {
		util.Abc_Print(1, "Networks are UNDECIDED after the new CEC engine.  ")
		util.Abc_PrintTime(1, "Time", util.Abc_Clock()-clk)
	}
	// fDumpUndecided is 0 in the C source
	if pPars.TimeLimit != 0 && (util.Abc_Clock()-clkTotal)/util.CLOCKS_PER_SEC >= int64(pPars.TimeLimit) {
		gia.Gia_ManStop(pNew)
		return -1
	}
	// call other solver
	if pPars.FVerbose != 0 {
		util.Abc_Print(1, "Calling the old CEC engine.\n")
	}
	RetValue = Cec_ManVerifyOld(pNew, pPars.FVerbose, &pPars.iOutFail, clkTotal, pPars.FSilent)
	p.PCexComb = pNew.PCexComb
	pNew.PCexComb = nil
	if p.PCexComb != nil && gia.Gia_ManVerifyCex(p, p.PCexComb, 1) == 0 {
		util.Abc_Print(1, "Counter-example simulation has failed.\n")
	}
	gia.Gia_ManStop(pNew)
	return RetValue
}

// Cec_ManVerifySimple is a simple SAT run to check equivalence.
func Cec_ManVerifySimple(p *Gia_Man_t) int32 {
	var ParsCec Cec_ParCec_t
	pPars := &ParsCec
	Cec_ManCecSetDefaultParams(pPars)
	pPars.FSilent = 1
	if gia.Gia_ManCoNum(p) != 2 {
		panic("cec: Cec_ManVerifySimple: Gia_ManCoNum(p) == 2")
	}
	if gia.Gia_ManRegNum(p) != 0 {
		panic("cec: Cec_ManVerifySimple: Gia_ManRegNum(p) == 0")
	}
	return Cec_ManVerify(p, pPars)
}

// Cec_ManVerifyTwo applies the new CEC engine to two circuits.
func Cec_ManVerifyTwo(p0, p1 *Gia_Man_t, fVerbose int32) int32 {
	if p0 == nil || p1 == nil {
		return -1 // Go-only guard: C dereferences both managers
	}
	var ParsCec Cec_ParCec_t
	pPars := &ParsCec
	Cec_ManCecSetDefaultParams(pPars)
	pPars.FVerbose = fVerbose
	pMiter := gia.Gia_ManMiter(p0, p1, 0, 1, 0, 0, pPars.FVerbose)
	if pMiter == nil {
		return -1
	}
	RetValue := Cec_ManVerify(pMiter, pPars)
	p0.PCexComb = pMiter.PCexComb
	pMiter.PCexComb = nil
	gia.Gia_ManStop(pMiter)
	return RetValue
}

// cecRequireAIG rejects native nodes that the source pipeline does not
// preserve and that the current sweep port cannot process faithfully.
func cecRequireAIG(graph *Gia_Man_t) error {
	if graph == nil {
		return fmt.Errorf("cec: nil graph")
	}
	for id := int32(0); int(id) < graph.ObjectCount(); id++ {
		_, _, _, kind := graph.Gate(id)
		if kind == 3 || kind == 4 || kind == 5 {
			return fmt.Errorf("cec: unsupported native gate kind %d at object %d", kind, id)
		}
	}
	return nil
}
