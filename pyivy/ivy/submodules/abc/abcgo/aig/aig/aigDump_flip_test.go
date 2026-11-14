package aig_test

import (
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestDumpWritersObserveFirstPoFlip(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b.Not()))
	aig.Aig_ManFlipFirstPo(g)
	for _, format := range []string{"blif", "v"} {
		t.Run(format, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "flipped."+format)
			var err error
			if format == "blif" {
				err = aig.Aig_ManDumpBlif(g, path, nil, nil)
			} else {
				err = aig.Aig_ManDumpVerilog(g, path)
			}
			if err != nil {
				t.Fatal(err)
			}
			n := readDumpNetwork(t, path, format, 0)
			checkDumpNetwork(t, g, n, false)
		})
	}
}
