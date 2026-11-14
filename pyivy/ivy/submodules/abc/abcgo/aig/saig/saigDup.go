package saig

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// dupGraph keeps source object IDs separate from destination edges. Register
// inputs and outputs remain trailing, as required by the sequential interface.
type dupGraph struct {
	src, dst *aig.Graph
	copies   map[int]aig.Edge
}

func newDupGraph(p *aig.Graph) *dupGraph {
	q := aig.NewGraph()
	q.Name = p.Name
	q.Spec = p.Spec
	return &dupGraph{p, q, map[int]aig.Edge{0: q.Const1()}}
}
func (d *dupGraph) edge(e aig.Edge) aig.Edge {
	r, ok := d.copies[e.NodeID()]
	if !ok {
		n := d.src.Node(e)
		switch n.Kind {
		case aig.And:
			r = d.dst.AND(d.edge(n.Fanins[0]), d.edge(n.Fanins[1]))
		case aig.ExclusiveOr:
			r = d.dst.NewXor(d.edge(n.Fanins[0]), d.edge(n.Fanins[1]))
		case aig.Buffer:
			r = aig.Aig_ObjCreate(d.dst, &aig.Node{Kind: aig.Buffer, Fanins: [2]aig.Edge{d.edge(n.Fanins[0]), aig.InvalidEdge}})
		case aig.Output:
			r = d.edge(n.Fanins[0])
		default:
			panic("saig: unmapped input")
		}
		d.copies[e.NodeID()] = r
	}
	if e.Complemented() {
		r = r.Not()
	}
	return r
}
func (d *dupGraph) inputs() {
	for _, e := range d.src.Inputs() {
		d.copies[e.NodeID()] = d.dst.NewInput()
	}
}

// allNodes preserves the original duplication pass, including dangling logic.
func (d *dupGraph) allNodes() {
	for _, e := range d.src.TopologicalObjects() {
		n := d.src.Node(e)
		if n.Kind == aig.And || n.Kind == aig.ExclusiveOr || n.Kind == aig.Buffer {
			d.edge(e)
		}
	}
}
func (d *dupGraph) finish(regs int32) *aig.Graph {
	aig.Aig_ManSetRegNum(d.dst, regs)
	aig.Aig_ManCleanup(d.dst)
	return d.dst
}
func Saig_ManDupOrpos(p *aig.Graph) *aig.Graph {
	if aig.Aig_ManConstrNum(p) > 0 {
		return nil
	}
	d := newDupGraph(p)
	d.inputs()
	d.allNodes()
	m := d.dst.Const0()
	for _, e := range p.Outputs()[:Saig_ManPoNum(p)] {
		m = d.dst.OR(m, d.edge(e))
	}
	d.dst.AddOutput(m)
	for _, e := range p.Outputs()[Saig_ManPoNum(p):] {
		d.dst.AddOutput(d.edge(e))
	}
	return d.finish(p.RegisterCount())
}
func Saig_ManCreateEquivMiter(p *aig.Graph, pairs *vec.Vec_Int_t, addOutputs int32) *aig.Graph {
	if aig.Aig_ManConstrNum(p) > 0 {
		return nil
	}
	if vec.Vec_IntSize(pairs)%2 != 0 {
		panic("saig: odd equivalence pairs")
	}
	d := newDupGraph(p)
	d.inputs()
	d.allNodes()
	for i := 0; i < vec.Vec_IntSize(pairs); i += 2 {
		a, b := aig.Edge(vec.Vec_IntEntry(pairs, i))<<1, aig.Edge(vec.Vec_IntEntry(pairs, i+1))<<1
		r := d.dst.NewXor(d.edge(a), d.edge(b))
		if p.Node(a).Phase != p.Node(b).Phase {
			r = r.Not()
		}
		d.dst.AddOutput(r)
	}
	if addOutputs != 0 {
		for _, e := range p.Outputs()[Saig_ManPoNum(p):] {
			d.dst.AddOutput(d.edge(e))
		}
		return d.finish(p.RegisterCount())
	}
	return d.finish(0)
}
func Saig_ManTrimPis(p *aig.Graph) *aig.Graph {
	d := newDupGraph(p)
	allUnused := true
	for _, e := range p.Inputs()[:Saig_ManPiNum(p)] {
		if p.Node(e).Refs > 0 {
			allUnused = false
		}
	}
	for i, e := range p.Inputs() {
		if allUnused || p.Node(e).Refs > 0 || i >= int(Saig_ManPiNum(p)) {
			d.copies[e.NodeID()] = d.dst.NewInput()
			original := int32(i)
			if p.InputNumbers != nil {
				original = p.InputNumbers[i]
			}
			d.dst.InputNumbers = append(d.dst.InputNumbers, original)
		}
	}
	d.allNodes()
	for _, e := range p.Outputs() {
		d.dst.AddOutput(d.edge(e))
	}
	aig.Aig_ManSetConstrNum(d.dst, aig.Aig_ManConstrNum(p))
	aig.Aig_ManSetRegNum(d.dst, p.RegisterCount())
	return d.dst
}
func Saig_ManDupAbstraction(p *aig.Graph, flops *vec.Vec_Int_t) *aig.Graph {
	selected := make([]bool, p.RegisterCount())
	for i := 0; i < vec.Vec_IntSize(flops); i++ {
		k := vec.Vec_IntEntry(flops, i)
		if k < 0 || k >= p.RegisterCount() || selected[k] {
			panic("saig: invalid abstraction register")
		}
		selected[k] = true
	}
	d := newDupGraph(p)
	pis := int(Saig_ManPiNum(p))
	for pass := 0; pass < 2; pass++ {
		for i, e := range p.Inputs() {
			keep := i >= pis && selected[i-pis]
			if keep != (pass == 1) {
				continue
			}
			d.copies[e.NodeID()] = d.dst.NewInput()
			d.dst.InputNumbers = append(d.dst.InputNumbers, int32(i))
		}
	}
	for _, e := range p.Outputs()[:Saig_ManPoNum(p)] {
		d.dst.AddOutput(d.edge(e))
	}
	for i, keep := range selected {
		if keep {
			d.dst.AddOutput(d.edge(p.Outputs()[int(Saig_ManPoNum(p))+i]))
		}
	}
	d.finish(int32(vec.Vec_IntSize(flops)))
	before := d.dst.Inputs()
	numbers := append([]int32(nil), d.dst.InputNumbers...)
	aig.Aig_ManSeqCleanup(d.dst)
	byID := make(map[int]int32, len(before))
	for i, e := range before {
		byID[e.NodeID()] = numbers[i]
	}
	d.dst.InputNumbers = nil
	for _, e := range d.dst.Inputs() {
		d.dst.InputNumbers = append(d.dst.InputNumbers, byID[e.NodeID()])
	}
	return d.dst
}
func Saig_ManDupWithPhase(p *aig.Graph, initial *vec.Vec_Int_t) *aig.Graph {
	if vec.Vec_IntSize(initial) < int(p.RegisterCount()) {
		panic("saig: missing initial register values")
	}
	d := newDupGraph(p)
	d.inputs()
	pis := int(Saig_ManPiNum(p))
	pos := int(Saig_ManPoNum(p))
	for i, e := range p.Inputs()[pis:] {
		if vec.Vec_IntEntry(initial, i) != 0 {
			d.copies[e.NodeID()] = d.copies[e.NodeID()].Not()
		}
	}
	d.allNodes()
	for i, e := range p.Outputs() {
		r := d.edge(e)
		if i >= pos && vec.Vec_IntEntry(initial, i-pos) != 0 {
			r = r.Not()
		}
		d.dst.AddOutput(r)
	}
	aig.Aig_ManSetConstrNum(d.dst, aig.Aig_ManConstrNum(p))
	return d.finish(p.RegisterCount())
}

