package nwk

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

func Nwl_ManComputeTotalSwitching(p *Nwk_Man_t) float32 {
	g := Nwk_ManStrash(p)
	probabilities, err := saig.SwitchingProbabilities(g, 48, 16, false)
	if err != nil {
		panic(err)
	}
	var total float32
	for _, o := range p.Objects {
		if o == nil {
			continue
		}
		e, ok := o.Copy.(aig.Edge)
		if ok {
			total += float32(len(o.Fanouts)) * probabilities[e.NodeID()]
		}
	}
	return total
}
