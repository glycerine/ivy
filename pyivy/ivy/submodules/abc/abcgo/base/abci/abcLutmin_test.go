package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/bdd/cudd/core"
)

func evalLutTestBDD(dd *core.DdManager, edge *core.DdNode, assignment []bool) bool {
	if core.Cudd_IsConstant(edge) {
		return edge == dd.One
	}
	regular := core.Cudd_Regular(edge)
	branch := regular.E
	if assignment[regular.Index] {
		branch = regular.T
	}
	if core.Cudd_IsComplement(edge) {
		branch = core.Cudd_Not(branch)
	}
	return evalLutTestBDD(dd, branch, assignment)
}

func TestAbcNtkBddMux21TruthTable(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD, 1)
	dd := n.FunctionManager.(*core.DdManager)
	pins := []*abc.Object{abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)}
	obj := Abc_NtkBddMux21(n, pins)
	if obj == nil {
		t.Fatal("mux construction returned nil")
	}
	function := obj.Data.(*core.DdNode)
	for code := 0; code < 8; code++ {
		values := []bool{code&1 != 0, code&2 != 0, code&4 != 0}
		want := values[0] && values[2] || !values[0] && values[1]
		if got := evalLutTestBDD(dd, function, values); got != want {
			t.Fatalf("mux(%03b) = %v, want %v", code, got, want)
		}
	}
}

func TestAbcNtkBddCofactorsUsesSourceLevelOrder(t *testing.T) {
	dd := core.NewBDDManager(2)
	x := core.Cudd_bddIthVar(dd, 0)
	y := core.Cudd_bddIthVar(dd, 1)
	f := core.Cudd_bddXor(dd, x, y)
	cofactors := Abc_NtkBddCofactors(dd, f, 1)
	if len(cofactors) != 2 {
		t.Fatalf("cofactor count = %d, want 2", len(cofactors))
	}
	if cofactors[0] != y || cofactors[1] != core.Cudd_Not(y) {
		t.Fatalf("first-level cofactors = (%p,%p), want (%p,%p)", cofactors[0], cofactors[1], y, core.Cudd_Not(y))
	}
}

func TestAbcNtkBddCurtisReconstructsBoundColumns(t *testing.T) {
	old := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD, 1)
	oldDD := old.FunctionManager.(*core.DdManager)
	inputs := make([]*abc.Object, 4)
	for i := range inputs {
		inputs[i] = abc.Abc_NtkCreatePi(old)
	}
	vars := []*core.DdNode{core.Cudd_bddIthVar(oldDD, 0), core.Cudd_bddIthVar(oldDD, 1), core.Cudd_bddIthVar(oldDD, 2), core.Cudd_bddIthVar(oldDD, 3)}
	boundParity := core.Cudd_bddXor(oldDD, core.Cudd_bddXor(oldDD, vars[0], vars[1]), vars[2])
	function := core.Cudd_bddIte(oldDD, boundParity, core.Cudd_Not(vars[3]), vars[3])
	oldNode := abc.Abc_NtkCreateNode(old)
	oldNode.Data = function
	for _, input := range inputs {
		abc.Abc_ObjAddFanin(oldNode, input)
	}

	cofactors := Abc_NtkBddCofactors(oldDD, function, 3)
	unique := lutUniqueCofactors(cofactors)
	if len(unique) != 2 {
		t.Fatalf("unique Curtis columns = %d, want 2", len(unique))
	}
	dst := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_BDD, 1)
	for _, input := range inputs {
		input.Copy = abc.Abc_NtkCreatePi(dst)
	}
	result := Abc_NtkBddCurtis(dst, oldNode, cofactors, unique)
	if result == nil || result.Data == nil {
		t.Fatal("Curtis decomposition returned no composition node")
	}

	for code := 0; code < 8; code++ {
		bound := []bool{code&1 != 0, code&2 != 0, code&4 != 0}
		column := -1
		for i, candidate := range unique {
			if candidate == cofactors[code] {
				column = i
				break
			}
		}
		if column < 0 {
			t.Fatalf("cofactor %d is absent from unique columns", code)
		}
		for bit := 0; bit < 1; bit++ {
			want := column&(1<<uint(bit)) != 0
			if got := evalLutTestBDD(dst.FunctionManager.(*core.DdManager), result.Fanins()[1].Object.Data.(*core.DdNode), append(bound, false)); got != want {
				t.Fatalf("bound-set code bit at row %03b = %v, want %v", code, got, want)
			}
		}
		for _, free := range []bool{false, true} {
			assignment := make([]bool, len(dst.FunctionManager.(*core.DdManager).Vars))
			assignment[0], assignment[1] = free, column&1 != 0
			got := evalLutTestBDD(dst.FunctionManager.(*core.DdManager), result.Data.(*core.DdNode), assignment)
			want := evalLutTestBDD(oldDD, function, []bool{bound[0], bound[1], bound[2], free})
			if got != want {
				t.Fatalf("composition row %03b free=%v = %v, want %v", code, free, got, want)
			}
		}
	}
}
