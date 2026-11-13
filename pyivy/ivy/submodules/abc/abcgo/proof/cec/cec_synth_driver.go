package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// Cec_SequentialSynthesisPart computes equivalences independently for each
// clock-domain partition and transfers them back to the original graph.
func Cec_SequentialSynthesisPart(p *gia.Gia_Man_t, pars *Cec_ParSeq_t) int32 {
	start := time.Now()
	reprs := make([]int32, gia.Gia_ManObjNum(p))
	for i := range reprs {
		reprs[i] = -1
	}
	for index, part := range p.ClockDomains() {
		var nPIs, nRegs int32
		var mapping []int32
		sub := Gia_ManRegCreatePart(p, part, &nPIs, &nRegs, &mapping)
		if nPIs == 0 {
			continue
		}
		switch {
		case pars.fUseScorr != 0:
			var cor Cec_ParCor_t
			Cec_ManCorSetDefaultParams(&cor)
			cor.nBTLimit = pars.nBTLimit
			cor.nLevelMax = pars.nLevelMax
			cor.fVerbose = pars.fVeryVerbose
			cor.fUseCSat = 1
			Cec_ManLSCorrespondenceClasses(sub, &cor)
		case pars.fUseLcorr != 0:
			var cor Cec_ParCor_t
			Cec_ManCorSetDefaultParams(&cor)
			cor.fLatchCorr = 1
			cor.nBTLimit = pars.nBTLimit
			cor.fVerbose = pars.fVeryVerbose
			cor.fUseCSat = 1
			Cec_ManLSCorrespondenceClasses(sub, &cor)
		default:
			gia.Gia_ManSeqCleanupClasses(sub, pars.fConsts, pars.fEquivs, pars.fVerbose)
		}
		classes := Gia_TransferMappedClasses(sub, mapping, reprs)
		if pars.fVerbose != 0 {
			fmt.Printf("%3d : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d. Cl = %5d.\n", index, vec.Vec_IntSize(part), gia.Gia_ManCiNum(sub)-int32(vec.Vec_IntSize(part)), nPIs, nRegs, gia.Gia_ManAndNum(sub), classes)
		}
	}
	Gia_ManNormalizeEquivalences(p, reprs)
	if pars.fVerbose != 0 {
		fmt.Printf("Total time: %s\n", time.Since(start))
	}
	return 1
}
