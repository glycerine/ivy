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

func TestDecompositionSimulationPatterns(t *testing.T) {
	p := sfmDecPatterns{}
	values := []uint64{0b11010110, 0b01010101}
	p.setupSimulation(0xff, 0b11110000, values, true)
	if p.counts != [2]int{4, 4} || p.words != [2]int{1, 1} {
		t.Fatalf("counts %v words %v", p.counts, p.words)
	}
	if p.divisor(0, 0)[0] != 0xd || p.divisor(1, 0)[0] != 6 || p.divisor(0, 1)[0] != 5 {
		t.Fatalf("sets %x %x %x", p.divisor(0, 0)[0], p.divisor(1, 0)[0], p.divisor(0, 1)[0])
	}
	out := []uint64{0xffffffffffffffff, 0xffffffffffffffff}
	p.setdownSimulation(out, true)
	if out[0]&15 != 0xd || (out[0]>>32)&15 != 6 || out[1]&15 != 5 {
		t.Fatalf("setdown %x", out)
	}
	p.setupSimulation(0, ^uint64(0), values, true)
	if p.words != [2]int{} || p.counts != [2]int{} {
		t.Fatal("empty care should reset patterns")
	}
}
