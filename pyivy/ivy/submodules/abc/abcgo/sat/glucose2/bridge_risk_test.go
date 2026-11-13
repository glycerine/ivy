package glucose2

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestABCBridgeIncrementalCoreAndReset(t *testing.T) {
	s := bmcg2_sat_solver_start()
	bmcg2_sat_solver_set_nvars(s, 2)
	if bmcg2_sat_solver_addclause(s, []int32{0, 2}, 2) != 1 {
		t.Fatal("clause rejected")
	}
	if got := bmcg2_sat_solver_solve(s, []int32{1, 3}, 2); got != GLUCOSE_UNSAT {
		t.Fatalf("contradictory assumptions: %d", got)
	}
	var core []int32
	if bmcg2_sat_solver_final(s, &core) != 2 {
		t.Fatalf("core %v", core)
	}
	if got := bmcg2_sat_solver_solve(s, []int32{0}, 1); got != GLUCOSE_SAT {
		t.Fatalf("base formula poisoned: %d", got)
	}
	if bmcg2_sat_solver_read_cex_varvalue(s, 0) != 1 {
		t.Fatal("wrong model")
	}
	bmcg2_sat_solver_reset(s)
	if bmcg2_sat_solver_varnum(s) != 0 {
		t.Fatal("reset retained variables")
	}
	if got := bmcg2_sat_solver_solve(s, nil, 0); got != GLUCOSE_SAT {
		t.Fatalf("reset solve %d", got)
	}
}

func TestABCBridgeEliminationAndModelReconstruction(t *testing.T) {
	s := bmcg2_sat_solver_start()
	for i := int32(0); i < 3; i++ {
		if got := bmcg2_sat_solver_addvar(s); got != i {
			t.Fatalf("addvar = %d, want %d", got, i)
		}
	}
	// (x or a) and (!x or b), with a and b retained for assumptions.
	for _, clause := range [][]int32{{0, 2}, {1, 4}} {
		if bmcg2_sat_solver_addclause(s, clause, int32(len(clause))) == 0 {
			t.Fatal("clause rejected")
		}
	}
	bmcg2_sat_solver_var_set_frozen(s, 1, 1)
	bmcg2_sat_solver_var_set_frozen(s, 2, 1)
	if bmcg2_sat_solver_eliminate(s, 0) == 0 || bmcg2_sat_solver_var_is_elim(s, 0) != 1 {
		t.Fatal("expected x to be eliminated")
	}
	if bmcg2_sat_solver_elim_varnum(s) != 1 || bmcg2_sat_solver_var_is_elim(s, 1) != 0 || bmcg2_sat_solver_var_is_elim(s, 2) != 0 {
		t.Fatal("frozen variables or elimination count incorrect")
	}
	if bmcg2_sat_solver_solve(s, []int32{3, 4}, 2) != GLUCOSE_SAT {
		t.Fatal("expected satisfiable assumptions")
	}
	model := bmcg2_sat_solver_read_cex(s)
	if len(model) != 3 || model[0] != 1 || model[1] != 0 || model[2] != 1 {
		t.Fatalf("reconstructed model %v", model)
	}
	model[0] = 0
	if bmcg2_sat_solver_read_cex_varvalue(s, 0) != 1 {
		t.Fatal("model copy aliases solver state")
	}
	if bmcg2_sat_solver_solve(s, []int32{3, 5}, 2) != GLUCOSE_UNSAT {
		t.Fatal("elimination lost the resolvent")
	}
	if bmcg2_sat_solver_solve(s, []int32{3, 4}, 2) != GLUCOSE_SAT {
		t.Fatal("unsatisfiable assumptions poisoned subsequent solve")
	}
}

func TestQuantificationCNFReloadAndGraphSupport(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	root := g.AND(x, y.Not())
	var objects, inputs []int32
	rootIndex := Gia_ManSatAndCollect2_rec(g, int32(root.NodeID()), &objects, &inputs)
	if objects[rootIndex] != int32(root.NodeID()) || len(inputs) != 2 {
		t.Fatalf("support objects=%v inputs=%v root=%d", objects, inputs, rootIndex)
	}
	nObjects, nInputs := len(objects), len(inputs)
	Gia_ManSatAndCollect2_rec(g, int32(root.NodeID()), &objects, &inputs)
	if len(objects) != nObjects || len(inputs) != nInputs {
		t.Fatal("repeated collection duplicated graph support")
	}
	solvers := [2]*SimpSolver{bmcg2_sat_solver_start(), bmcg2_sat_solver_start()}
	// Loading must discard old clauses and encode the graph without asserting root.
	for _, s := range solvers {
		bmcg2_sat_solver_set_nvars(s, 1)
		bmcg2_sat_solver_addclause(s, []int32{0}, 1)
		bmcg2_sat_solver_addclause(s, []int32{1}, 1)
	}
	Gia_ManQuantLoadCnf2(g, objects, solvers)
	for solverIndex, s := range solvers {
		for mask := 0; mask < 8; mask++ {
			assumptions := make([]int32, 3)
			for i, e := range []aig.Edge{x, y, root} {
				assumptions[i] = int32(e.NodeID() << 1)
				if mask&(1<<i) == 0 {
					assumptions[i]++
				}
			}
			want := (mask&4 != 0) == (mask&1 != 0 && mask&2 == 0)
			got := bmcg2_sat_solver_solve(s, assumptions, 3)
			if (got == GLUCOSE_SAT) != want {
				t.Fatalf("solver=%d assignment=%03b status=%d want SAT=%v", solverIndex, mask, got, want)
			}
		}
	}
}

