package reo

// Port of src/bdd/reo/reoApi.c: implementation of API functions.

import (
	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

// Extra_ReorderInit initializes the reordering engine. The first argument
// is the max number of variables in all the BDDs to be reordered; the
// second, the max number of BDD nodes. These limits are soft.
func Extra_ReorderInit(nDdVarsMax, nNodesMax int) *reo_man {
	// allocate and clean the data structure
	p := &reo_man{}
	// resize the manager to meet user's needs
	reoResizeStructures(p, nDdVarsMax, nNodesMax, 100)
	// set the defaults
	p.fMinApl = 0
	p.fMinWidth = 0
	p.fRemapUp = 0
	p.fVerbose = 0
	p.fVerify = 0
	p.nIters = 1
	return p
}

// Extra_ReorderQuit disposes of the reordering engine.
func Extra_ReorderQuit(p *reo_man) {
	if p == nil {
		return
	}
	p.pTops = nil
	p.pSupp = nil
	p.pOrderInt = nil
	p.pWidthCofs = nil
	p.pMapToPlanes = nil
	p.pMapToDdVarsOrig = nil
	p.pMapToDdVarsFinal = nil
	p.pPlanes = nil
	p.pVarCosts = nil
	p.pLevelOrder = nil
	p.HTable = nil
	p.pRefNodes = nil
	reoUnitsStopDispenser(p)
	p.visited = nodeSet{}
}

// Extra_ReorderSetMinimizationType sets the type of DD minimization.
func Extra_ReorderSetMinimizationType(p *reo_man, fMinType reo_min_type) {
	if fMinType == REO_MINIMIZE_NODES {
		p.fMinWidth = 0
		p.fMinApl = 0
	} else if fMinType == REO_MINIMIZE_WIDTH {
		p.fMinWidth = 1
		p.fMinApl = 0
	} else if fMinType == REO_MINIMIZE_APL {
		p.fMinWidth = 0
		p.fMinApl = 1
	} else {
		assert(false, "valid minimization type")
	}
}

// Extra_ReorderSetRemapping sets the type of remapping performed by the engine.
func Extra_ReorderSetRemapping(p *reo_man, fRemapUp int) { p.fRemapUp = fRemapUp }

// Extra_ReorderSetIterations sets the number of iterations of sifting performed.
func Extra_ReorderSetIterations(p *reo_man, nIters int) { p.nIters = nIters }

// Extra_ReorderSetVerification sets the verification mode.
func Extra_ReorderSetVerification(p *reo_man, fVerify int) { p.fVerify = fVerify }

// Extra_ReorderSetVerbosity sets the verbosity level.
func Extra_ReorderSetVerbosity(p *reo_man, fVerbose int) { p.fVerbose = fVerbose }

// Extra_Reorder performs reordering of the function. If pOrder is not nil,
// returns the resulting variable permutation: permuting the result with it
// gives back Func. The returned DD is not referenced.
func Extra_Reorder(p *reo_man, dd *DdManager, Func *DdNode, pOrder []int) *DdNode {
	FuncRes := []*DdNode{nil}
	Extra_ReorderArray(p, dd, []*DdNode{Func}, FuncRes, 1, pOrder)
	cudd.Cudd_Deref(FuncRes[0])
	return FuncRes[0]
}

// Extra_ReorderArray performs reordering of the array of functions; the
// resulting DDs are returned referenced.
func Extra_ReorderArray(p *reo_man, dd *DdManager, Funcs []*DdNode, FuncsRes []*DdNode, nFuncs int, pOrder []int) {
	reoReorderArray(p, dd, Funcs, FuncsRes, nFuncs, pOrder)
}
