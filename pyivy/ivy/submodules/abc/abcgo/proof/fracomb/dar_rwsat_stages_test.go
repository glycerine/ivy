package fracomb_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
)

func dumpDarStage(out *strings.Builder, g *aig.Graph, label string) {
	fmt.Fprintf(out, "%s ci=%d nodes=%d co=%d\n", label, g.InputCount(), aig.Aig_ManNodeNum(g), g.OutputCount())
	phase := func(v bool) int {
		if v {
			return 1
		}
		return 0
	}
	fmt.Fprintf(out, "const=%d\n", phase(g.Node(g.Const1()).Phase))
	for _, edge := range g.Inputs() {
		fmt.Fprintf(out, "i%d=%d\n", edge.NodeID(), phase(g.Node(edge).Phase))
	}
	for _, edge := range g.LiveEdges() {
		n := g.Node(edge)
		if n.Kind == aig.And {
			fmt.Fprintf(out, "n%d=%d,%d,%d\n", edge.NodeID(), n.Fanins[0], n.Fanins[1], phase(n.Phase))
		}
	}
	for _, edge := range g.Outputs() {
		n := g.Node(edge)
		fmt.Fprintf(out, "o%d=%d,%d\n", edge.NodeID(), n.Fanins[0], phase(n.Phase))
	}
}

// The fixture records every live node ID, ordered fanin literal and phase
// after each stage of the original C rwsat script, starting from the same AIG.
func TestDarRwsatLargerOrderedStagesAgainstSource(t *testing.T) {
	g := aig.NewGraph()
	var inputs [8]aig.Edge
	var outputs [4]aig.Edge
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	parity := inputs[0]
	for i := 1; i <= 24; i++ {
		left := g.AND(parity, inputs[i%8].Not())
		right := g.AND(parity.Not(), inputs[i%8])
		parity = g.OR(left, right)
		if i%6 == 0 {
			outputs[i/6-1] = parity
		}
	}
	for _, edge := range outputs {
		g.AddOutput(edge)
	}
	original := g
	var actual strings.Builder
	dumpDarStage(&actual, g, "initial")
	g = aig.Aig_ManDupDfs(g)
	dumpDarStage(&actual, g, "dup")
	var err error
	g, err = dar.Balance(context.Background(), g, false)
	if err != nil {
		t.Fatal(err)
	}
	dumpDarStage(&actual, g, "balance1")
	rw := dar.DefaultRewriteParams()
	rw.UpdateLevel = false
	if err = dar.Dar_ManRewrite(g, &rw); err != nil {
		t.Fatal(err)
	}
	g = aig.Aig_ManDupDfs(g)
	dumpDarStage(&actual, g, "rewrite1")
	ref := dar.DefaultRefactorParams()
	ref.UpdateLevel = false
	if err = dar.Dar_ManRefactor(g, &ref); err != nil {
		t.Fatal(err)
	}
	g = aig.Aig_ManDupDfs(g)
	dumpDarStage(&actual, g, "refactor")
	g, err = dar.Balance(context.Background(), g, false)
	if err != nil {
		t.Fatal(err)
	}
	dumpDarStage(&actual, g, "balance2")
	if err = dar.Dar_ManRewrite(g, &rw); err != nil {
		t.Fatal(err)
	}
	g = aig.Aig_ManDupDfs(g)
	dumpDarStage(&actual, g, "rewrite2")
	_, file, _, _ := runtime.Caller(0)
	golden, err := os.ReadFile(filepath.Join(filepath.Dir(file), "test_vectors/dar_rwsat_stages.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if actual.String() != string(golden) {
		t.Fatalf("Go rewrite stages differ from source golden:\n%s", actual.String())
	}
	direct, err := dar.Dar_ManRwsat(original, true, false)
	if err != nil {
		t.Fatal(err)
	}
	var directDump strings.Builder
	dumpDarStage(&directDump, direct, "rewrite2")
	final := strings.LastIndex(string(golden), "rewrite2 ci=")
	if final < 0 || directDump.String() != string(golden)[final:] {
		t.Fatalf("Dar_ManRwsat output differs from source final stage:\n%s", directDump.String())
	}
	if c, ok := runDarCOracle(t, "dar_rwsat_stages.c"); ok && c != string(golden) {
		t.Fatalf("compiled C rewrite stages differ from golden:\n%s", c)
	}
}
