package cnf

import "github.com/glycerine/abcgo/aig/aig"

// MappingGraph is the graph whose cuts are owned by this manager.
func (p *Cnf_Man_t_) MappingGraph() *aig.Graph { return p.pManAig }

// TruthCover returns the packed four-variable SOP and its cube count.
func (p *Cnf_Man_t_) TruthCover(truth uint16) (string, int32) {
	return p.pSops[truth], int32(p.pSopSizes[truth])
}

// PositiveCover borrows the positive-polarity ISOP for a larger cut.
func (p *Cnf_Cut_t_) PositiveCover() *Vec_Int_t { return p.vIsop[1] }

func Cnf_DataLiteralNum(p *Cnf_Dat_t) int32 { return p.nLiterals }
