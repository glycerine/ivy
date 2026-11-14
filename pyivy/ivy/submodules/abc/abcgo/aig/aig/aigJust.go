package aig

const (
	AIG_VAL0 int32 = 1
	AIG_VAL1 int32 = 2
	AIG_VALX int32 = 3
)

func Aig_ObjSetTerValue(p *Graph, e Edge, value int32) int32 {
	if value < AIG_VAL0 || value > AIG_VALX {
		panic("aig: invalid ternary value")
	}
	p.node(e).MarkA = value&1 != 0
	p.node(e).MarkB = value&2 != 0
	return value
}
func Aig_ObjGetTerValue(p *Graph, e Edge) int32 {
	var value int32
	if p.node(e).MarkA {
		value |= 1
	}
	if p.node(e).MarkB {
		value |= 2
	}
	return value
}
func Aig_ObjNotTerValue(value int32) int32 {
	switch value {
	case AIG_VAL0:
		return AIG_VAL1
	case AIG_VAL1:
		return AIG_VAL0
	default:
		return AIG_VALX
	}
}
func Aig_ObjAndTerValue(a, b int32) int32 {
	if a == AIG_VAL0 || b == AIG_VAL0 {
		return AIG_VAL0
	}
	if a == AIG_VAL1 && b == AIG_VAL1 {
		return AIG_VAL1
	}
	return AIG_VALX
}
func Aig_ObjNotCondTerValue(value, condition int32) int32 {
	if condition != 0 {
		return Aig_ObjNotTerValue(value)
	}
	return value
}
func Aig_ObjSatValue(p *Graph, e Edge, complement int32) int32 {
	n := p.node(e)
	if n.Kind == Constant {
		return Aig_ObjNotCondTerValue(AIG_VAL1, complement^int32(e&1))
	}
	if n.TravID != p.travID {
		return AIG_VALX
	}
	value := AIG_VAL0
	if n.MarkA {
		value = AIG_VAL1
	}
	return Aig_ObjNotCondTerValue(value, complement^int32(e&1))
}

// Aig_NtkFindSatAssign_rec performs ABC's bounded greedy justification. A false
// result means this attempt failed; it is not an UNSAT proof.
func Aig_NtkFindSatAssign_rec(p *Graph, e Edge, value int32, support *[]int32, heur int32) bool {
	if value != 0 && value != 1 {
		panic("aig: justification requires Boolean target")
	}
	if e.Complemented() {
		value ^= 1
		e = Aig_Regular(e)
	}
	n := p.node(e)
	if n.Kind == Constant {
		return value == 1
	}
	if n.TravID == p.travID {
		return n.MarkA == (value != 0)
	}
	n.TravID = p.travID
	n.MarkA = value != 0
	if n.Kind == Input {
		if support != nil {
			index := Aig_ObjCioId(p, e)
			if index < 0 || int(index) >= len(p.inputs) || p.inputs[index] != e {
				panic("aig: justification requires CI IDs indexing the current interface")
			}
			*support = append(*support, index*2+(1-value))
		}
		return true
	}
	a, b := n.Fanins[0], n.Fanins[1]
	if n.Kind == Output || n.Kind == Buffer {
		return Aig_NtkFindSatAssign_rec(p, a, value, support, heur)
	}
	if n.Kind != And && n.Kind != ExclusiveOr {
		panic("aig: invalid justification node")
	}
	va, vb := Aig_ObjSatValue(p, a, 0), Aig_ObjSatValue(p, b, 0)
	if n.Kind == ExclusiveOr {
		first := int32(0)
		if heur&1 != 0 {
			first = 1
		}
		if va == AIG_VAL0 {
			first = 0
		} else if va == AIG_VAL1 {
			first = 1
		} else if vb == AIG_VAL0 {
			first = value
		} else if vb == AIG_VAL1 {
			first = value ^ 1
		}
		return Aig_NtkFindSatAssign_rec(p, a, first, support, heur) && Aig_NtkFindSatAssign_rec(p, b, first^value, support, heur)
	}
	if value == 1 {
		return Aig_NtkFindSatAssign_rec(p, a, 1, support, heur) && Aig_NtkFindSatAssign_rec(p, b, 1, support, heur)
	}
	if va == AIG_VAL0 || vb == AIG_VAL0 {
		return true
	}
	if va == AIG_VAL1 && vb == AIG_VAL1 {
		return false
	}
	if va == AIG_VAL1 {
		return Aig_NtkFindSatAssign_rec(p, b, 0, support, heur)
	}
	if vb == AIG_VAL1 {
		return Aig_NtkFindSatAssign_rec(p, a, 0, support, heur)
	}
	heur++
	if heur%8 == 0 {
		return Aig_NtkFindSatAssign_rec(p, b, 0, support, heur)
	}
	return Aig_NtkFindSatAssign_rec(p, a, 0, support, heur)
}

