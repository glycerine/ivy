package acec

import (
	"reflect"
	"testing"
)

func TestSourceSignatureDistributionSignedCoefficients(t *testing.T) {
	// Encoded exponents are -(2+1) and +(3+1); source uses -3+4-1=0.
	got := poRows(Acec_ParseSignature("(-2*o1)*(3*i0)"))
	want := [][]int32{{1, -1, 0}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("source signed distribution got %v want %v", got, want)
	}
}
