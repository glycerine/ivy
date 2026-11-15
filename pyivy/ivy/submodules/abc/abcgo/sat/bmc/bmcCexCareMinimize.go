package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// Bmc_CexCareMinimizeAig uses the shared original priority implementation.
func Bmc_CexCareMinimizeAig(p *gia.Gia_Man_t, realPis int32, cex *util.Abc_Cex_t, tries, check, verbose int32) *util.Abc_Cex_t {
	return gia.Gia_ManCexCareMinimize(p, realPis, cex, tries, check, verbose)
}

// Bmc_CexCareMinimize applies the same minimization to an AIG manager.
func Bmc_CexCareMinimize(p *aig.Graph, realPis int32, cex *util.Abc_Cex_t, tries, check, verbose int32) *util.Abc_Cex_t {
	if p == nil {
		return nil
	}
	return Bmc_CexCareMinimizeAig(gia.Gia_ManFromAigSimple(p), realPis, cex, tries, check, verbose)
}
