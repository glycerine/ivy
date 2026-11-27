package io

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
)

// TestBenchReadsGndVddGatesAsConstants: Io_ReadBenchNetwork reads "x = gnd"
// and "x = vdd" as constant nodes without fanins (the file ABC writes for a
// LUT with a constant input), not as buffers of extra "gnd"/"vdd" nets.
func TestBenchReadsGndVddGatesAsConstants(t *testing.T) {
	text := "INPUT(a)\nOUTPUT(z)\nz           = LUT 0x6 ( a, new_gnd_ )\nnew_gnd_    = gnd\n"
	n, err := ParseBench("c.bench", text, true)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(n.ObjectsOfKind(abc.ObjectNode)); got != 2 {
		t.Fatalf("%d nodes, want 2 (the LUT and a constant)", got)
	}
	checkBooleanNetwork(t, n, func(a []bool) bool { return a[0] })
}

// TestReadersNameNetworkLikeAbcNtkStartRead: bench, eqn and pla readers
// name the network Extra_FileNameGeneric(file) and keep the file as spec.
func TestReadersNameNetworkLikeAbcNtkStartRead(t *testing.T) {
	b, err := ParseBench("dir/x.bench", "INPUT(a)\nOUTPUT(z)\nz = NOT(a)\n", true)
	if err != nil {
		t.Fatal(err)
	}
	e, err := ParseEqn("dir/y.eqn", "INORDER = a; OUTORDER = z; z = !a;", true)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		n          *abc.Network
		name, spec string
	}{{b, "dir/x", "dir/x.bench"}, {e, "dir/y", "dir/y.eqn"}} {
		if c.n.Name != c.name || c.n.Spec != c.spec {
			t.Errorf("name %q spec %q, want %q %q", c.n.Name, c.n.Spec, c.name, c.spec)
		}
	}
}

// TestBenchLutsCollapseThroughBdds: Io_ReadBenchNetwork turns LUT truth
// tables into covers through BDDs (Abc_NtkToBdd, Abc_NtkToSop), so a
// majority LUT reads as its three-cube ISOP, not four minterms.
func TestBenchLutsCollapseThroughBdds(t *testing.T) {
	n, err := ParseBench("m.bench", "INPUT(a)\nINPUT(b)\nINPUT(c)\nOUTPUT(z)\nz = LUT 0xe8 ( a, b, c )\n", true)
	if err != nil {
		t.Fatal(err)
	}
	nodes := n.ObjectsOfKind(abc.ObjectNode)
	if len(nodes) != 1 {
		t.Fatalf("%d nodes", len(nodes))
	}
	if cubes := abc.Abc_SopGetCubeNum(nodes[0].Data.(string)); cubes != 3 {
		t.Fatalf("%d cubes in %q, want 3", cubes, nodes[0].Data)
	}
}
