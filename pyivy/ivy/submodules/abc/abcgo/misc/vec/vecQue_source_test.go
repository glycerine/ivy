package vec

import "testing"

func TestQueueEmptyAndPriorityReattachment(t *testing.T) {
	p := Vec_QueAlloc(4)
	if Vec_QueTop(p) != -1 || Vec_QueTopPriority(p) != -1000000000 {
		t.Fatal("empty queue sentinel differs from C")
	}
	costs := []float32{1, 2}
	Vec_QueSetPriority(p, &costs)
	Vec_QuePush(p, 0)
	Vec_QuePush(p, 1)
	costs = append(costs, 3)
	Vec_QueSetPriority(p, &costs)
	Vec_QuePush(p, 2)
	for _, want := range []int{2, 1, 0} {
		if got := Vec_QuePop(p); got != want {
			t.Fatalf("pop=%d want %d", got, want)
		}
	}
	if Vec_QueTopPriority(p) != -1000000000 {
		t.Fatal("drained queue sentinel")
	}
}
