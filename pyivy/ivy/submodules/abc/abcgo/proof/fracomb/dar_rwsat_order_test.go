package fracomb_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
)

const darRwsatSourceGolden = "initial in=3 and=6 out=1\nn4=2,5\nn5=3,4\nn6=9,11\nn7=7,13\nn8=6,12\nn9=15,17\no10=19\n" +
	"unbalanced in=3 and=6 out=1\nn4=2,5\nn5=3,4\nn6=9,11\nn7=7,13\nn8=6,12\nn9=15,17\no10=19\n" +
	"balanced in=3 and=6 out=1\nn4=2,5\nn5=3,4\nn6=9,11\nn7=7,13\nn8=6,12\nn9=15,17\no10=19\n" +
	"factor_initial in=2 and=3 out=1\nn3=2,4\nn4=2,5\nn5=7,9\no6=11\n" +
	"factor_balanced in=2 and=0 out=1\no3=2\n"

func darRwsatDump(b *strings.Builder, g *aig.Graph, label string) {
	fmt.Fprintf(b, "%s in=%d and=%d out=%d\n", label, g.InputCount(), aig.Aig_ManNodeNum(g), g.OutputCount())
	for _, e := range g.LiveEdges() {
		n := g.Node(e)
		switch n.Kind {
		case aig.And:
			fmt.Fprintf(b, "n%d=%d,%d\n", e.NodeID(), n.Fanins[0], n.Fanins[1])
		case aig.Output:
			fmt.Fprintf(b, "o%d=%d\n", e.NodeID(), n.Fanins[0])
		}
	}
}

// Ordered intermediate operations remove C's unspecified argument-evaluation
// order from the input graph. The golden was produced by ABC's compiled C.
func TestDarRwsatOrderedStructureAgainstSource(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	x0, x1 := g.AND(a, b.Not()), g.AND(a.Not(), b)
	x := g.OR(x0, x1)
	y0, y1 := g.AND(x, c.Not()), g.AND(x.Not(), c)
	g.AddOutput(g.OR(y0, y1))
	var out strings.Builder
	darRwsatDump(&out, g, "initial")
	for _, stage := range []struct {
		label   string
		balance bool
	}{{"unbalanced", false}, {"balanced", true}} {
		result, err := dar.Dar_ManRwsat(g, stage.balance, false)
		if err != nil {
			t.Fatal(err)
		}
		darRwsatDump(&out, result, stage.label)
	}
	g = aig.NewGraph()
	a, b = g.NewInput(), g.NewInput()
	x0, x1 = g.AND(a, b), g.AND(a, b.Not())
	g.AddOutput(g.OR(x0, x1))
	darRwsatDump(&out, g, "factor_initial")
	result, err := dar.Dar_ManRwsat(g, true, false)
	if err != nil {
		t.Fatal(err)
	}
	darRwsatDump(&out, result, "factor_balanced")
	if out.String() != darRwsatSourceGolden {
		t.Fatalf("Go graph order differs from source:\n%s", out.String())
	}

	// Compare with the compiled checked-in C objects when available; the golden
	// keeps the same assertion active on systems without a C toolchain.
	if runtime.GOOS != "linux" {
		t.Log("C oracle skipped: Linux required")
		return
	}
	cc, e1 := exec.LookPath("gcc")
	cxx, e2 := exec.LookPath("g++")
	if e1 != nil || e2 != nil {
		t.Log("C oracle skipped: compiler unavailable")
		return
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	if _, err := os.Stat(filepath.Join(src, "opt/dar/darScript.o")); err != nil {
		t.Log("C oracle skipped: ABC object files unavailable")
		return
	}
	var objects []string
	if err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".o") && path != filepath.Join(src, "base/main/main.o") && path != filepath.Join(src, "misc/st/st.o") {
			objects = append(objects, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	fixture := filepath.Join(filepath.Dir(file), "test_vectors/dar_rwsat_order.c")
	obj := filepath.Join(dir, "oracle.o")
	st := filepath.Join(dir, "st.o")
	bin := filepath.Join(dir, "oracle")
	for _, step := range [][]string{{"-DABC_USE_STDINT_H=1", "-I", src, "-c", fixture, "-o", obj}, {"-DABC_USE_STDINT_H=1", "-I", src, "-c", filepath.Join(src, "misc/st/st.c"), "-o", st}} {
		if result, err := exec.Command(cc, step...).CombinedOutput(); err != nil {
			t.Fatalf("compile oracle: %v\n%s", err, result)
		}
	}
	args := append([]string{"-o", bin, obj, st}, objects...)
	args = append(args, "-lm", "-ldl", "-lpthread", "-lz", "-lreadline")
	if result, err := exec.Command(cxx, args...).CombinedOutput(); err != nil {
		t.Fatalf("link oracle: %v\n%s", err, result)
	}
	actual, err := exec.Command(bin).CombinedOutput()
	if err != nil {
		t.Fatalf("C oracle: %v\n%s", err, actual)
	}
	if string(actual) != darRwsatSourceGolden {
		t.Fatalf("compiled C graph order differs from golden:\n%s", actual)
	}
}
