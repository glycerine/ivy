package ivy

import "testing"

func TestFraigParamsSetBmc(t *testing.T) {
	var params Ivy_FraigParams_t
	Ivy_FraigParamsDefault(&params)
	Ivy_FraigParamsSetBmc(&params, 0x3fffffff, 0, 0)
	if params.nBTLimitNode != 0x3fffffff {
		t.Fatalf("node backtrack limit = %d, want %d", params.nBTLimitNode, 0x3fffffff)
	}
	if params.fVerbose != 0 || params.fProve != 0 {
		t.Fatalf("verbose/prove = %d/%d, want 0/0", params.fVerbose, params.fProve)
	}
}
