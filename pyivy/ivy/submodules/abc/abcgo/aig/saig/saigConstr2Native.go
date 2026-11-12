package saig

import (
	"errors"
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

// FunctionalCandidates groups candidate constraints by the reverse frame
// level used by ABC's sequential induction check.
type FunctionalCandidates [][]aig.Edge

// UnrollProperty creates a combinational graph for the property output at the
// final frame. Initial latch values are unconstrained inputs.
func UnrollProperty(src *aig.Graph, frames int) (*aig.Graph, error) {
	if src == nil || frames < 1 {
		return nil, errors.New("saig: property unroll needs a graph and at least one frame")
	}
	outputs, regs := src.Outputs(), int(aig.Aig_ManRegNum(src))
	poCount := len(outputs) - regs
	if regs < 0 || poCount < 1 || regs > len(src.Inputs()) {
		return nil, errors.New("saig: malformed sequential graph")
	}
	return unrollGraph(src, frames, 0, true)
}

// InductionMiter creates one output per candidate, each encoding candidate at
// frame zero and its negation at frame one.
func InductionMiter(src *aig.Graph, candidates FunctionalCandidates) (*aig.Graph, error) {
	if src == nil {
		return nil, errors.New("saig: nil graph")
	}
	graph, maps, err := unrollFrames(src, 2)
	if err != nil {
		return nil, err
	}
	for _, group := range candidates {
		for _, candidate := range group {
			if !src.ValidEdge(candidate) {
				return nil, errors.New("saig: invalid candidate edge")
			}
			a := mappedEdge(maps, candidate, 0)
			b := mappedEdge(maps, candidate, 1).Not()
			graph.AddOutput(graph.AND(a, b))
		}
	}
	return graph, nil
}

// DetectFunctionalConstraints finds source nodes implied in either polarity
// by the property at each unrolled frame, then retains only inductive choices.
// SAT-unknown candidates are omitted because implication was not established.
func DetectFunctionalConstraints(src *aig.Graph, frameCount int, conflicts, propagations int64) (FunctionalCandidates, error) {
	if src == nil || frameCount < 1 {
		return nil, errors.New("saig: functional detection needs a graph and at least one frame")
	}
	poCount := len(src.Outputs()) - int(aig.Aig_ManRegNum(src))
	if poCount != 1 {
		return nil, fmt.Errorf("saig: functional detection requires one property output, got %d", poCount)
	}
	frames, maps, err := unrollFrames(src, frameCount)
	if err != nil {
		return nil, err
	}
	property := mappedEdge(maps, aig.Aig_ObjChild0(src, src.Outputs()[0]), frameCount-1)
	frames.AddOutput(property)
	cnf, err := encodeAIG(frames)
	if err != nil {
		return nil, err
	}
	propertyLit := cnf.lit(aig.Aig_ObjChild0(frames, frames.Outputs()[0]))
	cnf.solver.SetLimits(conflicts, propagations)
	if cnf.solver.Solve(propertyLit) != bsat.Satisfiable {
		return nil, nil
	}
	candidates := make(FunctionalCandidates, frameCount)
	nodes := src.Nodes()
	for frame := 0; frame < frameCount; frame++ {
		level := frameCount - 1 - frame
		for id, node := range nodes {
			if !src.ValidEdge(aig.Edge(id) << 1) {
				continue
			}
			if node.Kind != aig.Input && node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
				continue
			}
			mapped := maps[frame][id]
			if impliedBy(cnf, propertyLit, cnf.lit(mapped), conflicts, propagations) {
				candidates[level] = appendUniqueEdge(candidates[level], aig.Edge(id)<<1)
			} else if impliedBy(cnf, propertyLit, cnf.lit(mapped.Not()), conflicts, propagations) {
				candidates[level] = appendUniqueEdge(candidates[level], aig.Edge(id)<<1|1)
			}
		}
	}
	return FilterInductiveCandidates(src, candidates, conflicts, 0)
}

func impliedBy(cnf *aigCNF, property, candidate bsat.Literal, conflicts, propagations int64) bool {
	cnf.solver.SetLimits(conflicts, propagations)
	return cnf.solver.Solve(property, candidate^1) == bsat.Unsatisfiable
}

func appendUniqueEdge(edges []aig.Edge, edge aig.Edge) []aig.Edge {
	for _, existing := range edges {
		if existing == edge {
			return edges
		}
	}
	return append(edges, edge)
}

