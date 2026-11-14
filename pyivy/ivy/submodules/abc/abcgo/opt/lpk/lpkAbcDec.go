package lpk

import (
	"time"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
)

// Lpk_ImplementFun builds the Hop logic for one decomposition component.
func Lpk_ImplementFun(pMan *Lpk_Man_t, pNtk *Abc_Ntk_t, leaves *Vec_Ptr_t, fun *Lpk_Fun_t) *Abc_Obj_t {
	if fun.fMark != 0 {
		pMan.nMuxes++
	} else {
		pMan.nDsds++
	}
	obj := abc.Abc_NtkCreateNode(pNtk)
	for i := 0; i < int(fun.nVars); i++ {
		abc.Abc_ObjAddFanin(obj, abc.Abc_ObjRegular(vec.Vec_PtrEntry(leaves, int(fun.pFanins[i]))))
	}
	abc.Abc_ObjSetLevel(obj, abcLevelNew(obj))
	truth := Lpk_FunTruth(fun, 0)
	hopMan, ok := pNtk.FunctionManager.(*hop.Hop_Man_t)
	if !ok || hopMan == nil {
		panic("lpk: network function manager is not a Hop manager")
	}
	switch fun.nVars {
	case 0:
		obj.Data = hop.Hop_NotCond(hop.Hop_ManConst1(hopMan), int32((truth[0]&1)^1))
	case 1:
		obj.Data = hop.Hop_NotCond(hop.Hop_ManPi(hopMan, 0), int32(truth[0]&1))
	default:
		obj.Data = kit.Kit_TruthToHop(hopMan, truth, int32(fun.nVars), nil)
	}
	return obj
}

func abcLevelNew(obj *Abc_Obj_t) int32 {
	var level int32
	for _, fanin := range obj.Fanins() {
		if fanin.Object != nil && int32(fanin.Object.Level) > level {
			level = int32(fanin.Object.Level)
		}
	}
	if len(obj.Fanins()) > 0 {
		level++
	}
	return level
}

func Lpk_Implement_rec(pMan *Lpk_Man_t, pNtk *Abc_Ntk_t, leaves *Vec_Ptr_t, fun *Lpk_Fun_t) *Abc_Obj_t {
	for i := 0; i < int(fun.nVars); i++ {
		faninIndex := int(fun.pFanins[i])
		fanin := vec.Vec_PtrEntry(leaves, faninIndex)
		if child, ok := fanin.(*Lpk_Fun_t); ok {
			Lpk_Implement_rec(pMan, pNtk, leaves, child)
		}
		if abc.Abc_ObjIsComplement(vec.Vec_PtrEntry(leaves, faninIndex)) == 0 {
			panic("lpk: implementation did not replace a decomposition fanin")
		}
	}
	res := Lpk_ImplementFun(pMan, pNtk, leaves, fun)
	vec.Vec_PtrWriteEntry(leaves, int(fun.Id), abc.Abc_ObjNot(res))
	Lpk_FunFree(fun)
	return res
}

func Lpk_Implement(pMan *Lpk_Man_t, pNtk *Abc_Ntk_t, leaves *Vec_Ptr_t, nLeavesOld int32) *Abc_Obj_t {
	if nLeavesOld < 0 || int(nLeavesOld) >= vec.Vec_PtrSize(leaves) {
		panic("lpk: invalid decomposition leaf boundary")
	}
	for i := 0; i < int(nLeavesOld); i++ {
		fanin := vec.Vec_PtrEntry(leaves, i)
		vec.Vec_PtrWriteEntry(leaves, i, abc.Abc_ObjNot(fanin))
	}
	fun, ok := vec.Vec_PtrEntry(leaves, int(nLeavesOld)).(*Lpk_Fun_t)
	if !ok {
		panic("lpk: missing root decomposition function")
	}
	res := Lpk_Implement_rec(pMan, pNtk, leaves, fun)
	vec.Vec_PtrShrink(leaves, int(nLeavesOld))
	return res
}

