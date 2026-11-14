package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bool/dec"
	"github.com/glycerine/abcgo/misc/vec"
)

func resubTestNetwork() *abc.Network {
	return abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 0)
}

func TestAbcManResubSimulateAndNormalize(t *testing.T) {
	n := resubTestNetwork()
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	and := abc.Abc_NtkCreateNode(n)
	abc.Abc_ObjAddFanin(and, abc.Abc_ObjNot(a))
	abc.Abc_ObjAddFanin(and, abc.Abc_ObjNot(b))
	vDivs := vec.Vec_PtrAlloc(3)
	vec.Vec_PtrPush(vDivs, a)
	vec.Vec_PtrPush(vDivs, b)
	vec.Vec_PtrPush(vDivs, and)
	leaf0, leaf1, internal := []uint32{0x97}, []uint32{0x3d}, []uint32{0}
	vSims := vec.Vec_PtrAlloc(3)
	vec.Vec_PtrPush(vSims, leaf0)
	vec.Vec_PtrPush(vSims, leaf1)
	vec.Vec_PtrPush(vSims, internal)

	Abc_ManResubSimulate(vDivs, 2, vSims, 2, 1)
	want := ^uint32(0x97) & ^uint32(0x3d)
	if got := internal[0]; got != want {
		t.Fatalf("normalized simulation = %08x, want %08x", got, want)
	}
	if and.Phase {
		t.Fatal("simulation phase bit should be zero for this complemented AND")
	}
	if leaf0[0] != ^uint32(0x97) || leaf1[0] != ^uint32(0x3d) {
		t.Fatalf("leaf simulations were not normalized in-place: %08x %08x", leaf0[0], leaf1[0])
	}
	data, ok := abc.Abc_ObjData(and).([]uint32)
	if !ok || len(data) == 0 || &data[0] != &internal[0] {
		t.Fatal("simulation slice was not associated with the object scratch field")
	}
}

func TestAbcManResubDivsSUsesCareSetAndComplementedCandidate(t *testing.T) {
	n := resubTestNetwork()
	objects := make([]*abc.Abc_Obj_t, 4)
	for i := range objects {
		objects[i] = abc.Abc_NtkCreatePi(n)
	}
	root := &Abc_ManRes_t{pRoot: abc.Abc_NtkCreatePi(n), nDivs: 4, nWords: 1, pCareSet: []uint32{0xf}}
	root.pRoot.Data = []uint32{0x3}
	root.vDivs = vec.Vec_PtrAlloc(4)
	root.vDivs1UP, root.vDivs1UN, root.vDivs1B = vec.Vec_PtrAlloc(4), vec.Vec_PtrAlloc(4), vec.Vec_PtrAlloc(4)
	for i, words := range [][]uint32{{0x3}, {0xc}, {0x7}, {0x5}} {
		objects[i].Data = words
		vec.Vec_PtrPush(root.vDivs, objects[i])
	}
	Abc_ManResubDivsS(root, 9)
	if got := vec.Vec_PtrSize(root.vDivs1UP); got != 2 {
		t.Fatalf("positive unate count = %d, want 2", got)
	}
	if got := vec.Vec_PtrEntry(root.vDivs1UP, 1); abc.Abc_ObjIsComplement(got) == 0 {
		t.Fatalf("second positive candidate = %#v, want complemented divisor", got)
	}
	if got := vec.Vec_PtrEntry(root.vDivs1UN, 0); got != objects[2] {
		t.Fatalf("negative unate candidate = %T %v, want object %p", got, got, objects[2])
	}
	if got := vec.Vec_PtrEntry(root.vDivs1B, 0); got != objects[3] {
		t.Fatalf("binate candidate = %T %v, want object %p", got, got, objects[3])
	}
}

