package core
import (
	"context"
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"testing"
)

func TestRewritePreservesRandomFunctions(t *testing.T) {
	random := rand.New(rand.NewSource(74))
	for trial := 0; trial < 40; trial++ {
		g := aig.NewGraph()
		edges := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 25; i++ {
			a, b := edges[random.Intn(len(edges))], edges[random.Intn(len(edges))]
			if random.Intn(2) == 0 {
				a = a.Not()
			}
			if random.Intn(2) == 0 {
				b = b.Not()
			}
			edges = append(edges, g.AND(a, b))
		}
		g.AddOutput(edges[len(edges)-1])
		before := make([]bool, 16)
		for n := 0; n < 16; n++ {
			out, err := g.Evaluate([]bool{n&1 != 0, n&2 != 0, n&4 != 0, n&8 != 0})
			if err != nil {
				t.Fatal(err)
			}
			before[n] = out[0]
		}
		p := DefaultRewriteParams()
		p.UseZeros = true
		if err := Dar_ManRewrite(g, &p); err != nil {
			t.Fatal(trial, err)
		}
		for n := 0; n < 16; n++ {
			out, err := g.Evaluate([]bool{n&1 != 0, n&2 != 0, n&4 != 0, n&8 != 0})
			if err != nil || out[0] != before[n] {
				t.Fatalf("trial %d assignment %d: %v %v", trial, n, out, err)
			}
		}
	}
}
func TestRewriteCancellationAndValidation(t *testing.T) {
	g := aig.NewGraph()
	g.AddOutput(g.AND(g.NewInput(), g.NewInput()))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := DefaultRewriteParams()
	p.Context = ctx
	if err := Dar_ManRewrite(g, &p); err != context.Canceled {
		t.Fatal(err)
	}
	p = DefaultRewriteParams()
	p.CutsMax = 0
	if _, err := Dar_ManStart(g, &p); err == nil {
		t.Fatal("accepted invalid limit")
	}
}

func TestRewriteActuallyReducesDistributedLogic(t *testing.T) {
	g := aig.NewGraph()
	a, b, c, d := g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(g.OR(g.AND(a, c), g.AND(b, c)), d))
	before := g.NodeCount()
	m, err := Dar_ManStart(g, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer Dar_ManStop(m)
	if err := m.Rewrite(); err != nil {
		t.Fatal(err)
	}
	if m.Stats.Replacements == 0 || aig.Aig_ManDupDfs(g).NodeCount() >= before {
		t.Fatalf("no reduction: before=%d after=%d stats=%+v", before, g.NodeCount(), m.Stats)
	}
}

func TestRewriteRejectsRootDependentReplacement(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b)
	dependent := g.AND(root, g.NewInput())
	if !graphDependsOn(g, dependent, root) || graphDependsOn(g, a, root) {
		t.Fatal("incorrect root dependency")
	}
}
func TestRewriteTrivialSupport(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.OR(a, g.AND(a, g.OR(b, c))))
	before := g.NodeCount()
	if err := Dar_ManRewrite(g, nil); err != nil {
		t.Fatal(err)
	}
	if aig.Aig_ManDupDfs(g).NodeCount() >= before {
		t.Fatal("support-one simplification missed")
	}
	for n := 0; n < 8; n++ {
		out, err := g.Evaluate([]bool{n&1 != 0, n&2 != 0, n&4 != 0})
		if err != nil || out[0] != (n&1 != 0) {
			t.Fatal(n, out, err)
		}
	}
}
