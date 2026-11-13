package abs

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

type Saig_RefMan_t struct {
	pAig, pFrames     *aig.Graph
	pCex              *util.Abc_Cex_t
	nInputs, fVerbose int32
	vMapPiF2A         *vec.Vec_Int_t
	phase             []bool
}

// Saig_RefManOrderLiterals groups frame literals by their underlying input.
func Saig_RefManOrderLiterals(p *Saig_RefMan_t, varToPI, assumptions *vec.Vec_Int_t) *vec.Vec_Vec_t {
	result := vec.Vec_VecAlloc(100)
	groups := make(map[int32]int)
	for _, lit := range vec.Vec_IntArray(assumptions) {
		pi := vec.Vec_IntEntry(varToPI, int(lit>>1))
		if pi < 0 || int(pi) >= len(p.pFrames.Inputs()) {
			panic("invalid frame input")
		}
		input := vec.Vec_IntEntry(p.vMapPiF2A, int(2*pi))
		group, ok := groups[input]
		if !ok {
			group = vec.Vec_VecSize(result)
			groups[input] = group
		}
		vec.Vec_VecPushInt(result, group, lit)
	}
	return result
}

// Saig_RefManCreateCex maps SAT assumption literals back to care bits.
func Saig_RefManCreateCex(p *Saig_RefMan_t, varToPI, assumptions *vec.Vec_Int_t) *util.Abc_Cex_t {
	care := util.Abc_CexDup(p.pCex, p.pCex.NRegs)
	clear(care.PData)
	for _, lit := range vec.Vec_IntArray(assumptions) {
		pi := vec.Vec_IntEntry(varToPI, int(lit>>1))
		if pi < 0 || int(pi) >= len(p.pFrames.Inputs()) {
			panic("invalid frame input")
		}
		input, frame := vec.Vec_IntEntry(p.vMapPiF2A, int(2*pi)), vec.Vec_IntEntry(p.vMapPiF2A, int(2*pi+1))
		util.Abc_InfoSetBit(care.PData, care.NRegs+care.NPis*frame+input)
	}
	return care
}

// Saig_RefManReason2Inputs returns distinct original inputs in reason order.
func Saig_RefManReason2Inputs(p *Saig_RefMan_t, reasons *vec.Vec_Int_t) *vec.Vec_Int_t {
	result := vec.Vec_IntAlloc(len(p.pAig.Inputs()))
	visited := make(map[int32]bool)
	for _, framePI := range vec.Vec_IntArray(reasons) {
		input := vec.Vec_IntEntry(p.vMapPiF2A, int(2*framePI))
		if input < 0 || int(input) >= len(p.pAig.Inputs()) {
			panic("invalid original input")
		}
		if !visited[input] {
			vec.Vec_IntPush(result, input)
			visited[input] = true
		}
	}
	return result
}

// Saig_RefManReason2Cex maps frame input IDs to a counterexample care mask.
func Saig_RefManReason2Cex(p *Saig_RefMan_t, reasons *vec.Vec_Int_t) *util.Abc_Cex_t {
	care := util.Abc_CexDup(p.pCex, p.pCex.NRegs)
	clear(care.PData)
	for _, pi := range vec.Vec_IntArray(reasons) {
		if pi < 0 || int(pi) >= len(p.pFrames.Inputs()) {
			panic("invalid frame input")
		}
		input, frame := vec.Vec_IntEntry(p.vMapPiF2A, int(2*pi)), vec.Vec_IntEntry(p.vMapPiF2A, int(2*pi+1))
		util.Abc_InfoSetBit(care.PData, care.NRegs+care.NPis*frame+input)
	}
	return care
}

// Saig_RefManSetPhases simulates the unrolled AIG with the counterexample's
// inputs, replacing don't-care inputs with value1. Phases belong to this
// refinement manager so concurrent users of the graph do not overwrite them.
func Saig_RefManSetPhases(p *Saig_RefMan_t, care *util.Abc_Cex_t, value1 int32) int32 {
	g := p.pFrames
	p.phase = make([]bool, len(g.Nodes()))
	p.phase[g.Const1().NodeID()] = true
	for i, e := range g.Inputs() {
		input, frame := vec.Vec_IntEntry(p.vMapPiF2A, 2*i), vec.Vec_IntEntry(p.vMapPiF2A, 2*i+1)
		bit := p.pCex.NRegs + p.pCex.NPis*frame + input
		p.phase[e.NodeID()] = util.Abc_InfoHasBit(p.pCex.PData, bit)
		if care != nil && !util.Abc_InfoHasBit(care.PData, bit) {
			p.phase[e.NodeID()] = value1 != 0
		}
	}
	value := func(e aig.Edge) bool { return p.phase[e.NodeID()] != e.Complemented() }
	for _, e := range g.LiveEdges() {
		n := g.Node(e)
		switch n.Kind {
		case aig.And:
			p.phase[e.NodeID()] = value(n.Fanins[0]) && value(n.Fanins[1])
		case aig.Buffer:
			p.phase[e.NodeID()] = value(n.Fanins[0])
		case aig.ExclusiveOr:
			panic("refinement expects an AND-inverter graph")
		}
	}
	for _, e := range g.Outputs() {
		p.phase[e.NodeID()] = value(g.Node(e).Fanins[0])
	}
	if p.phase[g.Outputs()[0].NodeID()] {
		return 1
	}
	return 0
}
