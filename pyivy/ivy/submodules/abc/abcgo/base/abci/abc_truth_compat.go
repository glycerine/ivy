package abci

import (
	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/bdd/cudd"
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

// These compatibility entry points mirror ABC's C symbol names while keeping
// the underlying package implementations and representations intact.

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
