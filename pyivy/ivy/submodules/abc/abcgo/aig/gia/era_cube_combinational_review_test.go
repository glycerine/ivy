package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func TestCubeReachabilityNativeCombinationalDomain(t *testing.T) {
	exe := compileGiaSourceOracle(t, "era_cube_review.c")
	native, err := testoracle.Command(t, exe, "--combinational").CombinedOutput()
	if err != nil {
		t.Fatalf("C zero-register domain: %v\n%s", err, native)
	}
	if strings.Contains(string(native), "INVALID") {
		t.Fatal("C generated an invalid witness")
	}
	goText := hashReviewStdout(t, func() {
		for inputs := int32(1); inputs <= 2; inputs++ {
			for kind := int32(0); kind < 3; kind++ {
				for miter := int32(0); miter < 2; miter++ {
					g := Gia_ManStart(16)
					lit := int32(0)
					for i := int32(0); i < inputs; i++ {
						lit = Gia_ManAppendCi(g)
					}
					if kind < 2 {
						lit = kind
					}
					Gia_ManAppendCo(g, lit)
					fmt.Printf("combinational %d %d %d\n", inputs, kind, miter)
					result := Gia_ManArePerform(g, 100, miter, 0)
					fmt.Printf("result %d\n", result)
					if g.pCexSeq != nil {
						if Gia_ManVerifyCex(g, g.pCexSeq, 0) == 0 {
							t.Fatal("Go generated an invalid zero-register witness")
						}
						fmt.Printf("witness %d %d\n", g.pCexSeq.IPo, g.pCexSeq.IFrame)
					}
					Gia_ManStop(g)
				}
			}
		}
	})
	records := func(text string) string {
		var out strings.Builder
		for _, line := range strings.Split(text, "\n") {
			if strings.HasPrefix(line, "combinational ") || strings.HasPrefix(line, "result ") || strings.HasPrefix(line, "witness ") {
				out.WriteString(line)
				out.WriteByte('\n')
			}
		}
		return out.String()
	}
	if got, want := records(goText), records(string(native)); got != want {
		t.Fatalf("Go domain records\n%s\nC domain records\n%s", got, want)
	}
}

func TestCubeReachabilityConstantNoInputs(t *testing.T) {
	// C's MostUsedPi helper dereferences a missing CI in this degenerate case;
	// the Go nil guard already handles it. Validate by constant truth directly.
	for output := int32(0); output < 2; output++ {
		for miter := int32(0); miter < 2; miter++ {
			g := Gia_ManStart(2)
			Gia_ManAppendCo(g, output)
			want := int32(1)
			if output != 0 && miter != 0 {
				want = 0
			}
			var got int32
			hashReviewStdout(t, func() { got = Gia_ManArePerform(g, 100, miter, 0) })
			if got != want {
				t.Fatalf("output%d miter%d result%d want%d", output, miter, got, want)
			}
			if want == 0 && (g.pCexSeq == nil || g.pCexSeq.IFrame != 0 || Gia_ManVerifyCex(g, g.pCexSeq, 0) == 0) {
				t.Fatal("invalid zero-bit witness")
			}
			Gia_ManStop(g)
		}
	}
}
