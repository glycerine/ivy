package cec

import "fmt"

// cec4PatternPacker owns the simulation bits and reservation masks used by
// Cec4 to combine compatible partial assignments in the same pattern slot.
type cec4PatternPacker struct {
	values, reserved *cec3Simulation
	cursor           int
}

func newCec4PatternPacker(objects, words int) *cec4PatternPacker {
	return &cec4PatternPacker{values: newCec3Simulation(objects, words), reserved: newCec3Simulation(objects, words)}
}
func (p *cec4PatternPacker) bit(s *cec3Simulation, id int32, bit int) bool {
	return s.object(id)[bit>>6]&(uint64(1)<<uint(bit&63)) != 0
}
func (p *cec4PatternPacker) try(bit int, literals []int32) bool {
	for _, lit := range literals {
		id := lit >> 1
		if p.bit(p.reserved, id, bit) && p.bit(p.values, id, bit) != (lit&1 != 0) {
			return false
		}
	}
	for _, lit := range literals {
		id := lit >> 1
		p.reserved.setInputBit(id, bit, true)
		p.values.setInputBit(id, bit, lit&1 != 0)
	}
	return true
}
func (p *cec4PatternPacker) extend(bit int, literals []int32) {
	limit := len(literals)
	if capBits := p.values.words*64 - 1; limit > capBits {
		limit = capBits
	}
	for k := 0; k < limit; k++ {
		local := (bit+k+1)%limit + 1
		for i, lit := range literals {
			id := lit >> 1
			if !p.bit(p.reserved, id, local) {
				p.values.setInputBit(id, local, (lit&1 != 0) != (i == k))
			}
		}
	}
}

// add returns the number of slots examined; a full pack returns words*64.
func (p *cec4PatternPacker) add(literals []int32, extend bool) int {
	limit := p.values.words * 64
	for k := 1; k < limit-1; k++ {
		p.cursor++
		if p.cursor == limit-1 {
			p.cursor = 1
		}
		if p.try(p.cursor, literals) {
			if extend {
				p.extend(p.cursor, literals)
			}
			return k
		}
	}
	p.cursor = limit - 1
	if !p.try(p.cursor, literals) {
		fmt.Println("Internal error.")
		return limit
	}
	if extend {
		p.extend(p.cursor, literals)
	}
	return limit
}
