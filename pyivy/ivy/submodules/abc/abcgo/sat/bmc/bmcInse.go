package bmc

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

// inseState packs ternary simulations into separate known-zero and known-one
// bitsets for each GIA object. A bit clear in both sets represents X.
type inseState struct {
	p               *gia.Gia_Man_t
	words           int
	zero, one       [][]uint64
	inputs, outputs []int32
	npi, npo        int
}

func newInseState(p *gia.Gia_Man_t, nwords int) *inseState {
	if p == nil || nwords <= 0 {
		panic("bmc: invalid initial-state simulation")
	}
	n := gia.Gia_ManObjNum(p)
	s := &inseState{p: p, words: nwords, zero: make([][]uint64, n), one: make([][]uint64, n), inputs: p.InputLiterals(), outputs: p.OutputLiterals(), npi: int(gia.Gia_ManPiNum(p)), npo: int(gia.Gia_ManPoNum(p))}
	for i := range s.zero {
		s.zero[i] = make([]uint64, nwords)
		s.one[i] = make([]uint64, nwords)
	}
	return s
}
func (s *inseState) init(values []int32) {
	if len(values) != len(s.inputs)-s.npi {
		panic("bmc: wrong initial state length")
	}
	for i, value := range values {
		id := gia.Gia_ObjId(s.p, gia.Gia_ManRi(s.p, int32(i)))
		for w := 0; w < s.words; w++ {
			s.zero[id][w], s.one[id][w] = 0, 0
			if value == 0 {
				s.zero[id][w] = ^uint64(0)
			} else if value == 1 {
				s.one[id][w] = ^uint64(0)
			}
		}
	}
}
func (s *inseState) fanin(lit int32, w int) (uint64, uint64) {
	a, b := s.zero[lit>>1][w], s.one[lit>>1][w]
	if lit&1 != 0 {
		return b, a
	}
	return a, b
}
func (s *inseState) simulate() {
	for id := int32(0); id < gia.Gia_ManObjNum(s.p); id++ {
		a, b, c, kind := s.p.Gate(id)
		for w := 0; w < s.words; w++ {
			switch kind {
			case 0:
				s.zero[id][w], s.one[id][w] = ^uint64(0), 0
			case 1:
				pi := -1
				for i, lit := range s.inputs {
					if lit>>1 == id {
						pi = i
						break
					}
				}
				if pi < 0 {
					panic("bmc: invalid input")
				}
				if pi < s.npi {
					s.zero[id][w] = gia.Gia_ManRandomW(0)
					s.one[id][w] = ^s.zero[id][w]
				} else {
					ri := gia.Gia_ObjId(s.p, gia.Gia_ManRi(s.p, int32(pi-s.npi)))
					s.zero[id][w], s.one[id][w] = s.zero[ri][w], s.one[ri][w]
				}
			case 2:
				a0, a1 := s.fanin(a, w)
				b0, b1 := s.fanin(b, w)
				s.zero[id][w], s.one[id][w] = a0|b0, a1&b1
			case 3:
				a0, a1 := s.fanin(a, w)
				b0, b1 := s.fanin(b, w)
				s.zero[id][w], s.one[id][w] = (a0&b0)|(a1&b1), (a0&b1)|(a1&b0)
			case 4:
				c0, c1 := s.fanin(a, w)
				t0, t1 := s.fanin(b, w)
				e0, e1 := s.fanin(c, w)
				s.zero[id][w], s.one[id][w] = (c1&t0)|(c0&e0)|(t0&e0), (c1&t1)|(c0&e1)|(t1&e1)
			case 5, 6:
				s.zero[id][w], s.one[id][w] = s.fanin(a, w)
			default:
				panic("bmc: invalid GIA object")
			}
		}
	}
}
func (s *inseState) best() (int, int32) {
	counts := make([]int32, 64*s.words)
	for i := 0; i < len(s.inputs)-s.npi; i++ {
		id := gia.Gia_ObjId(s.p, gia.Gia_ManRi(s.p, int32(i)))
		for w := 0; w < s.words; w++ {
			known := s.zero[id][w] | s.one[id][w]
			for bit := 0; bit < 64; bit++ {
				counts[64*w+bit] += int32((known >> bit) & 1)
			}
		}
	}
	best := 0
	for i := 1; i < len(counts); i++ {
		if counts[i] > counts[best] {
			best = i
		}
	}
	return best, int32(len(s.inputs)-s.npi) - counts[best]
}
func (s *inseState) extract(pattern int, init, inputs *vec.Vec_Int_t) {
	val := func(id int32) int32 {
		mask := uint64(1) << uint(pattern%64)
		if s.zero[id][pattern/64]&mask != 0 {
			return 0
		}
		if s.one[id][pattern/64]&mask != 0 {
			return 1
		}
		return 2
	}
	vec.Vec_IntClear(init)
	for i := 0; i < len(s.inputs)-s.npi; i++ {
		vec.Vec_IntPush(init, val(gia.Gia_ObjId(s.p, gia.Gia_ManRi(s.p, int32(i)))))
	}
	for i := 0; i < s.npi; i++ {
		vec.Vec_IntPush(inputs, val(s.inputs[i]>>1))
	}
}

