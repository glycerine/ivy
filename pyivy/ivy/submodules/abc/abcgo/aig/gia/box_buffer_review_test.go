package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"github.com/glycerine/abcgo/misc/tim"
	"github.com/glycerine/abcgo/misc/vec"
	"slices"
	"strings"
	"testing"
)

func boxBufferReviewGraph(signs int, future bool) *Gia_Man_t {
	g := Gia_ManStart(1)
	a, b, u, w := Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	a ^= int32(signs & 1)
	alternate := a
	if future {
		yes, no := Gia_ManAppendAnd(g, a, w), Gia_ManAppendAnd(g, a, w^1)
		alternate = Gia_ManAppendAnd(g, yes^1, no^1) ^ 1 // a, with a structural future-box dependency
	}
	sibling := Gia_ManAppendBuf(g, alternate)
	root := Gia_ManAppendBuf(g, a)
	out := Gia_ManAppendAnd(g, u, b^int32(signs>>1&1))
	Gia_ManAppendCo(g, root)
	Gia_ManAppendCo(g, u)
	Gia_ManAppendCo(g, out)
	g.pSibls = make([]int32, g.nObjs)
	g.pSibls[root>>1] = sibling >> 1
	timing := tim.Tim_ManStart(4, 3)
	tim.Tim_ManCreateBox(timing, 0, 1, 2, 1, -1, 0)
	tim.Tim_ManCreateBox(timing, 1, 1, 3, 1, -1, 0)
	g.PManTime = timing
	return g
}

func TestBoxBufferTraversalIgnoresChoiceSiblings(t *testing.T) {
	for _, future := range []bool{false, true} {
		g := boxBufferReviewGraph(0, future)
		root := Gia_ObjFanin0(Gia_ManCo(g, 0)).id
		external := Gia_ObjFanin0(Gia_ManCo(g, 2)).id
		want := []int32{0, Gia_ManCi(g, 0).id, Gia_ManCi(g, 1).id, root, Gia_ManCo(g, 0).id, Gia_ManCi(g, 2).id, Gia_ManCo(g, 1).id, Gia_ManCi(g, 3).id, external, Gia_ManCo(g, 2).id}
		got := Gia_ManOrderWithBoxes(g)
		if got == nil || !slices.Equal(vec.Vec_IntArray(got), want) {
			t.Fatalf("future%v: order %v, want %v", future, got, want)
		}
	}
}

func TestBoxBufferNativeOracle(t *testing.T) {
	exe := compileGiaSourceOracle(t, "box_buffer_order.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native: %v\n%s", err, native)
	}
	var got strings.Builder
	for signs := 0; signs < 4; signs++ {
		for _, future := range []bool{false, true} {
			g := boxBufferReviewGraph(signs, future)
			order := Gia_ManOrderWithBoxes(g)
			if order == nil {
				t.Fatalf("C accepts signs%d future%v; Go rejects it", signs, future)
			}
			fmt.Fprintf(&got, "case %d %d order", signs, giaBool(future))
			for _, id := range vec.Vec_IntArray(order) {
				fmt.Fprintf(&got, " %d", id)
			}
			got.WriteString(" truth")
			root := Gia_ObjFanin0(Gia_ManCo(g, 0)).id
			sibling := g.pSibls[root]
			for mask := 0; mask < 16; mask++ {
				inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				nodes, outputs, err := g.evaluateObjects(inputs)
				if err != nil {
					t.Fatal(err)
				}
				if nodes[root] != nodes[sibling] {
					t.Fatal("fixture buffer alternatives are not equivalent")
				}
				packed := 0
				for i, value := range outputs {
					if value {
						packed |= 1 << i
					}
				}
				fmt.Fprintf(&got, " %d", packed)
			}
			got.WriteByte('\n')
		}
	}
	if got.String() != string(native) {
		t.Fatalf("C:\n%sGo:\n%s", native, got.String())
	}
}
