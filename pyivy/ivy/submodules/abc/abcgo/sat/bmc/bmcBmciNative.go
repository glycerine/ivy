package bmc

import (
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/cnf"
)

// CompareInitialStatesAIG checks whether two register initializations produce
// the same register state for every shared primary-input sequence, through the
// requested number of transitions. It returns 1 if equivalent, 0 if a state
// difference is reachable, and -1 for invalid input or a solver timeout.
func CompareInitialStatesAIG(graph *aig.Graph, init0, init1 []bool, frames int32, timeout time.Duration) int32 {
	values0, values1 := make([]int32, len(init0)), make([]int32, len(init1))
	for i, value := range init0 {
		if value {
			values0[i] = 1
		}
	}
	for i, value := range init1 {
		if value {
			values1[i] = 1
		}
	}
	return compareInitialStateLiteralsAIG(graph, values0, values1, frames, timeout)
}

// compareInitialStateLiteralsAIG follows GIA literal conventions for initial
// states: 0 is false, 1 is true, and values above 1 request independent inputs.
func compareInitialStateLiteralsAIG(graph *aig.Graph, init0, init1 []int32, frames int32, timeout time.Duration) int32 {
	if graph == nil || frames < 0 {
		return -1
	}
	regs := int(aig.Aig_ManRegNum(graph))
	inputs := graph.Inputs()
	outputs := graph.Outputs()
	if regs < 0 || regs > len(inputs) || regs > len(outputs) || len(init0) != regs || len(init1) != regs {
		return -1
	}
	pis := len(inputs) - regs
	combined := aig.NewGraph()
	state0, state1 := make([]aig.Edge, regs), make([]aig.Edge, regs)
	for i := 0; i < regs; i++ {
		state0[i] = initialAIGEdge(combined, init0[i])
		state1[i] = initialAIGEdge(combined, init1[i])
	}
	for frame := int32(0); frame < frames; frame++ {
		shared := make([]aig.Edge, pis)
		for i := range shared {
			shared[i] = combined.NewInput()
		}
		_, next0 := cloneSequentialFrame(combined, graph, shared, state0, pis, regs)
		_, next1 := cloneSequentialFrame(combined, graph, shared, state1, pis, regs)
		for i := 0; i < regs; i++ {
			combined.AddOutput(combined.XOR(next0[i], next1[i]))
		}
		state0, state1 = next0, next1
	}
	data, err := cnf.EncodeAIG(combined, 0)
	if err != nil {
		return -1
	}
	solver := bsat.NewSolver()
	for solver.NumVars() < int32(data.Variables) {
		solver.NewVariable()
	}
	for _, clause := range data.Clauses {
		lits := make([]bsat.Literal, len(clause))
		for i, lit := range clause {
			if lit&1 != 0 {
				lits[i] = bsat.Negative(lit >> 1)
			} else {
				lits[i] = bsat.Positive(lit >> 1)
			}
		}
		if !solver.AddClause(lits...) {
			return 1
		}
	}
	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}
	for _, output := range combined.Outputs() {
		if !deadline.IsZero() {
			solver.SetDeadline(deadline)
		}
		driver := combined.Node(output).Fanins[0]
		if driver.NodeID() == 0 {
			if !driver.Complemented() {
				return 0
			}
			continue
		}
		variable := data.VarNums[driver.NodeID()]
		if variable < 0 {
			return -1
		}
		literal := bsat.Positive(variable)
		if driver.Complemented() {
			literal = bsat.Negative(variable)
		}
		switch solver.Solve(literal) {
		case bsat.Satisfiable:
			return 0
		case bsat.Unsatisfiable:
		case bsat.Unknown:
			return -1
		}
	}
	return 1
}

func initialAIGEdge(graph *aig.Graph, value int32) aig.Edge {
	switch value {
	case 0:
		return graph.Const0()
	case 1:
		return graph.Const1()
	default:
		return graph.NewInput()
	}
}

func cloneSequentialFrame(dst, src *aig.Graph, primaryInputs, state []aig.Edge, pis, regs int) ([]aig.Edge, []aig.Edge) {
	inputs := src.Inputs()
	outputs := src.Outputs()
	mapEdges := make([]aig.Edge, len(src.Nodes()))
	for i, in := range inputs {
		if i < pis {
			mapEdges[in.NodeID()] = primaryInputs[i]
		} else {
			mapEdges[in.NodeID()] = state[i-pis]
		}
	}
	for id, node := range src.Nodes() {
		switch node.Kind {
		case aig.Constant:
			mapEdges[id] = dst.Const1()
		case aig.Input:
		case aig.And:
			mapEdges[id] = dst.AND(mappedAIGEdge(mapEdges, node.Fanins[0]), mappedAIGEdge(mapEdges, node.Fanins[1]))
		case aig.Buffer:
			mapEdges[id] = mappedAIGEdge(mapEdges, node.Fanins[0])
		case aig.ExclusiveOr:
			mapEdges[id] = dst.XOR(mappedAIGEdge(mapEdges, node.Fanins[0]), mappedAIGEdge(mapEdges, node.Fanins[1]))
		}
	}
	next := make([]aig.Edge, regs)
	for i := 0; i < regs; i++ {
		next[i] = mappedAIGEdge(mapEdges, src.Node(outputs[len(outputs)-regs+i]).Fanins[0])
	}
	return mapEdges, next
}

func mappedAIGEdge(mapping []aig.Edge, edge aig.Edge) aig.Edge {
	result := mapping[edge.NodeID()]
	if edge.Complemented() {
		return result.Not()
	}
	return result
}
