package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestTemporalDriverChecksBMC(t *testing.T) {
	defer RegisterTemporalBMC(nil)
	g := temporalFixture()
	RegisterTemporalBMC(nil)
	if Saig_ManTempor(g, 2, 3, 4, 1, 0, 0, 0) != nil {
		t.Fatal("decomposition accepted without BMC")
	}
	called := false
	RegisterTemporalBMC(func(graph *aig.Graph, frames, timeout, conflicts, verbose int32, finished *int32) int32 {
		called = true
		if graph != g || frames != 2 || timeout != 3 || conflicts != 4 {
			t.Fatal("incorrect BMC arguments")
		}
		*finished = 1
		return -1
	})
	if Saig_ManTempor(g, 2, 3, 4, 1, 0, 0, 0) == nil || !called {
		t.Fatal("successful BMC did not permit decomposition")
	}
	RegisterTemporalBMC(func(_ *aig.Graph, _, _, _, _ int32, finished *int32) int32 { *finished = 0; return 0 })
	if Saig_ManTempor(g, 2, 0, 0, 1, 0, 0, 0) != nil {
		t.Fatal("counterexample ignored")
	}
	RegisterTemporalBMC(func(_ *aig.Graph, _, _, _, _ int32, finished *int32) int32 { *finished = 0; return -1 })
	if Saig_ManTempor(g, 2, 0, 0, 1, 0, 0, 0) != nil {
		t.Fatal("unfinished BMC accepted")
	}
}
func TestLastTransientBeforeLimit(t *testing.T) {
	v := vec.Vec_IntAlloc(4)
	for _, x := range []int32{0, 1, 0, 2} {
		vec.Vec_IntPush(v, x)
	}
	for _, test := range []struct{ limit, want int32 }{{0, -1}, {1, -1}, {3, 1}, {4, 3}, {10, 3}} {
		if got := Vec_IntLastNonZeroBeforeLimit(v, test.limit); got != test.want {
			t.Fatalf("limit %d got %d", test.limit, got)
		}
	}
}
