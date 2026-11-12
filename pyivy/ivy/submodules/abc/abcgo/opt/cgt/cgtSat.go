package cgt

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/satoko"
)

func (p *Manager) encodePart() {
	if p.solver == nil {
		p.solver = satoko.New()
		p.encoded = make(map[int]bool)
	}
	p.solver.SetVariables(len(p.Part.Nodes()))
	for _, e := range p.Part.LiveEdges() {
		id := e.NodeID()
		if p.encoded[id] {
			continue
		}
		p.encoded[id] = true
		n := p.Part.Node(e)
		z := int(e)
		a, b := int(n.Fanins[0]), int(n.Fanins[1])
		add := func(lits ...int) { p.solver.AddClause(lits) }
		switch n.Kind {
		case aig.Constant:
			add(z)
		case aig.Input:
		case aig.And:
			add(a^1, b^1, z)
			add(a, z^1)
			add(b, z^1)
		case aig.ExclusiveOr:
			add(a, b, z^1)
			add(a^1, b^1, z^1)
			add(a, b^1, z)
			add(a^1, b, z)
		case aig.Buffer, aig.Output:
			add(a^1, z)
			add(a, z^1)
		}
	}
}
func Cgt_CheckImplication(p *Manager, gate, miter aig.Edge) int {
	if p.Part == nil || !p.Part.ValidEdge(gate) || !p.Part.ValidEdge(miter) {
		panic("cgt: implication literals are outside the active partition")
	}
	p.encodePart()
	p.Stats.Calls++
	status := p.solver.SolveAssumptionsLimit([]int{int(gate), int(miter)}, int64(p.Params.ConflictMax))
	switch status {
	case satoko.SATOKO_UNSAT:
		p.Stats.UNSAT++
		p.solver.AddClause([]int{int(gate) ^ 1, int(miter) ^ 1})
		return 1
	case satoko.SATOKO_SAT:
		p.Stats.SAT++
		inputs := p.Part.Inputs()
		assignment := make([]bool, len(inputs))
		for i, e := range inputs {
			assignment[i] = p.solver.ModelValue(e.NodeID()) == 1
		}
		p.lastModel = evaluateNodes(p.Part, assignment)
		return 0
	default:
		p.Stats.Unknown++
		return -1
	}
}
func Cgt_SimulationFilter(p *Manager, gate, miter aig.Edge) bool {
	for _, pattern := range p.patterns {
		if gate.NodeID() >= len(pattern) || miter.NodeID() >= len(pattern) {
			continue
		}
		g := pattern[gate.NodeID()] != (gate&1 != 0)
		m := pattern[miter.NodeID()] != (miter&1 != 0)
		if g && m {
			return false
		}
	}
	return true
}
func Cgt_SimulationRecord(p *Manager) {
	if p.lastModel == nil {
		panic("cgt: no satisfiable pattern to record")
	}
	p.patterns = append(p.patterns, append([]bool(nil), p.lastModel...))
	p.Stats.Patterns++
}
