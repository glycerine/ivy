package fra

// SAT queries use the canonical FRAIG manager implementation.
import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/framgr"
)

func Fra_NodesAreEquiv(p *Fra_Man_t, old, new aig.Edge) int32 {
	return framgr.Fra_NodesAreEquiv(p, old, new)
}
func Fra_NodesAreImp(p *Fra_Man_t, old, new aig.Edge, negOld, negNew int32) int32 {
	return framgr.Fra_NodesAreImp(p, old, new, negOld, negNew)
}
func Fra_NodesAreClause(p *Fra_Man_t, old, new aig.Edge, negOld, negNew int32) int32 {
	return framgr.Fra_NodesAreClause(p, old, new, negOld, negNew)
}
func Fra_NodeIsConst(p *Fra_Man_t, node aig.Edge) int32 { return framgr.Fra_NodeIsConst(p, node) }
func Fra_SetActivityFactors_rec(p *Fra_Man_t, node aig.Edge, min, max int32, seen map[int]bool) int32 {
	return framgr.Fra_SetActivityFactors_rec(p, node, min, max, seen)
}
func Fra_SetActivityFactors(p *Fra_Man_t, old, new aig.Edge) int32 {
	return framgr.Fra_SetActivityFactors(p, old, new)
}
