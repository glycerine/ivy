package msat

import "testing"

func TestQueueDoesNotRecycleUntilCleared(t *testing.T) {
	q := Msat_QueueAlloc(1)
	Msat_QueueInsert(q, 0)
	if Msat_QueueExtract(q) != 0 || Msat_QueueExtract(q) != -1 {
		t.Fatal("extract")
	}
	defer func() {
		if recover() == nil {
			t.Fatal("C asserts when fixed queue reaches capacity, even after extraction")
		}
	}()
	Msat_QueueInsert(q, 1)
}

func TestQueueAcceptsNegativeLiteral(t *testing.T) {
	q := Msat_QueueAlloc(1)
	Msat_QueueInsert(q, -2)
	if got := Msat_QueueExtract(q); got != -2 {
		t.Fatalf("literal = %d", got)
	}
}
