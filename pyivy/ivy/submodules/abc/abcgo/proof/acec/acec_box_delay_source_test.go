package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"reflect"
	"testing"
)

// Values are from test_vectors/acec_box_delay.c linked against ABC.
func TestSourceArtificialBoxTimingVariants(t *testing.T) {
	for _, invert := range []bool{false, true} {
		for _, fan := range []int32{0, 1} {
			for _, trick := range []int32{0, 1} {
				g, x, y, z := carryFixture(invert)
				markCarry(g, y, x)
				markCarry(g, z, y)
				p := Gia_ManDupWithArtificalFaddBoxes(g, fan, trick)
				if gia.Gia_ManCiNum(p) != 8 || gia.Gia_ManCoNum(p) != 8 || gia.Gia_ManBoxNum(p) != 2 {
					t.Fatalf("invert=%v fan=%d trick=%d: ci=%d co=%d boxes=%d", invert, fan, trick, gia.Gia_ManCiNum(p), gia.Gia_ManCoNum(p), gia.Gia_ManBoxNum(p))
				}
				want, ignore := int32(120), int32(100)
				if fan == 0 && trick == 0 {
					want, ignore = 200, 200
				}
				if fan == 0 && trick == 1 {
					want = 300
				}
				var boxes int32
				if got := Gia_ManFindAnnotatedDelay(p, 10, &boxes, 0); got != want || boxes != 0 {
					t.Fatalf("invert=%v fan=%d trick=%d delay=%d boxes=%d want=%d,0", invert, fan, trick, got, boxes, want)
				}
				if got := Gia_ManFindAnnotatedDelay(p, 10, &boxes, 1); got != ignore || boxes != 0 {
					t.Fatalf("invert=%v fan=%d trick=%d ignored delay=%d boxes=%d want=%d,0", invert, fan, trick, got, boxes, ignore)
				}
			}
		}
	}
}

func TestSourceSevenBoxCarryChain(t *testing.T) {
	for _, fan := range []int32{0, 1} {
		for _, trick := range []int32{0, 1} {
			g := gia.Gia_ManStart(128)
			in := make([]int32, 9)
			for i := range in {
				in[i] = gia.Gia_ManAppendCi(g)
			}
			prior := gia.Gia_ManAppendAnd(g, in[0], in[1])
			penultimate := prior
			for i := 2; i < 9; i++ {
				node := gia.Gia_ManAppendAnd(g, prior, in[i])
				markCarry(g, node>>1, prior>>1)
				penultimate = prior
				prior = node
			}
			gia.Gia_ManAppendCo(g, prior)
			gia.Gia_ManAppendCo(g, penultimate)
			p := Gia_ManDupWithArtificalFaddBoxes(g, fan, trick)
			if gia.Gia_ManCiNum(p) != 23 || gia.Gia_ManCoNum(p) != 23 || gia.Gia_ManBoxNum(p) != 7 {
				t.Fatalf("fan=%d trick=%d: ci=%d co=%d boxes=%d", fan, trick, gia.Gia_ManCiNum(p), gia.Gia_ManCoNum(p), gia.Gia_ManBoxNum(p))
			}
			want := int32(170)
			if fan == 0 && trick == 0 {
				want = 250
			}
			if fan == 0 && trick == 1 {
				want = 350
			}
			var count int32
			if got := Gia_ManFindAnnotatedDelay(p, 10, &count, 0); got != want || count != 0 {
				t.Fatalf("fan=%d trick=%d delay=%d count=%d, C=%d,0", fan, trick, got, count, want)
			}
			if got := Gia_ManFindAnnotatedDelay(p, 10, &count, 1); got != 100 || count != 0 {
				t.Fatalf("fan=%d trick=%d ignored delay=%d count=%d", fan, trick, got, count)
			}
			_, extra := p.BoxTiming()
			collapsed := gia.Gia_ManDupCollapse(p, extra, nil, 0)
			for mask := 0; mask < 512; mask++ {
				inputs := make([]bool, 9)
				for i := range inputs {
					inputs[i] = mask&(1<<i) != 0
				}
				a, e1 := g.Evaluate(inputs)
				b, e2 := collapsed.Evaluate(inputs)
				if e1 != nil || e2 != nil || !reflect.DeepEqual(a, b) {
					t.Fatalf("fan=%d trick=%d input=%d original=%v collapsed=%v errors=%v %v", fan, trick, mask, a, b, e1, e2)
				}
			}
		}
	}
}
