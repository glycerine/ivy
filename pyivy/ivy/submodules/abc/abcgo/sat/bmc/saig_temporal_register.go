package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

func init() {
	saig.RegisterTemporalBMC(func(g *aig.Graph, frames, timeout, conflicts, verbose int32, finished *int32) int32 {
		return Saig_BmcPerform(g, 0, frames, 2000, timeout, conflicts, 0, verbose, 0, finished, 0, 0)
	})
}
