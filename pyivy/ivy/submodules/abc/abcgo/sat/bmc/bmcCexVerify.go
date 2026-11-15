package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// Bmc_CexVerify shares the source ternary verification with GIA's minimizer.
func Bmc_CexVerify(p *gia.Gia_Man_t, cex, care *util.Abc_Cex_t) int32 {
	return gia.Gia_ManCexCareVerify(p, cex, care)
}
