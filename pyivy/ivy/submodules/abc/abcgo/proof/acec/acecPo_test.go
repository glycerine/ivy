package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"reflect"
	"testing"
)

func poRows(w *vec.Vec_Wec_t) [][]int32 {
	out := make([][]int32, vec.Vec_WecSize(w))
	for i := range out {
		out[i] = append([]int32(nil), vec.Vec_IntArray(vec.Vec_WecEntry(w, i))...)
	}
	return out
}
func TestSignatureParsingAndDistribution(t *testing.T) {
	got := Acec_ParseSignature("(4*o1+2*o2)*(3*i0+1*i1)+(5*o0)")
	want := [][]int32{{1, -1, 8}, {1, -2, 6}, {2, -1, 6}, {2, -2, 4}, {0, 6}}
	if !reflect.DeepEqual(poRows(got), want) {
		t.Fatalf("got %v want %v", poRows(got), want)
	}
	if Acec_ParseSignature("(1*o0)*(1*i0") == nil {
	} else {
		t.Fatal("accepted unterminated signature")
	}
}
func polynomialValue(p *vec.Vec_Wec_t, inputs map[int32]bool) int64 {
	var sum int64
	for i := 0; i < vec.Vec_WecSize(p); i += 2 {
		coef := vec.Vec_IntArray(vec.Vec_WecEntry(p, i))
		vars := vec.Vec_IntArray(vec.Vec_WecEntry(p, i+1))
		active := true
		for _, v := range vars {
			if !inputs[v] {
				active = false
				break
			}
		}
		if !active {
			continue
		}
		for _, v := range coef {
			n := int64(1) << uint(abs32(v)-1)
			if v < 0 {
				n = -n
			}
			sum += n
		}
	}
	return sum
}
func TestPolynomialBothBuildersAllAndPolarities(t *testing.T) {
	for flags := 0; flags < 8; flags++ {
		g := gia.Gia_ManStart(20)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		left, right := a, b
		if flags&1 != 0 {
			left ^= 1
		}
		if flags&2 != 0 {
			right ^= 1
		}
		root := gia.Gia_ManAppendAnd(g, left, right)
		if flags&4 != 0 {
			root ^= 1
		}
		roots := vec.Vec_IntAllocArrayCopy([]int32{root}, 1)
		leaves := vec.Vec_IntAllocArrayCopy([]int32{a >> 1, b >> 1}, 2)
		nodes := vec.Vec_IntAllocArrayCopy([]int32{(root &^ 1) >> 1}, 1)
		objectPoly := Gia_PolynBuildNew(g, nil, roots, 0, leaves, nodes, 0, 0, 0)
		literalPoly := Gia_PolynBuildNew2(g, roots, 0, leaves, nodes, 0, 0, 0)
		for mask := 0; mask < 4; mask++ {
			x, y := mask&1 != 0, mask&2 != 0
			xx, yy := x, y
			if flags&1 != 0 {
				xx = !xx
			}
			if flags&2 != 0 {
				yy = !yy
			}
			want := xx && yy
			if flags&4 != 0 {
				want = !want
			}
			expected := int64(0)
			if want {
				expected = 1
			}
			if got := polynomialValue(objectPoly, map[int32]bool{a >> 1: x, b >> 1: y}); got != expected {
				t.Fatalf("object flags=%d mask=%d got=%d want=%d poly=%v", flags, mask, got, expected, poRows(objectPoly))
			}
			if got := polynomialValue(literalPoly, map[int32]bool{a: x, b: y}); got != expected {
				t.Fatalf("literal flags=%d mask=%d got=%d want=%d poly=%v", flags, mask, got, expected, poRows(literalPoly))
			}
		}
	}
}

func TestPolynomialWeightedOutputsAndCancellation(t *testing.T) {
	g := gia.Gia_ManStart(10)
	a := gia.Gia_ManAppendCi(g)
	leaves := vec.Vec_IntAllocArrayCopy([]int32{a >> 1}, 1)
	nodes := vec.Vec_IntAlloc(0)
	for _, test := range []struct {
		roots  []int32
		signed int32
		want   int64
	}{
		{[]int32{a, a}, 0, 3}, {[]int32{a, a}, 1, -1}, {[]int32{a, a ^ 1}, 0, -1},
	} {
		roots := vec.Vec_IntAllocArrayCopy(test.roots, len(test.roots))
		for _, literal := range []bool{false, true} {
			var poly *vec.Vec_Wec_t
			key := a >> 1
			if literal {
				poly = Gia_PolynBuildNew2(g, roots, 0, leaves, nodes, test.signed, 0, 0)
				key = a
			} else {
				poly = Gia_PolynBuildNew(g, nil, roots, 0, leaves, nodes, test.signed, 0, 0)
			}
			// The all-zero assignment may include a constant due to complemented roots;
			// the difference between assignments is the expected linear coefficient.
			delta := polynomialValue(poly, map[int32]bool{key: true}) - polynomialValue(poly, map[int32]bool{key: false})
			if delta != test.want {
				t.Fatalf("roots=%v signed=%d literal=%v got=%d want=%d terms=%v", test.roots, test.signed, literal, delta, test.want, poRows(poly))
			}
		}
	}
}
func TestPolynomialBothBuildersRealXorPolarities(t *testing.T) {
	for flags := 0; flags < 8; flags++ {
		g := gia.Gia_ManStart(16)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		root := gia.Gia_ManAppendXorReal(g, a^int32(flags&1), b^int32(flags>>1&1)) ^ int32(flags>>2&1)
		roots := vec.Vec_IntAllocArrayCopy([]int32{root}, 1)
		leaves := vec.Vec_IntAllocArrayCopy([]int32{a >> 1, b >> 1}, 2)
		nodes := vec.Vec_IntAllocArrayCopy([]int32{root >> 1}, 1)
		obj := Gia_PolynBuildNew(g, nil, roots, 0, leaves, nodes, 0, 0, 0)
		lit := Gia_PolynBuildNew2(g, roots, 0, leaves, nodes, 0, 0, 0)
		for mask := 0; mask < 4; mask++ {
			x, y := mask&1 != 0, mask&2 != 0
			want := (x != (flags&1 != 0)) != (y != (flags&2 != 0))
			if flags&4 != 0 {
				want = !want
			}
			expected := int64(0)
			if want {
				expected = 1
			}
			if got := polynomialValue(obj, map[int32]bool{a >> 1: x, b >> 1: y}); got != expected {
				t.Fatalf("object flags %d mask %d got %d want %d", flags, mask, got, expected)
			}
			if got := polynomialValue(lit, map[int32]bool{a: x, b: y}); got != expected {
				t.Fatalf("literal flags %d mask %d got %d want %d", flags, mask, got, expected)
			}
		}
	}
}
