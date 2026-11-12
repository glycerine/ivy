package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestGiaQbfQuantificationAndCofactor(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b))
	for _, test := range []struct {
		and, or int32
		want    bool
	}{{1, 0, false}, {0, 1, true}} {
		q := Gia_QbfQuantifyOne(p, 1, test.and, test.or)
		for _, v := range []bool{false, true} {
			got, err := q.Evaluate([]bool{v, false})
			if err != nil || got[0] != (v && test.want) {
				t.Fatalf("quantification: %v %v", got, err)
			}
		}
	}
	both := Gia_QbfQuantifyOne(p, 1, 0, 0)
	got, err := both.Evaluate([]bool{true, false})
	if err != nil || len(got) != 2 || got[0] || !got[1] {
		t.Fatalf("cofactors %v %v", got, err)
	}
	values := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(values, 1)
	mapping := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(mapping, -1)
	q := Gia_QbfCofactor(p, 1, values, mapping)
	got, err = q.Evaluate([]bool{true})
	if err != nil || !got[0] {
		t.Fatalf("cofactor %v %v", got, err)
	}
	all := Gia_QbfQuantifyAll(p, 1, 0, 1)
	got, err = all.Evaluate([]bool{true, false})
	if err != nil || !got[0] {
		t.Fatalf("quantify all %v %v", got, err)
	}
}
