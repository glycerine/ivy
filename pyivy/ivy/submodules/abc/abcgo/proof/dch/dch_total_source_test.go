package dch

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

// C dchAig.c checks CI/CO counts, then builds a fresh manager without
// transferring sequential and constraint metadata from the first snapshot.
func TestSourceTotalAigDoesNotCopyMetadata(t *testing.T) {
	first, second := aig.NewGraph(), aig.NewGraph()
	for _, g := range []*aig.Graph{first, second} {
		a, b := g.NewInput(), g.NewInput()
		g.AddOutput(a)
		g.AddOutput(b)
	}
	aig.Aig_ManSetRegNum(first, 1)
	aig.Aig_ManSetConstrNum(first, 1)
	first.Name = "original"
	if _, err := DeriveTotalAig([]*aig.Graph{first, second}); err == nil {
		t.Fatal("Go convenience accepted different sequential metadata")
	}
	out := Dch_DeriveTotalAig([]*aig.Graph{first, second})
	if out.InputCount() != 2 || out.OutputCount() != 2 || aig.Aig_ManRegNum(out) != 0 || aig.Aig_ManConstrNum(out) != 0 || out.Name != "" {
		t.Fatalf("unexpected C total interface: inputs=%d outputs=%d regs=%d constraints=%d name=%q", out.InputCount(), out.OutputCount(), aig.Aig_ManRegNum(out), aig.Aig_ManConstrNum(out), out.Name)
	}
}
