package gia

import "testing"

func TestAbstractionFlopClassOwnership(t *testing.T) {
	p := Gia_ManStart(10)
	Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, 0)
	Gia_ManSetRegNum(p, 1)
	source := []int32{1}
	if err := p.SetAbstractionFlopClasses(source); err != nil {
		t.Fatal(err)
	}
	source[0] = 0
	got := p.AbstractionFlopClasses()
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("classes %v", got)
	}
	got[0] = 0
	if p.AbstractionFlopClasses()[0] != 1 {
		t.Fatal("getter aliases manager")
	}
	if err := p.SetAbstractionFlopClasses([]int32{1, 0}); err == nil {
		t.Fatal("accepted wrong class count")
	}
	if err := p.SetAbstractionFlopClasses(nil); err != nil || p.AbstractionFlopClasses() != nil {
		t.Fatal("clearing map failed")
	}
}
