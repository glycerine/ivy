package acec

import (
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/util"
	"testing"
)

// Compiled C oracle: test_vectors/acec_timing_shapes.c.
func TestSourceTimingBoxShapes(t *testing.T) {
	cases := []struct{ pis, pos, boxes, ins, outs, ci, co, negative int32 }{
		{1, 2, 2, 2, 1, 3, 6, 0}, {3, 1, 2, 4, 3, 9, 9, 8}, {0, 0, 3, 5, 4, 12, 15, 15}, {2, 3, 0, 3, 2, 2, 3, 3},
	}
	for _, tc := range cases {
		m := Gia_ManGenerateTim(tc.pis, tc.pos, tc.boxes, tc.ins, tc.outs)
		if tim.Tim_ManCiNum(m) != int(tc.ci) || tim.Tim_ManCoNum(m) != int(tc.co) || tim.Tim_ManBoxNum(m) != int(tc.boxes) {
			t.Fatalf("shape %+v interface mismatch", tc)
		}
		for i := 0; i < int(tc.boxes); i++ {
			if tim.Tim_ManBoxInputFirst(m, i) != i*int(tc.ins) || tim.Tim_ManBoxInputNum(m, i) != int(tc.ins) || tim.Tim_ManBoxOutputFirst(m, i) != int(tc.pis)+i*int(tc.outs) || tim.Tim_ManBoxOutputNum(m, i) != int(tc.outs) {
				t.Fatalf("shape %+v box %d topology mismatch", tc, i)
			}
			table := tim.Tim_ManBoxDelayTable(m, i)
			if len(table) != 3+int(tc.ins*tc.outs) || table[1] != float32(tc.ins) || table[2] != float32(tc.outs) {
				t.Fatalf("shape %+v box %d table %v", tc, i, table)
			}
			for k, delay := range table[3:] {
				want := float32(1)
				if k == int(tc.negative) {
					want = -float32(util.ABC_INFINITY)
				}
				if delay != want {
					t.Fatalf("shape %+v box %d delay[%d]=%g want %g", tc, i, k, delay, want)
				}
			}
		}
	}
}
