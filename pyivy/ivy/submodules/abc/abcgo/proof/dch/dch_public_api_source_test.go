package dch

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// These public spellings mirror dch.h; test them alongside their typed
// counterparts so callers do not silently lose the source entry points.
func TestSourcePublicEntryPoints(t *testing.T) {
	params := DefaultParams()
	if Dch_ManReadVerbose(&params) != 0 {
		t.Fatal("default verbosity is enabled")
	}
	params.Verbose = true
	if Dch_ManReadVerbose(&params) != 1 {
		t.Fatal("verbose parameter was not read")
	}

	variants, _, _ := associativeSnapshots()
	total := Dch_DeriveTotalAig(variants)
	if err := Dch_ComputeEquivalences(total, nil); err != nil {
		t.Fatal(err)
	}
	choices, err := Dch_DeriveChoiceAig(total, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := Dch_CheckChoices(choices, 0); err != nil {
		t.Fatalf("source choice checker rejected derived graph: %v", err)
	}
	if err := Dch_CheckChoices(nil, 0); err == nil {
		t.Fatal("source checker accepted nil graph")
	}

	variants, _, _ = associativeSnapshots()
	total = Dch_DeriveTotalAig(variants)
	computed, err := Dch_ComputeChoices(total, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := CheckChoices(computed, false); err != nil {
		t.Fatal(err)
	}

	graph := aig.NewGraph()
	x := graph.NewInput()
	graph.AddOutput(x)
	if err := Dch_ComputeEquivalences(graph, nil); err != nil {
		t.Fatal(err)
	}
}
