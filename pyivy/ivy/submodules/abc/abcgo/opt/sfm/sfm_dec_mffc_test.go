package sfm

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/map/mio"
)

func TestSfmMappedMffcArea(t *testing.T) {
	p := testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 0}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	p.Mapping[2], p.Mapping[3] = &Gate{Area: 3}, &Gate{Area: 5}
	before := append([]int(nil), p.Fanouts[2]...)
	area, cone := sfmMappedMffcArea(p, 3, nil)
	if area != 8 || !reflect.DeepEqual(cone, []int{2, 3}) {
		t.Fatalf("area %d cone %v", area, cone)
	}
	area, cone = sfmMappedMffcArea(p, 3, []int{2})
	if area != 5 || !reflect.DeepEqual(cone, []int{3}) || !reflect.DeepEqual(before, p.Fanouts[2]) {
		t.Fatalf("cut: area %d cone %v", area, cone)
	}
}

func TestSfmFindAreaNodeMatch(t *testing.T) {
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
	_, _, _, handles := Sfm_LibPreprocess(lib)
	prepared := sfmLibPrepareCells(6, 0, 0, 0, 0, mio.Mio_CollectRootsNew2(lib, 6, nil, 0))
	defer Sfm_LibStop(prepared)
	p := testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 0}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	p.Mapping[2], p.Mapping[3] = &Gate{Area: 2000}, &Gate{Area: 2000}
	pars := Sfm_Par_t{TfoLevMax: 2, TfiLevMax: 3, FanoutMax: 10, MffcMax: 3, VarMax: 6, UseAndOr: true}
	gates, fanins, order, cone, ok := sfmFindAreaNodeMatch(p, 3, pars, prepared)
	if !ok || len(gates) != 1 || len(fanins) != 1 || len(fanins[0]) != 2 || !reflect.DeepEqual(cone, []int{2, 3}) || len(order) != 4 {
		t.Fatalf("match: %v %v %v %v %v", gates, fanins, order, cone, ok)
	}
	want := coreTable(t, p)
	if !sfmInsertAreaMatch(p, 3, gates, fanins, order, cone, handles) {
		t.Fatal("replacement rejected")
	}
	if !reflect.DeepEqual(coreTable(t, p), want) || !p.Deleted[2] || !p.Deleted[3] {
		t.Fatalf("replacement changed behavior or retained old cone: %v", p.Fanins)
	}
	// A two-gate replacement must distinguish its bottom gate from input 0.
	p = testNetwork(t, [][]int{nil, nil, {0, 1}, {2}}, 2, 1, []uint64{0, 0, 8, 0})
	want = coreTable(t, p)
	var and, buf int32 = -1, -1
	for i, gate := range handles {
		switch mio.Mio_GateReadName(gate) {
		case "and2":
			and = int32(i)
		case "buf":
			buf = int32(i)
		}
	}
	if !sfmInsertAreaMatch(p, 2, []int32{and, buf}, [][]int32{{0, 1}, {-1}}, []int{0, 1}, []int{2}, handles) || !reflect.DeepEqual(coreTable(t, p), want) {
		t.Fatal("two-gate replacement changed behavior")
	}
	p = testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 0}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	p.Mapping[2], p.Mapping[3] = &Gate{Area: 2000}, &Gate{Area: 2000}
	want = coreTable(t, p)
	pars.NodeOne = 3
	changed, err := sfmOptimizeMappedArea(p, pars, prepared, handles)
	if err != nil || changed != 1 || !reflect.DeepEqual(coreTable(t, p), want) {
		t.Fatalf("area driver: changed=%d err=%v", changed, err)
	}
	p = testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 0}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	p.Mapping[2], p.Mapping[3] = &Gate{Area: 2000}, &Gate{Area: 2000}
	pars.AreaRev = true
	pars.DecMax = 3
	changed, err = sfmOptimizeMappedArea(p, pars, prepared, handles)
	if err != nil || changed != 1 || !reflect.DeepEqual(coreTable(t, p), want) {
		t.Fatalf("reverse area driver: changed=%d err=%v", changed, err)
	}
	pars.DecMax = SFM_DEC_MAX + 1
	if _, err := sfmOptimizeMappedArea(p, pars, prepared, handles); err == nil {
		t.Fatal("unsupported option was silently ignored")
	}
	p = testNetwork(t, [][]int{nil, nil, {0, 1}, {2, 0}, {3}}, 2, 1, []uint64{0, 0, 8, 8, 0})
	slow := &Gate{Area: 2000, Inputs: 2, Pins: []Pin{{Phase: 2, Rise: 5000, Fall: 5000}, {Phase: 2, Rise: 5000, Fall: 5000}}}
	p.Mapping[2], p.Mapping[3] = slow, slow
	timing := Sfm_TimStart(p, 6000)
	delayLib := sfmLibPrepareCells(6, 0, 1, 0, 0, mio.Mio_CollectRootsNew2(lib, 6, nil, 0))
	defer Sfm_LibStop(delayLib)
	pars.DecMax = 1
	gates, fanins, order, cone, ok = sfmFindDelayNodeMatch(p, 3, pars, delayLib, timing)
	if !ok || !sfmInsertAreaMatch(p, 3, gates, fanins, order, cone, handles) {
		t.Fatal("timing match failed")
	}
	Sfm_TimTrace(timing)
	if timing.Delay >= 10000 || !reflect.DeepEqual(coreTable(t, p), want) {
		t.Fatalf("timing replacement: delay=%d", timing.Delay)
	}
}
