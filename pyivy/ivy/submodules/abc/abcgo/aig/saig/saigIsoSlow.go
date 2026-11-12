package saig

import "github.com/glycerine/abcgo/aig/aig"

// Saig_ManFindIsoPerm keeps the original SAIG entry point while delegating to
// the Go graph-isomorphism implementation. The original file's Iso_Man_t
// arrays, prime tables, manual hash chains, and timing counters were private
// implementation details; deterministic Go color refinement replaces them.
func Saig_ManFindIsoPerm(pAig *aig.Graph, fVerbose int32) []int32 {
	return isoFindInputPermutation(pAig, fVerbose)
}

// Iso_ManObjCount reports unique internal nodes and complemented fanin edges
// in the cone rooted at root. It preserves the original helper's traversal
// semantics without leaving traversal marks on the graph.
func Iso_ManObjCount(p *aig.Graph, root aig.Edge) (nodes, complementedEdges int32) {
	visited := make(map[aig.Edge]struct{})
	var visit func(aig.Edge)
	visit = func(edge aig.Edge) {
		edge = aig.Aig_Regular(edge)
		if aig.Aig_ObjIsConst1(p, edge) || aig.Aig_ObjIsCi(p, edge) {
			return
		}
		if _, ok := visited[edge]; ok {
			return
		}
		visited[edge] = struct{}{}
		if aig.Aig_ObjIsBuf(p, edge) {
			fanin := aig.Aig_ObjChild0(p, edge)
			if aig.Aig_IsComplement(fanin) != 0 {
				complementedEdges++
			}
			visit(fanin)
			nodes++
			return
		}
		fanin0, fanin1 := aig.Aig_ObjChild0(p, edge), aig.Aig_ObjChild1(p, edge)
		if aig.Aig_IsComplement(fanin0) != 0 {
			complementedEdges++
		}
		if aig.Aig_IsComplement(fanin1) != 0 {
			complementedEdges++
		}
		visit(fanin0)
		visit(fanin1)
		nodes++
	}
	visit(root)
	return nodes, complementedEdges
}
