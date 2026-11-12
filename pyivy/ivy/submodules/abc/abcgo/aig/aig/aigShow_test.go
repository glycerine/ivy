package aig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDOTPreservesGraphAndMarks(t *testing.T) {
	p := NewGraph()
	p.Name = "test \"name\""
	a, b := p.NewInput(), p.NewInput()
	root := p.AND(a, b.Not())
	out := p.AddOutput(root)
	p.node(a).MarkB = true
	path := filepath.Join(t.TempDir(), "graph.dot")
	if err := Aig_WriteDotAig(p, path, 0, []Edge{root}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"digraph AIG", `style="filled,bold"`, "style=dotted", "shape=triangle", "shape=invtriangle", `label="test \"name\""`} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q:\n%s", want, text)
		}
	}
	if !p.node(a).MarkB || p.node(out).Level != p.node(root).Level {
		t.Fatal("DOT mutated graph")
	}
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
}
