package fra

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecTrivialWitnessAgainstCompiledABC(t *testing.T) {
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	ar, err := exec.LookPath("ar")
	if err != nil {
		t.Skip("ar unavailable")
	}
	src, err := filepath.Abs("../../../src")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(src, "proof/fra/fraSec.o")); err != nil {
		t.Skip("prebuilt ABC objects unavailable")
	}
	var objects []string
	if err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "misc/st/st.o") {
			objects = append(objects, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st, archive := filepath.Join(dir, "st.o"), filepath.Join(dir, "abc.a")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", st).CombinedOutput(); err != nil {
		t.Fatalf("compile st: %v: %s", err, out)
	}
	if out, err := exec.Command(ar, append([]string{"qcs", archive}, append(objects, st)...)...).CombinedOutput(); err != nil {
		t.Fatalf("archive: %v: %s", err, out)
	}
	binary := filepath.Join(dir, "oracle")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "test_vectors/sec_trivial_witness_oracle.c", archive, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("link: %v: %s", err, out)
	}
	if out, err := exec.Command(binary).CombinedOutput(); err != nil || !strings.Contains(string(out), "oracle:0,0,0") {
		t.Fatalf("ABC source result: %v: %s", err, out)
	}
}
