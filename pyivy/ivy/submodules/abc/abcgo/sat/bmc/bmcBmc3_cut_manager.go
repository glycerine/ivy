package bmc

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

type bmcCutFrame struct {
	lits      []int32
	known     []bool
	tern      []bmcTernary
	ternKnown []bool
	active    []bool
	piLits    []int32
	piKnown   []bool
}

// bmc3CutManager is the dynamic object/frame map used by the scalable BMC
// path. Its literals use C's encoding (0=false, 1=true, 2*v|phase otherwise).
// SAT variable zero is reserved in the C manager; the other backends retain
// the same zero-based variable IDs and C literal phases.
type bmc3CutManager struct {
	graph      *aig.Graph
	nodes      []aig.Node
	inputs     []aig.Edge
	outputs    []aig.Edge
	inputIndex map[int]int
	regs, pis  int32
	pos        int32
	mapping    *vec.Vec_Int_t
	encoder    *bmcCutFrameEncoder
	frames     []*bmcCutFrame
	pending    [][]int32
}

func newBmc3CutManager(graph *aig.Graph, regs, pis, pos int32) (*bmc3CutManager, error) {
	if graph == nil {
		return nil, fmt.Errorf("bmc: nil cut graph")
	}
	sourceGraph, err := bmcNormalizeAIG(graph)
	if err != nil {
		return nil, err
	}
	mapping := cnf.Cnf_DeriveMappingArray(sourceGraph)
	if mapping == nil {
		return nil, fmt.Errorf("bmc: failed to derive cut mapping")
	}
	m := &bmc3CutManager{graph: sourceGraph, nodes: sourceGraph.Nodes(), inputs: sourceGraph.Inputs(), outputs: sourceGraph.Outputs(), inputIndex: make(map[int]int), regs: regs, pis: pis, pos: pos, mapping: mapping, encoder: newBmcCutFrameEncoder(1)}
	for i, e := range m.inputs {
		m.inputIndex[e.NodeID()] = i
	}
	return m, nil
}

// bmcNormalizeAIG lowers Go's native XOR nodes into the AND/inverter graph
// accepted by Cnf_DeriveMappingArray and BMC3's source cut manager.
func bmcNormalizeAIG(src *aig.Graph) (*aig.Graph, error) {
	nodes, inputs, outputs := src.Nodes(), src.Inputs(), src.Outputs()
	needsCopy := false
	for _, n := range nodes {
		if n.Kind == aig.ExclusiveOr {
			needsCopy = true
			break
		}
	}
	if !needsCopy {
		return src, nil
	}
	dst := aig.NewGraph()
	dst.Name, dst.Spec = src.Name, src.Spec
	mapped := make([]aig.Edge, len(nodes))
	mapped[0] = dst.Const1()
	for _, input := range inputs {
		mapped[input.NodeID()] = dst.NewInput()
	}
	child := func(e aig.Edge) aig.Edge {
		v := mapped[e.NodeID()]
		if e.Complemented() {
			v = v.Not()
		}
		return v
	}
	for id, n := range nodes {
		switch n.Kind {
		case aig.Constant, aig.Input:
		case aig.And:
			mapped[id] = dst.AND(child(n.Fanins[0]), child(n.Fanins[1]))
		case aig.ExclusiveOr:
			a, b := child(n.Fanins[0]), child(n.Fanins[1])
			mapped[id] = dst.OR(dst.AND(a, b.Not()), dst.AND(a.Not(), b))
		case aig.Buffer, aig.Output:
			mapped[id] = child(n.Fanins[0])
		}
	}
	for _, output := range outputs {
		n := nodes[output.NodeID()]
		dst.AddOutput(child(n.Fanins[0]))
	}
	aig.Aig_ManSetRegNum(dst, aig.Aig_ManRegNum(src))
	return dst, nil
}

func (m *bmc3CutManager) ensureFrame(frame int) *bmcCutFrame {
	for len(m.frames) <= frame {
		f := &bmcCutFrame{lits: make([]int32, len(m.nodes)), known: make([]bool, len(m.nodes)), tern: make([]bmcTernary, len(m.nodes)), ternKnown: make([]bool, len(m.nodes)), active: make([]bool, len(m.nodes)), piLits: make([]int32, m.pis), piKnown: make([]bool, m.pis)}
		f.lits[0], f.known[0] = 1, true
		f.tern[0], f.ternKnown[0] = bmcTernaryOne, true
		for i, e := range m.inputs[int(m.pis):] {
			if len(m.frames) == 0 {
				f.lits[e.NodeID()], f.known[e.NodeID()] = 0, true
				f.tern[e.NodeID()], f.ternKnown[e.NodeID()] = bmcTernaryZero, true
			}
			_ = i
		}
		m.frames = append(m.frames, f)
	}
	return m.frames[frame]
}

