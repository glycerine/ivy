package aig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestDOTLabelsUseSourceObjectIDs(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	out := g.AddOutput(g.AND(a, b))
	data, err := aigDOT(g, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range []Edge{a, b, out} {
		label := regexp.MustCompile(fmt.Sprintf(`(?m)^\s*Node%d \[.*label\s*=\s*"%d"`, e.NodeID(), e.NodeID()))
		if !label.Match(data) {
			t.Errorf("object %d lacks its source ID label:\n%s", e.NodeID(), data)
		}
	}
}

func TestDOTOmitsUnusedConstant(t *testing.T) {
	g := NewGraph()
	g.AddOutput(g.NewInput())
	data, err := aigDOT(g, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Node0") {
		t.Error("unused constant appears in visualization")
	}
}

func TestShowReplacesSourceExtensionAndKeepsDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("version.2", 0755); err != nil {
		t.Fatal(err)
	}
	g := NewGraph()
	g.AddOutput(g.NewInput())
	g.Name = filepath.Join("version.2", "circuit.blif")
	got, err := Aig_ManShow(g, 0, nil)
	want := filepath.Join("version.2", "circuit.dot")
	if err != nil || got != want {
		t.Fatalf("Show = %q (%v), want %q", got, err, want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Fatal(err)
	}
}

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
	for _, want := range []string{"digraph AIG", "style = filled", "style = dotted", "shape = triangle", "shape = invtriangle", `label="test \"name\""`} {
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

func TestDOTPreservesAllNodeScratch(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	root := g.AND(a, b.Not())
	out := g.AddOutput(root)
	g.node(out).Level = 99 // CO cached levels are not their display rank.
	for i := range g.nodes {
		g.nodes[i].MarkA = i%2 != 0
		g.nodes[i].MarkB = i%3 != 0
		g.nodes[i].TravID = int32(17 + i)
	}
	before := append([]Node(nil), g.nodes...)
	if _, err := aigDOT(g, []Edge{root, out, a}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, g.nodes) || g.travID != 1 {
		t.Fatal("DOT writing changed node caches, marks or traversal IDs")
	}
}

func TestDOTLimitAndWriteFailuresPreserveExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "graph.dot")
	if err := os.WriteFile(path, []byte("existing"), 0600); err != nil {
		t.Fatal(err)
	}
	g := NewGraph()
	root := g.NewInput()
	for i := 0; i < 200; i++ {
		root = g.AND(root, g.NewInput())
	}
	g.AddOutput(root)
	if _, err := aigDOT(g, nil); err != nil {
		t.Fatalf("200 logic nodes should be allowed: %v", err)
	}
	if err := Aig_WriteDotAig(g, path, 0, []Edge{InvalidEdge}); err == nil {
		t.Fatal("invalid highlight accepted")
	}
	g.AND(root, g.NewInput()) // A dangling gate still counts toward the limit.
	if err := Aig_WriteDotAig(g, path, 0, nil); err == nil {
		t.Fatal("201 logic nodes accepted")
	}
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "existing" {
		t.Fatalf("failed preflight overwrote existing file: %q (%v)", data, err)
	}
	small := NewGraph()
	small.AddOutput(small.NewInput())
	if err := Aig_WriteDotAig(small, filepath.Dir(path), 0, nil); err == nil {
		t.Fatal("writing to a directory succeeded")
	}
}

func TestDOTExtensionsParseWithGraphviz(t *testing.T) {
	dot, err := exec.LookPath("dot")
	if err != nil {
		t.Skip("Graphviz unavailable")
	}
	g := NewGraph()
	g.Name = "quoted \"name\"\nUnicode: café \\ path"
	a, b := g.NewInput(), g.NewInput()
	buffer := Aig_ObjCreate(g, &Node{Kind: Buffer, Fanins: [2]Edge{a, InvalidEdge}})
	xor := g.NewXor(buffer, b.Not())
	g.AddOutput(xor)
	// A higher dangling cone must have explicit labels/ranks, rather than the
	// implicit nodes produced by the source writer beyond its output depth.
	dangling := g.AND(xor, a)
	dangling = g.AND(dangling, b)
	data, err := aigDOT(g, []Edge{xor, dangling.Not()})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"shape = doublecircle", "shape = diamond", fmt.Sprintf("Node%d [label", dangling.NodeID())} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("missing %q", want)
		}
	}
	command := exec.Command(dot, "-Tdot")
	command.Stdin = strings.NewReader(string(data))
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("Graphviz rejected DOT: %v: %s", err, output)
	}
}
