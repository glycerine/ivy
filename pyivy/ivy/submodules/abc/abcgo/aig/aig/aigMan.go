package aig

import "fmt"

func Aig_ManStart(capacity int32) *Graph {
	if capacity < 0 {
		panic("aig: negative capacity")
	}
	p := NewGraph()
	if capacity > 1 {
		nodes := make([]Node, 1, int(capacity))
		nodes[0] = p.nodes[0]
		p.nodes = nodes
	}
	return p
}
func Aig_ManStartFrom(p *Graph) *Graph {
	q := Aig_ManStart(int32(len(p.nodes)))
	q.Name, q.Spec = p.Name, p.Spec
	p.copies = map[int]Edge{0: 0}
	for _, e := range p.inputs {
		x := q.NewInput()
		q.node(x).Level = p.node(e).Level
		p.copies[e.NodeID()] = x
	}
	return q
}
func Aig_ManDup_rec(destination, source *Graph, root Edge) Edge {
	id := root.NodeID()
	if copy, ok := source.copies[id]; ok {
		if root.Complemented() {
			return copy.Not()
		}
		return copy
	}
	n := *source.node(root)
	var copy Edge
	switch n.Kind {
	case And:
		copy = destination.AND(Aig_ManDup_rec(destination, source, n.Fanins[0]), Aig_ManDup_rec(destination, source, n.Fanins[1]))
	case ExclusiveOr:
		copy = destination.NewXor(Aig_ManDup_rec(destination, source, n.Fanins[0]), Aig_ManDup_rec(destination, source, n.Fanins[1]))
	case Buffer, Output:
		copy = Aig_ManDup_rec(destination, source, n.Fanins[0])
	default:
		panic("aig: unbound source terminal")
	}
	if source.copies == nil {
		source.copies = make(map[int]Edge)
	}
	source.copies[id] = copy
	if root.Complemented() {
		return copy.Not()
	}
	return copy
}
func Aig_ManExtractMiter(p *Graph, a, b Edge) *Graph {
	q := Aig_ManStartFrom(p)
	q.AddOutput(q.XOR(Aig_ManDup_rec(q, p, a), Aig_ManDup_rec(q, p, b)))
	Aig_ManCleanup(q)
	return q
}

