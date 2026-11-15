package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func framesReviewDependencyGraph() *Gia_Man_t {
	g := Gia_ManStart(64)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	r0, r1, r2 := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, r0)
	y := Gia_ManAppendAnd(g, b, r1)
	z := Gia_ManAppendAnd(g, x, r2^1)
	u := Gia_ManAppendAnd(g, y, r0)
	v := Gia_ManAppendAnd(g, a, r1^1)
	w := Gia_ManAppendAnd(g, v, b)
	for _, lit := range []int32{x, z, y, u, w} {
		Gia_ManAppendCo(g, lit)
	}
	Gia_ManSetRegNum(g, 3)
	return g
}

func TestFramesDependencyTents(t *testing.T) {
	g := framesReviewDependencyGraph()
	defer Gia_ManStop(g)
	limits := Vec_IntAlloc(0)
	q := Gia_ManUnrollDup(g, limits)
	defer Gia_ManStop(q)
	if got, want := Vec_IntArray(limits), []int32{1, 8, 15, 17}; !slices.Equal(got, want) {
		t.Fatalf("limits%v want%v", got, want)
	}
}

func TestFramesSparseAbstraction(t *testing.T) {
	g := framesReviewDependencyGraph()
	defer Gia_ManStop(g)
	abs := Gia_ManUnrollAbs(g, 3)
	want := [][]int32{{8, 6, 5, 3}, {40, 38, 37, 35, 11, 10, 7, 4}, {72, 70, 69, 67, 43, 42, 39, 36, 9}}
	for i, row := range Vec_PtrArray(abs) {
		if got := Vec_IntArray(row.(*Vec_Int_t)); !slices.Equal(got, want[i]) {
			t.Fatalf("row%d got%v want%v", i, got, want[i])
		}
	}
}

func TestFrames2SourceCombinationalContract(t *testing.T) {
	g := framesReviewDependencyGraph()
	defer Gia_ManStop(g)
	q := Gia_ManFrames2(g, &Gia_ParFra_t{NFrames: 3})
	defer Gia_ManStop(q)
	if Gia_ManCiNum(q) != 9 || Gia_ManCoNum(q) != 6 || q.nRegs != 0 {
		t.Fatalf("ci%d co%d regs%d want9/6/0", Gia_ManCiNum(q), Gia_ManCoNum(q), q.nRegs)
	}
}

func TestFramesNativeDependencyContract(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frames_review.c")
	for _, mode := range []string{"order", "--abs", "--frames2", "--incremental", "--fixed"} {
		t.Run(mode, func(t *testing.T) {
			var args []string
			if mode != "order" {
				args = []string{mode}
			}
			native, err := exec.Command(exe, args...).CombinedOutput()
			if err != nil {
				t.Fatalf("C unrolling probe: %v\n%s", err, native)
			}
			g := framesReviewDependencyGraph()
			defer Gia_ManStop(g)
			var got strings.Builder
			dump := func(q *Gia_Man_t, tag string) {
				fmt.Fprintf(&got, "%s ci%d co%d regs%d objs%d\n", tag, Gia_ManCiNum(q), Gia_ManCoNum(q), q.nRegs, q.nObjs)
				for id, o := range q.pObjs {
					fmt.Fprintf(&got, "%d %d %d %d %d %d %d\n", id, o.iDiff0, o.iDiff1, o.fCompl0, o.fCompl1, o.fTerm, o.Value)
				}
			}
			switch mode {
			case "order":
				limits := Vec_IntAlloc(0)
				q := Gia_ManUnrollDup(g, limits)
				defer Gia_ManStop(q)
				got.WriteString("limits")
				for _, id := range Vec_IntArray(limits) {
					fmt.Fprintf(&got, " %d", id)
				}
				got.WriteByte('\n')
				dump(q, "ordered")
			case "--abs":
				for i, row := range Vec_PtrArray(Gia_ManUnrollAbs(g, 3)) {
					fmt.Fprintf(&got, "abs %d", i)
					for _, id := range Vec_IntArray(row.(*Vec_Int_t)) {
						fmt.Fprintf(&got, " %d", id)
					}
					got.WriteByte('\n')
				}
			case "--frames2":
				q := Gia_ManFrames2(g, &Gia_ParFra_t{NFrames: 3})
				defer Gia_ManStop(q)
				dump(q, "frames2")
			case "--incremental", "--fixed":
				save := int32(0)
				if mode == "--incremental" {
					save = 1
				}
				raw := Gia_ManUnrollStart(g, &Gia_ParFra_t{NFrames: 3, fSaveLastLit: save})
				defer Gia_ManUnrollStop(raw)
				var q *Gia_Man_t
				defer func() { Gia_ManStop(q) }()
				for f := int32(1); f <= 3; f++ {
					q = Gia_ManUnrollAdd(raw, f).(*Gia_Man_t)
					lit := Gia_ManUnrollLastLit(raw)
					fmt.Fprintf(&got, "step%d ci%d co%d objs%d lit%d\n", f, Gia_ManCiNum(q), Gia_ManCoNum(q), q.nObjs, lit)
					for bits := 0; bits < 1<<Gia_ManCiNum(q); bits++ {
						values := cofReviewValues(q, bits)
						if values[lit>>1] != (lit&1 != 0) {
							got.WriteByte('1')
						} else {
							got.WriteByte('0')
						}
					}
					got.WriteByte('\n')
				}
			}
			esopReviewCompare(t, native, got.String())
		})
	}
}

