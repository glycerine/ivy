package aig

const (
	// SupportMinSuccess means a care-equivalent cofactored cone was constructed.
	SupportMinSuccess        int32 = 0
	SupportMinNotOrCondition int32 = -1
	SupportMinInapplicable   int32 = -2
	SupportMinNoReduction    int32 = -3
	aigSupportLimit                = 16
)

// The factoring routines use typed cone frontiers and local truth tables rather
// than placing pointers, simulation words and temporary copies in one C union.
// Public traversal helpers retain ABC's current/previous traversal-ID protocol.

func Aig_ManFindImplications_rec(p *Graph, root Edge, implications *[]Edge) {
	seen := map[Edge]bool{}
	for _, e := range *implications {
		seen[e] = true
	}
	active, done := map[int]bool{}, map[int]bool{}
	var visit func(Edge)
	visit = func(e Edge) {
		n := p.node(e)
		if e.Complemented() || n.Kind != And {
			if !seen[e] {
				*implications = append(*implications, e)
				seen[e] = true
			}
			return
		}
		if active[e.NodeID()] {
			panic("aig: cyclic implication cone")
		}
		if done[e.NodeID()] {
			return
		}
		active[e.NodeID()] = true
		visit(n.Fanins[0])
		visit(n.Fanins[1])
		delete(active, e.NodeID())
		done[e.NodeID()] = true
	}
	visit(root)
}
func Aig_ManFindImplications(p *Graph, root Edge) []Edge {
	result := []Edge{}
	Aig_ManFindImplications_rec(p, root, &result)
	return result
}
func Aig_ManFindConeOverlap_rec(p *Graph, root Edge) bool {
	n := p.node(root)
	if n.TravID == p.travID-1 {
		return true
	}
	if n.TravID == p.travID {
		return false
	}
	n.TravID = p.travID
	for _, e := range nodeFanins(*n) {
		if Aig_ManFindConeOverlap_rec(p, e) {
			return true
		}
	}
	return false
}
func Aig_ManFindConeOverlap(p *Graph, implications []Edge, root Edge) bool {
	Aig_ManIncrementTravId(p)
	for _, e := range implications {
		p.node(e).TravID = p.travID
	}
	Aig_ManIncrementTravId(p)
	return Aig_ManFindConeOverlap_rec(p, root)
}
func aigFactRebuild(p *Graph, n Node, a, b Edge) Edge {
	switch n.Kind {
	case And:
		return p.AND(a, b)
	case ExclusiveOr:
		return p.NewXor(a, b)
	case Buffer, Output:
		return a
	}
	panic("aig: cannot rebuild a terminal as a gate")
}
func Aig_ManDeriveNewCone_rec(p *Graph, root Edge) Edge {
	e := Aig_Regular(root)
	n := *p.node(e)
	var result Edge
	if n.TravID == p.travID {
		var ok bool
		result, ok = p.copies[e.NodeID()]
		if !ok {
			panic("aig: missing current-traversal copy")
		}
	} else {
		p.node(e).TravID = p.travID
		if n.Kind == Input || n.Kind == Constant {
			result = e
		} else {
			fanins := nodeFanins(n)
			a := Aig_ManDeriveNewCone_rec(p, fanins[0])
			b := InvalidEdge
			if len(fanins) > 1 {
				b = Aig_ManDeriveNewCone_rec(p, fanins[1])
			}
			result = aigFactRebuild(p, n, a, b)
		}
		p.copies[e.NodeID()] = result
	}
	if root.Complemented() {
		result = result.Not()
	}
	return result
}
func Aig_ManDeriveNewCone(p *Graph, implications []Edge, root Edge) Edge {
	Aig_ManIncrementTravId(p)
	if p.copies == nil {
		p.copies = map[int]Edge{}
	}
	for _, e := range implications {
		p.node(e).TravID = p.travID
		value := p.Const1()
		if e.Complemented() {
			value = value.Not()
		}
		p.copies[e.NodeID()] = value
	}
	return Aig_ManDeriveNewCone_rec(p, root)
}
func Aig_ManFactorAlgebraic_int(p *Graph, outputA, outputB Edge, value int32) Edge {
	if value != 0 && value != 1 {
		panic("aig: factoring value must be zero or one")
	}
	a, b := p.outputDriver(outputA), p.outputDriver(outputB)
	if p.node(a).Kind == Constant || p.node(b).Kind == Constant || p.node(b).Kind == Input {
		return InvalidEdge
	}
	if value == 0 {
		a = a.Not()
	}
	implications := Aig_ManFindImplications(p, a)
	if !Aig_ManFindConeOverlap(p, implications, Aig_Regular(b)) {
		return InvalidEdge
	}
	return Aig_ManDeriveNewCone(p, implications, b)
}
func Aig_ManFactorAlgebraic(p *Graph, outputA, outputB, value int32) Edge {
	if outputA < 0 || int(outputA) >= len(p.outputs) || outputB < 0 || int(outputB) >= len(p.outputs) {
		panic("aig: factoring output index")
	}
	return Aig_ManFactorAlgebraic_int(p, p.outputs[outputA], p.outputs[outputB], value)
}
func Aig_ManFactorAlgebraicTest(p *Graph) Edge {
	result := Aig_ManFactorAlgebraic(p, 0, 1, 0)
	if result != InvalidEdge {
		Aig_ObjPrintVerbose(p, Aig_Regular(result), 0)
	}
	return result
}

