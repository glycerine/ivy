package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRexAlphabetAndPreprocess(t *testing.T) {
	s := Gia_ManRexPreprocess(" (a|b)* a ")
	if s != "((((a|b))*a))" {
		t.Fatalf("preprocess = %q", s)
	}
	var alpha, states *Vec_Int_t
	if n := Gia_ManRexNumInputs(s, &alpha, &states); n != 3 {
		t.Fatalf("states = %d", n)
	}
	if got := vec.Vec_IntArray(alpha); len(got) != 2 || got[0] != 'a' || got[1] != 'b' {
		t.Fatalf("alphabet = %v", got)
	}
	if got := vec.Vec_IntEntry(states, len(s)); got != 3 {
		t.Fatalf("terminal state = %d", got)
	}
}
func TestRexOneHotTruth(t *testing.T) {
	p := Gia_ManStart(16)
	lits := vec.Vec_IntAlloc(3)
	for i := 0; i < 3; i++ {
		vec.Vec_IntPush(lits, Gia_ManAppendCi(p))
	}
	Gia_ManHashStart(p)
	result := Gia_ManBuild1Hot(p, lits)
	Gia_ManHashStop(p)
	Gia_ManAppendCo(p, result)
	got := Gia_LutComputeTruth66_rec(p, Gia_ManCo(p, 0))
	if byte(got) != 0x16 {
		t.Fatalf("one-hot truth = %02x", byte(got))
	}
}
