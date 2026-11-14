// Forward retiming frontier, ported from ABC aigRetF.c (Alan Mishchenko).
package aig

// Aig_ManRetimeMark_rec marks the combinational cone reached by any already
// marked CI. Inversion does not affect structural reachability.
func Aig_ManRetimeMark_rec(p *Graph, e Edge) int32 {
	n := p.node(e)
	if n.MarkB {
		return 1
	}
	if n.Kind == Input || n.Kind == Constant {
		return 0
	}
	if n.TravID == p.travID {
		return 0
	}
	n.TravID = p.travID
	for _, fanin := range nodeFanins(*n) {
		if Aig_ManRetimeMark_rec(p, fanin) != 0 {
			p.node(e).MarkB = true
			return 1
		}
	}
	return 0
}

// Aig_ManRetimeMark computes the fixed point of true-PI reachability across
// register boundaries. Register correspondence is local data, replacing C's
// temporary pNext pointer links without disturbing other graph scratch maps.
func Aig_ManRetimeMark(p *Graph) {
	registers, pis, pos := sclLayout(p)
	for id, n := range p.nodes {
		if !p.deleted[id] && n.MarkB {
			panic("aig: retiming requires clear MarkB flags")
		}
	}
	for _, input := range p.inputs[:pis] {
		p.node(input).MarkB = true
	}
	registerByOutput := make(map[int]Edge, registers)
	for i := 0; i < registers; i++ {
		registerByOutput[p.outputs[pos+i].NodeID()] = p.inputs[pis+i]
	}
	for changed := true; changed; {
		changed = false
		Aig_ManIncrementTravId(p)
		for _, output := range p.outputs {
			if p.node(output).MarkB {
				continue
			}
			if Aig_ManRetimeMark_rec(p, output) != 0 {
				if input, ok := registerByOutput[output.NodeID()]; ok {
					p.node(input).MarkB = true
				}
				changed = true
			}
		}
	}
}

// Retiming combines register state, so original flop-ID correspondence no
// longer applies after a move. Clock domains remain meaningful and are kept:
// only register pairs from the same domain may be crossed.
func retimeFrontierDomains(p *Graph, pis int) map[int]int {
	domains := make(map[int]int, int(p.nRegs))
	for _, input := range p.inputs[pis:] {
		domains[input.NodeID()] = -1
	}
	for domain, registers := range p.ClockDomains {
		for _, register := range registers {
			input := p.inputs[pis+int(register)]
			if domains[input.NodeID()] != -1 {
				panic("aig: register appears in multiple clock domains")
			}
			domains[input.NodeID()] = domain
		}
	}
	return domains
}

// Aig_ManRetimeFrontier moves gates forward across zero-initialized registers
// and returns a structurally rehashed copy. The source graph is mutable work
// storage. C checks its limit after a move, so nonpositive limits permit one.
func Aig_ManRetimeFrontier(p *Graph, steps int32) *Graph {
	if p == nil || p.nRegs <= 0 {
		panic("aig: forward retiming requires registers")
	}
	if err := p.Check(); err != nil {
		panic(err)
	}
	registers, pis, pos := sclLayout(p)
	steps = max(steps, 1)
	domains := retimeFrontierDomains(p, pis)
	Aig_ManRetimeMark(p)
	liByLo := make(map[int]Edge, registers)
	for i := 0; i < registers; i++ {
		lo := p.inputs[pis+i]
		p.node(lo).MarkA = true
		liByLo[lo.NodeID()] = p.outputs[pos+i]
	}
	cleanMarks := func() {
		for id := range liByLo {
			if !p.deleted[id] {
				p.nodes[id].MarkA = false
			}
		}
		for id := range p.nodes {
			if !p.deleted[id] {
				p.nodes[id].MarkB = false
			}
		}
	}
	defer cleanMarks()
	Aig_TableClear(p)
	var moves int32
	for changed := true; changed && moves < steps; {
		changed = false
		// New gates are eligible later in this same scan, just as ABC's object
		// iterator observes the growing manager. Never retain node pointers across
		// append: Graph owns a slice whose backing array can move.
		for id := 0; id < len(p.nodes) && moves < steps; id++ {
			if p.deleted[id] {
				continue
			}
			node := p.nodes[id]
			if !node.MarkB || (node.Kind != And && node.Kind != ExclusiveOr) {
				continue
			}
			first := Aig_ObjReal_rec(p, node.Fanins[0])
			second := Aig_ObjReal_rec(p, node.Fanins[1])
			li0, ok0 := liByLo[first.NodeID()]
			li1, ok1 := liByLo[second.NodeID()]
			if !ok0 || !ok1 || domains[first.NodeID()] != domains[second.NodeID()] {
				continue
			}
			a, b := p.outputDriver(li0), p.outputDriver(li1)
			if first.Complemented() {
				a = a.Not()
			}
			if second.Complemented() {
				b = b.Not()
			}
			complemented := first.Complemented() && second.Complemented()
			var result Edge
			if node.Kind == ExclusiveOr {
				result = p.NewXor(a, b)
				complemented = first.Complemented() != second.Complemented()
			} else {
				result = p.AND(a, b)
			}
			if p.node(result).Kind != Constant {
				p.node(result).MarkB = true
			}
			lo := Aig_ObjCreateCi(p)
			p.node(lo).MarkA = true
			next := result
			if complemented {
				next = next.Not()
			}
			li := Aig_ObjCreateCo(p, next)
			domain := domains[first.NodeID()]
			if domain >= 0 {
				p.ClockDomains[domain] = append(p.ClockDomains[domain], p.nRegs)
			}
			domains[lo.NodeID()] = domain
			p.nRegs++
			liByLo[lo.NodeID()] = li
			p.FlopNumbers = nil
			p.FlopReprs = nil
			p.RetimingChanged = true
			old := Edge(id) << 1
			Aig_ObjDisconnect(p, old)
			p.node(old).Kind = Buffer
			replacement := lo
			if complemented {
				replacement = replacement.Not()
			}
			p.node(old).Phase = complemented
			Aig_ObjConnect(p, old, replacement, InvalidEdge)
			moves++
			changed = true
		}
	}
	cleanMarks()
	Aig_ManSeqCleanup(p)
	return Aig_ManDupOrdered(p)
}
