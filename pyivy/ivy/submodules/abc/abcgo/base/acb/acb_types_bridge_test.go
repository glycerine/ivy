package acb

import (
	"reflect"
	"testing"
)

func TestCNFBridgeImportsPackedClausesAndLifts(t *testing.T) {
	data := &Cnf_Dat_t{}
	Cnf_DataSetPackedClauses(data, 2, []int32{0, 2}, []int32{0, 3})
	if got := Cnf_DataVarNum(data); got != 2 {
		t.Fatalf("variable count = %d, want 2", got)
	}
	if got := Cnf_DataClauseNum(data); got != 1 {
		t.Fatalf("clause count = %d, want 1", got)
	}
	Cnf_DataLift(data, 2)
	if got, want := Cnf_DataClause(data, 0), []int32{4, 7}; !reflect.DeepEqual(got, want) {
		t.Fatalf("lifted clause = %v, want %v", got, want)
	}
}

func TestTruthPointerBridgeAddressesLaterWords(t *testing.T) {
	words := [2]word{}
	Abc_TtSetBit(&words[0], 70)
	if got := Abc_TtGetBit(&words[0], 70); got != 1 {
		t.Fatalf("bit 70 = %d, want 1", got)
	}
	if words[0] != 0 || words[1] != word(1)<<6 {
		t.Fatalf("word storage = %#x, %#x, want %#x, %#x", words[0], words[1], 0, word(1)<<6)
	}
}

func TestBSATBridgeModelAndFinalConflict(t *testing.T) {
	s := sat_solver_new()
	defer sat_solver_delete(s)
	sat_solver_setnvars(s, 2)
	if !sat_solver_addclause(s, []int32{Abc_Var2Lit(0, 0), Abc_Var2Lit(1, 0)}) {
		t.Fatal("failed to add source-encoded disjunction")
	}
	if got := sat_solver_solve(s, []int32{Abc_Var2Lit(0, 0), Abc_Var2Lit(1, 1)}, nil, 0, 0, 0, 0); got != 1 {
		t.Fatalf("solve status = %d, want SAT", got)
	}
	if got := sat_solver_var_value(s, 1); got != 0 {
		t.Fatalf("model value for variable 1 = %d, want false", got)
	}
	if !sat_solver_addclause(s, []int32{Abc_Var2Lit(0, 0)}) {
		t.Fatal("failed to add positive unit clause")
	}
	if got := sat_solver_solve(s, []int32{Abc_Var2Lit(0, 1)}, nil, 0, 0, 0, 0); got != -1 {
		t.Fatalf("contradictory assumption status = %d, want UNSAT", got)
	}
	var conflict []int32
	if n := sat_solver_final(s, &conflict); n == 0 || len(conflict) != int(n) {
		t.Fatalf("final conflict = %v (count %d), want nonempty matching count", conflict, n)
	}
}

func TestSatokoBridgeClauseSolveAndFinalConflict(t *testing.T) {
	s := satoko_create()
	defer satoko_destroy(s)
	satoko_setnvars(s, 2)
	if satoko_add_clause(s, []int32{Abc_Var2Lit(0, 0), Abc_Var2Lit(1, 0)}) != 1 {
		t.Fatal("failed to add source-encoded disjunction")
	}
	if satoko_add_clause(s, []int32{Abc_Var2Lit(1, 1)}) != 1 {
		t.Fatal("failed to add unit clause")
	}
	if status := satoko_solve_assumptions(s, []int32{Abc_Var2Lit(0, 0)}, 1); status != 1 {
		t.Fatalf("satisfiable assumptions status = %d, want SAT", status)
	}
	if status := satoko_solve_assumptions(s, []int32{Abc_Var2Lit(0, 1)}, 1); status != -1 {
		t.Fatalf("conflicting assumption status = %d, want UNSAT", status)
	}
	var conflict []int32
	if n := satoko_final_conflict(s, &conflict); n == 0 || len(conflict) != int(n) {
		t.Fatalf("final conflict = %v (count %d), want nonempty matching count", conflict, n)
	}
	satoko_setnvars(s, 3)
	assumptions := []int32{Abc_Var2Lit(0, 1), Abc_Var2Lit(2, 0)} // !x and irrelevant z
	if kept := satoko_minimize_assumptions(s, assumptions, 2, 0); kept != 1 || assumptions[0] != Abc_Var2Lit(0, 1) {
		t.Fatalf("minimized assumptions = %v (count %d), want [!x]", assumptions, kept)
	}
}
