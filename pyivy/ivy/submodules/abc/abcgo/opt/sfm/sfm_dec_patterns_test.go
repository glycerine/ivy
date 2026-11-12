package sfm

import "testing"

func TestDecompositionPatternCosts(t *testing.T) {
	p := sfmDecPatterns{words: [2]int{1, 1}, divisors: 2, implications: [2][]int{{0, 2}, {1}}, decomposition: []int{0}}
	p.sets[0] = make([]uint64, 2*SFM_SIM_WORDS)
	p.sets[1] = make([]uint64, 2*SFM_SIM_WORDS)
	p.sets[0][0] = 0x0f
	p.sets[1][0] = 0xf0
	p.sets[0][SFM_SIM_WORDS] = 0x33
	p.sets[1][SFM_SIM_WORDS] = 0xcc
	mask := [2][]uint64{{0xff}, {0xff}}
	if c := p.variableCost(mask, 0); c != [2][2]int{{4, 4}, {4, 4}} {
		t.Fatalf("counts %v", c)
	}
	p.sets[0][0] = 0x03
	p.sets[1][0] = 0xfc
	if got := p.bestBalanced(mask); got != 0 {
		t.Fatalf("balanced %d", got)
	}
	if got := p.bestImplication(mask); got != 1 {
		t.Fatalf("implication %d", got)
	}
	if got := p.cost(0, 1, mask[1]); got != 2 {
		t.Fatalf("complement cost %d", got)
	}
}
