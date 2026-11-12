package saig

import (
	"errors"
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

type dualRails [2]aig.Edge // [false, true]

// Saig_ManDupDual constructs a two-rail encoding. The first nDualPis primary
// inputs are unconstrained rails; other PIs use complementary rails. Latch
// initialization can be set per register with dcFlops, or uniformly with
// dualFlops when dcFlops is nil.
func Saig_ManDupDual(src *aig.Graph, dcFlops *vec.Vec_Int_t, nDualPis, dualFlops, miterFlops, complementPO, checkZero, checkOne int32) *aig.Graph {
	graph, err := DuplicateDualRail(src, dcFlops, int(nDualPis), dualFlops != 0, miterFlops != 0, complementPO != 0, checkZero != 0, checkOne != 0)
	if err != nil {
		panic(err)
	}
	return graph
}

// DuplicateDualRail builds the dual-rail miter directly on the native graph.
func DuplicateDualRail(src *aig.Graph, dcFlops *vec.Vec_Int_t, dualPI int, dualFlops, miterFlops, complementPO, checkZero, checkOne bool) (*aig.Graph, error) {
	if src == nil {
		return nil, errors.New("saig: nil dual-rail source")
	}
	inputs, outputs, nodes := src.Inputs(), src.Outputs(), src.Nodes()
	regs := int(aig.Aig_ManRegNum(src))
	piCount := len(inputs) - regs
	poCount := len(outputs) - regs
	if regs < 0 || piCount < 0 || poCount <= 0 || dualPI < 0 || dualPI > piCount {
		return nil, errors.New("saig: malformed dual-rail interface")
	}
	if dcFlops != nil && vec.Vec_IntSize(dcFlops) != regs {
		return nil, errors.New("saig: DC-flop vector must match register count")
	}
	dst := aig.NewGraph()
	dst.Name, dst.Spec = src.Name, src.Spec
	copy := make([]dualRails, len(nodes))
	copy[0] = dualRails{dst.Const0(), dst.Const1()}
	for i, input := range inputs {
		var falseRail, trueRail aig.Edge
		switch {
		case i < dualPI:
			falseRail, trueRail = dst.NewInput(), dst.NewInput()
		case i < piCount:
			trueRail = dst.NewInput()
			falseRail = trueRail.Not()
		default:
			falseRail, trueRail = dst.NewInput(), dst.NewInput()
			invertFalse := !dualFlops
			if dcFlops != nil {
				invertFalse = vec.Vec_IntEntry(dcFlops, i-piCount) == 0
			}
			if invertFalse {
				falseRail = falseRail.Not()
			}
		}
		copy[input.NodeID()] = dualRails{
			dst.AND(falseRail, trueRail.Not()),
			dst.AND(trueRail, falseRail.Not()),
		}
	}
	for id, node := range nodes {
		if !src.ValidEdge(aig.Edge(id)<<1) || id == 0 {
			continue
		}
		switch node.Kind {
		case aig.Input, aig.Output, aig.Buffer:
			if node.Kind != aig.Input {
				copy[id] = dualFanin(dst, copy, node.Fanins[0])
			}
		case aig.And:
			a, b := dualFanin(dst, copy, node.Fanins[0]), dualFanin(dst, copy, node.Fanins[1])
			copy[id] = dualRails{dst.OR(a[0], b[0]), dst.AND(a[1], b[1])}
		case aig.ExclusiveOr:
			a, b := dualFanin(dst, copy, node.Fanins[0]), dualFanin(dst, copy, node.Fanins[1])
			copy[id] = dualRails{
				dst.OR(dst.AND(a[1], b[0]), dst.AND(a[0], b[1])),
				dst.OR(dst.AND(a[1], b[1]), dst.AND(a[0], b[0])),
			}
		}
	}
	violation := dst.Const0()
	if miterFlops {
		for _, li := range outputs[poCount:] {
			rails := dualFanin(dst, copy, aig.Aig_ObjChild0(src, li))
			violation = dst.OR(violation, dualCheck(dst, rails, checkZero, checkOne))
		}
	} else {
		for _, po := range outputs[:poCount] {
			rails := dualFanin(dst, copy, aig.Aig_ObjChild0(src, po))
			violation = dst.OR(violation, dualCheck(dst, rails, checkZero, checkOne))
		}
	}
	if complementPO {
		violation = violation.Not()
	}
	dst.AddOutput(violation)
	for i, li := range outputs[poCount:] {
		falseRail := dualFanin(dst, copy, aig.Aig_ObjChild0(src, li))[0]
		invertFalse := !dualFlops
		if dcFlops != nil {
			invertFalse = vec.Vec_IntEntry(dcFlops, i) == 0
		}
		if invertFalse {
			falseRail = falseRail.Not()
		}
		dst.AddOutput(falseRail)
		dst.AddOutput(dualFanin(dst, copy, aig.Aig_ObjChild0(src, li))[1])
	}
	aig.Aig_ManSetRegNum(dst, int32(2*regs))
	return dst, nil
}

func dualFanin(dst *aig.Graph, mapped []dualRails, edge aig.Edge) dualRails {
	pair := mapped[edge.NodeID()]
	if edge.Complemented() {
		care := dst.OR(pair[0], pair[1])
		return dualRails{dst.AND(pair[1], care), dst.AND(pair[0], care)}
	}
	return pair
}

func dualCheck(dst *aig.Graph, rails dualRails, checkZero, checkOne bool) aig.Edge {
	switch {
	case checkZero:
		return dst.AND(rails[0], rails[1].Not())
	case checkOne:
		return dst.AND(rails[0].Not(), rails[1])
	default:
		return dst.AND(rails[0].Not(), rails[1].Not())
	}
}

// Saig_ManBlockPo forces all primary outputs to zero for the first nCycles
// cycles by adding a one-hot countdown chain.
func Saig_ManBlockPo(graph *aig.Graph, nCycles int32) {
	if graph == nil || nCycles <= 0 {
		panic(fmt.Errorf("saig: cycle count must be positive"))
	}
	outputs := graph.Outputs()
	poCount := len(outputs) - int(aig.Aig_ManRegNum(graph))
	if poCount < 0 {
		panic("saig: malformed sequential graph")
	}
	previous, condition := graph.Const1(), graph.Const1()
	for i := int32(0); i < nCycles; i++ {
		graph.AddOutput(previous)
		previous = graph.NewInput()
		condition = graph.AND(condition, previous)
	}
	for _, output := range outputs[:poCount] {
		driver := aig.Aig_ObjChild0(graph, output)
		aig.Aig_ObjPatchFanin0(graph, output, graph.AND(driver, condition))
	}
	aig.Aig_ManSetRegNum(graph, aig.Aig_ManRegNum(graph)+nCycles)
}
