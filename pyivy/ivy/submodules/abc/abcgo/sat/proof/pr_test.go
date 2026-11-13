package proof

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/sat/bsat2"
)

func TestLiteralEncoding(t *testing.T) {
	for variable := int32(0); variable < 100; variable++ {
		for _, negative := range []bool{false, true} {
			l := ToLit(variable)
			if negative {
				l = LitNeg(l)
			}
			printed := LitToDIMACS(l)
			if LitFromDIMACS(printed) != l || LitVar(l) != variable || LitSign(l) != negative {
				t.Fatalf("literal round trip failed: var=%d neg=%v lit=%d printed=%d", variable, negative, l, printed)
			}
		}
	}
	if LitValid(LIT_UNDEF, 4) || LitValid(ToLit(4), 4) || !LitValid(ToLit(3), 4) {
		t.Fatal("literal range check returned unexpected result")
	}
}

func add(t *testing.T, p *Pr_Man_t, root, inA bool, dimacs ...int32) {
	t.Helper()
	lits := make([]Lit, len(dimacs))
	for i, value := range dimacs {
		lits[i] = LitFromDIMACS(value)
	}
	r, a := int32(0), int32(0)
	if root {
		r = 1
	}
	if inA {
		a = 1
	}
	if Pr_ManAddClause(p, lits, r, a) != 1 {
		t.Fatalf("add clause %v: %v", dimacs, p.LastError())
	}
}

func unsatProof(t *testing.T) *Pr_Man_t {
	t.Helper()
	p := Pr_ManAlloc(2)
	// This four-clause core is unsatisfiable but has no root-level units. The
	// propagated variable is the second literal in each implication clause.
	add(t, p, true, false, 1, 2)
	add(t, p, true, false, 1, -2)
	add(t, p, true, false, -1, 2)
	add(t, p, true, false, -1, -2)
	// Under not-y, unit propagation derives both x and not-x. Adding y then
	// triggers the terminal conflict under y; the empty clause is the sentinel.
	add(t, p, false, false, 1)
	add(t, p, false, false)
	return p
}

func TestProofRecorderDerivesResolutionChain(t *testing.T) {
	p := unsatProof(t)
	defer Pr_ManFree(p)
	p.SetProofVerification(true)
	if got := Pr_ManProofWrite(p); got != 1 {
		t.Fatalf("proof failed: %v", p.LastError())
	}
	steps := p.ProofSteps()
	if len(steps) < 2 {
		t.Fatalf("expected resolution steps, got %#v", steps)
	}
	if steps[len(steps)-1].Literals != nil && len(steps[len(steps)-1].Literals) != 0 {
		t.Fatalf("terminal resolution did not derive empty clause: %#v", steps[len(steps)-1])
	}
	if p.pEmpty.pProof != steps[len(steps)-1].ID {
		t.Fatalf("empty clause proof id=%d, final step=%d", p.pEmpty.pProof, steps[len(steps)-1].ID)
	}
	// Independent MiniSAT check: the same root formula is unsatisfiable.
	solver := bsat2.NewSolver()
	x, y := solver.NewVariable(), solver.NewVariable()
	for _, clause := range [][]bsat2.Lit{{bsat2.MakeLit(x, false), bsat2.MakeLit(y, false)}, {bsat2.MakeLit(x, false), bsat2.MakeLit(y, true)}, {bsat2.MakeLit(x, true), bsat2.MakeLit(y, false)}, {bsat2.MakeLit(x, true), bsat2.MakeLit(y, true)}} {
		if !solver.AddClause(clause...) {
			continue
		}
	}
	if got := solver.Solve(); got != bsat2.False {
		t.Fatalf("native SAT cross-check returned %v", got)
	}
}

