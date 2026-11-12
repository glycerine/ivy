package ivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIvyWriteDotRestoresMarksAndWritesEdges(t *testing.T) {
	p := Ivy_ManStart()
	defer Ivy_ManStop(p)
	a, b := Ivy_ObjCreatePi(p), Ivy_ObjCreatePi(p)
	root := Ivy_And(p, a, Ivy_Not(b))
	Ivy_ObjCreatePo(p, root)

	bold := Vec_PtrAlloc(1)
	defer Vec_PtrFree(bold)
	Vec_PtrPush(bold, root)
	filename := filepath.Join(t.TempDir(), "graph.dot")
	Ivy_WriteDotAig(p, filename, 0, bold)
	if root.fMarkB != 0 {
		t.Fatal("writer did not restore the prior bold mark")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	for _, expected := range []string{"digraph AIG", "Node3 [label = \"3\", shape = ellipse, style = filled]", "Node3 -> Node1", "style = dotted", "Node3 -> Node2"} {
		if !strings.Contains(out, expected) {
			t.Errorf("DOT output missing %q", expected)
		}
	}
}
