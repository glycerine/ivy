package gia

import (
	"testing"

	"github.com/glycerine/abcgo/sat/bsat"
)

func TestFalsePathRecheckBudgetDoesNotCollectHook(t *testing.T) {
	g, root := falseReviewGraph(12)
	defer Gia_ManStop(g)
	Gia_ManLevelNum(g)
	query := giaFalseSolveQuery(g, root, 1, true)
	if query.status != bsat.Unknown {
		t.Fatalf("expected exhausted recheck budget, got%d", query.status)
	}
	hooks := Vec_WecStart(int(g.nObjs))
	if query.hook() != nil || query.addBatchHook(hooks) || Vec_WecSizeUsed(hooks) != 0 {
		t.Fatal("resource exhaustion yielded a rewrite certificate")
	}
	// With an unlimited budget the same query is decided, and any certificate
	// must no longer rely on the original input-transition assumption.
	query = giaFalseSolveQuery(g, root, 0, true)
	if query.status == bsat.Unknown {
		t.Fatal("unbounded query stayed unknown")
	}
	for _, lit := range query.core {
		if lit>>1 == 0 {
			t.Fatal("retained unsafe original-PI assumption")
		}
	}
}

func TestFalsePathBatchCertificatesPreserveFunctions(t *testing.T) {
	checked := 0
	for seed := uint32(0); seed < 128; seed++ {
		g, root := falseReviewGraph(seed)
		Gia_ManLevelNum(g)
		query := giaFalseSolveQuery(g, root, 0, true)
		if query.status == bsat.Unknown {
			t.Fatal("unexpected unlimited unknown")
		}
		for _, lit := range query.core {
			if lit>>1 == 0 {
				t.Fatal("unsafe original-PI condition in final certificate")
			}
		}
		if path := query.hook(); path != nil {
			out := Gia_ManFalseRebuildPath(g, path, 0, 0)
			cofReviewCheckRebuild(t, g, out)
			Gia_ManStop(out)
			checked++
		}
		Gia_ManStop(g)
	}
	t.Logf("%d individual batch certificates checked at all assignments", checked)
}

func TestFalsePathDeepCriticalPath(t *testing.T) {
	g := Gia_ManStart(6100)
	defer Gia_ManStop(g)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, b)
	root = Gia_ManAppendAnd(g, root, a^1)
	for i := 0; i < 6000; i++ {
		root = Gia_ManAppendAnd(g, root, b)
	}
	Gia_ManAppendCo(g, root)
	Gia_ManLevelNum(g)
	path := Gia_ManCollectPath(g, Gia_ManCo(g, 0))
	if Vec_IntSize(path) != 6003 || Vec_IntEntry(path, 0) != b>>1 || Vec_IntEntry(path, 6002) != root>>1 {
		t.Fatal("deep critical path truncated or reversed")
	}
	out := Gia_ManCheckOne(g, 0, root>>1, 0, 0, 0)
	if out == nil {
		t.Fatal("deep constant-false path was not eliminated")
	}
	defer Gia_ManStop(out)
	cofReviewCheckRebuild(t, g, out)
}

func TestFalsePathEmptyAndConstantOutputs(t *testing.T) {
	for _, outputs := range []bool{false, true} {
		for _, driver := range []string{"batch", "outputs", "all"} {
			g := Gia_ManStart(8)
			if outputs {
				a := Gia_ManAppendCi(g)
				Gia_ManAppendCo(g, 0)
				Gia_ManAppendCo(g, 1)
				Gia_ManAppendCo(g, a)
				Gia_ManAppendCo(g, a^1)
			}
			var out *Gia_Man_t
			hashReviewStdout(t, func() {
				switch driver {
				case "batch":
					out = Gia_ManCheckFalse2(g, 32, 0, 0, 0)
				case "outputs":
					out = Gia_ManCheckFalse(g, 32, 0, 0, 0)
				case "all":
					out = Gia_ManCheckFalseAll(g, 32, 0, 0, 0)
				}
			})
			cofReviewCheckRebuild(t, g, out)
			Gia_ManStop(out)
			Gia_ManStop(g)
		}
	}
}
