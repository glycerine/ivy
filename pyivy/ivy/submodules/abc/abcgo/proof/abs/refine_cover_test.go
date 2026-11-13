package abs

import (
	"math/bits"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestRefinementCoverAgainstSubsetOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(517032))
	for trial := 0; trial < 1000; trial++ {
		limit := rng.Intn(9)
		input := make([]int32, rng.Intn(40))
		p := vec.Vec_IntAlloc(len(input))
		for i := range input {
			input[i] = int32(rng.Intn(256))
			vec.Vec_IntPush(p, input[i])
		}
		set := map[int32]bool{}
		for _, candidate := range input {
			if bits.OnesCount32(uint32(candidate)) > limit {
				continue
			}
			minimal := true
			for _, other := range input {
				if other != candidate && other&candidate == other {
					minimal = false
					break
				}
			}
			if minimal {
				set[candidate] = true
			}
		}
		want := make([]int32, 0, len(set))
		for mask := range set {
			want = append(want, mask)
		}
		sort.Slice(want, func(i, j int) bool { return want[i] < want[j] })
		Rf2_ManProcessVector(p, int32(limit))
		got := append([]int32{}, vec.Vec_IntArray(p)...)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("trial=%d limit=%d input=%v got=%v want=%v", trial, limit, input, got, want)
		}
	}
}
