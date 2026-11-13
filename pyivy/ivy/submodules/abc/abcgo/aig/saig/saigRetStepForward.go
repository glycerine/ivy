package saig

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
)

// Saig_ManRetimeNodeFwd moves an AND across its two input registers.
// The caller marks autonomous nodes before attempting retiming.
func Saig_ManRetimeNodeFwd(p *aig.Graph, obj aig.Edge, makeBug int32) aig.Edge {
	if p == nil || Saig_ManRegNum(p) == 0 || !aig.Aig_ObjIsNode(p, obj) {
		panic("saig: invalid forward retiming node")
	}
	f0, f1 := aig.Aig_ObjFanin0(p, obj), aig.Aig_ObjFanin1(p, obj)
	if !Saig_ObjIsLo(p, f0) || !Saig_ObjIsLo(p, f1) ||
		(!aig.Aig_ObjIsTravIdCurrent(p, f0) && !aig.Aig_ObjIsTravIdCurrent(p, f1)) {
		return aig.InvalidEdge
	}
	in0 := aig.Aig_ObjChild0(p, Saig_ObjLoToLi(p, f0))
	in1 := aig.Aig_ObjChild0(p, Saig_ObjLoToLi(p, f1))
	c0, c1 := aig.Aig_ObjFaninC0(p, obj), aig.Aig_ObjFaninC1(p, obj)
	in0, in1 = aig.Aig_NotCond(in0, c0), aig.Aig_NotCond(in1, c1)
	if makeBug != 0 {
		fmt.Println("Introducing bug during retiming.")
		in1 = in1.Not()
	}
	compl := c0 & c1
	driver := aig.Aig_NotCond(p.AND(in0, in1), compl)
	p.AddOutput(driver)
	lo := p.NewInput()
	aig.Aig_ManSetRegNum(p, Saig_ManRegNum(p)+1)
	aig.Aig_ObjSetTravIdCurrent(p, lo)
	return aig.Aig_NotCond(lo, compl)
}
