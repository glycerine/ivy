package int2

import "testing"

// The checked-in constructor references an undeclared pFrames; no complete
// source contract exists for allocating its structural hash table.
func TestSourceManagerConstructorUnsupported(t *testing.T) {
	if got := Int2_ManCreate(nil, nil); got != nil {
		t.Fatal("incomplete source constructor unexpectedly returned a manager")
	}
}
