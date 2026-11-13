package rwr

import (
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/extra/npn"
)

// rwrLibrary holds the initialized non-precompute branch of Rwr_ManStart.
// Each entry in graphs is indexed by its corresponding forest node ID.
type rwrLibrary struct {
	forest                         []rwrForestNode
	classes                        [][]int
	inverse                        []uint16
	canonical                      []uint16
	phases, permutations, classMap []byte
	graphs                         []*dec.Dec_Graph_t
}

func rwrStartLibrary() *rwrLibrary {
	p := &rwrLibrary{}
	p.canonical, p.phases, p.permutations, p.classMap = npn.Extra_Truth4VarNPN()
	p.forest = rwrLoadForest(s_RwtAigSubgraphs)
	p.classes, p.inverse = rwrClassifyForest(p.forest, p.canonical, p.classMap)
	p.graphs = make([]*dec.Dec_Graph_t, len(p.forest))
	for _, nodes := range p.classes {
		for _, id := range nodes {
			g := rwrDecomposition(p.forest, id)
			if uint16(dec.Dec_GraphDeriveTruth(g)) != p.forest[id].truth {
				panic("rwr: incorrect decomposition truth")
			}
			p.graphs[id] = g
		}
	}
	return p
}
