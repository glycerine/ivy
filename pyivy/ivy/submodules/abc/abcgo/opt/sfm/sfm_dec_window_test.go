package sfm

import "testing"

func TestSfmDecomposeWindow(t *testing.T) {
	// Two free divisors feed an AND pivot, then a buffer root.
	cnfs := [][][]int{Sfm_TruthToCnf(0x8888888888888888, nil, 2), Sfm_TruthToCnf(0xaaaaaaaaaaaaaaaa, nil, 1)}
	w := &sfmDecWindow{gates: []int{-1, -1, 0, 1}, fanins: [][]int{nil, nil, {0, 1}, {2}}, roots: []int{3}, pivot: 2}
	truth, support, ok := sfmDecomposeWindow(w, cnfs, 2, Sfm_Par_t{VarMax: 6, UseAndOr: true}, 0, 0, []uint64{0, 0}, []int{0, 1})
	if !ok {
		t.Fatal("no decomposition")
	}
	for assignment := 0; assignment < 4; assignment++ {
		index := 0
		for i, id := range support {
			if assignment&(1<<uint(id)) != 0 {
				index |= 1 << uint(i)
			}
		}
		got := truth[0]&(1<<uint(index)) != 0
		if got != (assignment == 3) {
			t.Fatalf("assignment %d truth %x support %v", assignment, truth, support)
		}
	}
}
