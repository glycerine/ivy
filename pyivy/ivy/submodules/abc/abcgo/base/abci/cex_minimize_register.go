package abci

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	abcio "github.com/glycerine/abcgo/base/io"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/proof/abs"
	"github.com/glycerine/abcgo/sat/bmc"
)

// write_cex -m minimizes as Abc_NtkDumpOneCex does: -o with
// Saig_ManCbaFindCexCareBits (checked with -c), -u SAT-based, otherwise
// Bmc_CexCareMinimize with four tries.
func init() {
	abcio.RegisterCexMinimizer(func(g *aig.Graph, cex *util.Abc_Cex_t, oldMin, satBased, highEffort, check, verbose bool) *util.Abc_Cex_t {
		b := func(v bool) int32 {
			if v {
				return 1
			}
			return 0
		}
		switch {
		case oldMin:
			care := abs.Saig_ManCbaFindCexCareBits(g, cex, 0, b(verbose))
			if check {
				bmc.Bmc_CexCareVerify(g, cex, care, b(verbose))
			}
			return care
		case satBased:
			return bmc.Bmc_CexCareSatBasedMinimize(g, saig.Saig_ManPiNum(g), cex, b(highEffort), b(check), b(verbose))
		default:
			return bmc.Bmc_CexCareMinimize(g, saig.Saig_ManPiNum(g), cex, 4, b(check), b(verbose))
		}
	})
}
