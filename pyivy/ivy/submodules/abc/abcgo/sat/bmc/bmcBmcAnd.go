package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
)

// Gia_ManBmcPerform checks the GIA properties through the native sequential
// AIG unroller and shared SAT-backed scalable BMC engine. GIA's object fields
// are private; translating it once to Graph keeps the BMC implementation
// independent of pointer arithmetic and scratch fields in Gia_Obj_t.
func Gia_ManBmcPerform(graph *gia.Gia_Man_t, params *Bmc_AndPar_t) int32 {
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
	// This adapter reports bounded safety as 1 when it completed the caller's
	// explicit frame bound without a resource interruption. The lower-level
	// scalable-BMC API preserves C's -1 result for finite-bound exhaustion.
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