// supportTruth builds cut truth tables with one bit per assignment. A support
// node is an independent boundary variable even when it is an internal gate.
func supportTruth(p *Graph, nodes, support []Edge) map[int][]bool {
	if len(support) > aigSupportLimit {
		panic("aig: support truth limit is sixteen variables")
	}
	assignments := 1 << len(support)
	tables := make(map[int][]bool, len(nodes)+len(support)+1)
	one := make([]bool, assignments)
	for i := range one {
		one[i] = true
	}
	tables[0] = one
	seenSupport := map[int]bool{}
	for i, e := range support {
		if e.NodeID() == 0 || seenSupport[e.NodeID()] {
			panic("aig: support boundaries must be distinct nonconstant nodes")
		}
		seenSupport[e.NodeID()] = true
		if e.Complemented() {
			panic("aig: complemented support boundary")
		}
		values := make([]bool, assignments)
		for mask := range values {
			values[mask] = mask&(1<<i) != 0
		}
		tables[e.NodeID()] = values
	}
	for _, e := range nodes {
		n := *p.node(e)
		fanins := nodeFanins(n)
		if len(fanins) == 0 {
			panic("aig: terminal in reconstructed cone")
		}
		a := tables[fanins[0].NodeID()]
		if a == nil {
			panic("aig: support does not bound the cone")
		}
		var b []bool
		if len(fanins) > 1 {
			b = tables[fanins[1].NodeID()]
			if b == nil {
				panic("aig: support does not bound the cone")
			}
		}
		values := make([]bool, assignments)
		for mask := range values {
			av := a[mask] != fanins[0].Complemented()
			value := av
			if len(fanins) > 1 {
				bv := b[mask] != fanins[1].Complemented()
				if n.Kind == ExclusiveOr {
					value = av != bv
				} else {
					value = av && bv
				}
			}
			values[mask] = value
		}
		tables[e.NodeID()] = values
	}
	return tables
}