// Saig_ManDupCones closes selected output cones through register dependencies.
// Primary inputs and registers follow first discovery order, as in the C DFS.
func Saig_ManDupCones(p *aig.Graph, outputs []int32, nOutputs int32) *aig.Graph {
	if nOutputs < 0 || int(nOutputs) > len(outputs) {
		panic("saig: invalid cone output count")
	}
	roots := make([]aig.Edge, nOutputs)
	for i := range roots {
		roots[i] = p.Outputs()[outputs[i]]
	}
	seen := map[int]bool{0: true}
	var leaves, regs []aig.Edge
	var visit func(aig.Edge)
	visit = func(e aig.Edge) {
		if seen[e.NodeID()] {
			return
		}
		seen[e.NodeID()] = true
		n := p.Node(e)
		switch n.Kind {
		case aig.Input:
			if Saig_ObjIsLo(p, e) {
				regs = append(regs, e)
				roots = append(roots, Saig_ObjLoToLi(p, e))
			} else {
				leaves = append(leaves, e)
			}
		case aig.And, aig.ExclusiveOr:
			visit(n.Fanins[0])
			visit(n.Fanins[1])
		case aig.Buffer, aig.Output:
			visit(n.Fanins[0])
		}
	}
	for i := 0; i < len(roots); i++ {
		visit(roots[i])
	}
	d := newDupGraph(p)
	for _, e := range append(leaves, regs...) {
		d.copies[e.NodeID()] = d.dst.NewInput()
	}
	for _, e := range roots {
		d.dst.AddOutput(d.edge(e))
	}
	return d.finish(int32(len(regs)))
}

