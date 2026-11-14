package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Bmc_ChainFailOneOutput returns a bounded witness for the first property that
// fails. The former implementation advanced mutable GIA state after each
// witness; independent frame unrolling gives the same bounded reachability
// result without storing temporary marks in graph objects.
func Bmc_ChainFailOneOutput(graph *gia.Gia_Man_t, maxFrames, conflictLimit, verbose, veryVerbose int32) *util.Abc_Cex_t {
	if graph == nil {
		return nil
	}
	native := gia.Gia_ManToAig(graph, 0)
	options := Saig_ParBmc_t{NFramesMax: maxFrames, NConfLimit: conflictLimit, FVerbose: veryVerbose}
	if Saig_ManBmcScalable(native, &options) != 0 {
		return nil
	}
	return native.TakeSequentialCounterexample()
}

// Bmc_ChainTest checks each property with the same bounded engine and returns
// one output-indexed CEX vector. Nil slots mean no failure was found at the
// requested bound. It preserves source output numbering in every witness.
func Bmc_ChainTest(graph *gia.Gia_Man_t, maxFrames, conflictLimit, verbose, veryVerbose int32, cexes **vec.Vec_Ptr_t) int32 {
	if graph == nil {
		return -1
	}
	native := gia.Gia_ManToAig(graph, 0)
	properties := native.OutputCount() - int(native.RegisterCount())
	if properties < 0 {
		return -1
	}
	results := vec.Vec_PtrAlloc(properties)
	for output := 0; output < properties; output++ {
		model := checkOneProperty(native, output, maxFrames, conflictLimit, 0, 0)
		if model != nil {
			model.IPo = int32(output)
		}
		if model == nil {
			vec.Vec_PtrPush(results, nil)
		} else {
			vec.Vec_PtrPush(results, model)
		}
	}
	if cexes != nil {
		*cexes = results
	}
	return 0
}

func checkOneProperty(source *aig.Graph, output int, maxFrames, conflictLimit, timeout, timeoutPerOutput int32) *util.Abc_Cex_t {
	projected, err := projectAIGProperties(source, []int{output})
	if err != nil {
		return nil
	}
	options := Saig_ParBmc_t{
		NFramesMax:  maxFrames,
		NConfLimit:  conflictLimit,
		NTimeOut:    timeout,
		NTimeOutOne: timeoutPerOutput,
	}
	if Saig_ManBmcScalable(projected, &options) != 0 {
		return nil
	}
	model := projected.TakeSequentialCounterexample()
	if model != nil {
		model.IPo = int32(output)
	}
	return model
}
