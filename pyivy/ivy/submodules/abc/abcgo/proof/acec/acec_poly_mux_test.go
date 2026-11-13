package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestPolynomialBuildersNativeMux(t *testing.T) {
	for flags := 0; flags < 16; flags++ {
		p := gia.Gia_ManStart(16)
		p.EnableMuxes()
		a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
		root := gia.Gia_ManAppendMuxReal(p, a^int32(flags&1), b^int32(flags>>1&1), c^int32(flags>>2&1)) ^ int32(flags>>3&1)
		roots := vec.Vec_IntAllocArrayCopy([]int32{root}, 1)
		leaves := vec.Vec_IntAllocArrayCopy([]int32{a >> 1, b >> 1, c >> 1}, 3)
		nodes := vec.Vec_IntAllocArrayCopy([]int32{root >> 1}, 1)
		object := Gia_PolynBuildNew(p, nil, roots, 0, leaves, nodes, 0, 0, 0)
		literal := Gia_PolynBuildNew2(p, roots, 0, leaves, nodes, 0, 0, 0)
		gia.Gia_ManAppendCo(p, root)
		for mask := 0; mask < 8; mask++ {
			input := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			result, err := p.Evaluate(input)
			if err != nil {
				t.Fatal(err)
			}
			want := int64(0)
			if result[0] {
				want = 1
			}
			if got := polynomialValue(object, map[int32]bool{a >> 1: input[0], b >> 1: input[1], c >> 1: input[2]}); got != want {
				t.Fatalf("object flags %d mask %d got %d want %d", flags, mask, got, want)
			}
			if got := polynomialValue(literal, map[int32]bool{a: input[0], b: input[1], c: input[2]}); got != want {
				t.Fatalf("literal flags %d mask %d got %d want %d", flags, mask, got, want)
			}
		}
	}
}
