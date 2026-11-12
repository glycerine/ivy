package saig

import (
	"errors"
	"fmt"
	"sort"

	"github.com/glycerine/abcgo/aig/aig"
)

// StructuralConstraint is the factorization recognized by ABC's structural
// constraint detector. PropertyNegations and ConstraintNegations retain the
// complement polarity used by the original supergate algorithm.
type StructuralConstraint struct {
	PropertyNegations   []aig.Edge
	ConstraintNegations []aig.Edge
	ConstantFalse       bool
}

// DetectStructuralConstraint recognizes an output whose AND supergate contains
// the negated next-state function of one latch. It returns false when the
// output does not have the source algorithm's required structural shape.
func DetectStructuralConstraint(g *aig.Graph, output int) (StructuralConstraint, bool, error) {
	if g == nil {
		return StructuralConstraint{}, false, errors.New("saig: nil graph")
	}
	regs := int(aig.Aig_ManRegNum(g))
	inputs := g.Inputs()
	outputs := g.Outputs()
	if regs < 0 || regs > len(inputs) || regs > len(outputs) {
		return StructuralConstraint{}, false, errors.New("saig: malformed sequential interface")
	}
	poCount := len(outputs) - regs
	if output < 0 || output >= poCount {
		return StructuralConstraint{}, false, fmt.Errorf("saig: output %d outside [0,%d)", output, poCount)
	}
	driver := aig.Aig_ObjChild0(g, outputs[output])
	if driver == g.Const0() {
		return StructuralConstraint{PropertyNegations: []aig.Edge{g.Const1()}, ConstantFalse: true}, true, nil
	}
	if driver.Complemented() || !aig.Aig_ObjIsAnd(g, driver) {
		return StructuralConstraint{}, false, nil
	}

	frontier := collectConstraintSuper(g, driver)
	if len(frontier) < 2 {
		return StructuralConstraint{}, false, nil
	}
	piCount := len(inputs) - regs
	for _, candidate := range frontier {
		reg := -1
		for i, lo := range inputs[piCount:] {
			if lo.NodeID() == candidate.NodeID() {
				reg = i
				break
			}
		}
		if reg < 0 {
			continue
		}
		liNode := outputs[poCount+reg]
		liDriver := aig.Aig_ObjChild0(g, liNode)
		if !liDriver.Complemented() {
			continue
		}
		liRegular := aig.Aig_Regular(liDriver)
		if !aig.Aig_ObjIsAnd(g, liRegular) {
			continue
		}
		flopSuper := collectConstraintSuper(g, liRegular)
		propertyNegations, ok := differenceIfSubset(frontier, flopSuper)
		if !ok {
			continue
		}
		// The C implementation abandons this output if a candidate latch term
		// matches a contained next-state supergate with the wrong phase or is
		// absent from that supergate.
		if candidate.Complemented() || !containsEdge(flopSuper, candidate) {
			return StructuralConstraint{}, false, nil
		}
		flopSuper = removeEdge(flopSuper, candidate)
		return StructuralConstraint{
			PropertyNegations:   propertyNegations,
			ConstraintNegations: flopSuper,
		}, true, nil
	}
	return StructuralConstraint{}, false, nil
}

