package gia

import (
	"slices"
	"testing"

	"github.com/glycerine/abcgo/sat/bsat"
)

func falseReviewQuery(begin, end int) giaFalseQuery {
	ids := make([]int32, 24)
	for i := range ids {
		ids[i] = int32(100 + i)
	}
	core := []bsat.Literal{int32(2*end + 1)}
	if begin != end {
		core = append(core, int32(2*begin+1))
	}
	return giaFalseQuery{ids: ids, core: core, status: bsat.Unsatisfiable}
}

func TestFalsePathBatchSourceIntervalBoundary(t *testing.T) {
	// C tests the final-core span, not the length after adding the preceding
	// node. A span of nineteen can therefore produce a twenty-one-node hook.
	for _, test := range []struct {
		begin, end int
		accept     bool
	}{{1, 20, true}, {0, 19, true}, {0, 20, false}, {1, 21, false}, {19, 20, true}} {
		q := falseReviewQuery(test.begin, test.end)
		hooks := Vec_WecStart(200)
		if got := q.addBatchHook(hooks); got != test.accept {
			t.Errorf("core interval%d..%d (hook length%d): accepted%t want%t", test.begin, test.end, Vec_IntSize(q.hook()), got, test.accept)
			continue
		}
		if test.accept {
			want := q.ids[max(test.begin-1, 0) : test.end+1]
			if got := Vec_IntArray(Vec_WecEntry(hooks, int(want[len(want)-1]))); !slices.Equal(got, want) {
				t.Fatalf("hook%v want%v", got, want)
			}
		} else if Vec_WecSizeUsed(hooks) != 0 {
			t.Fatal("rejected interval still changed hooks")
		}
	}
}
func TestFalsePathBatchHookOverlapAndStatus(t *testing.T) {
	q := falseReviewQuery(1, 5)
	hooks := Vec_WecStart(200)
	Vec_WecPush(hooks, 103, 99)
	if q.addBatchHook(hooks) || Vec_WecSizeUsed(hooks) != 1 {
		t.Fatal("overlapping hook was collected")
	}
	hooks = Vec_WecStart(200)
	for _, status := range []bsat.Status{bsat.Unknown, bsat.Satisfiable} {
		q.status = status
		if q.addBatchHook(hooks) {
			t.Fatal("collected a hook without UNSAT certificate")
		}
	}
	q.status = bsat.Unsatisfiable
	if !q.addBatchHook(hooks) || q.addBatchHook(hooks) {
		t.Fatal("did not retain exactly one non-overlapping hook")
	}
}
