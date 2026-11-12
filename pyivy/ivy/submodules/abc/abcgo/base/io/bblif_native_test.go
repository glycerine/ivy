package io

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/bbl"
)

func TestBBLIFNetlistRoundTrip(t *testing.T) {
	n, err := ParseBLIF("xor.blif", ".model xor\n.inputs a b\n.outputs y\n.names a b y\n01 1\n10 1\n.end\n", false, true)
	if err != nil {
		t.Fatal(err)
	}
	manager, err := BBLIFfromNetwork(n)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err = manager.WriteBinary(&output); err != nil {
		t.Fatal(err)
	}
	manager, err = bbl.ParseBinary(output.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	logic, err := NetworkFromBBLIF(manager)
	if err != nil {
		t.Fatal(err)
	}
	network, err := BBLIFToAIG(manager)
	if err != nil {
		t.Fatal(err)
	}
	for _, copy := range []*abc.Network{logic, network} {
		pi := copy.ObjectsOfKind(abc.ObjectPI)
		po := copy.ObjectsOfKind(abc.ObjectPO)[0]
		for bits := 0; bits < 4; bits++ {
			got, err := copy.Evaluate(po.Signal(), map[*abc.Object]bool{pi[0]: bits&1 != 0, pi[1]: bits&2 != 0})
			if err != nil || got != (bits == 1 || bits == 2) {
				t.Fatal(bits, got, err)
			}
		}
	}
	path := filepath.Join(t.TempDir(), "network.bblif")
	if err = WriteBBLIF(n, path); err != nil {
		t.Fatal(err)
	}
	if _, err = ReadBBLIF(path, true); err != nil {
		t.Fatal(err)
	}
}
