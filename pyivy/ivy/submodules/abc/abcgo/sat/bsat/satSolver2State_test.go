package bsat

import (
	"math/rand"
	"testing"
)

func TestSolver2StateUnitsAndBacktrack(t *testing.T) {
	s := sat_solver2_new()
	sat_solver2_setnvars(s, 3)
	if s.fProofLogging != 1 || s.hProofLast != -1 || s.Mem.nPageSize != 14 {
		t.Fatal("initial state")
	}
	if solver2_enqueue(s, 0, 0) == 0 {
		t.Fatal("enqueue")
	}
	s.qhead = s.qtail
	if solver2_assume(s, 2) == 0 {
		t.Fatal("assume")
	}
	s.units[1] = 42
	s.qhead = s.qtail
	solver2_canceluntil(s, 0)
	if s.units[1] != 0 || s.assigns[1] != satUnassigned || s.assigns[0] != 0 {
		t.Fatal("cancel state")
	}
	var_set_partA(s, 2, 1)
	sat_solver2_setnvars(s, 4)
	if var_is_partA(s, 2) != 1 || s.units[3] != 0 {
		t.Fatal("grow state")
	}
}

func TestSolver2ProofChainAndClauseIdentity(t *testing.T) {
	s := sat_solver2_new()
	sat_solver2_setnvars(s, 2)
	s.pPrf1 = NewProofGraph()
	first := clause2_create_new(s, []lit{0, 2}, 0, 0)
	clause2_set_id(s, first, 1)
	second := clause2_create_new(s, []lit{1}, 0, 0)
	clause2_set_id(s, second, 2)
	s.units[0] = second
	s.variableA[0] = 1
	proof_chain_start(s, clause2_read(s, first))
	proof_chain_resolve(s, nil, 0)
	proof := proof_chain_stop(s)
	if proof <= 0 {
		t.Fatal("missing proof")
	}
	learned := clause2_create_new(s, []lit{2}, 1, proof)
	if clause2_proofid(s, clause2_read(s, learned), 0) != proof<<2 {
		t.Fatal("proof identity")
	}
	s.activity[0] = 0x7fffffff
	act_var2_bump(s, 0)
	if s.var_inc < 16 || s.activity[0]&0x80000000 != 0 {
		t.Fatal("variable activity rescale")
	}
	s.act_clas.ptr[0] = 0x7fffffff
	act_clause2_bump(s, clause2_read(s, learned))
	if s.cla_inc < 1024 {
		t.Fatal("clause activity rescale")
	}
}

func TestSolver2RootPropagationLogsEmptyProof(t *testing.T) {
	s := sat_solver2_new()
	sat_solver2_setnvars(s, 2)
	s.pPrf1 = NewProofGraph()
	for index, c := range [][]lit{{0, 2}, {1}, {3}} {
		h := clause2_create_new(s, c, 0, 0)
		clause2_set_id(s, h, int32(index+1))
		if len(c) == 1 {
			s.units[lit_var(c[0])] = h
			solver2_enqueue(s, c[0], 0)
		}
	}
	if solver2_propagate(s) == nil || s.hProofLast <= 0 {
		t.Fatal("missing root refutation")
	}
	if s.trail_lim.size != 0 {
		t.Fatal("root propagation created decisions")
	}
}

func TestSolver2FirstUIPHasResolutionProof(t *testing.T) {
	s := sat_solver2_new()
	s.pPrf1 = NewProofGraph()
	sat_solver2_addclause(s, []lit{1, 2}, 1)
	sat_solver2_addclause(s, []lit{1, 3}, 2)
	solver2_assume(s, 0)
	conflict := solver2_propagate(s)
	if conflict == nil {
		t.Fatal("expected conflict")
	}
	var learned veci
	veci_new(&learned)
	proof := solver2_analyze(s, conflict, &learned)
	if learned.size != 1 || learned.ptr[0] != 1 || proof <= 0 || s.tagged.size != 0 {
		t.Fatal(veci_begin(&learned), proof)
	}
	core := Proof_DeriveCore(s.pPrf1, proof)
	if len(core) != 2 {
		t.Fatal(core)
	}
}

func TestSolver2CDCLAgainstEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(7742))
	for trial := 0; trial < 200; trial++ {
		s := sat_solver2_new()
		s.pPrf1 = NewProofGraph()
		formula := [][]lit{}
		for count := 0; count < 24; count++ {
			omitted := rng.Intn(4)
			c := []lit{}
			for v := 0; v < 4; v++ {
				if v != omitted {
					c = append(c, lit(2*v+rng.Intn(2)))
				}
			}
			formula = append(formula, c)
			sat_solver2_addclause(s, c, int32(len(formula)))
		}
		expected := formulaHasModel(formula, nil, 4)
		got := sat_solver2_solve(s, nil, 0, 0, 0, 0)
		if (got == l_True) != expected {
			t.Fatalf("trial %d result %d expected SAT %v", trial, got, expected)
		}
		if !expected {
			if s.hProofLast <= 0 {
				t.Fatal("missing unsat proof")
			}
			ids := Proof_DeriveCore(s.pPrf1, s.hProofLast)
			subset := [][]lit{}
			for _, id := range ids {
				subset = append(subset, formula[id-1])
			}
			if formulaHasModel(subset, nil, 4) {
				t.Fatal("proof core is satisfiable", ids)
			}
		}
	}
}

