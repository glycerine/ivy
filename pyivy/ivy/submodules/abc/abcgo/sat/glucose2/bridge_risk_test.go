package glucose2

import (
	"sync/atomic"
	"testing"
	"time"

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
