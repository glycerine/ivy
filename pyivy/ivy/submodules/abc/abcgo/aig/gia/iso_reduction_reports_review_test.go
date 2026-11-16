package gia

import (
	"fmt"
	"strings"
	"testing"
)

func TestIsoReductionVerboseReportsSourceClassesAndStats(t *testing.T) {
	for mode := 0; mode < 3; mode++ {
		t.Run(fmt.Sprintf("mode%d", mode), func(t *testing.T) {
			graph := iso2ClassesReviewGraph(0)
			got := isoReviewCapture(t, func() {
				if mode == 0 {
					Gia_ManIsoReduce(graph, nil, nil, 0, 0, 1, 0)
				} else {
					Gia_ManIsoReduce2(graph, nil, nil, 0, int32(mode-1), 0, 1, 0)
				}
			})
			want := "Nontrivial classes:\n    0 : { 0 1 10 }\n    1 : { 2 3 11 }\n    2 : { 4 5 }\n    5 : { 8 9 }\n"
			if !strings.HasSuffix(got, want) {
				t.Errorf("missing indexed source class report: %q", got)
			}
			if !strings.Contains(got, "candidate   classes") {
				t.Error("candidate report does not preserve source format")
			}
			if mode == 0 {
				for _, row := range []struct{ index, size, first int }{{0, 3, 0}, {1, 3, 2}, {2, 2, 4}, {5, 2, 8}} {
					if !strings.Contains(got, fmt.Sprintf("%6d %6d %6d : ", row.index, row.size, row.first)) {
						t.Errorf("missing candidate statistics for class%d", row.index)
					}
				}
				if strings.Count(got, "i/o =") != 4 {
					t.Error("missing source per-cone graph statistics")
				}
			}
		})
	}
}

func TestIsoReductionReportsSourceHundredOutputProgress(t *testing.T) {
	for _, singletons := range []bool{false, true} {
		t.Run(fmt.Sprintf("singletons%t", singletons), func(t *testing.T) {
			graph := Gia_ManStart(260)
			a, b := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
			driver := a
			var roots []int32
			for range 120 {
				if singletons {
					driver = Gia_ManAppendAnd(graph, driver, b)
				}
				roots = append(roots, driver)
			}
			for _, root := range roots {
				Gia_ManAppendCo(graph, root)
			}
			got := isoReviewCapture(t, func() { Gia_ManIsoReduce(graph, nil, nil, 0, 0, 0, 0) })
			if strings.Count(got, "   100 finished...\r") != 1 {
				t.Fatalf("missing source hundred-output progress: %q", got)
			}
		})
	}
}
