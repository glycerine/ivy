package abc

// Go-only behavior checks for the active shell in opt/dau/dauCore.c.

import "testing"

func TestDauTryPerform(t *testing.T) {
	if !DauTryPerform(nil, DauTryParameters{}) {
		t.Fatal("nil source inputs must retain the unconditional success result")
	}
	graph := NewAigManager()
	before := graph.ObjectCount()
	for _, parameter := range []int{minInt(), -1, 0, 1, int(^uint(0) >> 1)} {
		if !DauTryPerform(graph, DauTryParameters{Parameter: parameter}) {
			t.Fatalf("parameter %d failed", parameter)
		}
		if graph.ObjectCount() != before {
			t.Fatalf("parameter %d mutated graph", parameter)
		}
	}
}

func minInt() int { return -int(^uint(0)>>1) - 1 }
