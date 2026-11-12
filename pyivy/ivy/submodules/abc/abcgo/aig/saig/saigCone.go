package saig

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
)

// Saig_ManSupport returns the latch-input outputs in the transitive support
// of the supplied AIG outputs. Primary inputs and constants are not state
// support. The result is unique and follows first-visit traversal order.
func Saig_ManSupport(p *aig.Graph, outputs []aig.Edge) []aig.Edge {
	seenNodes := make(map[aig.Edge]struct{})
	seenRegs := make(map[aig.Edge]struct{})
	support := make([]aig.Edge, 0)
	nPis := int32(p.InputCount()) - aig.Aig_ManRegNum(p)
	var visit func(aig.Edge)
	visit = func(edge aig.Edge) {
		edge = aig.Aig_Regular(edge)
		if _, ok := seenNodes[edge]; ok {
			return
		}
		seenNodes[edge] = struct{}{}
		if aig.Aig_ObjIsConst1(p, edge) {
			return
		}
		if aig.Aig_ObjIsCi(p, edge) {
			if aig.Aig_ObjCioId(p, edge) >= nPis {
				li := aig.Aig_ManLi(p, aig.Aig_ObjCioId(p, edge)-nPis)
				if _, exists := seenRegs[li]; !exists {
					seenRegs[li] = struct{}{}
					support = append(support, li)
				}
			}
			return
		}
		if aig.Aig_ObjIsNode(p, edge) || aig.Aig_ObjIsBuf(p, edge) {
			visit(aig.Aig_ObjChild0(p, edge))
			if aig.Aig_ObjIsNode(p, edge) {
				visit(aig.Aig_ObjChild1(p, edge))
			}
		}
	}
	for _, output := range outputs {
		output = aig.Aig_Regular(output)
		if aig.Aig_ObjIsCo(p, output) {
			visit(aig.Aig_ObjFanin0(p, output))
		} else {
			visit(output)
		}
	}
	return support
}

// Saig_ManPrintConeOne prints successive state-support frontiers for one
// primary output object.
func Saig_ManPrintConeOne(p *aig.Graph, output aig.Edge) {
	frontier := Saig_ManSupport(p, []aig.Edge{output})
	previous := make(map[aig.Edge]struct{})
	total := make(map[aig.Edge]struct{})
	fmt.Printf("    PO %3d  ", aig.Aig_ObjCioId(p, output))
	for depth := 0; ; depth++ {
		var fresh, repeatedPrevious, repeatedOlder int
		for _, edge := range frontier {
			if _, ok := total[edge]; !ok {
				total[edge] = struct{}{}
				fresh++
			} else if _, ok := previous[edge]; ok {
				repeatedPrevious++
			} else {
				repeatedOlder++
			}
		}
		fmt.Printf("%d:%d %d=%d+%d+%d  ", depth, len(total), len(frontier), fresh, repeatedPrevious, repeatedOlder)
		if fresh == 0 {
			break
		}
		previous = make(map[aig.Edge]struct{}, len(frontier))
		for _, edge := range frontier {
			previous[edge] = struct{}{}
		}
		frontier = Saig_ManSupport(p, frontier)
	}
	fmt.Println()
}

// Saig_ManPrintCones reports the sequential support frontier of every primary
// output. Register outputs are excluded from the primary-output prefix.
func Saig_ManPrintCones(p *aig.Graph) {
	nPos := int32(p.OutputCount()) - aig.Aig_ManRegNum(p)
	for i := int32(0); i < nPos; i++ {
		Saig_ManPrintConeOne(p, aig.Aig_ManCo(p, i))
	}
}
