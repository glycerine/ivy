package abci

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteCexLikeNative: after bmc3 finds s27's counterexample (kept as
// pAbc->pCex), write_cex -n writes it with names and write_cex -n -m with
// only the care bits of Bmc_CexCareMinimize. The port found no
// counterexample (it read another store) and had no minimizer registered.
func TestWriteCexLikeNative(t *testing.T) {
	dir := t.TempDir()
	full, minimized := filepath.Join(dir, "full.cex"), filepath.Join(dir, "min.cex")
	runScript(t, "read test_vectors/native_s27_drw.blif; strash; bmc3; write_cex -n "+full+"; write_cex -n -m "+minimized)
	for path, want := range map[string]string{
		full:      "# FALSIFYING OUTPUTS: G17\n# COUNTEREXAMPLE LENGTH: 1\nG5@0=0\nG6@0=0\nG7@0=0\nG0@0=0\nG1@0=0\nG2@0=0\nG3@0=0\n# DONE\n",
		minimized: "# FALSIFYING OUTPUTS: G17\n# COUNTEREXAMPLE LENGTH: 1\nG5@0=0\nG6@0=0\nG7@0=0\nG3@0=0\n# DONE\n",
	} {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Errorf("%s:\n%s\nwant:\n%s", filepath.Base(path), got, want)
		}
	}
}
