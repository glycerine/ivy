package aig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Original C Aig_ObjReplace transplants a positive, unused replacement node
// into the old node's object slot and deletes the replacement object.
func TestSourceReplaceMovesUnusedPositiveNode(t *testing.T) {
	g := NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	old := g.AND(a, b)
	replacement := g.AND(a, c)
	g.AddOutput(old)
	Aig_ObjReplace(g, old, replacement, 0)
	n := g.Node(old)
	if n.Kind != And || n.Fanins != [2]Edge{a, c} || n.Refs != 1 || g.ValidEdge(replacement) || g.AND(a, c) != old {
		t.Fatalf("old=%+v replacement-valid=%v lookup=%v", n, g.ValidEdge(replacement), g.AND(a, c))
	}
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestSourceReplaceBranchesOracle(t *testing.T) {
	const golden = "0 kind=and f0=2 f1=6 oldref=1 newpresent=0 lookup_old=1\n1 kind=buffer f0=11 f1=-1 oldref=1 newpresent=1 lookup_old=0\n2 kind=buffer f0=10 f1=-1 oldref=1 newpresent=1 lookup_old=0\n"
	var got strings.Builder
	for mode := 0; mode < 3; mode++ {
		g := NewGraph()
		a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
		old, replacement := g.AND(a, b), g.AND(a, c)
		g.AddOutput(old)
		if mode == 2 {
			g.AddOutput(replacement)
		}
		target := replacement
		if mode == 1 {
			target = target.Not()
		}
		Aig_ObjReplace(g, old, target, 0)
		n := g.Node(old)
		kind := "and"
		f1 := int(n.Fanins[1])
		if n.Kind == Buffer {
			kind = "buffer"
			f1 = -1
		}
		present, lookup := 0, 0
		if g.ValidEdge(replacement) {
			present = 1
		}
		if g.AND(a, c) == old {
			lookup = 1
		}
		fmt.Fprintf(&got, "%d kind=%s f0=%d f1=%d oldref=%d newpresent=%d lookup_old=%d\n", mode, kind, n.Fanins[0], f1, n.Refs, present, lookup)
		if err := g.Check(); err != nil {
			t.Fatal(err)
		}
	}
	if got.String() != golden {
		t.Fatalf("replacement branches differ from C:\n%s", got.String())
	}
	if c, ok := runReplaceCOracle(t, "aig_replace_source.c"); ok && c != golden {
		t.Fatalf("compiled C differs:\n%s", c)
	}
}

func runReplaceCOracle(t *testing.T, fixtureName string) (string, bool) {
	t.Helper()
	// Compare with the compiled checked-in C objects when available; the golden
	// keeps the same assertion active on systems without a C toolchain.
	if runtime.GOOS != "linux" {
		t.Log("C oracle skipped: Linux required")
		return "", false
	}
	cc, e1 := exec.LookPath("gcc")
	cxx, e2 := exec.LookPath("g++")
	if e1 != nil || e2 != nil {
		t.Log("C oracle skipped: compiler unavailable")
		return "", false
	}
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../src"))
	if _, err := os.Stat(filepath.Join(src, "aig/aig/aigObj.o")); err != nil {
		t.Log("C oracle skipped: ABC object files unavailable")
		return "", false
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
	fixture := filepath.Join(filepath.Dir(file), "test_vectors", fixtureName)
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
	return string(actual), true
}
