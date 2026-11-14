package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

func Lf_ManPrintQuit(p *Lf_Man_t, result *Gia_Man_t) {
	if p.pPars.FVerbose == 0 {
		if p.pPars.FCutGroup != 0 {
			var regular, dual int
			lut := p.pPars.NLutSize / 2
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
	if pars.FUseMux7 != 0 {
		pars.FCoarsen = 1
		pars.NRoundsEla = 0
	}
	if Gia_ManHasChoices(g) != 0 || pars.NLutSizeMux != 0 {
		pars.FCutMin = 1
	}
	coarse := g
	if pars.FCoarsen != 0 {
		coarse = Gia_ManDupMuxes(g, pars.NCoarseLimit)
		coarse.PManTime = g.PManTime
		g.PManTime = nil
	}
	p := Lf_ManAlloc(coarse, pars)
	Lf_ManPrintInit(p)
	for p.Iter = 0; p.Iter < pars.NRounds; p.Iter++ {
		Lf_ManComputeMapping(p)
	}
	p.fUseEla = 1
	for ; p.Iter < pars.NRounds+pars.NRoundsEla; p.Iter++ {
		Lf_ManComputeMapping(p)
	}
	if pars.FVeryVerbose != 0 && pars.FCutMin != 0 {
		vec.Vec_MemDumpTruthTables(p.vTtMem, coarse.pName, int(pars.NLutSize))
	}
	var result *Gia_Man_t
	if pars.FCutMin != 0 {
		result = Lf_ManDeriveMappingGia(p)
	} else if pars.FCoarsen != 0 {
		result = Lf_ManDeriveMappingCoarse(p)
	} else {
		result = Lf_ManDeriveMapping(p)
	}
	Gia_ManMappingVerify(result)
	Lf_ManPrintQuit(p, result)
	Lf_ManFree(p)
	if coarse != g {
		g.PManTime = coarse.PManTime
		coarse.PManTime = nil
		Gia_ManStop(coarse)
	}
	return result
}
func Lf_ManPerformMapping(g *Gia_Man_t, pars *Jf_Par_t) *Gia_Man_t {
	if g.PManTime != nil {
		timing, ok := g.PManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("LF invalid timing manager")
		}
		if tim.Tim_ManBoxNum(timing) != 0 && Gia_ManIsNormalized(g) != 0 {
			g.PManTime = tim.Tim_ManDup(timing, 1)
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
			if normalized.PManTime != nil {
				tim.Tim_ManStop(normalized.PManTime.(*tim.Tim_Man_t))
			}
			normalized.PManTime = timing
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
	if pars.FCutGroup != 0 && pars.NLutSize != 9 && pars.NLutSize != 11 && pars.NLutSize != 13 {
		panic("LF invalid grouped LUT size")
	}
	if pars.pTimesArr != nil || pars.pTimesReq != nil {
		panic("LF arrival arrays already set")
	}
	source := g
	if g.PManTime != nil && normalized != 0 {
		source = Gia_ManDupUnnormalize(g)
		if source == nil {
			return nil
		}
		Gia_ManTransferTiming(source, g)
		tm, ok := source.PManTime.(*tim.Tim_Man_t)
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
		result.PSpec = source.PSpec
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
	if g.PMuxes != nil {
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
