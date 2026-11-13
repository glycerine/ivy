package cec

// cec3Simulation owns the per-object simulation blocks that Cec3 stores on
// the GIA manager in C. Each block has the same number of 64-bit words.
type cec3Simulation struct {
	words int
	data  []uint64
}

func newCec3Simulation(objects, words int) *cec3Simulation {
	if objects < 1 || words < 1 {
		panic("cec: invalid simulation dimensions")
	}
	return &cec3Simulation{words: words, data: make([]uint64, objects*words)}
}

func (s *cec3Simulation) object(id int32) []uint64 {
	if id < 0 || int(id) >= len(s.data)/s.words {
		panic("cec: simulation object out of range")
	}
	start := int(id) * s.words
	return s.data[start : start+s.words : start+s.words]
}

// setInputBit implements Cec3_ObjSimSetInputBit over 64-bit Go words.
func (s *cec3Simulation) setInputBit(id int32, bit int, value bool) {
	words := s.object(id)
	if bit < 0 || bit >= 64*len(words) {
		panic("cec: simulation bit out of range")
	}
	mask := uint64(1) << uint(bit&63)
	if value {
		words[bit>>6] |= mask
	} else {
		words[bit>>6] &^= mask
	}
}

// copyDriver and andFanins implement the original output and AND simulation.
func (s *cec3Simulation) copyDriver(out, driver int32, complemented bool) {
	dst, src := s.object(out), s.object(driver)
	for i, word := range src {
		if complemented {
			word = ^word
		}
		dst[i] = word
	}
}

func (s *cec3Simulation) andFanins(out, a, b int32, complementA, complementB bool) {
	dst, left, right := s.object(out), s.object(a), s.object(b)
	for i := range dst {
		x, y := left[i], right[i]
		if complementA {
			x = ^x
		}
		if complementB {
			y = ^y
		}
		dst[i] = x & y
	}
}

// equivalent matches Cec3_ObjSimEqual, accepting an overall complement.
func (s *cec3Simulation) equivalent(a, b int32) bool {
	x, y := s.object(a), s.object(b)
	phase := uint64(0)
	if (x[0]^y[0])&1 != 0 {
		phase = ^uint64(0)
	}
	for i := range x {
		if x[i] != y[i]^phase {
			return false
		}
	}
	return true
}
