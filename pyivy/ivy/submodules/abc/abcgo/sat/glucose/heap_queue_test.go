package glucose

import (
	"math/rand"
	"testing"
)

func TestHeapMutablePriorities(t *testing.T) {
	priorities := make([]int, 100)
	h := NewHeap(func(a, b int) bool { return priorities[a] < priorities[b] })
	rng := rand.New(rand.NewSource(17))
	for i := range priorities {
		priorities[i] = rng.Intn(1000)
		h.Insert(i)
	}
	for step := 0; step < 1000; step++ {
		i := rng.Intn(100)
		priorities[i] = rng.Intn(1000)
		h.Update(i)
	}
	previous := -1
	for !h.Empty() {
		i := h.RemoveMin()
		if priorities[i] < previous {
			t.Fatal("priority order", i)
		}
		previous = priorities[i]
		if h.InHeap(i) {
			t.Fatal("removed member retained")
		}
	}
	h.Build([]int{99, 30, 45})
	if h.Size() != 3 {
		t.Fatal(h.Size())
	}
	h.Clear()
	for _, i := range []int{99, 30, 45} {
		if h.InHeap(i) {
			t.Fatal(i)
		}
	}
}

func TestBoundedQueueAgainstSlice(t *testing.T) {
	q := NewBoundedQueue(7)
	var reference []uint64
	for i := uint64(0); i < 100; i++ {
		q.Push(i)
		reference = append(reference, i)
		if len(reference) > 7 {
			reference = reference[1:]
		}
		if i%3 == 0 {
			q.Pop()
			reference = reference[1:]
		}
		var sum uint64
		for _, v := range reference {
			sum += v
		}
		if q.Sum() != sum || q.Size() != len(reference) {
			t.Fatal(i, q.Sum(), sum)
		}
		if len(reference) > 0 && (q.Peek() != reference[0] || q.Avg() != sum/uint64(len(reference))) {
			t.Fatal("queue order/average", i)
		}
	}
	q.GrowTo(3)
	if q.Sum() != 0 || q.Size() != 0 {
		t.Fatal("resize retained sum")
	}
	q.Push(9)
	q.FastClear()
	q.Push(4)
	if q.Sum() != 4 || q.Peek() != 4 {
		t.Fatal("clear retained sample")
	}
}
