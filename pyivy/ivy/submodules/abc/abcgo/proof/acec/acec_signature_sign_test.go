package acec

import (
	"reflect"
	"testing"
)

func TestSignatureDistributionPreservesSigns(t *testing.T) {
	got := Acec_ParseSignature("(1*i0)*(1*i1-1*i2)")
	want := [][]int32{{-1, -2, 3}, {-1, -3, -3}}
	if rows := poRows(got); !reflect.DeepEqual(rows, want) {
		t.Fatalf("distributed monomials %v want %v", rows, want)
	}
}

func TestSignatureDistributionSignCombinations(t *testing.T) {
	got := Acec_ParseSignature("(1*i0-2*i1)*(3*i2-4*i3)")
	want := [][]int32{{-1, -3, 5}, {-1, -4, -6}, {-2, -3, -6}, {-2, -4, 7}}
	if rows := poRows(got); !reflect.DeepEqual(rows, want) {
		t.Fatalf("distributed monomials %v want %v", rows, want)
	}
}
