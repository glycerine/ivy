package saig

import "github.com/glycerine/abcgo/aig/aig"

// Saig_ManRetimeNodeBwd moves a register backwards across its driving gate.
// The original register remains in the graph until the caller replaces it.
func Saig_ManRetimeNodeBwd(p *aig.Graph, lo aig.Edge) aig.Edge {
	if p == nil || Saig_ManRegNum(p) == 0 || !Saig_ObjIsLo(p, lo) {
		panic("saig: invalid backward retiming register")
	}
	li := Saig_ObjLoToLi(p, lo)
	node := aig.Aig_ObjFanin0(p, li)
	if !aig.Aig_ObjIsNode(p, node) {
		return aig.InvalidEdge
	}
	c0 := aig.Aig_ObjFaninC0(p, node) ^ aig.Aig_ObjFaninC0(p, li)
	c1 := aig.Aig_ObjFaninC1(p, node) ^ aig.Aig_ObjFaninC0(p, li)
	p.AddOutput(aig.Aig_NotCond(aig.Aig_ObjFanin0(p, node), c0))
	p.AddOutput(aig.Aig_NotCond(aig.Aig_ObjFanin1(p, node), c1))
	lo0, lo1 := p.NewInput(), p.NewInput()
	aig.Aig_ManSetRegNum(p, Saig_ManRegNum(p)+2)
	return p.AND(aig.Aig_NotCond(lo0, c0), aig.Aig_NotCond(lo1, c1))
}
