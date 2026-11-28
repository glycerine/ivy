package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func TestFramesLongNativeHistory(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frames_review.c")
	native, err := testoracle.Command(t, exe, "--long-history").CombinedOutput()
	if err != nil {
		t.Fatalf("C long history: %v\n%s", err, native)
	}
	g := Gia_ManStart(64)
	defer Gia_ManStop(g)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	r := make([]int32, 12)
	for i := range r {
		r[i] = Gia_ManAppendCi(g)
	}
	Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, r[0]))
	for i := 0; i < 11; i++ {
		Gia_ManAppendCo(g, Gia_ManAppendAnd(g, a, r[i+1]))
	}
	Gia_ManAppendCo(g, b)
	Gia_ManSetRegNum(g, 12)
	p := Gia_ManUnrollStart(g, &Gia_ParFra_t{NFrames: 16, fSaveLastLit: 1}).(*Gia_ManUnr_t)
	defer Gia_ManStop(p.pNew)
	defer Gia_ManUnrollStop(p)
	if size := p.store[p.first[g.pObjs[a>>1].Value>>1]]; size != 12 {
		t.Fatalf("expected twelve-cell cross-tent PI history, got%d", size)
	}
	var got strings.Builder
	vector := func(tag string, values []int32) {
		got.WriteString(tag)
		for _, v := range values {
			fmt.Fprintf(&got, " %d", v)
		}
		got.WriteByte('\n')
	}
	snapshot := func() {
		vector("limits", p.limits)
		vector("delays", p.delays)
		got.WriteString("first")
		for _, offset := range p.first {
			fmt.Fprintf(&got, " %d", offset)
		}
		got.WriteByte('\n')
		vector("history", p.store)
	}
	got.WriteString("long step0\n")
	snapshot()
	framesReviewDump(&got, p.pOrder, "ordered")
	var rng uint64 = 17
	for step := int32(1); step <= 16; step++ {
		q := Gia_ManUnrollAdd(p, step).(*Gia_Man_t)
		fmt.Fprintf(&got, "long step%d lit%d\n", step, p.LastLit)
		snapshot()
		framesReviewDump(&got, q, "step")
		for trace := 0; trace < 32; trace++ {
			inputs := make([]bool, Gia_ManCiNum(q))
			for i := range inputs {
				rng = rng*6364136223846793005 + 1442695040888963407
				inputs[i] = rng>>63 != 0
			}
			state := append([]bool(nil), inputs[:12]...)
			last := false
			for f := int32(0); f < step; f++ {
				piA, piB := inputs[12+2*f], inputs[13+2*f]
				last = piA && state[0]
				next := make([]bool, 12)
				for i := 0; i < 11; i++ {
					next[i] = piA && state[i+1]
				}
				next[11] = piB
				state = next
			}
			values, _, err := q.evaluateObjects(inputs)
			if err != nil {
				t.Fatal(err)
			}
			if actual := values[p.LastLit>>1] != (p.LastLit&1 != 0); actual != last {
				t.Fatalf("step%d trace%d changed the streamed property", step, trace)
			}
		}
	}
	esopReviewCompare(t, native, got.String())
}