// UnfoldStructuralConstraints splits compatible structural constraint factors
// into trailing constraint outputs. If detection fails or outputs disagree, it
// returns a faithful graph copy and reports changed=false.
func UnfoldStructuralConstraints(src *aig.Graph) (*aig.Graph, bool, error) {
	if src == nil {
		return nil, false, errors.New("saig: nil graph")
	}
	regs := int(aig.Aig_ManRegNum(src))
	outputs := src.Outputs()
	if regs < 0 || regs > len(src.Inputs()) || regs > len(outputs) {
		return nil, false, errors.New("saig: malformed sequential interface")
	}
	poCount := len(outputs) - regs
	splits := make([]StructuralConstraint, poCount)
	for i := 0; i < poCount; i++ {
		split, ok, err := DetectStructuralConstraint(src, i)
		if err != nil || !ok {
			copyGraph, _, copyErr := duplicateSequentialGraph(src, nil)
			return copyGraph, false, errors.Join(err, copyErr)
		}
		sort.Slice(split.PropertyNegations, func(i, j int) bool { return split.PropertyNegations[i] < split.PropertyNegations[j] })
		sort.Slice(split.ConstraintNegations, func(i, j int) bool { return split.ConstraintNegations[i] < split.ConstraintNegations[j] })
		splits[i] = split
	}
	chosen := []aig.Edge(nil)
	for _, split := range splits {
		if !split.ConstantFalse && len(split.ConstraintNegations) != 0 {
			chosen = split.ConstraintNegations
		}
	}
	for _, split := range splits {
		if split.ConstantFalse {
			continue
		}
		if !equalEdges(chosen, split.ConstraintNegations) {
			copyGraph, _, err := duplicateSequentialGraph(src, nil)
			return copyGraph, false, err
		}
	}
	graph, _, err := duplicateSequentialGraph(src, func(dst *aig.Graph, copyEdge func(aig.Edge) aig.Edge) error {
		for _, split := range splits {
			property := dst.Const1()
			for _, negated := range split.PropertyNegations {
				property = dst.AND(property, copyEdge(negated.Not()))
			}
			dst.AddOutput(property)
		}
		for _, constraint := range chosen {
			dst.AddOutput(copyEdge(constraint))
		}
		for _, li := range outputs[poCount:] {
			dst.AddOutput(copyEdge(aig.Aig_ObjChild0(src, li)))
		}
		aig.Aig_ManSetRegNum(dst, int32(regs))
		aig.Aig_ManSetConstrNum(dst, int32(len(chosen)))
		return nil
	})
	return graph, true, err
}

// FoldStructuralConstraints gates every property output with a new sticky
// register that records whether any selected constraint has failed.
func FoldStructuralConstraints(src *aig.Graph, constraintOutputs []int) (*aig.Graph, error) {
	if src == nil {
		return nil, errors.New("saig: nil graph")
	}
	regs := int(aig.Aig_ManRegNum(src))
	inputs, outputs := src.Inputs(), src.Outputs()
	if regs == 0 || regs > len(inputs) || regs > len(outputs) {
		return nil, errors.New("saig: folding requires a valid sequential graph")
	}
	poCount := len(outputs) - regs
	selected := make([]int, 0, len(constraintOutputs))
	selectedSet := make(map[int]struct{}, len(constraintOutputs))
	for _, i := range constraintOutputs {
		// Match the source API: output zero is reserved as the property.
		if i <= 0 || i >= poCount {
			return nil, fmt.Errorf("saig: constraint output %d outside [1,%d)", i, poCount)
		}
		if _, exists := selectedSet[i]; !exists {
			selectedSet[i] = struct{}{}
			selected = append(selected, i)
		}
	}
	graph, _, err := duplicateSequentialGraph(src, func(dst *aig.Graph, copyEdge func(aig.Edge) aig.Edge) error {
		violation := dst.Const0()
		for _, i := range selected {
			driver := aig.Aig_ObjChild0(src, outputs[i])
			violation = dst.OR(violation, copyEdge(driver))
		}
		flopOut := dst.NewInput()
		flopIn := dst.OR(violation, flopOut)
		for _, out := range outputs[:poCount] {
			dst.AddOutput(dst.AND(copyEdge(aig.Aig_ObjChild0(src, out)), flopIn.Not()))
		}
		for _, li := range outputs[poCount:] {
			dst.AddOutput(copyEdge(aig.Aig_ObjChild0(src, li)))
		}
		dst.AddOutput(flopIn)
		aig.Aig_ManSetRegNum(dst, int32(regs+1))
		aig.Aig_ManSetConstrNum(dst, aig.Aig_ManConstrNum(src))
		return nil
	})
	return graph, err
}

