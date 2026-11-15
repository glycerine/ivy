package gia

import (
	"fmt"
	"testing"
)

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
