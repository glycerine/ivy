package gia

import (
	"fmt"
	"testing"

	"github.com/glycerine/abcgo/aig/miniaig"
)

func TestMiniEquivalenceReductionAndMiterSequentialTraces(t *testing.T) {
	for signs := 0; signs < 4; signs++ {
		t.Run(fmt.Sprint(signs), func(t *testing.T) {
			p := miniaig.Mini_AigStart()
			pi := miniaig.Mini_AigCreatePi(p)
			r0, r1 := miniaig.Mini_AigCreatePi(p), miniaig.Mini_AigCreatePi(p)
			x := miniaig.Mini_AigAnd(p, pi^(signs&1), r0^((signs>>1)&1))
			y := miniaig.Mini_AigAnd(p, pi^(signs&1), r0^((signs>>1)&1))
			for _, lit := range []int{x, y ^ 1, r0, r1, pi, pi} {
				miniaig.Mini_AigCreatePo(p, lit)
			}
			miniaig.Mini_AigSetRegNum(p, 2)
			equivs := make([]int32, miniaig.Mini_AigNodeNum(p))
			for i := range equivs {
				equivs[i] = -1
			}
			equivs[y>>1], equivs[r1>>1] = int32(x), int32(r0)
			reduced, miter := Gia_MiniAigReduce(p, equivs), Gia_MiniAigMiter(p, equivs)
			if Gia_ManPiNum(reduced) != 1 || Gia_ManPoNum(reduced) != 4 || Gia_ManRegNum(reduced) != 1 {
				t.Fatalf("reduced interface: PI=%d PO=%d registers=%d", Gia_ManPiNum(reduced), Gia_ManPoNum(reduced), Gia_ManRegNum(reduced))
			}
			if Gia_ManPiNum(miter) != 1 || Gia_ManPoNum(miter) != 2 || Gia_ManRegNum(miter) != 2 {
				t.Fatalf("miter interface: PI=%d assertions=%d registers=%d", Gia_ManPiNum(miter), Gia_ManPoNum(miter), Gia_ManRegNum(miter))
			}
			for trace := 0; trace < 16; trace++ {
				state := false
				for frame := 0; frame < 4; frame++ {
					input := trace&(1<<uint(frame)) != 0
					xv := (input != (signs&1 != 0)) && (state != (signs&2 != 0))
					got, err := reduced.Evaluate([]bool{input, state})
					want := []bool{xv, !xv, state, state, input}
					if err != nil || len(got) != len(want) {
						t.Fatalf("trace %x frame %d: reduced=%v (%v)", trace, frame, got, err)
					}
					for i := range want {
						if got[i] != want[i] {
							t.Fatalf("trace %x frame %d output %d: reduced=%v, want %v", trace, frame, i, got, want)
						}
					}
					checks, err := miter.Evaluate([]bool{input, state, state})
					if err != nil || len(checks) != 4 || checks[0] || checks[1] || checks[2] != input || checks[3] != input {
						t.Fatalf("trace %x frame %d: assertions/next state=%v (%v)", trace, frame, checks, err)
					}
					state = input
				}
			}
		})
	}
}
