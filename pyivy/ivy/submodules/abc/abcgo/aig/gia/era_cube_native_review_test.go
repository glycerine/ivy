package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func TestCubeReachabilityNativeDriverTrace(t *testing.T) {
	exe := compileGiaSourceOracle(t, "era_cube_review.c")
	cmd := testoracle.Command(t, exe)
	cmd.Dir = t.TempDir()
	native, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("C cube driver: %v\n%s", err, native)
	}
	if strings.Contains(string(native), "INVALID") || strings.Contains(string(native), "Failure...") {
		t.Fatalf("native witness failed: %s", native)
	}
	goText := hashReviewStdout(t, func() {
		for _, ni := range []int32{0, 1, 3, 5, 6} {
			for _, nr := range []int32{1, 3, 33, 65} {
				for seed := int32(0); seed < 4; seed++ {
					for miter := int32(0); miter < 2; miter++ {
						g := eraReviewGraph(ni, nr, seed)
						fmt.Printf("case %d %d %d 100 %d 0\n", ni, nr, seed, miter)
						result := Gia_ManArePerform(g, 100, miter, 0)
						fmt.Printf("result %d\n", result)
						fmt.Print("cex")
						if c := g.pCexSeq; c != nil {
							if Gia_ManVerifyCex(g, c, 0) == 0 {
								t.Fatal("invalid Go counterexample")
							}
							fmt.Printf(" %d %d ", c.IPo, c.IFrame)
							for bit := int32(0); bit < c.NBits; bit++ {
								fmt.Print((c.PData[bit>>5] >> uint(bit&31)) & 1)
							}
						}
						fmt.Println()
						Gia_ManStop(g)
					}
				}
			}
		}
	})
	records := func(text string) []string {
		var rows []string
		for _, line := range strings.Split(text, "\n") {
			if strings.HasPrefix(line, "case ") || strings.HasPrefix(line, "result ") || strings.HasPrefix(line, "cex") {
				rows = append(rows, line)
			}
		}
		return rows
	}
	want, got := records(string(native)), records(goText)
	if len(want) != 480 || len(got) != len(want) {
		t.Fatalf("native/Go records%d/%d, want480", len(want), len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("record%d\nGo %s\n C %s", i, got[i], want[i])
		}
	}
}

func TestCubeReachabilityNativeRelations(t *testing.T) {
	exe := compileGiaSourceOracle(t, "era_cube_review.c")
	native, err := testoracle.Command(t, exe, "--relations").CombinedOutput()
	if err != nil {
		t.Fatalf("C cube relations: %v\n%s", err, native)
	}
	var out strings.Builder
	for mode, positions := range [][3]int32{{0, 1, 2}, {0, 16, 32}, {15, 16, 64}} {
		for i := 0; i < 27; i++ {
			for j := 0; j < 27; j++ {
				a, b := &Gia_StaAre_t{pData: make([]uint32, 5)}, &Gia_StaAre_t{pData: make([]uint32, 5)}
				x, y := i, j
				for _, pos := range positions {
					if x%3 == 1 {
						Gia_StaSetValue0(a, pos)
					}
					if x%3 == 2 {
						Gia_StaSetValue1(a, pos)
					}
					if y%3 == 1 {
						Gia_StaSetValue0(b, pos)
					}
					if y%3 == 2 {
						Gia_StaSetValue1(b, pos)
					}
					x /= 3
					y /= 3
				}
				fmt.Fprintf(&out, "%d %d %d: %d %d %d %d %d %d\n", mode, i, j, Gia_StaAreEqual(a, b, 5), Gia_StaAreDisjoint(a, b, 5), Gia_StaAreContain(a, b, 5), Gia_StaAreDashNum(a, b, 5), Gia_StaAreSharpVar(a, b, 5), Gia_StaAreDisjointVar(a, b, 5))
			}
		}
	}
	if out.String() != string(native) {
		t.Fatal("cube relations disagree with actual C across 2187 pairs")
	}
}
