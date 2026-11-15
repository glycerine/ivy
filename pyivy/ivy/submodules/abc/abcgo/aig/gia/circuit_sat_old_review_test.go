package gia

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestChronologicalMiterStatsExcludeConstants(t *testing.T) {
	g := Gia_ManStart(16)
	lit := Gia_ManAppendCi(g)
	for _, root := range []int32{0, 1, lit, lit ^ 1} {
		Gia_ManAppendCo(g, root)
	}
	text := hashReviewStdout(t, func() { Cbs_ManSolveMiter(g, 1000, nil, 1) })
	for _, pattern := range []string{`Unsat calls\s+0`, `Sat\s+calls\s+2`, `Undef calls\s+0`, `\(\s*50\.00 %\)`, `Ave conf =\s+0\.0`, `Total time =`} {
		if !regexp.MustCompile(pattern).MatchString(text) {
			t.Fatalf("source statistics %q missing: %q", pattern, text)
		}
	}
	if !strings.Contains(text, "Constant 1 output of SRM!!!") {
		t.Fatal("source constant-output diagnostic missing")
	}
}

func BenchmarkChronologicalLeafQueries(b *testing.B) {
	for _, size := range []int{10, 10000} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			g := Gia_ManStart(int32(size + 8))
			a, x := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
			root := a
			for i := 0; i < size; i++ {
				root = Gia_ManAppendAnd(g, root, x^int32(i&1))
			}
			p := Cbs0_ManAlloc(g)
			defer Cbs0_ManStop(p)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if Cbs0_ManSolve(p, Gia_ObjFromLit(g, a)) != 0 {
					b.Fatal("leaf query failed")
				}
			}
		})
	}
}

func TestChronologicalRejectsTypedGatesAfterAllocation(t *testing.T) {
	for _, kind := range []string{"xor", "mux"} {
		g := Gia_ManStart(16)
		a, b, c := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		p := Cbs0_ManAlloc(g)
		if kind == "xor" {
			Gia_ManAppendXorReal(g, a, b)
		} else {
			g.EnableMuxes()
			Gia_ManAppendMuxReal(g, a, b, c)
		}
		cexReviewMustPanic(t, func() { Cbs0_ManSolve(p, Gia_ObjFromLit(g, a)) })
		Cbs0_ManStop(p)
	}
}

func TestChronologicalNestedQueuesAndStop(t *testing.T) {
	g := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	queue := Cbs0_Que_t{}
	Cbs0_QuePush(&queue, Gia_ObjFromLit(g, a))
	Cbs0_QuePush(&queue, Gia_ObjFromLit(g, b))
	var head0, tail0, head1, tail1 int32
	Cbs0_QueStore(&queue, &head0, &tail0)
	// Force backing-slice growth while a parent frontier remains suspended.
	for i := 0; i < 3000; i++ {
		Cbs0_QuePush(&queue, Gia_ObjFromLit(g, a))
	}
	Cbs0_QueStore(&queue, &head1, &tail1)
	queue.pData[queue.iHead] = nil
	Cbs0_QueRestore(&queue, head1, tail1)
	if Cbs0_QueHasNode(&queue, Gia_ObjFromLit(g, b)) != 1 {
		t.Fatal("nested frontier lost its parent copy")
	}
	Cbs0_QueRestore(&queue, head0, tail0)
	if queue.iHead != 0 || queue.iTail != 2 || queue.pData[0] != Gia_ObjFromLit(g, a) || queue.pData[1] != Gia_ObjFromLit(g, b) {
		t.Fatal("outer frontier lost")
	}
	for _, bounds := range [][2]int32{{-1, 0}, {2, 1}, {0, int32(len(queue.pData) + 1)}} {
		cexReviewMustPanic(t, func() { Cbs0_QueRestore(&queue, bounds[0], bounds[1]) })
	}
	p := Cbs0_ManAlloc(g)
	Cbs0_ManAssign(p, Gia_ObjFromLit(g, a))
	Cbs0_ManAssign(p, Gia_ObjFromLit(g, b^1))
	if Cbs0_VarValue(Gia_ObjFromLit(g, a)) != 1 || Cbs0_VarValue(Gia_ObjFromLit(g, b)) != 0 {
		t.Fatal("assignment polarity")
	}
	Cbs0_ManCancelUntil(p, 1)
	if Cbs0_VarIsAssigned(Gia_ObjFromLit(g, a)) != 1 || Cbs0_VarIsAssigned(Gia_ObjFromLit(g, b)) != 0 {
		t.Fatal("partial cancellation")
	}
	cexReviewMustPanic(t, func() { Cbs0_ManCancelUntil(p, 2) })
	Cbs0_ManStop(p)
	if g.pObjs[a>>1].fMark0 != 0 || p.pAig != nil || p.vModel != nil || p.pProp.pData != nil || p.pJust.pData != nil {
		t.Fatal("active stop retained state")
	}
	Cbs0_ManStop(p)
	Cbs0_ManStop(nil)
}