func TestResubDivisorWordPredicatesCoverPolarityCases(t *testing.T) {
	a, b, c, d := uint32(0x01234567), uint32(0x89abcdef), uint32(0x55aa55aa), uint32(0xdeadbeef)
	care := []uint32{^uint32(0)}
	root := []uint32{0}
	for flags := 0; flags < 16; flags++ {
		va, vb, vc, vd := flags&8 != 0, flags&4 != 0, flags&2 != 0, flags&1 != 0
		got := resubSourceDivs3Word(a, b, c, d, va, vb, vc, vd)
		root[0] = got
		if !resubDivs3Equal([]uint32{a}, []uint32{b}, []uint32{c}, []uint32{d}, resubPolarityEdge(va), resubPolarityEdge(vb), resubPolarityEdge(vc), resubPolarityEdge(vd), root, care, 1) {
			t.Fatalf("four-input divisor polarity %04b did not match its source truth expression", flags)
		}
		root[0] = got ^ 1
		if resubDivs3Equal([]uint32{a}, []uint32{b}, []uint32{c}, []uint32{d}, resubPolarityEdge(va), resubPolarityEdge(vb), resubPolarityEdge(vc), resubPolarityEdge(vd), root, care, 1) {
			t.Fatalf("four-input divisor polarity %04b accepted a one-bit mismatch", flags)
		}
	}
}

func TestResubTwoPlusOnePredicatesCoverPolarityAndCare(t *testing.T) {
	a, b, c := uint32(0x12345678), uint32(0x89abcdef), uint32(0x0f0ff0f0)
	care := []uint32{0x00ffffff}
	root := []uint32{0}
	for flags := 0; flags < 8; flags++ {
		f0, f1, f2 := flags&4 != 0, flags&2 != 0, flags&1 != 0
		for _, positive := range []bool{true, false} {
			got := resubSourceDivs2Word(a, b, c, f0, f1, f2, positive)
			root[0] = got
			if !resubDivs2Equal([]uint32{a}, []uint32{b}, []uint32{c}, resubPolarityEdge(f0), resubPolarityEdge(f1), resubPolarityEdge(f2), root, care, 1, positive) {
				t.Fatalf("two-plus-one divisor flags %03b positive=%t did not match source expression", flags, positive)
			}
			root[0] = got ^ 1
			if resubDivs2Equal([]uint32{a}, []uint32{b}, []uint32{c}, resubPolarityEdge(f0), resubPolarityEdge(f1), resubPolarityEdge(f2), root, care, 1, positive) {
				t.Fatalf("two-plus-one divisor flags %03b positive=%t accepted a cared mismatch", flags, positive)
			}
			root[0] = got ^ 0x01000000
			if !resubDivs2Equal([]uint32{a}, []uint32{b}, []uint32{c}, resubPolarityEdge(f0), resubPolarityEdge(f1), resubPolarityEdge(f2), root, care, 1, positive) {
				t.Fatalf("two-plus-one divisor flags %03b positive=%t did not mask don't-care bits", flags, positive)
			}
		}
	}
}

func resubSourceDivs2Word(a, b, c uint32, f0, f1, f2, positive bool) uint32 {
	b0, c0 := b, c
	if f0 {
		a = ^a
	}
	if positive {
		if f1 && f2 {
			return a | (b0 | c0)
		}
		if f1 {
			return a | (^b0 & c0)
		}
		if f2 {
			return a | (b0 & ^c0)
		}
		return a | (b0 & c0)
	}
	if f1 && f2 {
		return a & (b0 | c0)
	}
	if f1 {
		return a & (^b0 & c0)
	}
	if f2 {
		return a & (b0 & ^c0)
	}
	return a & (b0 & c0)
}

func resubPolarityEdge(complement bool) any {
	n := &abc.Object{}
	if complement {
		return abc.Edge{Object: n, Complement: true}
	}
	return n
}

func resubSourceDivs3Word(a, b, c, d uint32, f0, f1, f2, f3 bool) uint32 {
	if f0 {
		a = ^a
	}
	if f1 {
		b = ^b
	}
	if f0 && f1 {
		left := ^a | ^b
		if f2 {
			c = ^c
		}
		if f3 {
			d = ^d
		}
		right := c & d
		if f2 && f3 {
			right = ^c | ^d
		}
		return left | right
	} else {
		left := a & b
		if f2 {
			c = ^c
		}
		if f3 {
			d = ^d
		}
		right := c & d
		if f2 && f3 {
			right = ^c | ^d
		}
		return left | right
	}
}

