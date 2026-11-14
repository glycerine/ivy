package wln

import (
	"github.com/glycerine/abcgo/aig/miniaig"
	"testing"
)

// wlnNdr.c exports WLN object IDs as NDR signal IDs, including the source's
// use of the CO object ID for its single input.
func TestNetworkToNdrSourceObjectMapping(t *testing.T) {
	p := Wln_NtkAlloc("", 8)
	a := Wln_ObjAlloc(p, ABC_OPER_CI, 0, 3, 0)
	n := Wln_ObjAlloc(p, miniaig.ABC_OPER_LOGIC_NOT, 0, 3, 0)
	Wln_ObjAddFanin(p, n, a)
	co := Wln_ObjAlloc(p, ABC_OPER_CO, 0, 3, 0)
	Wln_ObjAddFanin(p, co, n)
	d := Wln_NtkToNdr(p)
	objects := wlnNdrObjects(d, 2)
	if len(objects) != 3 || miniaig.Ndr_DataCiNum(d, 2) != 1 || miniaig.Ndr_DataCoNum(d, 2) != 1 {
		t.Fatalf("objects %v", objects)
	}
	var in []int
	miniaig.Ndr_ObjReadArray(d, objects[1], miniaig.NDR_INPUT, &in)
	if len(in) != 1 || in[0] != int(a) {
		t.Fatalf("internal fanins %v", in)
	}
	miniaig.Ndr_ObjReadArray(d, objects[2], miniaig.NDR_INPUT, &in)
	if len(in) != 1 || in[0] != int(n) {
		t.Fatalf("CO input %v", in)
	}
	var end, beg int32
	if Ndr_ObjGetRange(d, int32(objects[1]), &end, &beg) != 0 || end != 3 || beg != 0 {
		t.Fatalf("range %d:%d", end, beg)
	}
}

func TestNetworkToNdrPreservesZeroFanin(t *testing.T) {
	p := Wln_NtkAlloc("", 4)
	ci := Wln_ObjAlloc(p, ABC_OPER_CI, 0, 0, 0)
	node := Wln_ObjAlloc(p, miniaig.ABC_OPER_LOGIC_AND, 0, 0, 0)
	Wln_ObjAddFanin(p, node, 0)
	Wln_ObjAddFanin(p, node, ci)
	d := Wln_NtkToNdr(p)
	objects := wlnNdrObjects(d, 2)
	var got []int
	miniaig.Ndr_ObjReadArray(d, objects[1], miniaig.NDR_INPUT, &got)
	if len(got) != 2 || got[0] != 0 || got[1] != int(ci) {
		t.Fatalf("internal fanins %v, want [0 %d]", got, ci)
	}
}

func TestNdrObjectRangeForms(t *testing.T) {
	d := miniaig.Ndr_Create(1)
	mod := miniaig.Ndr_AddModule(d, 1)
	cases := []struct{ end, beg, signed int }{{0, 0, 0}, {5, 5, 0}, {7, 2, 0}, {7, 2, 1}}
	for i, tc := range cases {
		miniaig.Ndr_AddObject(d, mod, miniaig.ABC_OPER_CI, 0, tc.end, tc.beg, tc.signed, 0, nil, 1, []int{i + 1}, nil)
	}
	objects := wlnNdrObjects(d, 2)
	if len(objects) != len(cases) {
		t.Fatalf("got %d objects", len(objects))
	}
	for i, tc := range cases {
		var end, beg int32
		signed := Ndr_ObjGetRange(d, int32(objects[i]), &end, &beg)
		if int(end) != tc.end || int(beg) != tc.beg || int(signed) != tc.signed {
			t.Errorf("range %d: got (%d,%d,%d), want (%d,%d,%d)", i, end, beg, signed, tc.end, tc.beg, tc.signed)
		}
	}
}

func TestReadNdrReturnsImportedNetwork(t *testing.T) {
	t.Chdir(t.TempDir())
	p := Wln_NtkAlloc("", 4)
	ci := Wln_ObjAlloc(p, ABC_OPER_CI, 0, 0, 0)
	co := Wln_ObjAlloc(p, ABC_OPER_CO, 0, 0, 0)
	Wln_ObjAddFanin(p, co, ci)
	d := Wln_NtkToNdr(p)
	miniaig.Ndr_Write("roundtrip.ndr", d)
	miniaig.Ndr_Delete(d)

	got := Wln_ReadNdr("roundtrip.ndr")
	if got == nil {
		t.Fatal("Wln_ReadNdr returned nil for a valid NDR file")
	}
	if Wln_NtkCiNum(got) != 1 || Wln_NtkCoNum(got) != 1 {
		t.Fatalf("imported interface: %d CIs, %d COs", Wln_NtkCiNum(got), Wln_NtkCoNum(got))
	}
}
