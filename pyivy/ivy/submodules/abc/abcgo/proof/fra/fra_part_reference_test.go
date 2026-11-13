package fra

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func partitionReferenceGraph() *aig.Graph {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	ab := g.AND(a, b)
	g.AddOutput(g.AND(ab, g.OR(a, c)))
	return g
}

func TestPartitionFraigAgainstCompiledABC(t *testing.T) {
	cc, err := exec.LookPath("gcc")
	if err != nil { t.Skip("C compiler unavailable") }
	ar, err := exec.LookPath("ar")
	if err != nil { t.Skip("ar unavailable") }
	src, err := filepath.Abs("../../../src")
	if err != nil { t.Fatal(err) }
	if _, err := os.Stat(filepath.Join(src, "aig/aig/aigPart.o")); err != nil { t.Skip("prebuilt ABC objects unavailable") }
	var objects []string
	err = filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil { return err }
		if !d.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "misc/st/st.o") { objects = append(objects, path) }
		return nil
	})
	if err != nil { t.Fatal(err) }
	dir := t.TempDir()
	st := filepath.Join(dir, "st.o")
	if out, err := exec.Command(cc, "-DABC_USE_STDINT_H", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", st).CombinedOutput(); err != nil { t.Fatalf("compile st: %v: %s", err, out) }
	objects = append(objects, st)
	archive := filepath.Join(dir, "abc.a")
	if out, err := exec.Command(ar, append([]string{"qcs", archive}, objects...)...).CombinedOutput(); err != nil { t.Fatalf("archive: %v: %s", err, out) }
	exe := filepath.Join(dir, "oracle")
	args := []string{"-DABC_USE_STDINT_H", "-I", src, "test_vectors/partition_fraig_oracle.c", archive, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", exe}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil { t.Fatalf("link: %v: %s", err, out) }
	out, err := exec.Command(exe).CombinedOutput()
	if err != nil { t.Fatalf("C oracle: %v: %s", err, out) }
	var wantNodes, wantOutputs int
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "partition-result ") {
			if _, err := fmt.Sscanf(line, "partition-result nodes=%d outputs=%d", &wantNodes, &wantOutputs); err != nil { t.Fatal(err) }
			goto found
		}
	}
	t.Fatalf("missing C result: %s", out)
found:
	g := partitionReferenceGraph()
	got := Aig_ManFraigPartitioned(g, 1, 1000, 0, 0)
	if int(aig.Aig_ManNodeNum(got)) != wantNodes || got.OutputCount() != wantOutputs {
		t.Fatalf("Go nodes/outputs %d/%d; C %d/%d", aig.Aig_ManNodeNum(got), got.OutputCount(), wantNodes, wantOutputs)
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, _ := g.Evaluate(in)
		actual, err := got.Evaluate(in)
		if err != nil || len(actual) != 1 || actual[0] != want[0] { t.Fatalf("inputs %v: %v != %v: %v", in, actual, want, err) }
	}
}
