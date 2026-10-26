package abc

// Packed clause metadata from sat/cnf/cnfUtil.c:Cnf_DataDeriveLitPolarities.

// LiteralPolarities derives two-bit masks for the first four tail literals of
// each object-owned clause. A negative occurrence contributes bit 0; a positive
// occurrence bit 1. Masks combine all clauses of the same owner/head polarity.
// Wider clauses participate in collection but only four pairs fit the byte.
func (c *AigObjectCNF) LiteralPolarities(o CNFTransformOptions) ([]uint8, error) {
	if c == nil || c.AigMappedCNF == nil {
		return nil, ErrDimension
	}
	if e := checkCNFTransform(c.AigMappedCNF, o); e != nil {
		return nil, e
	}
	if len(c.starts) != c.variables || len(c.counts) != c.variables {
		return nil, ErrDimension
	}
	stop := func() bool { return o.Stop != nil && o.Stop() }
	pols := [2][]uint8{make([]uint8, c.variables), make([]uint8, c.variables)}
	result := make([]uint8, len(c.clauses))
	covered := make([]bool, len(c.clauses))
	for owner, count := range c.counts {
		if stop() {
			return nil, ErrCancelled
		}
		if count < -1 {
			return nil, ErrDimension
		}
		if count <= 0 {
			continue
		}
		start := c.starts[owner]
		if start < 0 || count > len(c.clauses)-start {
			return nil, ErrDimension
		}
		for index := start; index < start+count; index++ {
			if stop() {
				return nil, ErrCancelled
			}
			clause := c.clauses[index]
			if covered[index] || len(clause) == 0 || len(clause) > 17 || int(clause[0].Var()) != owner {
				return nil, ErrDimension
			}
			covered[index] = true
			polarity := int(clause[0] & 1)
			for _, l := range clause[1:] {
				pols[polarity][l.Var()] |= 2 - uint8(l&1)
			}
		}
		for index := start; index < start+count; index++ {
			clause := c.clauses[index]
			polarity := int(clause[0] & 1)
			for slot, l := range clause[1:min(len(clause), 5)] {
				result[index] |= pols[polarity][l.Var()] << uint(2*slot)
			}
		}
		for index := start; index < start+count; index++ {
			for _, l := range c.clauses[index][1:] {
				pols[0][l.Var()] = 0
				pols[1][l.Var()] = 0
			}
		}
	}
	for _, ok := range covered {
		if !ok {
			return nil, ErrDimension
		}
	}
	if stop() {
		return nil, ErrCancelled
	}
	return result, nil
}
