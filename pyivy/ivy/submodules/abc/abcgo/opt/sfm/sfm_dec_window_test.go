package sfm

import (
	"github.com/glycerine/abcgo/map/mio"
	"testing"
)

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

func TestSfmDecomposeWindowExhaustiveThreeInputs(t *testing.T) {
	successes := 0
	for function := 0; function < 256; function++ {
		cnfs := [][][]int{Sfm_TruthToCnf(uint64(function), nil, 3)}
		w := &sfmDecWindow{gates: []int{-1, -1, -1, 0}, fanins: [][]int{nil, nil, nil, {0, 1, 2}}, roots: []int{3}, pivot: 3}
		truth, support, ok := sfmDecomposeWindow(w, cnfs, 3, Sfm_Par_t{VarMax: 6, UseAndOr: true}, 0, 0, []uint64{0, 0, 0}, []int{0, 1, 2})
		if !ok {
			continue
		}
		successes++
		for assignment := 0; assignment < 8; assignment++ {
			index := 0
			for i, id := range support {
				if assignment&(1<<uint(id)) != 0 {
					index |= 1 << uint(i)
				}
			}
			got := truth[0]&(1<<uint(index)) != 0
			want := function&(1<<uint(assignment)) != 0
			if got != want {
				t.Fatalf("function %02x assignment %d truth %x support %v", function, assignment, truth, support)
			}
		}
	}
	if successes < 100 {
		t.Fatalf("only %d functions decomposed", successes)
	}
}

func TestSfmMatchWindowArea(t *testing.T) {
	const genlib = `GATE zero 0 O=CONST0;
GATE one 0 O=CONST1;
GATE buf 1 O=a; PIN * NONINV 1 999 1 0 2 0
GATE inv 1 O=!a; PIN * INV 1 999 1 0 2 0
GATE and2 2 O=a*b; PIN * NONINV 1 999 1 0 2 0
.end
`
	lib, err := mio.ReadLibrary("fixture", genlib, nil)
	if err != nil {
		t.Fatal(err)
	}
	Sfm_LibPreprocess(lib)
	prepared := sfmLibPrepareCells(6, 0, 0, 0, 0, mio.Mio_CollectRootsNew2(lib, 6, nil, 0))
	defer Sfm_LibStop(prepared)
	cnfs := [][][]int{Sfm_TruthToCnf(0x8888888888888888, nil, 2)}
	w := &sfmDecWindow{gates: []int{-1, -1, 0}, fanins: [][]int{nil, nil, {0, 1}}, roots: []int{2}, pivot: 2}
	gates, fanins, ok := sfmMatchAreaReplacement(w, cnfs, 2, Sfm_Par_t{VarMax: 6, UseAndOr: true}, prepared, 0, 0, []uint64{0, 0}, []int{0, 1}, 100000)
	if !ok || len(gates) != 1 || len(fanins) != 1 || len(fanins[0]) != 2 {
		t.Fatalf("gates %v fanins %v ok %v", gates, fanins, ok)
	}
}

func TestSfmNetworkWindowDecomposition(t *testing.T) {
	p := testNetwork(t, [][]int{nil, nil, {0, 1}, {2}}, 2, 1, []uint64{0, 0, 8, 0})
	window, cnfs := sfmDecWindowFromNetwork(p, []int{0, 1, 2}, []int{2}, nil, 2, 2)
	truth, support, ok := sfmDecomposeWindow(window, cnfs, 2, Sfm_Par_t{VarMax: 6, UseAndOr: true}, 0, 0, []uint64{0, 0}, []int{0, 1})
	if !ok || len(support) != 2 {
		t.Fatalf("truth %x support %v ok %v", truth, support, ok)
	}
	for a := 0; a < 4; a++ {
		index := 0
		for i, id := range support {
			if a&(1<<uint(id)) != 0 {
				index |= 1 << uint(i)
			}
		}
		if (truth[0]&(1<<uint(index)) != 0) != (a == 3) {
			t.Fatalf("assignment %d", a)
		}
	}
}

func TestSfmNetworkWindowInternalBoundary(t *testing.T) {
	// The AND node 2 is a cut input: its fanins are outside this window.
	p := testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 1}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	window, cnfs := sfmDecWindowFromNetwork(p, []int{2, 1, 3}, []int{3}, []int{2}, 3, 2)
	if window.gates[0] != -1 || len(window.fanins[0]) != 0 {
		t.Fatalf("internal boundary encoded as gate: %+v", window)
	}
	truth, support, ok := sfmDecomposeWindow(window, cnfs, 2, Sfm_Par_t{VarMax: 6, UseAndOr: true}, 0, 0, []uint64{0, 0}, []int{0, 1})
	if !ok || len(support) != 2 {
		t.Fatalf("truth %x support %v ok %v", truth, support, ok)
	}
	for a := 0; a < 4; a++ {
		index := 0
		for i, id := range support {
			if a&(1<<uint(id)) != 0 {
				index |= 1 << uint(i)
			}
		}
		if (truth[0]&(1<<uint(index)) != 0) != (a == 3) {
			t.Fatalf("assignment %d", a)
		}
	}
}
