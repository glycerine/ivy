package abci

import (
	"github.com/glycerine/abcgo/aig/hop"
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
	"github.com/glycerine/abcgo/opt/sim"
	"github.com/glycerine/abcgo/proof/fraig"
)

// These compatibility entry points mirror ABC's C symbol names while keeping
// the underlying package implementations and representations intact.

func Sim_ComputeFunSupp(network *Abc_Ntk_t, verbose int32) *Vec_Ptr_t {
	graph := baseabc.Abc_NtkStrashToGia(network)
	info := sim.Sim_ComputeFunSuppSourceWithSolver(graph, verbose != 0, func(input, output int) (bool, []uint32) {
		miter := Abc_NtkMiterForCofactors(network, int32(output), int32(input), -1)
		if miter == nil {
			panic("Sim_ComputeFunSupp: cofactor miter construction failed")
		}
		defer baseabc.Abc_NtkDelete(miter)

		params := &Fraig_Params_t{}
		fraig.Fraig_ParamsSetDefault(params)
		params.NSeconds = util.ABC_INFINITY
		params.FInternal = 1
		manager := Abc_NtkToFraig(miter, params, 0, 0).(*Fraig_Man_t)
		defer fraig.Fraig_ManFree(manager)
		fraig.Fraig_ManProveMiter(manager)
		result := fraig.Fraig_ManCheckMiter(manager)
		if result < 0 {
			panic("Sim_ComputeFunSupp: FRAIG left cofactor miter undecided")
		}
		if result == 1 {
			return false, nil
		}

		values := fraig.Fraig_ManReadModel(manager)
		if len(values) != len(network.Inputs()) {
			panic("Sim_ComputeFunSupp: satisfiable cofactor miter has no model")
		}
		assignment := make([]bool, len(network.Inputs()))
		packed := make([]uint32, sim.SIM_NUM_WORDS(len(assignment)))
		for i := range assignment {
			assignment[i] = values[i] != 0
			if assignment[i] {
				sim.Sim_SetBit(packed, i)
			}
		}
		before, err := graph.Evaluate(assignment)
		if err != nil {
			panic(err)
		}
		assignment[input] = !assignment[input]
		after, err := graph.Evaluate(assignment)
		if err != nil {
			panic(err)
		}
		if before[output] == after[output] {
			panic("Sim_ComputeFunSupp: FRAIG model failed cofactor verification")
		}
		return true, packed
	})
	result := vec.Vec_PtrAlloc(len(info))
	for _, support := range info {
		vec.Vec_PtrPush(result, support)
	}
	return result
}

func Cudd_CountMinterm(manager *cudd.DdManager, node *cudd.DdNode, nvars int32) float64 {
	return cudd.Cudd_CountMinterm(manager, node, nvars)
}

func Hop_ManConvertAigToTruth(manager *hop.Hop_Man_t, root *hop.Hop_Obj_t, n int32, scratch *vec.Vec_Int_t, msb int32) []uint32 {
	return hop.Hop_ManConvertAigToTruth(manager, root, n, scratch, msb)
}

func Hop_Regular(obj *hop.Hop_Obj_t) *hop.Hop_Obj_t { return hop.Hop_Regular(obj) }

func Hop_IsComplement(obj *hop.Hop_Obj_t) int32 { return hop.Hop_IsComplement(obj) }

func Extra_TruthNot(out, in []uint32, n int32) { extra.Extra_TruthNot(out, in, n) }

func Abc_Tt6Stretch(truth uint64, n int32) uint64 { return util.Abc_Tt6Stretch(truth, n) }

func Dau_DecTrySets(truth []uint64, n, verbose int32) { dau.Dau_DecTrySets(truth, n, verbose) }

func Kit_TruthPrintProfile(truth []uint32, nVars int32) { kit.Kit_TruthPrintProfile(truth, nVars) }

func Kit_DsdPrintCofactors(truth []uint32, nVars, level, verbose int32) {
	kit.Kit_DsdPrintCofactors(truth, nVars, level, verbose)
}

func Kit_DsdTest(truth []uint32, nVars int32) { kit.Kit_DsdTest(truth, nVars) }
