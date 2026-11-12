package bmc

import "github.com/glycerine/abcgo/aig/gia"

// Bmcg_ManPerformOne runs bounded checking through the common native frame
// encoder and SAT solver. The historical BMCg implementation differed in its
// graph traversal, while its public contract is the same bounded status.
func Bmcg_ManPerformOne(graph *gia.Gia_Man_t, params *Bmc_AndPar_t) int32 {
	return Gia_ManBmcPerform(graph, params)
}

func Bmcg_ManPerform(graph *gia.Gia_Man_t, params *Bmc_AndPar_t) int32 {
	return Gia_ManBmcPerform(graph, params)
}
