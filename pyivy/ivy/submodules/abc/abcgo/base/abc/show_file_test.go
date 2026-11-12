package abc

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShowGetFileName(t *testing.T) {
	if got := Abc_ShowGetFileName("a/b:c é.xyz"); got != "a_b_c___.xyz.dot" {
		t.Fatalf("unexpected filename %q", got)
	}
}

func TestWriteNetworkVisualization(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_SOP, 1)
	n.Name = "visualization"
	input := n.NewObject(ObjectPI)
	input.SetName("input")
	node := Abc_NtkCreateNodeBuf(n, input)
	output := n.NewObject(ObjectPO)
	output.SetName("output")
	Abc_ObjAddFanin(output, node)
	var out bytes.Buffer
	if err := WriteNetworkVisualization(&out, n, false, false, false); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"digraph network", "visualization", "input", "output"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("missing %q from DOT output", want)
		}
	}
	if n.Function != uint8(ABC_FUNC_SOP) || output.fanins[0].Object != node {
		t.Fatal("source network changed")
	}
}

func TestWriteCutVisualization(t *testing.T) {
	n := NewNetwork("cutview")
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	left := n.AndEdges(a.Signal(), b.Signal()).Object
	right := n.AndEdges(a.Signal().Not(), c.Signal()).Object
	root := n.AndEdges(left.Signal(), right.Signal()).Object
	Abc_ObjAddFanin(n.NewObject(ObjectPO), root)
	var out bytes.Buffer
	if err := WriteCutVisualization(&out, root, 2, 4); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "digraph network") || !strings.Contains(out.String(), "style = filled") {
		t.Fatalf("missing highlighted cut in DOT: %s", out.String())
	}
}

func TestShowDotFilePreservesInputOnFailure(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "input with spaces.dot")
	if err := os.WriteFile(filename, []byte("digraph G {}"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := showDotFile(filename, filepath.Join(t.TempDir(), "missing"), "unused"); err == nil {
		t.Fatal("expected a Graphviz execution error")
	}
	if _, err := os.Stat(filename); err != nil {
		t.Fatalf("input was removed after failure: %v", err)
	}
}
