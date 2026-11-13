package bmc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestDeriveOneForAnd(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, a, b))
	for pattern := 0; pattern < 4; pattern++ {
		v := vec.Vec_IntAlloc(2)
		vec.Vec_IntPush(v, int32(pattern&1))
		vec.Vec_IntPush(v, int32(pattern>>1))
		q := Gia_ManDeriveOne(g, v)
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			out, err := q.Evaluate(in)
			if err != nil {
				t.Fatal(err)
			}
			a0, b0 := pattern&1 != 0, pattern&2 != 0
			want := false
			switch {
			case a0 && b0:
				want = in[0] && in[1]
			case !a0 && !b0:
				want = !in[0] || !in[1]
			case !a0:
				want = !in[0]
			default:
				want = !in[1]
			}
			if out[0] != want {
				t.Fatalf("pattern=%d input=%v result=%v want=%v", pattern, in, out, want)
			}
		}
	}
}

func TestEnumerateDerivedPatterns(t *testing.T) {
	g := gia.Gia_ManStart(16)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, a, b))
	if count := enumerateDerivedPatterns(g, 10); count < 1 || count > 3 {
		t.Fatalf("unexpected count %d", count)
	}
}
