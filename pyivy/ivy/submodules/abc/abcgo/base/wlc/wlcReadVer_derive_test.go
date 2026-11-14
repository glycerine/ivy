package wlc

import "testing"

func TestWlcReadVerDerivesSimpleCombinationalModule(t *testing.T) {
	const source = `module top(input a, input b, output y);
assign y = a & b;
endmodule
`
	n := Wlc_ReadVer("", source)
	if n == nil {
		t.Fatal("Wlc_ReadVer returned nil for a simple module")
	}
	defer Wlc_NtkFree(n)
	if got := Wlc_NtkPiNum(n); got != 2 {
		t.Fatalf("PI count = %d, want 2", got)
	}
	if got := Wlc_NtkPoNum(n); got != 1 {
		t.Fatalf("PO count = %d, want 1", got)
	}
	po := Wlc_NtkPo(n, 0)
	if po == nil || Wlc_ObjName(n, Wlc_ObjId(n, po)) != "y" {
		t.Fatalf("first PO = %v, want y", po)
	}
}
