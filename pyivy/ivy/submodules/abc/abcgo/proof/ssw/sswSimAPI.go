package ssw

import "github.com/glycerine/abcgo/aig/aig"

func Ssw_SmlCheckNonConstOutputs(p *Ssw_Sml_t) int32 {
	outputs := p.pAig.Outputs()
	ordinary := len(outputs) - int(p.pAig.RegisterCount()) - int(aig.Aig_ManConstrNum(p.pAig))
	for _, obj := range outputs[:ordinary] {
		if Ssw_SmlNodeIsZero(p, obj) == 0 {
			return 1
		}
	}
	return 0
}
func Ssw_SmlSimulateComb(g *aig.Graph, words int32) *Ssw_Sml_t {
	p := Ssw_SmlStart(g, 0, 1, words)
	Ssw_SmlInitialize(p, 0)
	Ssw_SmlSimulateOne(p)
	return p
}
func Ssw_SmlSimulateSeq(g *aig.Graph, prefix, frames, words int32) *Ssw_Sml_t {
	p := Ssw_SmlStart(g, prefix, frames, words)
	Ssw_SmlInitialize(p, 1)
	Ssw_SmlSimulateOne(p)
	p.fNonConstOut = Ssw_SmlCheckNonConstOutputs(p)
	return p
}
func Ssw_SmlResimulateSeq(p *Ssw_Sml_t) {
	Ssw_SmlReinitialize(p)
	Ssw_SmlSimulateOne(p)
	p.fNonConstOut = Ssw_SmlCheckNonConstOutputs(p)
}
