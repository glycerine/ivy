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

func TestTimeframeSimplifyUsesFRA(t *testing.T) {
	g := aig.NewGraph()
	input, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(g.XOR(input, state))
	aig.Aig_ManSetRegNum(g, 1)
	got := Saig_ManTimeframeSimplify(g, 2, 3, 1, 0)
	if got == nil {
		t.Fatal("missing timeframe graph")
	}
	if err := got.Check(); err != nil {
		t.Fatal(err)
	}
	for bits := 0; bits < 1<<got.InputCount(); bits++ {
		inputs := make([]bool, got.InputCount())
		for i := range inputs {
			inputs[i] = bits&(1<<i) != 0
		}
		out, err := got.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		state := false
		for i := 0; i < 3; i++ {
			if out[i] != state {
				t.Fatalf("inputs %v, output %v", inputs, out)
			}
			state = state != inputs[i]
		}
	}
}

func TestPartitionFraigAgainstCompiledABC(t *testing.T) {
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
	if _, err := os.Stat(filepath.Join(src, "aig/aig/aigPart.o")); err != nil {
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
	exe := filepath.Join(dir, "oracle")
	args := []string{"-DABC_USE_STDINT_H", "-I", src, "test_vectors/partition_fraig_oracle.c", archive, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", exe}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("link: %v: %s", err, out)
	}
	out, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C oracle: %v: %s", err, out)
	}
	var wantNodes, wantOutputs, choiceNodes, choiceOutputs, choiceCount int
	foundPartition, foundChoice := false, false
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "partition-result ") {
			if _, err := fmt.Sscanf(line, "partition-result nodes=%d outputs=%d", &wantNodes, &wantOutputs); err != nil {
				t.Fatal(err)
			}
			foundPartition = true
		}
		if strings.HasPrefix(line, "choice-result ") {
			if _, err := fmt.Sscanf(line, "choice-result nodes=%d outputs=%d choices=%d", &choiceNodes, &choiceOutputs, &choiceCount); err != nil {
				t.Fatal(err)
			}
			foundChoice = true
		}
	}
	if !foundPartition || !foundChoice {
		t.Fatalf("missing C result: %s", out)
	}
	g := partitionReferenceGraph()
	got := Aig_ManFraigPartitioned(g, 1, 1000, 0, 0)
	if int(aig.Aig_ManNodeNum(got)) != wantNodes || got.OutputCount() != wantOutputs {
		t.Fatalf("Go nodes/outputs %d/%d; C %d/%d", aig.Aig_ManNodeNum(got), got.OutputCount(), wantNodes, wantOutputs)
	}
	for bits := 0; bits < 8; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, _ := g.Evaluate(in)
		actual, err := got.Evaluate(in)
		if err != nil || len(actual) != 1 || actual[0] != want[0] {
			t.Fatalf("inputs %v: %v != %v: %v", in, actual, want, err)
		}
	}
	versions := make([]*aig.Graph, 2)
	for i := range versions {
		g := aig.NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		root := g.AND(g.AND(a, b), c)
		if i != 0 {
			root = g.AND(a, g.AND(b, c))
		}
		g.AddOutput(root)
		versions[i] = g
	}
	choices := Aig_ManChoicePartitioned(versions, 2, 1000, 0, 0)
	if int(aig.Aig_ManNodeNum(choices)) != choiceNodes || choices.OutputCount() != choiceOutputs || int(aig.Aig_ManChoiceNum(choices)) != choiceCount {
		t.Fatalf("Go choice nodes/outputs/choices %d/%d/%d; C %d/%d/%d", aig.Aig_ManNodeNum(choices), choices.OutputCount(), aig.Aig_ManChoiceNum(choices), choiceNodes, choiceOutputs, choiceCount)
	}
}
