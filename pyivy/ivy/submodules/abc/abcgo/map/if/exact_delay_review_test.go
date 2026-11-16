package if_

import (
	abcutil "github.com/glycerine/abcgo/misc/util"
	"slices"
	"testing"
)

func exactDelayReviewCut(t *testing.T) (*If_Man_t, *If_Obj_t, *If_Cut_t) {
	t.Helper()
	p := If_ManStart(&If_Par_t{NLutSize: 6, NCutsMax: 4, FTruth: 1, FUsePerm: 1})
	t.Cleanup(func() { If_ManStop(p) })
	a, b := If_ManCreateCi(p), If_ManCreateCi(p)
	for _, obj := range []*If_Obj_t{a, b} {
		If_ManSetupCutTriv(p, If_ObjCutBest(obj), obj.Id)
	}
	a.CutBest.Delay, b.CutBest.Delay = 2.75, 5.25
	obj := If_ManCreateAnd(p, a, b)
	cut := If_ObjCutBest(obj)
	if If_CutMergeOrdered(p, &a.CutBest, &b.CutBest, cut) != 1 {
		t.Fatal("fixture merge failed")
	}
	If_CutComputeTruth(p, cut, &a.CutBest, &b.CutBest, 0, 0)
	cut.Cost = 37
	return p, obj, cut
}

func TestExactDelayDispatchAndSourceFailureMarkers(t *testing.T) {
	previous := exactDelayBackend()
	defer RegisterExactDelayCost(previous)
	for _, failed := range []bool{false, true} {
		p, obj, cut := exactDelayReviewCut(t)
		calls := 0
		RegisterExactDelayCost(func(truth *uint64, vars int32, arrivals *int32, pins []byte, cost *int32, aig int32) int32 {
			calls++
			if vars != 2 || *truth != 0x8888888888888888 || arrivals != &p.pArrTimeProfile[0] || !slices.Equal(p.pArrTimeProfile[:2], []int32{2, 5}) || *cost != 0 || aig != 6 || cut.fUser != 1 {
				t.Fatal("exact backend arguments/order changed")
			}
			for _, entry := range Vec_PtrArray(p.vObjs) {
				if entry.(*If_Obj_t).fVisit != 0 {
					t.Fatal("AIG bound retained visit marks")
				}
			}
			pins[0], pins[1] = 3, 1
			// Dispatch must release its registration lock before invoking ABC.
			RegisterExactDelayCost(nil)
			if failed {
				*cost = abcutil.ABC_INFINITY
				return abcutil.ABC_INFINITY
			}
			*cost = 3
			return 9
		})
		delay := ifCutExactDelay(p, obj, cut)
		if calls != 1 || cut.fUser != 1 {
			t.Fatal("exact backend was not called once")
		}
		if failed {
			if delay != float32(abcutil.ABC_INFINITY) || cut.Cost != IF_COST_MAX || cut.fUseless != 1 || !slices.Equal(If_CutPerm(cut), []byte{IF_BIG_CHAR, IF_BIG_CHAR}) {
				t.Fatal("source failure markers lost")
			}
		} else if delay != 9 || cut.Cost != 37 || cut.fUseless != 0 || !slices.Equal(If_CutPerm(cut), []byte{3, 1}) {
			t.Fatal("source successful cost/pin semantics changed")
		}
	}
}

func TestExactDelayMissingBackend(t *testing.T) {
	previous := exactDelayBackend()
	defer RegisterExactDelayCost(previous)
	RegisterExactDelayCost(nil)
	p, obj, cut := exactDelayReviewCut(t)
	defer func() {
		if failure := recover(); failure != "if: exact-delay backend is not registered" {
			t.Fatalf("missing backend: %v", failure)
		}
	}()
	ifCutExactDelay(p, obj, cut)
}

func TestExactDelayRetainsSourceUnsupportedGiaReconstruction(t *testing.T) {
	p := fromIfReviewRecordedManager()
	defer If_ManStop(p)
	p.pPars.FUserRecLib, p.pPars.FUserSesLib = 0, 1
	defer func() {
		if failure := recover(); failure != "if: unsupported recorded library" {
			t.Fatalf("source GIA assertion boundary: %v", failure)
		}
	}()
	Gia_ManFromIfAig(p)
}
