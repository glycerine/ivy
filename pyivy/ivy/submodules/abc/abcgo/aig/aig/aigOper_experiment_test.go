package aig

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestMuxExperimentWritesSourceArtifact(t *testing.T) {
	t.Chdir(t.TempDir())
	out := captureAigOutput(t, Aig_MuxTest)
	if strings.Count(out, "Number of nodes =") != 2 {
		t.Errorf("missing source experiment statistics: %q", out)
	}
	data, err := os.ReadFile("test1.blif")
	if err != nil {
		t.Fatal("missing source MUX experiment circuit:", err)
	}
	if !strings.Contains(string(data), ".outputs") || !strings.Contains(string(data), ".end") {
		t.Fatal("incomplete experiment circuit")
	}
}

func TestMuxExperimentMatchesOriginalC(t *testing.T) {
	want := runAigSourceFixtureOptions(t, "mux_experiment_compare.c", "", 10*time.Second, []string{"-Wl,--wrap=Aig_ManDumpBlif"}, "aigMan.c", "aigObj.c", "aigOper.c", "aigTable.c", "aigMem.c", "aigFanout.c", "aigUtil.c", "../../misc/tim/timMan.c", "../../misc/mem/mem.c")
	p := muxExperimentGraph()
	var got strings.Builder
	fmt.Fprintf(&got, "Number of nodes = %6d.\n", Aig_ManObjNum(p))
	Aig_ManCleanup(p)
	fmt.Fprintf(&got, "Number of nodes = %6d.\n", Aig_ManObjNum(p))
	// C argument evaluation order can create the two MUX product gates in
	// either order. Compare canonical structure, not sibling allocation IDs.
	lines := strings.Split(strings.TrimSpace(want), "\n")
	if got.String() != strings.Join(lines[:2], "\n")+"\n" {
		t.Fatalf("source object counts differ: Go %q C %q", got.String(), lines[:2])
	}
	table := map[[2]uint64]uint64{}
	next := uint64(21)
	canonical := func(a, b uint64) uint64 {
		if a > b {
			a, b = b, a
		}
		key := [2]uint64{a, b}
		if id, ok := table[key]; ok {
			return id
		}
		id := next << 1
		next++
		table[key] = id
		return id
	}
	cValues, goValues := make([]uint64, len(p.nodes)), make([]uint64, len(p.nodes))
	for i := 1; i <= 20; i++ {
		cValues[i], goValues[i] = uint64(i*2), uint64(i*2)
	}
	var cOutputs, goOutputs []uint64
	for _, line := range lines[2:] {
		var kind string
		var id, a, b int
		if _, err := fmt.Sscanf(line, "%s %d %d %d", &kind, &id, &a, &b); err != nil && kind != "out" {
			t.Fatal(err)
		}
		left := cValues[a/2] ^ uint64(a&1)
		if kind == "and" {
			cValues[id] = canonical(left, cValues[b/2]^uint64(b&1))
		} else {
			cOutputs = append(cOutputs, left)
		}
	}
	for _, e := range p.LiveEdges() {
		n := p.Node(e)
		a, b := n.Fanins[0], n.Fanins[1]
		switch n.Kind {
		case And:
			goValues[e.NodeID()] = canonical(goValues[a.NodeID()]^uint64(a&1), goValues[b.NodeID()]^uint64(b&1))
		case Output:
			goOutputs = append(goOutputs, goValues[a.NodeID()]^uint64(a&1))
		}
	}
	if !slices.Equal(goOutputs, cOutputs) {
		t.Fatal("seeded MUX output structures differ from original C")
	}
	if p.InputCount() != 20 || p.OutputCount() != 2000 {
		t.Fatal("experiment dimensions changed")
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}