func Lpk_Decompose_rec(pMan *Lpk_Man_t, p *Lpk_Fun_t) int32 {
	if p.nLutK < 3 || p.nLutK > 6 || p.nVars <= p.nLutK {
		return 0
	}
	if Lpk_LutNumLuts(int32(p.nVars), int32(p.nLutK)) > int32(p.nAreaLim) || Lpk_SuppDelay(p.uSupp, p.pDelays[:]) > int32(p.nDelayLim) {
		return 0
	}
	if p.fSupports == 0 {
		Lpk_FunComputeCofSupps(p)
	}
	clk := time.Now()
	resDSD := Lpk_DsdAnalize(pMan, p, pMan.pPars.NVarsShared)
	pMan.timeEvalDsdAn += abctime(time.Since(clk))
	if resDSD != nil && (resDSD.nBSVars == int32(p.nLutK) || resDSD.nBSVars == int32(p.nLutK)-1) && resDSD.AreaEst <= int32(p.nAreaLim) && resDSD.DelayEst <= int32(p.nDelayLim) {
		clk = time.Now()
		p2 := Lpk_DsdSplit(pMan, p, resDSD.pCofVars[:], resDSD.nCofVars, resDSD.BSVars)
		pMan.timeEvalDsdSp += abctime(time.Since(clk))
		if p2 == nil || p2.nVars > p.nLutK {
			return 0
		}
		if p.nVars > p.nLutK && Lpk_Decompose_rec(pMan, p) == 0 {
			return 0
		}
		return 1
	}
	clk = time.Now()
	resMux := Lpk_MuxAnalize(pMan, p)
	pMan.timeEvalMuxAn += abctime(time.Since(clk))
	if resMux != nil && (resMux.DelayEst > int32(p.nDelayLim) || resMux.AreaEst > int32(p.nAreaLim)) {
		resMux = nil
	}
	if resMux != nil && resMux.nSuppSizeS <= int32(p.nLutK) && resMux.nSuppSizeL <= int32(p.nLutK) {
		resDSD = nil
	} else if resMux != nil && resDSD != nil {
		muxBetter := resMux.AreaEst < resDSD.AreaEst ||
			(resMux.AreaEst == resDSD.AreaEst && resMux.nSuppSizeL < resDSD.nSuppSizeL) ||
			(resMux.AreaEst == resDSD.AreaEst && resMux.nSuppSizeL == resDSD.nSuppSizeL && resMux.DelayEst < resDSD.DelayEst)
		if muxBetter {
			resDSD = nil
		} else {
			resMux = nil
		}
	}
	if resMux != nil {
		clk = time.Now()
		p2 := Lpk_MuxSplit(pMan, p, resMux.Variable, resMux.Polarity)
		pMan.timeEvalMuxSp += abctime(time.Since(clk))
		if p2 == nil {
			return 0
		}
		if p2.nVars > p.nLutK && Lpk_Decompose_rec(pMan, p2) == 0 {
			return 0
		}
		if p.nVars > p.nLutK && Lpk_Decompose_rec(pMan, p) == 0 {
			return 0
		}
		return 1
	}
	if resDSD != nil {
		clk = time.Now()
		p2 := Lpk_DsdSplit(pMan, p, resDSD.pCofVars[:], resDSD.nCofVars, resDSD.BSVars)
		pMan.timeEvalDsdSp += abctime(time.Since(clk))
		if p2 == nil || p2.nVars > p.nLutK {
			return 0
		}
		if p.nVars > p.nLutK && Lpk_Decompose_rec(pMan, p) == 0 {
			return 0
		}
		return 1
	}
	return 0
}

func Lpk_DecomposeClean(leaves *Vec_Ptr_t, nLeavesOld int32) {
	for i := int(nLeavesOld); i < vec.Vec_PtrSize(leaves); i++ {
		if fun, ok := vec.Vec_PtrEntry(leaves, i).(*Lpk_Fun_t); ok {
			Lpk_FunFree(fun)
		}
	}
	vec.Vec_PtrShrink(leaves, int(nLeavesOld))
}

func Lpk_Decompose(p *Lpk_Man_t, pNtk *Abc_Ntk_t, leaves *Vec_Ptr_t, truth []uint32, supps []uint32, nLutK, areaLim, delayLim int32) *Abc_Obj_t {
	nLeaves := vec.Vec_PtrSize(leaves)
	fun := Lpk_FunCreate(pNtk, leaves, truth, nLutK, areaLim, delayLim)
	if fun == nil {
		return nil
	}
	if len(supps) >= 2*nLeaves && (supps[0] != 0 || supps[1] != 0) {
		copy(fun.puSupps[:2*nLeaves], supps[:2*nLeaves])
		fun.fSupports = 1
	}
	Lpk_FunSuppMinimize(fun)
	var result *Abc_Obj_t
	if int32(fun.nVars) <= int32(fun.nLutK) {
		result = Lpk_ImplementFun(p, pNtk, leaves, fun)
	} else if Lpk_Decompose_rec(p, fun) != 0 {
		result = Lpk_Implement(p, pNtk, leaves, int32(nLeaves))
	}
	Lpk_DecomposeClean(leaves, int32(nLeaves))
	return result
}
