package cec

import "github.com/glycerine/abcgo/aig/gia"

// SweepCorrespondence implements gia.SweepCorrespondence for CEC parameters.
func (p *Cec_ParCor_t) SweepCorrespondence(graph *gia.Gia_Man_t) {
	Cec_ManLSCorrespondenceClasses(graph, p)
}
