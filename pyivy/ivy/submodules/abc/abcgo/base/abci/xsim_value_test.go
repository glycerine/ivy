package abci

import "testing"

func TestXSimulationSourceTruthTables(t *testing.T) {
	vals := []int32{XVS0, XVS1, XVSX}
	inv := []int32{XVS1, XVS0, XVSX}
	and := [][]int32{{XVS0, XVS0, XVS0}, {XVS0, XVS1, XVSX}, {XVS0, XVSX, XVSX}}
	for i, a := range vals {
		if got := Abc_XsimInv(a); got != inv[i] {
			t.Fatalf("inv %d=%d", a, got)
		}
		for j, b := range vals {
			if got := Abc_XsimAnd(a, b); got != and[i][j] {
				t.Fatalf("%d and %d=%d", a, b, got)
			}
		}
	}
	// Source short-circuit order accepts an otherwise invalid value here.
	if Abc_XsimAnd(XVS0, 42) != XVS0 || Abc_XsimAnd(XVSX, 42) != XVSX {
		t.Fatal("source short circuit changed")
	}
}
