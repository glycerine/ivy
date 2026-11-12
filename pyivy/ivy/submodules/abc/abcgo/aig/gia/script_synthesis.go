package gia

import "github.com/glycerine/abcgo/misc/util"

func Gia_ManAigSyn3(p *Gia_Man_t, verbose, veryVerbose int32) *Gia_Man_t {
	var pars Jf_Par_t
	Jf_ManSetDefaultPars(&pars)
	pars.nRelaxRatio = 40
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
		pars.nLutSize = size
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
	pars.nRelaxRatio = 40
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
		pars.nLutSize = size
		q = Jf_ManPerformMapping(q, &pars)
		q = Gia_ManPerformFx(q, util.ABC_INFINITY, 0, 0, veryVerbose, 0)
		q = Gia_ManAreaBalance(q, 0, util.ABC_INFINITY, veryVerbose, 0)
		if verbose != 0 {
			Gia_ManPrintStats(q, nil)
		}
	}
	return q
}
