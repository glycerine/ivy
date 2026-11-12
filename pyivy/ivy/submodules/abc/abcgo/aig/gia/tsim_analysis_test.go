package gia

import "testing"

func TestTernaryRegisterReplacementMap(t *testing.T) {
	p := Gia_ManStart(16)
	pi := Gia_ManAppendCi(p)
	r0, r1, r2, r3 := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	for _, r := range []int32{r0, r1, r2, r3} {
		Gia_ManAppendCo(p, r)
	}
	Gia_ManAppendCo(p, 0)
	Gia_ManAppendCo(p, r1^1)
	Gia_ManAppendCo(p, r2^1)
	Gia_ManAppendCo(p, pi)
	Gia_ManSetRegNum(p, 4)
	s := Gia_ManTerSimulate(p, 0)
	defer Gia_ManTerDelete(s)
	m := Gia_ManTerCreateMap(s, 0)
	if m[1] != 0 || m[2] != -1 || m[3] != r1 || m[4] != -1 {
		t.Fatalf("mapping %v", m)
	}
	q := Gia_ManDupDfsCiMap(p, m, nil)
	if Gia_ManRegNum(q) != 4 {
		t.Fatal("register interface changed")
	}
	reduced := Gia_ManReduceConst(p, 0)
	out, err := reduced.Evaluate([]bool{false, true, false, true, false})
	if err != nil || out[0] || out[1] != out[2] {
		t.Fatalf("reduced outputs %v: %v", out, err)
	}
}
