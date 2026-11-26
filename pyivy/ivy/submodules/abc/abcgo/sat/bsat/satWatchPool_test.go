package bsat

import "testing"

// A solver built from recycled watch lists must start with every list
// empty, exactly like a fresh one.
func TestRecycledWatchListsStartEmpty(t *testing.T) {
	s := sat_solver_new()
	sat_solver_setnvars(s, 50)
	for i := int32(0); i < 49; i++ {
		sat_solver_addclause(s, []lit{toLitCond(i, 0), toLitCond(i+1, 1)})
	}
	sat_solver_delete(s)
	r := sat_solver_new()
	sat_solver_setnvars(r, 60)
	for i := range r.wlists {
		if r.wlists[i].size != 0 {
			t.Fatal("recycled watch list not empty", i)
		}
	}
	sat_solver_addclause(r, []lit{toLitCond(0, 0)})
	if sat_solver_solve(r, nil, 0, 0, 0, 0) != l_True {
		t.Fatal("recycled solver failed")
	}
}
