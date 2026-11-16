package gia

import (
	"fmt"
	"slices"
	"strings"
	"testing"
)

func TestIsoFullRefinementPreservesFinalStabilityRound(t *testing.T) {
	graph := Gia_ManStart(5)
	a, b := Gia_ManAppendCi(graph), Gia_ManAppendCi(graph)
	root := Gia_ManAppendAnd(graph, a, b^1)
	Gia_ManAppendCo(graph, root)
	got := isoReviewCapture(t, func() { Gia_IsoDeriveEquivPos(graph, 0, 1) })
	var iterations []int
	for _, line := range strings.Split(got, "\n") {
		if !strings.HasPrefix(line, "Iter ") {
			continue
		}
		var iteration int
		if _, err := fmt.Sscanf(line, "Iter %d", &iteration); err != nil {
			t.Fatal(err)
		}
		iterations = append(iterations, iteration)
	}
	// C runs two forward passes, splits the CI class in the first reverse
	// pass, settles with two more reverse passes, then confirms stability
	// with two forward and two reverse passes. Its final snapshot is i+2.
	want := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 11}
	if !slices.Equal(iterations, want) {
		t.Fatalf("refinement iterations%v want source trace%v", iterations, want)
	}
}
