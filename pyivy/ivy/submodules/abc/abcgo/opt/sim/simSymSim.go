package sim

// Sim_SymmsCreateSquare puts P with input i flipped in simulation column i.
func Sim_SymmsCreateSquare(p *Sym_Man_t, pattern []uint32) {
	for i, e := range p.Inputs {
		row := p.Sim[e.NodeID()]
		Sim_UtilSetConst(row, p.Words, Sim_HasBit(pattern, i))
		Sim_XorBit(row, i)
	}
}
func Sim_SymmsSimulate(p *Sym_Man_t, pattern []uint32, nonSym []Matrix) {
	Sim_SymmsCreateSquare(p, pattern)
	for _, e := range p.Order {
		Sim_UtilSimulateNodeOne(p.Graph, e, p.Sim, p.Words, 0)
	}
	for o, e := range p.Outputs {
		Sim_SymmsDeriveInfo(p, pattern, p.Sim[e.NodeID()], nonSym[o], o)
	}
}

// For equal P bits, differing single-flip outputs witness non-symmetry. For
// unequal P bits those flips produce 00/11, which say nothing about a swap.
func Sim_SymmsDeriveInfo(p *Sym_Man_t, pattern, values []uint32, m Matrix, output int) {
	words := SIM_NUM_WORDS(len(m))
	col, row := make([]uint32, words), make([]uint32, words)
	// Restrict padding and irrelevant variables through the functional support.
	var support []uint32
	if p.Functional != nil {
		support = p.Functional[output]
	} else {
		support = make([]uint32, words)
		for _, i := range p.Supports[output] {
			Sim_SetBit(support, i)
		}
	}
	for polarity := 0; polarity < 2; polarity++ {
		for w := 0; w < words; w++ {
			base := pattern[w]
			if polarity != 0 {
				base = ^base
			}
			col[w] = support[w] & base & values[w]
			row[w] = support[w] & base &^ values[w]
		}
		for _, i := range p.Supports[output] {
			if Sim_HasBit(col, i) {
				Sim_UtilInfoAdd(m[i], row, words)
			}
			if Sim_HasBit(row, i) {
				Sim_UtilInfoAdd(m[i], col, words)
			}
		}
	}
}
