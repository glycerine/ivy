package gia

import (
	"math"
	"os/exec"
	"strings"
	"testing"
)

func TestGlitchRejectsNaNTransitionProbability(t *testing.T) {
	exe := compileGiaSourceOracle(t, "glitch_review.c")
	g := glitchReviewGraph{cis: 1, nodes: []glitchReviewNode{{[]int32{0}, 0xaaaaaaaaaaaaaaaa}}, outputs: []int32{1}}
	file := g.write(t, t.TempDir())
	for _, tc := range []struct {
		name, mode string
		set        func(*Gli_Man_t, float32)
	}{{"combinational", "0", Gli_ManSetPiRandom}, {"sequential", "1", Gli_ManSetPiRandomSeq}} {
		t.Run(tc.name, func(t *testing.T) {
			source, err := exec.Command(exe, file, tc.mode, "1", "0", "NaN", "0").CombinedOutput()
			if err == nil || !strings.Contains(string(source), "PiTransProb < 1.0") {
				t.Fatalf("original probability precondition changed: %v %s", err, source)
			}
			p := g.build()
			defer Gli_ManStop(p)
			cexReviewMustPanic(t, func() { tc.set(p, float32(math.NaN())) })
		})
	}
}
