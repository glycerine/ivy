package cba

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestVerilogPrimitiveTables(t *testing.T) {
	primitives := []struct {
		typ  int32
		name string
	}{
		{int32(CBA_BOX_BUF), "buf"},
		{int32(CBA_BOX_INV), "not"},
		{int32(CBA_BOX_AND), "and"},
		{int32(CBA_BOX_NAND), "nand"},
		{int32(CBA_BOX_OR), "or"},
		{int32(CBA_BOX_NOR), "nor"},
		{int32(CBA_BOX_XOR), "xor"},
		{int32(CBA_BOX_XNOR), "xnor"},
		{int32(CBA_BOX_TRI), "bufif1"},
	}
	for i, want := range primitives {
		got := s_VerilogPrims[i]
		if got.Type != want.typ || got.pName != want.name {
			t.Errorf("primitive %d = (%d, %q), want (%d, %q)", i, got.Type, got.pName, want.typ, want.name)
		}
	}
	if got := s_VerilogPrims[len(primitives)]; got.Type != 0 || got.pName != "" {
		t.Errorf("primitive table sentinel = (%d, %q), want (0, empty)", got.Type, got.pName)
	}

	if got := s_VerInfo[0]; got.Type != -1 || got.nInputs != 0 || got.pTypeName != "" {
		t.Errorf("operator table sentinel = (%d, %d, %q), want (-1, 0, empty)", got.Type, got.nInputs, got.pTypeName)
	}
	if got := s_VerInfo[7]; got.Type != int32(CBA_BOX_AND) || got.nInputs != 1 || got.pTypeName != "VERIFIC_AND" || !reflect.DeepEqual(got.pSigNames[:3], []string{"a0", "a1", "o"}) {
		t.Errorf("VERIFIC_AND metadata = (%d, %d, %q, %v)", got.Type, got.nInputs, got.pTypeName, got.pSigNames)
	}
	if got := s_VerInfo[53]; got.pTypeName != "ReadPort_" || !reflect.DeepEqual(got.pSigNames[:4], []string{"read_enable", "read_address", "Ram", "read_data"}) {
		t.Errorf("ReadPort metadata = (%q, %v)", got.pTypeName, got.pSigNames)
	}
}

func TestVerilogBuilderMapsInputsAndOutput(t *testing.T) {
	const source = `module demo(input a, input b, output y);
assign y = a & b;
endmodule
`
	path := filepath.Join(t.TempDir(), "demo.v")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	manager := Cba_ManReadVerilog(path)
	if manager == nil {
		t.Fatal("Cba_ManReadVerilog returned nil")
	}
	defer Cba_ManFree(manager)
	ntk := Cba_ManNtk(manager, 1)
	if ntk == nil {
		t.Fatal("builder did not create the top-level network")
	}
	if got := Cba_NtkPiNum(ntk); got != 2 {
		t.Fatalf("input count = %d, want 2", got)
	}
	if got := Cba_NtkPoNum(ntk); got != 1 {
		t.Fatalf("output count = %d, want 1", got)
	}
	po := Cba_NtkPo(ntk, 0)
	if got := Cba_ObjFinFon(ntk, po, 0); got == 0 {
		t.Fatal("output y was not connected to its driver")
	}
}
