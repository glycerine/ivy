package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManMultiProve checks all properties independently through the native
// BMC engine. The output-indexed witness list remains available on the GIA.
// It returns 0 when at least one property is disproved and -1 when all are
// bounded-safe or solving stopped without a witness.
func Gia_ManMultiProve(graph *gia.Gia_Man_t, params *Bmc_MulPar_t) int32 {
	if graph == nil || params == nil {
		return -1
	}
	native := gia.Gia_ManToAig(graph, 0)
	models := multiResultValues(Gia_ManMultiProveAig(native, params))
	if err := graph.SetSequentialCounterexamples(models); err != nil {
		return -1
	}
	for _, model := range models {
		if model != nil {
			return 0
		}
	}
	return -1
}

// Gia_ManMultiProveAig returns a vector with one slot per property. Each slot
// contains an owned sequential witness or nil when no bounded witness exists.
func Gia_ManMultiProveAig(graph *aig.Graph, params *Bmc_MulPar_t) *vec.Vec_Ptr_t {
	if graph == nil || params == nil {
		return nil
	}
	properties := graph.OutputCount() - int(graph.RegisterCount())
	if properties < 0 {
		return nil
	}
	results := vec.Vec_PtrAlloc(properties)
	for output := 0; output < properties; output++ {
		timeout := params.TimeOutLoc
		if timeout <= 0 {
			timeout = params.TimeOutGlo
		}
		model := checkOneProperty(graph, output, 0, 0, timeout, params.TimePerOut)
		if model == nil {
			vec.Vec_PtrPush(results, nil)
		} else {
			vec.Vec_PtrPush(results, model)
		}
	}
	return results
}

// ValuesAsCex returns an owned slice of the witness slots stored in a vector.
func multiResultValues(p *vec.Vec_Ptr_t) []*util.Abc_Cex_t {
	if p == nil {
		return nil
	}
	result := make([]*util.Abc_Cex_t, vec.Vec_PtrSize(p))
	for i := range result {
		if model, ok := vec.Vec_PtrEntry(p, i).(*util.Abc_Cex_t); ok {
			result[i] = model
		}
	}
	return result
}
