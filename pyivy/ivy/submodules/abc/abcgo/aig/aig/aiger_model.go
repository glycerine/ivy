package aig

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/ioa"
)

// AigerModel normalizes rewriting buffers, explicit XORs, and arbitrary DAG
// order into AIGER's AND-only representation. The source manager is untouched.
func (p *Graph) AigerModel() (*ioa.Model, error) {
	if p == nil {
		return nil, fmt.Errorf("aig: nil manager")
	}
	if err := p.Check(); err != nil {
		return nil, err
	}
	if p.nRegs < 0 || int(p.nRegs) > len(p.inputs) || int(p.nRegs) > len(p.outputs) {
		return nil, fmt.Errorf("aig: invalid register count")
	}
	ordinary := len(p.outputs) - int(p.nRegs)
	if p.nConstrs < 0 || int(p.nConstrs) > ordinary {
		return nil, fmt.Errorf("aig: invalid constraint count")
	}
	q := NewGraph()
	mapping := map[int]Edge{0: 0}
	for _, e := range p.inputs {
		mapping[e.NodeID()] = q.NewInput()
	}
	var copyEdge func(Edge) (Edge, error)
	active := make(map[int]bool)
	copyEdge = func(e Edge) (Edge, error) {
		id := e.NodeID()
		value, ok := mapping[id]
		if !ok {
			if active[id] {
				return InvalidEdge, fmt.Errorf("aig: cycle during AIGER conversion")
			}
			active[id] = true
			n := p.nodes[id]
			a, err := copyEdge(n.Fanins[0])
			if err != nil {
				return InvalidEdge, err
			}
			switch n.Kind {
			case Output, Buffer:
				value = a
			case And, ExclusiveOr:
				b, err := copyEdge(n.Fanins[1])
				if err != nil {
					return InvalidEdge, err
				}
				if n.Kind == And {
					value = q.AND(a, b)
				} else {
					value = q.XOR(a, b)
				}
			default:
				return InvalidEdge, fmt.Errorf("aig: unsupported AIGER node %d", id)
			}
			delete(active, id)
			mapping[id] = value
		}
		if e.Complemented() {
			value = value.Not()
		}
		return value, nil
	}
	roots := make([]Edge, len(p.outputs))
	for i, e := range p.outputs {
		value, err := copyEdge(e)
		if err != nil {
			return nil, err
		}
		roots[i] = value
	}
	literal := func(e Edge) uint32 {
		if e.NodeID() == 0 {
			return uint32(e) ^ 1
		}
		return uint32(e)
	}
	result := &ioa.Model{MaxVar: uint32(len(q.nodes) - 1), Name: p.Name}
	piCount := len(p.inputs) - int(p.nRegs)
	for _, e := range q.inputs[:piCount] {
		result.Inputs = append(result.Inputs, literal(e))
	}
	for i, e := range q.inputs[piCount:] {
		result.Latches = append(result.Latches, ioa.Latch{Literal: literal(e), Next: literal(roots[ordinary+i]), Init: 0})
	}
	for i, e := range roots[:ordinary] {
		if p.nConstrs == 0 {
			result.Outputs = append(result.Outputs, literal(e))
		} else if i < ordinary-int(p.nConstrs) {
			result.Bad = append(result.Bad, literal(e))
		} else {
			result.Constraints = append(result.Constraints, literal(e.Not()))
		}
	}
	for id, n := range q.nodes {
		if n.Kind == And {
			left, right := literal(n.Fanins[0]), literal(n.Fanins[1])
			if left < right {
				left, right = right, left
			}
			result.ANDs = append(result.ANDs, ioa.And{Literal: uint32(id) << 1, Left: left, Right: right})
		}
	}
	if err := result.Check(); err != nil {
		return nil, err
	}
	return result, nil
}
