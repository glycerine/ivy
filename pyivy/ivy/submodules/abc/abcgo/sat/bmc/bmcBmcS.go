package bmc

import "github.com/glycerine/abcgo/aig/gia"

// Bmcs_ManPerformOne uses the shared native unroller and typed SAT solver.
// This keeps the established entry point while avoiding the C port's mutable
// per-object frame copies.
func Bmcs_ManPerformOne(graph *gia.Gia_Man_t, params *Bmc_AndPar_t) int32 {
	return bmcPerformScalable(graph, params)
}

func Bmcs_ManPerform(graph *gia.Gia_Man_t, params *Bmc_AndPar_t) int32 {
	return bmcPerformScalable(graph, params)
}

// bmcPerformScalable checks the GIA properties with the bmc3 engine
// (Saig_ManBmcScalable) for the &bmcs entry points, which are not ported
// from bmcBmcS.c/bmcBmcG.c. It reports bounded safety as 1 when it completed
// the caller's explicit frame bound without a resource interruption.
func bmcPerformScalable(graph *gia.Gia_Man_t, params *Bmc_AndPar_t) int32 {
	if graph == nil || params == nil {
		return -1
	}
	graph.ClearSequentialCounterexample()
	native := gia.Gia_ManToAig(graph, 0)
	options := Saig_ParBmc_t{
		NStart:      params.NStart,
		NFramesMax:  params.NFramesMax,
		NConfLimit:  params.NConfLimit,
		NTimeOut:    params.NTimeOut,
		FVerbose:    params.FVerbose,
		FNotVerbose: params.FNotVerbose,
	}
	result := Saig_ManBmcScalable(native, &options)
	if result == -1 && params.NFramesMax > 0 && params.NConfLimit == 0 && params.NTimeOut == 0 {
		result = 1
	}
	params.IFrame = options.IFrame
	params.NFailOuts = options.NFailOuts
	if cex := native.SequentialCounterexample(); cex != nil {
		_ = graph.SetSequentialCounterexample(cex)
	}
	return result
}
