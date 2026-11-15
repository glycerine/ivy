package gia_test

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	// Register the actual IF-owned SOP and DSD algorithms, not test substitutes.
	_ "github.com/glycerine/abcgo/map/if"
)

func TestWindowDriversWithOriginalBalancingEngines(t *testing.T) {
	cases := 0
	for _, engine := range []string{"SOP", "DSD"} {
		for signs := 0; signs < 8; signs++ {
			for _, choice := range [][2]int32{{1, 0}, {2, 0}, {0, 50}} {
				t.Run(fmt.Sprintf("%s/phase%d/target%d-window%d", engine, signs, choice[0], choice[1]), func(t *testing.T) {
					// IF prints progress even with verbose=0; retain it in a file.
					progress, err := os.CreateTemp(t.TempDir(), "engine-progress")
					if err != nil {
						t.Fatal(err)
					}
					original := os.Stdout
					os.Stdout = progress
					defer func() { os.Stdout = original; progress.Close() }()
					src := gia.Gia_ManStart(32)
					a, b, c, r := gia.Gia_ManAppendCi(src), gia.Gia_ManAppendCi(src), gia.Gia_ManAppendCi(src), gia.Gia_ManAppendCi(src)
					gia.Gia_ManHashStart(src)
					x := gia.Gia_ManHashXor(src, a^int32(signs&1), b^int32(signs>>1&1))
					y := gia.Gia_ManHashAnd(src, x, c^int32(signs>>2&1))
					m := gia.Gia_ManHashMux(src, r, y, a)
					for _, lit := range []int32{y ^ 1, m, x, r ^ 1} {
						gia.Gia_ManAppendCo(src, lit)
					}
					gia.Gia_ManHashStop(src)
					gia.Gia_ManSetRegNum(src, 1)
					if err := src.SetAigerProperties(1, 1); err != nil {
						t.Fatal(err)
					}
					if err := src.SetAigerLatchInitialStates([]int32{1}); err != nil {
						t.Fatal(err)
					}
					src.SetAigerComment([]byte("engine-window"))
					var out *gia.Gia_Man_t
					if engine == "SOP" {
						out = gia.Gia_ManPerformSopBalanceWin(src, choice[0], choice[1], 8, 0, 0)
					} else {
						out = gia.Gia_ManPerformDsdBalanceWin(src, choice[0], choice[1], 6, 8, 0, 0)
					}
					if out.RegisterCount() != 1 || gia.Gia_ManConstrNum(out) != 1 || !slices.Equal(out.AigerLatchInitialStates(), []int32{1}) || string(out.AigerComment()) != "engine-window" {
						t.Fatal("engine window lost interface metadata")
					}
					plain := gia.Gia_ManDupNoMuxes(out, 1)
					data, err := gia.MarshalAiger(plain, gia.AigerWriteOptions{OmitExtensions: true, OmitSymbols: true})
					if err != nil {
						t.Fatal(err)
					}
					header := strings.Fields(strings.SplitN(string(data), "\n", 2)[0])
					if len(header) < 8 || header[6] != "1" || header[7] != "1" {
						t.Fatalf("lost property partition in header %v", header)
					}
					restored, err := gia.ParseAiger(data, gia.AigerReadOptions{})
					if err != nil {
						t.Fatal(err)
					}
					for mask := 0; mask < 16; mask++ {
						inputs := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
						want, err := src.Evaluate(inputs)
						if err != nil {
							t.Fatal(err)
						}
						for _, g := range []*gia.Gia_Man_t{out, restored} {
							got, err := g.Evaluate(inputs)
							if err != nil {
								t.Fatal(err)
							}
							if !slices.Equal(got, want) {
								t.Fatalf("assignment%d: got%v want%v", mask, got, want)
							}
						}
					}
				})
				cases++
			}
		}
	}
	t.Logf("%d real SOP/DSD window runs preserve all16 assignments and serialized properties", cases)
}
