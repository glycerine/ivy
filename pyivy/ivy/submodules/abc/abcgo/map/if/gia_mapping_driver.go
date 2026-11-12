package if_

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/tim"
)

// SetDsdManager supplies the shared decomposition manager used during mapping.
func (p *If_Par_t_) SetDsdManager(manager *If_DsdMan_t) { p.pDsdManager = manager }

// SetSwitchingProvider supplies switching probabilities when power mapping is enabled.
func (p *If_Par_t_) SetSwitchingProvider(provider func(*If_Man_t)) { p.pFuncSwitching = provider }

// SetDelayTrace supplies the optional post-mapping delay-trace reporter.
func (p *If_Par_t_) SetDelayTrace(trace func(*Gia_Man_t)) { p.pFuncDelayTrace = trace }

// Gia_ManPerformMappingInt maps one GIA in its current interface order.
func Gia_ManPerformMappingInt(src *Gia_Man_t, pars *If_Par_t) *Gia_Man_t {
	if src == nil || pars == nil {
		panic("if: missing mapping input")
	}
	assert(pars.pTimesArr == nil && pars.pTimesReq == nil, "timing arrays start empty")
	arrivals, required, attrs, timing := src.MappingInputs()
	if arrivals != nil {
		assert(len(arrivals) == int(Gia_ManCiNum(src)), "CI arrivals match interface")
		pars.pTimesArr = make([]float32, len(arrivals))
		for i, value := range arrivals {
			pars.pTimesArr[i] = float32(value)
		}
	}
	if required != nil {
		assert(len(required) == int(Gia_ManCoNum(src)), "CO requirements match interface")
		pars.pTimesReq = make([]float32, len(required))
		for i, value := range required {
			pars.pTimesReq[i] = float32(value)
		}
	}
	defer func() { pars.pTimesArr, pars.pTimesReq = nil, nil }()
	src.ClearCellConfiguration()
	if pars.fDelayOpt == 0 && pars.fDelayOptLut == 0 && pars.fDsdBalance == 0 && pars.fUserRecLib == 0 && pars.fUserSesLib == 0 && pars.fDeriveLuts == 0 && pars.fUseDsd == 0 && pars.fUseTtPerm == 0 && pars.pFuncCell2 == nil {
		pars.fCutMin = 0
	}
	manager := Gia_ManToIf(src, pars)
	defer If_ManStop(manager)
	if pars.fUseDsd != 0 {
		dsd := pars.pDsdManager
		assert(dsd != nil && pars.nLutSize <= If_DsdManVarNum(dsd), "compatible DSD manager is provided")
		assert(pars.pLutStruct == "" && If_DsdManLutSize(dsd) == 0 || pars.pLutStruct != "" && int32(pars.pLutStruct[0]-'0') == If_DsdManLutSize(dsd), "DSD LUT structure matches")
		manager.pIfDsdMan = dsd
		if pars.fDsdBalance != 0 {
			If_DsdManAllocIsops(dsd, pars.nLutSize)
		}
	}
	if pars.fPower != 0 {
		if timing != nil {
			fmt.Println("Switching activity computation for designs with boxes is disabled.")
		} else {
			if pars.pFuncSwitching == nil {
				panic("if: power mapping requires a switching provider")
			}
			pars.pFuncSwitching(manager)
		}
	}
	manager.pUserMan = pars.pReoMan
	if timing != nil {
		unit := 0
		if pars.fDelayOpt != 0 || pars.fDelayOptLut != 0 || pars.fDsdBalance != 0 || pars.fUserRecLib != 0 || pars.fUserSesLib != 0 {
			unit = 1
		}
		manager.pManTim = tim.Tim_ManDup(timing, unit)
	}
	if attrs != nil {
		assert(len(attrs) == int(If_ManCoNum(manager)), "CO attributes match interface")
		for i, value := range attrs {
			If_ObjFanin0(If_ManCo(manager, int32(i))).fSpec = uint32(ifTuneBoolInt(value != 0))
		}
	}
	if If_ManPerformMapping(manager) == 0 {
		return nil
	}
	if pars.pFuncWrite != nil {
		pars.pFuncWrite(manager)
	}
	var result *Gia_Man_t
	if pars.fDelayOpt != 0 || pars.fDsdBalance != 0 || pars.fUserRecLib != 0 || pars.fUserSesLib != 0 {
		result = Gia_ManFromIfAig(manager)
	} else {
		result = Gia_ManFromIfLogic(manager)
	}
	if result == nil {
		return nil
	}
	if arrivals != nil || required != nil {
		out := make([]int32, If_ManCoNum(manager))
		for i := range out {
			out[i] = int32(If_ObjArrTime(If_ObjFanin0(If_ManCo(manager, int32(i)))))
		}
		src.SetMappingOutputArrivals(out)
	}
	name, spec := src.Identity()
	result.SetIdentity(name, spec)
	Gia_ManSetRegNum(result, Gia_ManRegNum(src))
	if pars.fVerboseTrace != 0 {
		if pars.pFuncDelayTrace == nil {
			panic("if: verbose delay trace requires a reporter")
		}
		pars.pFuncDelayTrace(result)
	}
	return result
}

