package vec

import (
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

// TestVecWrdStartRandomDrawsAbcRandomW: vecWrd.h fills the vector with
// Abc_RandomW(0), ABC's own generator, so simulation patterns (and
// &sim_gen and friends) follow ABC's stream.
func TestVecWrdStartRandomDrawsAbcRandomW(t *testing.T) {
	util.Abc_Random(1)
	want := make([]uint64, 5)
	for i := range want {
		want[i] = util.Abc_RandomW(0)
	}
	util.Abc_Random(1)
	got := Vec_WrdStartRandom(5)
	for i := range want {
		if Vec_WrdEntry(got, i) != want[i] {
			t.Fatalf("word %d: %#x, want Abc_RandomW %#x", i, Vec_WrdEntry(got, i), want[i])
		}
	}
}
