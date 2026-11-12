package sfm

import "testing"

func TestSfmDecomposeAreaNode(t *testing.T) {
	p := testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 0}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	pars := Sfm_Par_t{TfoLevMax: 2, TfiLevMax: 3, FanoutMax: 10, MffcMax: 3, VarMax: 6, UseAndOr: true, UseSim: true}
	truth, support, order, ok := sfmDecomposeAreaNode(p, 3, pars)
	if !ok || len(support) != 2 {
		t.Fatalf("decomposition: %x %v %v", truth, support, ok)
	}
	for a := 0; a < 4; a++ {
		index := 0
		for i, id := range support {
			if a&(1<<uint(order[id])) != 0 {
				index |= 1 << uint(i)
			}
		}
		if (truth[0]&(1<<uint(index)) != 0) != (a == 3) {
			t.Fatalf("assignment %d", a)
		}
	}
	pars.WinSizeMax = 1
	if _, _, _, ok := sfmDecomposeAreaNode(p, 3, pars); ok {
		t.Fatal("oversized window accepted")
	}
}
