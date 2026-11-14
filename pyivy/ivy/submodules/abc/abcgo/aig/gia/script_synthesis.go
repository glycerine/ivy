package gia

import "github.com/glycerine/abcgo/misc/util"

// giaAigSyn2 carries out the area/delay script using the supplied DSD mapper.
// The mapper lives in the IF package, which itself imports GIA.
func giaAigSyn2(input *Gia_Man_t, old, coarsen, cutMin, relax, delayMin, verbose, veryVerbose int32, dsdBalance func(*Gia_Man_t, int32, int32, int32, int32) *Gia_Man_t) *Gia_Man_t {
	var pars Jf_Par_t
	if old != 0 {
		Jf_ManSetDefaultPars(&pars)
		pars.FCutMin = cutMin
	} else {
		Lf_ManSetDefaultPars(&pars)
		pars.FCutMin, pars.FCoarsen, pars.NRelaxRatio = cutMin, coarsen, relax
		pars.NAreaTuner, pars.NCutNum = 1, 4
	}
	if verbose != 0 {
		Gia_ManPrintStats(input, nil)
	}
	p := Gia_ManDup(input)
	Gia_ManTransferTiming(p, input)
	if Gia_ManAndNum(p) == 0 {
		result := Gia_ManDup(p)
		Gia_ManTransferTiming(result, p)
		Gia_ManStop(p)
		return result
	}
	mapGraph := func(g *Gia_Man_t) *Gia_Man_t {
		if old != 0 {
			return Jf_ManPerformMapping(g, &pars)
		}
		return Lf_ManPerformMapping(g, &pars)
	}
	if delayMin != 0 && p.PManTime == nil {
		savedCutMin, savedCoarsen, savedRelax := pars.FCutMin, pars.FCoarsen, pars.NRelaxRatio
		pars.FCutMin, pars.FCoarsen, pars.NRelaxRatio = 0, 0, 0
		mapGraph(p)
		area0, delay0 := int32(pars.Area), int32(pars.Delay)
		balanced := dsdBalance(p, 6, 4, 0, 0)
		mapGraph(balanced)
		area1, delay1 := int32(pars.Area), int32(pars.Delay)
		if delay1 < delay0-1 || (delay1 == delay0+1 && 100.0*float64(area1-area0)/float64(area1) < 3.0) {
			Gia_ManStop(p)
			p = balanced
		} else {
			Gia_ManStop(balanced)
			p.VMapping = nil
		}
		pars.FCutMin, pars.FCoarsen, pars.NRelaxRatio = savedCutMin, savedCoarsen, savedRelax
	}
	result := Gia_ManAreaBalance(p, 0, util.ABC_INFINITY, veryVerbose, 0)
	if verbose != 0 {
		Gia_ManPrintStats(result, nil)
	}
	Gia_ManStop(p)
	previous := result
	result = mapGraph(previous)
	if verbose != 0 {
		Gia_ManPrintStats(result, nil)
	}
	if previous != result {
		Gia_ManStop(previous)
	}
	previous = result
	result = Gia_ManAreaBalance(previous, 0, util.ABC_INFINITY, veryVerbose, 0)
	if verbose != 0 {
		Gia_ManPrintStats(result, nil)
	}
	Gia_ManStop(previous)
	return result
}

func Gia_ManAigSyn3(p *Gia_Man_t, verbose, veryVerbose int32) *Gia_Man_t {
	var pars Jf_Par_t
	Jf_ManSetDefaultPars(&pars)
	pars.NRelaxRatio = 40
	if verbose != 0 {
		Gia_ManPrintStats(p, nil)
	}
	if Gia_ManAndNum(p) == 0 {
		q := Gia_ManDup(p)
		Gia_ManTransferTiming(q, p)
		return q
	}
	q := Gia_ManAreaBalance(p, 0, util.ABC_INFINITY, veryVerbose, 0)
	for _, size := range []int32{6, 4} {
		pars.NLutSize = size
		q = Jf_ManPerformMapping(q, &pars)
		if verbose != 0 {
			Gia_ManPrintStats(q, nil)
		}
		q = Gia_ManAreaBalance(q, 0, util.ABC_INFINITY, veryVerbose, 0)
	}
	if verbose != 0 {
		Gia_ManPrintStats(q, nil)
	}
	return q
}
func Gia_ManAigSyn4(p *Gia_Man_t, verbose, veryVerbose int32) *Gia_Man_t {
	var pars Jf_Par_t
	Jf_ManSetDefaultPars(&pars)
	pars.NRelaxRatio = 40
	if verbose != 0 {
		Gia_ManPrintStats(p, nil)
	}
	if Gia_ManAndNum(p) == 0 {
		q := Gia_ManDup(p)
		Gia_ManTransferTiming(q, p)
		return q
	}
	q := Gia_ManAreaBalance(p, 0, util.ABC_INFINITY, veryVerbose, 0)
	for _, size := range []int32{7, 5} {
		pars.NLutSize = size
		q = Jf_ManPerformMapping(q, &pars)
		q = Gia_ManPerformFx(q, util.ABC_INFINITY, 0, 0, veryVerbose, 0)
		q = Gia_ManAreaBalance(q, 0, util.ABC_INFINITY, veryVerbose, 0)
		if verbose != 0 {
			Gia_ManPrintStats(q, nil)
		}
	}
	return q
}