func collectConstraintSuper(g *aig.Graph, root aig.Edge) []aig.Edge {
	var result []aig.Edge
	seen := make(map[aig.Edge]struct{})
	var visit func(aig.Edge)
	visit = func(e aig.Edge) {
		if e.Complemented() || !aig.Aig_ObjIsAnd(g, e) {
			term := e.Not()
			if _, ok := seen[term]; !ok {
				seen[term] = struct{}{}
				result = append(result, term)
			}
			return
		}
		visit(aig.Aig_ObjChild0(g, e))
		visit(aig.Aig_ObjChild1(g, e))
	}
	node := aig.Aig_ObjChild0(g, root)
	visit(node)
	node = aig.Aig_ObjChild1(g, root)
	visit(node)
	return result
}

func differenceIfSubset(all, subset []aig.Edge) ([]aig.Edge, bool) {
	set := make(map[aig.Edge]struct{}, len(all))
	for _, e := range all {
		set[e] = struct{}{}
	}
	for _, e := range subset {
		if _, ok := set[e]; !ok {
			return nil, false
		}
	}
	result := make([]aig.Edge, 0, len(all)-len(subset))
	remove := make(map[aig.Edge]int, len(subset))
	for _, e := range subset {
		remove[e]++
	}
	for _, e := range all {
		if remove[e] > 0 {
			remove[e]--
		} else {
			result = append(result, e)
		}
	}
	return result, true
}

func containsEdge(edges []aig.Edge, target aig.Edge) bool {
	for _, e := range edges {
		if e == target {
			return true
		}
	}
	return false
}

func removeEdge(edges []aig.Edge, target aig.Edge) []aig.Edge {
	for i, e := range edges {
		if e == target {
			return append(edges[:i:i], edges[i+1:]...)
		}
	}
	return edges
}

func cloneEdges(edges []aig.Edge) []aig.Edge { return append([]aig.Edge(nil), edges...) }

func equalEdges(a, b []aig.Edge) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func duplicateSequentialGraph(src *aig.Graph, addOutputs func(*aig.Graph, func(aig.Edge) aig.Edge) error) (*aig.Graph, bool, error) {
	dst := aig.NewGraph()
	dst.Name, dst.Spec = src.Name, src.Spec
	nodes, inputs, outputs := src.Nodes(), src.Inputs(), src.Outputs()
	mapping := make([]aig.Edge, len(nodes))
	known := make([]bool, len(nodes))
	mapping[0], known[0] = dst.Const1(), true
	for _, input := range inputs {
		copy := dst.NewInput()
		mapping[input.NodeID()], known[input.NodeID()] = copy, true
	}
	var copyEdge func(aig.Edge) aig.Edge
	copyEdge = func(e aig.Edge) aig.Edge {
		id := e.NodeID()
		if id >= len(nodes) || !src.ValidEdge(e) {
			panic("saig: invalid edge while duplicating graph")
		}
		if !known[id] {
			n := nodes[id]
			switch n.Kind {
			case aig.Input:
				panic("saig: unmapped input")
			case aig.Output, aig.Buffer:
				mapping[id] = copyEdge(n.Fanins[0])
			case aig.And:
				mapping[id] = dst.AND(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
			case aig.ExclusiveOr:
				mapping[id] = dst.NewXor(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
			case aig.Constant:
				mapping[id] = dst.Const1()
			default:
				panic("saig: unsupported node during graph duplication")
			}
			known[id] = true
		}
		mapped := mapping[id]
		if e.Complemented() {
			mapped = mapped.Not()
		}
		return mapped
	}
	if addOutputs == nil {
		for _, output := range outputs {
			dst.AddOutput(copyEdge(aig.Aig_ObjChild0(src, output)))
		}
		aig.Aig_ManSetRegNum(dst, aig.Aig_ManRegNum(src))
		aig.Aig_ManSetConstrNum(dst, aig.Aig_ManConstrNum(src))
		return dst, false, nil
	}
	if err := addOutputs(dst, copyEdge); err != nil {
		return nil, false, err
	}
	return dst, true, nil
}
