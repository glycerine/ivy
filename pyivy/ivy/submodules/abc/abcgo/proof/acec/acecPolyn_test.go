package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func plnEval(p *Pln_Man_t, values map[int32]bool) int64 {
	var result int64
	for id, coef := range vec.Vec_IntArray(p.coefs) {
		if coef == 0 {
			continue
		}
		active := true
		for _, v := range vec.Hsh_VecReadArray(p.monomials, int32(id)) {
			if !values[v] {
				active = false
				break
			}
		}
		if !active {
			continue
		}
		for _, c := range vec.Hsh_VecReadArray(p.constants, coef) {
			x := int64(1) << uint(abs32(c)-1)
			if c < 0 {
				x = -x
			}
			result += x
		}
	}
	return result
}
func TestPriorityPolynomialExpansion(t *testing.T) {
	for flags := 0; flags < 4; flags++ {
		g := gia.Gia_ManStart(20)
		a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x, y := a, b
		if flags&1 != 0 {
			x ^= 1
		}
		if flags&2 != 0 {
			y ^= 1
		}
		root := gia.Gia_ManAppendAnd(g, x, y)
		p := Pln_ManAlloc(g, nil)
		vec.Vec_IntFill(p.tempC[0], 1, 1)
		vec.Vec_IntFill(p.tempM[0], 1, root>>1)
		Gia_PolynBuildAdd(p, p.tempC[0], p.tempM[0])
		for vec.Vec_QueSize(p.queue) > 0 {
			Gia_PolynBuildOne(p, int32(vec.Vec_QuePop(p.queue)))
		}
		for mask := 0; mask < 4; mask++ {
			xx, yy := mask&1 != 0, mask&2 != 0
			aValue, bValue := xx, yy
			if flags&1 != 0 {
				aValue = !aValue
			}
			if flags&2 != 0 {
				bValue = !bValue
			}
			want := int64(0)
			if aValue && bValue {
				want = 1
			}
			if got := plnEval(p, map[int32]bool{a >> 1: xx, b >> 1: yy}); got != want {
				t.Fatalf("flags=%d mask=%d got=%d want=%d", flags, mask, got, want)
			}
		}
	}
}
func TestPriorityPolynomialXor(t *testing.T) {
	g := gia.Gia_ManStart(20)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	root := gia.Gia_ManAppendXorReal(g, a, b)
	p := Pln_ManAlloc(g, nil)
	vec.Vec_IntFill(p.tempC[0], 1, 1)
	vec.Vec_IntFill(p.tempM[0], 1, root>>1)
	Gia_PolynBuildAdd(p, p.tempC[0], p.tempM[0])
	for vec.Vec_QueSize(p.queue) > 0 {
		Gia_PolynBuildOne(p, int32(vec.Vec_QuePop(p.queue)))
	}
	for mask := 0; mask < 4; mask++ {
		x, y := mask&1 != 0, mask&2 != 0
		want := int64(0)
		if x != y {
			want = 1
		}
		if got := plnEval(p, map[int32]bool{a >> 1: x, b >> 1: y}); got != want {
			t.Fatalf("mask=%d got=%d want=%d", mask, got, want)
		}
	}
}
func TestPolynomialCoefficientCarryAndCancel(t *testing.T) {
	v := vec.Vec_IntAlloc(4)
	Gia_PolynMergeConstOne(v, 1)
	Gia_PolynMergeConstOne(v, 1)
	if vec.Vec_IntSize(v) != 1 || vec.Vec_IntEntry(v, 0) != 2 {
		t.Fatalf("carry: %v", vec.Vec_IntArray(v))
	}
	Gia_PolynMergeConstOne(v, -2)
	if vec.Vec_IntSize(v) != 0 {
		t.Fatalf("cancel: %v", vec.Vec_IntArray(v))
	}
}
