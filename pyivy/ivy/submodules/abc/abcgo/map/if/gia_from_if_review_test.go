package if_

import (
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestGiaFromIfAigRecTruthAndMemoization(t *testing.T) {
	for seed := uint32(1); seed <= 64; seed++ {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			p := If_ManStart(&If_Par_t{NLutSize: 4, NCutsMax: 4})
			defer If_ManStop(p)
			pool := []*If_Obj_t{If_ManCreateCi(p), If_ManCreateCi(p), If_ManCreateCi(p), If_ManCreateCi(p)}
			state := seed
			next := func() uint32 { state = 1664525*state + 1013904223; return state }
			for i := 0; i < 12; i++ {
				a, b := pool[int(next()%uint32(len(pool)))], pool[int(next()%uint32(len(pool)))]
				pool = append(pool, If_ManCreateAnd(p, If_NotCond(a, int32(next()&1)), If_NotCond(b, int32(next()&1))))
			}
			root := pool[len(pool)-1]
			If_ManCreateCo(p, root)
			If_ManCreateCo(p, If_Not(root))
			If_ManCreateCo(p, p.pConst1)
			dst := gia.Gia_ManStart(32)
			defer gia.Gia_ManStop(dst)
			gia.Gia_ManHashAlloc(dst)
			p.pConst1.iCopy = 1
			for i := int32(0); i < 4; i++ {
				If_ManCi(p, i).iCopy = gia.Gia_ManAppendCi(dst)
			}
			lit := Gia_ManFromIfAigRec(dst, If_Regular(root)) ^ If_IsComplement(root)
			before := gia.Gia_ManObjNum(dst)
			if again := Gia_ManFromIfAigRec(dst, If_Regular(root)) ^ If_IsComplement(root); again != lit || gia.Gia_ManObjNum(dst) != before {
				t.Fatal("repeated cone reconstruction changed result or allocated duplicate nodes")
			}
			gia.Gia_ManAppendCo(dst, lit)
			gia.Gia_ManAppendCo(dst, lit^1)
			gia.Gia_ManAppendCo(dst, 1)
			for assignment := 0; assignment < 16; assignment++ {
				input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
				want := bridgeReviewEvaluate(p, input)
				got, err := dst.Evaluate(input)
				if err != nil || !slices.Equal(got, want) {
					t.Fatalf("assignment%d got%v want%v err%v", assignment, got, want, err)
				}
			}
		})
	}
}

func fromIfReviewRecordedManager() *If_Man_t {
	pars := &If_Par_t{NLutSize: 4, NCutsMax: 4, FUserRecLib: 1}
	pars.SetRecordedGiaBuilder(func(dst *Gia_Man_t, _ *If_Man_t, _ *If_Cut_t, leaves *Vec_Int_t, hash int32) int32 {
		return gia.Gia_ManAppendAnd(dst, vec.Vec_IntEntry(leaves, 0), vec.Vec_IntEntry(leaves, 1))
	})
	p := If_ManStart(pars)
	a, b := If_ManCreateCi(p), If_ManCreateCi(p)
	root := If_ManCreateAnd(p, a, b)
	root.CutBest = If_Cut_t{nLeaves: 2, pLeaves: []int32{a.Id, b.Id}}
	If_ManCreateCo(p, root)
	If_ManCreateCo(p, If_Not(root))
	If_ManCreateCo(p, p.pConst1)
	return p
}

func TestGiaFromIfRecordedReconstructionCanRepeat(t *testing.T) {
	p := fromIfReviewRecordedManager()
	defer If_ManStop(p)
	for repeat := 0; repeat < 3; repeat++ {
		dst := Gia_ManFromIfAig(p)
		for assignment := 0; assignment < 4; assignment++ {
			input := []bool{assignment&1 != 0, assignment&2 != 0}
			got, err := dst.Evaluate(input)
			want := []bool{input[0] && input[1], !(input[0] && input[1]), true}
			if err != nil || !slices.Equal(got, want) {
				t.Fatalf("repeat%d assignment%d got%v want%v err%v", repeat, assignment, got, want, err)
			}
		}
		gia.Gia_ManStop(dst)
	}
}

func BenchmarkGiaFromIfRecordedScratch(b *testing.B) {
	p := fromIfReviewRecordedManager()
	defer If_ManStop(p)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := Gia_ManFromIfAig(p)
		gia.Gia_ManStop(dst)
	}
}

func TestGiaFromIfDsdLargerMiniGraph(t *testing.T) {
	// Like the source, this request needs a manager with at least seven inputs.
	balanceReviewIsolateFrame(t)
	src := gia.Gia_ManStart(64)
	defer gia.Gia_ManStop(src)
	var root int32
	for pin := 0; pin < 7; pin++ {
		input := gia.Gia_ManAppendCi(src)
		if pin == 0 {
			root = input
		} else {
			root = gia.Gia_ManAppendXor(src, root, input)
		}
	}
	gia.Gia_ManAppendCo(src, root)
	var result *gia.Gia_Man_t
	dsdReportCapture(t, func() { result = Gia_ManPerformDsdBalance(src, 7, 8, 0, 0) })
	if result == nil {
		t.Fatal("parity balancing failed")
	}
	defer gia.Gia_ManStop(result)
	for assignment := 0; assignment < 128; assignment++ {
		input := make([]bool, 7)
		want := false
		for pin := range input {
			input[pin] = assignment>>pin&1 != 0
			want = want != input[pin]
		}
		got, err := result.Evaluate(input)
		if err != nil || !slices.Equal(got, []bool{want}) {
			t.Fatalf("parity assignment%d got%v want%t err%v", assignment, got, want, err)
		}
	}
}
