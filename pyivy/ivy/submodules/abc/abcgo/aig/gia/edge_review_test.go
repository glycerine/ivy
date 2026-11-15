package gia

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func edgeReviewChain() (p *Gia_Man_t, x, y int32) {
	p = Gia_ManStart(8)
	a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x = Gia_ManAppendAnd(p, a, b) >> 1
	y = Gia_ManAppendAnd(p, x<<1, a) >> 1
	Gia_ManAppendCo(p, y<<1)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, record := range []struct {
		id     int32
		fanins []int32
	}{{x, []int32{a >> 1, b >> 1}}, {y, []int32{x, a >> 1}}} {
		Vec_IntWriteEntry(p.VMapping, int(record.id), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, 2)
		for _, fanin := range record.fanins {
			Vec_IntPush(p.VMapping, fanin)
		}
		Vec_IntPush(p.VMapping, record.id)
	}
	Edg_ManToMapping(p)
	return p, x, y
}

func TestEdgeImproveReportsDirectionAndCount(t *testing.T) {
	for _, fanins := range []bool{false, true} {
		t.Run(fmt.Sprint(fanins), func(t *testing.T) {
			p, x, y := edgeReviewChain()
			defer Gia_ManStop(p)
			p.VEdge1, p.VEdge2 = Vec_IntStart(int(p.nObjs)), Vec_IntStart(int(p.nObjs))
			Edg_ManEvalEdgeDelay(p)
			Edg_ManEvalEdgeDelayR(p)
			node, direction := x, "fanouts"
			if fanins {
				node, direction = y, "fanins"
			}
			text := hashReviewStdout(t, func() {
				if Edg_ObjImprove(p, node, 2, 2, 1) != 1 {
					t.Error("expected critical edge improvement")
				}
			})
			if !strings.HasSuffix(text, "Fixed 1 critical "+direction+"\n") {
				t.Fatalf("missing count/direction: %q", text)
			}
		})
	}
}

func TestEdgeWindowSecondaryMappingAndRestoration(t *testing.T) {
	for _, incremental := range []bool{false, true} {
		t.Run(fmt.Sprintf("incremental%t", incremental), func(t *testing.T) {
			p, x, y := edgeReviewChain()
			defer Gia_ManStop(p)
			p.VMapping = nil // full window evaluator and SAT LUT caller use this form
			Gia_ManStaticFanoutStart(p)
			Gia_ManComputeEdgeDelay(p, 1)
			leaves, nodes, temp := Vec_IntAlloc(1), Vec_IntAlloc(1), Vec_IntAlloc(2)
			Vec_IntPush(leaves, x)
			Vec_IntPush(nodes, y)
			window := Vec_WecStart(1)
			Vec_WecPush(window, 0, 1)
			Vec_WecPush(window, 0, 2)
			original := append([]int32(nil), Vec_IntArray(Gia_ObjLutFanins2(p, y))...)
			var got int32
			if incremental {
				got = Gia_ManEvalWindowInc(p, leaves, nodes, window, temp, 1)
			} else {
				got = Gia_ManEvalWindow(p, leaves, nodes, window, temp, 1)
			}
			if got != 10 {
				t.Fatalf("two-input replacement should remove the second LUT level: %d", got)
			}
			if !reflect.DeepEqual(Vec_IntArray(Gia_ObjLutFanins2(p, y)), original) || !reflect.DeepEqual(Vec_IntArray(Vec_WecEntry(window, 0)), []int32{1, 2}) {
				t.Fatal("window evaluation did not restore both mapping records")
			}
		})
	}
}

