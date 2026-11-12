package res

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
)

// GraphSAT is a Tseitin encoding of an AIG graph in an incremental BSAT
// solver. It owns no graph or solver storage; callers keep both alive.
type GraphSAT struct {
	Graph  *aig.Graph
	Solver *bsat.Solver
	vars   []int32
}

// Res_SatEncodeGraph adds a complete, phase-aware Tseitin encoding of g to s.
// Existing solver variables are respected, so the encoding can share a solver
// with other independently constructed formulas.
func Res_SatEncodeGraph(s *bsat.Solver, g *aig.Graph) (*GraphSAT, error) {
	if s == nil || g == nil {
		return nil, fmt.Errorf("res: nil SAT solver or AIG graph")
	}
	if err := g.Check(); err != nil {
		return nil, fmt.Errorf("res: invalid AIG graph: %w", err)
	}

	p := &GraphSAT{Graph: g, Solver: s, vars: make([]int32, len(g.Nodes()))}
	for i := range p.vars {
		p.vars[i] = s.NewVariable()
	}
	for id, node := range g.Nodes() {
		edge := aig.Edge(id << 1)
		if !g.ValidEdge(edge) {
			continue
		}
		out := p.vars[id]
		switch node.Kind {
		case aig.Constant:
			if Res_SatAddConst1(s, out, 0) == 0 {
				return nil, fmt.Errorf("res: conflicting constant encoding")
			}
		case aig.Input:
			// Primary inputs are unconstrained SAT variables.
		case aig.And:
			a, b := node.Fanins[0], node.Fanins[1]
			if Res_SatAddAnd(s, out, p.vars[a.NodeID()], p.vars[b.NodeID()], edgePhase(a), edgePhase(b)) == 0 {
				return nil, fmt.Errorf("res: conflicting AND encoding at node %d", id)
			}
		case aig.ExclusiveOr:
			if !p.addXor(out, node.Fanins[0], node.Fanins[1]) {
				return nil, fmt.Errorf("res: conflicting XOR encoding at node %d", id)
			}
		case aig.Buffer, aig.Output:
			fanin := node.Fanins[0]
			if Res_SatAddEqual(s, out, p.vars[fanin.NodeID()], edgePhase(fanin)) == 0 {
				return nil, fmt.Errorf("res: conflicting wire encoding at node %d", id)
			}
		default:
			return nil, fmt.Errorf("res: unsupported AIG node kind %d at node %d", node.Kind, id)
		}
	}
	return p, nil
}

func edgePhase(edge aig.Edge) int32 {
	if edge.Complemented() {
		return 1
	}
	return 0
}

func (p *GraphSAT) addXor(output int32, a, b aig.Edge) bool {
	litA, litB := resLiteral(p.vars[a.NodeID()], edgePhase(a)), resLiteral(p.vars[b.NodeID()], edgePhase(b))
	clauses := [][]bsat.Literal{
		{litA, litB, bsat.Negative(output)},
		{litA ^ 1, litB ^ 1, bsat.Negative(output)},
		{litA, litB ^ 1, bsat.Positive(output)},
		{litA ^ 1, litB, bsat.Positive(output)},
	}
	for _, clause := range clauses {
		if !p.Solver.AddClause(clause...) {
			return false
		}
	}
	return true
}

// LiteralFor returns a SAT literal that requires edge to have value want.
func (p *GraphSAT) LiteralFor(edge aig.Edge, want bool) (bsat.Literal, error) {
	if p == nil || p.Graph == nil || !p.Graph.ValidEdge(edge) {
		return 0, fmt.Errorf("res: invalid graph edge")
	}
	negative := edge.Complemented() == want
	if negative {
		return bsat.Negative(p.vars[edge.NodeID()]), nil
	}
	return bsat.Positive(p.vars[edge.NodeID()]), nil
}

// Solve finds an input assignment satisfying every requested edge value.
// Unknown is returned with no assignment when the solver reaches a resource
// limit; Unsatisfiable is a valid result and is not converted to an error.
func (p *GraphSAT) Solve(edges []aig.Edge, values []bool) ([]bool, bsat.Status, error) {
	if p == nil || p.Solver == nil || p.Graph == nil {
		return nil, bsat.Unknown, fmt.Errorf("res: nil graph SAT encoding")
	}
	if len(edges) != len(values) {
		return nil, bsat.Unknown, fmt.Errorf("res: got %d edge constraints and %d values", len(edges), len(values))
	}
	assumptions := make([]bsat.Literal, len(edges))
	for i, edge := range edges {
		literal, err := p.LiteralFor(edge, values[i])
		if err != nil {
			return nil, bsat.Unknown, err
		}
		assumptions[i] = literal
	}
	status := p.Solver.Solve(assumptions...)
	if status != bsat.Satisfiable {
		return nil, status, nil
	}
	model := p.Solver.Model()
	inputs := make([]bool, len(p.Graph.Inputs()))
	for i, edge := range p.Graph.Inputs() {
		inputs[i] = model[p.vars[edge.NodeID()]]
	}
	return inputs, status, nil
}

// Res_SatProveGraphEquivalent returns true exactly when a and b agree for all
// primary-input assignments. Both mismatch phases are checked under solver
// assumptions, so the encoded graph and clauses are reused between solves.
func Res_SatProveGraphEquivalent(g *aig.Graph, a, b aig.Edge) (bool, error) {
	solver := bsat.NewSolver()
	p, err := Res_SatEncodeGraph(solver, g)
	if err != nil {
		return false, err
	}
	for _, pair := range [][2]bool{{true, false}, {false, true}} {
		x, err := p.LiteralFor(a, pair[0])
		if err != nil {
			return false, err
		}
		y, err := p.LiteralFor(b, pair[1])
		if err != nil {
			return false, err
		}
		switch solver.Solve(x, y) {
		case bsat.Satisfiable:
			return false, nil
		case bsat.Unsatisfiable:
		default:
			return false, fmt.Errorf("res: SAT solver returned unknown while checking equivalence")
		}
	}
	return true, nil
}
