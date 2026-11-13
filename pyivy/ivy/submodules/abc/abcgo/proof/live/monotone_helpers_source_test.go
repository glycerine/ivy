package live

import (
	"fmt"
	"strings"
	"testing"
)

func TestMonotoneSourceHelperResults(t *testing.T) {
	if got := checkSanityOfKnownMonotone([]int32{99}, []int32{42}, nil); got != 1 {
		t.Fatalf("C helper always returns 1, got %d", got)
	}
	if got := findHintOutputs(liveFixture(false)); got != nil {
		t.Fatalf("C helper returns NULL without hint outputs, got %v", got)
	}
}

func TestSourceKnownMonotoneUsesOriginalInvalidOffset(t *testing.T) {
	n := liveFixture(false)
	indexes := &aigPoIndices{0, 1, 1, -1}
	vectors := &monotoneVectorsStruct{attrKnownMonotone: []int32{1}, attrHintMonotone: []int32{1}}
	defer func() {
		r := recover()
		if r == nil || !strings.Contains(fmt.Sprint(r), "index out of range [-1]") {
			t.Fatalf("expected C's negative delayed-flop index, got %v", r)
		}
	}()
	createMonotoneTester(n.Graph, indexes, vectors, nil)
}
