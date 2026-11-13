package xsat

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
)

func brute(n int, clauses [][]int) bool {
	for mask := 0; mask < 1<<n; mask++ {
		ok := true
		for _, c := range clauses {
			sat := false
			for _, lit := range c {
				if ((mask>>(lit>>1))&1)^(lit&1) == 1 {
					sat = true
					break
				}
			}
			if !sat {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}
func checkModel(t *testing.T, s *Solver, clauses [][]int) {
	t.Helper()
	for _, c := range clauses {
		sat := false
		for _, lit := range c {
			if s.ModelValue(lit>>1)^(lit&1) == 1 {
				sat = true
			}
		}
		if !sat {
			t.Fatal("invalid model", c, s.model)
		}
	}
}
func TestRandomCNFAgainstExhaustive(t *testing.T) {
	rng := rand.New(rand.NewSource(8173))
	for run := 0; run < 1500; run++ {
		n := 1 + rng.Intn(8)
		s := New()
		for v := 0; v < n; v++ {
			s.AddVariable(1)
		}
		var clauses [][]int
		count := rng.Intn(30)
		for c := 0; c < count; c++ {
			lits := make([]int, 1+rng.Intn(5))
			for j := range lits {
				lits[j] = rng.Intn(2 * n)
			}
			clauses = append(clauses, lits)
			s.AddClause(lits)
		}
		want := brute(n, clauses)
		for trial := 0; trial < 3; trial++ {
			got := s.Solve()
			if got == LBoolUndef || (got == LBoolTrue) != want {
				t.Fatalf("case %d trial %d result %d SAT %v clauses %v", run, trial, got, want, clauses)
			}
			if got == LBoolTrue {
				checkModel(t, s, clauses)
			}
		}
	}
}

func TestIncrementalClausesAndSimplification(t *testing.T) {
	rng := rand.New(rand.NewSource(91812))
	for trial := 0; trial < 400; trial++ {
		const n = 6
		s := New()
		for i := 0; i < n; i++ {
			s.AddVariable(1)
		}
		var clauses [][]int
		for step := 0; step < 14; step++ {
			clause := make([]int, 1+rng.Intn(4))
			for i := range clause {
				clause[i] = rng.Intn(2 * n)
			}
			clauses = append(clauses, append([]int(nil), clause...))
			s.AddClause(clause)
			if step%3 == 0 {
				s.Simplify()
			}
			if step%4 == 0 {
				xSAT_SolverGarbageCollect(s)
			}
			want, got := brute(n, clauses), s.Solve()
			if got == LBoolUndef || (got == LBoolTrue) != want {
				t.Fatalf("trial=%d step=%d got=%d want=%v clauses=%v", trial, step, got, want, clauses)
			}
			if got == LBoolTrue {
				checkModel(t, s, clauses)
			}
		}
	}
}
func TestHardRandomThreeSAT(t *testing.T) {
	rng := rand.New(rand.NewSource(35235))
	for run := 0; run < 400; run++ {
		n := 4 + rng.Intn(6)
		s := New()
		for v := 0; v < n; v++ {
			s.AddVariable(1)
		}
		var clauses [][]int
		for c := 0; c < 4*n; c++ {
			vars := rng.Perm(n)[:3]
			lits := make([]int, 3)
			for i, v := range vars {
				lits[i] = 2*v + rng.Intn(2)
			}
			clauses = append(clauses, lits)
			s.AddClause(lits)
		}
		want := brute(n, clauses)
		for trial := 0; trial < 3; trial++ {
			got := s.Solve()
			if got == LBoolUndef || (got == LBoolTrue) != want {
				t.Fatalf("case %d trial %d result %d SAT %v clauses %v", run, trial, got, want, clauses)
			}
			if got == LBoolTrue {
				checkModel(t, s, clauses)
			}
		}
	}
}
func TestPigeonholeLearning(t *testing.T) {
	const pigeons = 5
	const holes = 4
	s := New()
	for p := 0; p < pigeons; p++ {
		var lits []int
		for h := 0; h < holes; h++ {
			lits = append(lits, 2*(p*holes+h))
		}
		s.AddClause(lits)
	}
	for h := 0; h < holes; h++ {
		for p := 0; p < pigeons; p++ {
			for q := p + 1; q < pigeons; q++ {
				s.AddClause([]int{2*(p*holes+h) + 1, 2*(q*holes+h) + 1})
			}
		}
	}
	if s.Solve() != LBoolFalse || s.LearntCount() == 0 {
		t.Fatal("learning failed", s.Stats)
	}
}
func TestDIMACSAndOriginalVectorAPI(t *testing.T) {
	s, err := ParseDIMACS(strings.NewReader("c test\np cnf 3 3\n1 -2 0\n2 3 0\n-1 0\n"))
	if err != nil || s.Solve() != LBoolTrue {
		t.Fatal(err)
	}
	p := xSAT_SolverCreate()
	lits := vec.Vec_IntAllocArrayCopy([]int32{0, 2}, 2)
	if xSAT_SolverAddClause(p, lits) == 0 || xSAT_SolverSolve(p) != LBoolTrue {
		t.Fatal("original vector API")
	}
	xSAT_SolverDestroy(p)
	if p.memory != nil || p.assigns != nil {
		t.Fatal("destroy retained storage")
	}
	for _, bad := range []string{"p cnf 1 1\n2 0\n", "p cnf 1 1\n1\n", "p cnf 1 2\n1 0\n"} {
		if _, err := ParseDIMACS(strings.NewReader(bad)); err == nil {
			t.Fatal("accepted bad DIMACS", bad)
		}
	}
}
func TestLimitsAndIncremental(t *testing.T) {
	s := New()
	s.AddClause([]int{0, 2})
	s.Config.Deadline = time.Now().Add(-time.Second)
	if s.Solve() != LBoolUndef {
		t.Fatal("deadline ignored")
	}
	s.Config.Deadline = time.Time{}
	if s.Solve() != LBoolTrue {
		t.Fatal("deadline poisoned solver")
	}
	s.AddClause([]int{1})
	if s.Solve() != LBoolTrue || s.ModelValue(1) != 1 {
		t.Fatal("incremental unit")
	}
	if s.AddClause([]int{3}) || s.Solve() != LBoolFalse {
		t.Fatal("incremental contradiction")
	}
}
func TestQueueMixedWrapPopPush(t *testing.T) {
	p := xSAT_BQueueNew(3)
	for _, v := range []uint32{1, 2, 3, 4} {
		xSAT_BQueuePush(p, v)
	}
	if xSAT_BQueueAvg(p) != 3 || xSAT_BQueuePop(p) != 2 {
		t.Fatal("queue wrap")
	}
	xSAT_BQueuePush(p, 5)
	var got []int
	for len(p.data) > 0 {
		got = append(got, xSAT_BQueuePop(p))
	}
	if !reflect.DeepEqual(got, []int{3, 4, 5}) {
		t.Fatal(got)
	}
	xSAT_BQueueClean(p)
	xSAT_BQueuePush(p, ^uint32(0))
	xSAT_BQueuePush(p, ^uint32(0))
	if xSAT_BQueueAvg(p) != ^uint32(0) {
		t.Fatal("sum overflow")
	}
}
func TestWatchRemoval(t *testing.T) {
	w := &xSAT_WatchList_t{data: []xSAT_Watcher_t{{CRef: 1}, {CRef: 2}, {CRef: 3}}}
	xSAT_WatchListRemove(w, 2)
	if len(w.data) != 2 || w.data[1].CRef != 3 {
		t.Fatal(w)
	}
}

func TestOriginalVectorNormalizationAndNativeOwnership(t *testing.T) {
	s := New()
	v := vec.Vec_IntAllocArrayCopy([]int32{4, 0, 4, 2}, 4)
	if xSAT_SolverAddClause(s, v) == 0 || !reflect.DeepEqual(vec.Vec_IntArray(v), []int32{0, 2, 4}) {
		t.Fatal("compatibility vector not normalized", vec.Vec_IntArray(v))
	}
	raw := []int{8, 6, 8}
	s.AddClause(raw)
	if !reflect.DeepEqual(raw, []int{8, 6, 8}) {
		t.Fatal("native AddClause mutated input", raw)
	}
}
func TestLongWatchTailPreservedOnConflict(t *testing.T) {
	s := New()
	for _, lits := range [][]int{{1, 2, 4}, {1, 6, 8}, {1, 10, 12}} {
		s.AddClause(lits)
	}
	s.limits = append(s.limits, len(s.trail))
	s.enqueue(3, CRefUndef)
	s.enqueue(5, CRefUndef)
	s.qhead = len(s.trail)
	s.enqueue(0, CRefUndef)
	if s.propagate() == CRefUndef {
		t.Fatal("expected conflict")
	}
	if len(s.watches.data[0].data) != 3 {
		t.Fatal("lost unprocessed watch tail", s.watches.data[0].data)
	}
}
func TestGarbageCollectionRewritesAllHandles(t *testing.T) {
	s := New()
	for _, lits := range [][]int{{1, 2, 4}, {0, 6, 8}, {3, 6, 10}} {
		s.AddClause(lits)
	}
	dead := s.clauses[0]
	s.memory.clauses[dead].marked = true
	s.unwatch(dead)
	s.clauses = s.clauses[1:]
	live := s.memory.clauses[s.clauses[0]]
	s.enqueue(0, s.clauses[0])
	xSAT_SolverGarbageCollect(s)
	if len(s.memory.clauses) != 2 || s.memory.clauses[s.reasons[0]] != live {
		t.Fatal("root reason not relocated")
	}
	for _, watches := range []*xSAT_VecWatchList_t{s.binary, s.watches} {
		for _, list := range watches.data {
			for _, w := range list.data {
				if int(w.CRef) >= len(s.memory.clauses) || s.memory.clauses[w.CRef].marked {
					t.Fatal("stale watch after collection", w)
				}
			}
		}
	}
	if s.Solve() != LBoolTrue {
		t.Fatal("relocated solver unusable")
	}
}
