package fra

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestFraSmlManagerSimulationAndRefinement(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, y))
	m := &Fra_Man_t{PManAig: g, PPars: &Fra_Par_t{DSimSatur: 0.1}, PCla: Fra_ClassesStart(g), PSml: Fra_SmlStart(g, 0, 1, 1), PPatWords: make([]uint32, 1), NFramesAll: 1}
	Fra_SmlSimulate(m, 0)
	if m.PCla.Simulation != m.PSml || m.PSml.Rounds < 4 {
		t.Fatalf("manager did not run simulation and attach class partition: rounds %d", m.PSml.Rounds)
	}
	Fra_SmlSavePattern1(m, 0)
	if m.PPatWords[0] != ^uint32(0) {
		t.Fatal("all-one pattern not set")
	}
	Fra_SmlSavePattern0(m, 0)
	if m.PPatWords[0] != 0 {
		t.Fatal("all-zero pattern not set")
	}
}

func TestFraSmlSimulationFromFile(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	output := g.AddOutput(g.XOR(x, y))
	file := filepath.Join(t.TempDir(), "patterns")
	if err := os.WriteFile(file, []byte("00 01\n10\t11"), 0600); err != nil {
		t.Fatal(err)
	}
	s := Fra_SmlSimulateCombGiven(g, file, 1, 0)
	if s == nil || !s.NonconstantOutput {
		t.Fatal("failed to detect nonzero output")
	}
	if got := s.NodeWords(output.NodeID())[0] & 15; got != 6 {
		t.Fatalf("output patterns got %04b, want 0110", got)
	}
	if err := os.WriteFile(file, []byte("01x1"), 0600); err != nil {
		t.Fatal(err)
	}
	if Fra_SmlSimulateCombGiven(g, file, 0, 0) != nil {
		t.Fatal("accepted invalid input")
	}
}

func TestFraSmlCheckOutputSavesWitness(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	out := g.AddOutput(x)
	s := Fra_SmlStart(g, 0, 1, 1)
	s.FrameWords(x.NodeID(), 0)[0] = 2
	s.SimulateOne()
	m := &Fra_Man_t{PManAig: g, PManFraig: g, PSml: s}
	if Fra_SmlCheckOutput(m) != 1 {
		t.Fatal("failed to find output witness")
	}
	model, ok := g.Data.([]int32)
	if !ok || len(model) != 2 || model[0] != 1 || model[1] != int32(out.NodeID()) {
		t.Fatalf("wrong witness: %v", g.Data)
	}
}

func TestFraSmlSavePatternOneClearsInitialRegisters(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 1)
	m := &Fra_Man_t{PManAig: g, NFramesAll: 3, PPatWords: make([]uint32, 1)}
	Fra_SmlSavePattern1(m, 1)
	if m.PPatWords[0] != ^uint32(8) {
		t.Fatalf("initial latch bit not cleared: %08x", m.PPatWords[0])
	}
}
