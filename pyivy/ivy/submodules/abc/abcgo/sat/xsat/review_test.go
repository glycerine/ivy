package xsat

import "testing"

func TestInvalidStoredClauseCannotPartiallyMutateSolver(t *testing.T) {
	s := New()
	s.AddVariable(1)
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected singleton storage rejection")
			}
		}()
		s.newClause([]int{0}, false)
	}()
	if len(s.memory.clauses) != 0 || len(s.clauses) != 0 || len(s.watches.data[1].data) != 0 {
		t.Fatal("failed allocation left a partial clause", s.memory, s.clauses)
	}
}
func TestMissingWatchRemovalIsDetected(t *testing.T) {
	p := &xSAT_WatchList_t{data: []xSAT_Watcher_t{{CRef: 3}}}
	defer func() {
		if recover() == nil {
			t.Error("missing watch removal silently accepted")
		}
	}()
	xSAT_WatchListRemove(p, 9)
}
func TestInvalidLiteralCannotGrowVariables(t *testing.T) {
	s := New()
	func() {
		defer func() {
			if recover() == nil {
				t.Error("expected literal rejection")
			}
		}()
		s.AddClause([]int{100, -1})
	}()
	if s.VariableCount() != 0 || len(s.memory.clauses) != 0 {
		t.Fatal("invalid clause changed solver")
	}
}

func TestDestroyAndReleasePathsAreIdempotent(t *testing.T) {
	s := New()
	s.AddClause([]int{0, 2})
	s.Solve()
	xSAT_SolverDestroy(s)
	xSAT_SolverDestroy(s)
	if s.memory != nil || s.watches != nil || s.binary != nil || s.activity != nil || s.model != nil {
		t.Fatal("destroy retained owner references")
	}
	p := xSAT_MemAlloc(1)
	ref := xSAT_MemAppend(p, 4)
	p.clauses[ref].lits = append(p.clauses[ref].lits, 0, 2)
	xSAT_MemRestart(p)
	if len(p.clauses) != 0 {
		t.Fatal("restart retained live clauses")
	}
	xSAT_MemFree(p)
	xSAT_MemFree(p)
	if p.clauses != nil {
		t.Fatal("memory free retained slice")
	}
	w := xSAT_VecWatchListAlloc(1)
	xSAT_VecWatchListPush(w)
	xSAT_WatchListPush(w.data[0], xSAT_Watcher_t{CRef: 1})
	xSAT_VecWatchListFree(w)
	xSAT_VecWatchListFree(w)
	if w.data != nil {
		t.Fatal("watch free retained storage")
	}
}
func TestPropagationBudgetCanResume(t *testing.T) {
	s := New()
	s.AddClause([]int{0, 2})
	s.Config.PropagationLimit = 1
	if s.Solve() != LBoolUndef || s.model != nil {
		t.Fatal("propagation budget ignored")
	}
	s.Config.PropagationLimit = 0
	if s.Solve() != LBoolTrue {
		t.Fatal("budget corrupted resumed solver")
	}
}