// A cofactor is a literal forced true: a complemented Edge forces its node
// to zero, and an uncomplemented Edge forces its node to one.
// Each accepted cofactor is checked against the *updated* function on the care
// set. Individually valid cofactors need not remain jointly valid; progressive
// checking prevents that source-level heuristic from changing cared values.
func Aig_SuppMinPerform(p *Graph, orGate, nodes, support []Edge) []Edge {
	if len(nodes) == 0 {
		return nil
	}
	tables := supportTruth(p, nodes, support)
	function := tables[nodes[len(nodes)-1].NodeID()]
	care := make([]bool, len(function))
	for _, e := range orGate {
		values := tables[e.NodeID()]
		if values == nil {
			panic("aig: care gate outside support")
		}
		for mask, v := range values {
			care[mask] = care[mask] || (v != e.Complemented())
		}
	}
	cofactors := []Edge{}
	for i, e := range support {
		bit := 1 << i
		for _, value := range []bool{false, true} {
			legal := true
			for mask := range function {
				cofactor := mask &^ bit
				if value {
					cofactor = mask | bit
				}
				if care[mask] && function[mask] != function[cofactor] {
					legal = false
					break
				}
			}
			if !legal {
				continue
			}
			next := make([]bool, len(function))
			for mask := range next {
				cofactor := mask &^ bit
				if value {
					cofactor = mask | bit
				}
				next[mask] = function[cofactor]
			}
			function = next
			if !value {
				e = e.Not()
			}
			cofactors = append(cofactors, e)
			break
		}
	}
	return cofactors
}
func Aig_SuppMinReconstruct(p *Graph, cofactors, nodes, support []Edge) Edge {
	if len(nodes) == 0 {
		return InvalidEdge
	}
	if p.copies == nil {
		p.copies = make(map[int]Edge)
	}
	copies := p.copies
	copies[0] = p.Const1()
	for _, e := range support {
		if e.Complemented() {
			panic("aig: complemented support boundary")
		}
		copies[e.NodeID()] = e
	}
	for _, e := range cofactors {
		value := p.Const1()
		if e.Complemented() {
			value = value.Not()
		}
		copies[e.NodeID()] = value
	}
	child := func(e Edge) Edge {
		copy, ok := copies[e.NodeID()]
		if !ok {
			panic("aig: reconstruction misses a cut boundary")
		}
		if e.Complemented() {
			copy = copy.Not()
		}
		return copy
	}
	var result Edge
	for _, e := range nodes {
		n := *p.node(e)
		fanins := nodeFanins(n)
		a := child(fanins[0])
		b := InvalidEdge
		if len(fanins) > 1 {
			b = child(fanins[1])
		}
		result = aigFactRebuild(p, n, a, b)
		copies[e.NodeID()] = result
	}
	return result
}
func Aig_SuppMinGateIsInSupport(p *Graph, orGate, support []Edge) bool {
	Aig_ManIncrementTravId(p)
	for _, e := range support {
		p.node(e).TravID = p.travID
	}
	for _, e := range orGate {
		if p.node(e).TravID != p.travID {
			return false
		}
	}
	return true
}
func Aig_SuppMinCollectSupport(p *Graph, nodes []Edge) []Edge {
	support := []Edge{}
	for _, e := range nodes {
		for _, f := range nodeFanins(*p.node(e)) {
			if f.NodeID() == 0 || p.node(f).TravID == p.travID {
				continue
			}
			f = Aig_Regular(f)
			p.node(f).TravID = p.travID
			support = append(support, f)
		}
	}
	return support
}
func Aig_SuppMinCollectCone_rec(p *Graph, root Edge, nodes *[]Edge) {
	e := Aig_Regular(root)
	n := p.node(e)
	if n.TravID != p.travID-1 {
		return
	}
	n.TravID = p.travID
	if len(nodeFanins(*n)) == 0 {
		panic("aig: terminal marked as internal cone node")
	}
	for _, f := range nodeFanins(*n) {
		Aig_SuppMinCollectCone_rec(p, f, nodes)
	}
	*nodes = append(*nodes, e)
}
func Aig_SuppMinCollectCone(p *Graph, root Edge) []Edge {
	Aig_ManIncrementTravId(p)
	nodes := []Edge{}
	Aig_SuppMinCollectCone_rec(p, root, &nodes)
	return nodes
}
func Aig_SuppMinHighlightCone_rec(p *Graph, root Edge) bool {
	n := p.node(root)
	if n.TravID == p.travID {
		return true
	}
	if n.TravID == p.travID-1 {
		return false
	}
	n.TravID = p.travID - 1
	overlaps := false
	for _, f := range nodeFanins(*n) {
		overlaps = Aig_SuppMinHighlightCone_rec(p, f) || overlaps
	}
	if overlaps {
		n.TravID = p.travID
	}
	return overlaps
}
func Aig_SuppMinHighlightCone(p *Graph, root Edge, orGate []Edge) bool {
	Aig_ManIncrementTravId(p)
	Aig_ManIncrementTravId(p)
	for _, e := range orGate {
		p.node(e).TravID = p.travID
	}
	overlaps := Aig_SuppMinHighlightCone_rec(p, root)
	for _, e := range orGate {
		p.node(e).TravID = p.travID - 1
	}
	return overlaps
}
func Aig_SuppMinCollectSuper_rec(p *Graph, root Edge, super *[]Edge) {
	implications := Aig_ManFindImplications(p, root)
	known := map[Edge]bool{}
	for _, e := range *super {
		known[e] = true
	}
	for _, e := range implications {
		e = e.Not()
		if !known[e] {
			known[e] = true
			*super = append(*super, e)
		}
	}
}
func Aig_SuppMinCollectSuper(p *Graph, root Edge) []Edge {
	super := []Edge{}
	for _, e := range nodeFanins(*p.node(root)) {
		Aig_SuppMinCollectSuper_rec(p, e, &super)
	}
	return super
}
func Aig_ManSupportMinimization(p *Graph, condition, function Edge) (Edge, int32) {
	if !condition.Complemented() || p.node(condition).Kind != And {
		return InvalidEdge, SupportMinNotOrCondition
	}
	kind := p.node(function).Kind
	if kind != And && kind != ExclusiveOr {
		return InvalidEdge, SupportMinInapplicable
	}
	orGate := Aig_SuppMinCollectSuper(p, Aig_Regular(condition))
	if len(orGate) > aigSupportLimit {
		return InvalidEdge, SupportMinInapplicable
	}
	if !Aig_SuppMinHighlightCone(p, Aig_Regular(function), orGate) {
		return InvalidEdge, SupportMinInapplicable
	}
	nodes := Aig_SuppMinCollectCone(p, Aig_Regular(function))
	if len(nodes) == 0 {
		return InvalidEdge, SupportMinNoReduction
	}
	support := Aig_SuppMinCollectSupport(p, nodes)
	if len(support) > aigSupportLimit {
		return InvalidEdge, SupportMinInapplicable
	}
	if !Aig_SuppMinGateIsInSupport(p, orGate, support) {
		return InvalidEdge, SupportMinNoReduction
	}
	cofactors := Aig_SuppMinPerform(p, orGate, nodes, support)
	if len(cofactors) == 0 {
		return InvalidEdge, SupportMinNoReduction
	}
	result := Aig_SuppMinReconstruct(p, cofactors, nodes, support)
	if function.Complemented() {
		result = result.Not()
	}
	return result, SupportMinSuccess
}
func Aig_ManSupportMinimizationTest() (*Graph, Edge, int32) {
	p := NewGraph()
	for i := int32(0); i < 5; i++ {
		Aig_IthVar(p, i)
	}
	function := p.Mux(Aig_IthVar(p, 3), Aig_IthVar(p, 1), Aig_IthVar(p, 0))
	function = p.Mux(Aig_IthVar(p, 4), Aig_IthVar(p, 2), function)
	condition := p.OR(Aig_IthVar(p, 3), Aig_IthVar(p, 4))
	result, status := Aig_ManSupportMinimization(p, condition, function)
	return p, result, status
}
func Aig_ManSupportMinimizationTest2() (*Graph, Edge, int32) {
	p := NewGraph()
	a, b, c := p.NewInput(), p.NewInput(), p.NewInput()
	n9 := p.AND(a, b.Not())
	n10 := p.AND(n9.Not(), c.Not())
	function := p.AND(n10, b)
	condition := p.OR(p.OR(b, c), a)
	result, status := Aig_ManSupportMinimization(p, condition, function)
	return p, result, status
}