// UnfoldFunctionalConstraints adds one constraint output for each candidate
// and enough pipeline registers to delay it by its reverse-frame level.
func UnfoldFunctionalConstraints(src *aig.Graph, candidates FunctionalCandidates) (*aig.Graph, error) {
	if src == nil {
		return nil, errors.New("saig: nil graph")
	}
	if len(candidates) == 0 || candidateCount(candidates) == 0 {
		copy, _, err := duplicateSequentialGraph(src, nil)
		return copy, err
	}
	outputs := src.Outputs()
	regs := int(aig.Aig_ManRegNum(src))
	poCount := len(outputs) - regs
	if regs < 0 || regs > len(src.Inputs()) || poCount < int(aig.Aig_ManConstrNum(src)) {
		return nil, errors.New("saig: malformed sequential interface")
	}
	newFlops := 0
	for level, group := range candidates {
		for _, edge := range group {
			if !src.ValidEdge(edge) {
				return nil, errors.New("saig: invalid functional constraint candidate")
			}
			newFlops += level
		}
	}
	var pipeline []aig.Edge
	copy, _, err := duplicateSequentialGraph(src, func(dst *aig.Graph, copyEdge func(aig.Edge) aig.Edge) error {
		for _, output := range outputs[:poCount] {
			dst.AddOutput(copyEdge(aig.Aig_ObjChild0(src, output)))
		}
		for level, group := range candidates {
			for _, candidate := range group {
				candidateCopy := copyEdge(candidate)
				current := candidateCopy
				if level > 0 {
					pipeline = append(pipeline, candidateCopy)
				}
				for step := 0; step < level; step++ {
					current = dst.NewInput()
					if step+1 < level {
						pipeline = append(pipeline, current)
					}
				}
				dst.AddOutput(current)
			}
		}
		for _, li := range outputs[poCount:] {
			dst.AddOutput(copyEdge(aig.Aig_ObjChild0(src, li)))
		}
		for _, edge := range pipeline {
			dst.AddOutput(edge)
		}
		aig.Aig_ManSetRegNum(dst, int32(regs+newFlops))
		aig.Aig_ManSetConstrNum(dst, aig.Aig_ManConstrNum(src)+int32(candidateCount(candidates)))
		return nil
	})
	return copy, err
}

// FoldFunctionalConstraints gates each property with a sticky register that
// records whether any constraint has failed. If complemented is true, each
// constraint's complement is treated as its failure condition.
func FoldFunctionalConstraints(src *aig.Graph, complemented bool) (*aig.Graph, error) {
	if src == nil {
		return nil, errors.New("saig: nil graph")
	}
	regs := int(aig.Aig_ManRegNum(src))
	outputs := src.Outputs()
	poCount := len(outputs) - regs
	constraints := int(aig.Aig_ManConstrNum(src))
	if constraints == 0 {
		copy, _, err := duplicateSequentialGraph(src, nil)
		return copy, err
	}
	if regs < 0 || regs > len(src.Inputs()) || constraints >= poCount {
		return nil, errors.New("saig: malformed functional constraint interface")
	}
	copy, _, err := duplicateSequentialGraph(src, func(dst *aig.Graph, copyEdge func(aig.Edge) aig.Edge) error {
		violation := dst.Const0()
		for _, output := range outputs[poCount-constraints : poCount] {
			constraint := copyEdge(aig.Aig_ObjChild0(src, output))
			if complemented {
				constraint = constraint.Not()
			}
			violation = dst.OR(violation, constraint)
		}
		var healthy aig.Edge
		if regs > 0 {
			previous := dst.NewInput()
			healthy = dst.OR(violation, previous)
		} else {
			healthy = violation
		}
		for _, output := range outputs[:poCount-constraints] {
			dst.AddOutput(dst.AND(copyEdge(aig.Aig_ObjChild0(src, output)), healthy.Not()))
		}
		for _, li := range outputs[poCount:] {
			dst.AddOutput(copyEdge(aig.Aig_ObjChild0(src, li)))
		}
		if regs > 0 {
			dst.AddOutput(healthy)
			aig.Aig_ManSetRegNum(dst, int32(regs+1))
		}
		return nil
	})
	return copy, err
}

func candidateCount(candidates FunctionalCandidates) int {
	count := 0
	for _, group := range candidates {
		count += len(group)
	}
	return count
}

// FilterInductiveCandidates keeps candidates for which candidate(t) and
// !candidate(t+1) is proved unsatisfiable. SAT-unknown results are discarded.
func FilterInductiveCandidates(src *aig.Graph, candidates FunctionalCandidates, conflicts, propagations int64) (FunctionalCandidates, error) {
	miter, err := InductionMiter(src, candidates)
	if err != nil {
		return nil, err
	}
	cnf, err := encodeAIG(miter)
	if err != nil {
		return nil, err
	}
	filtered := make(FunctionalCandidates, len(candidates))
	output := 0
	for level, group := range candidates {
		for _, candidate := range group {
			assumption := cnf.lit(aig.Aig_ObjChild0(miter, miter.Outputs()[output]))
			cnf.solver.SetLimits(conflicts, propagations)
			status := cnf.solver.Solve(assumption)
			if status == bsat.Unsatisfiable {
				filtered[level] = append(filtered[level], candidate)
			}
			output++
		}
	}
	return filtered, nil
}

type aigCNF struct {
	solver *bsat.Solver
	vars   []int32
}

