package abc

import (
	"bytes"
	"strings"
	"testing"
)

func TestConeComparisonPreservesCopies(t *testing.T) {
	n := NewNetwork("cones")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	root := n.AndEdges(a.Signal(), b.Signal()).Object
	out := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(out, root)
	setCopy(root, a.Signal().Not())
	var result bytes.Buffer
	if err := WriteConeComparison(&result, n); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.String(), "Supp =     2") || !strings.Contains(result.String(), "COs =   1") {
		t.Fatalf("bad cone report: %s", result.String())
	}
	if getCopy(root) != a.Signal().Not() {
		t.Fatal("cone analysis changed copy mapping")
	}
}

func TestInvertConstraintsAndBufferFanouts(t *testing.T) {
	n := NewNetwork("constraints")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	root := n.AndEdges(a.Signal(), b.Signal()).Object
	low := n.NewObject(ObjectPO)
	high := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(low, root)
	Abc_ObjAddFanin(high, root.Signal().Not())
	n.Constraints = 1
	Abc_NtkInvertConstraints(n)
	if low.fanins[0].Complement || high.fanins[0].Complement {
		t.Fatal("constraint inversion changed wrong output")
	}
	root.Level, low.Level, high.Level = 1, 2, 3
	if Abc_NtkAddBuffsEval(low, root) != 0 || Abc_NtkAddBuffsEval2(low, root) != 1 {
		t.Fatal("buffer fanout level checks failed")
	}
}
