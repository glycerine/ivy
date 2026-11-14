package aig

// Aig_ManCheck verifies structural invariants and preserves ABC's 1/0 result.
func Aig_ManCheck(p *Graph) int32 {
	if p == nil || p.Check() != nil {
		return 0
	}
	return 1
}

// Aig_ManCheckMarkA detects marks left behind by temporary traversals.
func Aig_ManCheckMarkA(p *Graph) {
	for id, n := range p.nodes {
		if !p.deleted[id] && n.MarkA {
			panic("aig: traversal mark A was not cleared")
		}
	}
}

// Aig_ManCheckPhase validates each node's all-zero-input simulation value.
func Aig_ManCheckPhase(p *Graph) {
	for id, n := range p.nodes {
		if p.deleted[id] {
			continue
		}
		phase := false
		switch n.Kind {
		case Constant:
			phase = true
		case Input:
			phase = false
		case And, ExclusiveOr, Buffer, Output:
			for _, fanin := range nodeFanins(n) {
				if !p.valid(fanin) {
					panic("aig: invalid phase-check fanin")
				}
			}
			phase = Aig_ObjPhaseReal(p, n.Fanins[0])
			if n.Kind == And {
				phase = phase && Aig_ObjPhaseReal(p, n.Fanins[1])
			} else if n.Kind == ExclusiveOr {
				phase = phase != Aig_ObjPhaseReal(p, n.Fanins[1])
			}
		default:
			panic("aig: unknown node type")
		}
		if n.Phase != phase {
			panic("aig: inconsistent phase assignment")
		}
	}
}
