package rwr

import (
	"fmt"
	"os"
	"testing"
)

func TestLibrarySourceInsertionDominanceAndVolume(t *testing.T) {
	p := &Rwr_Man_t{pTable: make([]*Rwr_Node_t, 1<<16), puCanons: make([]uint16, 1<<16), pPractical: make([]bool, 1<<16)}
	for i := range p.puCanons {
		p.puCanons[i] = uint16(i)
		p.pPractical[i] = true
	}
	a := Rwr_ManAddVar(p, 0xaaaa, 0)
	b := Rwr_ManAddVar(p, 0xcccc, 0)
	x, y := Rwr_NodeRef{node: a}, Rwr_NodeRef{node: b}
	first := Rwr_ManTryNode(p, x, Rwr_Not(y), 0, 1, 1)
	if first == nil || first.uTruth != 0x2222 || p.pTable[0x2222] != first {
		t.Fatal("failed to insert complemented AND")
	}
	if Rwr_ManTryNode(p, x, Rwr_Not(y), 0, 2, 2) != nil {
		t.Fatal("source-dominated candidate accepted")
	}
	equal := Rwr_ManTryNode(p, x, Rwr_Not(y), 0, 1, 1)
	if equal == nil || first.pNext != equal {
		t.Fatal("source equal-volume candidate must append")
	}
	xor := Rwr_ManAddNode(p, Rwr_NodeRef{node: first}, x, 1, 2, 3)
	volume := Rwr_ManNodeVolume(p, xor, first)
	if volume != 3 {
		t.Fatalf("shared XOR volume %d", volume)
	}
	expected, err := os.ReadFile("test_vectors/library_results.txt")
	if err != nil {
		t.Fatal(err)
	}
	got := fmt.Sprintf("%d %d %d %d %d %d %d %d %d\n", first.uTruth, 1, 1, volume, p.nConsidered, p.nAdded, p.nClasses, len(p.vForest), xor.uTruth)
	if got != string(expected) {
		t.Fatalf("source: %q Go: %q", expected, got)
	}
}