func TestProofOutputFormatAndRead(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "proof.cnf")
	out := filepath.Join(dir, "proof.steps")
	input := "p 2 6 4\n1 2 0\n-1 2 0\n1 -2 0\n-1 -2 0\n2 0\n0\n"
	if err := os.WriteFile(in, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := ReadProofLog(in)
	if err != nil {
		t.Fatal(err)
	}
	p.SetProofOutput(out)
	if Pr_ManProofWrite(p) != 1 {
		t.Fatalf("write failed: %v", p.LastError())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, line := range []string{"1 1 2 0 0", "2 -1 2 0 0", "5 * 2 1 0", "7 * 6 5 0"} {
		if !strings.Contains(text, line) {
			t.Errorf("proof output missing %q:\n%s", line, text)
		}
	}
	Pr_ManFree(p)
}

func TestRejectsInvalidClauseAndInvalidProof(t *testing.T) {
	p := Pr_ManAlloc(2)
	if Pr_ManAddClause(p, []Lit{ToLit(0), LitNeg(ToLit(0))}, 1, 0) != 0 {
		t.Fatal("accepted repeated variable")
	}
	if p.LastError() == nil {
		t.Fatal("missing repeated-variable error")
	}
	Pr_ManFree(p)

	p = Pr_ManAlloc(2)
	add(t, p, true, false, 1, 2)
	add(t, p, true, false, -1, 2)
	add(t, p, false, false, 1) // not RUP: assumptions not-x do not conflict
	add(t, p, false, false)
	if Pr_ManProofWrite(p) != 0 {
		t.Fatal("accepted a non-RUP learned clause")
	}
	if p.LastError() == nil {
		t.Fatal("missing proof error")
	}
	Pr_ManFree(p)
}

func TestRejectsClausesAfterEmptyWithoutMutation(t *testing.T) {
	p := Pr_ManAlloc(1)
	add(t, p, true, false, 1)
	add(t, p, false, false)
	_, _, before := p.Counts()
	if Pr_ManAddClause(p, nil, 0, 0) != 0 || p.LastError() == nil {
		t.Fatal("accepted a second empty clause")
	}
	if Pr_ManAddClause(p, []Lit{ToLit(0)}, 0, 0) != 0 || p.LastError() == nil {
		t.Fatal("accepted a clause after the terminal empty clause")
	}
	_, _, after := p.Counts()
	if before != after {
		t.Fatalf("rejected additions changed count from %d to %d", before, after)
	}
}

func TestInterpolantPartitionAndBound(t *testing.T) {
	p := Pr_ManAlloc(1)
	add(t, p, true, true, 1)
	add(t, p, true, false, -1)
	add(t, p, false, false)
	Pr_ManPrepareInter(p)
	if p.LastError() != nil {
		t.Fatal(p.LastError())
	}
	if p.nVarsAB != 1 || p.clauses[0].uTruth != 0xAAAAAAAA || p.clauses[1].uTruth != ^uint32(0) {
		t.Fatalf("unexpected partition interpolants: common=%d A=%08x B=%08x", p.nVarsAB, p.clauses[0].uTruth, p.clauses[1].uTruth)
	}
	Pr_ManFree(p)

	p = Pr_ManAlloc(6)
	vars := []int32{1, 2, 3, 4, 5, 6}
	add(t, p, true, true, vars...)
	neg := make([]int32, len(vars))
	for i, v := range vars {
		neg[i] = -v
	}
	add(t, p, true, false, neg...)
	add(t, p, false, false)
	Pr_ManPrepareInter(p)
	if p.LastError() == nil {
		t.Fatal("expected source truth-table bound error")
	}
	Pr_ManFree(p)
}

func TestParserReportsMalformedInput(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "bad.cnf")
	for _, content := range []string{"", "1 2 0\n", "p 2 1 1\n3 0\n", "p 2 2 1\n1 0\n"} {
		if err := os.WriteFile(file, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if p, err := ReadProofLog(file); err == nil {
			Pr_ManFree(p)
			t.Errorf("accepted malformed input %q", content)
		}
	}
}
