package fra

// Packed counterexample simulation from Fra_ClausSimInfoRecord/Check.
type clausModels struct {
	values [][]uint32
	count  int
}

func newClausModels(variables int) *clausModels {
	if variables < 0 {
		panic("negative variable count")
	}
	return &clausModels{values: make([][]uint32, variables)}
}
func (p *clausModels) clear() {
	for _, v := range p.values {
		clear(v)
	}
	p.count = 0
}
func (p *clausModels) record(model []bool) {
	if len(model) != len(p.values) {
		panic("invalid counterexample model")
	}
	word, bit := p.count/32, uint(p.count%32)
	for i, value := range model {
		if word == len(p.values[i]) {
			p.values[i] = append(p.values[i], 0)
		}
		if value {
			p.values[i][word] |= uint32(1) << bit
		}
	}
	p.count++
}

// disproves reports whether any recorded assignment falsifies every literal.
// frameOffset is the variable offset of the last unrolled timeframe.
func (p *clausModels) disproves(clause []int32, frameOffset int32) bool {
	for word := 0; word < (p.count+31)/32; word++ {
		mask := ^uint32(0)
		for _, lit := range clause {
			variable := int((lit >> 1) - frameOffset)
			if lit < 0 || variable < 0 || variable >= len(p.values) {
				panic("invalid simulated clause variable")
			}
			bits := p.values[variable][word]
			if lit&1 == 0 {
				bits = ^bits
			}
			mask &= bits
		}
		if word == p.count/32 && p.count%32 != 0 {
			mask &= (uint32(1) << uint(p.count%32)) - 1
		}
		if mask != 0 {
			return true
		}
	}
	return false
}
