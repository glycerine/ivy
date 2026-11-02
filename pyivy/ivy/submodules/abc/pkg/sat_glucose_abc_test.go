package abc

import "testing"

// TestGlucoseBmcgAPI exercises the bmcg_sat_solver_* wrappers: clauses, assumptions,
// the final conflict, assumption minimisation, elimination and the callback.
func TestGlucoseBmcgAPI(t *testing.T) {
	s := bmcg_sat_solver_start()
	defer bmcg_sat_solver_stop(s)
	bmcg_sat_solver_set_nvars(s, 4)
	// a(0) -> b(1), b -> c(2), c -> d(3)  as clauses over literals 2*v / 2*v+1 (+1 = negated)
	for _, c := range [][]int{{1, 2}, {3, 4}, {5, 6}} { // (!a|b), (!b|c), (!c|d)
		bmcg_sat_solver_addclause(s, c, 2)
	}
	called := 0
	bmcg_sat_solver_setcallback(s, "man", func(p any, msg int, cex []int) int {
		called++
		if p.(string) != "man" {
			t.Errorf("callback got %v", p)
		}
		return 0
	})
	if r := bmcg_sat_solver_solve(s, []int{0}, 1); r != GlucoseSat { // assume a
		t.Fatalf("solve(a) = %d", r)
	}
	if bmcg_sat_solver_read_cex_varvalue(s, 3) != 1 {
		t.Fatalf("d must be true when a is assumed")
	}
	if called != 1 {
		t.Fatalf("callback called %d times", called)
	}
	// assume a and !d: unsatisfiable; the final conflict is a subset of the assumptions
	assum := []int{0, 7, 2} // a, !d, b (b is redundant for the contradiction)
	if r := bmcg_sat_solver_solve(s, assum, 3); r != GlucoseUnsat {
		t.Fatalf("solve(a,!d,b) = %d", r)
	}
	var arr []int
	n := bmcg_sat_solver_final(s, &arr)
	if n != len(arr) || n == 0 {
		t.Fatalf("final conflict %v (%d)", arr, n)
	}
	// minimise the assumptions {b, a, !d}: pivot 0
	m := []int{2, 0, 7}
	k := bmcg_sat_solver_minimize_assumptions(s, m, 3, 0)
	if k < 1 || k > 3 {
		t.Fatalf("minimize_assumptions kept %d", k)
	}
	if r := bmcg_sat_solver_solve(s, m, k); r != GlucoseUnsat {
		t.Fatalf("minimised assumptions %v are not unsat", m[:k])
	}
	if bmcg_sat_solver_varnum(s) != 4 || bmcg_sat_solver_clausenum(s) < 3 {
		t.Fatalf("counts: %d vars %d clauses", bmcg_sat_solver_varnum(s), bmcg_sat_solver_clausenum(s))
	}
	// AND gate encoding: x4 = x0 & x1
	v := bmcg_sat_solver_addvar(s)
	if bmcg_sat_solver_add_and(s, v, 0, 1, 0, 0, 0) != 1 {
		t.Fatal("add_and failed")
	}
	if r := bmcg_sat_solver_solve(s, []int{2 * v, 1}, 2); r != GlucoseUnsat { // x4 & !x0
		t.Fatalf("x4 & !x0 = %d", r)
	}
}

func TestGlucose2BmcgAPI(t *testing.T) {
	s := bmcg2_sat_solver_start()
	defer bmcg2_sat_solver_stop(s)
	bmcg2_sat_solver_set_nvars(s, 3)
	bmcg2_sat_solver_addclause(s, []int{1, 2}, 2)
	bmcg2_sat_solver_addclause(s, []int{3, 4}, 2)
	if r := bmcg2_sat_solver_solve(s, []int{0}, 1); r != Glucose2Sat {
		t.Fatalf("solve = %d", r)
	}
	if r := bmcg2_sat_solver_solve(s, []int{0, 5}, 2); r != Glucose2Unsat {
		t.Fatalf("solve(a, !c) = %d", r)
	}
}
