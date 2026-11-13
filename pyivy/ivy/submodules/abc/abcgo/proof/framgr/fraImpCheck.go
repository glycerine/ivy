package framgr

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

// Fra_ImpCheckForNode checks the sorted implications whose largest original
// node ID is node. It returns the first position belonging to a later node.
func Fra_ImpCheckForNode(p *Fra_Man_t, imps *vec.Vec_Int_t, node aig.Edge, pos int32) int32 {
	for i := int(pos); i < vec.Vec_IntSize(imps); i++ {
		imp := vec.Vec_IntEntry(imps, i)
		if imp == 0 {
			continue
		}
		leftID, rightID := Fra_ImpLeft(imp), Fra_ImpRight(imp)
		maxID := leftID
		if rightID > maxID {
			maxID = rightID
		}
		if maxID < int32(node.NodeID()) {
			panic("fra: implications out of node order")
		}
		if maxID > int32(node.NodeID()) {
			return int32(i)
		}
		left, right := aig.Edge(leftID<<1), aig.Edge(rightID<<1)
		mappedLeft, mappedRight := Fra_ObjFraig(p, left, p.pPars.NFramesK), Fra_ObjFraig(p, right, p.pPars.NFramesK)
		if mappedLeft == aig.InvalidEdge || mappedRight == aig.InvalidEdge {
			panic("fra: missing implication mapping")
		}
		negLeft := p.pManAig.Node(left).Phase != mappedLeft.Complemented()
		negRight := p.pManAig.Node(right).Phase != mappedRight.Complemented()
		if aig.Aig_Regular(mappedLeft) == aig.Aig_Regular(mappedRight) {
			if negLeft == negRight || (mappedLeft.NodeID() == 0 && negLeft) {
				continue
			}
			p.pCla.FRefinement = 1
			vec.Vec_IntWriteEntry(imps, i, 0)
			continue
		}
		result := Fra_NodesAreImp(p, aig.Aig_Regular(mappedLeft), aig.Aig_Regular(mappedRight), fraBool(negLeft), fraBool(negRight))
		if result != 1 {
			p.pCla.FRefinement = 1
			if result == 0 {
				Fra_SmlResimulate(p)
			}
			if vec.Vec_IntEntry(imps, i) != 0 {
				fmt.Println("Fra_ImpCheckForNode(): Implication is not refined!")
				panic("fra: implication was not refined")
			}
		}
	}
	return int32(vec.Vec_IntSize(imps))
}
