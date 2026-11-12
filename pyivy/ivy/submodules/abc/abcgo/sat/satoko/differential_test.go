package satoko

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

// The production package is pure Go. This opt-in audit compiles the original
// repository C sources solely as an independent reference implementation.
func TestOriginalCDifferential(t *testing.T) {
	if os.Getenv("SATOKO_C_DIFFERENTIAL") != "1" {
		t.Skip("set SATOKO_C_DIFFERENTIAL=1 to compare with original C")
	}
	compiler, err := exec.LookPath("cc")
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	binary := filepath.Join(tmp, "original")
	source := filepath.Join(root, "src/sat/satoko")
	args := []string{"-O2", "-DABC_USE_STDINT_H", "-I" + filepath.Join(root, "src"), "test_vectors/original_driver.c", filepath.Join(source, "solver.c"), filepath.Join(source, "solver_api.c"), filepath.Join(source, "cnf_reader.c"), "-lm", "-o", binary}
	if output, err := exec.Command(compiler, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile C reference: %v\n%s", err, output)
	}
	life, err := exec.Command(binary, "--lifecycle").CombinedOutput()
	if err != nil || !strings.Contains(string(life), "LIFE=1,-1,1") {
		t.Fatalf("C lifecycle reference: %v %s", err, life)
	}
	rng := rand.New(rand.NewSource(7489))
	for run := 0; run < 300; run++ {
		n := 2 + rng.Intn(13)
		count := 1 + rng.Intn(60)
		s := New()
		s.SetVariables(n)
		var data strings.Builder
		fmt.Fprintf(&data, "p cnf %d %d\n", n, count)
		for c := 0; c < count; c++ {
			lits := make([]int, 1+rng.Intn(5))
			for i := range lits {
				v := rng.Intn(n)
				negative := rng.Intn(2) == 1
				lits[i] = Literal(v, negative)
				dimacs := v + 1
				if negative {
					dimacs = -dimacs
				}
				fmt.Fprintf(&data, "%d ", dimacs)
			}
			data.WriteString("0\n")
			s.AddClause(lits)
		}
		path := filepath.Join(tmp, "formula.cnf")
		if err := os.WriteFile(path, []byte(data.String()), 0600); err != nil {
			t.Fatal(err)
		}
		for trial := 0; trial < 3; trial++ {
			var assumptions []int
			args := []string{path}
			for k := 0; k < rng.Intn(4); k++ {
				lit := rng.Intn(2 * n)
				assumptions = append(assumptions, lit)
				args = append(args, strconv.Itoa(lit))
			}
			output, err := exec.Command(binary, args...).CombinedOutput()
			if err != nil {
				t.Fatalf("C reference: %v\n%s", err, output)
			}
			text := string(output)
			at := strings.LastIndex(text, "RESULT=")
			if at < 0 {
				t.Fatal(text)
			}
			want, err := strconv.Atoi(strings.TrimSpace(text[at+7:]))
			if err != nil {
				t.Fatal(err, text)
			}
			if got := s.SolveAssumptions(assumptions); got != want {
				t.Fatalf("case %d trial %d got %d C %d assumptions %v\n%s", run, trial, got, want, assumptions, data.String())
			}
		}
	}
}
