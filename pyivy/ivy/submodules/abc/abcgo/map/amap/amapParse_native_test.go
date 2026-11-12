package amap

import "testing"

func TestAmapLibraryEquationTruthTables(t *testing.T) {
	lib := &Amap_Lib_t_{vGates: Vec_PtrAlloc(4)}
	and := &Amap_Gat_t_{pName: "AND2", pForm: "A * B", nPins: 2, Pins: []Amap_Pin_t{{pName: "A"}, {pName: "B"}}}
	xor := &Amap_Gat_t_{pName: "XOR2", pForm: "A ^ B", nPins: 2, Pins: []Amap_Pin_t{{pName: "A"}, {pName: "B"}}}
	constant := &Amap_Gat_t_{pName: "CONST1", pForm: AMAP_STRING_CONST1}
	ignored := &Amap_Gat_t_{pName: "BUF_WITH_UNUSED_PIN", pForm: "A", nPins: 2, Pins: []Amap_Pin_t{{pName: "A"}, {pName: "B"}}}
	for _, gate := range []*Amap_Gat_t{and, xor, constant, ignored} {
		Vec_PtrPush(lib.vGates, gate)
	}
	if got := Amap_LibParseEquations(lib, 0); got != 1 {
		t.Fatal("library equation parsing failed")
	}
	if len(and.pFunc) != 1 || and.pFunc[0]&0xf != 0x8 {
		t.Fatalf("AND2 truth = %08x", and.pFunc)
	}
	if len(xor.pFunc) != 1 || xor.pFunc[0]&0xf != 0x6 {
		t.Fatalf("XOR2 truth = %08x", xor.pFunc)
	}
	if len(constant.pFunc) != 1 || constant.pFunc[0] != ^uint32(0) {
		t.Fatalf("CONST1 truth = %08x", constant.pFunc)
	}
	if ignored.pFunc != nil {
		t.Fatalf("gate with unused pin was retained: %08x", ignored.pFunc)
	}
}

func TestAmapLibraryEquationRejectsInvalidFormula(t *testing.T) {
	lib := &Amap_Lib_t_{vGates: Vec_PtrAlloc(1)}
	Vec_PtrPush(lib.vGates, &Amap_Gat_t_{pName: "BROKEN", pForm: "A +", nPins: 1, Pins: []Amap_Pin_t{{pName: "A"}}})
	if got := Amap_LibParseEquations(lib, 0); got != 0 {
		t.Fatal("invalid formula unexpectedly accepted")
	}
}
