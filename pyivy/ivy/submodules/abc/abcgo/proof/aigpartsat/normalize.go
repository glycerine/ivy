package aigpartsat

import "github.com/glycerine/abcgo/aig/aig"

// Original C partitions are AND-only. Normalize typed XOR extensions through
// the existing AIGER adapter before applying the same cut-mapped CNF pipeline.
func normalizePart(p *aig.Graph) (*aig.Graph, error) {
	if aig.Aig_ManExorNum(p) == 0 && aig.Aig_ManBufNum(p) == 0 {
		return p, nil
	}
	m, err := p.AigerModel()
	if err != nil {
		return nil, err
	}
	q := aig.NewGraph()
	q.Name, q.Spec = p.Name, p.Spec
	mapping := map[uint32]aig.Edge{0: q.Const0()}
	original := p.Inputs()
	for i, lit := range m.Inputs {
		e := q.NewInput()
		mapping[lit>>1] = e
		n := p.Node(original[i])
		if n.MarkA {
			aig.Aig_ObjSetMarkA(q, e)
		}
		if n.MarkB {
			aig.Aig_ObjSetMarkB(q, e)
		}
	}
	edge := func(lit uint32) aig.Edge {
		e := mapping[lit>>1]
		if lit&1 != 0 {
			return e.Not()
		}
		return e
	}
	for _, gate := range m.ANDs {
		mapping[gate.Literal>>1] = q.AND(edge(gate.Left), edge(gate.Right))
	}
	for _, lit := range m.Outputs {
		q.AddOutput(edge(lit))
	}
	return q, nil
}
