package core

// Cudd_addResidue builds the ADD whose leaves are the residue modulo modulus
// of the selected bit variables. By default bit index top is the least
// significant bit; CUDD_RESIDUE_MSB reverses that order and CUDD_RESIDUE_TC
// interprets the most significant bit as a two's-complement sign bit.
func Cudd_addResidue(dd *DdManager, bits, modulus, options, top int) *DdNode {
	if dd == nil || dd.closed || bits < 1 || modulus < 2 || top < 0 || uint64(top)+uint64(bits) > uint64(CUDD_MAXINDEX) {
		return nil
	}
	values := make([]*DdNode, modulus)
	for residue := range values {
		values[residue] = Cudd_addConst(dd, float64(residue))
		if values[residue] == nil {
			return nil
		}
	}
	weight := 1 % modulus
	for bit := 0; bit < bits; bit++ {
		index := top + bit
		if options&CUDD_RESIDUE_MSB != 0 {
			index = top + bits - bit - 1
		}
		variable := Cudd_addIthVar(dd, index)
		if variable == nil {
			return nil
		}
		if options&CUDD_RESIDUE_TC != 0 && bit == bits-1 {
			weight = (modulus - weight) % modulus
		}
		next := make([]*DdNode, modulus)
		for value := range next {
			whenSet := (value + weight) % modulus
			next[value] = Cudd_addIte(dd, variable, values[whenSet], values[value])
			if next[value] == nil {
				return nil
			}
		}
		values = next
		weight = (2 * weight) % modulus
	}
	return values[0]
}
