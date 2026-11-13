package fracomb_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	dar "github.com/glycerine/abcgo/opt/dar/core"
)

// The fixture records every live node ID, ordered fanin literal and phase
// after each stage of the original C rwsat script, starting from the same AIG.
func TestDarRwsatMuxStagesAgainstSource(t *testing.T) {
	g := aig.NewGraph()
	var inputs [8]aig.Edge
	var outputs [4]aig.Edge
	for i := range inputs {
		inputs[i] = g.NewInput()
	}
	parity := inputs[0]
	for i := 1; i <= 24; i++ {
		left := g.AND(inputs[(i+1)%8], parity)
		right := g.AND(inputs[(i+1)%8].Not(), inputs[(i+2)%8])
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
	golden, err := os.ReadFile(filepath.Join(filepath.Dir(file), "test_vectors/dar_rwsat_mux_stages.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if c, ok := runDarCOracle(t, "dar_rwsat_mux_stages.c"); ok && c != string(golden) {
		t.Fatalf("compiled C rewrite stages differ from golden:\n%s", c)
	}
	if actual.String() != string(golden) {
		got, want := actual.String(), string(golden)
		i := 0
		for i < len(got) && i < len(want) && got[i] == want[i] { i++ }
		t.Fatalf("Go rewrite stages differ from source golden at byte %d (got %q, want %q)", i, got[max(0, i-80):min(len(got), i+120)], want[max(0, i-80):min(len(want), i+120)])
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
}