// Stop releases owned storage. Callers must not use manager edges afterwards.
func Aig_ManStop(p *Graph) {
	if p != nil {
		*p = Graph{}
	}
}
func Aig_ManStopP(p **Graph) {
	if p != nil {
		Aig_ManStop(*p)
		*p = nil
	}
}
func Aig_ManCleanup(p *Graph) int32 {
	live := make([]bool, len(p.nodes))
	live[0] = true
	for _, e := range p.inputs {
		live[e.NodeID()] = true
	}
	var visit func(Edge)
	visit = func(e Edge) {
		id := e.NodeID()
		if live[id] {
			return
		}
		live[id] = true
		n := p.nodes[id]
		if n.Kind == And || n.Kind == ExclusiveOr || (n.Kind == Buffer || n.Kind == Output) {
			visit(n.Fanins[0])
			if n.Kind != Buffer && n.Kind != Output {
				visit(n.Fanins[1])
			}
		}
	}
	for _, e := range p.outputs {
		visit(e)
	}
	var removed int32
	for id, n := range p.nodes {
		if p.deleted[id] || live[id] {
			continue
		}
		p.deleted[id] = true
		removed++
		if n.Kind == And || n.Kind == ExclusiveOr {
			a, b := n.Fanins[0], n.Fanins[1]
			if a > b {
				a, b = b, a
			}
			delete(p.nodeTable(n.Kind), edgePair{a, b})
		}
	}
	p.recountReferences()
	p.refreshFanouts()
	return removed
}
func (p *Graph) recountReferences() {
	for i := range p.nodes {
		p.nodes[i].Refs = 0
	}
	for id, n := range p.nodes {
		if p.deleted[id] {
			continue
		}
		if n.Kind == And || n.Kind == ExclusiveOr {
			p.node(n.Fanins[0]).Refs++
			p.node(n.Fanins[1]).Refs++
		} else if n.Kind == Buffer || n.Kind == Output {
			p.node(n.Fanins[0]).Refs++
		}
	}
	for _, e := range p.outputs {
		if p.node(e).Kind != Output {
			p.node(e).Refs++
		}
	}
}
func Aig_ManAntiCleanup(p *Graph) int32 {
	old := len(p.outputs)
	for id, n := range p.nodes {
		if !p.deleted[id] && (n.Kind == And || n.Kind == ExclusiveOr) && n.Refs == 0 {
			p.AddOutput(Edge(id) << 1)
		}
	}
	return int32(old - len(p.outputs))
}
func Aig_ManCiCleanup(p *Graph) int32 {
	old := len(p.inputs)
	keep := p.inputs[:0]
	for i, e := range p.inputs {
		if i >= old-int(p.nRegs) || p.node(e).Refs > 0 {
			keep = append(keep, e)
		} else {
			p.deleted[e.NodeID()] = true
		}
	}
	p.inputs = keep
	if p.nRegs != 0 {
		p.nTruePis = int32(len(p.inputs)) - p.nRegs
	}
	p.refreshFanouts()
	return int32(old - len(keep))
}
func Aig_ManCoCleanup(p *Graph) int32 {
	old := len(p.outputs)
	original := append([]Edge(nil), p.outputs...)
	keep := p.outputs[:0]
	for i, e := range p.outputs {
		if i >= old-int(p.nRegs) || p.outputDriver(e) != p.Const0() {
			keep = append(keep, e)
		}
	}
	removed := make(map[int]bool)
	for _, e := range original {
		if p.node(e).Kind == Output && p.outputDriver(e) == p.Const0() {
			removed[e.NodeID()] = true
		}
	}
	for _, e := range keep {
		delete(removed, e.NodeID())
	}
	for id := range removed {
		p.deleted[id] = true
	}
	p.outputs = keep
	if p.nRegs != 0 {
		p.nTruePos = int32(len(p.outputs)) - p.nRegs
	}
	p.recountReferences()
	p.refreshFanouts()
	return int32(old - len(keep))
}
func Aig_ManPrintStats(p *Graph) {
	fmt.Printf("%-15s : pi = %5d  po = %5d  ", p.Name, len(p.inputs)-int(p.nRegs), len(p.outputs)-int(p.nRegs))
	if p.nRegs != 0 {
		fmt.Printf("lat = %5d  ", p.nRegs)
	}
	fmt.Printf("and = %7d  ", Aig_ManAndNum(p))
	if count := Aig_ManExorNum(p); count != 0 {
		fmt.Printf("xor = %5d  ", count)
	}
	if count := Aig_ManChoiceNum(p); count != 0 {
		fmt.Printf("ch = %5d  ", count)
	}
	if count := Aig_ManBufNum(p); count != 0 {
		fmt.Printf("buf = %5d  ", count)
	}
	fmt.Printf("lev = %3d\n", Aig_ManLevels(p))
}
func Aig_ManReportImprovement(p, q *Graph) {
	reduction := func(a, b int32) float64 {
		if a == 0 {
			return 0
		}
		return 100 * float64(a-b) / float64(a)
	}
	fmt.Printf("REG: Beg = %5d. End = %5d. (R =%5.1f %%)  AND: Beg = %6d. End = %6d. (R =%5.1f %%)\n", p.nRegs, q.nRegs, reduction(p.nRegs, q.nRegs), Aig_ManNodeNum(p), Aig_ManNodeNum(q), reduction(Aig_ManNodeNum(p), Aig_ManNodeNum(q)))
}
func Aig_ManSetRegNum(p *Graph, n int32) {
	p.nRegs = n
	p.nTruePis = int32(len(p.inputs)) - n
	p.nTruePos = int32(len(p.outputs)) - n
	Aig_ManSetCioIds(p)
}
func Aig_ManFlipFirstPo(p *Graph) {
	if len(p.outputs) == 0 {
		panic("aig: no output")
	}
	if output := p.node(p.outputs[0]); output.Kind == Output {
		// C flips the CO's child, preserving the terminal's identity. Consumers
		// such as CNF derivation and dump writers read that child directly.
		output.Fanins[0] = output.Fanins[0].Not()
	} else {
		// Legacy raw output edges have no separate CO object.
		p.outputs[0] = p.outputs[0].Not()
	}
}
func Aig_ManReleaseData(p *Graph) any { data := p.Data; p.Data = nil; return data }

func (p *Graph) outputDriver(e Edge) Edge {
	if p.node(e).Kind == Output {
		driver := p.node(e).Fanins[0]
		if e.Complemented() {
			driver = driver.Not()
		}
		return driver
	}
	return e
}
