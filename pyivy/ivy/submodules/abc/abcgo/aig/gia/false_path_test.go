package gia

import (
	"github.com/glycerine/abcgo/sat/satoko"
	"testing"
)

func TestFalsePathDetectionAndRebuild(t *testing.T) {
	p := Gia_ManStart(12)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y := Gia_ManAppendAnd(p, x, a^1)
	Gia_ManAppendCo(p, y)
	Gia_ManLevelNum(p)
	path, status := giaFalsePathSolve(p, y>>1, 0, false)
	if status != satoko.SATOKO_UNSAT || path == nil {
		t.Fatalf("expected false path, status=%d", status)
	}
	q := Gia_ManFalseRebuildPath(p, path, 0, 0)
	for bits := 0; bits < 4; bits++ {
		in := []bool{bits&1 != 0, bits&2 != 0}
		out, err := q.Evaluate(in)
		if err != nil || out[0] {
			t.Fatalf("output=%v error=%v", out, err)
		}
	}
}

func TestFalsePathDrivers(t *testing.T) {
	for _, batch := range []bool{false, true} {
		p := Gia_ManStart(12)
		a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a, b)
		y := Gia_ManAppendAnd(p, x, a^1)
		Gia_ManAppendCo(p, y)
		var q *Gia_Man_t
		if batch {
			q = Gia_ManCheckFalse2(p, 0, 0, 0, 0)
		} else {
			q = Gia_ManCheckFalse(p, 0, 0, 0, 0)
		}
		for bits := 0; bits < 4; bits++ {
			in := []bool{bits&1 != 0, bits&2 != 0}
			got, err := q.Evaluate(in)
			if err != nil || got[0] {
				t.Fatalf("batch=%v output=%v error=%v", batch, got, err)
			}
		}
	}
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	root := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, root)
	Gia_ManLevelNum(p)
	if path, status := giaFalsePathSolve(p, root>>1, 0, false); status != satoko.SATOKO_SAT || path != nil {
		t.Fatalf("real path rejected: status=%d path=%v", status, path)
	}
}
