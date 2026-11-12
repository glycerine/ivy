package gia

import "testing"

func TestBuildFromMini(t *testing.T) {
	for _, hash := range []int32{0, 1} {
		p := Gia_ManStart(12)
		leaves := Vec_IntAlloc(3)
		for i := 0; i < 3; i++ {
			Vec_IntPush(leaves, Gia_ManAppendCi(p))
		}
		if hash != 0 {
			Gia_ManHashAlloc(p)
		}
		mini := Vec_IntAlloc(5)
		for _, x := range []int32{0, 3, 6, 4, 1} {
			Vec_IntPush(mini, x)
		}
		root := Gia_ManBuildFromMiniInt(p, leaves, mini, hash)
		Gia_ManAppendCo(p, root)
		if Vec_IntSize(mini) != 0 {
			t.Fatal("mini scratch not cleared")
		}
		for bits := 0; bits < 8; bits++ {
			got, err := p.Evaluate([]bool{bits&1 != 0, bits&2 != 0, bits&4 != 0})
			want := !(bits&1 != 0 && bits&2 == 0 && bits&4 != 0)
			if err != nil || got[0] != want {
				t.Fatalf("hash=%d bits=%d got=%v want=%v err=%v", hash, bits, got, want, err)
			}
		}
	}
}
func TestBuildFromMiniConstantAndVariable(t *testing.T) {
	p := Gia_ManStart(4)
	leaf := Gia_ManAppendCi(p)
	leaves := Vec_IntAlloc(1)
	Vec_IntPush(leaves, leaf)
	mini := Vec_IntAlloc(2)
	Vec_IntPush(mini, 0)
	Vec_IntPush(mini, 1)
	if got := Gia_ManBuildFromMiniInt(p, leaves, mini, 0); got != leaf^1 {
		t.Fatalf("variable=%d", got)
	}
	Vec_IntClear(mini)
	Vec_IntPush(mini, 1)
	if Gia_ManBuildFromMiniInt(p, leaves, mini, 0) != 1 {
		t.Fatal("constant")
	}
}
