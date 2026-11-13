package live

import "testing"

func TestMonotoneSourceHelperResults(t *testing.T) {
	if got := checkSanityOfKnownMonotone([]int32{99}, []int32{42}, nil); got != 1 {
		t.Fatalf("C helper always returns 1, got %d", got)
	}
	if got := findHintOutputs(liveFixture(false)); got != nil {
		t.Fatalf("C helper returns NULL without hint outputs, got %v", got)
	}
}
