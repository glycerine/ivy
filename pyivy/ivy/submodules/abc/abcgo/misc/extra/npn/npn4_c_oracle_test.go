package npn

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestTruth4VarNPNMatchesC compiles ABC's Extra_Truth4VarNPN from ../src and
// compares its canon, phase, permutation and class tables with the port's for
// all 65,536 four-input functions. It also requires every function to share
// its canonical representative's class: the original C left the complement of
// each class representative in class 0 (see CBUGS_FIXED.md).
func TestTruth4VarNPNMatchesC(t *testing.T) {
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable for oracle")
	}
	src := filepath.Join("..", "..", "..", "..", "src")
	binary := filepath.Join(t.TempDir(), "npn4")
	args := []string{"-O1", "-DLIN64", "-DABC_USE_STDINT_H", "-I" + src, "-o", binary,
		filepath.Join("test_vectors", "npn4_tables_oracle.c")}
	for _, f := range []string{"misc/extra/extraUtilMisc.c", "misc/extra/extraUtilFile.c", "misc/extra/extraUtilUtil.c", "misc/util/utilFile.c"} {
		args = append(args, filepath.Join(src, f))
	}
	if out, err := exec.Command(cc, append(args, "-lm")...).CombinedOutput(); err != nil {
		t.Fatalf("C oracle build: %v\n%s", err, out)
	}
	out, err := exec.Command(binary).Output()
	if err != nil {
		t.Fatal(err)
	}
	canons, phases, perms, classes := Extra_Truth4VarNPN()
	scanner := bufio.NewScanner(bytes.NewReader(out))
	rows, mismatches := 0, 0
	for scanner.Scan() {
		var truth, canon, phase, perm, class int
		if _, err := fmt.Sscan(scanner.Text(), &truth, &canon, &phase, &perm, &class); err != nil {
			t.Fatal(err)
		}
		rows++
		if class != int(classes[canon]) {
			if mismatches++; mismatches <= 5 {
				t.Errorf("C truth %04x has class %d, but its canon %04x has class %d", truth, class, canon, classes[canon])
			}
		}
		if canon != int(canons[truth]) || phase != int(phases[truth]) || perm != int(perms[truth]) || class != int(classes[truth]) {
			if mismatches++; mismatches <= 5 {
				t.Errorf("truth %04x: C (%04x %d %d %d), Go (%04x %d %d %d)", truth, canon, phase, perm, class, canons[truth], phases[truth], perms[truth], classes[truth])
			}
		}
	}
	if rows != 1<<16 {
		t.Fatalf("oracle printed %d rows", rows)
	}
	if mismatches != 0 {
		t.Fatalf("%d mismatches", mismatches)
	}
}
