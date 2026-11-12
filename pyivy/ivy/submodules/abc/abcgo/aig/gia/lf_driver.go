package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Lf_ManPrintQuit(p *Lf_Man_t, result *Gia_Man_t) {
	if p.pPars.fVerbose == 0 {
		if p.pPars.fCutGroup != 0 {
			var regular, dual int
			lut := p.pPars.nLutSize / 2
			for id := int32(0); id < result.nObjs; id++ {
				if Gia_ObjIsLut(result, id) != 0 {
					if Gia_ObjLutSize(result, id) > lut {
						dual++
					} else {
						regular++
					}
				}
			}
			fmt.Printf("Created %d regular %d-LUTs and %d dual %d-LUTs. The total of %d %d-LUTs.\n", regular, lut, dual, lut, regular+2*dual, lut)
		}
		return
	}
	fmt.Printf("CutPair = %.0f  Merge = %.0f  Eval = %.0f  Cut = %.0f\n", p.CutCount[0], p.CutCount[1], p.CutCount[2], p.CutCount[3])
	fmt.Printf("CoDrvs = %d  CoInvs = %d  Front = %d  TimeFails = %d   Time = %s\n", p.nCoDrivers, p.nInverters, p.nFrontMax, p.nTimeFails, time.Since(p.clkStart))
}
func Lf_ManPerformMappingInt(g *Gia_Man_t, pars *Jf_Par_t) *Gia_Man_t {
	if pars.fUseMux7 != 0 {
		pars.fCoarsen = 1
		pars.nRoundsEla = 0
	}
	if Gia_ManHasChoices(g) != 0 || pars.nLutSizeMux != 0 {
		pars.fCutMin = 1
	}
	coarse := g
	if pars.fCoarsen != 0 {
		coarse = Gia_ManDupMuxes(g, pars.nCoarseLimit)
		coarse.pManTime = g.pManTime
		g.pManTime = nil
	}
	p := Lf_ManAlloc(coarse, pars)
	Lf_ManPrintInit(p)
	for p.Iter = 0; p.Iter < pars.nRounds; p.Iter++ {
		Lf_ManComputeMapping(p)
	}
	p.fUseEla = 1
	for ; p.Iter < pars.nRounds+pars.nRoundsEla; p.Iter++ {
		Lf_ManComputeMapping(p)
	}
	if pars.fVeryVerbose != 0 && pars.fCutMin != 0 {
		vec.Vec_MemDumpTruthTables(p.vTtMem, coarse.pName, int(pars.nLutSize))
	}
	var result *Gia_Man_t
	if pars.fCutMin != 0 {
		result = Lf_ManDeriveMappingGia(p)
	} else if pars.fCoarsen != 0 {
		result = Lf_ManDeriveMappingCoarse(p)
	} else {
		result = Lf_ManDeriveMapping(p)
	}
	Gia_ManMappingVerify(result)
	Lf_ManPrintQuit(p, result)
	Lf_ManFree(p)
	if coarse != g {
		g.pManTime = coarse.pManTime
		coarse.pManTime = nil
		Gia_ManStop(coarse)
	}
	return result
}
func Lf_ManPerformMapping(g *Gia_Man_t, pars *Jf_Par_t) *Gia_Man_t {
	if g.pManTime != nil {
		timing, ok := g.pManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("LF invalid timing manager")
		}
		if tim.Tim_ManBoxNum(timing) != 0 && Gia_ManIsNormalized(g) != 0 {
			g.pManTime = tim.Tim_ManDup(timing, 1)
			source := Gia_ManDupUnnormalize(g)
			if source == nil {
				return nil
			}
			Gia_ManTransferTiming(source, g)
			result := Lf_ManPerformMappingInt(source, pars)
			if result != source {
				Gia_ManTransferTiming(result, source)
				Gia_ManStop(source)
			}
			normalized := Gia_ManDupNormalize(result, 0)
			Gia_ManTransferMapping(normalized, result)
			Gia_ManTransferTiming(normalized, result)
			Gia_ManStop(result)
			if normalized.pManTime != nil {
				tim.Tim_ManStop(normalized.pManTime.(*tim.Tim_Man_t))
			}
			normalized.pManTime = timing
			if Gia_ManIsNormalized(normalized) == 0 {
				panic("LF normalization failed")
			}
			return normalized
		}
	}
	result := Lf_ManPerformMappingInt(g, pars)
	Gia_ManTransferTiming(result, g)
	return result
}
func Gia_ManPerformLfMapping(g *Gia_Man_t, pars *Jf_Par_t, normalized int32) *Gia_Man_t {
	if pars.fCutGroup != 0 && pars.nLutSize != 9 && pars.nLutSize != 11 && pars.nLutSize != 13 {
		panic("LF invalid grouped LUT size")
	}
	if pars.pTimesArr != nil || pars.pTimesReq != nil {
		panic("LF arrival arrays already set")
	}
	source := g
	if g.pManTime != nil && normalized != 0 {
		source = Gia_ManDupUnnormalize(g)
		if source == nil {
			return nil
		}
		Gia_ManTransferTiming(source, g)
		tm, ok := source.pManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("LF invalid timing manager")
		}
		pars.pTimesArr = tim.Tim_ManGetArrTimes(tm)
		pars.pTimesReq = tim.Tim_ManGetReqTimes(tm)
	} else {
		source = Gia_ManDup(g)
	}
	result := Lf_ManPerformMappingInt(source, pars)
	if result != source {
		result.pName = source.pName
		result.pSpec = source.pSpec
		Gia_ManSetRegNum(result, Gia_ManRegNum(source))
		Gia_ManTransferTiming(result, source)
		Gia_ManStop(source)
	}
	out := Gia_ManDupNormalize(result, 0)
	Gia_ManTransferMapping(out, result)
	Gia_ManTransferTiming(out, result)
	Gia_ManStop(result)
	return out
}

func Lf_ManComputeCrossCut(g *Gia_Man_t) int32 {
	if g.pMuxes != nil {
		panic("LF crosscut expects ordinary AIG")
	}
	Gia_ManCleanValue(g)
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		for _, fanin := range []*Gia_Obj_t{Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)} {
			if Gia_ObjIsAnd(fanin) != 0 {
				fanin.Value++
			}
		}
	}
	var current, maximum int32
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		if obj.Value != 0 {
			current++
		}
		maximum = max(maximum, current)
		for _, fanin := range []*Gia_Obj_t{Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)} {
			if Gia_ObjIsAnd(fanin) != 0 {
				fanin.Value--
				if fanin.Value == 0 {
					current--
				}
			}
		}
	}
	if current != 0 {
		panic("LF crosscut did not drain")
	}
	for _, obj := range g.pObjs {
		if obj.Value != 0 {
			panic("LF crosscut reference remains")
		}
	}
	fmt.Printf("CutMax = %d\n", maximum)
	return maximum
}
