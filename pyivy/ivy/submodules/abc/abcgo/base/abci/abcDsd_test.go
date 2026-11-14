package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func TestDsdGlobalPreservesTruthTable(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_STRASH, abc.ABC_FUNC_AIG, 1)
	a, b, c := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	a.SetName("a")
	b.SetName("b")
	c.SetName("c")
	and := n.AndEdges(a.Signal(), b.Signal())
	// f = (a & b) | c, expressed with complemented AIG edges.
	root := n.AndEdges(and.Not(), c.Signal().Not()).Not()
	po := n.NewObject(abc.ObjectPO)
	po.SetName("f")
	if err := n.AddFanin(po, root); err != nil {
		t.Fatal(err)
	}
	got := Abc_NtkDsdGlobal(n, 0, 0, 0)
	if got == nil {
		t.Fatal("global DSD conversion failed")
	}
	defer abc.Abc_NtkDelete(got)
	for mask := 0; mask < 8; mask++ {
		inputs := got.Inputs()
		values := map[*abc.Object]bool{inputs[0]: mask&1 != 0, inputs[1]: mask&2 != 0, inputs[2]: mask&4 != 0}
		want := mask&3 == 3 || mask&4 != 0
		value, err := got.Evaluate(got.Outputs()[0].Fanins()[0], values)
		if err != nil {
			t.Fatalf("assignment %03b: %v", mask, err)
		}
		if value != want {
			t.Errorf("assignment %03b: got %v, want %v", mask, value, want)
		}
	}
}

func makeDsdBDDNetwork(t *testing.T) *abc.Network {
	t.Helper()
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD, 1)
	dd := n.FunctionManager.(*core.DdManager)
	inputs := make([]*abc.Object, 3)
	f := core.Cudd_Not(core.Cudd_ReadOne(dd))
	for i := range inputs {
		inputs[i] = n.NewObject(abc.ObjectPI)
		inputs[i].SetName(string(rune('a' + i)))
		v := core.Cudd_bddIthVar(dd, i)
		f = core.Cudd_bddOr(dd, f, v)
	}
	node := n.NewObject(abc.ObjectNode)
	node.Data = f
	for _, input := range inputs {
		if err := n.AddFanin(node, abc.Edge{Object: input}); err != nil {
			t.Fatal(err)
		}
	}
	po := n.NewObject(abc.ObjectPO)
	po.SetName("f")
	if err := n.AddFanin(po, abc.Edge{Object: node}); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestDsdLocalAndSparsifyTruthTables(t *testing.T) {
	n := makeDsdBDDNetwork(t)
	if Abc_NtkDsdLocal(n, 0, 0) == 0 {
		t.Fatal("local DSD failed")
	}
	inputs, output := n.Inputs(), n.Outputs()[0].Fanins()[0]
	for mask := 0; mask < 8; mask++ {
		values := map[*abc.Object]bool{inputs[0]: mask&1 != 0, inputs[1]: mask&2 != 0, inputs[2]: mask&4 != 0}
		got, err := n.Evaluate(output, values)
		if err != nil {
			t.Fatal(err)
		}
		if got != (mask != 0) {
			t.Errorf("local DSD mask %03b = %v", mask, got)
		}
	}
	sparse := Abc_NtkSparsify(makeDsdBDDNetwork(t), 100, 0)
	if sparse == nil {
		t.Fatal("sparsification failed")
	}
	defer abc.Abc_NtkDelete(sparse)
	for mask := 0; mask < 8; mask++ {
		values := map[*abc.Object]bool{sparse.Inputs()[0]: mask&1 != 0, sparse.Inputs()[1]: mask&2 != 0, sparse.Inputs()[2]: mask&4 != 0}
		for i, want := range []bool{mask != 0, mask == 0} {
			got, err := sparse.Evaluate(sparse.Outputs()[i].Fanins()[0], values)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("sparsify output %d mask %03b = %v, want %v", i, mask, got, want)
			}
		}
	}
}
