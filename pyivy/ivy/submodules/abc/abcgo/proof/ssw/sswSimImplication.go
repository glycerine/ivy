package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/bits"
)

func sswImplicationWords(p *Ssw_Sml_t, li, lo, cand aig.Edge) ([]uint32, []uint32, []uint32, bool) {
	if p.pAig.Node(lo).Phase {
		panic("ssw: latch output phase must be zero")
	}
	return Ssw_ObjSim(p, int32(li.NodeID())), Ssw_ObjSim(p, int32(lo.NodeID())), Ssw_ObjSim(p, int32(cand.NodeID())), p.pAig.Node(cand).Phase != cand.Complemented()
}
func Ssw_SmlCheckXorImplication(p *Ssw_Sml_t, li, lo, cand aig.Edge) int32 {
	x, y, c, invert := sswImplicationWords(p, li, lo, cand)
	for i := p.nWordsPref; i < p.nWordsTotal; i++ {
		value := c[i]
		if invert {
			value = ^value
		}
		if value&(x[i]^y[i]) != 0 {
			return 0
		}
	}
	return 1
}
func Ssw_SmlCountXorImplication(p *Ssw_Sml_t, li, lo, cand aig.Edge) int32 {
	x, y, c, invert := sswImplicationWords(p, li, lo, cand)
	var count int32
	for i := p.nWordsPref; i < p.nWordsTotal; i++ {
		value := c[i]
		if invert {
			value = ^value
		}
		count += int32(bits.OnesCount32(value & ^(x[i] ^ y[i])))
	}
	return count
}
