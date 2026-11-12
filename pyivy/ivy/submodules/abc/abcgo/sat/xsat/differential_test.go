package xsat

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

// The compiler is an optional independent test oracle, not a production dependency.
func TestOriginalCDifferential(t *testing.T) {
	if os.Getenv("XSAT_C_DIFFERENTIAL") != "1" {
		t.Skip("set XSAT_C_DIFFERENTIAL=1 to compare freshly compiled original C")
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
	binary := filepath.Join(tmp, "original-xsat")
	args := []string{"-O2", "-DABC_USE_STDINT_H", "-I" + filepath.Join(root, "src"), "testdata/original_driver.c"}
	for _, file := range []string{"xsatSolver.c", "xsatSolverAPI.c", "xsatCnfReader.c"} {
		args = append(args, filepath.Join(root, "src/sat/xsat", file))
	}
	args = append(args, "-lm", "-o", binary)
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("C oracle build: %v\n%s", err, out)
	}
	rng := rand.New(rand.NewSource(1467))
	for run := 0; run < 600; run++ {
		n := 4 + rng.Intn(9)
		count := 1 + rng.Intn(6*n)
		s := New()
		for v := 0; v < n; v++ {
			s.AddVariable(1)
		}
		var input strings.Builder
		fmt.Fprintf(&input, "p cnf %d %d\n", n, count)
		for c := 0; c < count; c++ {
			widths := []int{2, 3, 3, 3, 4}
			width := widths[rng.Intn(len(widths))]
			vars := rng.Perm(n)[:width]
			lits := make([]int, width)
			for i, v := range vars {
				sign := rng.Intn(2)
				lits[i] = 2*v + sign
				dimacs := v + 1
				if sign != 0 {
					dimacs = -dimacs
				}
				fmt.Fprintf(&input, "%d ", dimacs)
			}
			input.WriteString("0\n")
			s.AddClause(lits)
		}
		path := filepath.Join(tmp, "formula.cnf")
		if err := os.WriteFile(path, []byte(input.String()), 0600); err != nil {
			t.Fatal(err)
		}
		out, err := exec.Command(binary, path).CombinedOutput()
		if err != nil {
			t.Fatalf("C oracle %d: %v\n%s\n%s", run, err, out, input.String())
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
		if got := s.Solve(); got != want {
			t.Fatalf("case %d got %d C %d\n%s", run, got, want, input.String())
		}
	}
}
