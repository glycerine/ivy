package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestFaultFormulaCount(t *testing.T) {
	for _, tc := range []struct {
		form         string
		ok           bool
		vars, params int32
	}{
		{"((a&b)^p)", true, 2, 1}, {"(~a?(b&s):p)", true, 2, 4},
		{"", false, 0, 0}, {"(a&b", false, 0, 0}, {"(a&b))(", false, 0, 0},
		{"(a&b)", false, 2, 0}, {"(a&p)", false, 1, 1}, {"(a&b&~)", false, 0, 0}, {"(a&b&z)", false, 0, 0},
	} {
		var v, p int32
		ok := Gia_FormStrCount(tc.form, &v, &p) == 0
		if ok != tc.ok || ok && (v != tc.vars || p != tc.params) {
			t.Errorf("%q: ok=%v vars=%d params=%d", tc.form, ok, v, p)
		}
	}
}

func TestRealizeFaultFormula(t *testing.T) {
	for _, tc := range []struct {
		form string
		f    func(bool, bool, bool) bool
	}{
		{"((a&b)^p)", func(a, b, p bool) bool { return (a && b) != p }},
		{"(p?~a:b)", func(a, b, p bool) bool {
			if p {
				return !a
			}
			return b
		}},
		{"((a|p)&~b)", func(a, b, p bool) bool { return (a || p) && !b }},
	} {
		g := gia.Gia_ManStart(16)
		gia.Gia_ManHashStart(g)
		a, b, p := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		out := Gia_ManRealizeFormula(g, []int32{a, b}, []int32{p}, Gia_FormStrTransform(tc.form), 1)
		gia.Gia_ManAppendCo(g, out)
		gia.Gia_ManHashStop(g)
		for bits := 0; bits < 8; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			got, err := g.Evaluate(in)
			if err != nil || got[0] != tc.f(in[0], in[1], in[2]) {
				t.Fatalf("%s %v: %v %v", tc.form, in, got, err)
			}
		}
	}
}

func TestFormulaUnfold(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b, r := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	first := gia.Gia_ManAppendAnd(g, a, b)
	second := gia.Gia_ManAppendAnd(g, first, r)
	gia.Gia_ManAppendCo(g, first)
	gia.Gia_ManAppendCo(g, second)
	gia.Gia_ManSetRegNum(g, 1)
	for _, ff := range []int32{0, 1} {
		q := Gia_ManFormulaUnfold(g, "((a&b)^p)", ff)
		n := 5
		if ff != 0 {
			n = 4
		}
		if len(q.InputLiterals()) != n {
			t.Fatalf("ff=%d inputs=%d", ff, len(q.InputLiterals()))
		}
		for bits := 0; bits < 1<<n; bits++ {
			in := make([]bool, n)
			for i := range in {
				in[i] = bits&(1<<i) != 0
			}
			x := in[0] && in[1]
			if ff == 0 {
				x = x != in[3]
			}
			y := x && in[2]
			if ff == 0 {
				y = y != in[4]
			} else {
				y = y != in[3]
			}
			got, err := q.Evaluate(in)
			if err != nil || got[0] != x || got[1] != y {
				t.Fatalf("ff=%d in=%v got=%v err=%v want=%v,%v", ff, in, got, err, x, y)
			}
		}
	}
}

func TestDeriveFormula(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, a, b^1)^1)
	if got := Gia_DeriveFormula(g, []string{"a", "b"}); got != "((~a|b))" {
		t.Fatalf("got %q", got)
	}
}
