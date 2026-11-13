package minisatp

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"testing"
)

// Set MINISATP_ORACLE to an executable built from the official 1.14p archive
// to run the differential test. The regular test suite needs no C++ compiler.
func TestAgainstOriginalMiniSat(t *testing.T) {
	oracle := os.Getenv("MINISATP_ORACLE")
	if oracle == "" {
		t.Skip("MINISATP_ORACLE not set")
	}
	rng := rand.New(rand.NewSource(1142006))
	for trial := 0; trial < 100; trial++ {
		n := 1 + rng.Intn(7)
		m := 1 + rng.Intn(35)
		var input bytes.Buffer
		fmt.Fprintf(&input, "p cnf %d %d\n", n, m)
		s := newSolver(n)
		for i := 0; i < m; i++ {
			clause := make([]int32, 1+rng.Intn(5))
			for j := range clause {
				lit := int32(rng.Intn(2 * n))
				clause[j] = lit
				dimacs := int(lit>>1) + 1
				if lit&1 != 0 {
					dimacs = -dimacs
				}
				fmt.Fprintf(&input, "%d ", dimacs)
			}
			fmt.Fprintln(&input, "0")
			s.addClause(clause)
		}
		command := exec.Command(oracle)
		command.Stdin = bytes.NewReader(input.Bytes())
		out, err := command.CombinedOutput()
		exit, ok := err.(*exec.ExitError)
		if !ok || (exit.ExitCode() != 10 && exit.ExitCode() != 20) {
			t.Fatalf("oracle failed: %v\n%s", err, out)
		}
		if got, want := s.solve(), exit.ExitCode() == 10; (got == 1) != want {
			t.Fatalf("trial %d: Go=%d original=%d\n%s", trial, got, exit.ExitCode(), input.String())
		}
	}
}
