package sim

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestSourceRandMatchesLibc(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("source rand oracle currently targets glibc")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	program := "#include <stdio.h>\n#include <stdlib.h>\nint main(void){srand(0xABC); for(int i=0;i<32;i++) printf(\"%u\\n\", (unsigned)rand()); return 0;}\n"
	source := t.TempDir() + "/rand.c"
	binary := t.TempDir() + "/rand"
	if err := os.WriteFile(source, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(cc, source, "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile libc oracle: %v\n%s", err, out)
	}
	out, err := exec.Command(binary).Output()
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Fields(string(out))
	rng := newSourceRand(0xABC)
	for i, line := range lines {
		var want uint32
		if _, err := fmt.Sscan(line, &want); err != nil {
			t.Fatal(err)
		}
		if got := rng.Uint31(); got != want {
			t.Fatalf("rand[%d] = %d, want libc %d", i, got, want)
		}
	}
}

func TestSourceFunctionalSupportExhaustiveSmallFunctions(t *testing.T) {
	for truth := uint32(0); truth < 256; truth++ {
		g := aig.NewGraph()
		x := [3]aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
		// Build the truth table as a sum of minterms using the graph's AIG nodes.
		var terms []aig.Edge
		for assignment := 0; assignment < 8; assignment++ {
			if truth&(1<<uint(assignment)) == 0 {
				continue
			}
			term := aig.Edge(g.Const1())
			for v := 0; v < 3; v++ {
				lit := x[v]
				if assignment&(1<<uint(v)) == 0 {
					lit = lit.Not()
				}
				term = g.AND(term, lit)
			}
			terms = append(terms, term)
		}
		root := aig.Edge(g.Const0())
		for _, term := range terms {
			root = g.AND(root.Not(), term.Not()).Not()
		}
		g.AddOutput(root)
		support := Sim_ComputeFunSuppSource(g, false)[0]
		for variable := 0; variable < 3; variable++ {
			want := false
			for assignment := 0; assignment < 8; assignment++ {
				if assignment&(1<<uint(variable)) != 0 {
					continue
				}
				if (truth>>uint(assignment))&1 != (truth>>uint(assignment|1<<uint(variable)))&1 {
					want = true
				}
			}
			if got := Sim_HasBit(support, variable); got != want {
				t.Fatalf("truth=%08b variable=%d: got %v want %v", truth, variable, got, want)
			}
		}
	}
}

func TestSourceSolverCallbackSATLifecycle(t *testing.T) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 20)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	root := g.Const1()
	for _, input := range inputs {
		root = g.AND(root, input)
	}
	g.AddOutput(root)

	satCalls := 0
	solver := func(input, output int) (bool, []uint32) {
		sat, model := cofactorDifference(g, output, input, -1)
		if sat {
			satCalls++
		}
		return sat, model
	}
	got := Sim_ComputeFunSuppSourceWithSolver(g, false, solver)[0]
	if satCalls == 0 {
		t.Fatal("source target flow never requested a SAT witness")
	}
	for i := range inputs {
		if !Sim_HasBit(got, i) {
			t.Fatalf("input %d missing from functional support", i)
		}
	}
}

func TestSourceSolverCallbackUNSATLifecycle(t *testing.T) {
	g := aig.NewGraph()
	inputs := make([]aig.Edge, 3)
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	// (x1 | x2) & (x1 | !x2) is x1, while retaining x2 in
	// the structural support so the target reaches the SAT callback.
	left := g.OR(inputs[1], inputs[2])
	right := g.OR(inputs[1], inputs[2].Not())
	g.AddOutput(g.AND(inputs[0], g.AND(left, right)))

	unsatCalls := 0
	solver := func(input, output int) (bool, []uint32) {
		sat, model := cofactorDifference(g, output, input, -1)
		if !sat {
			unsatCalls++
		}
		return sat, model
	}
	got := Sim_ComputeFunSuppSourceWithSolver(g, false, solver)[0]
	if unsatCalls == 0 {
		t.Fatal("source target flow never removed an UNSAT target")
	}
	if !Sim_HasBit(got, 0) || !Sim_HasBit(got, 1) || Sim_HasBit(got, 2) {
		t.Fatalf("unexpected support bits: %032b", got)
	}
}
