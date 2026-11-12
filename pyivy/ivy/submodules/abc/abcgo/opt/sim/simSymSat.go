package sim

func Sim_SymmsSatProveOne(p *Sym_Man_t, out, a, b int, pattern []uint32) bool {
	different, model := cofactorDifference(p.Graph, out, a, b)
	if different {
		copy(pattern, model)
	}
	return !different
}

// Sim_SymmsGetPatternUsingSat proves unresolved pairs until it finds a swap
// counterexample. Equivalence classes propagate both proofs and disproofs.
func Sim_SymmsGetPatternUsingSat(p *Sym_Man_t, pattern []uint32) bool {
	for o := p.Output; o < len(p.Outputs); o++ {
		s := p.Supports[o]
		sym, non := p.Symmetric[o], p.NonSymmetric[o]
		for i, a := range s {
			for _, b := range s[i+1:] {
				if sym.Has(a, b) || non.Has(a, b) {
					continue
				}
				var ca, cb []int
				for _, v := range s {
					if sym.Has(a, v) {
						ca = append(ca, v)
					}
					if sym.Has(b, v) {
						cb = append(cb, v)
					}
				}
				p.SATRuns++
				if Sim_SymmsSatProveOne(p, o, a, b, pattern) {
					p.UNSAT++
					// Non-symmetry rows of equivalent variables must be united before class merge.
					union := make([]uint32, p.Words)
					merged := append(append([]int(nil), ca...), cb...)
					for _, v := range merged {
						Sim_UtilInfoAdd(union, non[v], p.Words)
					}
					for _, x := range merged {
						for _, y := range s {
							if Sim_HasBit(union, y) {
								non.Set(x, y)
							}
						}
					}
					for _, x := range ca {
						for _, y := range cb {
							sym.Set(x, y)
						}
					}
				} else {
					p.SAT++
					for _, x := range ca {
						for _, y := range cb {
							non.Set(x, y)
						}
					}
					p.Output, p.Var1, p.Var2 = o, a, b
					return true
				}
			}
		}
	}
	p.Output = len(p.Outputs)
	return false
}
