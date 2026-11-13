package fraig

import "testing"

// fraigFeed.c skips non-sparse dynamic classes only when they exceed 20
// nodes; a full 20-node class contributes every pair when primary bits agree.
func TestFeedbackCoveringClassSizeBoundary(t *testing.T) {
	for _, size := range []int{19, 20, 21} {
		p := Fraig_ManCreate(&Fraig_Params_t{NPatsRand: 128, NPatsDyna: 128})
		p.iWordStart = 1
		p.pSimsReal[0] = 1
		nodes := make([]*Fraig_Node_t, size)
		for i := range nodes {
			nodes[i] = &Fraig_Node_t{puSimD: []uint32{uint32(i) << 1}}
			if i > 0 {
				nodes[i-1].pNextD = nodes[i]
			}
		}
		p.pTableF.pBins[0] = nodes[0]
		columns := Fraig_FeedBackCoveringStart(p)
		want := size * (size - 1) / 2
		if size > 20 {
			want = 0
		}
		if len(columns) != want {
			t.Fatalf("class %d: %d columns want %d", size, len(columns), want)
		}
		for _, column := range columns {
			if len(column) != 1 || column[0]&1 != 0 {
				t.Fatalf("class %d invalid masked column %v", size, column)
			}
		}
		p.pTableF.pBins[0] = nil
		Fraig_ManFree(p)
	}
}
