package glucose2

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestGlucose2CNFFileSatisfiability(t *testing.T) {
	for _, test := range []struct {
		name, input string
		want        int32
	}{
		{"sat", "p cnf 2 2\n1 2 0\n-1 2 0\n", 10},
		{"unsat", "p cnf 1 2\n1 0\n-1 0\n", 20},
	} {
		t.Run(test.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "formula.cnf")
			if err := os.WriteFile(filename, []byte(test.input), 0o600); err != nil {
				t.Fatal(err)
			}
			pars := Glucose_CreatePars(0, 0, 0, 0)
			status, model, err := Glucose2_SolveCnf(filename, &pars)
			if err != nil {
				t.Fatal(err)
			}
			if status != test.want {
				t.Fatalf("status=%d, want %d", status, test.want)
			}
			if status == 10 && len(model) != 2 {
				t.Fatalf("SAT model has %d variables, want 2: %v", len(model), model)
			}
		})
	}
}

func TestGlucose2SolveAIGAllTwoInputFunctions(t *testing.T) {
	for table := 0; table < 16; table++ {
		g := aig.NewGraph()
		x, y := g.NewInput(), g.NewInput()
		root := g.Const0()
		for bits := 0; bits < 4; bits++ {
			if table&(1<<bits) == 0 {
				continue
			}
			cube := g.Const1()
			for i, input := range []aig.Edge{x, y} {
				if bits&(1<<i) == 0 {
					input = input.Not()
				}
				cube = g.AND(cube, input)
			}
			root = g.OR(root, cube)
		}
		g.AddOutput(root)
		pars := Glucose_CreatePars(1, 0, 0, 0)
		status, model := Glucose2_SolveAig(g, &pars)
		if (status == 10) != (table != 0) {
			t.Fatalf("truth table %04b: status=%d, want satisfiable=%v", table, status, table != 0)
		}
		if status == 10 {
			inputs := []bool{model[0] != 0, model[1] != 0}
			outputs, err := g.Evaluate(inputs)
			if err != nil || len(outputs) != 1 || !outputs[0] {
				t.Fatalf("truth table %04b returned invalid model %v, outputs=%v, err=%v", table, model, outputs, err)
			}
		}
	}
}

func TestGlucose2AIGSolverInstallsCircuitGateMetadata(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	and := g.AND(y, x.Not())
	xor := g.NewXor(x, y)
	g.AddOutput(and)
	g.AddOutput(xor)
	s := NewABCSolver()
	s.JustificationMode = 2
	Glucose2_SolverFromAig(g, s)
	if !s.isTwoFanin(Var(and.NodeID())) || !s.isAND(Var(and.NodeID())) {
		t.Fatal("AND gate metadata was not installed in ascending fanin order")
	}
	if !s.isTwoFanin(Var(xor.NodeID())) || s.isAND(Var(xor.NodeID())) {
		t.Fatal("XOR gate metadata was not installed in descending fanin order")
	}
	if got := s.SolveLimited(nil, false); got != LTrue {
		t.Fatalf("SAT result = %v, want SAT", got)
	}
}

func TestGlucose2CompatibilityAssumptionMinimization(t *testing.T) {
	s := NewABCSolver()
	s.EnsureABCVariables(3)
	if !s.AddABCClause([]int32{abcLit(0, true), abcLit(1, true)}) {
		t.Fatal("could not add base clause")
	}
	assumptions := []int32{abcLit(0, false), abcLit(1, false), abcLit(2, false)}
	count := bmcg2_sat_solver_minimize_assumptions(s, assumptions, int32(len(assumptions)), 0)
	if count <= 0 || count >= int32(len(assumptions)) {
		t.Fatalf("minimized count=%d, assumptions=%v", count, assumptions)
	}
	if got := s.SolveABCLiterals(assumptions[:count]); got != GLUCOSE_UNSAT {
		t.Fatalf("minimized core status=%d, want UNSAT; core=%v", got, assumptions[:count])
	}
	for i := 0; i < int(count); i++ {
		trial := append([]int32(nil), assumptions[:count]...)
		trial = append(trial[:i], trial[i+1:]...)
		if got := s.SolveABCLiterals(trial); got != GLUCOSE_SAT {
			t.Fatalf("core is not minimal after removing %d: status=%d core=%v", i, got, assumptions[:count])
		}
	}
}

func TestGlucose2CommandRegistrationAndCNFOutput(t *testing.T) {
	frame := abcmain.NewFrame()
	var out, stderr bytes.Buffer
	frame.Out, frame.Err = &out, &stderr
	Glucose2_Init(frame)
	command, ok := frame.LookupCommand("&glucose2")
	if !ok {
		t.Fatal("&glucose2 was not registered")
	}
	filename := filepath.Join(t.TempDir(), "formula.cnf")
	if err := os.WriteFile(filename, []byte("p cnf 1 1\n1 0\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if status := command.Handler(frame, []string{"&glucose2", filename}); status != 0 {
		t.Fatalf("command status=%d, stderr=%q", status, stderr.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("SATISFIABLE")) {
		t.Fatalf("command output = %q", out.String())
	}
}

func TestGlucose2CommandUsesAIGResourceAndStoresCounterexample(t *testing.T) {
	frame := abcmain.NewFrame()
	var out, stderr bytes.Buffer
	frame.Out, frame.Err = &out, &stderr
	graph := aig.NewGraph()
	x := graph.NewInput()
	graph.AddOutput(x)
	frame.SetResource("pGia", graph)
	Glucose2_Init(frame)
	command, ok := frame.LookupCommand("&glucose2")
	if !ok {
		t.Fatal("&glucose2 was not registered")
	}
	if status := command.Handler(frame, []string{"&glucose2"}); status != 0 {
		t.Fatalf("command status=%d, stderr=%q", status, stderr.String())
	}
	if frame.Cex == nil || frame.Cex.PIs != 1 || len(frame.Cex.Bits) != 1 || frame.Cex.Bits[0]&1 == 0 {
		t.Fatalf("unexpected counterexample: %+v", frame.Cex)
	}
	if !bytes.Contains(out.Bytes(), []byte("SATISFIABLE")) {
		t.Fatalf("command output = %q", out.String())
	}
}
