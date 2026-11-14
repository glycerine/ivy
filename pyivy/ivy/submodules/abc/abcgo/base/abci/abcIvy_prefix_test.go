package abci

import (
	"testing"

	baseabc "github.com/glycerine/abcgo/base/abc"
)

func TestIvyEdgeEncodingAndNetworkDecode(t *testing.T) {
	network := baseabc.NewNetwork("ivy-edge")
	object := baseabc.Abc_NtkCreateNode(network)
	for _, complement := range []int32{0, 1} {
		encoded := Abc_EdgeCreate(int32(object.ID), complement)
		if got := Abc_EdgeId(encoded); got != int32(object.ID) {
			t.Fatalf("edge id = %d, want %d", got, object.ID)
		}
		if got := Abc_EdgeIsComplement(encoded); got != complement {
			t.Fatalf("edge complement = %d, want %d", got, complement)
		}
		decoded := Abc_EdgeToNode(network, encoded)
		if decoded.Object != object || decoded.Complement != (complement != 0) {
			t.Fatalf("decoded edge = (%p, %v), want (%p, %v)", decoded.Object, decoded.Complement, object, complement != 0)
		}
	}
}

func TestIvyEdgeRegularAndNot(t *testing.T) {
	const edge = Abc_Edge_t(0x35)
	if got := Abc_EdgeRegular(edge); got != 0x34 {
		t.Fatalf("regular edge = %#x, want %#x", got, 0x34)
	}
	if got := Abc_EdgeNot(edge); got != 0x34 {
		t.Fatalf("complemented edge = %#x, want %#x", got, 0x34)
	}
	if got := Abc_EdgeNotCond(edge, 0); got != edge {
		t.Fatalf("conditional complement with false condition = %#x, want %#x", got, edge)
	}
}

func TestAbcNtkIvyMatchesActiveCStubBody(t *testing.T) {
	// The C routine stops after validating its Ivy manager and returns NULL;
	// the rewrite and conversion-back code below it is commented out.
	network := baseabc.NewNetwork("ivy-active-body")
	if got := Abc_NtkIvy(network); got != nil {
		t.Fatalf("Abc_NtkIvy returned %p, want nil per active C body", got)
	}
}
