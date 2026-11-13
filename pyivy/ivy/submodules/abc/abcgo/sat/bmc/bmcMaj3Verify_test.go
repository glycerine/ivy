package bmc

import "testing"

func TestZyxReadNodes(t *testing.T) {
	truth := make([]uint64, 4)
	var vars, lut, nodes int32
	if Zyx_TestGetTruthTablePars("8-2-1.txt", truth, &vars, &lut, &nodes) != 1 || vars != 2 || lut != 2 || nodes != 1 {
		t.Fatalf("parameters %d %d %d", vars, lut, nodes)
	}
	info := Zyx_TestCreateTruthTables(vars, nodes)
	if Zyx_TestReadNode("C1000ab\n", info, vars, lut, 2) != 1 {
		t.Fatal("node rejected")
	}
	got := Zyx_TestTruth(info, 2, 1)
	if got[0] != 0x8888888888888888 {
		t.Fatalf("and truth %x", got[0])
	}
}
