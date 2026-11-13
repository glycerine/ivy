package int

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestFramesInterTruthTables(t *testing.T) {
	src := aig.NewGraph()
	x, lo := src.NewInput(), src.NewInput()
	src.AddOutput(lo)
	src.AddOutput(x) // constraint
	src.AddOutput(src.NewXor(x, lo))
	aig.Aig_ManSetRegNum(src, 1)
	aig.Aig_ManSetConstrNum(src, 1)
	for _, addRegs := range []int32{0, 1} {
		for _, twoFrames := range []int32{0, 1} {
			g := Inter_ManFramesInter(src, 2, addRegs, twoFrames)
			for mask := 0; mask < 8; mask++ {
				initial, x0, x1 := mask&1 != 0, mask&2 != 0, mask&4 != 0
				inputs := []bool{initial, x0, x1}
				if addRegs != 0 {
					initial = false
					inputs = []bool{x0, x1}
				}
				state1 := initial != x0
				last := state1
				if addRegs != 0 {
					last = state1 != x1
				} else if twoFrames != 0 {
					last = initial || state1
				}
				got, err := g.Evaluate(inputs)
				if err != nil || len(got) != 3 || got[0] != !x0 || got[1] != !x1 || got[2] != last {
					t.Fatalf("regs=%d two=%d inputs=%v got=%v want=[%v %v %v] err=%v", addRegs, twoFrames, inputs, got, !x0, !x1, last, err)
				}
			}
		}
	}
}
