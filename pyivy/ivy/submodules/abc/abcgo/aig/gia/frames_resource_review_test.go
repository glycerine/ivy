package gia

import (
	"fmt"
	"math/bits"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func TestFramesHistoryStopReleasesOwnedStorage(t *testing.T) {
	g := framesReviewGenerated(9)
	defer Gia_ManStop(g)
	p := Gia_ManUnrollStart(g, &Gia_ParFra_t{NFrames: 4, fSaveLastLit: 1}).(*Gia_ManUnr_t)
	order := p.pOrder
	q := Gia_ManUnrollAdd(p, 3).(*Gia_Man_t)
	defer Gia_ManStop(q)
	before := cofReviewValues(q, 7)
	if len(p.store) == 0 || len(p.delays) == 0 || len(p.first) == 0 {
		t.Fatal("missing live history storage")
	}
	Gia_ManUnrollStop(p)
	Gia_ManUnrollStop(p)
	if order.nObjs != 0 || p.pOrder != nil || p.pAig != nil || p.pPars != nil || p.pNew != nil || p.limits != nil || p.delays != nil || p.first != nil || p.store != nil {
		t.Fatal("stopped handle retains owned scheduler storage")
	}
	if Gia_ManUnrollAdd(p, 4) != nil || Gia_ManUnrollLastLit(p) != -1 {
		t.Fatal("stopped scheduler accepted work")
	}
	Gia_ManLevelNum(q)
	if !slices.Equal(before, cofReviewValues(q, 7)) {
		t.Fatal("stopping scheduler changed its returned graph")
	}
	// Source scratch can be reused for another unroller; the first returned
	// graph must remain independent of both the source and the new scheduler.
	p2 := Gia_ManUnrollStart(g, &Gia_ParFra_t{NFrames: 1}).(*Gia_ManUnr_t)
	q2 := Gia_ManUnrollAdd(p2, 1).(*Gia_Man_t)
	Gia_ManUnrollStop(p2)
	Gia_ManStop(q2)
	Gia_ManStop(g)
	if !slices.Equal(before, cofReviewValues(q, 7)) {
		t.Fatal("returned graph depends on reused/disposed source scratch")
	}
}

func TestFramesAbsPackedIDsAtPowersOfTwo(t *testing.T) {
	for _, size := range []int32{4, 8, 16, 32} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			g := Gia_ManStart(size)
			defer Gia_ManStop(g)
			r := Gia_ManAppendCi(g)
			Gia_ManAppendCo(g, r)
			Gia_ManAppendCo(g, r^1)
			Gia_ManSetRegNum(g, 1)
			for g.nObjs < size {
				Gia_ManAppendAnd(g, r, 0) // unreachable: changes ID width, not tents
			}
			shift := uint(bits.Len(uint(size - 1)))
			rows := Vec_PtrArray(Gia_ManUnrollAbs(g, 3))
			if len(rows) != 3 {
				t.Fatal("wrong abstraction frame count")
			}
			for frame, row := range rows {
				want := []int32{int32(frame)<<shift | r>>1}
				if got := Vec_IntArray(row.(*Vec_Int_t)); !slices.Equal(got, want) {
					t.Fatalf("frame%d packed IDs%v want%v", frame, got, want)
				}
			}
		})
	}
}

func TestFramesEmptyBoundaryDrivers(t *testing.T) {
	for _, init := range []int32{0, 1} {
		g := Gia_ManStart(1)
		out := Gia_ManFrames(g, &Gia_ParFra_t{NFrames: 3, FInit: init})
		if out == nil || out.nObjs != 1 || Gia_ManCiNum(out) != 0 || Gia_ManCoNum(out) != 0 {
			t.Fatal("empty graph did not remain empty")
		}
		Gia_ManStop(out)
		out = Gia_ManFrames2(g, &Gia_ParFra_t{NFrames: 0, FInit: init})
		if out == nil || out.nObjs != 1 {
			t.Fatal("zero-frame empty combinational boundary failed")
		}
		Gia_ManStop(out)
		Gia_ManStop(g)
	}
}

func TestFramesCompleteWrapperChecksOutputContract(t *testing.T) {
	g := frameTestGraph()
	defer Gia_ManStop(g)
	defer func() {
		if recover() == nil {
			t.Fatal("complete wrapper accepted save mode without its required frame outputs")
		}
	}()
	q := Gia_ManUnroll(g, &Gia_ParFra_t{NFrames: 1, fSaveLastLit: 1})
	Gia_ManStop(q)
}

func TestFramesOriginalBoundaryAssertions(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frames_review.c")
	for _, tc := range []struct{ mode, assertion string }{{"--abs-power", "nObjMask"}, {"--complete-save", "Gia_ManPoNum"}} {
		t.Run(tc.mode, func(t *testing.T) {
			cmd := exec.Command(exe, tc.mode)
			cmd.Dir = t.TempDir()
			out, err := cmd.CombinedOutput()
			if err == nil || !strings.Contains(string(out), "Assertion") || !strings.Contains(string(out), tc.assertion) {
				t.Fatalf("expected original C boundary assertion, got%v\n%s", err, out)
			}
		})
	}
}

func TestFramesSmallStreamingHashTableGrows(t *testing.T) {
	g := framesReviewDependencyGraph()
	defer Gia_ManStop(g)
	p := Gia_ManUnrollStart(g, &Gia_ParFra_t{NFrames: 512, fSaveLastLit: 1}).(*Gia_ManUnr_t)
	defer Gia_ManStop(p.pNew)
	defer Gia_ManUnrollStop(p)
	initial := Vec_IntSize(&p.pNew.vHTable)
	q := Gia_ManUnrollAdd(p, 512).(*Gia_Man_t)
	gates, buckets := int(Gia_ManAndNum(q)), Vec_IntSize(&q.vHTable)
	if gates <= 2*initial {
		t.Fatal("fixture did not outgrow its small initial hash table")
	}
	if gates > 2*buckets {
		t.Fatalf("streaming table failed to grow: %d gates in %d buckets (initial%d)", gates, buckets, initial)
	}
}
