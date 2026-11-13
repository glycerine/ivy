package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"io"
	"os"
	"strings"
	"testing"
)

func TestCec4EntryVerboseReportsSimulationAndSAT(t *testing.T) {
	g := gia.Gia_ManStart(16)
	x, y, z := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	a := gia.Gia_ManAppendAnd(g, x, y)
	b := gia.Gia_ManAppendAnd(g, a, x)
	gia.Gia_ManAppendCo(g, b)
	gia.Gia_ManAppendCo(g, z)
	pars := Cec_ParFra_t{jType: 0, nWords: 1, nRounds: 1, nItersMax: 4, nBTLimit: 1000, fVerbose: 1}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	var out *Gia_Man_t
	status := Cec4_ManPerformSweeping(g, &pars, &out, 0)
	w.Close()
	os.Stdout = old
	output, err := io.ReadAll(r)
	r.Close()
	if err != nil {
		t.Fatal(err)
	}
	if status != 1 || out == nil {
		t.Fatalf("status %d", status)
	}
	text := string(output)
	for _, part := range []string{"Solver type =", "Sim    0", "SAT ", "SAT calls =", "Sim = 2", "Generate CNF", "TOTAL"} {
		if !strings.Contains(text, part) {
			t.Fatalf("missing %q in report %q", part, text)
		}
	}
}

func TestCec4GeneratedPatternsCountedBeforeEarlyMiterExit(t *testing.T) {
	g := gia.Gia_ManStart(40)
	inputs := make([]int32, 12)
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(g)
	}
	root := inputs[0]
	for _, lit := range inputs[1:] {
		root = gia.Gia_ManAppendAnd(g, root, lit)
	}
	gia.Gia_ManAppendCo(g, root)
	pars := Cec_ParFra_t{jType: 0, nWords: 1, nRounds: 0, nGenIters: 1, nItersMax: 2000, fCheckMiter: 1}
	manager := Cec4_ManCreate(g, &pars)
	defer Cec4_ManDestroy(manager)
	_, witness := cec4PrepareSimulation(g, &pars, false, manager)
	if witness == nil {
		t.Fatal("expected generated miter counterexample")
	}
	if manager.nSatSat <= 0 {
		t.Fatal("generated pattern count was not recorded before miter exit")
	}
}

func TestCec4GeneratedPatternCountResetBeforeSAT(t *testing.T) {
	g := gia.Gia_ManStart(40)
	inputs := make([]int32, 12)
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(g)
	}
	root := inputs[0]
	for _, lit := range inputs[1:] {
		root = gia.Gia_ManAppendAnd(g, root, lit)
	}
	gia.Gia_ManAppendCo(g, root)
	pars := Cec_ParFra_t{jType: 0, nWords: 1, nRounds: 0, nGenIters: 1, nItersMax: 2000, nBTLimit: 1000}
	manager := Cec4_ManCreate(g, &pars)
	defer Cec4_ManDestroy(manager)
	_, witness, stats := cec4Sweep(g, &pars, false, false, manager)
	if witness != nil || manager.nSatSat != int32(stats.sat) {
		t.Fatalf("generated patterns leaked into SAT count: manager %d SAT %d", manager.nSatSat, stats.sat)
	}
}
