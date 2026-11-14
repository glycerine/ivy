package cba

import "testing"

func TestSourceSopMappings(t *testing.T) {
	for _, tc := range []struct {
		s     string
		kind  Cba_ObjType_t
		label string
	}{
		{" 0\n", CBA_BOX_CF, "CBA_BOX_C0"}, {" 1\n", CBA_BOX_CT, "CBA_BOX_C1"}, {"1 1\n", CBA_BOX_BUF, "CBA_BOX_BUF"},
		{"0 1\n", CBA_BOX_INV, "CBA_BOX_INV"}, {"11 1\n", CBA_BOX_AND, "CBA_BOX_AND"}, {"00 1\n", CBA_BOX_NOR, "CBA_BOX_NOR"},
		{"00 0\n", CBA_BOX_OR, "CBA_BOX_OR"}, {"-1 1\n1- 1\n", CBA_BOX_OR, "CBA_BOX_OR"}, {"1- 1\n-1 1\n", CBA_BOX_OR, "CBA_BOX_OR"},
		{"01 1\n10 1\n", CBA_BOX_XOR, "CBA_BOX_XOR"}, {"10 1\n01 1\n", CBA_BOX_XOR, "CBA_BOX_XOR"},
		{"11 1\n00 1\n", CBA_BOX_XNOR, "CBA_BOX_XNOR"}, {"00 1\n11 1\n", CBA_BOX_XNOR, "CBA_BOX_XNOR"},
		{"10 1\n", CBA_BOX_SHARP, "CBA_BOX_SHARP"}, {"01 1\n", CBA_BOX_SHARPL, "CBA_BOX_SHARPL"},
	} {
		if got := Ptr_SopToType(tc.s); got != tc.kind || Ptr_SopToTypeName(tc.s) != tc.label {
			t.Fatalf("%q kind=%d", tc.s, got)
		}
	}
	if Ptr_SopToType("11 1\n\x00ignored") != CBA_BOX_AND {
		t.Fatal("C strcmp terminator")
	}
	if Ptr_TypeToSop(CBA_BOX_NAND) != "11 0\n" || Ptr_TypeToName(CBA_BOX_TRI) != "bufifl" {
		t.Fatal("source output mapping")
	}
}
func TestSourceUnsupportedSopAsserts(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected source assertion")
		}
	}()
	Ptr_SopToType("11 0\n")
}
