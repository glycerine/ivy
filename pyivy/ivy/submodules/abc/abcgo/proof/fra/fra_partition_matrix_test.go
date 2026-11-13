package fra

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// TestPartitionMatrixAgainstCompiledABC builds the checked-in C sources in
// an isolated directory when FRA_BUILD_ORACLE=1; no checkout objects are used.
func TestPartitionMatrixAgainstCompiledABC(t *testing.T) {
	if os.Getenv("FRA_BUILD_ORACLE") != "1" {
		t.Skip("set FRA_BUILD_ORACLE=1 to build the C oracle from source")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	if _, err = exec.LookPath("g++"); err != nil {
		t.Skip("C++ compiler unavailable")
	}
	if _, err = exec.LookPath("make"); err != nil {
		t.Skip("make unavailable")
	}
	src, err := filepath.Abs("../../../src")
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Dir(src)
	dir := t.TempDir()
	build := filepath.Join(dir, "abc")
	if err = os.MkdirAll(build, 0755); err != nil {
		t.Fatal(err)
	}
	copyFile := func(from, to string) error {
		in, e := os.Open(from)
		if e != nil {
			return e
		}
		defer in.Close()
		out, e := os.Create(to)
		if e != nil {
			return e
		}
		_, e = io.Copy(out, in)
		closeErr := out.Close()
		if e != nil {
			return e
		}
		return closeErr
	}
	for _, name := range []string{"Makefile", "arch_flags.c", "depends.sh"} {
		if err = copyFile(filepath.Join(root, name), filepath.Join(build, name)); err != nil {
			t.Fatal(err)
		}
	}
	if err = os.Chmod(filepath.Join(build, "depends.sh"), 0755); err != nil {
		t.Fatal(err)
	}
	err = filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, e := filepath.Rel(src, path)
		if e != nil {
			return e
		}
		target := filepath.Join(build, "src", rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if !d.Type().IsRegular() || strings.HasSuffix(path, ".o") || strings.HasSuffix(path, ".d") {
			return nil
		}
		return copyFile(path, target)
	})
	if err != nil {
		t.Fatal(err)
	}
	buildCmd := exec.Command("make", "-j4", "ABC_MAKE_NO_DEPS=1", "ABC_USE_STDINT_H=1", "ABC_USE_NO_READLINE=1", "ABC_USE_NO_PTHREADS=1", "OPTFLAGS=-O0", "libabc.a")
	buildCmd.Dir = build
	if out, e := buildCmd.CombinedOutput(); e != nil {
		t.Fatalf("build checked-in ABC sources: %v\n%s", e, out)
	}
	src = filepath.Join(build, "src")
	archive := filepath.Join(build, "libabc.a")
	exe := filepath.Join(dir, "oracle")
	args := []string{"-DABC_USE_STDINT_H=1", "-DABC_USE_CUDD=1", "-I", src, "test_vectors/partition_matrix_oracle.c", archive, "-lstdc++", "-lm", "-ldl", "-lz", "-o", exe}
	if out, e := exec.Command(cc, args...).CombinedOutput(); e != nil {
		t.Fatalf("link C oracle: %v: %s", e, out)
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
