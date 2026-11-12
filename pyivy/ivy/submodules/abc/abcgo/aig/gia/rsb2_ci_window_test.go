package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestRsb2CIWindow(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	window := Gia_RsbCiWindow(p, 2)
	data := vec.Vec_IntArray(window)
	if len(data) != 10 || data[6] != 2 || data[7] != 4 || data[8] != 6 || data[9] != 6 {
		t.Fatalf("window %v", data)
	}
	q := Gia_ManFromResub(data, int32(len(data)/2), 2)
	if Gia_ManVerifyTwoTruths(p, q) != 1 {
		t.Fatal("window changed function")
	}
	if p.vFanout != nil {
		t.Fatal("fanout scratch not released")
	}
}
