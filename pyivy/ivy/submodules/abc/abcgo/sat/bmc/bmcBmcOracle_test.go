package bmc

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

// oracleTruth uses bit (input + 2*state), independent of the AIG graph.
func oracleTruth(table, input, state int) int { return (table >> (input + 2*state)) & 1 }
func oracleAIGTruth(g *aig.Graph, input, state aig.Edge, table int) aig.Edge {
	out := g.Const0()
	for stateBit := 0; stateBit < 2; stateBit++ {
		for inputBit := 0; inputBit < 2; inputBit++ {
			if oracleTruth(table, inputBit, stateBit) == 0 {
				continue
			}
			a, b := input, state
			if inputBit == 0 {
				a = a.Not()
			}
			if stateBit == 0 {
				b = b.Not()
			}
			out = g.OR(out, g.AND(a, b))
		}
	}
	return out
}
func TestScalableBMCExhaustiveStateOracle(t *testing.T) {
	const bound = 3
	for transition := 0; transition < 16; transition++ {
		for output := 0; output < 16; output++ {
			g := aig.NewGraph()
			in, reg := g.NewInput(), g.NewInput()
			g.AddOutput(oracleAIGTruth(g, in, reg, output))
			g.AddOutput(oracleAIGTruth(g, in, reg, transition))
			aig.Aig_ManSetRegNum(g, 1)
			for start := 0; start < 2; start++ {
				for _, all := range []int32{0, 1} {
					reachable := [2]bool{true, false}
					earliest := -1
					for frame := 0; frame < bound; frame++ {
						next := [2]bool{}
						for state := 0; state < 2; state++ {
							if !reachable[state] {
								continue
							}
							for input := 0; input < 2; input++ {
								if frame >= start && earliest < 0 && oracleTruth(output, input, state) != 0 {
									earliest = frame
								}
								next[oracleTruth(transition, input, state)] = true
							}
						}
						reachable = next
					}
					params := Saig_ParBmc_t{nStart: int32(start), nFramesMax: bound, fSolveAll: all}
					result := Saig_ManBmcScalable(g, &params)
					if earliest < 0 {
						if result != 1 {
							t.Fatalf("trans=%x output=%x start=%d all=%d got=%d, want safe", transition, output, start, all, result)
						}
						continue
					}
					if result != 0 {
						t.Fatalf("trans=%x output=%x start=%d all=%d got=%d, want fail at %d", transition, output, start, all, result, earliest)
					}
					cex := g.SequentialCounterexample()
					if cex == nil || int(cex.IFrame) != earliest || cex.IPo != 0 {
						t.Fatalf("trans=%x output=%x start=%d all=%d cex=%+v want frame %d", transition, output, start, all, cex, earliest)
					}
					state := 0
					for frame := 0; frame <= earliest; frame++ {
						input := 0
						if util.Abc_InfoHasBit(cex.PData, int32(1+frame)) {
							input = 1
						}
						if frame == earliest && oracleTruth(output, input, state) == 0 {
							t.Fatalf("counterexample fails replay: trans=%x output=%x start=%d", transition, output, start)
						}
						state = oracleTruth(transition, input, state)
					}
				}
			}
		}
	}
}

func TestScalableBMCMultiOutputOracle(t *testing.T) {
	const bound = 3
	for transition := 0; transition < 16; transition++ {
		for out0 := 0; out0 < 16; out0++ {
			out1 := (out0*7 + transition*3) & 15
			g := aig.NewGraph()
			in, reg := g.NewInput(), g.NewInput()
			g.AddOutput(oracleAIGTruth(g, in, reg, out0))
			g.AddOutput(oracleAIGTruth(g, in, reg, out1))
			g.AddOutput(oracleAIGTruth(g, in, reg, transition))
			aig.Aig_ManSetRegNum(g, 1)
			for start := 0; start < 2; start++ {
				want := [2]int{-1, -1}
				reachable := [2]bool{true, false}
				for frame := 0; frame < bound; frame++ {
					next := [2]bool{}
					for state := 0; state < 2; state++ {
						if !reachable[state] {
							continue
						}
						for input := 0; input < 2; input++ {
							if frame >= start {
								for i, table := range []int{out0, out1} {
									if want[i] < 0 && oracleTruth(table, input, state) != 0 {
										want[i] = frame
									}
								}
							}
							next[oracleTruth(transition, input, state)] = true
						}
					}
					reachable = next
				}
				got := [2]int{-1, -1}
				params := Saig_ParBmc_t{nStart: int32(start), nFramesMax: bound, fSolveAll: 1}
				params.pFuncOnFail = func(output int32, cex *util.Abc_Cex_t) int32 {
					if output < 0 || output > 1 || cex == nil || got[output] >= 0 || cex.IPo != output {
						t.Fatalf("invalid witness output=%d cex=%+v", output, cex)
					}
					got[output] = int(cex.IFrame)
					state := 0
					for frame := 0; frame <= int(cex.IFrame); frame++ {
						input := 0
						if util.Abc_InfoHasBit(cex.PData, int32(1+frame)) {
							input = 1
						}
						if frame == int(cex.IFrame) && oracleTruth([]int{out0, out1}[output], input, state) == 0 {
							t.Fatalf("failed witness replay")
						}
						state = oracleTruth(transition, input, state)
					}
					return 0
				}
				result := Saig_ManBmcScalable(g, &params)
				expected := int32(1)
				count := int32(0)
				for _, frame := range want {
					if frame >= 0 {
						expected = 0
						count++
					}
				}
				if result != expected || got != want || params.nFailOuts != count {
					t.Fatalf("trans=%x out=%x,%x start=%d got=%v result=%d count=%d want=%v result=%d count=%d", transition, out0, out1, start, got, result, params.nFailOuts, want, expected, count)
				}
			}
		}
	}
}
