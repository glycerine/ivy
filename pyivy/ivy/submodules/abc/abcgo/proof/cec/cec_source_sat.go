package cec

import "github.com/glycerine/abcgo/proof/cec/core"

// cecSourceSAT owns CEC query accounting around the shared source CNF builder.
type cecSourceSAT struct {
	cnf                                               *core.SourceCNF
	since                                             int32
	calls, satCount, unsatCount, undecCount, recycles int32
}

func newCecSourceSAT(g *Gia_Man_t, p *Cec_ParSat_t) *cecSourceSAT {
	return &cecSourceSAT{cnf: core.NewSourceCNF(g.SourceCNFGraph(), func() bool { return p.fPolarFlip != 0 })}
}
func (s *cecSourceSAT) reset()                  { s.cnf.Reset(); s.since = 0; s.recycles++ }
func (s *cecSourceSAT) literal(lit int32) int32 { return s.cnf.Literal(lit) }
func (s *cecSourceSAT) addClause(lits ...int32) { s.cnf.AddClause(lits...) }
func (s *cecSourceSAT) addNode(root int32) {
	if s.cnf.Solver() == nil {
		s.reset()
	}
	s.cnf.AddNode(root)
}