func TestFramesFixedModeHasNoSavedProperty(t *testing.T) {
	g := frameTestGraph()
	defer Gia_ManStop(g)
	raw := Gia_ManUnrollStart(g, &Gia_ParFra_t{NFrames: 2})
	q := Gia_ManUnrollAdd(raw, 1).(*Gia_Man_t)
	defer Gia_ManStop(q)
	defer Gia_ManUnrollStop(raw)
	if lit := Gia_ManUnrollLastLit(raw); lit != 0 {
		t.Fatalf("fixed mode saved property%d, source leaves LastLit0", lit)
	}
}

func TestFramesOriginalInitializedSaveIndexFailure(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frames_review.c")
	cmd := exec.Command(exe, "--initialized-save")
	cmd.Dir = t.TempDir()
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "Assertion") {
		t.Fatalf("expected original C PI-index assertion, got%v\n%s", err, out)
	}
	g := framesReviewDependencyGraph()
	defer Gia_ManStop(g)
	raw := Gia_ManUnrollStart(g, &Gia_ParFra_t{NFrames: 3, FInit: 1, fSaveLastLit: 1})
	q := Gia_ManUnrollAdd(raw, 1).(*Gia_Man_t)
	defer Gia_ManStop(q)
	defer Gia_ManUnrollStop(raw)
	if Gia_ManCiNum(q) != 2 || Gia_ManUnrollLastLit(raw) != 0 {
		t.Fatal("initialized first property should be zero with only two PI inputs")
	}
}

func TestFramesFixedZeroBoundRejectsAdd(t *testing.T) {
	g := frameTestGraph()
	defer Gia_ManStop(g)
	raw := Gia_ManUnrollStart(g, &Gia_ParFra_t{NFrames: 0})
	defer Gia_ManStop(raw.(*Gia_ManUnr_t).pNew)
	defer Gia_ManUnrollStop(raw)
	if Gia_ManUnrollAdd(raw, 1) != nil {
		t.Fatal("fixed zero-frame bound accepted a positive horizon")
	}
}

func TestFramesIncrementalZeroBoundStillExtends(t *testing.T) {
	g := frameTestGraph()
	defer Gia_ManStop(g)
	raw := Gia_ManUnrollStart(g, &Gia_ParFra_t{NFrames: 0, fSaveLastLit: 1})
	defer Gia_ManStop(raw.(*Gia_ManUnr_t).pNew)
	defer Gia_ManUnrollStop(raw)
	if q, ok := Gia_ManUnrollAdd(raw, 3).(*Gia_Man_t); !ok || q == nil || Gia_ManCiNum(q) != 4 {
		t.Fatal("zero-bound incremental mode did not extend through three frames")
	}
}
