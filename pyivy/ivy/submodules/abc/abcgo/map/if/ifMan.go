package if_

import (
	"fmt"
	"unsafe"

	"github.com/glycerine/abcgo/misc/tim"
	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// If_ManCutSetRecycle returns a no-longer-live cut set to the manager's pool.
func If_ManCutSetRecycle(p *If_Man_t, set *If_Set_t) {
	if set == nil {
		return
	}
	set.nCuts = 0
	set.pNext = p.pFreeList
	p.pFreeList = set
}

// If_ManCutSetFetch obtains one cut set from the manager's pool.
func If_ManCutSetFetch(p *If_Man_t) *If_Set_t {
	set := p.pFreeList
	if set == nil {
		panic("if: cut-set pool exhausted")
	}
	p.pFreeList = set.pNext
	set.pNext = nil
	return set
}

// If_ManStart initializes the graph manager and its enabled truth caches.
func If_ManStart(pars *If_Par_t) *If_Man_t {
	assert(pars != nil, "pars != nil")
	assert(pars.fUseDsd == 0 || pars.fUseTtPerm == 0, "pars.fUseDsd == 0 || pars.fUseTtPerm == 0")
	assert(pars.nLutSize > 0 && pars.nLutSize <= IF_MAX_LUTSIZE, "valid LUT size")
	if pars.fTruth != 0 {
		assert(pars.nLutSize <= IF_MAX_FUNC_LUTSIZE, "truth LUT size is within the supported truth-cache bound")
	}
	p := &If_Man_t_{pPars: pars, fEpsilon: pars.Epsilon}
	p.vCis = vec.Vec_PtrAlloc(100)
	p.vCos = vec.Vec_PtrAlloc(100)
	p.vObjs = vec.Vec_PtrAlloc(100)
	p.vTemp = vec.Vec_PtrAlloc(100)
	p.vVisited = vec.Vec_PtrAlloc(100)

	if pars.fTruth != 0 {
		maxVars := int(Abc_MaxInt(6, pars.nLutSize))
		for v := 0; v <= maxVars; v++ {
			p.nTruth6Words[v] = abcutil.Abc_Truth6WordNum(int32(v))
		}
		for v := 6; v <= maxVars; v++ {
			p.vTtMem[v] = vec.Vec_MemAllocForTT(v, int(pars.fUseTtPerm))
		}
		for v := 0; v < 6; v++ {
			p.vTtMem[v] = p.vTtMem[6]
			p.nTruth6Words[v] = p.nTruth6Words[5]
		}
		if pars.fDelayOpt != 0 || pars.nGateSize > 0 {
			for v := 6; v <= maxVars; v++ {
				p.vTtIsops[v] = vec.Vec_WecAlloc(1000)
				vec.Vec_WecInit(p.vTtIsops[v], 2)
				p.vTtIsopComplement[v] = make([]bool, 2)
			}
			for v := 0; v < 6; v++ {
				p.vTtIsops[v] = p.vTtIsops[6]
			}
		}
		if pars.fDelayOpt != 0 || pars.nGateSize > 0 || pars.fDsdBalance != 0 {
			p.vCover = vec.Vec_IntAlloc(0)
			p.vArray = vec.Vec_IntAlloc(1000)
		}
	}
	if pars.fUsePerm != 0 {
		p.nPermWords = If_CutPermWords(pars.nLutSize)
	}
	p.nObjBytes = int32(unsafe.Sizeof(If_Obj_t_{}))
	p.nCutBytes = int32(unsafe.Sizeof(If_Cut_t_{})) + 4*(pars.nLutSize+p.nPermWords)
	p.nSetBytes = int32(unsafe.Sizeof(If_Set_t_{})) + (int32(unsafe.Sizeof((*If_Cut_t)(nil)))+p.nCutBytes)*(pars.nCutsMax+1)

	if pars.fTruth != 0 {
		wordCount := int(p.nTruth6Words[pars.nLutSize])
		p.puTemp[0] = make([]uint32, 4*wordCount*2)
		for i := 1; i < 4; i++ {
			p.puTemp[i] = p.puTemp[0][i*wordCount*2:]
		}
		p.puTempW = make([]uint64, wordCount)
	}
	if pars.fUseDsd != 0 {
		for v := 6; v <= int(Abc_MaxInt(6, pars.nLutSize)); v++ {
			p.vTtDsds[v] = vec.Vec_IntAlloc(1000)
			vec.Vec_IntPush(p.vTtDsds[v], 0)
			vec.Vec_IntPush(p.vTtDsds[v], 2)
			p.vTtPerms[v] = vec.Vec_StrAlloc(10000)
			vec.Vec_StrFill(p.vTtPerms[v], 2*v, IF_BIG_CHAR)
			vec.Vec_StrWriteEntry(p.vTtPerms[v], v, 0)
		}
		for v := 0; v < 6; v++ {
			p.vTtDsds[v], p.vTtPerms[v] = p.vTtDsds[6], p.vTtPerms[6]
		}
	}
	if pars.fUseTtPerm != 0 {
		p.vPairHash = vec.Hash_IntManStart(10000)
		p.vPairPerms = vec.Vec_StrAlloc(10000)
		vec.Vec_StrFill(p.vPairPerms, int(pars.nLutSize), 0)
		p.vPairRes = vec.Vec_IntAlloc(1000)
		vec.Vec_IntPush(p.vPairRes, -1)
		for v := 6; v <= int(Abc_MaxInt(6, pars.nLutSize)); v++ {
			p.vTtOccurs[v] = vec.Vec_IntAlloc(1000)
			vec.Vec_IntPushTwo(p.vTtOccurs[v], 0, 0)
		}
		for v := 0; v < 6; v++ {
			p.vTtOccurs[v] = p.vTtOccurs[6]
		}
	}
	if pars.fUseCofVars != 0 {
		for v := 6; v <= int(Abc_MaxInt(6, pars.nLutSize)); v++ {
			p.vTtVars[v] = vec.Vec_StrAlloc(1000)
			vec.Vec_StrPush(p.vTtVars[v], 0)
			vec.Vec_StrPush(p.vTtVars[v], 0)
		}
		for v := 0; v < 6; v++ {
			p.vTtVars[v] = p.vTtVars[6]
		}
	}
	if pars.fUseAndVars != 0 {
		for v := 6; v <= int(Abc_MaxInt(6, pars.nLutSize)); v++ {
			p.vTtDecs[v] = vec.Vec_IntAlloc(1000)
			vec.Vec_IntPushTwo(p.vTtDecs[v], 0, 0)
		}
		for v := 0; v < 6; v++ {
			p.vTtDecs[v] = p.vTtDecs[6]
		}
	}
	if pars.fVerbose != 0 {
		fmt.Printf("K = %d. Cut record bytes = %d. Object bytes = %d. Set bytes = %d.\n",
			pars.nLutSize, p.nCutBytes, p.nObjBytes, p.nSetBytes)
	}

	p.pConst1 = If_ManSetupObj(p)
	p.pConst1.Type = IF_CONST1
	p.pConst1.fPhase = 1
	p.nObjs[IF_CONST1]++
	return p
}

// If_ManRestart clears the graph while retaining reusable manager caches.
func If_ManRestart(p *If_Man_t) {
	assert(p != nil && !p.stopped, "manager is live")
	vec.Vec_PtrClear(p.vCis)
	vec.Vec_PtrClear(p.vCos)
	vec.Vec_PtrClear(p.vObjs)
	vec.Vec_PtrClear(p.vTemp)
	vec.Vec_PtrClear(p.vVisited)
	p.ciCutSets = nil
	p.pMemCi = nil
	p.nObjs = [IF_VOID + 1]int32{}
	p.nLevelMax, p.nChoices = 0, 0
	p.pConst1 = If_ManSetupObj(p)
	p.pConst1.Type = IF_CONST1
	p.pConst1.fPhase = 1
	p.nObjs[IF_CONST1]++
}

// If_ManStop releases Go-owned graph and cache references; it is safe to repeat.
func If_ManStop(p *If_Man_t) {
	if p == nil || p.stopped {
		return
	}
	p.stopped = true
	if p.pPars != nil && p.pPars.fVerbose != 0 && p.vCutData != nil {
		If_ManCacheAnalize(p)
	}
	for v := 6; v <= int(Abc_MaxInt(6, p.pPars.nLutSize)); v++ {
		vec.Vec_MemFree(p.vTtMem[v])
		vec.Vec_WecFree(p.vTtIsops[v])
		vec.Vec_IntFree(p.vTtDsds[v])
		vec.Vec_StrFree(p.vTtPerms[v])
		vec.Vec_StrFree(p.vTtVars[v])
		vec.Vec_IntFree(p.vTtDecs[v])
		vec.Vec_IntFree(p.vTtOccurs[v])
	}
	vec.Vec_IntFree(p.vCoAttrs)
	vec.Vec_PtrFree(p.vCis)
	vec.Vec_PtrFree(p.vCos)
	vec.Vec_PtrFree(p.vObjs)
	vec.Vec_PtrFree(p.vTemp)
	vec.Vec_IntFree(p.vCover)
	vec.Vec_IntFree(p.vArray)
	vec.Vec_WrdFree(p.vAnds)
	vec.Vec_WrdFree(p.vAndGate)
	vec.Vec_WrdFree(p.vOrGate)
	vec.Vec_PtrFree(p.vObjsRev)
	vec.Vec_PtrFree(p.vLatchOrder)
	vec.Vec_IntFree(p.vLags)
	vec.Vec_IntFree(p.vDump)
	vec.Vec_IntFree(p.vCutData)
	vec.Vec_IntFree(p.vPairRes)
	vec.Vec_StrFree(p.vPairPerms)
	vec.Vec_PtrFree(p.vVisited)
	if p.vPairHash != nil {
		vec.Hash_IntManStop(p.vPairHash)
	}
	if p.vTtMem6 != nil {
		vec.Vec_MemFree(p.vTtMem6)
	}
	if p.pManTim != nil {
		tim.Tim_ManStop(p.pManTim)
	}
	vec.Vec_IntFree(p.vSwitching)
	for i := range p.pHashTable {
		p.pHashTable[i] = nil
	}
	p.pMemObj, p.pMemSet, p.pMemEntries = nil, nil, nil
	p.pMemCi, p.pMemAnd, p.pFreeList = nil, nil, nil
	p.puTemp, p.puTempW, p.pTruth32 = [4][]uint32{}, nil, nil
	p.vTtMem, p.vTtIsops = [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Mem_t{}, [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Wec_t{}
	p.vTtDsds, p.vTtDecs, p.vTtOccurs = [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Int_t{}, [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Int_t{}, [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Int_t{}
	p.vTtPerms, p.vTtVars = [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Str_t{}, [IF_MAX_FUNC_LUTSIZE + 1]*Vec_Str_t{}
	p.pPars = nil
	p.pConst1 = nil
	p.ciCutSets, p.ownedCutSets = nil, nil
}

// If_ManCreateCi appends one primary input.
func If_ManCreateCi(p *If_Man_t) *If_Obj_t {
	obj := If_ManSetupObj(p)
	obj.Type = IF_CI
	obj.IdPio = int32(vec.Vec_PtrSize(p.vCis))
	vec.Vec_PtrPush(p.vCis, obj)
	p.nObjs[IF_CI]++
	return obj
}

// If_ManCreateCo appends an output whose driver may be complemented.
func If_ManCreateCo(p *If_Man_t, driver *If_Obj_t) *If_Obj_t {
	obj := If_ManSetupObj(p)
	obj.IdPio = int32(vec.Vec_PtrSize(p.vCos))
	vec.Vec_PtrPush(p.vCos, obj)
	obj.Type = IF_CO
	obj.fCompl0 = uint32(If_IsComplement(driver))
	driver = If_Regular(driver)
	obj.pFanin0 = driver
	driver.nRefs++
	obj.fPhase = obj.fCompl0 ^ driver.fPhase
	obj.Level = driver.Level
	if p.nLevelMax < int32(obj.Level) {
		p.nLevelMax = int32(obj.Level)
	}
	p.nObjs[IF_CO]++
	return obj
}

// If_ManCreateAnd constructs an AND node with constant and duplicate folding.
func If_ManCreateAnd(p *If_Man_t, fan0, fan1 *If_Obj_t) *If_Obj_t {
	if fan0 == fan1 {
		return fan0
	}
	if fan0 == If_Not(fan1) {
		return If_Not(p.pConst1)
	}
	if If_Regular(fan0) == p.pConst1 {
		if fan0 == p.pConst1 {
			return fan1
		}
		return If_Not(p.pConst1)
	}
	if If_Regular(fan1) == p.pConst1 {
		if fan1 == p.pConst1 {
			return fan0
		}
		return If_Not(p.pConst1)
	}
	obj := If_ManSetupObj(p)
	obj.Type = IF_AND
	obj.fCompl0 = uint32(If_IsComplement(fan0))
	obj.fCompl1 = uint32(If_IsComplement(fan1))
	fan0, fan1 = If_Regular(fan0), If_Regular(fan1)
	obj.pFanin0, obj.pFanin1 = fan0, fan1
	for _, fanin := range []*If_Obj_t{fan0, fan1} {
		fanin.nRefs++
		fanin.nVisits++
		fanin.nVisitsCopy++
	}
	obj.fPhase = (obj.fCompl0 ^ fan0.fPhase) & (obj.fCompl1 ^ fan1.fPhase)
	obj.Level = 1 + uint32(max(int(fan0.Level), int(fan1.Level)))
	if int32(obj.Level) > p.nLevelMax {
		p.nLevelMax = int32(obj.Level)
	}
	p.nObjs[IF_AND]++
	return obj
}

func If_ManCreateXor(p *If_Man_t, fan0, fan1 *If_Obj_t) *If_Obj_t {
	a := If_ManCreateAnd(p, If_Not(fan0), fan1)
	b := If_ManCreateAnd(p, fan0, If_Not(fan1))
	return If_Not(If_ManCreateAnd(p, If_Not(a), If_Not(b)))
}

func If_ManCreateMux(p *If_Man_t, fan0, fan1, control *If_Obj_t) *If_Obj_t {
	a := If_ManCreateAnd(p, fan0, If_Not(control))
	b := If_ManCreateAnd(p, fan1, control)
	return If_Not(If_ManCreateAnd(p, If_Not(a), If_Not(b)))
}

// If_ManCreateChoice marks a linked equivalent-node class as one mapping choice.
func If_ManCreateChoice(p *If_Man_t, obj *If_Obj_t) {
	assert(obj.fRepr == 0, "obj.fRepr == 0")
	obj.fRepr = 1
	for choice := obj; choice != nil; choice = choice.pEquiv {
		if choice.Level > obj.Level {
			obj.Level = choice.Level
		}
		choice.nVisits++
		choice.nVisitsCopy++
	}
	if int32(obj.Level) > p.nLevelMax {
		p.nLevelMax = int32(obj.Level)
	}
	p.nChoices++
}

func If_ManSetupSet(p *If_Man_t, set *If_Set_t) {
	set.nCuts = 0
	set.nCutsMax = int16(p.pPars.nCutsMax)
	set.ppCuts = make([]*If_Cut_t, int(set.nCutsMax)+1)
	for i := range set.ppCuts {
		cut := &If_Cut_t_{}
		If_CutSetup(p, cut)
		set.ppCuts[i] = cut
	}
}

func If_ManSetupCutTriv(p *If_Man_t, cut *If_Cut_t, objID int32) {
	cut.fCompl = 0
	cut.nLimit = uint32(p.pPars.nLutSize)
	cut.nLeaves = 1
	if p.pPars.fLiftLeaves != 0 {
		objID <<= 8
	}
	cut.pLeaves = []int32{objID}
	cut.uSign = If_ObjCutSign(uint32(objID))
	if p.pPars.fUseTtPerm != 0 {
		cut.iCutFunc = 3
	} else if p.pPars.fTruth != 0 {
		cut.iCutFunc = 2
	} else {
		cut.iCutFunc = -1
	}
	cut.uMaskFunc = 0
	assert(objID >= 0 && int(objID) < vec.Vec_PtrSize(p.vObjs), "valid trivial cut object")
}

func If_ManSetupObj(p *If_Man_t) *If_Obj_t {
	obj := &If_Obj_t_{}
	If_CutSetup(p, &obj.CutBest)
	obj.Id = int32(vec.Vec_PtrSize(p.vObjs))
	obj.Required = IF_FLOAT_LARGE
	vec.Vec_PtrPush(p.vObjs, obj)
	return obj
}

func If_ManSetupCiCutSets(p *If_Man_t) {
	assert(p.pMemCi == nil, "p.pMemCi == nil")
	p.ciCutSets = make([]*If_Set_t, vec.Vec_PtrSize(p.vCis))
	for i, entry := range vec.Vec_PtrArray(p.vCis) {
		obj := entry.(*If_Obj_t)
		If_ManSetupCutTriv(p, &obj.CutBest, obj.Id)
		set := &If_Set_t{nCuts: 1, nCutsMax: int16(p.pPars.nCutsMax)}
		set.ppCuts = make([]*If_Cut_t, int(set.nCutsMax)+1)
		set.ppCuts[0] = &obj.CutBest
		p.ciCutSets[i] = set
		obj.pCutSet = set
	}
	if len(p.ciCutSets) != 0 {
		p.pMemCi = p.ciCutSets[0]
	}
}

func If_ManSetupNodeCutSet(p *If_Man_t, obj *If_Obj_t) *If_Set_t {
	assert(If_ObjIsAnd(obj) != 0 && obj.pCutSet == nil, "AND node has no live cut set")
	set := If_ManCutSetFetch(p)
	set.nCuts = 0
	set.nCutsMax = int16(p.pPars.nCutsMax)
	obj.pCutSet = set
	return set
}

func If_ManDerefNodeCutSet(p *If_Man_t, obj *If_Obj_t) {
	assert(If_ObjIsAnd(obj) != 0, "If_ObjIsAnd(obj)")
	assert(obj.nVisits >= 0, "obj.nVisits >= 0")
	if obj.nVisits == 0 {
		If_ManCutSetRecycle(p, obj.pCutSet)
		obj.pCutSet = nil
	}
	for _, fanin := range []*If_Obj_t{obj.pFanin0, obj.pFanin1} {
		assert(fanin.nVisits > 0, "fanin.nVisits > 0")
		if If_ObjIsCi(fanin) == 0 {
			fanin.nVisits--
			if fanin.nVisits == 0 {
				If_ManCutSetRecycle(p, fanin.pCutSet)
				fanin.pCutSet = nil
			}
		}
	}
}

func If_ManDerefChoiceCutSet(p *If_Man_t, obj *If_Obj_t) {
	assert(If_ObjIsAnd(obj) != 0 && obj.fRepr != 0 && obj.nVisits > 0, "valid representative choice node")
	for choice := obj; choice != nil; choice = choice.pEquiv {
		choice.nVisits--
		if choice.nVisits == 0 {
			If_ManCutSetRecycle(p, choice.pCutSet)
			choice.pCutSet = nil
		}
	}
}

func If_ManSetupSetAll(p *If_Man_t, crossCut int32) {
	assert(crossCut >= 0, "crossCut >= 0")
	n := int64(128) + int64(crossCut)
	if n > int64(^uint32(0)>>1) {
		panic("if: cut-set pool size overflows int32")
	}
	p.ownedCutSets = make([]*If_Set_t, int(n))
	for i := range p.ownedCutSets {
		set := &If_Set_t{}
		If_ManSetupSet(p, set)
		if i+1 < len(p.ownedCutSets) {
			set.pNext = nil
		}
		p.ownedCutSets[i] = set
	}
	for i := 0; i+1 < len(p.ownedCutSets); i++ {
		p.ownedCutSets[i].pNext = p.ownedCutSets[i+1]
	}
	p.pFreeList = p.ownedCutSets[0]
	p.pMemAnd = p.pFreeList
	if p.pPars.fVerbose != 0 {
		fmt.Printf("Node = %d. Choice nodes = %d. Cut-set pool = %d.\n",
			If_ManAndNum(p), p.nChoices, len(p.ownedCutSets))
	}
}