// On failure support retains the last attempt's partial assignments, as in C;
// these are diagnostic state, not a satisfying witness or an UNSAT certificate.
// Collecting support requires current CI indices (Aig_ManSetCioIds).
func Aig_ObjFindSatAssign(p *Graph, e Edge, value int32, support *[]int32) bool {
	if n := p.node(e); n.Kind == Output {
		driver := n.Fanins[0]
		if e.Complemented() {
			driver = driver.Not()
		}
		return Aig_ObjFindSatAssign(p, driver, value, support)
	}
	for attempt := int32(0); attempt < 8; attempt++ {
		if support != nil {
			*support = (*support)[:0]
		}
		Aig_ManIncrementTravId(p)
		if Aig_NtkFindSatAssign_rec(p, e, value, support, attempt) {
			return true
		}
	}
	return false
}
func Aig_ObjTerSimulate_rec(p *Graph, e Edge) int32 {
	n := p.node(e)
	if n.Kind == Constant {
		return Aig_ObjNotCondTerValue(AIG_VAL1, int32(e&1))
	}
	if n.TravID == p.travID {
		return Aig_ObjNotCondTerValue(Aig_ObjGetTerValue(p, e), int32(e&1))
	}
	n.TravID = p.travID
	if n.Kind == Input {
		return Aig_ObjSetTerValue(p, e, AIG_VALX)
	}
	a := Aig_ObjTerSimulate_rec(p, n.Fanins[0])
	value := a
	if n.Kind == And {
		if a != AIG_VAL0 {
			value = Aig_ObjAndTerValue(a, Aig_ObjTerSimulate_rec(p, n.Fanins[1]))
		}
	} else if n.Kind == ExclusiveOr {
		b := Aig_ObjTerSimulate_rec(p, n.Fanins[1])
		if a == AIG_VALX || b == AIG_VALX {
			value = AIG_VALX
		} else if a == b {
			value = AIG_VAL0
		} else {
			value = AIG_VAL1
		}
	} else if n.Kind != Output && n.Kind != Buffer {
		panic("aig: invalid ternary node")
	}
	Aig_ObjSetTerValue(p, e, value)
	return Aig_ObjNotCondTerValue(value, int32(e&1))
}
func Aig_ObjTerSimulate(p *Graph, e Edge, support []int32) int32 {
	Aig_ManIncrementTravId(p)
	for _, literal := range support {
		index := literal >> 1
		if index < 0 || int(index) >= len(p.inputs) {
			panic("aig: invalid support literal")
		}
		input := p.inputs[index]
		value := AIG_VAL1
		if literal&1 != 0 {
			value = AIG_VAL0
		}
		if p.node(input).TravID == p.travID && Aig_ObjGetTerValue(p, input) != value {
			panic("aig: conflicting support literals")
		}
		Aig_ObjSetTerValue(p, input, value)
		p.node(input).TravID = p.travID
	}
	return Aig_ObjTerSimulate_rec(p, e)
}
