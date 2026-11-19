package goivy

import "testing"

// Python sig.interp is a dict: compile_theories iterates it in insertion
// order, and sig.copy() keeps that order. goivy's Sig.Interp was a Go map
// (random order); it is now an InsMap, and Copy must keep the order.
func TestSigInterpKeepsInsertionOrderThroughCopy(t *testing.T) {
	s := NewSig()
	for _, k := range []string{"zeta", "alpha", "mid"} {
		s.Interp.Set(k, "int")
	}
	c := s.Copy()
	var got []string
	for k := range c.Interp.All() {
		got = append(got, k)
	}
	if len(got) != 3 || got[0] != "zeta" || got[1] != "alpha" || got[2] != "mid" {
		t.Fatalf("copied interp order = %v, want [zeta alpha mid]", got)
	}
	c.Interp.Set("extra", "int")
	if _, ok := s.Interp.Get2("extra"); ok {
		t.Fatal("Copy shares Interp with the original")
	}
}
