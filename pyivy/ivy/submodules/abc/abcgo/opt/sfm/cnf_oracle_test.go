package sfm

import (
	"bytes"
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

// cnfSatisfiedExactly reports whether the CNF of an nVars-input node holds
// exactly on the assignments where the output (variable nVars) equals the
// truth table.
func cnfSatisfiedExactly(cnf []int8, words []uint64, nVars int) error {
	for a := 0; a < 1<<uint(nVars+1); a++ {
		satisfied, clauseOK := true, false
		for _, e := range cnf {
			if e == -1 {
				satisfied = satisfied && clauseOK
				clauseOK = false
				continue
			}
			value := a&(1<<uint(e>>1)) != 0
			clauseOK = clauseOK || value != (e&1 != 0)
		}
		wanted := ((a>>uint(nVars))&1 != 0) == bit(words, a&((1<<uint(nVars))-1))
		if satisfied != wanted {
			return fmt.Errorf("CNF %v of %d-input function %x: assignment %b satisfied=%v want %v", cnf, nVars, words, a, satisfied, wanted)
		}
	}
	return nil
}

// A node whose function is constant but which still has fanins (as
// Sfm_NodeResubOne produces) must get a CNF that fixes its output, not its
// first fanin (CBUGS_FIXED.md, Sfm_TruthToCnf).
func TestTruthToCnfConstantWithFanins(t *testing.T) {
	for nVars := 0; nVars <= 8; nVars++ {
		for _, c := range []uint64{0, ^uint64(0)} {
			words := make([]uint64, truthWords(nVars))
			for i := range words {
				words[i] = c
			}
			var cover []uint32
			var cnf []int8
			Sfm_TruthToCnf(words[0], words, nVars, &cover, &cnf)
			if err := cnfSatisfiedExactly(cnf, words, nVars); err != nil {
				t.Error(err)
			}
		}
	}
}

// TestTruthToCnfMatchesC compares Sfm_TruthToCnf with the C function
// compiled from src/opt/sfm/sfmCnf.c on random and constant functions.
func TestTruthToCnfMatchesC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section collection")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("source oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	binary := filepath.Join(t.TempDir(), "cnf_oracle")
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", filepath.Join(root, "src"), filepath.Join(root, "test_vectors/sfm/cnf_oracle.c"), "-o", binary)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile original C: %v\n%s", err, out)
	}
	rng := rand.New(rand.NewSource(77))
	var input bytes.Buffer
	var expected []string
	add := func(nVars int, words []uint64) {
		fmt.Fprintf(&input, "%d", nVars)
		for _, w := range words {
			fmt.Fprintf(&input, " %x", w)
		}
		fmt.Fprintln(&input)
		var cover []uint32
		var cnf []int8
		n := Sfm_TruthToCnf(words[0], words, nVars, &cover, &cnf)
		var line strings.Builder
		fmt.Fprintf(&line, "%d", n)
		for _, e := range cnf {
			fmt.Fprintf(&line, " %d", e)
		}
		expected = append(expected, line.String())
		if err := cnfSatisfiedExactly(cnf, words, nVars); err != nil {
			t.Error(err)
		}
	}
	for nVars := 0; nVars <= 8; nVars++ {
		for _, c := range []uint64{0, ^uint64(0)} {
			words := make([]uint64, truthWords(nVars))
			for i := range words {
				words[i] = c
			}
			add(nVars, words)
		}
		if nVars == 0 {
			continue
		}
		for trial := 0; trial < 30; trial++ {
			words := make([]uint64, truthWords(nVars))
			for i := range words {
				words[i] = rng.Uint64()
			}
			if nVars < 6 {
				words[0] = util.Abc_Tt6Stretch(words[0], int32(nVars))
			}
			add(nVars, words)
		}
	}
	cmd = exec.Command(binary)
	cmd.Stdin = &input
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("source oracle: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
	if len(lines) != len(expected) {
		t.Fatalf("cases %d want %d", len(lines), len(expected))
	}
	for i, line := range lines {
		if line != expected[i] {
			t.Errorf("case %d\nC  %s\nGo %s", i, line, expected[i])
		}
	}
}
