package gia

import (
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

func cycleReviewCex(g *Gia_Man_t, trace int) *Abc_Cex_t {
	if trace == 0 {
		return nil
	}
	cex := util.Abc_CexAlloc(g.nRegs, Gia_ManPiNum(g), 4)
	cex.IFrame = 3
	cex.PData[0] = uint32(trace+1) * 0x9e3779b9
	return cex
}

// Evaluate each frame through the read-only Boolean evaluator, then apply
// simultaneous register updates. Also independently reproduce C's PI stream.
func cycleReviewExpected(t *testing.T, g *Gia_Man_t, cex *Abc_Cex_t, frames int32, initial int) ([]bool, uint32) {
	t.Helper()
	z, w := uint32(3716960521), uint32(2174103536)
	random := func() uint32 { z = 36969*(z&65535) + (z >> 16); w = 18000*(w&65535) + (w >> 16); return z<<16 + w }
	random() // Cycle resets the stream and discards its first generated word.
	values := make([]bool, g.nObjs)
	states := make([]bool, g.nRegs)
	for r := range states {
		states[r] = initial>>r&1 != 0
		values[Gia_ManRo(g, int32(r)).id] = states[r]
	}
	inputs := make([]bool, Gia_ManCiNum(g))
	for frame := int32(0); frame < frames; frame++ {
		for pi := int32(0); pi < Gia_ManPiNum(g); pi++ {
			if cex == nil {
				inputs[pi] = random()&1 != 0
			} else {
				bit := g.nRegs + frame*Gia_ManPiNum(g) + pi
				inputs[pi] = cex.PData[bit>>5]>>uint(bit&31)&1 != 0
			}
		}
		copy(inputs[Gia_ManPiNum(g):], states)
		all, outputs, err := g.evaluateObjects(inputs)
		if err != nil {
			t.Fatal(err)
		}
		values = all
		copy(states, outputs[Gia_ManPoNum(g):])
		for r, state := range states {
			values[Gia_ManRo(g, int32(r)).id] = state
		}
	}
	return values, random()
}

func cycleReviewSetInitial(g *Gia_Man_t, initial int) {
	Gia_ManCleanMark0(g)
	for r := int32(0); r < g.nRegs; r++ {
		Gia_ManRo(g, r).fMark0 = uint32(initial >> r & 1)
	}
	for _, obj := range g.pObjs {
		obj.fMark1 = uint32(obj.id & 1)
	}
}

func TestCycleAgainstTransitionTable(t *testing.T) {
	cases := 0
	for _, signs := range []int{0, 2, 5, 7} {
		for _, typed := range []bool{false, true} {
			for _, layout := range []bool{false, true} {
				for regs := int32(0); regs <= 5; regs++ {
					for initial := 0; initial < 1<<regs; initial++ {
						for trace := 0; trace < 5; trace++ {
							for frames := int32(0); frames <= 3; frames++ {
								g := partitionReviewGraph(signs, typed, layout)
								g.nRegs = regs // Reinterpret the fixture's existing terminal suffix.
								cycleReviewSetInitial(g, initial)
								cex := cycleReviewCex(g, trace)
								want, next := cycleReviewExpected(t, g, cex, frames, initial)
								Gia_ManCycle(g, cex, frames)
								if got := Gia_ManRandom(0); got != next {
									t.Fatalf("random stream %d, want%d", got, next)
								}
								for i, obj := range g.pObjs {
									if (obj.fMark0 != 0) != want[i] {
										t.Fatalf("signs%d typed%v dfs%v regs%d initial%d trace%d frames%d node%d mark%d want%v", signs, typed, layout, regs, initial, trace, frames, i, obj.fMark0, want[i])
									}
									if obj.fMark1 != uint32(obj.id&1) {
										t.Fatal("Cycle changed mark1")
									}
								}
								cases++
							}
						}
					}
				}
			}
		}
	}
	t.Logf("%d cycle runs match complete object marks and next random word", cases)
}

func TestCycledDuplicationAgainstTransitionTable(t *testing.T) {
	cases, rows := 0, 0
	for _, signs := range []int{0, 2, 5, 7} {
		for _, typed := range []bool{false, true} {
			for _, layout := range []bool{false, true} {
				for regs := int32(0); regs <= 5; regs++ {
					for trace := 0; trace < 5; trace++ {
						for frames := int32(0); frames <= 3; frames++ {
							src := partitionReviewGraph(signs, typed, layout)
							src.nRegs = regs // Reinterpret the fixture's existing terminal suffix.
							if err := src.SetAigerProperties(1, 1); err != nil {
								t.Fatal(err)
							}
							cex := cycleReviewCex(src, trace)
							want, next := cycleReviewExpected(t, src, cex, frames, 0)
							initial := make([]int32, (regs+31)/32)
							for r := int32(0); r < regs; r++ {
								if want[Gia_ManRo(src, r).id] {
									initial[r>>5] |= int32(uint32(1) << uint(r&31))
								}
							}
							// DupCycled starts from zero and cleans scratch even if the caller's marks
							// were dirty. The CEX register prefix does not replace that zero state.
							for _, obj := range src.pObjs {
								obj.fMark0 = 1
								obj.fMark1 = uint32(obj.id & 1)
							}
							out := Gia_ManDupCycled(src, cex, frames)
							if got := Gia_ManRandom(0); got != next {
								t.Fatalf("random stream %d, want%d", got, next)
							}
							for _, obj := range src.pObjs {
								if obj.fMark0 != 0 || obj.fMark1 != uint32(obj.id&1) {
									t.Fatal("DupCycled scratch cleanup changed")
								}
							}
							if out.aigerBadOutputs != 1 || out.nConstrs != 1 || out.nRegs != regs {
								t.Fatal("cycled output roles changed")
							}
							for assignment := 0; assignment < 32; assignment++ {
								inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0, assignment&16 != 0}
								checkQuantificationOutputs(t, out, inputs, stateFlipReviewExpected(t, src, inputs, initial))
								rows++
							}
							cases++
						}
					}
				}
			}
		}
	}
	t.Logf("%d cycled copies, %d input/state rows", cases, rows)
}

