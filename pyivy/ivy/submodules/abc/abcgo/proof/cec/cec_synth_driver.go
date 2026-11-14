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
	domains := p.ClockDomains()
	if domains == nil {
		panic("cec: missing clock-domain partitions")
	}
	reprs := make([]int32, gia.Gia_ManObjNum(p))
	for i := range reprs {
		reprs[i] = -1
	}
	for index, part := range domains {
		var nPIs, nRegs int32
		var mapping []int32
		sub := Gia_ManRegCreatePart(p, part, &nPIs, &nRegs, &mapping)
		if nPIs == 0 {
			// The C driver releases partitions even when they have no primary
			// inputs and therefore skip synthesis and class transfer.
			gia.Gia_ManStop(sub)
			continue
		}
		switch {
		case pars.fUseScorr != 0:
			var cor Cec_ParCor_t
			Cec_ManCorSetDefaultParams(&cor)
			cor.NBTLimit = pars.nBTLimit
			cor.nLevelMax = pars.nLevelMax
			cor.FVerbose = pars.fVeryVerbose
			cor.FUseCSat = 1
			Cec_ManLSCorrespondenceClasses(sub, &cor)
		case pars.fUseLcorr != 0:
			var cor Cec_ParCor_t
			Cec_ManCorSetDefaultParams(&cor)
			cor.FLatchCorr = 1
			cor.NBTLimit = pars.nBTLimit
			cor.FVerbose = pars.fVeryVerbose
			cor.FUseCSat = 1
			Cec_ManLSCorrespondenceClasses(sub, &cor)
		default:
			gia.Gia_ManSeqCleanupClasses(sub, pars.fConsts, pars.fEquivs, pars.fVerbose)
		}
		classes := Gia_TransferMappedClasses(sub, mapping, reprs)
		if pars.fVerbose != 0 {
			fmt.Printf("%3d : Reg = %4d. PI = %4d. (True = %4d. Regs = %4d.) And = %5d. Cl = %5d.\n", index, vec.Vec_IntSize(part), gia.Gia_ManCiNum(sub)-int32(vec.Vec_IntSize(part)), nPIs, nRegs, gia.Gia_ManAndNum(sub), classes)
		}
		gia.Gia_ManStop(sub)
	}
	Gia_ManNormalizeEquivalences(p, reprs)
	if pars.fVerbose != 0 {
		fmt.Printf("Total time: %s\n", time.Since(start))
	}
	return 1
}
