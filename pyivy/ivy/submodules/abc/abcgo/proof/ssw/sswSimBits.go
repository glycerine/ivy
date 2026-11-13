package ssw

import "github.com/glycerine/abcgo/aig/aig"

// The callback data is the graph owning the object edges.
func Ssw_SmlObjIsConstBit(data any, obj aig.Edge) int32 {
	n := data.(*aig.Graph).Node(obj)
	if n.Phase == n.MarkB {
		return 1
	}
	return 0
}
func Ssw_SmlObjsAreEqualBit(data any, a, b aig.Edge) int32 {
	g := data.(*aig.Graph)
	x, y := g.Node(a), g.Node(b)
	if (x.Phase == y.Phase) == (x.MarkB == y.MarkB) {
		return 1
	}
	return 0
}
func Ssw_SmlSavePattern0(p *Ssw_Man_t, _ int32) { clear(p.pPatWords) }
func Ssw_SmlSavePattern1(p *Ssw_Man_t, init int32) {
	for i := range p.pPatWords {
		p.pPatWords[i] = ^uint32(0)
	}
	if init == 0 {
		return
	}
	nRegs := int(p.pAig.RegisterCount())
	nPis := len(p.pAig.Inputs()) - nRegs
	for i := 0; i < nRegs; i++ {
		sswTogglePatternBit(p.pPatWords, nPis*int(p.nFrames)+i)
	}
}
