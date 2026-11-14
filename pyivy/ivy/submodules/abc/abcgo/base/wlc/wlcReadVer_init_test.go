package wlc

import "testing"

func addInitTestFanin(t *testing.T, ntk *Wlc_Ntk_t, obj *Wlc_Obj_t, fanins ...int32) {
	t.Helper()
	v := Vec_IntAlloc(len(fanins))
	for _, fanin := range fanins {
		Vec_IntPush(v, fanin)
	}
	Wlc_ObjAddFanins(ntk, obj, v)
	Vec_IntFree(v)
}

func newInitTestNetwork() *Wlc_Ntk_t {
	ntk := Wlc_NtkAlloc("init-test", 8)
	ntk.vInits = Vec_IntAlloc(1)
	return ntk
}

func TestWlcPrsConvertInitValuesNegativeWidth(t *testing.T) {
	ntk := newInitTestNetwork()
	Vec_IntPush(ntk.vInits, -3)

	got := Wlc_PrsConvertInitValues(ntk)
	if want := "000\x00"; got != want {
		t.Fatalf("Wlc_PrsConvertInitValues() = %q, want %q", got, want)
	}
	if gotWidth := Vec_IntEntry(ntk.vInits, 0); gotWidth != -3 {
		t.Fatalf("negative init entry changed to %d, want -3", gotWidth)
	}
}

func TestWlcPrsConvertInitValuesBufferedConstantZeroExtend(t *testing.T) {
	ntk := newInitTestNetwork()
	constID := Wlc_ObjAlloc(ntk, WLC_OBJ_CONST, 0, 1, 0)
	addInitTestFanin(t, ntk, Wlc_NtkObj(ntk, constID), 2) // bits 1:0 = 10
	bufID := Wlc_ObjAlloc(ntk, WLC_OBJ_BUF, 0, 3, 0)
	addInitTestFanin(t, ntk, Wlc_NtkObj(ntk, bufID), constID)
	Vec_IntPush(ntk.vInits, bufID)

	got := Wlc_PrsConvertInitValues(ntk)
	if want := "0100\x00"; got != want {
		t.Fatalf("Wlc_PrsConvertInitValues() = %q, want %q", got, want)
	}
	if gotWidth := Vec_IntEntry(ntk.vInits, 0); gotWidth != -4 {
		t.Fatalf("constant init entry = %d, want -4", gotWidth)
	}
}

func TestWlcPrsConvertInitValuesNonconstantInput(t *testing.T) {
	ntk := newInitTestNetwork()
	piID := Wlc_ObjAlloc(ntk, WLC_OBJ_PI, 0, 2, 0)
	Vec_IntPush(ntk.vInits, piID)

	got := Wlc_PrsConvertInitValues(ntk)
	if want := "xxx\x00"; got != want {
		t.Fatalf("Wlc_PrsConvertInitValues() = %q, want %q", got, want)
	}
	if gotCI := Vec_IntEntry(ntk.vInits, 0); gotCI != 0 {
		t.Fatalf("nonconstant init entry = %d, want CI index 0", gotCI)
	}
}
