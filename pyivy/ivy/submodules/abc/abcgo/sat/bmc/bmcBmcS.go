package bmc

import "github.com/glycerine/abcgo/aig/gia"

// Bmcs_ManPerformOne uses the shared native unroller and typed SAT solver.
// This keeps the established entry point while avoiding the C port's mutable
// per-object frame copies.
func Bmcs_ManPerformOne(graph *gia.Gia_Man_t, params *Bmc_AndPar_t) int32 {
	return Gia_ManBmcPerform(graph, params)
}

func Bmcs_ManPerform(graph *gia.Gia_Man_t, params *Bmc_AndPar_t) int32 {
	return Gia_ManBmcPerform(graph, params)
}
