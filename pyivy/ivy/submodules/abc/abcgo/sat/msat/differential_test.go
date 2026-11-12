package msat

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestOriginalCDifferential(t *testing.T) {
	if os.Getenv("MSAT_C_DIFFERENTIAL") != "1" {
		t.Skip("set MSAT_C_DIFFERENTIAL=1 to compile original C oracle")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	binary := filepath.Join(tmp, "original-msat")
	args := []string{"-O2", "-DABC_USE_STDINT_H", "-I" + filepath.Join(root, "src"), "test_vectors/original_driver.c"}
	for _, name := range []string{"Activity", "Clause", "ClauseVec", "Mem", "OrderH", "Queue", "Read", "SolverApi", "SolverCore", "SolverIo", "SolverSearch", "Sort", "Vec"} {
		args = append(args, filepath.Join(root, "src/sat/msat/msat"+name+".c"))
	}
	args = append(args, "-lm", "-o", binary)
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("C oracle build: %v\n%s", err, out)
	}
	// The source itself documents a tautology-normalization bug. Use proper
	// 3-SAT clauses (distinct variables) for parity; unrestricted clauses are
	// independently checked against exhaustive enumeration in solver_test.go.
	rng := rand.New(rand.NewSource(98473))
	for run := 0; run < 300; run++ {
		n := 4 + rng.Intn(11)
		count := 1 + rng.Intn(5*n)
		p := New(n)
		var data strings.Builder
		fmt.Fprintf(&data, "p cnf %d %d\n", n, count)
		for c := 0; c < count; c++ {
			variables := rng.Perm(n)[:3]
			var lits []int
			for _, v := range variables {
				sign := rng.Intn(2)
				lit := 2*v + sign
				lits = append(lits, lit)
				value := v + 1
				if sign != 0 {
					value = -value
				}
				fmt.Fprintf(&data, "%d ", value)
			}
			data.WriteString("0\n")
			p.AddClause(lits)
		}
		path := filepath.Join(tmp, "formula.cnf")
		if err := os.WriteFile(path, []byte(data.String()), 0600); err != nil {
			t.Fatal(err)
		}
		for trial := 0; trial < 3; trial++ {
			variables := rng.Perm(n)[:rng.Intn(4)]
			as := make([]int, len(variables))
			args := []string{path}
			for i, v := range variables {
				as[i] = 2*v + rng.Intn(2)
				args = append(args, strconv.Itoa(as[i]))
			}
			out, err := exec.Command(binary, args...).CombinedOutput()
			if err != nil {
				t.Fatalf("original C: %v %s", err, out)
			}
			text := string(out)
			at := strings.LastIndex(text, "RESULT=")
			if at < 0 {
				t.Fatal(text)
			}
			want, err := strconv.Atoi(strings.TrimSpace(text[at+7:]))
			if err != nil {
				t.Fatal(err, text)
			}
			if got := p.Solve(as, 0, 0); int(got) != want {
				t.Fatalf("case %d assumptions %v got %d C %d\n%s", run, as, got, want, data.String())
			}
		}
	}
}
