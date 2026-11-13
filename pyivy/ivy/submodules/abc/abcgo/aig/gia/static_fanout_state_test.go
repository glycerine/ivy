package gia

import "testing"

func TestStaticFanoutState(t *testing.T) {
	p := Gia_ManStart(4)
	Gia_ManAppendCi(p)
	if Gia_ManHasStaticFanout(p) {
		t.Fatal("new manager has static fanouts")
	}
	Gia_ManStaticFanoutStart(p)
	if !Gia_ManHasStaticFanout(p) {
		t.Fatal("static fanouts not reported")
	}
	Gia_ManStaticFanoutStop(p)
	if Gia_ManHasStaticFanout(p) {
		t.Fatal("stopped fanouts still reported")
	}
}
