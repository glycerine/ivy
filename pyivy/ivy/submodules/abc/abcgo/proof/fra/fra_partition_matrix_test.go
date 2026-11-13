package fra

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// This is a direct comparison with the source C partition driver and FRA
// engine, using prebuilt ABC objects. It covers multiple output partitions
// with overlapping support, complemented edges and repeated representatives.
func TestPartitionMatrixAgainstCompiledABC(t *testing.T) {
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
	if _, err = os.Stat(filepath.Join(src, "aig/aig/aigPart.o")); err != nil {
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
	args := []string{"-DABC_USE_STDINT_H", "-I", src, "test_vectors/partition_matrix_oracle.c", archive, "-lstdc++", "-lm", "-ldl", "-lpthread", "-lreadline", "-o", exe}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("link: %v: %s", err, out)
	}
	data, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C oracle: %v: %s", err, data)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	cases, choiceCases, sharedRepresentatives := 0, 0, 0
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "choice-matrix ") {
			var size, limit, level, nodes, choices, ci, co int32
			var truth [3]uint32
			_, err := fmt.Sscanf(scanner.Text(), "choice-matrix %d %d %d %d %d %d %d %d %d %d", &size, &limit, &level, &nodes, &choices, &ci, &co, &truth[0], &truth[1], &truth[2])
			if err != nil {
				t.Fatalf("parse %q: %v", scanner.Text(), err)
			}
			versions := make([]*aig.Graph, 2)
			for i := range versions {
				g := aig.NewGraph()
				a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
				root := g.AND(g.AND(a, b), c)
				if i != 0 {
					root = g.AND(a, g.AND(b, c))
				}
				g.AddOutput(root)
				g.AddOutput(root.Not())
				g.AddOutput(g.AND(root, d))
				versions[i] = g
			}
			out := Aig_ManChoicePartitioned(versions, size, limit, level, 0)
			if int32(out.InputCount()) != ci || int32(out.OutputCount()) != co || aig.Aig_ManNodeNum(out) != nodes || aig.Aig_ManChoiceNum(out) != choices {
				t.Fatalf("choice size=%d limit=%d level=%d Go nodes/choices/ci/co=%d/%d/%d/%d C=%d/%d/%d/%d", size, limit, level, aig.Aig_ManNodeNum(out), aig.Aig_ManChoiceNum(out), out.InputCount(), out.OutputCount(), nodes, choices, ci, co)
			}
			for mask := 0; mask < 16; mask++ {
				in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				values, err := out.Evaluate(in)
				if err != nil {
					t.Fatal(err)
				}
				for i, value := range values {
					if value != (truth[i]&(1<<mask) != 0) {
						t.Fatalf("choice size=%d limit=%d level=%d mask=%d output=%d Go=%v Ctruth=%x", size, limit, level, mask, i, value, truth[i])
					}
				}
			}
			choiceCases++
			continue
		}
		if !strings.HasPrefix(scanner.Text(), "matrix ") {
			continue
		}
		var size, limit, level, nodes, co, ci int32
		var repr [4]int
		var truth [4]uint32
		_, err := fmt.Sscanf(scanner.Text(), "matrix %d %d %d %d %d %d %d %d %d %d %d %d %d %d", &size, &limit, &level, &nodes, &repr[0], &repr[1], &repr[2], &repr[3], &co, &ci, &truth[0], &truth[1], &truth[2], &truth[3])
		if err != nil {
			t.Fatalf("parse %q: %v", scanner.Text(), err)
		}
		g := aig.NewGraph()
		a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
		ab, bc := g.AND(a, b), g.AND(b, c)
		r1, r2 := g.AND(ab, c), g.AND(a, bc)
		r3, r4 := g.AND(r1, g.OR(a, d)), g.AND(r2.Not(), d.Not())
		roots := []aig.Edge{r1, r2, r3, r4}
		g.AddOutput(r1)
		g.AddOutput(r3)
		g.AddOutput(r2.Not())
		g.AddOutput(r4)
		out := Aig_ManFraigPartitioned(g, size, limit, level, 0)
		if int32(out.InputCount()) != ci || int32(out.OutputCount()) != co || aig.Aig_ManNodeNum(out) != nodes {
			t.Fatalf("size=%d limit=%d level=%d interface/nodes Go=%d/%d/%d C=%d/%d/%d", size, limit, level, out.InputCount(), out.OutputCount(), aig.Aig_ManNodeNum(out), ci, co, nodes)
		}
		for i := range repr {
			for j := 0; j < i; j++ {
				if repr[i] >= 0 && repr[i] == repr[j] {
					sharedRepresentatives++
				}
			}
		}
		for i, root := range roots {
			r := aig.Aig_ObjFindRepr(g, root)
			got := -1
			if r != aig.InvalidEdge {
				got = r.NodeID()
			}
			if got != repr[i] {
				t.Fatalf("size=%d limit=%d level=%d root=%d Go repr=%d C=%d", size, limit, level, i, got, repr[i])
			}
		}
		for mask := 0; mask < 16; mask++ {
			in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
			values, err := out.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			for i, value := range values {
				if value != (truth[i]&(1<<mask) != 0) {
					t.Fatalf("size=%d limit=%d level=%d mask=%d output=%d Go=%v Ctruth=%x", size, limit, level, mask, i, value, truth[i])
				}
			}
		}
		cases++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if sharedRepresentatives == 0 {
		t.Fatal("fixture did not produce repeated representatives")
	}
	if cases != 18 || choiceCases != 8 {
		t.Fatalf("got %d FRAIG and %d choice cases; oracle output: %s", cases, choiceCases, data)
	}
}
