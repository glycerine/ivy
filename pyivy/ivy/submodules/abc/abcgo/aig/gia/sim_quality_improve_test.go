package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestPatternQualitySelectsImprovement(t *testing.T) {
	p := Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	Gia_ManAppendCo(p, x)
	pat := vec.Vec_IntStart(2)
	stats := vec.Vec_IntStart(int(p.nObjs))
	vec.Vec_IntWriteEntry(stats, int(x>>1), 100)
	selected := Gia_SimQualityPatternsMax(p, pat, 0, 0, stats)
	if selected < 0 || selected > 1 || vec.Vec_IntEntry(pat, 0) != 0 || vec.Vec_IntEntry(pat, 1) != 0 {
		t.Fatalf("wrong selected input %d or modified pattern", selected)
	}
	vec.Vec_IntWriteEntry(pat, int(selected), 1)
	if Gia_SimComputeScore(p, stats, Gia_SimQualityOne(p, pat, 0)) <= 1 {
		t.Fatal("selected flip did not improve score")
	}
}
