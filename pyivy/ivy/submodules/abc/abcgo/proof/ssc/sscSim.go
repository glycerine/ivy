package ssc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"math/bits"
)

func Ssc_SimAnd(out, a, b []uint64, n int, ca, cb bool) {
	for i := 0; i < n; i++ {
		x, y := a[i], b[i]
		if ca {
			x = ^x
		}
		if cb {
			y = ^y
		}
		out[i] = x & y
	}
}
func Ssc_SimDup(out, a []uint64, n int, c bool) {
	for i := 0; i < n; i++ {
		out[i] = a[i]
		if c {
			out[i] = ^out[i]
		}
	}
}
func Ssc_SimConst(out []uint64, n int, c bool) {
	for i := 0; i < n; i++ {
		out[i] = 0
		if c {
			out[i] = ^uint64(0)
		}
	}
}
func Ssc_SimOr(out, a []uint64, n int, c bool) {
	for i := 0; i < n; i++ {
		x := a[i]
		if c {
			x = ^x
		}
		out[i] |= x
	}
}
func Ssc_SimFindBitWord(x uint64) int32 {
	if x == 0 {
		return -1
	}
	return int32(bits.TrailingZeros64(x))
}
func Ssc_SimFindBit(a []uint64, n int) int32 {
	for i, x := range a[:n] {
		if x != 0 {
			return int32(64*i) + Ssc_SimFindBitWord(x)
		}
	}
	return -1
}
func Ssc_SimCountBitsWord(x uint64) int32 { return int32(bits.OnesCount64(x)) }
func Ssc_SimCountBits(a []uint64, n int) int32 {
	var count int32
	for _, x := range a[:n] {
		count += Ssc_SimCountBitsWord(x)
	}
	return count
}
func NewSimulation(p *gia.Gia_Man_t, words int) *Simulation {
	s := &Simulation{Graph: p}
	Ssc_GiaResetPiPattern(s, words)
	return s
}
func Ssc_GiaResetPiPattern(s *Simulation, words int) {
	if words < 1 {
		panic("ssc: simulation needs positive words")
	}
	s.Words = words
	s.Patterns = 0
	s.Inputs = make([][]uint64, len(s.Graph.InputLiterals()))
	for i := range s.Inputs {
		s.Inputs[i] = make([]uint64, words)
	}
}
func Ssc_GiaSavePiPattern(s *Simulation, pattern []bool) {
	if len(pattern) != len(s.Inputs) {
		panic("ssc: pattern size mismatch")
	}
	if s.Patterns == 64*s.Words {
		for i := range s.Inputs {
			s.Inputs[i] = append(s.Inputs[i], make([]uint64, s.Words)...)
		}
		s.Words *= 2
	}
	for i, b := range pattern {
		if b {
			s.Inputs[i][s.Patterns/64] |= 1 << uint(s.Patterns%64)
		}
	}
	s.Patterns++
}
func Ssc_GiaRandomPiPattern(s *Simulation, words int, pivot []bool) {
	Ssc_GiaResetPiPattern(s, words)
	if pivot != nil && len(pivot) != len(s.Inputs) {
		panic("ssc: pivot size mismatch")
	}
	// The C routine fills primary inputs, not register outputs. The latter
	// retain the zero words installed by Ssc_GiaResetPiPattern.
	for i := 0; i < int(gia.Gia_ManPiNum(s.Graph)); i++ {
		if pivot == nil {
			s.Inputs[i][0] = uint64(gia.Gia_ManRandom(0))<<32 | uint64(gia.Gia_ManRandom(0))<<2 | 2
		} else {
			s.Inputs[i][0] = uint64(gia.Gia_ManRandom(0))<<32 | uint64(gia.Gia_ManRandom(0))<<1
			if pivot[i] {
				s.Inputs[i][0] |= 1
			}
		}
		for w := 1; w < words; w++ {
			s.Inputs[i][w] = uint64(gia.Gia_ManRandom(0))<<32 | uint64(gia.Gia_ManRandom(0))
		}
	}
}
func Ssc_GiaResetSimInfo(s *Simulation) {
	s.Values = make([][]uint64, s.Graph.ObjectCount())
	for i := range s.Values {
		s.Values[i] = make([]uint64, s.Words)
	}
}
func Ssc_GiaSimRound(s *Simulation) {
	Ssc_GiaResetSimInfo(s)
	for i, lit := range s.Graph.InputLiterals() {
		copy(s.Values[lit>>1], s.Inputs[i])
	}
	done := make([]bool, len(s.Values))
	done[0] = true
	for _, lit := range s.Graph.InputLiterals() {
		done[lit>>1] = true
	}
	var node func(int32)
	var lit func(int32, int) uint64
	lit = func(l int32, w int) uint64 {
		node(l >> 1)
		v := s.Values[l>>1][w]
		if l&1 != 0 {
			v = ^v
		}
		return v
	}
	node = func(id int32) {
		if done[id] {
			return
		}
		a, b, c, k := s.Graph.Gate(id)
		if k < 0 {
			done[id] = true
			return
		}
		for w := 0; w < s.Words; w++ {
			x := lit(a, w)
			switch k {
			case 2:
				s.Values[id][w] = x & lit(b, w)
			case 3:
				s.Values[id][w] = x ^ lit(b, w)
			case 4:
				s.Values[id][w] = (x & lit(b, w)) | (^x & lit(c, w))
			case 5, 6:
				s.Values[id][w] = x
			default:
				panic("ssc: invalid simulation node")
			}
		}
		done[id] = true
	}
	for i := range s.Values {
		node(int32(i))
	}
}

