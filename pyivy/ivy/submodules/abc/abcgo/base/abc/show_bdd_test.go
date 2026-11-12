package abc

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestBDDVisualization(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_LOGIC, ABC_FUNC_BDD, 1)
	in := n.NewObject(ObjectPI)
	in.SetName("source")
	node := n.NewObject(ObjectNode)
	node.SetName("function")
	Abc_ObjAddFanin(node, in)
	dd := n.FunctionManager.(*core.DdManager)
	node.Data = core.Cudd_Not(dd.Vars[0])
	for _, complemented := range []bool{false, true} {
		var out bytes.Buffer
		if err := WriteNodeBDDVisualization(&out, node, complemented); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(out.String(), `label="source"`) || !strings.Contains(out.String(), `label="function"`) {
			t.Fatalf("missing names: %s", out.String())
		}
	}
}

func TestNetworkBDDVisualization(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	a.SetName("a")
	b.SetName("b")
	root := n.AndEdges(a.Signal().Not(), b.Signal())
	out := n.NewObject(ObjectPO)
	out.SetName("result")
	Abc_ObjAddFanin(out, root.Not())
	for _, complemented := range []bool{false, true} {
		var output bytes.Buffer
		if err := WriteNetworkBDDVisualization(&output, n, complemented); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{`label="a"`, `label="b"`, `label="result"`} {
			if !strings.Contains(output.String(), name) {
				t.Fatalf("missing %s in %s", name, output.String())
			}
		}
	}
}
