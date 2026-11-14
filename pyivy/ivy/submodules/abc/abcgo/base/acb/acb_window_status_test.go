package acb

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
)

// These tests are intentionally sequential because the source diagnostics use
// process-wide stdout. Capture only the call whose diagnostic is under test.
func acbCaptureStdout(t *testing.T, fn func()) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	saved := os.Stdout
	defer func() { os.Stdout = saved }()
	os.Stdout = f
	fn()
	os.Stdout = saved
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestWindowSolverConstructionStatusAndRestoration(t *testing.T) {
	cases := []struct {
		name                string
		vars, pivot, divs   int32
		clauses             [][]int32
		firstOK, multipleOK bool
	}{
		{"observable_buffer", 2, 1, 1, [][]int32{{1, 2}, {0, 3}}, true, true},
		{"constant_pivot", 1, 0, 0, [][]int32{{0}}, true, false},
		{"contradictory_units", 2, 1, 1, [][]int32{{2}, {3}}, false, false},
		{"propagation_conflict", 2, 1, 1, [][]int32{{1, 2}, {1, 3}, {0}}, false, false},
		{"empty_clause", 1, 0, 0, [][]int32{{}}, false, false},
		{"empty_formula", 1, 0, 0, nil, true, true},
	}
	for _, tc := range cases {
		for _, copies := range []int32{1, 2, 6} {
			t.Run(fmt.Sprintf("%s_copies%d", tc.name, copies), func(t *testing.T) {
				var lits []int32
				offsets := []int32{0}
				for _, clause := range tc.clauses {
					lits = append(lits, clause...)
					offsets = append(offsets, int32(len(lits)))
				}
				cnf := new(Cnf_Dat_t)
				Cnf_DataSetPackedClauses(cnf, tc.vars, offsets, lits)
				defer Cnf_DataFree(cnf)
				before := make([][]int32, Cnf_DataClauseNum(cnf))
				for i := range before {
					before[i] = append([]int32(nil), Cnf_DataClause(cnf, int32(i))...)
				}
				flips := Vec_IntAlloc(0)
				Cnf_DataCollectFlipLits(cnf, tc.pivot, flips)
				s := sat_solver_new()
				defer sat_solver_delete(s)
				var status int32
				output := acbCaptureStdout(t, func() { status = Acb_NtkWindow2Solver(s, cnf, flips, tc.pivot, tc.divs, copies) })
				want := tc.firstOK
				if copies > 1 {
					want = tc.multipleOK
				}
				if (status != 0) != want {
					t.Errorf("construction status=%d want consistent=%v", status, want)
				}
				if got := sat_solver_solve(s, nil, nil, 0, 0, 0, 0) == l_True; got != want {
					t.Errorf("solver SAT=%v want%v", got, want)
				}
				if output != "" {
					t.Errorf("normal construction wrote diagnostic: %s", output)
				}
				for i := range before {
					got := append([]int32(nil), Cnf_DataClause(cnf, int32(i))...)
					if !reflect.DeepEqual(got, before[i]) {
						t.Fatalf("CNF clause%d changed from%v to%v", i, before[i], got)
					}
				}
			})
		}
	}
}

func TestUnobservableWindowIsAnOptimizationResult(t *testing.T) {
	for _, verbose := range []int32{0, 1} {
		t.Run(fmt.Sprint(verbose), func(t *testing.T) {
			p := acbNetwork(t, 2)
			pivot := acbNode(p, stretchTruth(8, 2), 1, 2)
			// The constant fanout retains an input edge, as in the randomized
			// circuits. Its reconvergent side input brings it into the window.
			masked := acbNode(p, 0, pivot, 1)
			acbOutput(p, masked)
			acbOutput(p, 1)
			want := acbTable(p)
			var pars Acb_Par_t
			Acb_ParSetDefault(&pars)
			pars.fVerbose = verbose
			m := Acb_MfsStart(p, &pars)
			defer Acb_MfsStop(m)
			Acb_NtkCreateFanout(p)
			Acb_NtkCleanObjFuncs(p)
			Acb_NtkCleanObjCnfs(p)
			Acb_NtkUpdateLevelD(p, -1)
			var changed int32
			output := acbCaptureStdout(t, func() { changed = Acb_NtkOptNode(m, pivot) })
			if strings.Contains(output, "Error:") {
				t.Fatalf("unobservable node diagnosed as an error: %s", output)
			}
			if verbose == 0 && output != "" {
				t.Fatalf("nonverbose optimization wrote %q", output)
			}
			if verbose != 0 && !strings.Contains(output, "unobservable") {
				t.Fatalf("missing unobservable-window explanation: %s", output)
			}
			if changed != 1 || Acb_ObjTruth(p, pivot) != 0 || Acb_ObjFaninNum(p, pivot) != 0 {
				t.Fatal("did not replace unobservable pivot with zero")
			}
			if !reflect.DeepEqual(acbTable(p), want) {
				t.Fatal("removing unobservable logic changed outputs")
			}
			acbCheckLinks(t, p)
		})
	}
}

func TestZeroSupportBudgetKeepsFeasibleSupport(t *testing.T) {
	s := sat_solver_new()
	defer sat_solver_delete(s)
	sat_solver_addclause(s, []int32{0, 2})
	weights := Vec_IntStart(2)
	Vec_IntWriteEntry(weights, 0, 1)
	Vec_IntWriteEntry(weights, 1, 2)
	initial := Vec_IntAlloc(2)
	Vec_IntPushTwo(initial, 1, 3)
	var result *Vec_Int_t
	output := acbCaptureStdout(t, func() { result = Acb_FindSupport(s, 0, weights, initial, 0) })
	if strings.Contains(output, "Timeout") || !strings.Contains(output, "zero time budget") {
		t.Fatalf("zero budget misreported: %s", output)
	}
	if result == initial || !reflect.DeepEqual(Vec_IntArray(result), Vec_IntArray(initial)) {
		t.Fatal("zero budget must retain a separate copy of the supplied support")
	}
	if sat_solver_solve(s, Vec_IntArray(result), nil, 0, 0, 0, 0) != l_False {
		t.Fatal("returned support is not feasible")
	}
}