func TestAbcManResubDivsDUsesCareSet(t *testing.T) {
	n := resubTestNetwork()
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	a.Data, b.Data = []uint32{0x3}, []uint32{0x5}
	root := &Abc_ManRes_t{
		pRoot: abc.Abc_NtkCreatePi(n), nWords: 1, pCareSet: []uint32{0xf},
		vDivs1B: vec.Vec_PtrAlloc(2), vDivs2UP0: vec.Vec_PtrAlloc(8), vDivs2UP1: vec.Vec_PtrAlloc(8),
		vDivs2UN0: vec.Vec_PtrAlloc(8), vDivs2UN1: vec.Vec_PtrAlloc(8),
	}
	root.pRoot.Data = []uint32{0x7}
	vec.Vec_PtrPush(root.vDivs1B, a)
	vec.Vec_PtrPush(root.vDivs1B, b)
	Abc_ManResubDivsD(root, 10)
	if vec.Vec_PtrSize(root.vDivs2UP0) == 0 {
		t.Fatal("expected source-compatible positive unate pair")
	}
	if vec.Vec_PtrSize(root.vDivs2UN0) == 0 {
		t.Fatal("expected source-compatible negative unate pair")
	}
	for i := 0; i < vec.Vec_PtrSize(root.vDivs2UP0); i++ {
		if vec.Vec_PtrEntry(root.vDivs2UP0, i) == nil || vec.Vec_PtrEntry(root.vDivs2UP1, i) == nil {
			t.Fatalf("pair %d has a nil divisor", i)
		}
	}
}

func TestAbcManResubQuitGraphsEncodeSourceForms(t *testing.T) {
	n := resubTestNetwork()
	a, b := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	root := abc.Abc_NtkCreatePi(n)
	if got := dec.Dec_GraphDeriveTruth(Abc_ManResubQuit1(root, a, b, 0)); got != 0x88888888 {
		t.Fatalf("two-input AND truth = %08x, want 88888888", got)
	}
	if got := dec.Dec_GraphDeriveTruth(Abc_ManResubQuit1(root, a, b, 1)); got != 0xeeeeeeee {
		t.Fatalf("two-input OR truth = %08x, want eeeeeeee", got)
	}
	root.Phase = true
	if got := dec.Dec_GraphDeriveTruth(Abc_ManResubQuit1(root, a, b, 0)); got != 0x77777777 {
		t.Fatalf("phase-complemented AND truth = %08x, want 77777777", got)
	}
}

func TestAbcManResubQuitZeroAndMultiInputForms(t *testing.T) {
	n := resubTestNetwork()
	a, b, c, d := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	root := abc.Abc_NtkCreatePi(n)
	if got := dec.Dec_GraphDeriveTruth(Abc_ManResubQuit0(root, a)); got != 0xaaaaaaaa {
		t.Fatalf("one-leaf graph truth = %08x, want aaaaaaaa", got)
	}
	if got := dec.Dec_GraphDeriveTruth(Abc_ManResubQuit21(root, a, b, c, 1)); got != 0xfefefefe {
		t.Fatalf("three-input OR truth = %08x, want fefefefe", got)
	}
	if got := dec.Dec_GraphDeriveTruth(Abc_ManResubQuit21(root, a, b, c, 0)); got != 0x80808080 {
		t.Fatalf("three-input AND truth = %08x, want 80808080", got)
	}
	wantOrAnd := uint32(0xaaaaaaaa) | (uint32(0xcccccccc) & uint32(0xf0f0f0f0))
	if got := dec.Dec_GraphDeriveTruth(Abc_ManResubQuit2(root, a, b, c, 1)); got != wantOrAnd {
		t.Fatalf("OR-AND truth = %08x, want %08x", got, wantOrAnd)
	}
	wantAndOr := uint32(0xaaaaaaaa) & (uint32(0xcccccccc) & uint32(0xf0f0f0f0))
	if got := dec.Dec_GraphDeriveTruth(Abc_ManResubQuit2(root, a, b, c, 0)); got != wantAndOr {
		t.Fatalf("AND-AND truth = %08x, want %08x", got, wantAndOr)
	}
	wantFour := (uint32(0xaaaaaaaa) & uint32(0xcccccccc)) | (uint32(0xf0f0f0f0) & uint32(0xff00ff00))
	if got := dec.Dec_GraphDeriveTruth(Abc_ManResubQuit3(root, a, b, c, d, 1)); got != wantFour {
		t.Fatalf("four-input OR-AND truth = %08x, want %08x", got, wantFour)
	}
}