func TestABCBridgeStopAndDeadline(t *testing.T) {
	s := bmcg2_sat_solver_start()
	bmcg2_sat_solver_set_nvars(s, 1)
	var stop int32 = 1
	bmcg2_sat_solver_set_stop(s, &stop)
	if got := bmcg2_sat_solver_solve(s, nil, 0); got != GLUCOSE_UNDEC {
		t.Fatalf("stopped result %d", got)
	}
	atomic.StoreInt32(&stop, 0)
	if got := bmcg2_sat_solver_solve(s, nil, 0); got != GLUCOSE_SAT {
		t.Fatalf("resumed result %d", got)
	}
	bmcg2_sat_solver_set_stop(s, nil)
	deadline := util.Abc_Clock() - int64(time.Second)
	if old := bmcg2_sat_solver_set_runtime_limit(s, deadline); old != 0 {
		t.Fatalf("old deadline %d", old)
	}
	if got := bmcg2_sat_solver_solve(s, nil, 0); got != GLUCOSE_UNDEC {
		t.Fatalf("expired result %d", got)
	}
	if old := bmcg2_sat_solver_set_runtime_limit(s, 0); old != deadline {
		t.Fatalf("old deadline %d want %d", old, deadline)
	}
	if got := bmcg2_sat_solver_solve(s, nil, 0); got != GLUCOSE_SAT {
		t.Fatalf("deadline clear result %d", got)
	}
}

func TestABCBridgeAndXorExhaustive(t *testing.T) {
	for _, xor := range []bool{false, true} {
		for phase := int32(0); phase < 8; phase++ {
			if xor && phase > 1 {
				continue
			}
			s := bmcg2_sat_solver_start()
			if xor {
				if bmcg2_sat_solver_add_xor(s, 2, 0, 1, phase) == 0 {
					t.Fatal("xor")
				}
			} else if bmcg2_sat_solver_add_and(s, 2, 0, 1, phase&1, (phase>>1)&1, (phase>>2)&1) == 0 {
				t.Fatal("and")
			}
			for mask := 0; mask < 8; mask++ {
				a, b, z := mask&1 != 0, mask&2 != 0, mask&4 != 0
				var want bool
				if xor {
					want = (a != b) != (phase != 0)
				} else {
					want = ((a != (phase&1 != 0)) && (b != (phase&2 != 0))) != (phase&4 != 0)
				}
				assumptions := make([]int32, 3)
				for i := range assumptions {
					assumptions[i] = int32(i * 2)
					if mask&(1<<i) == 0 {
						assumptions[i]++
					}
				}
				got := bmcg2_sat_solver_solve(s, assumptions, 3) == GLUCOSE_SAT
				if got != (z == want) {
					t.Fatalf("xor=%v phase=%d assignment=%03b sat=%v", xor, phase, mask, got)
				}
			}
		}
	}
}

func TestABCBridgeCallbackStatusAndModelOwnership(t *testing.T) {
	s := bmcg2_sat_solver_start()
	marker := new(int)
	var statuses []int32
	var snapshots [][]int32
	bmcg2_sat_solver_setcallback(s, marker, func(m any, status int32, model []int32) int32 {
		if m != marker {
			t.Fatal("wrong callback manager")
		}
		statuses = append(statuses, status)
		snapshots = append(snapshots, model)
		return 0
	})
	bmcg2_sat_solver_set_nvars(s, 1)
	if bmcg2_sat_solver_solve(s, []int32{0}, 1) != GLUCOSE_SAT {
		t.Fatal("first solve")
	}
	if bmcg2_sat_solver_solve(s, []int32{1}, 1) != GLUCOSE_SAT {
		t.Fatal("second solve")
	}
	if len(statuses) != 2 || statuses[0] != 1 || statuses[1] != 1 || len(snapshots[0]) != 1 || snapshots[0][0] != 1 || snapshots[1][0] != 0 {
		t.Fatalf("statuses=%v models=%v", statuses, snapshots)
	}
	snapshots[1][0] = 1
	if bmcg2_sat_solver_read_cex_varvalue(s, 0) != 0 {
		t.Fatal("callback model aliases solver")
	}
	bmcg2_sat_solver_setcallback(s, nil, nil)
	if bmcg2_sat_solver_solve(s, nil, 0) != GLUCOSE_SAT || len(statuses) != 2 {
		t.Fatal("callback removal failed")
	}
}