// Gia_ManInseSimulate simulates the selected sequence from a concrete state;
// bit 2 marks registers whose final ternary value is known.
func Gia_ManInseSimulate(p *gia.Gia_Man_t, initial, inputs, expected *vec.Vec_Int_t) *vec.Vec_Int_t {
	npi, nreg := int(gia.Gia_ManPiNum(p)), int(gia.Gia_ManRegNum(p))
	init, stim, want := vec.Vec_IntArray(initial), vec.Vec_IntArray(inputs), vec.Vec_IntArray(expected)
	if len(init) != nreg || len(want) != nreg || npi == 0 && len(stim) != 0 || npi > 0 && len(stim)%npi != 0 {
		panic("bmc: invalid simulation dimensions")
	}
	state := make([]bool, nreg)
	for i, v := range init {
		if v != 0 && v != 1 {
			panic("bmc: non-binary initial state")
		}
		state[i] = v != 0
	}
	frames := 0
	if npi > 0 {
		frames = len(stim) / npi
	}
	for f := 0; f < frames; f++ {
		values := make([]bool, npi+nreg)
		for i := 0; i < npi; i++ {
			if stim[f*npi+i] != 0 && stim[f*npi+i] != 1 {
				panic("bmc: non-binary input")
			}
			values[i] = stim[f*npi+i] != 0
		}
		copy(values[npi:], state)
		outputs, err := p.Evaluate(values)
		if err != nil {
			panic(err)
		}
		copy(state, outputs[int(gia.Gia_ManPoNum(p)):])
	}
	result := vec.Vec_IntAlloc(nreg)
	for i, v := range state {
		x := int32(0)
		if v {
			x = 1
		}
		if want[i] != 2 {
			if want[i] != x {
				panic("bmc: ternary state mismatch")
			}
			x |= 4
		}
		vec.Vec_IntPush(result, x)
	}
	return result
}

func Gia_ManInsePerform(p *gia.Gia_Man_t, initial *vec.Vec_Int_t, nFrames, nWords, verbose int32) *vec.Vec_Int_t {
	if p == nil || nFrames < 0 || nWords <= 0 {
		panic("bmc: invalid initial-state search")
	}
	nreg, npi := int(gia.Gia_ManRegNum(p)), int(gia.Gia_ManPiNum(p))
	if initial == nil {
		initial = vec.Vec_IntStart(nreg)
	}
	start := time.Now()
	gia.Gia_ManRandomW(1)
	init := vec.Vec_IntAlloc(nreg)
	vec.Vec_IntFill(init, nreg, 2)
	inputs := vec.Vec_IntAlloc(npi * int(nFrames))
	state := newInseState(p, int(nWords))
	state.init(vec.Vec_IntArray(init))
	cost := int32(nreg)
	for f := int32(0); f < nFrames; f++ {
		state.simulate()
		pat, c := state.best()
		cost = c
		state.extract(pat, init, inputs)
		state.init(vec.Vec_IntArray(init))
		if verbose != 0 {
			fmt.Printf("Frame =%6d : Values =%6d (out of %6d)\n", f+1, cost, nreg)
		}
	}
	result := Gia_ManInseSimulate(p, initial, inputs, init)
	fmt.Printf("After %d frames, found a sequence to produce %d x-values (out of %d). Total runtime: %s\n", nFrames, cost, nreg, time.Since(start))
	return result
}
func Gia_ManInseTest(p *gia.Gia_Man_t, initial *vec.Vec_Int_t, nFrames, nWords, nTimeOut, fSim, verbose int32) *vec.Vec_Int_t {
	_ = initial
	_ = nTimeOut
	_ = fSim
	return Gia_ManInsePerform(p, vec.Vec_IntStart(int(gia.Gia_ManRegNum(p))), nFrames, nWords, verbose)
}
