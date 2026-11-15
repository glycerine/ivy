package gia

import (
	"regexp"
	"strings"
	"testing"
)

func TestChronologicalMiterStatsExcludeConstants(t *testing.T) {
	g := Gia_ManStart(16)
	lit := Gia_ManAppendCi(g)
	for _, root := range []int32{0, 1, lit, lit ^ 1} {
		Gia_ManAppendCo(g, root)
	}
	text := hashReviewStdout(t, func() { Cbs_ManSolveMiter(g, 1000, nil, 1) })
	for _, pattern := range []string{`Unsat calls\s+0`, `Sat\s+calls\s+2`, `Undef calls\s+0`, `\(\s*50\.00 %\)`, `Ave conf =\s+0\.0`, `Total time =`} {
		if !regexp.MustCompile(pattern).MatchString(text) {
			t.Fatalf("source statistics %q missing: %q", pattern, text)
		}
	}
	if !strings.Contains(text, "Constant 1 output of SRM!!!") {
		t.Fatal("source constant-output diagnostic missing")
	}
}
