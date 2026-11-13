package framgr

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Set ABC_FRA_C_ORACLE=1 to compare the complete sweep with ABC. The six FRA
// translation units under test are freshly compiled; the remaining linked ABC
// objects come from a prebuilt source tree, matching other local C fixtures.
func TestCompleteSweepTimeoutAgainstCompiledC(t *testing.T) {
	if os.Getenv("ABC_FRA_C_ORACLE") != "1" {
		t.Skip("set ABC_FRA_C_ORACLE=1 with prebuilt ABC objects")
	}
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
	var objects []string
	err = filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "misc/st/st.o") {
			objects = append(objects, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) == 0 {
		t.Skip("prebuilt ABC objects unavailable")
	}
	dir := t.TempDir()
	archive := filepath.Join(dir, "abc.a")
	if out, err := exec.Command(ar, append([]string{"qcs", archive}, objects...)...).CombinedOutput(); err != nil {
		t.Fatalf("archive: %v: %s", err, out)
	}
	var fresh []string
	for _, name := range []string{"fraCore", "fraSat", "fraSim", "fraMan", "fraClass", "fraCnf"} {
		target := filepath.Join(dir, name+".o")
		if out, err := exec.Command(cc, "-std=gnu99", "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "proof/fra", name+".c"), "-o", target).CombinedOutput(); err != nil {
			t.Fatalf("compile %s: %v: %s", name, err, out)
		}
		fresh = append(fresh, target)
	}
	st := filepath.Join(dir, "st.o")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", st).CombinedOutput(); err != nil {
		t.Fatalf("compile st: %v: %s", err, out)
	}
	binary := filepath.Join(dir, "oracle")
	args := []string{"-DABC_USE_STDINT_H", "-I", src, "test_vectors/fra_sweep_timeout.c"}
	args = append(args, fresh...)
	args = append(args, st, archive, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", binary)
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("link oracle: %v: %s", err, out)
	}
	out, err := exec.Command(binary).CombinedOutput()
	if err != nil || !strings.Contains(string(out), "callback failures=2 logged=2\ndone nodes=39\ntimeout=35 repr=-1\ntimeout=47 repr=-1") {
		t.Fatalf("source oracle: %v: %s", err, out)
	}
	args[3] = "test_vectors/fra_sweep_timeout_followup.c"
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("link follow-up oracle: %v: %s", err, out)
	}
	out, err = exec.Command(binary).CombinedOutput()
	if err != nil || !strings.Contains(string(out), "failures=2 sat=7 ids=47,59,63,68,69,72,73,76,77\nrare=77 nodes=57") {
		t.Fatalf("follow-up source oracle: %v: %s", err, out)
	}
	for _, id := range []string{"47", "59", "63", "68", "69", "72", "73", "76", "77"} {
		if !strings.Contains(string(out), "timeout="+id+" repr=-1\n") {
			t.Fatalf("source retained candidate %s: %s", id, out)
		}
	}
}

// This counterexample fixture compiles every linked ABC translation unit
// from checked-in C sources; unused functions are discarded by the linker.
func TestMultiwordCounterexampleAgainstFreshC(t *testing.T) {
	if os.Getenv("ABC_FRA_C_ORACLE") != "1" {
		t.Skip("set ABC_FRA_C_ORACLE=1 to compile ABC sources")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	src, err := filepath.Abs("../../../src")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	sources := []string{"proof/fra/fraSim.c", "aig/saig/saigDup.c", "misc/util/utilCex.c", "aig/aig/aigOper.c", "aig/aig/aigMan.c", "aig/aig/aigObj.c", "aig/aig/aigMem.c", "aig/aig/aigUtil.c", "aig/aig/aigFanout.c", "aig/aig/aigTable.c", "misc/tim/timMan.c", "misc/mem/mem.c"}
	var objects []string
	for i, source := range sources {
		target := filepath.Join(dir, strings.ReplaceAll(source, "/", "_")+".o")
		if out, err := exec.Command(cc, "-std=gnu99", "-DABC_USE_STDINT_H", "-I", src, "-ffunction-sections", "-fdata-sections", "-O1", "-c", filepath.Join(src, source), "-o", target).CombinedOutput(); err != nil {
			t.Fatalf("compile source %d: %v: %s", i, err, out)
		}
		objects = append(objects, target)
	}
	binary := filepath.Join(dir, "oracle")
	args := []string{"-std=gnu99", "-DABC_USE_STDINT_H", "-I", src, "-ffunction-sections", "-fdata-sections", "-O1", "-Wl,--gc-sections", "test_vectors/fra_multiword_cex.c", "test_vectors/oracle_link_helpers.c"}
	args = append(args, objects...)
	args = append(args, "-lm", "-lpthread", "-o", binary)
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("link oracle: %v: %s", err, out)
	}
	out, err := exec.Command(binary).CombinedOutput()
	if err != nil || !strings.Contains(string(out), "frame=19 po=0 bits=41 words=00000000,00000060") {
		t.Fatalf("source oracle: %v: %s", err, out)
	}
}
