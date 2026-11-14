package abci

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/hop"
	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/bdc"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Bdc_Man_t = bdc.Bdc_Man_t
type Bdc_Fun_t = bdc.Bdc_Fun_t
type Bdc_Par_t = bdc.Bdc_Par_t

// Bdc_FunCopyHop is ABC's complemented-edge bridge from BDC nodes to Hop.
func Bdc_FunCopyHop(f *Bdc_Fun_t) *Hop_Obj_t {
	copy, ok := bdc.Bdc_FuncCopy(bdc.Bdc_Regular(f)).(*Hop_Obj_t)
	if !ok {
		panic("Bdc_FunCopyHop: BDC node has no Hop copy")
	}
	return hop.Hop_NotCond(copy, bdc.Bdc_IsComplement(f))
}

// Abc_NodeIfNodeResyn converts one Hop function to truth, decomposes it with
// BDC, and maps the resulting BDC graph back into the same Hop manager.
func Abc_NodeIfNodeResyn(p *Bdc_Man_t, pHop *Hop_Man_t, pRoot *Hop_Obj_t, nVars int32, vTruth *Vec_Int_t, puCare []uint32, dProb float32) *Hop_Obj_t {
	assert(nVars <= 16, "nVars <= 16")
	truth := hop.Hop_ManConvertAigToTruth(pHop, hop.Hop_Regular(pRoot), nVars, vTruth, 0)
	if hop.Hop_IsComplement(pRoot) != 0 {
		extra.Extra_TruthNot(truth, truth, nVars)
	}

	decompositionCare := puCare
	if dProb >= 0 {
		prob := 2 * dProb * (1 - dProb)
		assert(prob >= 0 && prob <= 0.5, "Prob >= 0.0 && Prob <= 0.5")
		if prob >= 0.4 {
			extra.Extra_TruthNot(puCare, puCare, nVars)
			if dProb > 0.5 { // more 1s than 0s
				extra.Extra_TruthOr(truth, truth, puCare, nVars)
			} else {
				extra.Extra_TruthSharp(truth, truth, puCare, nVars)
			}
			extra.Extra_TruthNot(puCare, puCare, nVars)
			decompositionCare = nil
		}
	}
	bdc.Bdc_ManDecompose(p, truth, decompositionCare, nVars, nil, 1000)

	bdc.Bdc_FuncSetCopy(bdc.Bdc_ManFunc(p, 0), hop.Hop_ManConst1(pHop))
	for i := int32(0); i < nVars; i++ {
		bdc.Bdc_FuncSetCopy(bdc.Bdc_ManFunc(p, i+1), hop.Hop_ManPi(pHop, i))
	}
	for i := nVars + 1; i < bdc.Bdc_ManNodeNum(p); i++ {
		f := bdc.Bdc_ManFunc(p, i)
		bdc.Bdc_FuncSetCopy(f, hop.Hop_And(pHop,
			Bdc_FunCopyHop(bdc.Bdc_FuncFanin0(f)),
			Bdc_FunCopyHop(bdc.Bdc_FuncFanin1(f))))
	}
	return Bdc_FunCopyHop(bdc.Bdc_ManRoot(p))
}

// Abc_NtkBidecResyn resynthesizes each eligible logic node using ABC's BDC
// decomposition and accumulates the change in Hop DAG size.
func Abc_NtkBidecResyn(pNtk *Abc_Ntk_t, fVerbose int32) {
	started := util.Abc_Clock()
	assert(abc.Abc_NtkIsLogic(pNtk) != 0, "Abc_NtkIsLogic(pNtk)")
	if abc.Abc_NtkToAig(pNtk) == 0 {
		return
	}

	nVarsMax := abc.Abc_NtkGetFaninMax(pNtk)
	if nVarsMax > 15 {
		if fVerbose != 0 {
			fmt.Println("Resynthesis is not performed for nodes with more than 15 inputs.")
		}
		nVarsMax = 15
	}
	// BDC's manager needs a two-variable search limit even when all actual
	// nodes have smaller support; individual decompositions retain their arity.
	if nVarsMax < 2 {
		nVarsMax = 2
	}
	pars := &bdc.Bdc_Par_t{NVarsMax: nVarsMax, FVerbose: fVerbose}
	manager := bdc.Bdc_ManAlloc(pars)
	defer bdc.Bdc_ManFree(manager)

	scratch := vec.Vec_IntAlloc(0)
	defer vec.Vec_IntFree(scratch)
	hopManager, ok := pNtk.FunctionManager.(*hop.Hop_Man_t)
	if !ok || hopManager == nil {
		panic("Abc_NtkBidecResyn: logic network has no Hop manager")
	}
	var gain int32
	for _, object := range pNtk.ObjectsOfKind(abc.ObjectNode) {
		nVars := abc.Abc_ObjFaninNum(object)
		if nVars > 15 {
			continue
		}
		root, ok := object.Data.(*hop.Hop_Obj_t)
		if !ok || root == nil {
			panic("Abc_NtkBidecResyn: node has no Hop function")
		}
		nodesBefore := hop.Hop_DagSize(root)
		object.Data = Abc_NodeIfNodeResyn(manager, hopManager, root, nVars, scratch, nil, -1)
		gain += nodesBefore - hop.Hop_DagSize(object.Data.(*hop.Hop_Obj_t))
	}
	if fVerbose != 0 {
		fmt.Printf("Total gain in AIG nodes = %d.  ", gain)
		fmt.Printf("Total runtime = %9.2f sec\n", float64(util.Abc_Clock()-started)/1e9)
	}
}
