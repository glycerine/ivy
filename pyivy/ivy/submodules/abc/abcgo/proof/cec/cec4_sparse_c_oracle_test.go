package cec

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Compile the checked-in source fixture against the available ABC objects.
func TestCec4SparseModelAgainstABC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ABC object oracle requires Linux")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	ar, err := exec.LookPath("ar")
	if err != nil {
		t.Skip("ar unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	if _, err = os.Stat(filepath.Join(src, "aig/gia/giaDup.o")); err != nil {
		t.Skip("prebuilt ABC objects unavailable")
	}
	var objects []string
	if err = filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".o") && p != filepath.Join(src, "misc/st/st.o") {
			objects = append(objects, p)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st := filepath.Join(dir, "st.o")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", st).CombinedOutput(); err != nil {
		t.Fatalf("compile st: %v\n%s", err, out)
	}
	archive := filepath.Join(dir, "abc.a")
	if out, err := exec.Command(ar, append([]string{"qcs", archive}, append(objects, st)...)...).CombinedOutput(); err != nil {
		t.Fatalf("archive: %v\n%s", err, out)
	}
	binary := filepath.Join(dir, "oracle")
	fixture := filepath.Join(filepath.Dir(file), "test_vectors/cec4_sparse.c")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, fixture, archive, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("link: %v\n%s", err, out)
	}
	if out, err := exec.Command(binary, "unknown").CombinedOutput(); err != nil || strings.TrimSpace(string(out)) != "unknown=2 candidate=31 marked=1" {
		t.Fatalf("C unknown: %v %s", err, out)
	}
	for _, mode := range []string{"1", "2"} {
		out, err := exec.Command(binary, mode).CombinedOutput()
		if err != nil || strings.TrimSpace(string(out)) != "result=0 c=1 pattern=3,4" {
			t.Fatalf("C mode %s: %v %s", mode, err, out)
		}
	}
	// jType=0 replays reduced CI object IDs directly, even when the
	// original graph has interleaved inputs and logic nodes.
	interleaved := filepath.Join(filepath.Dir(file), "test_vectors/cec4_interleaved.c")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, interleaved, archive, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("link interleaved fixture: %v\n%s", err, out)
	}
	if out, err := exec.Command(binary).CombinedOutput(); err != nil || strings.TrimSpace(string(out)) != "status=0 old_c=1 old_gate=1 pattern=3,4,7" {
		t.Fatalf("C interleaved: %v %s", err, out)
	}

}
