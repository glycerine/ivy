package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"
)

func TestExplicitReachabilitySTGZeroWidthColumns(t *testing.T) {
	for _, inputs := range []int32{0, 1} {
		for _, outputs := range []int32{0, 1} {
			for _, toggle := range []bool{false, true} {
				t.Run(fmt.Sprintf("inputs%d/outputs%d/toggle%t", inputs, outputs, toggle), func(t *testing.T) {
					t.Chdir(t.TempDir())
					g := Gia_ManStart(8)
					defer Gia_ManStop(g)
					if inputs != 0 {
						Gia_ManAppendCi(g)
					}
					ro := Gia_ManAppendCi(g)
					if outputs != 0 {
						Gia_ManAppendCo(g, ro)
					}
					ri := ro
					if toggle {
						ri ^= 1
					}
					Gia_ManAppendCo(g, ri)
					Gia_ManSetRegNum(g, 1)
					hashReviewStdout(t, func() {
						if got := Gia_ManCollectReachable(g, 100, 0, 1, 0); got != 1 {
							t.Fatalf("result%d", got)
						}
					})
					data, err := testoracle.ReadFile("test.stg")
					if err != nil {
						t.Fatal(err)
					}
					// Extra_PrintBinary in C prints zero characters for zero bits.
					want := ""
					nstates := 1
					if toggle {
						nstates = 2
					}
					for state := 1; state <= nstates; state++ {
						next := state
						if toggle {
							next = 3 - state
						}
						for condition := 0; condition < 1<<inputs; condition++ {
							if inputs != 0 {
								want += fmt.Sprint(condition)
							}
							want += fmt.Sprintf(" %d %d ", state, next)
							if outputs != 0 {
								want += fmt.Sprint(state - 1)
							}
							want += "\n"
						}
					}
					if string(data) != want {
						t.Fatalf("STG text\ngot  %q\nwant %q", data, want)
					}
				})
			}
		}
	}
}

func TestExplicitReachabilityNativeSTG(t *testing.T) {
	exe := compileGiaSourceOracle(t, "era_review.c")
	cmd := testoracle.Command(t, exe, "--dump")
	cmd.Dir = t.TempDir()
	native, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("original C STG: %v\n%s", err, native)
	}
	t.Chdir(t.TempDir())
	goText := hashReviewStdout(t, func() {
		for ni := int32(0); ni < 2; ni++ {
			for no := int32(0); no < 2; no++ {
				for toggle := int32(0); toggle < 2; toggle++ {
					g := Gia_ManStart(8)
					if ni != 0 {
						Gia_ManAppendCi(g)
					}
					ro := Gia_ManAppendCi(g)
					if no != 0 {
						Gia_ManAppendCo(g, ro)
					}
					Gia_ManAppendCo(g, ro^toggle)
					Gia_ManSetRegNum(g, 1)
					fmt.Printf("dump %d %d %d\n", ni, no, toggle)
					Gia_ManCollectReachable(g, 100, 0, 1, 0)
					fmt.Println("stg begin")
					data, err := testoracle.ReadFile("test.stg")
					if err != nil {
						t.Fatal(err)
					}
					fmt.Print(string(data))
					fmt.Println("stg end")
					Gia_ManStop(g)
				}
			}
		}
	})
	rows := func(text string) string {
		var result strings.Builder
		in := false
		for _, line := range strings.Split(text, "\n") {
			if line == "stg begin" {
				in = true
			}
			if in || strings.HasPrefix(line, "dump ") {
				result.WriteString(line)
				result.WriteByte('\n')
			}
			if line == "stg end" {
				in = false
			}
		}
		return result.String()
	}
	if got, want := rows(goText), rows(string(native)); got != want {
		t.Fatalf("Go STG\n%s\nC STG\n%s", got, want)
	}
}
