package sfm

import "github.com/glycerine/abcgo/map/mio"

// sfmDecSimulation holds normal and pivot-flipped 64-pattern simulations.
type sfmDecSimulation struct {
	normal, flipped []uint64
	affected        []bool
}

func newSfmDecSimulation(n int) *sfmDecSimulation {
	return &sfmDecSimulation{normal: make([]uint64, n), flipped: make([]uint64, n), affected: make([]bool, n)}
}
func (s *sfmDecSimulation) evaluate(id int, gate *mio.Mio_Gate_t, fanins []int, flipped bool) uint64 {
	if len(fanins) > 6 {
		panic("sfm: simulation supports at most six inputs")
	}
	values := make([]uint64, len(fanins))
	for i, fan := range fanins {
		values[i] = s.normal[fan]
		if flipped && s.affected[fan] {
			values[i] = s.flipped[fan]
		}
	}
	return mio.Exp_Truth6(int32(len(fanins)), mio.Mio_GateReadExpr(gate), values)
}
func (s *sfmDecSimulation) update(id int, gate *mio.Mio_Gate_t, fanins []int) {
	s.normal[id] = s.evaluate(id, gate, fanins, false)
	if s.affected[id] {
		s.flipped[id] = s.evaluate(id, gate, fanins, true)
	}
}
func (s *sfmDecSimulation) flip(id int) { s.affected[id] = true; s.flipped[id] = ^s.normal[id] }
func (s *sfmDecSimulation) care(roots []int) uint64 {
	var result uint64
	for _, id := range roots {
		result |= s.normal[id] ^ s.flipped[id]
	}
	return result
}
