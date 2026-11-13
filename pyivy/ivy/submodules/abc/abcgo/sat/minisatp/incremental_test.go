package minisatp

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// The source permits adding clauses after SAT and solving again. Check every
// incremental prefix against independent exhaustive assignments and replay
// any final contradiction instead of relying on the solver's own model.
func TestIncrementalClausesAgainstExhaustiveOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(114319))
	for trial := 0; trial < 500; trial++ {
		n := 1 + rng.Intn(6)
		s := New(n)
		var formula [][]int32
		for step := 0; step < 20; step++ {
			clause := make([]int32, 1+rng.Intn(4))
			for i := range clause {
				clause[i] = int32(rng.Intn(2 * n))
			}
			formula = append(formula, clause)
			s.AddClause(clause)
			sat := false
			for assignment := 0; assignment < 1<<n; assignment++ {
				all := true
				for _, c := range formula {
					any := false
					for _, lit := range c {
						if (assignment>>uint(lit>>1))&1 != int(lit&1) {
							any = true
							break
						}
					}
					if !any {
						all = false
						break
					}
				}
				if all {
					sat = true
					break
				}
			}
			got := s.Solve()
			if (got == 1) != sat {
				t.Fatalf("trial=%d step=%d formula=%v got=%d wantSAT=%v", trial, step, formula, got, sat)
			}
			if !sat {
				if err := checkProof(s.s); err != nil {
					t.Fatalf("trial=%d step=%d formula=%v invalid proof: %v", trial, step, formula, err)
				}
				if again := s.Solve(); again != 0 {
					t.Fatalf("trial=%d repeated UNSAT gave %d", trial, again)
				}
				break
			}
		}
	}
}

// MINISATP_INCREMENTAL_ORACLE may point to a harness compiled against the
// upstream MiniSat-p sources, allowing a direct incremental comparison.
func TestIncrementalAgainstOriginalMiniSat(t *testing.T) {
	oracle := os.Getenv("MINISATP_INCREMENTAL_ORACLE")
	if oracle == "" {
		t.Skip("MINISATP_INCREMENTAL_ORACLE not set")
	}
	rng := rand.New(rand.NewSource(114320))
	for trial := 0; trial < 100; trial++ {
		n := 1 + rng.Intn(6)
		count := 1 + rng.Intn(20)
		var input, expected bytes.Buffer
		fmt.Fprintf(&input, "%d %d\n", n, count)
		s := New(n)
		stopped := false
		for i := 0; i < count; i++ {
			clause := make([]int32, 1+rng.Intn(4))
			fmt.Fprintf(&input, "%d", len(clause))
			for j := range clause {
				clause[j] = int32(rng.Intn(2 * n))
				fmt.Fprintf(&input, " %d", clause[j])
			}
			fmt.Fprintln(&input)
			if stopped {
				continue
			}
			s.AddClause(clause)
			status := s.Solve()
			fmt.Fprintln(&expected, status)
			stopped = status == 0
		}
		cmd := exec.Command(oracle)
		cmd.Stdin = bytes.NewReader(input.Bytes())
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("trial=%d original failed: %v %s", trial, err, out)
		}
		if strings.TrimSpace(string(out)) != strings.TrimSpace(expected.String()) {
			t.Fatalf("trial=%d input=%s\noriginal=%s\nGo=%s", trial, input.String(), out, expected.String())
		}
	}
}
