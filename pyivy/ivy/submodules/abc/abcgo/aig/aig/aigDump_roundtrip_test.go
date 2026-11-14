package aig_test

import (
	"bytes"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/base/abc"
	abcio "github.com/glycerine/abcgo/base/io"
)

func TestBLIFPortNamesDoNotCollideWithGeneratedNets(t *testing.T) {
	cases := []struct {
		name    string
		inputs  []string
		outputs []string
	}{
		{"constant", []string{"n0", "b"}, nil},
		{"gate", []string{"n3", "b"}, nil},
		{"output", []string{"n4", "b"}, nil},
		{"default_input", nil, []string{"n1"}},
		{"output_gate", nil, []string{"n3"}},
		{"reserved_suffix", []string{"n0", "n0_1"}, []string{"n3"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			a, b := g.NewInput(), g.NewInput()
			g.AddOutput(g.AND(a, b.Not()))
			var text bytes.Buffer
			if err := aig.WriteBLIF(&text, g, tc.inputs, tc.outputs); err != nil {
				t.Fatal(err)
			}
			n, err := abcio.ParseBLIF("roundtrip.blif", text.String(), false, true)
			if err != nil {
				t.Fatalf("cannot read exported BLIF: %v\n%s", err, text.String())
			}
			pis := n.ObjectsOfKind(abc.ObjectPI)
			pos := n.ObjectsOfKind(abc.ObjectPO)
			if len(pis) != 2 || len(pos) != 1 {
				t.Fatalf("interface dimensions: %d PI, %d PO", len(pis), len(pos))
			}
			for mask := 0; mask < 4; mask++ {
				values := map[*abc.Object]bool{pis[0]: mask&1 != 0, pis[1]: mask&2 != 0}
				got, err := n.Evaluate(pos[0].Signal(), values)
				if err != nil || got != (mask == 1) {
					t.Fatalf("mask %d: got %t (%v), want %t", mask, got, err, mask == 1)
				}
			}
		})
	}
}
