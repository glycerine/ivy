package mfs

import (
	"bytes"
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOriginalWindowAndDivisorOracle(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("ELF section collection required")
	}
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("source oracle requires cc")
	}
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	binary := filepath.Join(t.TempDir(), "oracle")
	cmd := exec.Command(cc, "-O2", "-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-Wl,--gc-sections", "-I", filepath.Join(root, "src"), filepath.Join(root, "test_vectors/mfs/oracle.c"), "-o", binary)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("compile source oracle: %v\n%s", err, out)
	}
	rng := rand.New(rand.NewSource(932))
	var input bytes.Buffer
	var expected []string
	for trial := 0; trial < 250; trial++ {
		n := NewNetwork()
		for i := 0; i < 5; i++ {
			n.AddInput()
		}
		for i := 0; i < 18; i++ {
			fs := make([]int, 2+rng.Intn(3))
			for k := range fs {
				fs[k] = 1 + rng.Intn(len(n.Nodes)-1)
			}
			n.AddNode(fs, []uint32{rng.Uint32()})
		} // make all gates observable, while retaining nontrivial root cutoffs.
		for id := 6; id < len(n.Nodes); id++ {
			if rng.Intn(4) == 0 || id >= len(n.Nodes)-2 {
				n.AddOutput(Literal(id << 1))
			}
		}
		p, err := initialize(nil, n)
		if err != nil {
			t.Fatal(err)
		}
		target := 6 + rng.Intn(18)
		levels := rng.Intn(4)
		fanouts := rng.Intn(6)
		cap := rng.Intn(16)
		maxlev := rng.Intn(7)
		p.Params.WinMax, p.Params.FanoutsMax = cap, fanouts
		// The original root function intentionally yields no root for a dangling
		// cone. Compare only nodes whose fanout reaches an output (source precondition).
		reachable := make([]bool, len(n.Nodes))
		var mark func(int)
		mark = func(id int) {
			if reachable[id] {
				return
			}
			reachable[id] = true
			for _, f := range n.Nodes[id].Fanins {
				mark(f)
			}
		}
		for _, e := range n.Outputs {
			mark(e.ID())
		}
		if !reachable[target] {
			continue
		}
		fmt.Fprintf(&input, "5 18 %d %d %d %d %d %d", len(n.Outputs), target, levels, fanouts, cap, maxlev)
		for id := 6; id < len(n.Nodes); id++ {
			v := n.Nodes[id]
			fmt.Fprintf(&input, " %d", len(v.Fanins))
			for _, f := range v.Fanins {
				fmt.Fprintf(&input, " %d", f)
			}
		}
		for _, out := range n.Outputs {
			fmt.Fprintf(&input, " %d", out.ID())
		}
		fmt.Fprintln(&input)
		roots := Abc_MfsComputeRoots(p, target, levels, fanouts)
		divs := Abc_MfsComputeDivisors(p, target, maxlev)
		var line strings.Builder
		for _, id := range roots {
			fmt.Fprintf(&line, "%d,", id)
		}
		fmt.Fprint(&line, " ")
		for _, id := range divs {
			fmt.Fprintf(&line, "%d,", id)
		}
		expected = append(expected, strings.TrimSpace(line.String()))
	}
	cmd = exec.Command(binary)
	cmd.Stdin = &input
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("source oracle: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != len(expected) {
		t.Fatalf("got%d sourcecases want%d", len(lines), len(expected))
	}
	for i, line := range lines {
		if strings.TrimSpace(line) != expected[i] {
			t.Fatalf("case%d C%s Go%s", i, line, expected[i])
		}
	}
}
