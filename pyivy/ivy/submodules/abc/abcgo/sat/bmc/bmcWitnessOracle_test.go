package bmc

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

// witnessTruthGraph builds an AIG from a complete truth table, while the
// reachability and counterexample oracles below evaluate the table directly.
func witnessTruthGraph(g *aig.Graph, vars []aig.Edge, table uint16) aig.Edge {
	out := g.Const0()
	for assignment := 0; assignment < 16; assignment++ {
		if table&(1<<assignment) == 0 {
			continue
		}
		term := g.Const1()
		for j, v := range vars {
			if assignment&(1<<j) == 0 {
				v = v.Not()
			}
			term = g.AND(term, v)
		}
		out = g.OR(out, term)
	}
	return out
}

func TestBMCWitnessTwoInputsTwoRegisters(t *testing.T) {
	const bound = 4
	rng := rand.New(rand.NewSource(20261113))
	for trial := 0; trial < 96; trial++ {
		// Also exercise constant and complementary functions.
		tables := [4]uint16{uint16(rng.Uint32()), uint16(rng.Uint32()), uint16(rng.Uint32()), uint16(rng.Uint32())}
		if trial == 0 {
			tables = [4]uint16{0, 0xffff, 0xaaaa, 0xcccc}
		}
		if trial == 1 {
			tables = [4]uint16{0x8000, 0x0001, 0xaaaa, 0xcccc}
		}
		g := aig.NewGraph()
		vars := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for _, table := range tables {
			g.AddOutput(witnessTruthGraph(g, vars, table))
		}
		aig.Aig_ManSetRegNum(g, 2)
		for start := 0; start < 2; start++ {
			want := [2]int{-1, -1}
			reachable := [4]bool{true, false, false, false}
			for frame := 0; frame < bound; frame++ {
				next := [4]bool{}
				for state := 0; state < 4; state++ {
					if !reachable[state] {
						continue
					}
					for input := 0; input < 4; input++ {
						index := input | (state << 2)
						if frame >= start {
							for out := 0; out < 2; out++ {
								if want[out] < 0 && (tables[out]>>index)&1 != 0 {
									want[out] = frame
								}
							}
						}
						ns := int((tables[2]>>index)&1) | int((tables[3]>>index)&1)<<1
						next[ns] = true
					}
				}
				reachable = next
			}
			got := [2]int{-1, -1}
			params := Saig_ParBmc_t{nFramesMax: bound, nStart: int32(start), fSolveAll: 1, fStoreCex: 1}
			params.pFuncOnFail = func(out int32, cex *util.Abc_Cex_t) int32 {
				if out < 0 || out > 1 || cex == nil || cex.IPo != out || got[out] >= 0 || cex.NRegs != 2 || cex.NPis != 2 {
					t.Fatalf("trial %d: invalid witness output=%d cex=%+v", trial, out, cex)
				}
				got[out] = int(cex.IFrame)
				state := 0
				for r := 0; r < 2; r++ {
					if util.Abc_InfoHasBit(cex.PData, int32(r)) {
						state |= 1 << r
					}
				}
				if state != 0 {
					t.Fatalf("trial %d: nonzero initial state %d", trial, state)
				}
				for frame := 0; frame <= int(cex.IFrame); frame++ {
					input := 0
					for pi := 0; pi < 2; pi++ {
						if util.Abc_InfoHasBit(cex.PData, int32(2+2*frame+pi)) {
							input |= 1 << pi
						}
					}
					index := input | (state << 2)
					if frame == int(cex.IFrame) && (tables[out]>>index)&1 == 0 {
						t.Fatalf("trial %d output %d frame %d witness does not reproduce failure", trial, out, frame)
					}
					state = int((tables[2]>>index)&1) | int((tables[3]>>index)&1)<<1
				}
				return 0
			}
			status := Saig_ManBmcScalable(g, &params)
			for out := int32(0); out < 2; out++ {
				if cex := bmcStoredCex(g, out); cex != nil {
					got[out] = int(cex.IFrame)
					state := 0
					for r := 0; r < 2; r++ {
						if util.Abc_InfoHasBit(cex.PData, int32(r)) {
							state |= 1 << r
						}
					}
					if state != 0 {
						t.Fatalf("trial %d secondary witness has nonzero initial state %d", trial, state)
					}
					for frame := 0; frame <= int(cex.IFrame); frame++ {
						input := 0
						for pi := 0; pi < 2; pi++ {
							if util.Abc_InfoHasBit(cex.PData, int32(2+2*frame+pi)) {
								input |= 1 << pi
							}
						}
						index := input | (state << 2)
						if frame == int(cex.IFrame) && (tables[out]>>index)&1 == 0 {
							t.Fatalf("trial %d secondary output %d witness does not replay", trial, out)
						}
						state = int((tables[2]>>index)&1) | int((tables[3]>>index)&1)<<1
					}
				}
			}
			expected := int32(-1)
			count := int32(0)
			for _, frame := range want {
				if frame >= 0 {
					expected = 0
					count++
				}
			}
			if status != expected || got != want || params.nFailOuts != count {
				t.Fatalf("trial=%d start=%d tables=%04x got frames=%v status=%d count=%d want frames=%v status=%d count=%d", trial, start, tables, got, status, params.nFailOuts, want, expected, count)
			}
		}
	}
}
