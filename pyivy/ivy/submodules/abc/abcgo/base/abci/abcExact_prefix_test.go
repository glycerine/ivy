package abci

import "testing"

func exactTruth2(fn func(int, int) bool) word {
	var out word
	for assignment := 0; assignment < 64; assignment++ {
		x0, x1 := assignment&1, (assignment>>1)&1
		if fn(x0, x1) {
			out |= word(1) << uint(assignment)
		}
	}
	return out
}

func TestExactTopDecompositionTruthFamilies(t *testing.T) {
	mask := ^word(0)
	cases := []struct {
		name  string
		truth word
		want  int32
	}{
		{"and-positive", exactTruth2(func(x, y int) bool { return x&y != 0 }), 1},
		{"and-negative", exactTruth2(func(x, y int) bool { return x == 0 && y != 0 }), 2},
		{"or-positive", exactTruth2(func(x, y int) bool { return x|y != 0 }), 3},
		{"or-negative", exactTruth2(func(x, y int) bool { return x == 0 || y != 0 }), 4},
		{"xor", exactTruth2(func(x, y int) bool { return x != y }), 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Abc_TtIsTopDecomposable(&tc.truth, &mask, 1, 0)
			if got != tc.want {
				t.Fatalf("got decomposition %d, want %d", got, tc.want)
			}
		})
	}
}

func TestExactStairDecompositionPreservesClassification(t *testing.T) {
	truth := exactTruth2(func(x, y int) bool { return x&y != 0 })
	vars := []int32{0}
	funcs := []int32{0}
	if got := Abc_TtIsStairDecomposable(&truth, 1, &vars[0], 1, &funcs[0]); got != 1 || funcs[0] != 1 {
		t.Fatalf("stair result=(%d,%d), want (1,1)", got, funcs[0])
	}
	badVars := []int32{8}
	if got := Abc_TtIsStairDecomposable(&truth, 1, &badVars[0], 1, &funcs[0]); got != 0 {
		t.Fatalf("invalid variable accepted: %d", got)
	}
}

func TestExactArrivalNormalization(t *testing.T) {
	times := []int32{7, 4, 9}
	var maximum int32
	delta := Abc_NormalizeArrivalTimes(&times[0], 3, &maximum)
	if delta != 4 || maximum != 6 || times[0] != 3 || times[1] != 0 || times[2] != 5 {
		t.Fatalf("normalized=%v delta=%d max=%d", times, delta, maximum)
	}
}
