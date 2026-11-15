package gia

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/tim"
)

func TestEdgeWindowsAndBoxesMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "edge_context_review.c")
	source, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("original window/box timing: %v\n%s", err, source)
	}
	observed := hashReviewStdout(t, func() {
		emit := func(name string, v *Vec_Int_t) {
			fmt.Print(name)
			for _, value := range Vec_IntArray(v) {
				fmt.Printf(" %d", value)
			}
			fmt.Println()
		}
		snapshot := func(p *Gia_Man_t) {
			emit("first", p.VEdge1)
			emit("second", p.VEdge2)
			emit("delay", p.vEdgeDelay)
		}
		for mode := 0; mode < 2; mode++ {
			for two := int32(0); two < 2; two++ {
				for replace := 0; replace < 2; replace++ {
					p, x, y := edgeReviewChain()
					if mode != 0 {
						p.VMapping = nil
					}
					Gia_ManStaticFanoutStart(p)
					leaves, nodes, temp := Vec_IntAlloc(1), Vec_IntAlloc(1), Vec_IntAlloc(2)
					Vec_IntPush(leaves, x)
					Vec_IntPush(nodes, y)
					Vec_IntPush(temp, -7)
					window := Vec_WecStart(1)
					if replace == 0 {
						Vec_WecPush(window, 0, x)
						Vec_WecPush(window, 0, 1)
					} else {
						Vec_WecPush(window, 0, 1)
						Vec_WecPush(window, 0, 2)
					}
					Gia_ManComputeEdgeDelay(p, two)
					fmt.Printf("window %d %d %d full %d\n", mode, two, replace, Gia_ManEvalWindow(p, leaves, nodes, window, temp, two))
					snapshot(p)
					emit("mapping", Gia_ObjLutFanins2(p, y))
					emit("window", Vec_WecEntry(window, 0))
					emit("temp", temp)
					if mode == 0 {
						Gia_ManComputeEdgeDelay(p, two)
						fmt.Printf("inc %d\n", Gia_ManEvalWindowInc(p, leaves, nodes, window, temp, two))
						snapshot(p)
						emit("mapping", Gia_ObjLutFanins2(p, y))
						emit("window", Vec_WecEntry(window, 0))
						emit("temp", temp)
					}
					Gia_ManStop(p)
				}
			}
		}
		for mode := 0; mode < 2; mode++ {
			for two := int32(0); two < 2; two++ {
				for ai, arrival := range []int{0, 6, 20} {
					for di, delay := range []int{0, 3, 12} {
						p := Gia_ManStart(8)
						a, box := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
						root := Gia_ManAppendAnd(p, a, box)
						Gia_ManAppendCo(p, a)
						Gia_ManAppendCo(p, root)
						p.VMapping = Vec_IntStart(int(p.nObjs))
						Vec_IntWriteEntry(p.VMapping, int(root>>1), int32(Vec_IntSize(p.VMapping)))
						for _, v := range []int32{2, a >> 1, box >> 1, root >> 1} {
							Vec_IntPush(p.VMapping, v)
						}
						if mode != 0 {
							Edg_ManToMapping(p)
							p.VMapping = nil
						}
						timing := tim.Tim_ManStart(2, 2)
						tim.Tim_ManSetDelayTables(timing, [][]float32{{0, 1, 1, float32(delay)}})
						tim.Tim_ManCreateBox(timing, 0, 1, 1, 1, 0, 0)
						tim.Tim_ManInitPiArrival(timing, 0, float32(arrival))
						p.PManTime = timing
						fmt.Printf("box %d %d %d %d compute %d\n", mode, two, ai, di, Gia_ManComputeEdgeDelay(p, two))
						snapshot(p)
						fmt.Printf("eval %d\n", Gia_ManEvalEdgeDelay(p))
						snapshot(p)
						fmt.Printf("times %.0f %.0f %.0f %.0f\n", tim.Tim_ManGetCiArrival(timing, 0), tim.Tim_ManGetCiArrival(timing, 1), tim.Tim_ManGetCoArrival(timing, 0), tim.Tim_ManGetCoArrival(timing, 1))
						Gia_ManStop(p)
						tim.Tim_ManStop(timing)
					}
				}
			}
		}
	})
	got, want := strings.Split(observed, "\n"), strings.Split(string(source), "\n")
	if len(got) != len(want) {
		t.Fatalf("trace lengths: Go %d, C %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("trace line %d: Go %q; C %q", i+1, got[i], want[i])
		}
	}
}

func TestEdgeWindowRestoresAfterPanic(t *testing.T) {
	for _, incremental := range []bool{false, true} {
		t.Run(fmt.Sprint(incremental), func(t *testing.T) {
			p, x, y := edgeReviewChain()
			defer Gia_ManStop(p)
			p.VMapping = nil
			Gia_ManStaticFanoutStart(p)
			Gia_ManComputeEdgeDelay(p, 1)
			leaves, nodes, temp := Vec_IntAlloc(1), Vec_IntAlloc(1), Vec_IntAlloc(2)
			Vec_IntPush(leaves, x)
			Vec_IntPush(nodes, y)
			window := Vec_WecStart(1)
			for _, leaf := range []int32{1, 2, 1, 2, 1} {
				Vec_WecPush(window, 0, leaf)
			}
			func() {
				defer func() {
					if got := fmt.Sprint(recover()); got != "gia: edge delay supports at most four LUT fanins" {
						t.Fatalf("expected invalid-window panic, got %q", got)
					}
				}()
				if incremental {
					Gia_ManEvalWindowInc(p, leaves, nodes, window, temp, 1)
				} else {
					Gia_ManEvalWindow(p, leaves, nodes, window, temp, 1)
				}
			}()
			assertIDs(t, Gia_ObjLutFanins2(p, y), x, 1)
			assertIDs(t, Vec_WecEntry(window, 0), 1, 2, 1, 2, 1)
		})
	}
}
