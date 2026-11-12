package sfm

import (
	"reflect"
	"testing"
)

func TestSfmExtractAreaWindow(t *testing.T) {
	// The first AND belongs to the pivot's cone; its PI fanins are divisors.
	p := testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 0}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	pars := Sfm_Par_t{TfoLevMax: 2, TfiLevMax: 3, FanoutMax: 10, MffcMax: 3}
	w := sfmExtractAreaWindow(p, 3, pars)
	if !reflect.DeepEqual(w.roots, []int{3}) || !reflect.DeepEqual(w.mffc, []int{3, 2}) || !reflect.DeepEqual(w.mffcInputs, []int{0, 1}) {
		t.Fatalf("unexpected window: %+v", w)
	}
	if w.divisors != 2 || !reflect.DeepEqual(w.order, []int{0, 1, 2, 3}) {
		t.Fatalf("unexpected ordering: %+v", w)
	}
	window, cnfs := sfmDecWindowFromNetwork(p, w.order, w.roots, w.boundary, 3, w.divisors)
	truth, support, ok := sfmDecomposeWindow(window, cnfs, w.divisors, Sfm_Par_t{VarMax: 6, UseAndOr: true}, 0, 0, []uint64{0, 0}, []int{0, 1})
	if !ok || len(support) != 2 {
		t.Fatalf("decomposition failed: %x %v %v", truth, support, ok)
	}
	for a := 0; a < 4; a++ {
		index := 0
		for i, id := range support {
			if a&(1<<uint(id)) != 0 {
				index |= 1 << uint(i)
			}
		}
		if (truth[0]&(1<<uint(index)) != 0) != (a == 3) {
			t.Fatalf("incorrect truth for %d", a)
		}
	}
}

func TestSfmExtractTimingWindow(t *testing.T) {
	p := testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 0}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	gate := &Gate{Inputs: 2, Pins: []Pin{{Phase: 2, Rise: 2, Fall: 2}, {Phase: 2, Rise: 2, Fall: 2}}}
	p.Mapping[2], p.Mapping[3] = gate, gate
	timing := Sfm_TimStart(p, 3)
	pars := Sfm_Par_t{TfoLevMax: 2, TfiLevMax: 3, FanoutMax: 10, MffcMax: 3}
	w := sfmExtractWindow(p, 3, pars, timing)
	if !reflect.DeepEqual(w.mffc, []int{3}) || w.divisors != 2 || !reflect.DeepEqual(w.order, []int{0, 1, 2, 3}) {
		t.Fatalf("unexpected timing window: %+v", w)
	}
	if !reflect.DeepEqual(w.mffcInputs, []int{0, 1}) {
		t.Fatalf("noncritical inputs: %v", w.mffcInputs)
	}
	sfmDecWindowFromNetwork(p, w.order, w.roots, w.boundary, 3, w.divisors)
}
