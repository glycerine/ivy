package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
	"testing"
)

func TestInvariantVerifyBaseAndInduction(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		updateFromInput, positive bool
		want                      int32
	}{
		{"held-zero", false, false, 1},
		{"fails-base", false, true, 0},
		{"fails-induction", true, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			graph := aig.NewGraph()
			input, latch := graph.NewInput(), graph.NewInput()
			update := latch
			if tc.updateFromInput {
				update = input
			}
			graph.AddOutput(update)
			aig.Aig_ManSetRegNum(graph, 1)
			data := cnf.Cnf_DeriveSimple(graph, int32(graph.OutputCount()))
			literal := data.VarNums[latch.NodeID()] * 2
			if !tc.positive {
				literal ^= 1
			}
			ends, lits := vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(1)
			vec.Vec_IntPush(ends, 1)
			vec.Vec_IntPush(lits, literal)
			if got := Fra_InvariantVerify(graph, 1, ends, lits); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
			if vec.Vec_IntEntry(lits, 0) != literal {
				t.Fatal("invariant literals modified")
			}
		})
	}
}

func TestInvariantVerifyUnsupportedDepthReturnsSourceSuccess(t *testing.T) {
	graph := aig.NewGraph()
	ends, literals := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	for _, depth := range []int32{0, 2, 10} {
		if got := Fra_InvariantVerify(graph, depth, ends, literals); got != 1 {
			t.Fatalf("depth %d returned %d, C returns success without checking", depth, got)
		}
	}
}

func TestInvariantVerifyRejectsMalformedClauses(t *testing.T) {
	graph := aig.NewGraph()
	input, latch := graph.NewInput(), graph.NewInput()
	graph.AddOutput(input)
	aig.Aig_ManSetRegNum(graph, 1)
	_ = latch
	for _, tc := range []struct {
		name       string
		ends, lits []int32
	}{
		{"missing offsets", nil, []int32{0}},
		{"offset beyond literals", []int32{2}, []int32{0}},
		{"decreasing offsets", []int32{1, 0}, []int32{0}},
		{"negative literal", []int32{1}, []int32{-1}},
		{"out of range literal", []int32{1}, []int32{1 << 30}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ends, lits := vec.Vec_IntAlloc(len(tc.ends)), vec.Vec_IntAlloc(len(tc.lits))
			for _, x := range tc.ends {
				vec.Vec_IntPush(ends, x)
			}
			for _, x := range tc.lits {
				vec.Vec_IntPush(lits, x)
			}
			if Fra_InvariantVerify(graph, 1, ends, lits) != 0 {
				t.Fatal("malformed invariant reported verified")
			}
		})
	}
}
