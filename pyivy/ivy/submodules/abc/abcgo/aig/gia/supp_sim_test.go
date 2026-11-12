package gia

import "testing"

func TestMin2Simulation(t *testing.T) {
	p := Gia_ManStart(16)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	m := Gia_Man2SuppStart(p)
	defer Gia_Man2SuppStop(m)
	m.iLits = [2]int32{x, a ^ 1}
	Gia_Min2ManGatherSupp(m)
	Gia_Min2ManSimulate(m) // Allocate and seed input words.
	Vec_WrdWriteEntry(m.vSims, int(a>>1), 0xaaaaaaaaaaaaaaaa)
	Vec_WrdWriteEntry(m.vSims, int(b>>1), ^uint64(0))
	if Gia_Min2ManSimulate(m) != 1 {
		t.Fatal("a & b and !a cannot coincide")
	}
	m.iLits[1] = a
	Gia_Min2ManGatherSupp(m)
	if Gia_Min2ManSimulate(m) != 0 {
		t.Fatal("expected satisfying patterns")
	}
	Gia_Min2SimSetInputBit(m, a>>1, 1, 1)
	Gia_Min2SimSetInputBit(m, a>>1, 0, 3)
	word := Vec_WrdEntry(m.vSims, int(a>>1))
	if word&(1<<1) == 0 || word&(1<<3) != 0 {
		t.Fatal("input bit update failed")
	}
}
