package framgr

import (
	"fmt"
	"io"
	"math/bits"
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func TestSavePatternCopiesModelAndSourceCounterexampleOrder(t *testing.T) {
	original := aig.NewGraph()
	for i := 0; i < 3; i++ {
		original.NewInput()
	}
	original.AddOutput(original.Const0())
	original.AddOutput(original.Const0())
	aig.Aig_ManSetRegNum(original, 1)
	frames := aig.NewGraph()
	for i := 0; i < 5; i++ {
		frames.NewInput()
	}
	frames.AddOutput(frames.Const0())
	aig.Aig_ManSetRegNum(frames, 1)
	solver := bsat.NewSolver()
	p := &Fra_Man_t{PManAig: original, PManFraig: frames, PSat: solver, PMemSatNums: make([]int32, len(frames.Nodes())), PPatWords: []uint32{^uint32(0)}, VCex: vec.Vec_IntAlloc(4)}
	want := []bool{true, false, true, false, true}
	for i, input := range frames.Inputs() {
		variable := solver.NewVariable()
		Fra_ObjSetSatNum(p, input, variable)
		literal := bsat.Negative(variable)
		if want[i] {
			literal = bsat.Positive(variable)
		}
		if !solver.AddClause(literal) {
			t.Fatal("model fixture contradiction")
		}
	}
	if !solver.Simplify() || solver.Solve() != bsat.Satisfiable {
		t.Fatal("model fixture unsatisfiable")
	}
	Fra_SmlSavePattern(p)
	if p.PPatWords[0] != 0x15 {
		t.Fatalf("pattern = %#x, want 0x15", p.PPatWords[0])
	}
	got := vec.Vec_IntArray(p.VCex)
	if len(got) != 3 || got[0] != 1 || got[1] != 0 || got[2] != 1 {
		t.Fatalf("counterexample = %v", got)
	}
}

func TestOneHotCoverageUsesSourceRandomSequence(t *testing.T) {
	g := aig.NewGraph()
	for i := 0; i < 2; i++ {
		g.NewInput()
		g.AddOutput(g.Const0())
	}
	aig.Aig_ManSetRegNum(g, 2)
	clauses := vec.Vec_IntAlloc(2)
	vec.Vec_IntPushTwo(clauses, Fra_RegToLit(0, 1), Fra_RegToLit(1, 1))
	const words = 1 << 14
	aig.Aig_ManRandom(1)
	first := make([]uint32, words)
	for i := range first {
		first[i] = Fra_ObjRandomSim()
	}
	covered := 0
	for _, x := range first {
		covered += bits.OnesCount32(x & Fra_ObjRandomSim())
	}
	next := Fra_ObjRandomSim()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	Fra_OneHotEstimateCoverage(&Fra_Man_t{PManAig: g}, clauses)
	os.Stdout = old
	w.Close()
	output, err := io.ReadAll(r)
	r.Close()
	if err != nil {
		t.Fatal(err)
	}
	care := words*32 - covered
	if !strings.Contains(string(output), fmt.Sprintf("Care states ratio = %f. (%d out of %d patterns)", float64(care)/float64(words*32), care, words*32)) {
		t.Fatalf("coverage output = %q", output)
	}
	if got := Fra_ObjRandomSim(); got != next {
		t.Fatalf("next random word = %08x, want %08x", got, next)
	}
}