func TestAbcManResubQuitRecognizesCareSetConstant(t *testing.T) {
	n := resubTestNetwork()
	root := abc.Abc_NtkCreatePi(n)
	manager := &Abc_ManRes_t{pRoot: root, nWords: 1, pCareSet: []uint32{0xf}}
	root.Data = []uint32{0xf0}
	if graph := Abc_ManResubQuit(manager); graph == nil {
		t.Fatal("bits outside the care set should allow a constant-zero replacement")
	} else if got := dec.Dec_GraphDeriveTruth(graph); got != 0 {
		t.Fatalf("constant-zero replacement truth = %08x", got)
	}
	root.Phase = true
	if graph := Abc_ManResubQuit(manager); graph == nil {
		t.Fatal("phase-one root should produce a constant-one replacement")
	} else if got := dec.Dec_GraphDeriveTruth(graph); got != ^uint32(0) {
		t.Fatalf("constant-one replacement truth = %08x", got)
	}
}

func TestAbcManResubDivisorSearchStages(t *testing.T) {
	n := resubTestNetwork()
	vars := []*abc.Abc_Obj_t{abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)}
	data := []uint32{0xaaaaaaaa, 0xcccccccc, 0xf0f0f0f0, 0xff00ff00}
	for i := range vars {
		vars[i].Data = []uint32{data[i]}
	}
	root := abc.Abc_NtkCreatePi(n)
	care := []uint32{^uint32(0)}
	newManager := func(result uint32) *Abc_ManRes_t {
		root.Data = []uint32{result}
		return &Abc_ManRes_t{
			pRoot: root, nWords: 1, pCareSet: care,
			vDivs: vec.Vec_PtrAlloc(4), vDivs1UP: vec.Vec_PtrAlloc(4), vDivs1UN: vec.Vec_PtrAlloc(4),
			vDivs1B: vec.Vec_PtrAlloc(4), vDivs2UP0: vec.Vec_PtrAlloc(4), vDivs2UP1: vec.Vec_PtrAlloc(4),
			vDivs2UN0: vec.Vec_PtrAlloc(4), vDivs2UN1: vec.Vec_PtrAlloc(4),
		}
	}
	manager := newManager(data[0] | data[1])
	manager.vDivs = vec.Vec_PtrAlloc(1)
	manager.nDivs = 1
	vec.Vec_PtrPush(manager.vDivs, vars[0])
	if graph := Abc_ManResubDivs0(manager); graph != nil {
		t.Fatal("divisor search accepted a nonmatching root truth table")
	}
	manager.pRoot.Data = []uint32{data[0]}
	if graph := Abc_ManResubDivs0(manager); graph == nil || dec.Dec_GraphDeriveTruth(graph) != data[0] {
		t.Fatal("divisor search did not return a matching one-node graph")
	}
	manager = newManager(data[0] | data[1])
	manager.vDivs1UP = vec.Vec_PtrAlloc(2)
	vec.Vec_PtrPush(manager.vDivs1UP, vars[0])
	vec.Vec_PtrPush(manager.vDivs1UP, vars[1])
	graph := Abc_ManResubDivs1(manager, 8)
	if graph == nil || dec.Dec_GraphDeriveTruth(graph) != data[0]|data[1] || manager.nUsedNode1Or != 1 {
		t.Fatalf("single-node OR search failed: graph=%v count=%d", graph, manager.nUsedNode1Or)
	}
	manager = newManager(data[0] | data[1] | data[2])
	vars[0].Level, vars[1].Level, vars[2].Level = 1, 2, 3
	manager.vDivs1UP = vec.Vec_PtrAlloc(3)
	for _, v := range vars[:3] {
		vec.Vec_PtrPush(manager.vDivs1UP, v)
	}
	graph = Abc_ManResubDivs12(manager, 8)
	if graph == nil || dec.Dec_GraphDeriveTruth(graph) != data[0]|data[1]|data[2] || manager.nUsedNode2Or != 1 {
		t.Fatalf("two-node OR search failed: graph=%v count=%d", graph, manager.nUsedNode2Or)
	}
	manager = newManager(data[0] & data[1] & data[2])
	manager.vDivs1UN = vec.Vec_PtrAlloc(3)
	for _, v := range vars[:3] {
		vec.Vec_PtrPush(manager.vDivs1UN, v)
	}
	graph = Abc_ManResubDivs12(manager, 8)
	if graph == nil || dec.Dec_GraphDeriveTruth(graph) != data[0]&data[1]&data[2] || manager.nUsedNode2And != 1 {
		t.Fatalf("two-node AND search failed: graph=%v count=%d", graph, manager.nUsedNode2And)
	}
	manager = newManager(data[0] | (data[1] & data[2]))
	manager.vDivs1UP, manager.vDivs2UP0, manager.vDivs2UP1 = vec.Vec_PtrAlloc(1), vec.Vec_PtrAlloc(1), vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(manager.vDivs1UP, vars[0])
	vec.Vec_PtrPush(manager.vDivs2UP0, vars[1])
	vec.Vec_PtrPush(manager.vDivs2UP1, vars[2])
	graph = Abc_ManResubDivs2(manager, 8)
	if graph == nil || dec.Dec_GraphDeriveTruth(graph) != data[0]|(data[1]&data[2]) || manager.nUsedNode2OrAnd != 1 {
		t.Fatalf("two-node OR-AND search failed: graph=%v count=%d", graph, manager.nUsedNode2OrAnd)
	}
	manager = newManager(data[0] & (data[1] & data[2]))
	manager.vDivs1UN, manager.vDivs2UN0, manager.vDivs2UN1 = vec.Vec_PtrAlloc(1), vec.Vec_PtrAlloc(1), vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(manager.vDivs1UN, vars[0])
	vec.Vec_PtrPush(manager.vDivs2UN0, vars[1])
	vec.Vec_PtrPush(manager.vDivs2UN1, vars[2])
	graph = Abc_ManResubDivs2(manager, 8)
	if graph == nil || dec.Dec_GraphDeriveTruth(graph) != data[0]&(data[1]&data[2]) || manager.nUsedNode2AndOr != 1 {
		t.Fatalf("two-node AND-OR search failed: graph=%v count=%d", graph, manager.nUsedNode2AndOr)
	}
	manager = newManager((data[0] & data[1]) | (data[2] & data[3]))
	manager.vDivs2UP0, manager.vDivs2UP1 = vec.Vec_PtrAlloc(2), vec.Vec_PtrAlloc(2)
	vec.Vec_PtrPush(manager.vDivs2UP0, vars[0])
	vec.Vec_PtrPush(manager.vDivs2UP1, vars[1])
	vec.Vec_PtrPush(manager.vDivs2UP0, vars[2])
	vec.Vec_PtrPush(manager.vDivs2UP1, vars[3])
	graph = Abc_ManResubDivs3(manager, 8)
	if graph == nil || dec.Dec_GraphDeriveTruth(graph) != (data[0]&data[1])|(data[2]&data[3]) || manager.nUsedNode3OrAnd != 1 {
		t.Fatalf("three-node OR-AND search failed: graph=%v count=%d", graph, manager.nUsedNode3OrAnd)
	}
}