func encodeAIG(g *aig.Graph) (*aigCNF, error) {
	nodes := g.Nodes()
	cnf := &aigCNF{solver: bsat.NewSolver(), vars: make([]int32, len(nodes))}
	for id := range nodes {
		cnf.vars[id] = cnf.solver.NewVariable()
	}
	cnf.solver.AddClause(bsat.Positive(cnf.vars[0]))
	for id, node := range nodes {
		z := bsat.Positive(cnf.vars[id])
		switch node.Kind {
		case aig.Constant:
			continue
		case aig.Input:
		case aig.And:
			a, b := cnf.lit(node.Fanins[0]), cnf.lit(node.Fanins[1])
			cnf.solver.AddClause(bsat.Negative(cnf.vars[id]), a)
			cnf.solver.AddClause(bsat.Negative(cnf.vars[id]), b)
			cnf.solver.AddClause(z, a^1, b^1)
		case aig.ExclusiveOr:
			a, b := cnf.lit(node.Fanins[0]), cnf.lit(node.Fanins[1])
			cnf.solver.AddClause(a^1, b^1, z^1)
			cnf.solver.AddClause(a, b, z^1)
			cnf.solver.AddClause(a, b^1, z)
			cnf.solver.AddClause(a^1, b, z)
		case aig.Buffer, aig.Output:
			a := cnf.lit(node.Fanins[0])
			cnf.solver.AddClause(z^1, a)
			cnf.solver.AddClause(z, a^1)
		default:
			return nil, fmt.Errorf("saig: unsupported AIG node kind %d", node.Kind)
		}
	}
	return cnf, nil
}

func (c *aigCNF) lit(edge aig.Edge) bsat.Literal {
	lit := bsat.Positive(c.vars[edge.NodeID()])
	if edge.Complemented() {
		return lit ^ 1
	}
	return lit
}

// unrollGraph is the shared sequential-frame copier. propertyOutput < 0 means
// no outputs; otherwise only that primary output is exposed at the last frame.
func unrollGraph(src *aig.Graph, frameCount, propertyOutput int, includeProperty bool) (*aig.Graph, error) {
	dst, maps, err := unrollFrames(src, frameCount)
	if err != nil {
		return nil, err
	}
	if includeProperty {
		poCount := len(src.Outputs()) - int(aig.Aig_ManRegNum(src))
		if propertyOutput < 0 || propertyOutput >= poCount {
			return nil, fmt.Errorf("saig: property output %d outside [0,%d)", propertyOutput, poCount)
		}
		dst.AddOutput(mappedEdge(maps, aig.Aig_ObjChild0(src, src.Outputs()[propertyOutput]), frameCount-1))
	}
	return dst, nil
}

func unrollFrames(src *aig.Graph, frameCount int) (*aig.Graph, [][]aig.Edge, error) {
	if src == nil || frameCount < 1 {
		return nil, nil, errors.New("saig: invalid frame unroll")
	}
	inputs, outputs, nodes := src.Inputs(), src.Outputs(), src.Nodes()
	regs := int(aig.Aig_ManRegNum(src))
	poCount := len(outputs) - regs
	if regs < 0 || regs > len(inputs) || poCount < 0 {
		return nil, nil, errors.New("saig: malformed sequential interface")
	}
	piCount := len(inputs) - regs
	dst := aig.NewGraph()
	dst.Name, dst.Spec = src.Name, src.Spec
	maps := make([][]aig.Edge, frameCount)
	for frame := range maps {
		maps[frame] = make([]aig.Edge, len(nodes))
		maps[frame][0] = dst.Const1()
		for _, input := range inputs[:piCount] {
			maps[frame][input.NodeID()] = dst.NewInput()
		}
		if frame == 0 {
			for _, latch := range inputs[piCount:] {
				maps[frame][latch.NodeID()] = dst.NewInput()
			}
		}
	}
	for frame := range maps {
		for id, node := range nodes {
			if !src.ValidEdge(aig.Edge(id) << 1) {
				continue
			}
			switch node.Kind {
			case aig.Constant, aig.Input:
			case aig.And:
				maps[frame][id] = dst.AND(mappedEdge(maps, node.Fanins[0], frame), mappedEdge(maps, node.Fanins[1], frame))
			case aig.ExclusiveOr:
				maps[frame][id] = dst.NewXor(mappedEdge(maps, node.Fanins[0], frame), mappedEdge(maps, node.Fanins[1], frame))
			case aig.Buffer, aig.Output:
				maps[frame][id] = mappedEdge(maps, node.Fanins[0], frame)
			default:
				return nil, nil, fmt.Errorf("saig: unsupported node kind %d during unroll", node.Kind)
			}
		}
		if frame+1 < frameCount {
			for i := 0; i < regs; i++ {
				latch := inputs[piCount+i]
				li := outputs[poCount+i]
				maps[frame+1][latch.NodeID()] = mappedEdge(maps, aig.Aig_ObjChild0(src, li), frame)
			}
		}
	}
	return dst, maps, nil
}

func mappedEdge(maps [][]aig.Edge, edge aig.Edge, frame int) aig.Edge {
	mapped := maps[frame][edge.NodeID()]
	if edge.Complemented() {
		mapped = mapped.Not()
	}
	return mapped
}
