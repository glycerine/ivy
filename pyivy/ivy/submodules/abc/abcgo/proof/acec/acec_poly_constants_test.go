package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestPolynomialBuildersConstantOutputs(t *testing.T) {
	g := gia.Gia_ManStart(8)
	leaves, nodes := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	for lit := int32(0); lit <= 1; lit++ {
		roots := vec.Vec_IntAllocArrayCopy([]int32{lit}, 1)
		for _, p := range []*vec.Vec_Wec_t{Gia_PolynBuildNew(g, nil, roots, 0, leaves, nodes, 0, 0, 0), Gia_PolynBuildNew2(g, roots, 0, leaves, nodes, 0, 0, 0)} {
			if got := polynomialValue(p, nil); got != int64(lit) {
				t.Fatalf("constant literal %d got %d terms=%v", lit, got, poRows(p))
			}
		}
	}
}

func TestPolynomialBuildersConstantGateInputs(t *testing.T) {
	for _, xor := range []bool{false, true} {
		for constant := int32(0); constant <= 1; constant++ {
			g := gia.Gia_ManStart(8)
			a := gia.Gia_ManAppendCi(g)
			var root int32
			if xor {
				root = gia.Gia_ManAppendXorReal(g, a, constant)
			} else {
				root = gia.Gia_ManAppendAnd(g, a, constant)
			}
			roots := vec.Vec_IntAllocArrayCopy([]int32{root}, 1)
			leaves := vec.Vec_IntAllocArrayCopy([]int32{a >> 1}, 1)
			nodes := vec.Vec_IntAllocArrayCopy([]int32{root >> 1}, 1)
			object := Gia_PolynBuildNew(g, nil, roots, 0, leaves, nodes, 0, 0, 0)
			literal := Gia_PolynBuildNew2(g, roots, 0, leaves, nodes, 0, 0, 0)
			gia.Gia_ManAppendCo(g, root)
			for _, input := range []bool{false, true} {
				want, err := g.Evaluate([]bool{input})
				if err != nil {
					t.Fatal(err)
				}
				expected := int64(0)
				if want[0] {
					expected = 1
				}
				if got := polynomialValue(object, map[int32]bool{a >> 1: input}); got != expected {
					t.Fatalf("object XOR=%v const=%d input=%v got=%d want=%d", xor, constant, input, got, expected)
				}
				if got := polynomialValue(literal, map[int32]bool{a: input}); got != expected {
					t.Fatalf("literal XOR=%v const=%d input=%v got=%d want=%d", xor, constant, input, got, expected)
				}
			}
		}
	}
}