// replayCex evaluates the final frame without mutating graph marks. The caller
// chooses whether to expose final-frame values through the legacy MarkB API.
func replayCex(p *aig.Graph, c *util.Abc_Cex_t) ([]bool, []bool, []bool) {
	return replayCexFrames(p, c, nil)
}
func replayCexFrames(p *aig.Graph, c *util.Abc_Cex_t, onFrame func(int32, []bool)) ([]bool, []bool, []bool) {
	if c == nil || c.IFrame < 0 || c.NRegs != p.RegisterCount() || c.NPis != Saig_ManPiNum(p) || c.NBits != c.NRegs+(c.IFrame+1)*c.NPis || int64(len(c.PData))*32 < int64(c.NBits) {
		panic("saig: incompatible counterexample")
	}
	bit := func(i int32) bool { return c.PData[i/32]&(1<<uint(i%32)) != 0 }
	state := make([]bool, c.NRegs)
	for i := range state {
		state[i] = bit(int32(i))
	}
	var values, outs []bool
	for frame := int32(0); frame <= c.IFrame; frame++ {
		inputs := make([]bool, len(p.Inputs()))
		for i := int32(0); i < c.NPis; i++ {
			inputs[i] = bit(c.NRegs + frame*c.NPis + i)
		}
		copy(inputs[c.NPis:], state)
		if onFrame != nil {
			onFrame(frame, inputs)
		}
		values = make([]bool, len(p.Nodes()))
		values[0] = true
		for i, e := range p.Inputs() {
			values[e.NodeID()] = inputs[i]
		}
		var evaluate func(aig.Edge) bool
		done := make([]bool, len(values))
		done[0] = true
		for _, e := range p.Inputs() {
			done[e.NodeID()] = true
		}
		evaluate = func(e aig.Edge) bool {
			if !done[e.NodeID()] {
				n := p.Node(e)
				switch n.Kind {
				case aig.And:
					values[e.NodeID()] = evaluate(n.Fanins[0]) && evaluate(n.Fanins[1])
				case aig.ExclusiveOr:
					values[e.NodeID()] = evaluate(n.Fanins[0]) != evaluate(n.Fanins[1])
				case aig.Buffer, aig.Output:
					values[e.NodeID()] = evaluate(n.Fanins[0])
				}
				done[e.NodeID()] = true
			}
			return values[e.NodeID()] != e.Complemented()
		}
		outs = make([]bool, len(p.Outputs()))
		for i, e := range p.Outputs() {
			outs[i] = evaluate(e)
		}
		for _, e := range p.TopologicalObjects() {
			evaluate(e)
		}
		if frame < c.IFrame {
			copy(state, outs[Saig_ManPoNum(p):])
		}
	}
	return values, state, outs
}
func Saig_ManVerifyCexNoClear(p *aig.Graph, c *util.Abc_Cex_t) int32 {
	v, _, o := replayCex(p, c)
	aig.Aig_ManCleanMarkB(p)
	for i, x := range v {
		if x && p.ValidEdge(aig.Edge(i)<<1) {
			aig.Aig_ObjSetMarkB(p, aig.Edge(i)<<1)
		}
	}
	if c.IPo < 0 || int(c.IPo) >= len(o) {
		panic("saig: invalid failing output")
	}
	if o[c.IPo] {
		return 1
	}
	return 0
}
func Saig_ManVerifyCex(p *aig.Graph, c *util.Abc_Cex_t) int32 {
	r := Saig_ManVerifyCexNoClear(p, c)
	aig.Aig_ManCleanMarkB(p)
	return r
}
func Saig_ManReturnFailingState(p *aig.Graph, c *util.Abc_Cex_t, next int32) *vec.Vec_Int_t {
	_, state, o := replayCex(p, c)
	aig.Aig_ManCleanMarkB(p)
	if c.IPo < 0 || int(c.IPo) >= len(o) || !o[c.IPo] {
		return nil
	}
	if next != 0 {
		state = o[Saig_ManPoNum(p):]
	}
	v := vec.Vec_IntAlloc(len(state))
	for _, b := range state {
		var x int32
		if b {
			x = 1
		}
		vec.Vec_IntPush(v, x)
	}
	return v
}
func Saig_ManFindFailedPoCex(p *aig.Graph, c *util.Abc_Cex_t) int32 {
	_, _, o := replayCex(p, c)
	aig.Aig_ManCleanMarkB(p)
	for i, b := range o[:Saig_ManPoNum(p)] {
		if b {
			return int32(i)
		}
	}
	return -1
}
func Saig_ManExtendCex(p *aig.Graph, c *util.Abc_Cex_t) *util.Abc_Cex_t {
	if c == nil {
		panic("saig: missing counterexample")
	}
	q := util.Abc_CexAlloc(0, int32(len(p.Inputs())), c.IFrame+1)
	q.IPo = c.IPo
	q.IFrame = c.IFrame
	replayCexFrames(p, c, func(frame int32, inputs []bool) {
		for i, b := range inputs {
			if b {
				index := frame*q.NPis + int32(i)
				q.PData[index/32] |= 1 << uint(index%32)
			}
		}
	})
	aig.Aig_ManCleanMarkB(p)
	return q
}