// Care masks AND outputs. Sweeping uses a simulation view of inverted violation
// outputs; this helper itself retains the source all-outputs-true convention.
func Ssc_GiaGetCareMask(s *Simulation) []uint64 {
	mask := make([]uint64, s.Words)
	for i := range mask {
		mask[i] = ^uint64(0)
	}
	// Register inputs are COs but are not POs; the C mask iterates POs only.
	for _, o := range s.Graph.OutputLiterals()[:int(gia.Gia_ManPoNum(s.Graph))] {
		for w := range mask {
			x := s.Values[o>>1][w]
			if o&1 != 0 {
				x = ^x
			}
			mask[w] &= x
		}
	}
	return mask
}
func Ssc_GiaGetOneSim(s *Simulation) []bool {
	bit := Ssc_SimFindBit(Ssc_GiaGetCareMask(s), s.Words)
	if bit < 0 {
		return nil
	}
	p := make([]bool, len(s.Inputs))
	for i, lit := range s.Graph.InputLiterals() {
		p[i] = s.Values[lit>>1][bit/64]&(1<<uint(bit%64)) != 0
	}
	return p
}
func Ssc_GiaFindPivotSim(s *Simulation) []bool {
	Ssc_GiaRandomPiPattern(s, 1, nil)
	Ssc_GiaSimRound(s)
	return Ssc_GiaGetOneSim(s)
}
func Ssc_GiaCountCaresSim(s *Simulation) int32 {
	return Ssc_SimCountBits(Ssc_GiaGetCareMask(s), s.Words)
}
func Ssc_GiaEstimateCare(s *Simulation, words int) int32 {
	Ssc_GiaRandomPiPattern(s, words, nil)
	Ssc_GiaSimRound(s)
	return Ssc_GiaCountCaresSim(s)
}
func Ssc_GiaTransferPiPattern(dst, care *Simulation, pivot []bool) int32 {
	mask := Ssc_GiaGetCareMask(care)
	count := Ssc_SimCountBits(mask, care.Words)
	if count == 0 {
		return 0
	}
	if len(dst.Inputs) != len(care.Inputs) || len(pivot) != len(dst.Inputs) {
		panic("ssc: transfer interface mismatch")
	}
	Ssc_GiaResetPiPattern(dst, care.Words)
	for i, lit := range care.Graph.InputLiterals() {
		for w := range mask {
			if pivot[i] {
				dst.Inputs[i][w] = care.Values[lit>>1][w] | ^mask[w]
			} else {
				dst.Inputs[i][w] = care.Values[lit>>1][w] & mask[w]
			}
		}
	}
	return count
}
