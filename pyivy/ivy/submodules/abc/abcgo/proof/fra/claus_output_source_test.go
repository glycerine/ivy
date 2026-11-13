package fra

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/ioa"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

// The accompanying compiled-C fixture writes b_care.aig for /tmp/a-b.aig.
// The separator before the final alphanumeric component is discarded by ABC.
func TestWriteClauseOutputUsesSourceFilenameAndPolarity(t *testing.T) {
	tmp, err := os.CreateTemp(".", "fra-")
	if err != nil {
		t.Fatal(err)
	}
	name := filepath.Base(tmp.Name())
	tmp.Close()
	os.Remove(tmp.Name())
	token := strings.TrimPrefix(name, "fra-")
	want := token + "_care.aig"
	t.Cleanup(func() { os.Remove(want) })
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	g.Name = filepath.Join(t.TempDir(), "a-"+token+".aig")
	p := Fra_ClausAlloc(g, 1, 0, 4, 4, 1, 2, 1, 0, 0, 0, 0)
	p.pCnf = cnf.Cnf_DeriveSimple(g, 0)
	vec.Vec_IntPush(p.vLitsProven, 2*p.pCnf.VarNums[x.NodeID()])
	vec.Vec_IntPush(p.vClausesProven, 1)
	Fra_ClausWriteIndClauses(p)
	model, err := ioa.Ioa_ReadAiger(want, 1)
	if err != nil {
		t.Fatalf("missing source-named output %q: %v", want, err)
	}
	if len(model.Inputs) != 1 || len(model.Outputs) != 1 || model.Outputs[0] != model.Inputs[0] {
		t.Fatalf("unexpected care output: %+v", model)
	}
}

// Set FRA_BUILD_ORACLE=1 to compare with the checked-in C implementation
// linked against the checkout's prebuilt ABC objects.
func TestWriteClauseOutputCompiledC(t *testing.T) {
	if os.Getenv("FRA_BUILD_ORACLE") != "1" {
		t.Skip("set FRA_BUILD_ORACLE=1 to run the compiled-C fixture")
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
	if _, err := os.Stat(filepath.Join(src, "proof/fra/fraClaus.o")); err != nil {
		t.Skip("prebuilt ABC objects unavailable")
	}
	var objects []string
	err = filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "misc/st/st.o") {
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
	exe := filepath.Join(dir, "claus-output")
	fixture, err := filepath.Abs("test_vectors/claus_output_source.c")
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"-DABC_USE_STDINT_H", "-I", src, fixture, archive, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", exe}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("link: %v: %s", err, out)
	}
	cmd := exec.Command(exe)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run: %v: %s", err, out)
	}
	model, err := ioa.Ioa_ReadAiger(filepath.Join(dir, "b_care.aig"), 1)
	if err != nil {
		t.Fatalf("source filename: %v", err)
	}
	if len(model.Inputs) != 1 || len(model.Outputs) != 1 || model.Outputs[0] != model.Inputs[0] {
		t.Fatalf("source care output: %+v", model)
	}
}
