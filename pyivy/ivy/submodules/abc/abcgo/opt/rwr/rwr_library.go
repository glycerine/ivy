package rwr

import (
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/extra/npn"
)

// rwrLibrary holds the initialized non-precompute branch of Rwr_ManStart.
// Each entry in graphs is indexed by its corresponding forest node ID.
type rwrLibrary struct {
	forest                                          []rwrForestNode
	classes                                         [][]int
	inverse                                         []uint16
	canonical                                       []uint16
	phases, permutations, classMap                  []byte
	practical                                       []bool
	graphs                                          []*dec.Dec_Graph_t
	scores, gains, added                            []int16
	nSubgraphs                                      int
	nCutsGood, nCutsBad                             int
	nNodesConsidered, nNodesRewritten, nNodesGained int
	classScores                                     [222]int
}

func rwrStartLibrary() *rwrLibrary {
	p := &rwrLibrary{}
	p.canonical, p.phases, p.permutations, p.classMap = npn.Extra_Truth4VarNPN()
	p.practical = rwrPracticalClasses()
	p.forest = rwrLoadForest(s_RwtAigSubgraphs)
	p.classes, p.inverse = rwrClassifyForest(p.forest, p.canonical, p.classMap)
	p.graphs = make([]*dec.Dec_Graph_t, len(p.forest))
	p.scores = make([]int16, len(p.forest))
	p.gains = make([]int16, len(p.forest))
	p.added = make([]int16, len(p.forest))
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

// rwrPracticalClasses follows Rwr_ManGetPractical, including the zero
// representative, which is also the sentinel at the end of the source table.
func rwrPracticalClasses() []bool {
	result := make([]bool, 1<<16)
	result[0] = true
	for _, truth := range s_RwrPracticalClasses[1:] {
		if truth == 0 {
			break
		}
		result[truth] = true
	}
	return result
}
