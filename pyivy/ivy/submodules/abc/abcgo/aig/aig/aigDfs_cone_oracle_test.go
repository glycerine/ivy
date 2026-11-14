package aig

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestConeMarksAgainstCompiledSource(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("C fixture uses GNU linker section garbage collection")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	src := filepath.Clean(filepath.Join(dir, "../../../src"))
	binary := filepath.Join(t.TempDir(), "cone-marks")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(ctx, cc, "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src, filepath.Join(dir, "test_vectors/cone_marks_compare.c"), "-Wl,--gc-sections", "-o", binary).CombinedOutput(); err != nil {
		t.Fatalf("compile C oracle: %v: %s", err, out)
	}
	want, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("run C oracle: %v: %s", err, want)
	}
	var got strings.Builder
	writeBool := func(value bool) {
		if value {
			got.WriteByte('1')
		} else {
			got.WriteByte('0')
		}
	}
	for mode := 0; mode < 5; mode++ {
		g, a, lower, barrier, root := coneSourceFixture()
		objects := append(g.Inputs(), lower, barrier, root)
		g.copies = make(map[int]Edge)
		for _, e := range objects {
			g.SetObjectData(e, "payload")
			g.copies[e.NodeID()] = e
		}
		g.node(barrier).MarkA = true
		var count int32
		switch mode {
		case 0:
			Aig_ConeMark_rec(g, root)
		case 1:
			Aig_ConeCleanAndMark_rec(g, root)
		case 2:
			count = Aig_ConeCountAndMark_rec(g, root)
		case 3:
			g.node(barrier).MarkA = false
			for _, e := range []Edge{a, lower, root} {
				g.node(e).MarkA = true
			}
			Aig_ConeUnmark_rec(g, root)
		case 4:
			g.node(a).MarkA = true
			count = Aig_DagSize(g, root.Not())
		}
		fmt.Fprintf(&got, "%d %d ", mode, count)
		for _, e := range objects {
			writeBool(g.node(e).MarkA)
		}
		got.WriteByte(' ')
		for _, e := range objects {
			_, copyPresent := g.copies[e.NodeID()]
			dataPresent := g.ObjectData(e) != nil
			if copyPresent != dataPresent {
				t.Fatalf("mode %d object %d: pData representations disagree", mode, e)
			}
			writeBool(dataPresent)
		}
		got.WriteByte('\n')
	}
	if got.String() != string(want) {
		t.Fatalf("Go cone states:\n%sC cone states:\n%s", got.String(), want)
	}
}
