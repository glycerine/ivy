package gia

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

func topAndReviewSatCount(t *testing.T, g *Gia_Man_t) int {
	t.Helper()
	count := 0
	for mask := 0; mask < 1<<uint(Gia_ManCiNum(g)); mask++ {
		inputs := make([]bool, Gia_ManCiNum(g))
		for i := range inputs {
			inputs[i] = mask&(1<<uint(i)) != 0
		}
		values, err := g.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		all := true
		for _, value := range values {
			all = all && value
		}
		if all {
			count++
		}
	}
	return count
}

func TestTopAndRetainsFalseOutput(t *testing.T) {
	for _, full := range []bool{false, true} {
		t.Run(fmt.Sprint(full), func(t *testing.T) {
			src := Gia_ManStart(1)
			a, b := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
			Gia_ManAppendCo(src, 0)
			Gia_ManAppendCo(src, Gia_ManAppendAnd(src, a, b))
			var out *Gia_Man_t
			if full {
				out = Gia_ManDupTopAnd(src, 0)
			} else {
				out = Gia_ManDupTopAnd_iter(src, 0)
			}
			if topAndReviewSatCount(t, src) != 0 || topAndReviewSatCount(t, out) != 0 {
				t.Fatal("AND decomposition removed a false conjunct and made UNSAT satisfiable")
			}
		})
	}
}

func TestDupClassesPreservesConstraintRoles(t *testing.T) {
	src := Gia_ManStart(1)
	a, r := Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	Gia_ManAppendCo(src, a)
	Gia_ManAppendCo(src, a^1)
	Gia_ManAppendCo(src, r)
	Gia_ManSetRegNum(src, 1)
	src.nConstrs = 1
	src.pReprsOld = make([]int32, src.nObjs)
	for i := range src.pReprsOld {
		src.pReprsOld[i] = -1
	}
	if out := Gia_ManDupDfsClasses(src); out.nConstrs != 1 {
		t.Fatal("class duplication reclassified a retained constraint")
	}
}

func topAndReviewSharedGraph(depth int) *Gia_Man_t {
	src := Gia_ManStart(1)
	a, b, c := Gia_ManAppendCi(src), Gia_ManAppendCi(src), Gia_ManAppendCi(src)
	x, y := Gia_ManAppendAnd(src, a, b), Gia_ManAppendAnd(src, a, c)
	for i := 0; i < depth; i++ {
		old := x
		x = Gia_ManAppendAnd(src, x, y)
		y = Gia_ManAppendAnd(src, old, a)
	}
	Gia_ManAppendCo(src, x)
	return src
}

func TestTopAndSharedConeTerminates(t *testing.T) {
	const key = "ABCGO_TOPAND_SHARED_REVIEW_CHILD"
	if os.Getenv(key) == "1" {
		debug.SetMemoryLimit(32 << 20)
		timer := time.AfterFunc(2*time.Second, func() { panic("shared cone expansion did not finish") })
		defer timer.Stop()
		out := Gia_ManDupTopAnd_iter(topAndReviewSharedGraph(48), 0)
		if Gia_ManCiNum(out) != 3 || Gia_ManCoNum(out) != 0 || Gia_ManAndNum(out) != 0 {
			t.Fatal("shared conjunction was not reduced to its satisfiable CI assignments")
		}
		return
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestTopAndSharedConeTerminates$", "-test.count=1")
	cmd.Env = append(os.Environ(), key+"=1")
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("shared cone: %v (context %v): %s", err, ctx.Err(), data)
	}
}

func TestPrintOldRepresentativesUsesOldTable(t *testing.T) {
	const key = "ABCGO_PRINT_OLD_REVIEW_CHILD"
	if mode := os.Getenv(key); mode != "" {
		src := Gia_ManStart(1)
		for i := 0; i < 3; i++ {
			Gia_ManAppendCi(src)
		}
		src.pReprsOld = []int32{-1, -1, -1, 1}
		if mode == "both" {
			src.StartRepresentatives()
			Gia_ObjSetRepr(src, 3, 2)
		}
		Gia_ManPrintRepr(src)
		return
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"old-only", "both"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run=^TestPrintOldRepresentativesUsesOldTable$", "-test.count=1")
			cmd.Env = append(os.Environ(), key+"="+mode)
			data, err := cmd.CombinedOutput()
			if err != nil || !strings.HasPrefix(string(data), "3->1 \n") {
				t.Fatalf("old representatives: %v: %s", err, data)
			}
		})
	}
}

var topAndReviewSink *Gia_Man_t

func BenchmarkTopAndSharedFrontier(b *testing.B) {
	src := topAndReviewSharedGraph(12)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		topAndReviewSink = Gia_ManDupTopAnd_iter(src, 0)
	}
}
