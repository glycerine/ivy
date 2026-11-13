package cnf

import "github.com/glycerine/abcgo/aig/aig"

// SourceGraph returns the borrowed AIG from which this CNF was derived, or nil
// for a CNF constructed without an originating graph.
func (cnf *Cnf_Dat_t_) SourceGraph() *aig.Graph {
	if cnf == nil {
		return nil
	}
	return cnf.pMan
}
