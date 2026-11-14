package abci

import (
	"testing"

	baseabc "github.com/glycerine/abcgo/base/abc"
)

func TestExactMapperManagerLifecycle(t *testing.T) {
	if Abc_ExactIsRunning() != 0 {
		t.Fatal("exact manager unexpectedly active at test start")
	}
	if Abc_ExactInputNum() != 8 {
		t.Fatalf("exact input limit = %d, want 8", Abc_ExactInputNum())
	}
	Abc_ExactStart(100, 0, 0, 0, "")
	first := s_pSesStore
	if Abc_ExactIsRunning() != 1 || first == nil {
		t.Fatal("start did not activate exact manager")
	}
	Abc_ExactStart(200, 1, 1, 0, "")
	if s_pSesStore != first {
		t.Fatal("second start replaced active manager")
	}
	Abc_ExactStop("")
	if Abc_ExactIsRunning() != 0 || s_pSesStore != nil {
		t.Fatal("stop did not clear exact manager")
	}
	Abc_ExactStart(100, 0, 0, 0, "")
	if Abc_ExactIsRunning() != 1 {
		t.Fatal("manager could not be restarted after stop")
	}
	Abc_ExactStop("")
}

func TestExactTruthTableNetworkShape(t *testing.T) {
	truth := [4]word{0xcafe, 0, 0, 0}
	ntk := Abc_NtkFromTruthTable(&truth[0], 4)
	if ntk == nil {
		t.Fatal("truth-table conversion returned nil")
	}
	defer baseabc.Abc_NtkDelete(ntk)
	if got := baseabc.Abc_NtkPiNum(ntk); got != 4 {
		t.Fatalf("truth network input count = %d, want 4", got)
	}
	if got := baseabc.Abc_NtkPoNum(ntk); got != 1 {
		t.Fatalf("truth network output count = %d, want 1", got)
	}
}
