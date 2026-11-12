package llb

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

// Llb_Mnn_t carries the graph and parameters for the alternative nonlinear
// reachability entry point. Its symbolic state representation is the same
// canonical BDD set used by Llb_ManReachability.
type Llb_Mnn_t struct {
	pInit, pAig                         *Aig_Man_t
	pPars                               *Gia_ParLlb_t
	dd, ddG, ddR                        *DdManager
	vLeaves, vRoots                     *Vec_Ptr_t
	pVars2Q, pOrderL, pOrderL2, pOrderG *Vec_Int_t
	vCs2Glo, vNs2Glo, vGlo2Cs, vGlo2Ns  *Vec_Int_t
}

// Llb_NonlinComputeInitState represents the all-zero register valuation.
func Llb_NonlinComputeInitState(g *Aig_Man_t, dd *DdManager) *DdNode {
	if g == nil || dd == nil {
		return nil
	}
	state := core.Cudd_ReadOne(dd)
	inputs := g.Inputs()
	registers := int(g.RegisterCount())
	for i := len(inputs) - registers; i < len(inputs); i++ {
		if i < 0 {
			continue
		}
		state = core.Cudd_bddAnd(dd, state, core.Cudd_Not(core.Cudd_bddIthVar(dd, inputs[i].NodeID())))
		if state == nil {
			return nil
		}
	}
	return state
}

// Llb_NonlinPrepareVarMap builds current and next register-variable maps.
func Llb_NonlinPrepareVarMap(p *Llb_Mnn_t) {
	if p == nil || p.pAig == nil {
		return
	}
	registers := int(p.pAig.RegisterCount())
	inputs, outputs := p.pAig.Inputs(), p.pAig.Outputs()
	p.vCs2Glo = vec.Vec_IntAlloc(registers)
	p.vNs2Glo = vec.Vec_IntAlloc(registers)
	p.vGlo2Cs = vec.Vec_IntAlloc(registers)
	p.vGlo2Ns = vec.Vec_IntAlloc(registers)
	for i := 0; i < registers; i++ {
		cs, ns := inputs[len(inputs)-registers+i].NodeID(), outputs[len(outputs)-registers+i].NodeID()
		vec.Vec_IntPush(p.vCs2Glo, int32(cs))
		vec.Vec_IntPush(p.vNs2Glo, int32(ns))
		vec.Vec_IntPush(p.vGlo2Cs, int32(cs))
		vec.Vec_IntPush(p.vGlo2Ns, int32(ns))
	}
}

func Llb_NonlinFindBestVar(dd *DdManager, function *DdNode, g *Aig_Man_t) int32 {
	if dd == nil || function == nil || g == nil {
		return -1
	}
	best, score := int32(-1), -1
	inputs := g.Inputs()
	firstLatch := len(inputs) - int(g.RegisterCount())
	for i := firstLatch; i < len(inputs); i++ {
		id := inputs[i].NodeID()
		refs := aig.Aig_ObjRefs(g, inputs[i])
		if int(refs) > score {
			best, score = int32(i-firstLatch), int(refs)
		}
		_ = id
	}
	return best
}
func Llb_NonlinTrySubsetting(dd *DdManager, function *DdNode) {}

// Llb_NonlinReachability uses exact symbolic image computation and returns the
// same status values as the primary reachability engine.
func Llb_NonlinReachability(p *Llb_Mnn_t) int32 {
	if p == nil || p.pPars == nil || p.pAig == nil {
		return -1
	}
	manager := &Llb_Man_t_{pAig: p.pAig, pAigGlo: p.pInit, pPars: p.pPars}
	return Llb_ManReachability(manager, nil, nil)
}

func Llb_MnnStart(initial, working *Aig_Man_t, params *Gia_ParLlb_t) *Llb_Mnn_t {
	if initial == nil || working == nil || params == nil {
		return nil
	}
	p := &Llb_Mnn_t{pInit: initial, pAig: working, pPars: params}
	Llb_NonlinPrepareVarMap(p)
	return p
}
func Llb_MnnStop(p *Llb_Mnn_t) {
	if p != nil {
		p.dd = nil
		p.ddG = nil
		p.ddR = nil
	}
}
func Llb_NonlinExperiment(g *Aig_Man_t, maxNodes int32) {
	params := &Gia_ParLlb_t{nIterMax: 100000, nBddMax: maxNodes, fSilent: 1}
	Llb_NonlinCoreReach(g, params)
}
func Llb_NonlinCoreReach(g *Aig_Man_t, params *Gia_ParLlb_t) int32 {
	if g == nil || params == nil {
		return -1
	}
	return Llb_NonlinReachability(Llb_MnnStart(g, g, params))
}