func TestSolver2ReductionRemapsUnitAndProofHandles(t *testing.T) {
	s := sat_solver2_new()
	sat_solver2_setnvars(s, 3)
	s.pPrf1 = NewProofGraph()
	root := clause2_create_new(s, []lit{0, 2, 4}, 0, 0)
	clause2_set_id(s, root, 1)
	for i := 0; i < 40; i++ {
		proof_chain_start(s, clause2_read(s, root))
		proof := proof_chain_stop(s)
		h := clause2_create_new(s, []lit{0, 2, 4}, 1, proof)
		c := clause2_read(s, h)
		c.lbd = 4
		s.act_clas.ptr[clause_id(c)] = int32(i * 100)
	}
	proof_chain_start(s, clause2_read(s, root))
	proof := proof_chain_stop(s)
	unit := clause2_create_new(s, []lit{0}, 1, proof)
	s.units[0] = unit
	s.nLearntStart, s.nLearntDelta, s.nLearntRatio = 20, 0, 25
	sat_solver2_reducedb(s)
	if s.stats.learnts >= 41 || s.stats.learnts != uint32(s.act_clas.size) || s.claProofs.size != s.act_clas.size {
		t.Fatal(s.stats.learnts, s.act_clas.size, s.claProofs.size)
	}
	c := clause2_read(s, s.units[0])
	if c.size != 1 || c.lits[0] != 0 {
		t.Fatal("unit handle lost")
	}
	for _, h := range veci_begin(&s.claProofs) {
		if core := Proof_DeriveCore(s.pPrf1, h); len(core) != 1 || core[0] != 1 {
			t.Fatal(core)
		}
	}
	for i := range s.wlists {
		for _, h := range veci_begin(&s.wlists[i]) {
			if clause2_read(s, h) == nil {
				t.Fatal("stale watch")
			}
		}
	}
}
func TestSolver2RollbackRestoresClauseAndProofState(t *testing.T) {
	s := sat_solver2_new()
	s.pPrf1 = NewProofGraph()
	sat_solver2_addclause(s, []lit{0, 2}, 1)
	sat_solver2_bookmark(s)
	sat_solver2_addclause(s, []lit{1, 4}, 2)
	sat_solver2_addclause(s, []lit{3, 5}, 3)
	sat_solver2_rollback(s)
	if s.size != 2 || s.stats.clauses != 1 || len(s.Mem.handles(0)) != 1 {
		t.Fatal(s.size, s.stats.clauses)
	}
	if got := sat_solver2_solve(s, nil, 0, 0, 0, 0); got != l_True {
		t.Fatal(got)
	}
}

func TestSolver2AssumptionsAgainstEnumeration(t *testing.T) {
	rng := rand.New(rand.NewSource(544))
	for trial := 0; trial < 100; trial++ {
		s := sat_solver2_new()
		s.pPrf1 = NewProofGraph()
		formula := [][]lit{}
		for count := 0; count < 15; count++ {
			c := []lit{}
			for _, v := range rng.Perm(5)[:3] {
				c = append(c, lit(2*v+rng.Intn(2)))
			}
			formula = append(formula, c)
			sat_solver2_addclause(s, c, int32(len(formula)))
		}
		assumptions := []lit{}
		for i := 0; i < 2; i++ {
			assumptions = append(assumptions, lit(rng.Intn(10)))
		}
		expected := formulaHasModel(formula, assumptions, 5)
		got := sat_solver2_solve(s, assumptions, 0, 0, 0, 0)
		if (got == l_True) != expected {
			t.Fatal(trial, assumptions, got, expected)
		}
		if got == l_False && s.hProofLast > 0 {
			subset := [][]lit{}
			for _, id := range Sat_ProofCore(s) {
				subset = append(subset, formula[id-1])
			}
			if formulaHasModel(subset, assumptions, 5) {
				t.Fatal("assumption proof core invalid")
			}
		}
	}
}

func TestSolver2RepeatedUnsatSolve(t *testing.T) {
	s := sat_solver2_new()
	s.pPrf1 = NewProofGraph()
	for i, c := range [][]lit{{0, 2}, {1, 2}, {0, 3}, {1, 3}} {
		sat_solver2_addclause(s, c, int32(i+1))
	}
	for repeat := 0; repeat < 3; repeat++ {
		if got := sat_solver2_solve(s, nil, 0, 0, 0, 0); got != l_False {
			t.Fatal(repeat, got)
		}
	}
}

func TestSolver2ReductionPreservesFinalProof(t *testing.T) {
	s := sat_solver2_new()
	sat_solver2_setnvars(s, 3)
	s.pPrf1 = NewProofGraph()
	root := clause2_create_new(s, []lit{0, 2, 4}, 0, 0)
	clause2_set_id(s, root, 1)
	for i := 0; i < 40; i++ {
		proof_chain_start(s, clause2_read(s, root))
		proof := proof_chain_stop(s)
		h := clause2_create_new(s, []lit{0, 2, 4}, 1, proof)
		clause2_read(s, h).lbd = 4
		s.act_clas.ptr[i] = int32(i * 100)
	}
	proof_chain_start(s, clause2_read(s, root))
	s.hProofLast = proof_chain_stop(s)
	s.nLearntStart, s.nLearntDelta, s.nLearntRatio = 20, 0, 25
	sat_solver2_reducedb(s)
	if core := Proof_DeriveCore(s.pPrf1, s.hProofLast); len(core) != 1 || core[0] != 1 {
		t.Fatal(core)
	}
}
