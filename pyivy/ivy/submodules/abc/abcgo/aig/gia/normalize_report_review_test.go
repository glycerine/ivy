package gia

import (
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/tim"
)

func TestNormalizeReportsSequentialBoxInputShuffle(t *testing.T) {
	for regs := int32(0); regs < 2; regs++ {
		for hash := int32(0); hash < 2; hash++ {
			src := Gia_ManStart(8)
			a, r, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			root := Gia_ManAppendAnd(src, c, r)
			Gia_ManAppendCo(src, a)
			Gia_ManAppendCo(src, root)
			Gia_ManSetRegNum(src, regs)
			timing := tim.Tim_ManStart(3, 2)
			tim.Tim_ManCreateBox(timing, 0, 1, 2, 1, -1, 0)
			src.SetBoxTiming(timing, nil)
			var dst *Gia_Man_t
			output := hashReviewStdout(t, func() { dst = Gia_ManDupNormalize(src, hash) })
			wantOutput := ""
			if regs != 0 {
				wantOutput = "Warning: Shuffled CI order to be correct sequential AIG.\n"
			}
			if output != wantOutput {
				t.Errorf("regs%d/hash%d: output %q, want %q", regs, hash, output, wantOutput)
			}
			if Gia_ManIsNormalized(dst) == 0 || Gia_ManRegNum(dst) != regs {
				t.Fatal("normalized interface changed")
			}
			wantCopies := []uint32{2, 4, 6}
			if regs != 0 {
				wantCopies = []uint32{2, 6, 4}
			}
			gotCopies := []uint32{Gia_ManCi(src, 0).Value, Gia_ManCi(src, 1).Value, Gia_ManCi(src, 2).Value}
			if !slices.Equal(gotCopies, wantCopies) {
				t.Fatalf("CI copies %v, want %v", gotCopies, wantCopies)
			}
			for x := 0; x < 8; x++ {
				a, r, c := x&1 != 0, x&2 != 0, x&4 != 0
				inputs := []bool{a, r, c}
				if regs != 0 {
					inputs = []bool{a, c, r}
				}
				got, err := dst.Evaluate(inputs)
				if err != nil || !slices.Equal(got, []bool{a, c && r}) {
					t.Fatalf("input%d: got%v err%v", x, got, err)
				}
			}
			Gia_ManStop(dst)
			Gia_ManStop(src)
		}
	}
}