func TestChronologicalRootGuardsAndRepeatedFanins(t *testing.T) {
	g := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, b)
	co := Gia_ManAppendCo(g, root)
	p := Cbs0_ManAlloc(g)
	other := Gia_ManStart(8)
	foreign := Gia_ManAppendCi(other)
	for _, bad := range []*Gia_Obj_t{nil, Gia_ObjFromLit(other, foreign), Gia_ObjFromLit(g, co)} {
		cexReviewMustPanic(t, func() { Cbs0_ManSolve(p, bad) })
	}
	if Cbs0_ManSolve(p, Gia_ObjFromLit(g, root)) != 0 {
		t.Fatal("invalid query poisoned solver")
	}
	Cbs0_ManStop(p)
	// Buffers have repeated regular fanins. The Go extension must assign once.
	buf := Gia_ManAppendBuf(g, a^1)
	p = Cbs0_ManAlloc(g)
	for _, lit := range []int32{buf, buf ^ 1, 0, 1} {
		if status := Cbs0_ManSolve(p, Gia_ObjFromLit(g, lit)); status != giaBool(lit == 0) {
			t.Fatalf("buffer/constant literal%d status%d", lit, status)
		}
		if lit != 0 {
			for assignment := 0; assignment < 4; assignment++ {
				matches := true
				for _, entry := range Vec_IntArray(p.vModel) {
					if entry < 0 || entry>>1 >= 2 {
						t.Fatal("buffer model CI index")
					}
					matches = matches && (assignment>>uint(entry>>1)&1 != 0) == (entry&1 == 0)
				}
				want := lit == 1 || lit == buf && assignment&1 == 0 || lit == buf^1 && assignment&1 != 0
				if matches && !want {
					t.Fatal("buffer/constant cube does not imply root")
				}
			}
		}
		for _, obj := range g.pObjs {
			if obj.fMark0 != 0 || obj.fMark1 != 0 {
				t.Fatal("repeated-fanin query retained marks")
			}
		}
	}
	Cbs0_ManStop(p)
}

func TestChronologicalConstantFaninsAndSourceAllocation(t *testing.T) {
	for _, constant := range []int32{0, 1} {
		g := Gia_ManStart(16)
		a := Gia_ManAppendCi(g)
		root := Gia_ManAppendAnd(g, constant, a)
		Gia_ManCreateRefs(g)
		p := Cbs0_ManAlloc()
		p.pAig = g // Original allocation/setup convention.
		for phase := int32(0); phase < 2; phase++ {
			status := Cbs0_ManSolve(p, Gia_ObjFromLit(g, root^phase))
			if want := giaBool(constant == 0 && phase == 0); status != want {
				t.Fatalf("constant%d phase%d status%d want%d", constant, phase, status, want)
			}
			if status == 0 {
				for input := 0; input < 2; input++ {
					matches := true
					for _, lit := range Vec_IntArray(p.vModel) {
						if lit>>1 != 0 {
							t.Fatal("constant-fanin model CI index")
						}
						matches = matches && (input == 1) == (lit&1 == 0)
					}
					if matches && (constant != 0 && input != 0) == (phase != 0) {
						t.Fatal("constant-fanin cube does not force the root")
					}
				}
			}
			for _, obj := range g.pObjs {
				if obj.fMark0 != 0 || obj.fMark1 != 0 {
					t.Fatal("constant-fanin query retained marks")
				}
			}
		}
		Cbs0_ManStop(p)
	}
}

func TestChronologicalStopAfterInvalidHeuristic(t *testing.T) {
	g := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	root := Gia_ManAppendAnd(g, a, b)
	p := Cbs0_ManAlloc(g)
	p.Pars.fUseHighest, p.Pars.fUseLowest, p.Pars.fUseMaxFF = 0, 0, 0
	cexReviewMustPanic(t, func() { Cbs0_ManSolve(p, Gia_ObjFromLit(g, root^1)) })
	Cbs0_ManStop(p)
	for _, obj := range g.pObjs {
		if obj.fMark0 != 0 || obj.fMark1 != 0 {
			t.Fatal("stop after invalid heuristic retained marks")
		}
	}
}

func TestChronologicalRejectsNonzeroQueueHeads(t *testing.T) {
	for _, propagation := range []bool{false, true} {
		g := Gia_ManStart(16)
		a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
		root := Gia_ManAppendAnd(g, a, b)
		p := Cbs0_ManAlloc(g)
		if propagation {
			p.pProp.iHead = 1
		} else {
			p.pJust.iHead = 1
		}
		cexReviewMustPanic(t, func() { Cbs0_ManSolve(p, Gia_ObjFromLit(g, root)) })
		Cbs0_ManStop(p)
	}
}
