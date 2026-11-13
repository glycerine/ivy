package bmc

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
)

// Bmc_CexDepthTest checks a counterexample's care set and appends one induced
// state condition for each selected suffix of the trace.
func Bmc_CexDepthTest(p *gia.Gia_Man_t, cex *util.Abc_Cex_t, frames, verbose int32) *gia.Gia_Man_t {
	if p == nil || cex == nil || frames < 0 {
		return nil
	}
	var implied *util.Abc_Cex_t
	states := Bmc_CexInnerStates(p, cex, &implied, verbose)
	care := Bmc_CexCareBits(p, states, implied, nil, 1, verbose)
	if Bmc_CexVerify(p, cex, care) == 0 {
		fmt.Println("Counter-example care-set verification has failed.")
	}
	return Bmc_CexBuildNetwork2Test(p, states, frames)
}
