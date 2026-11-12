package gia

import (
	"testing"
	"time"
)

func TestSequentialMiterVerification(t *testing.T) {
	graph := func(fail bool) *Gia_Man_t {
		p := Gia_ManStart(8)
		in, reg := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		out := int32(0)
		if fail {
			out = reg
		}
		Gia_ManAppendCo(p, out)
		Gia_ManAppendCo(p, in)
		Gia_ManSetRegNum(p, 1)
		return p
	}
	if got := giaVerifySequentialMiter(graph(false), 4, time.Time{}); got != 1 {
		t.Fatalf("safe %d", got)
	}
	if got := giaVerifySequentialMiter(graph(true), 4, time.Time{}); got != 0 {
		t.Fatalf("unsafe %d", got)
	}
	if got := giaVerifySequentialMiter(graph(false), 1, time.Time{}); got != -1 {
		t.Fatalf("state limit %d", got)
	}
	if got := giaVerifySequentialMiter(graph(false), 4, time.Now().Add(-time.Second)); got != -1 {
		t.Fatalf("deadline %d", got)
	}
}
