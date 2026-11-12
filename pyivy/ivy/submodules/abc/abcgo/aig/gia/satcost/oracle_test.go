package satcost

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia/mf"
	"github.com/glycerine/abcgo/sat/satoko"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOriginalCConflictCost(t *testing.T) {
	if os.Getenv("SATCOST_C_DIFFERENTIAL") != "1" {
		t.Skip("set SATCOST_C_DIFFERENTIAL=1 for original C conflict-cost audit")
	}
	root, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	compiler, err := exec.LookPath("cc")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "original-cost")
	source := filepath.Join(root, "src/sat/satoko")
	args := []string{"-O2", "-DABC_USE_STDINT_H", "-I" + filepath.Join(root, "src"), "test_vectors/original_cost.c", filepath.Join(source, "solver.c"), filepath.Join(source, "solver_api.c"), filepath.Join(source, "cnf_reader.c"), "-lm", "-o", binary}
	if log, err := exec.Command(compiler, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile original: %v %s", err, log)
	}
	rng := rand.New(rand.NewSource(99827))
	for run := 0; run < 100; run++ {
		n := 4 + rng.Intn(21)
		cnf := &mf.CNF{Variables: n}
		for i := 0; i < 4*n; i++ {
			clause := make([]int32, 3)
			for k := range clause {
				clause[k] = int32(rng.Intn(2 * n))
			}
			cnf.Clauses = append(cnf.Clauses, clause)
		}
		var input strings.Builder
		fmt.Fprintf(&input, "%d %d\n", n, len(cnf.Clauses))
		for _, clause := range cnf.Clauses {
			fmt.Fprintf(&input, "%d", len(clause))
			for _, lit := range clause {
				fmt.Fprintf(&input, " %d", lit)
			}
			input.WriteByte('\n')
		}
		path := filepath.Join(dir, "input.txt")
		if err := os.WriteFile(path, []byte(input.String()), 0600); err != nil {
			t.Fatal(err)
		}
		log, err := exec.Command(binary, path).CombinedOutput()
		if err != nil {
			t.Fatalf("original: %v %s", err, log)
		}
		var status int
		var cost int64
		if _, err := fmt.Sscanf(string(log), "%d %d", &status, &cost); err != nil {
			t.Fatal(err, string(log))
		}
		got, err := RunCNF(cnf, satoko.DefaultOptions())
		if err != nil {
			t.Fatal(err)
		}
		if got.Status != status || got.Conflicts != cost {
			t.Fatalf("run %d: Go status/cost %d/%d, C %d/%d\n%s", run, got.Status, got.Conflicts, status, cost, input.String())
		}
	}
}
