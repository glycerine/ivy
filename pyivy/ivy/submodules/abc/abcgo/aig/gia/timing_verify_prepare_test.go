package gia

import (
	"path/filepath"
	"testing"
)

func TestPrepareBoxVerificationPlainGraphs(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	Gia_ManAppendCo(p, Gia_ManAppendAnd(p, a, b))
	path := filepath.Join(t.TempDir(), "spec.aig")
	if err := Gia_AigerWrite(p, path, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	left, right, err := giaPrepareBoxVerification(p, path, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer Gia_ManStop(left)
	defer Gia_ManStop(right)
	for mask := 0; mask < 4; mask++ {
		in := []bool{mask&1 != 0, mask&2 != 0}
		a, ae := left.Evaluate(in)
		b, be := right.Evaluate(in)
		if ae != nil || be != nil || a[0] != b[0] {
			t.Fatalf("input %v: %v %v", in, a, b)
		}
	}
	if _, _, err := giaPrepareBoxVerification(p, "", 0); err == nil {
		t.Fatal("missing spec accepted")
	}
}