func TestEdgeUnitEvaluationClearsPreviousTiming(t *testing.T) {
	p, x, y := edgeReviewChain()
	defer Gia_ManStop(p)
	p.VEdge1, p.VEdge2 = Vec_IntStart(int(p.nObjs)), Vec_IntStart(int(p.nObjs))
	p.vEdgeDelay, p.vEdgeDelayR = Vec_IntStart(1), Vec_IntStart(1)
	for _, fast := range []bool{false, true, false, true} {
		Vec_IntFill(p.vEdgeDelay, int(p.nObjs), 99)
		Vec_IntFill(p.vEdgeDelayR, int(p.nObjs), 99)
		Vec_IntFill(p.VEdge1, int(p.nObjs), 0)
		Vec_IntFill(p.VEdge2, int(p.nObjs), 0)
		forward, reverse := int32(2), int32(1)
		if fast {
			Gia_ObjEdgeAdd(x, y, p.VEdge1, p.VEdge2)
			Gia_ObjEdgeAdd(y, x, p.VEdge1, p.VEdge2)
			forward, reverse = 1, 0
		}
		if got := Edg_ManEvalEdgeDelay(p); got != forward {
			t.Fatalf("fast%t: forward %d, want %d", fast, got, forward)
		}
		if got := Edg_ManEvalEdgeDelayR(p); got != reverse {
			t.Fatalf("fast%t: reverse %d, want %d", fast, got, reverse)
		}
		for _, id := range []int{0, 1, 2, int(p.nObjs - 1)} {
			if Vec_IntEntry(p.vEdgeDelay, id) != 0 || Vec_IntEntry(p.vEdgeDelayR, id) != 0 {
				t.Fatalf("unmapped object %d kept stale timing", id)
			}
		}
	}
}

func edgeReviewWideLut(inputs int, secondary bool) *Gia_Man_t {
	p := Gia_ManStart(int32(2*inputs + 2))
	var leaves []int32
	for i := 0; i < inputs; i++ {
		leaves = append(leaves, Gia_ManAppendCi(p))
	}
	root := leaves[0]
	for _, lit := range leaves[1:] {
		root = Gia_ManAppendAnd(p, root, lit)
	}
	Gia_ManAppendCo(p, root)
	if secondary {
		p.vMapping2 = Vec_WecStart(int(p.nObjs))
		for _, lit := range leaves {
			Vec_WecPush(p.vMapping2, int(root>>1), lit>>1)
		}
	} else {
		p.VMapping = Vec_IntStart(int(p.nObjs))
		Vec_IntWriteEntry(p.VMapping, int(root>>1), int32(Vec_IntSize(p.VMapping)))
		Vec_IntPush(p.VMapping, int32(inputs))
		for _, lit := range leaves {
			Vec_IntPush(p.VMapping, lit>>1)
		}
		Vec_IntPush(p.VMapping, root>>1)
	}
	return p
}

func TestEdgeReverseAssignmentSupportsWideLuts(t *testing.T) {
	// giaEdge.c limits forward assignment to four inputs; its reverse routine
	// propagates over an arbitrary fanin list and has no such restriction.
	for _, inputs := range []int{5, 8, 16} {
		for _, secondary := range []bool{false, true} {
			t.Run(fmt.Sprintf("inputs%d/secondary%t", inputs, secondary), func(t *testing.T) {
				p := edgeReviewWideLut(inputs, secondary)
				defer Gia_ManStop(p)
				if delay := Gia_ManComputeEdgeDelay2(p); delay != 1 {
					t.Fatalf("reverse delay %d, want 1", delay)
				}
				for i := int32(0); i < Gia_ManCiNum(p); i++ {
					if got := Vec_IntEntry(p.vEdgeDelay, int(Gia_ManCi(p, i).id)); got != 1 {
						t.Fatalf("CI %d reverse delay %d, want 1", i, got)
					}
				}
				if Gia_ManEvalEdgeCount(p) != 0 {
					t.Fatal("single mapped LUT must not acquire internal fast edges")
				}
			})
		}
	}
}

func TestEdgeForwardRetainsFourInputLimit(t *testing.T) {
	for _, secondary := range []bool{false, true} {
		for _, op := range []string{"evaluate", "assign", "convert"} {
			if secondary && op == "convert" {
				continue // conversion takes the primary mapping as its input
			}
			t.Run(fmt.Sprintf("%s/secondary%t", op, secondary), func(t *testing.T) {
				p := edgeReviewWideLut(5, secondary)
				defer Gia_ManStop(p)
				p.VEdge1, p.VEdge2 = Vec_IntStart(int(p.nObjs)), Vec_IntStart(int(p.nObjs))
				defer func() {
					if got := fmt.Sprint(recover()); got != "gia: edge delay supports at most four LUT fanins" {
						t.Fatalf("expected four-input precondition, got %q", got)
					}
				}()
				switch op {
				case "evaluate":
					Gia_ManEvalEdgeDelay(p)
				case "assign":
					Gia_ManComputeEdgeDelay(p, 1)
				case "convert":
					Edg_ManToMapping(p)
				}
			})
		}
	}
}
