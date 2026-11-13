package saig

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

// The C entry point's error order depends on whether FRA sees an internal
// node: a node-free unrolling fails its representative assertion first.
func TestTimeframeBoundaryAgainstCompiledC(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("fixture checks Linux process signals")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	ar, err := exec.LookPath("ar")
	if err != nil {
		t.Skip("ar unavailable")
	}
	src, err := filepath.Abs("../../../src")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(src, "aig/saig/saigTrans.o")); err != nil {
		t.Skip("prebuilt ABC objects unavailable")
	}
	var objects []string
	err = filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "misc/st/st.o") {
			objects = append(objects, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	st := filepath.Join(dir, "st.o")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", st).CombinedOutput(); err != nil {
		t.Fatalf("compile st: %v: %s", err, out)
	}
	objects = append(objects, st)
	archive := filepath.Join(dir, "abc.a")
	if out, err := exec.Command(ar, append([]string{"qcs", archive}, objects...)...).CombinedOutput(); err != nil {
		t.Fatalf("archive: %v: %s", err, out)
	}
	exe := filepath.Join(dir, "timeframe-boundary")
	args := []string{"-DABC_USE_STDINT_H", "-I", src, "test_vectors/timeframe_boundary.c", archive, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", exe}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("link: %v: %s", err, out)
	}
	for _, tc := range []struct {
		name    string
		args    []string
		signal  syscall.Signal
		message string
	}{
		{"zero frames", []string{"0", "2", "1", "1"}, syscall.SIGSEGV, ""},
		{"negative frames", []string{"-1", "2", "1", "1"}, syscall.SIGSEGV, ""},
		{"no registers", []string{"1", "2", "0", "0"}, syscall.SIGABRT, "Saig_ManRegNum(pAig) > 0"},
		{"too many frames", []string{"2", "1", "1", "1"}, syscall.SIGABRT, "nFrames <= nFramesMax"},
		{"empty FRAIG before frame limit", []string{"2", "1", "1", "0"}, syscall.SIGABRT, "pFrames->pReprs != NULL"},
		{"valid", []string{"2", "3", "1", "1"}, 0, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := exec.Command(exe, tc.args...).CombinedOutput()
			if tc.signal == 0 {
				if err != nil {
					t.Fatalf("C failed: %v: %s", err, out)
				}
				return
			}
			exit, ok := err.(*exec.ExitError)
			if !ok {
				t.Fatalf("C returned %v, want signal %v: %s", err, tc.signal, out)
			}
			status, ok := exit.Sys().(syscall.WaitStatus)
			if !ok || !status.Signaled() || status.Signal() != tc.signal || !strings.Contains(string(out), tc.message) {
				t.Fatalf("C status %v, want signal %v and %q: %s", status, tc.signal, tc.message, out)
			}
		})
	}
}