func (m *bmc3CutManager) literal(frame, objectID int) (int32, bool) {
	if frame < 0 || objectID < 0 || objectID >= len(m.nodes) {
		return 0, false
	}
	f := m.ensureFrame(frame)
	if f.known[objectID] {
		return f.lits[objectID], true
	}
	if value, evaluated := m.ternary(frame, objectID); evaluated && value != bmcTernaryUnknown {
		lit := int32(value)
		f.lits[objectID], f.known[objectID] = lit, true
		return lit, true
	}
	if f.active[objectID] {
		return 0, false
	}
	f.active[objectID] = true
	defer func() { f.active[objectID] = false }()
	n := m.nodes[objectID]
	var lit int32
	switch n.Kind {
	case aig.Input:
		index, ok := m.inputIndex[objectID]
		if !ok {
			return 0, false
		}
		if index < int(m.pis) {
			v := m.encoder.nextVariable
			m.encoder.nextVariable++
			lit = v << 1
			f.piLits[index], f.piKnown[index] = lit, true
		} else {
			if frame == 0 {
				lit = 0
			} else {
				li := m.outputs[int(m.pos)+index-int(m.pis)]
				var ok bool
				lit, ok = m.literal(frame-1, li.NodeID())
				if !ok {
					return 0, false
				}
				if li.Complemented() {
					lit ^= 1
				}
			}
		}
	case aig.Output, aig.Buffer:
		child := n.Fanins[0]
		var ok bool
		lit, ok = m.literal(frame, child.NodeID())
		if !ok {
			return 0, false
		}
		if child.Complemented() {
			lit ^= 1
		}
	case aig.And, aig.ExclusiveOr:
		encoded, mapped, err := m.encoder.EncodeObject(m.mapping, objectID, frame, func(fr, id int) (int32, bool) { return m.literal(fr, id) })
		if err != nil {
			return 0, false
		}
		if !mapped {
			return 0, false
		}
		m.pending = append(m.pending, encoded.clauses...)
		lit = encoded.resultLiteral
	case aig.Constant:
		lit = 1
	default:
		return 0, false
	}
	f.lits[objectID], f.known[objectID] = lit, true
	return lit, true
}

func (m *bmc3CutManager) ternary(frame, objectID int) (bmcTernary, bool) {
	if frame < 0 || objectID < 0 || objectID >= len(m.nodes) {
		return bmcTernaryUnknown, false
	}
	f := m.ensureFrame(frame)
	if f.ternKnown[objectID] {
		return f.tern[objectID], true
	}
	n := m.nodes[objectID]
	var v bmcTernary
	switch n.Kind {
	case aig.Constant:
		v = bmcTernaryOne
	case aig.Input:
		index, ok := m.inputIndex[objectID]
		if !ok {
			return bmcTernaryUnknown, false
		}
		if index < int(m.pis) {
			v = bmcTernaryUnknown
		} else if frame == 0 {
			v = bmcTernaryZero
		} else {
			li := m.outputs[int(m.pos)+index-int(m.pis)]
			value, known := m.ternary(frame-1, li.NodeID())
			if !known {
				return bmcTernaryUnknown, false
			}
			v = value
			if li.Complemented() {
				v = bmcTernaryNot(v)
			}
		}
	case aig.Output, aig.Buffer:
		child := n.Fanins[0]
		value, known := m.ternary(frame, child.NodeID())
		if !known {
			return bmcTernaryUnknown, false
		}
		v = value
		if child.Complemented() {
			v = bmcTernaryNot(v)
		}
	case aig.And, aig.ExclusiveOr:
		tta, ok := m.ternary(frame, n.Fanins[0].NodeID())
		if !ok {
			return bmcTernaryUnknown, false
		}
		ttb, ok := m.ternary(frame, n.Fanins[1].NodeID())
		if !ok {
			return bmcTernaryUnknown, false
		}
		if n.Fanins[0].Complemented() {
			tta = bmcTernaryNot(tta)
		}
		if n.Fanins[1].Complemented() {
			ttb = bmcTernaryNot(ttb)
		}
		if n.Kind == aig.And {
			if tta == bmcTernaryZero || ttb == bmcTernaryZero {
				v = bmcTernaryZero
			} else if tta == bmcTernaryOne && ttb == bmcTernaryOne {
				v = bmcTernaryOne
			} else {
				v = bmcTernaryUnknown
			}
		} else if tta == bmcTernaryUnknown || ttb == bmcTernaryUnknown {
			v = bmcTernaryUnknown
		} else if tta == ttb {
			v = bmcTernaryZero
		} else {
			v = bmcTernaryOne
		}
	default:
		return bmcTernaryUnknown, false
	}
	f.tern[objectID], f.ternKnown[objectID] = v, true
	return v, true
}

func bmcTernaryNot(v bmcTernary) bmcTernary {
	if v == bmcTernaryZero {
		return bmcTernaryOne
	}
	if v == bmcTernaryOne {
		return bmcTernaryZero
	}
	return bmcTernaryUnknown
}

func (m *bmc3CutManager) outputLiteral(frame int, output int32) (int32, bool) {
	if output < 0 || output >= m.pos {
		return 0, false
	}
	return m.literal(frame, m.outputs[output].NodeID())
}

func (m *bmc3CutManager) drainClauses() [][]int32 { out := m.pending; m.pending = nil; return out }
func (m *bmc3CutManager) variableCount() int32    { return m.encoder.nextVariable - 1 }

func (m *bmc3CutManager) counterexample(s bmc3Solver, frame, output int32) *util.Abc_Cex_t {
	model := &util.Abc_Cex_t{NRegs: m.regs, NPis: m.pis, IFrame: frame, IPo: output, NBits: m.regs + (frame+1)*m.pis}
	model.PData = make([]uint32, (int(model.NBits)+31)/32)
	for f := int32(0); f <= frame; f++ {
		state := m.ensureFrame(int(f))
		for pi := int32(0); pi < m.pis; pi++ {
			if !state.piKnown[pi] {
				continue
			}
			variable := state.piLits[pi] >> 1
			value, known := s.Value(variable)
			if known && value {
				bit := int(m.regs + f*m.pis + pi)
				model.PData[bit>>5] |= 1 << uint(bit&31)
			}
		}
	}
	return model
}

func bmcCToBSATLiteral(lit int32) (int32, error) {
	if lit < 2 {
		return 0, fmt.Errorf("bmc: constant literal %d cannot be a clause member", lit)
	}
	return (lit>>1)<<1 | lit&1, nil
}
