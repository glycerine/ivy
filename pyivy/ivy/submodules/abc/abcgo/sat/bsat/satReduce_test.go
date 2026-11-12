package bsat

import "testing"

func TestCDCLDatabaseReductionUpdatesReferences(t *testing.T) {
	s := sat_solver_new()
	sat_solver_setnvars(s, 4)
	s.nLearntStart = 10
	s.nLearntDelta = 2
	s.nLearntMax = 10
	s.nLearntRatio = 50
	var handles []cla
	for i := 0; i < 8; i++ {
		h := sat_solver_clause_new(s, []lit{toLit(0), toLit(1), toLit(2)}, 1)
		handles = append(handles, h)
		Sat_MemClauseHand(&s.Mem, h).lbd = 7
		s.act_clas.ptr[i] = int32(i * 16)
	}
	locked := handles[0]
	s.reasons[0] = locked
	binary := sat_solver_clause_new(s, []lit{toLit(2), toLit(3)}, 1)
	sat_solver_reducedb(s)
	if s.stats.learnts < 2 || s.act_clas.size != s.Mem.nEntries[1] || s.act_clas.size != int32(s.stats.learnts) {
		t.Fatal(s.stats, s.act_clas.size)
	}
	if Sat_MemClauseHand(&s.Mem, s.reasons[0]).mark != 0 {
		t.Fatal("locked clause removed")
	}
	for _, w := range s.wlists {
		for _, h := range veci_begin(&w) {
			if clause_is_lit(h) == 0 {
				if Sat_MemClauseHand(&s.Mem, h).mark != 0 {
					t.Fatal("stale watch")
				}
			}
		}
	}
	if binary == 0 {
		t.Fatal("missing learned binary")
	}
}
func TestCDCLDatabaseReductionEmptyAndZeroRatio(t *testing.T) {
	s := sat_solver_new()
	sat_solver_setnvars(s, 3)
	sat_solver_reducedb(s)
	s.nLearntRatio = 0
	s.nLearntStart = 1
	s.nLearntDelta = 0
	s.nLearntMax = 1
	sat_solver_clause_new(s, []lit{0, 2, 4}, 1)
	sat_solver_reducedb(s)
	if s.stats.learnts != 0 || s.Mem.nEntries[1] != 0 || s.act_clas.size != 0 {
		t.Fatal("zero selection", s.stats)
	}
}
