package fra

// Structural CNF generation is shared with the canonical FRAIG manager.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/framgr"
	"github.com/glycerine/abcgo/sat/bsat"
)

func fraLit(variable int32, negative bool) bsat.Literal {
	if negative {
		return bsat.Negative(variable)
	}
	return bsat.Positive(variable)
}
func Fra_AddClausesMux(p *Fra_Man_t, node aig.Edge) { framgr.Fra_AddClausesMux(p, node) }
func Fra_AddClausesSuper(p *Fra_Man_t, node aig.Edge, fanins []aig.Edge) {
	framgr.Fra_AddClausesSuper(p, node, fanins)
}
func Fra_CollectSuper(graph *aig.Graph, root aig.Edge, useMuxes bool) []aig.Edge {
	return framgr.Fra_CollectSuper(graph, root, useMuxes)
}
func Fra_ObjAddToFrontier(p *Fra_Man_t, node aig.Edge, frontier *[]aig.Edge) {
	framgr.Fra_ObjAddToFrontier(p, node, frontier)
}
func Fra_CnfNodeAddToSolver(p *Fra_Man_t, old, new aig.Edge) {
	framgr.Fra_CnfNodeAddToSolver(p, old, new)
}
