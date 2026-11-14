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

func TestDominatorDriversAgainstCompiledSource(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("fixture uses GNU linker section collection")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	src := filepath.Clean(filepath.Join(dir, "../../../src"))
	binary := filepath.Join(t.TempDir(), "dominators")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	args := []string{"-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src, filepath.Join(dir, "test_vectors/dominators_compare.c")}
	for _, name := range []string{"aigDfs.c", "aigFanout.c", "aigMem.c", "aigUtil.c"} {
		args = append(args, filepath.Join(src, "aig/aig", name))
	}
	args = append(args, "-Wl,--gc-sections", "-o", binary)
	if out, err := exec.CommandContext(ctx, cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile source: %v: %s", err, out)
	}
	out, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err != nil {
		t.Fatalf("source driver: %v: %s", err, out)
	}
	for mode, name := range []string{"nodes", "flops", "pis"} {
		t.Run(name, func(t *testing.T) {
			g := NewGraph()
			a, b := g.NewInput(), g.NewInput()
			g.NewInput()
			lo := g.NewInput()
			x, y, z := g.AND(a, lo), g.AND(b, lo), g.AND(a, b)
			g.AddOutput(z)
			g.AddOutput(x)
			g.AddOutput(y)
			g.nRegs, g.travID = 1, 7
			for i := range g.nodes {
				g.nodes[i].TravID = 5
			}
			Aig_ManFanoutStart(g)
			Aig_ObjRemoveFanout(g, lo, x)
			Aig_ObjAddFanout(g, lo, x)
			s := domCompute(g, 2, name)
			var got strings.Builder
			fmt.Fprintf(&got, "state %d %d %d %d %d\n", g.travID, s.DomNodes, s.Total, s.FilteredSupersets, s.FilteredCandidates)
			for id, n := range g.nodes {
				fmt.Fprintf(&got, "%d@%d:", id, n.TravID)
				for _, d := range s.Dominators[id] {
					got.WriteByte('[')
					for i, v := range d.Nodes {
						if i > 0 {
							got.WriteByte(',')
						}
						fmt.Fprint(&got, v)
					}
					got.WriteByte(']')
				}
				got.WriteByte('\n')
			}
			marker := fmt.Sprintf("BEGIN %d\n", mode)
			_, rest, ok := strings.Cut(string(out), marker)
			if !ok {
				t.Fatalf("missing source marker: %s", out)
			}
			want, _, ok := strings.Cut(rest, "END\n")
			if !ok || got.String() != want {
				t.Fatalf("Go driver:\n%sC driver:\n%s", got.String(), want)
			}
			if name == "flops" {
				Aig_ObjSetMarkA(g, x)
				volume := Aig_ObjDomVolume(s, NewDominator(x.NodeID()))
				Aig_ObjClearMarkA(g, x)
				Aig_ObjDomCollectLos(s)
				marked := 0
				if Aig_ObjIsMarkA(g, lo) {
					marked = 1
				}
				Aig_ObjPoLogicDeref(s)
				refs := domTestRefs(g)
				Aig_ObjPoLogicRef(s)
				Aig_ObjSetMarkA(g, a)
				Aig_ObjDomFindGood(s)
				got.Reset()
				fmt.Fprintf(&got, "volume %d lo-mark %d generation %d\n", volume, marked, g.travID)
				for id, n := range g.nodes {
					mark := 0
					if n.MarkA {
						mark = 1
					}
					fmt.Fprintf(&got, "%d refs %d restored %d mark %d trav %d\n", id, refs[id], n.Refs, mark, n.TravID)
				}
				_, rest, ok := strings.Cut(string(out), "DIAG BEGIN\n")
				if !ok {
					t.Fatal("missing diagnostic source states")
				}
				want, _, ok := strings.Cut(rest, "DIAG END\n")
				if !ok || got.String() != want {
					t.Fatalf("Go diagnostics:\n%sC diagnostics:\n%s", got.String(), want)
				}
			}

		})
	}
}