// Keep the unpacked PI frames as the oracle so an incorrect packed-bit offset
// in Cycle cannot be repeated by the reference evaluator.
func TestCycleCounterexampleWordBoundaries(t *testing.T) {
	for _, regs := range []int{0, 1, 31, 32, 33} {
		for _, piCount := range []int{1, 31, 32, 33, 35} {
			for _, layout := range []bool{false, true} {
				for frames := 0; frames <= 5; frames++ {
					g := Gia_ManStart(1)
					pis, states := make([]int32, piCount), make([]int32, regs)
					for i := range pis {
						pis[i] = Gia_ManAppendCi(g)
					}
					for i := range states {
						states[i] = Gia_ManAppendCi(g)
					}
					next := make([]int32, regs)
					for i := range states {
						next[i] = Gia_ManAppendAnd(g, pis[i%piCount]^int32(i&1), states[(i+1)%regs]^int32((i>>1)&1))
					}
					Gia_ManAppendCo(g, pis[piCount-1]^1)
					for _, lit := range next {
						Gia_ManAppendCo(g, lit)
					}
					Gia_ManSetRegNum(g, int32(regs))
					if layout {
						g = Gia_ManDupOrderDfsReverse(g)
					}
					cex := util.Abc_CexAlloc(int32(regs), int32(piCount), 6)
					cex.IFrame = 5
					setBit := func(bit int) { cex.PData[bit>>5] |= uint32(1) << uint(bit&31) }
					for r := 0; r < regs; r++ {
						setBit(r)
					} // This prefix must not set Cycle's starting state.
					frameInputs := make([][]bool, 6)
					for frame := range frameInputs {
						frameInputs[frame] = make([]bool, piCount)
						for pi := range frameInputs[frame] {
							frameInputs[frame][pi] = (frame*13+pi*7+frame*pi)%11 < 5
							if frameInputs[frame][pi] {
								setBit(regs + frame*piCount + pi)
							}
						}
					}
					expected := make([]bool, g.nObjs)
					state := make([]bool, regs)
					for frame := 0; frame < frames; frame++ {
						inputs := append(append([]bool(nil), frameInputs[frame]...), state...)
						values, outputs, err := g.evaluateObjects(inputs)
						if err != nil {
							t.Fatal(err)
						}
						expected = values
						copy(state, outputs[1:])
						for r, value := range state {
							expected[Gia_ManRo(g, int32(r)).id] = value
						}
					}
					cycleReviewSetInitial(g, 0)
					Gia_ManCycle(g, cex, int32(frames))
					for id, obj := range g.pObjs {
						if (obj.fMark0 != 0) != expected[id] {
							t.Fatalf("regs%d pis%d dfs%v frames%d node%d", regs, piCount, layout, frames, id)
						}
						if obj.fMark1 != uint32(id&1) {
							t.Fatal("Cycle changed mark1")
						}
					}
				}
			}
		}
	}
}