// Gia_ManPerformMapping handles boxed timing normalization and optional hashing.
func Gia_ManPerformMapping(src *Gia_Man_t, pars *If_Par_t) *Gia_Man_t {
	if src == nil || pars == nil {
		panic("if: missing mapping input")
	}
	var result *Gia_Man_t
	if gia.Gia_ManBoxNum(src) != 0 && gia.Gia_ManIsNormalized(src) != 0 {
		work := gia.Gia_ManDupUnnormalize(src)
		if work == nil {
			return nil
		}
		gia.Gia_ManTransferTiming(work, src)
		mapped := Gia_ManPerformMappingInt(work, pars)
		if mapped == nil {
			Gia_ManStop(work)
			return nil
		}
		gia.Gia_ManTransferTiming(mapped, work)
		Gia_ManStop(work)
		result = gia.Gia_ManDupNormalize(mapped, pars.fHashMapping)
		gia.Gia_ManTransferMapping(result, mapped)
		gia.Gia_ManTransferPacking(result, mapped)
		gia.Gia_ManTransferTiming(result, mapped)
		Gia_ManStop(mapped)
		assert(gia.Gia_ManIsNormalized(result) != 0, "mapped graph is normalized")
	} else {
		result = Gia_ManPerformMappingInt(src, pars)
		if result == nil {
			return nil
		}
		gia.Gia_ManTransferTiming(result, src)
		if pars.fHashMapping != 0 {
			mapped := result
			result = gia.Gia_ManDupHashMapping(mapped)
			gia.Gia_ManTransferPacking(result, mapped)
			gia.Gia_ManTransferTiming(result, mapped)
			Gia_ManStop(mapped)
		}
	}
	result.MappedDelay, result.MappedArea = int32(pars.FinalDelay), int32(pars.FinalArea)
	return result
}

// Gia_ManPerformSopBalance maps and rebuilds an AIG using SOP balancing.
func Gia_ManPerformSopBalance(src *Gia_Man_t, cuts, relax, verbose int32) *Gia_Man_t {
	var pars If_Par_t
	If_ManSetDefaultPars(&pars)
	pars.nCutsMax, pars.nRelaxRatio, pars.fVerbose = cuts, relax, verbose
	pars.nLutSize, pars.fDelayOpt, pars.fCutMin, pars.fTruth, pars.fExpRed = 6, 1, 1, 1, 0
	result := Gia_ManPerformMappingInt(src, &pars)
	if result != nil {
		gia.Gia_ManTransferTiming(result, src)
	}
	return result
}

// Gia_ManPerformDsdBalance maps and rebuilds an AIG using DSD balancing.
func Gia_ManPerformDsdBalance(src *Gia_Man_t, lutSize, cuts, relax, verbose int32) *Gia_Man_t {
	var pars If_Par_t
	If_ManSetDefaultPars(&pars)
	pars.nCutsMax, pars.nRelaxRatio, pars.fVerbose = cuts, relax, verbose
	pars.nLutSize, pars.fDsdBalance, pars.fUseDsd, pars.fCutMin, pars.fTruth, pars.fExpRed = lutSize, 1, 1, 1, 1, 0
	manager := If_DsdManAlloc(lutSize, 0)
	defer If_DsdManFree(manager, 0)
	pars.SetDsdManager(manager)
	result := Gia_ManPerformMappingInt(src, &pars)
	if result != nil {
		gia.Gia_ManTransferTiming(result, src)
	}
	return result
}
