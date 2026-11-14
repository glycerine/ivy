package if_

import (
	"github.com/glycerine/abcgo/misc/mem"
	"github.com/glycerine/abcgo/misc/tim"
	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type word = uint64
type abctime = int64
type Vec_Ptr_t = vec.Vec_Ptr_t
type Vec_Int_t = vec.Vec_Int_t
type Vec_Wrd_t = vec.Vec_Wrd_t
type Vec_Str_t = vec.Vec_Str_t
type Vec_Mem_t = vec.Vec_Mem_t
type Vec_Wec_t = vec.Vec_Wec_t
type Hash_IntMan_t = vec.Hash_IntMan_t
type Mem_Fixed_t = mem.Mem_Fixed_t
type Mem_Flex_t = mem.Mem_Flex_t
type Tim_Man_t = tim.Tim_Man_t

func Vec_PtrSize(v *Vec_Ptr_t) int                      { return vec.Vec_PtrSize(v) }
func Vec_PtrEntry(v *Vec_Ptr_t, i int) any              { return vec.Vec_PtrEntry(v, i) }
func Vec_PtrArray(v *Vec_Ptr_t) []any                   { return vec.Vec_PtrArray(v) }
func Vec_IntArray(v *Vec_Int_t) []int32                 { return vec.Vec_IntArray(v) }
func Vec_MemReadEntry(v *Vec_Mem_t, i int) []uint64     { return vec.Vec_MemReadEntry(v, i) }
func Vec_StrEntryP(v *Vec_Str_t, i int) *byte           { return vec.Vec_StrEntryP(v, i) }
func Abc_Lit2Var(lit int32) int32                       { return abcutil.Abc_Lit2Var(lit) }
func Abc_LitIsCompl(lit int32) int32                    { return abcutil.Abc_LitIsCompl(lit) }
func Abc_Lit2LitL(values []int32, lit int32) int32      { return abcutil.Abc_Lit2LitL(values, lit) }
func Abc_MaxInt(a, b int32) int32                       { return abcutil.Abc_MaxInt(a, b) }
func Abc_Tt4Check(truth int32) int32                    { return abcutil.Abc_Tt4Check(truth) }
func Abc_Clock() int64                                  { return abcutil.Abc_Clock() }
func Abc_Print(level int32, format string, args ...any) { abcutil.Abc_Print(level, format, args...) }
func Abc_PrintTime(level int32, label string, elapsed int64) {
	abcutil.Abc_PrintTime(level, label, elapsed)
}

/**CFile****************************************************************

  FileName    [if.h]

  SystemName  [ABC: Logic synthesis and verification system.]

  PackageName [FPGA mapping based on priority cuts.]

  Synopsis    [External declarations.]

  Author      [Alan Mishchenko]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - November 21, 2006.]

  Revision    [$Id: if.h,v 1.00 2006/11/21 00:00:00 alanmi Exp $]

***********************************************************************/

////////////////////////////////////////////////////////////////////////
///                          INCLUDES                                ///
////////////////////////////////////////////////////////////////////////

//#include <stdio.h>
//#include <stdlib.h>
//#include <string.h>
//#include <assert.h>

//#include "misc/vec/vec.h"
//#include "misc/mem/mem.h"
//#include "misc/tim/tim.h"
//#include "misc/util/utilNam.h"
//#include "misc/vec/vecMem.h"
//#include "misc/util/utilTruth.h"
//#include "opt/dau/dau.h"
//#include "misc/vec/vecHash.h"
//#include "misc/vec/vecWec.h"

//ABC_NAMESPACE_HEADER_START

////////////////////////////////////////////////////////////////////////
///                         PARAMETERS                               ///
////////////////////////////////////////////////////////////////////////

// the maximum size of LUTs used for mapping (should be the same as FPGA_MAX_LUTSIZE defined in "fpga.h"!!!)
const IF_MAX_LUTSIZE = 32

// the largest possible number of LUT inputs when funtionality of the LUTs are computed
const IF_MAX_FUNC_LUTSIZE = 15

// a very large number
const IF_INFINITY = 100000000

// the largest possible user cut cost
const IF_COST_MAX = 4095 // ((1<<12)-1)
const IF_FLOAT_LARGE float32 = 1.0e20
const IF_FLOAT_SMALL float32 = 1.0e-20
const IF_INT_LARGE = 10000000

const IF_BIG_CHAR byte = 120

// object types
type If_Type_t int32

const (
	IF_NONE If_Type_t = iota
	// 0: non-existent object
	IF_CONST1
	// 1: constant 1
	IF_CI
	// 2: combinational input
	IF_CO
	// 3: combinational output
	IF_AND
	// 4: AND node
	IF_VOID // 5: unused object

)

////////////////////////////////////////////////////////////////////////
///                         BASIC TYPES                              ///
////////////////////////////////////////////////////////////////////////

type If_Man_t = If_Man_t_
type If_Par_t = If_Par_t_
type If_Obj_t = If_Obj_t_
type If_Cut_t = If_Cut_t_
type If_Set_t = If_Set_t_
type If_LibLut_t = If_LibLut_t_
type If_LibBox_t = If_LibBox_t_
type If_DsdMan_t = If_DsdMan_t_
type Ifn_Ntk_t = Ifn_Ntk_t_

type Ifif_Par_t struct {
	nLutSize                                  int32
	pLutLib                                   *If_LibLut_t
	pLutDelays                                [IF_MAX_LUTSIZE]float32
	DelayWire                                 float32
	nDegree, fCascade, fVerbose, fVeryVerbose int32
}

type If_Par_t_ struct {
	nLutSize, nCutsMax, nFlowIters, nAreaIters, nGateSize, nNonDecLimit               int32
	DelayTarget, Epsilon                                                              float32
	nRelaxRatio, nStructType, nAndDelay, nAndArea                                     int32
	fPreprocess, fArea, fFancy, fExpRed, fLatchPaths, fLut6Filter                     int32
	fEdge, fPower, fCutMin, fDelayOpt, fDelayOptLut, fDsdBalance                      int32
	fUserRecLib, fUserSesLib, fBidec, fUse34Spec, fUseBat, fUseBuffs                  int32
	fEnableCheck07, fEnableCheck08, fEnableCheck10, fEnableCheck75, fEnableCheck75u   int32
	fUseDsd, fUseDsdTune, fUseCofVars, fUseAndVars, fUseTtPerm, fDeriveLuts           int32
	fDoAverage, fHashMapping, fVerbose, fVerboseTrace                                 int32
	pLutStruct                                                                        string
	WireDelay                                                                         float32
	fSkipCutFilter, fAreaOnly, fTruth, fUsePerm, fUseBdds, fUseSops, fUseCnfs, fUseMv int32
	nLatchesCi, nLatchesCo, nLatchesCiBox, nLatchesCoBox, fLiftLeaves, fUseCoAttrs    int32
	DelayTargetNew, FinalDelay, FinalArea                                             float32
	pLutLib                                                                           *If_LibLut_t
	pTimesArr, pTimesReq                                                              []float32
	pFuncCost                                                                         func(*If_Man_t, *If_Cut_t) int32
	pFuncUser                                                                         func(*If_Man_t, *If_Obj_t, *If_Cut_t) int32
	pFuncCell                                                                         func(*If_Man_t, []uint32, int32, int32, string) int32
	pFuncCell2                                                                        func(*If_Man_t, []uint64, int32, *Vec_Str_t, *[]string) int32
	pFuncWrite                                                                        func(*If_Man_t) int32
	pFuncRecGia                                                                       func(*Gia_Man_t, *If_Man_t, *If_Cut_t, *Vec_Int_t, int32) int32
	pDsdManager                                                                       *If_DsdMan_t
	pFuncSwitching                                                                    func(*If_Man_t)
	pFuncDelayTrace                                                                   func(*Gia_Man_t)
	pReoMan                                                                           any
}

type If_LibLut_t_ struct {
	pName                 string
	LutMax, fVarPinDelays int32
	pLutAreas             [IF_MAX_LUTSIZE + 1]float32
	pLutDelays            [IF_MAX_LUTSIZE + 1][IF_MAX_LUTSIZE + 1]float32
}

type If_Man_t_ struct {
	pName                                           string
	pPars                                           *If_Par_t
	pConst1                                         *If_Obj_t
	vCis, vCos, vObjs, vObjsRev, vTemp              *Vec_Ptr_t
	nObjs                                           [IF_VOID + 1]int32
	nLevelMax                                       int32
	fEpsilon, RequiredGlo, RequiredGlo2, AreaGlo    float32
	nNets                                           int32
	dPower                                          float32
	nCutsUsed, nCutsMerged                          int32
	puTemp                                          [4][]uint32
	puTempW                                         []uint64
	pTruth32                                        []uint32
	SortMode, fNextRound, nChoices                  int32
	vSwitching                                      *Vec_Int_t
	pPerm                                           [3][IF_MAX_LUTSIZE]int32
	uSharedMask                                     uint32
	nShared, fReqTimeWarn                           int32
	vCover, vArray                                  *Vec_Int_t
	vAnds, vOrGate, vAndGate                        *Vec_Wrd_t
	vLatchOrder                                     *Vec_Ptr_t
	vLags                                           *Vec_Int_t
	nAttempts, nMaxIters, Period                    int32
	nTruth6Words                                    [IF_MAX_FUNC_LUTSIZE + 1]int32
	nPermWords, nObjBytes, nCutBytes, nSetBytes     int32
	pMemObj, pMemSet                                *Mem_Fixed_t
	pMemCi, pMemAnd, pFreeList                      *If_Set_t
	ciCutSets                                       []*If_Set_t
	ownedCutSets                                    []*If_Set_t
	stopped                                         bool
	nSmallSupp, nCutsTotal                          int32
	nCutsUseless, nCutsCount                        [32]int32
	nCutsCountAll, nCutsUselessAll, nCuts5, nCuts5a int32
	pIfDsdMan                                       *If_DsdMan_t
	vTtMem                                          [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Mem_t
	vTtIsops                                        [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Wec_t
	vTtIsopComplement                               [IF_MAX_FUNC_LUTSIZE + 1][]bool
	vTtDsds                                         [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Int_t
	vTtPerms, vTtVars                               [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Str_t
	vTtDecs, vTtOccurs                              [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Int_t
	vPairHash                                       *Hash_IntMan_t
	vPairRes                                        *Vec_Int_t
	vPairPerms                                      *Vec_Str_t
	vTtMem6                                         *Vec_Mem_t
	pCanonPerm                                      [IF_MAX_LUTSIZE]byte
	uCanonPhase                                     uint32
	nCacheHits, nCacheMisses                        int32
	timeCache                                       [6]abctime
	nBestCutSmall, nCountNonDec                     [2]int32
	vCutData                                        *Vec_Int_t
	pArrTimeProfile                                 [IF_MAX_FUNC_LUTSIZE]int32
	vVisited                                        *Vec_Ptr_t
	pUserMan                                        any
	vDump                                           *Vec_Int_t
	pDumpIns                                        [16]int32
	pManTim                                         *Tim_Man_t
	vCoAttrs                                        *Vec_Int_t
	nTableSize, nTableEntries                       [2]int32
	pHashTable                                      [2]map[uint32]any
	pMemEntries                                     *Mem_Fixed_t
}

type If_Cut_t_ struct {
	Area, Edge, Power, Delay                                float32
	iCutFunc, uMaskFunc                                     int32
	uSign                                                   uint32
	Cost, fCompl, fUser, fUseless, fAndCut, nLimit, nLeaves uint32
	pLeaves                                                 []int32
	permutation                                             []byte
	data                                                    any
	dataInt                                                 int32
}

type If_Set_t_ struct {
	nCutsMax, nCuts int16
	pNext           *If_Set_t
	ppCuts          []*If_Cut_t
}

type If_Obj_t_ struct {
	Type                                                                     If_Type_t
	fCompl0, fCompl1, fPhase, fRepr, fMark, fVisit, fSpec, fDriver, fSkipCut uint32
	Level                                                                    uint32
	Id, IdPio, nRefs, nVisits, nVisitsCopy                                   int32
	pFanin0, pFanin1, pEquiv                                                 *If_Obj_t
	EstRefs, Required, LValue                                                float32
	pCopy                                                                    any
	iCopy                                                                    int32
	pCutSet                                                                  *If_Set_t
	CutBest                                                                  If_Cut_t
	regular                                                                  *If_Obj_t
	complement                                                               bool
	complemented                                                             *If_Obj_t
}

type If_Box_t struct {
	pName                         string
	fSeq, fBlack, fOuter, fUnused byte
	Id, nPis, nPos                int32
	pDelays                       []int32
}

type If_LibBox_t_ struct {
	nBoxes int32
	vBoxes *Vec_Ptr_t
}

func If_Regular(p *If_Obj_t) *If_Obj_t {
	if p == nil {
		return nil
	}
	if p.regular != nil {
		return p.regular
	}
	return p
}
func If_Not(p *If_Obj_t) *If_Obj_t {
	if p == nil {
		return nil
	}
	if p.complement {
		return If_Regular(p)
	}
	p = If_Regular(p)
	if p.complemented == nil {
		q := *p
		q.regular = p
		q.complement = true
		q.complemented = p
		p.complemented = &q
	}
	return p.complemented
}
func If_NotCond(p *If_Obj_t, c int32) *If_Obj_t {
	if c != 0 {
		return If_Not(p)
	}
	return p
}
func If_IsComplement(p *If_Obj_t) int32 {
	if p != nil && p.complement {
		return 1
	}
	return 0
}

func If_ManCiNum(p *If_Man_t) int32      { return p.nObjs[IF_CI] }
func If_ManCoNum(p *If_Man_t) int32      { return p.nObjs[IF_CO] }
func If_ManAndNum(p *If_Man_t) int32     { return p.nObjs[IF_AND] }
func If_ManObjNum(p *If_Man_t) int32     { return int32(Vec_PtrSize(p.vObjs)) }
func If_ManObjBytes(p *If_Man_t) int32   { return p.nObjBytes }
func If_ManArray(p *If_Man_t) *Vec_Int_t { return p.vArray }

func If_ManConst1(p *If_Man_t) *If_Obj_t         { return p.pConst1 }
func If_ManSetName(p *If_Man_t, name string)     { p.pName = name }
func ifObjectAt(v *Vec_Ptr_t, i int32) *If_Obj_t { return Vec_PtrEntry(v, int(i)).(*If_Obj_t) }
func If_ManCi(p *If_Man_t, i int32) *If_Obj_t    { return ifObjectAt(p.vCis, i) }
func If_ManCo(p *If_Man_t, i int32) *If_Obj_t    { return ifObjectAt(p.vCos, i) }
func If_ManLi(p *If_Man_t, i int32) *If_Obj_t {
	return If_ManCo(p, If_ManCoNum(p)-p.pPars.nLatchesCo+i)
}
func If_ManLo(p *If_Man_t, i int32) *If_Obj_t {
	return If_ManCi(p, If_ManCiNum(p)-p.pPars.nLatchesCi+i)
}
func If_ManObj(p *If_Man_t, i int32) *If_Obj_t { return ifObjectAt(p.vObjs, i) }

func If_ObjIsConst1(p *If_Obj_t) int32 {
	if p.Type == IF_CONST1 {
		return 1
	}
	return 0
}
func If_ObjIsCi(p *If_Obj_t) int32 {
	if p.Type == IF_CI {
		return 1
	}
	return 0
}
func If_ObjIsCo(p *If_Obj_t) int32 {
	if p.Type == IF_CO {
		return 1
	}
	return 0
}
func If_ObjIsTerm(p *If_Obj_t) int32 {
	if p.Type == IF_CI || p.Type == IF_CO {
		return 1
	}
	return 0
}
func If_ObjIsLatch(p *If_Obj_t) int32 {
	if If_ObjIsCi(p) != 0 && p.pFanin0 != nil {
		return 1
	}
	return 0
}
func If_ObjIsAnd(p *If_Obj_t) int32 {
	if p.Type == IF_AND {
		return 1
	}
	return 0
}

func If_ObjId(pObj *If_Obj_t) int32                  { return pObj.Id }
func If_ObjFanin0(pObj *If_Obj_t) *If_Obj_t          { return pObj.pFanin0 }
func If_ObjFanin1(pObj *If_Obj_t) *If_Obj_t          { return pObj.pFanin1 }
func If_ObjFaninC0(pObj *If_Obj_t) int32             { return int32(pObj.fCompl0) }
func If_ObjFaninC1(pObj *If_Obj_t) int32             { return int32(pObj.fCompl1) }
func If_ObjCopy(pObj *If_Obj_t) any                  { return pObj.pCopy }
func If_ObjLevel(pObj *If_Obj_t) int32               { return int32(pObj.Level) }
func If_ObjSetLevel(pObj *If_Obj_t, level int32)     { pObj.Level = uint32(level) }
func If_ObjSetCopy(pObj *If_Obj_t, pCopy any)        { pObj.pCopy = pCopy }
func If_ObjSetChoice(pObj *If_Obj_t, pEqu *If_Obj_t) { pObj.pEquiv = pEqu }

func If_CutLeaveNum(pCut *If_Cut_t) int32          { return int32(pCut.nLeaves) }
func If_CutLeaves(pCut *If_Cut_t) []int32          { return pCut.pLeaves }
func If_CutLimit(pCut *If_Cut_t) int32             { return int32(pCut.nLimit) }
func If_CutCost(pCut *If_Cut_t) uint32             { return pCut.Cost }
func If_CutSetCost(pCut *If_Cut_t, cost uint32)    { pCut.Cost = cost }
func If_CutSetUser(pCut *If_Cut_t, value int32)    { pCut.fUser = uint32(value) }
func If_CutSetUseless(pCut *If_Cut_t, value int32) { pCut.fUseless = uint32(value) }
func If_CutComplement(pCut *If_Cut_t) int32        { return int32(pCut.fCompl) }
func If_CutLeaf(p *If_Man_t, pCut *If_Cut_t, i int32) *If_Obj_t {
	assert(i >= 0 && i < int32(pCut.nLeaves), "i >= 0 && i < int32(pCut.nLeaves)")
	id := pCut.pLeaves[i]
	if p.pPars != nil && p.pPars.fLiftLeaves != 0 {
		id >>= 8
	}
	return If_ManObj(p, id)
}
func If_CutSuppMask(pCut *If_Cut_t) uint32 {
	if pCut.nLeaves >= 32 {
		return ^uint32(0)
	}
	return ^uint32(0) >> (32 - pCut.nLeaves)
}
func If_CutTruthWords(nVarsMax int32) int32 {
	if nVarsMax <= 5 {
		return 2
	}
	return 1 << (nVarsMax - 5)
}
func If_CutPermWords(nVarsMax int32) int32        { return (nVarsMax + 3) / 4 }
func If_CutLeafBit(pCut *If_Cut_t, i int32) int32 { return int32((pCut.uMaskFunc >> uint32(i)) & 1) }
func If_CutPerm(pCut *If_Cut_t) []byte            { return pCut.permutation }
func If_CutCopy(_ *If_Man_t, dst, src *If_Cut_t) {
	leaves := append([]int32(nil), src.pLeaves...)
	permutation := append([]byte(nil), src.permutation...)
	*dst = *src
	dst.pLeaves, dst.permutation = leaves, permutation
}
func If_CutSetup(p *If_Man_t, cut *If_Cut_t) {
	limit := uint32(p.pPars.nLutSize)
	*cut = If_Cut_t_{nLimit: limit}
}

func If_ObjCutBest(pObj *If_Obj_t) *If_Cut_t { return &pObj.CutBest }
func If_ObjCutSign(id uint32) uint32         { return uint32(1) << (id % 31) }
func If_ObjCutSignCompute(p *If_Cut_t) uint32 {
	var sign uint32
	for _, id := range p.pLeaves {
		sign |= If_ObjCutSign(uint32(id))
	}
	return sign
}

func If_ObjArrTime(pObj *If_Obj_t) float32             { return If_ObjCutBest(pObj).Delay }
func If_ObjSetArrTime(pObj *If_Obj_t, ArrTime float32) { If_ObjCutBest(pObj).Delay = ArrTime }

func If_ObjLValue(pObj *If_Obj_t) float32           { return pObj.LValue }
func If_ObjSetLValue(pObj *If_Obj_t, value float32) { pObj.LValue = value }

func If_CutData(pCut *If_Cut_t) any               { return pCut.data }
func If_CutSetData(pCut *If_Cut_t, data any)      { pCut.data = data }
func If_CutDataInt(pCut *If_Cut_t) int32          { return pCut.dataInt }
func If_CutSetDataInt(pCut *If_Cut_t, data int32) { pCut.dataInt = data }

func If_CutTruthLit(pCut *If_Cut_t) int32 {
	assert(pCut.iCutFunc >= 0, "pCut.iCutFunc >= 0")
	return pCut.iCutFunc
}
func If_CutTruthIsCompl(pCut *If_Cut_t) int32 {
	assert(pCut.iCutFunc >= 0, "pCut.iCutFunc >= 0")
	return Abc_LitIsCompl(pCut.iCutFunc)
}
func If_CutTruthWR(p *If_Man_t, pCut *If_Cut_t) *word {
	assert(pCut.iCutFunc >= 0, "pCut.iCutFunc >= 0")
	v := p.vTtMem[pCut.nLeaves]
	if v == nil {
		return nil
	}
	entry := Vec_MemReadEntry(v, int(Abc_Lit2Var(pCut.iCutFunc)))
	if len(entry) == 0 {
		return nil
	}
	return &entry[0]
}
func If_CutTruthUR(p *If_Man_t, pCut *If_Cut_t) *uint32 {
	words := If_CutTruthW(p, pCut)
	if words == nil {
		return nil
	}
	_ = words
	n := int(p.nTruth6Words[pCut.nLeaves])
	if len(p.pTruth32) < 2*n {
		p.pTruth32 = make([]uint32, 2*n)
	}
	for i, w := range p.puTempW[:n] {
		p.pTruth32[2*i], p.pTruth32[2*i+1] = uint32(w), uint32(w>>32)
	}
	return &p.pTruth32[0]
}
func If_CutTruthW(p *If_Man_t, pCut *If_Cut_t) *word {
	assert(pCut.iCutFunc >= 0, "pCut.iCutFunc >= 0")
	n := int(p.nTruth6Words[pCut.nLeaves])
	if n == 0 {
		return nil
	}
	if len(p.puTempW) < n {
		p.puTempW = make([]uint64, n)
	}
	src := If_CutTruthWR(p, pCut)
	if src == nil {
		return nil
	}
	entry := Vec_MemReadEntry(p.vTtMem[pCut.nLeaves], int(Abc_Lit2Var(pCut.iCutFunc)))
	copy(p.puTempW[:n], entry)
	if If_CutTruthIsCompl(pCut) != 0 {
		for i := 0; i < n; i++ {
			p.puTempW[i] = ^p.puTempW[i]
		}
	}
	return &p.puTempW[0]
}

// If_CutTruthWordSlice returns the word slice backing the current cut truth.
// Calling it computes the truth using the same scratch buffer as If_CutTruthW.
func If_CutTruthWordSlice(p *If_Man_t, pCut *If_Cut_t) []word {
	If_CutTruthW(p, pCut)
	n := int(p.nTruth6Words[pCut.nLeaves])
	return p.puTempW[:n]
}

func If_ManCutMin(p *If_Man_t) int32  { return p.pPars.fCutMin }
func If_ManLutSize(p *If_Man_t) int32 { return p.pPars.nLutSize }

func If_ParSetRec3Defaults(p *If_Par_t, vars, cuts, useSOPB int32, userFunc func(*If_Man_t, *If_Obj_t, *If_Cut_t) int32) {
	*p = If_Par_t{}
	p.nLutSize = vars
	p.nCutsMax = cuts
	p.DelayTarget = -1
	p.Epsilon = 0.005
	p.fArea = 1
	p.fTruth = 1
	if useSOPB != 0 {
		p.fCutMin, p.fUsePerm, p.fDelayOpt = 0, 1, 1
	} else {
		p.fCutMin, p.fUsePerm, p.fDelayOpt = 1, 0, 0
	}
	p.fSkipCutFilter = 0
	p.pFuncCost = nil
	p.pFuncUser = userFunc
}

func If_ObjCopyLit(pObj *If_Obj_t) int32              { return pObj.iCopy }
func If_ObjSetCopyLit(pObj *If_Obj_t, lit int32)      { pObj.iCopy = lit }
func If_CutTruth(p *If_Man_t, pCut *If_Cut_t) *uint32 { return If_CutTruthUR(p, pCut) }

func If_CutDsdLit(p *If_Man_t, pCut *If_Cut_t) int32 {
	return Abc_Lit2LitL(Vec_IntArray(p.vTtDsds[pCut.nLeaves]), If_CutTruthLit(pCut))
}
func If_CutDsdIsCompl(p *If_Man_t, pCut *If_Cut_t) int32 {
	return Abc_LitIsCompl(If_CutDsdLit(p, pCut))
}
func If_CutDsdPerm(p *If_Man_t, pCut *If_Cut_t) []byte {
	start := int(Abc_Lit2Var(pCut.iCutFunc) * Abc_MaxInt(6, int32(pCut.nLeaves)))
	count := int(Abc_MaxInt(6, int32(pCut.nLeaves)))
	data := vec.Vec_StrArray(p.vTtPerms[pCut.nLeaves])
	assert(start >= 0 && start+count <= len(data), "DSD permutation range is allocated")
	return data[start : start+count]
}
func If_CutLutArea(p *If_Man_t, pCut *If_Cut_t) float32 {
	if pCut.fAndCut != 0 {
		return float32(p.pPars.nAndArea)
	}
	if pCut.fUser != 0 {
		return float32(pCut.Cost)
	}
	if p.pPars.pLutLib != nil {
		return p.pPars.pLutLib.pLutAreas[pCut.nLeaves]
	}
	return 1
}
func If_CutLutDelay(p *If_LibLut_t, size, pin int32) float32 {
	if p == nil {
		return 1
	}
	if p.fVarPinDelays != 0 {
		return p.pLutDelays[size][pin]
	}
	return p.pLutDelays[size][0]
}

////////////////////////////////////////////////////////////////////////
///                      MACRO DEFINITIONS                           ///
////////////////////////////////////////////////////////////////////////

// Original macro:
// #define IF_MIN(a,b)      (((a) < (b))? (a) : (b))

// Original macro:
// #define IF_MAX(a,b)      (((a) > (b))? (a) : (b))

// the small and large numbers (min/max float are 1.17e-38/3.40e+38)

// iterator over the primary inputs
// Original macro:
// #define If_ManForEachCi( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( If_Obj_t *, p->vCis, pObj, i )

// iterator over the primary outputs
// Original macro:
// #define If_ManForEachCo( p, pObj, i )                                          \
//     Vec_PtrForEachEntry( If_Obj_t *, p->vCos, pObj, i )

// iterator over the primary inputs
// Original macro:
// #define If_ManForEachPi( p, pObj, i )                                          \
//     Vec_PtrForEachEntryStop( If_Obj_t *, p->vCis, pObj, i, If_ManCiNum(p) - p->pPars->nLatchesCi - p->pPars->nLatchesCiBox )

// iterator over the primary outputs
// Original macro:
// #define If_ManForEachPo( p, pObj, i )                                          \
//     Vec_PtrForEachEntryStartStop( If_Obj_t *, p->vCos, pObj, i, p->pPars->nLatchesCoBox, If_ManCoNum(p) - p->pPars->nLatchesCo )

// iterator over the latches
// Original macro:
// #define If_ManForEachLatchInput( p, pObj, i )                                  \
//     Vec_PtrForEachEntryStart( If_Obj_t *, p->vCos, pObj, i, If_ManCoNum(p) - p->pPars->nLatchesCo )

// Original macro:
// #define If_ManForEachLatchOutput( p, pObj, i )                                 \
//     Vec_PtrForEachEntryStartStop( If_Obj_t *, p->vCis, pObj, i, If_ManCiNum(p) - p->pPars->nLatchesCi - p->pPars->nLatchesCiBox, If_ManCiNum(p) - p->pPars->nLatchesCiBox )

// iterator over all objects in topological order
// Original macro:
// #define If_ManForEachObj( p, pObj, i )                                         \
//     Vec_PtrForEachEntry( If_Obj_t *, p->vObjs, pObj, i )

// iterator over all objects in reverse topological order
// Original macro:
// #define If_ManForEachObjReverse( p, pObj, i )                                  \
//     Vec_PtrForEachEntry( If_Obj_t *, p->vObjsRev, pObj, i )

// iterator over logic nodes
// Original macro:
// #define If_ManForEachNode( p, pObj, i )                                        \
//     If_ManForEachObj( p, pObj, i ) if ( pObj->Type != IF_AND ) {} else

// iterator over cuts of the node
// Original macro:
// #define If_ObjForEachCut( pObj, pCut, i )                                      \
//     for ( i = 0; (i < (pObj)->pCutSet->nCuts) && ((pCut) = (pObj)->pCutSet->ppCuts[i]); i++ )

// iterator over the leaves of the cut
// Original macro:
// #define If_CutForEachLeaf( p, pCut, pLeaf, i )                                 \
//     for ( i = 0; (i < (int)(pCut)->nLeaves) && ((pLeaf) = If_ManObj(p, (pCut)->pLeaves[i])); i++ )

// Original macro:
// #define If_CutForEachLeafReverse( p, pCut, pLeaf, i )                                 \
//     for ( i = (int)(pCut)->nLeaves - 1; (i >= 0) && ((pLeaf) = If_ManObj(p, (pCut)->pLeaves[i])); i-- )

//#define If_CutForEachLeaf( p, pCut, pLeaf, i )                                 \ \\prevent multiline comment
//    for ( i = 0; (i < (int)(pCut)->nLeaves) && ((pLeaf) = If_ManObj(p, p->pPars->fLiftLeaves? (pCut)->pLeaves[i] >> 8 : (pCut)->pLeaves[i])); i++ )
// iterator over the leaves of the sequential cut
// Original macro:
// #define If_CutForEachLeafSeq( p, pCut, pLeaf, Shift, i )                       \
//     for ( i = 0; (i < (int)(pCut)->nLeaves) && ((pLeaf) = If_ManObj(p, (pCut)->pLeaves[i] >> 8)) && (((Shift) = ((pCut)->pLeaves[i] & 255)) >= 0); i++ )

////////////////////////////////////////////////////////////////////////
///                    FUNCTION DECLARATIONS                         ///
////////////////////////////////////////////////////////////////////////

/*=== ifCore.c ===========================================================*/
// extern void            If_ManSetDefaultPars( If_Par_t * pPars );

// extern int             If_ManPerformMapping( If_Man_t * p );

// extern int             If_ManPerformMappingComb( If_Man_t * p );

// extern void            If_ManComputeSwitching( If_Man_t * p );

/*=== ifCut.c ============================================================*/
// extern int             If_CutVerifyCuts( If_Set_t * pCutSet, int fOrdered );

// extern int             If_CutFilter( If_Set_t * pCutSet, If_Cut_t * pCut, int fSaveCut0 );

// extern void            If_CutSort( If_Man_t * p, If_Set_t * pCutSet, If_Cut_t * pCut );

// extern void            If_CutOrder( If_Cut_t * pCut );

// extern int             If_CutMergeOrdered( If_Man_t * p, If_Cut_t * pCut0, If_Cut_t * pCut1, If_Cut_t * pCut );

// extern int             If_CutMerge( If_Man_t * p, If_Cut_t * pCut0, If_Cut_t * pCut1, If_Cut_t * pCut );

// extern int             If_CutCheck( If_Cut_t * pCut );

// extern void            If_CutPrint( If_Cut_t * pCut );

// extern void            If_CutPrintTiming( If_Man_t * p, If_Cut_t * pCut );

// extern void            If_CutLift( If_Cut_t * pCut );

// extern void            If_CutCopy( If_Man_t * p, If_Cut_t * pCutDest, If_Cut_t * pCutSrc );

// extern float           If_CutAreaFlow( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutEdgeFlow( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutPowerFlow( If_Man_t * p, If_Cut_t * pCut, If_Obj_t * pRoot );

// extern float           If_CutAverageRefs( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutAreaDeref( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutAreaRef( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutAreaDerefed( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutAreaRefed( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutEdgeDeref( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutEdgeRef( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutEdgeDerefed( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutEdgeRefed( If_Man_t * p, If_Cut_t * pCut );

// extern float           If_CutPowerDeref( If_Man_t * p, If_Cut_t * pCut, If_Obj_t * pRoot );

// extern float           If_CutPowerRef( If_Man_t * p, If_Cut_t * pCut, If_Obj_t * pRoot );

// extern float           If_CutPowerDerefed( If_Man_t * p, If_Cut_t * pCut, If_Obj_t * pRoot );

// extern float           If_CutPowerRefed( If_Man_t * p, If_Cut_t * pCut, If_Obj_t * pRoot );

/*=== ifDec.c =============================================================*/
// extern word            If_CutPerformDerive07( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck07( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck08( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck10( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck16( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck45( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck54( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern int             If_CutPerformCheck75( If_Man_t * p, unsigned * pTruth, int nVars, int nLeaves, char * pStr );

// extern float           If_CutDelayLutStruct( If_Man_t * p, If_Cut_t * pCut, char * pStr, float WireDelay );

// extern int             If_CluCheckExt( void * p, word * pTruth, int nVars, int nLutLeaf, int nLutRoot,
//                            char * pLut0, char * pLut1, word * pFunc0, word * pFunc1 );

// extern int             If_CluCheckExt3( void * p, word * pTruth, int nVars, int nLutLeaf, int nLutLeaf2, int nLutRoot,
//                            char * pLut0, char * pLut1, char * pLut2, word * pFunc0, word * pFunc1, word * pFunc2 );

/*=== ifDelay.c =============================================================*/
// extern int             If_CutDelaySop( If_Man_t * p, If_Cut_t * pCut );

// extern int             If_CutSopBalanceEvalInt( Vec_Int_t * vCover, int * pTimes, int * pFaninLits, Vec_Int_t * vAig, int * piRes, int nSuppAll, int * pArea );

// extern int             If_CutSopBalanceEval( If_Man_t * p, If_Cut_t * pCut, Vec_Int_t * vAig );

// extern int             If_CutSopBalancePinDelaysInt( Vec_Int_t * vCover, int * pTimes, word * pFaninRes, int nSuppAll, word * pRes );

// extern int             If_CutSopBalancePinDelays( If_Man_t * p, If_Cut_t * pCut, char * pPerm );

// extern int             If_CutLutBalanceEval( If_Man_t * p, If_Cut_t * pCut );

// extern int             If_CutLutBalancePinDelays( If_Man_t * p, If_Cut_t * pCut, char * pPerm );

/*=== ifDsd.c =============================================================*/
// extern If_DsdMan_t *   If_DsdManAlloc( int nVars, int nLutSize );

// extern void            If_DsdManAllocIsops( If_DsdMan_t * p, int nLutSize );

// extern void            If_DsdManPrint( If_DsdMan_t * p, char * pFileName, int Number, int Support, int fOccurs, int fTtDump, int fVerbose );

// extern void            If_DsdManTune( If_DsdMan_t * p, int LutSize, int fFast, int fAdd, int fSpec, int fVerbose );

// extern void            Id_DsdManTuneStr( If_DsdMan_t * p, char * pStruct, int nConfls, int nProcs, int fVerbose );

// extern void            If_DsdManFree( If_DsdMan_t * p, int fVerbose );

// extern void            If_DsdManSave( If_DsdMan_t * p, char * pFileName );

// extern If_DsdMan_t *   If_DsdManLoad( char * pFileName );

// extern void            If_DsdManMerge( If_DsdMan_t * p, If_DsdMan_t * pNew );

// extern void            If_DsdManCleanOccur( If_DsdMan_t * p, int fVerbose );

// extern void            If_DsdManCleanMarks( If_DsdMan_t * p, int fVerbose );

// extern void            If_DsdManInvertMarks( If_DsdMan_t * p, int fVerbose );

// extern If_DsdMan_t *   If_DsdManFilter( If_DsdMan_t * p, int Limit );

// extern int             If_DsdManCompute( If_DsdMan_t * p, word * pTruth, int nLeaves, unsigned char * pPerm, char * pLutStruct );

// extern char *          If_DsdManFileName( If_DsdMan_t * p );

// extern int             If_DsdManVarNum( If_DsdMan_t * p );

// extern int             If_DsdManObjNum( If_DsdMan_t * p );

// extern int             If_DsdManLutSize( If_DsdMan_t * p );

// extern int             If_DsdManTtBitNum( If_DsdMan_t * p );

// extern int             If_DsdManPermBitNum( If_DsdMan_t * p );

// extern void            If_DsdManSetLutSize( If_DsdMan_t * p, int nLutSize );

// extern int             If_DsdManSuppSize( If_DsdMan_t * p, int iDsd );

// extern int             If_DsdManCheckDec( If_DsdMan_t * p, int iDsd );

// extern int             If_DsdManReadMark( If_DsdMan_t * p, int iDsd );

// extern void            If_DsdManSetNewAsUseless( If_DsdMan_t * p );

// extern word *          If_DsdManGetFuncConfig( If_DsdMan_t * p, int iDsd );

// extern char *          If_DsdManGetCellStr( If_DsdMan_t * p );

// extern unsigned        If_DsdManCheckXY( If_DsdMan_t * p, int iDsd, int LutSize, int fDerive, unsigned uMaskNot, int fHighEffort, int fVerbose );

// extern int             If_CutDsdBalanceEval( If_Man_t * p, If_Cut_t * pCut, Vec_Int_t * vAig );

// extern int             If_CutDsdBalancePinDelays( If_Man_t * p, If_Cut_t * pCut, char * pPerm );

// extern void            Id_DsdManTuneThresh( If_DsdMan_t * p, int fUnate, int fThresh, int fThreshHeuristic, int fVerbose );

/*=== ifLib.c =============================================================*/
// extern If_LibLut_t *   If_LibLutRead( char * FileName );

// extern If_LibLut_t *   If_LibLutDup( If_LibLut_t * p );

// extern void            If_LibLutFree( If_LibLut_t * pLutLib );

// extern void            If_LibLutPrint( If_LibLut_t * pLutLib );

// extern int             If_LibLutDelaysAreDiscrete( If_LibLut_t * pLutLib );

// extern int             If_LibLutDelaysAreDifferent( If_LibLut_t * pLutLib );

// extern If_LibLut_t *   If_LibLutSetSimple( int nLutSize );

// extern float           If_LibLutFastestPinDelay( If_LibLut_t * p );

// extern float           If_LibLutSlowestPinDelay( If_LibLut_t * p );

/*=== ifLibBox.c =============================================================*/
// extern If_LibBox_t *   If_LibBoxStart();

// extern void            If_LibBoxFree( If_LibBox_t * p );

// extern int             If_LibBoxNum( If_LibBox_t * p );

// extern If_Box_t *      If_LibBoxReadBox( If_LibBox_t * p, int Id );

// extern If_Box_t *      If_LibBoxFindBox( If_LibBox_t * p, char * pName );

// extern void            If_LibBoxAdd( If_LibBox_t * p, If_Box_t * pBox );

// extern If_LibBox_t *   If_LibBoxRead( char * pFileName );

// extern If_LibBox_t *   If_LibBoxRead2( char * pFileName );

// extern void            If_LibBoxPrint( FILE * pFile, If_LibBox_t * p );

// extern void            If_LibBoxWrite( char * pFileName, If_LibBox_t * p );

// extern int             If_LibBoxLoad( char * pFileName );

// extern If_Box_t *      If_BoxStart( char * pName, int Id, int nPis, int nPos, int fSeq, int fBlack, int fOuter );

/*=== ifMan.c =============================================================*/
// extern If_Man_t *      If_ManStart( If_Par_t * pPars );

// extern void            If_ManRestart( If_Man_t * p );

// extern void            If_ManStop( If_Man_t * p );

// extern If_Obj_t *      If_ManCreateCi( If_Man_t * p );

// extern If_Obj_t *      If_ManCreateCo( If_Man_t * p, If_Obj_t * pDriver );

// extern If_Obj_t *      If_ManCreateAnd( If_Man_t * p, If_Obj_t * pFan0, If_Obj_t * pFan1 );

// extern If_Obj_t *      If_ManCreateXor( If_Man_t * p, If_Obj_t * pFan0, If_Obj_t * pFan1 );

// extern If_Obj_t *      If_ManCreateMux( If_Man_t * p, If_Obj_t * pFan0, If_Obj_t * pFan1, If_Obj_t * pCtrl );

// extern void            If_ManCreateChoice( If_Man_t * p, If_Obj_t * pRepr );

// extern void            If_ManSetupCutTriv( If_Man_t * p, If_Cut_t * pCut, int ObjId );

// extern void            If_ManSetupCiCutSets( If_Man_t * p );

// extern If_Set_t *      If_ManSetupNodeCutSet( If_Man_t * p, If_Obj_t * pObj );

// extern void            If_ManDerefNodeCutSet( If_Man_t * p, If_Obj_t * pObj );

// extern void            If_ManDerefChoiceCutSet( If_Man_t * p, If_Obj_t * pObj );

// extern void            If_ManSetupSetAll( If_Man_t * p, int nCrossCut );

/*=== ifMap.c =============================================================*/
// extern int *           If_CutArrTimeProfile( If_Man_t * p, If_Cut_t * pCut );

// extern void            If_ObjPerformMappingAnd( If_Man_t * p, If_Obj_t * pObj, int Mode, int fPreprocess, int fFirst );

// extern void            If_ObjPerformMappingChoice( If_Man_t * p, If_Obj_t * pObj, int Mode, int fPreprocess );

// extern int             If_ManPerformMappingRound( If_Man_t * p, int nCutsUsed, int Mode, int fPreprocess, int fFirst, char * pLabel );

/*=== ifReduce.c ==========================================================*/
// extern void            If_ManImproveMapping( If_Man_t * p );

/*=== ifSat.c ==========================================================*/
// extern void *          If_ManSatBuildXY( int nLutSize );

// extern void *          If_ManSatBuildXYZ( int nLutSize );

// extern void            If_ManSatUnbuild( void * p );

// extern int             If_ManSatCheckXY( void * pSat, int nLutSize, word * pTruth, int nVars, unsigned uSet, word * pTBound, word * pTFree, Vec_Int_t * vLits );

// extern unsigned        If_ManSatCheckXYall( void * pSat, int nLutSize, word * pTruth, int nVars, Vec_Int_t * vLits );

/*=== ifSeq.c =============================================================*/
// extern int             If_ManPerformMappingSeq( If_Man_t * p );

/*=== ifTime.c ============================================================*/
// extern float           If_CutDelay( If_Man_t * p, If_Obj_t * pObj, If_Cut_t * pCut );

// extern void            If_CutPropagateRequired( If_Man_t * p, If_Obj_t * pObj, If_Cut_t * pCut, float Required );

// extern float           If_ManDelayMax( If_Man_t * p, int fSeq );

// extern void            If_ManComputeRequired( If_Man_t * p );

/*=== ifTruth.c ===========================================================*/
// extern void            If_CutRotatePins( If_Man_t * p, If_Cut_t * pCut );

// extern int             If_CutComputeTruth( If_Man_t * p, If_Cut_t * pCut, If_Cut_t * pCut0, If_Cut_t * pCut1, int fCompl0, int fCompl1 );

// extern int             If_CutComputeTruthPerm( If_Man_t * p, If_Cut_t * pCut, If_Cut_t * pCut0, If_Cut_t * pCut1, int fCompl0, int fCompl1 );

// extern Vec_Mem_t *     If_DeriveHashTable6( int nVars, word Truth );

// extern int             If_CutCheckTruth6( If_Man_t * p, If_Cut_t * pCut );

/*=== ifTune.c ===========================================================*/
// extern Ifn_Ntk_t *     Ifn_NtkParse( char * pStr );

// extern int             Ifn_NtkTtBits( char * pStr );

// extern int             Ifn_NtkMatch( Ifn_Ntk_t * p, word * pTruth, int nVars, int nConfls, int fVerbose, int fVeryVerbose, word * pConfig );

// extern void            Ifn_NtkPrint( Ifn_Ntk_t * p );

// extern int             Ifn_NtkLutSizeMax( Ifn_Ntk_t * p );

// extern int             Ifn_NtkInputNum( Ifn_Ntk_t * p );

// extern void *          If_ManSatBuildFromCell( char * pStr, Vec_Int_t ** pvPiVars, Vec_Int_t ** pvPoVars, Ifn_Ntk_t ** ppNtk );

// extern int             If_ManSatFindCofigBits( void * pSat, Vec_Int_t * vPiVars, Vec_Int_t * vPoVars, word * pTruth, int nVars, word Perm, int nInps, Vec_Int_t * vValues );

// extern int             If_ManSatDeriveGiaFromBits( void * pNew, Ifn_Ntk_t * p, word * pTtData, Vec_Int_t * vLeaves, Vec_Int_t * vValues );

// extern void *          If_ManDeriveGiaFromCells( void * p );

/*=== ifUtil.c ============================================================*/
// extern void            If_ManCleanNodeCopy( If_Man_t * p );

// extern void            If_ManCleanCutData( If_Man_t * p );

// extern void            If_ManCleanMarkV( If_Man_t * p );

// extern float           If_ManScanMapping( If_Man_t * p );

// extern float           If_ManScanMappingDirect( If_Man_t * p );

// extern float           If_ManScanMappingSeq( If_Man_t * p );

// extern void            If_ManResetOriginalRefs( If_Man_t * p );

// extern int             If_ManCrossCut( If_Man_t * p );

// extern Vec_Ptr_t *     If_ManReverseOrder( If_Man_t * p );

// extern void            If_ManMarkMapping( If_Man_t * p );

// extern Vec_Ptr_t *     If_ManCollectMappingDirect( If_Man_t * p );

// extern Vec_Int_t *     If_ManCollectMappingInt( If_Man_t * p );

// extern int             If_ManCountSpecialPos( If_Man_t * p );

// extern void            If_CutTraverse( If_Man_t * p, If_Obj_t * pRoot, If_Cut_t * pCut, Vec_Ptr_t * vNodes );

// extern void            If_ObjPrint( If_Obj_t * pObj );

//ABC_NAMESPACE_HEADER_END

////////////////////////////////////////////////////////////////////////
///                       END OF FILE                                ///
////////////////////////////////////////////////////////////////////////
