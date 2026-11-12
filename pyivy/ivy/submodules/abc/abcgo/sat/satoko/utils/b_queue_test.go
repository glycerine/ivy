package utils

import "testing"

func TestBoundedQueueRollingAverage(t *testing.T) {
	p := NewBoundedQueue(3)
	for i := uint32(1); i <= 20; i++ {
		p.Push(i)
		want := i - 1
		if i == 1 {
			want = 1
		}
		if p.Average() != want {
			t.Fatalf("after %d average %d want %d", i, p.Average(), want)
		}
	}
	if !p.Full() {
		t.Fatal("not full")
	}
	p.Clear()
	if p.Full() || p.sum != 0 || p.size != 0 {
		t.Fatal(p)
	}
	p.Push(^uint32(0))
	p.Push(^uint32(0))
	if p.Average() != ^uint32(0) {
		t.Fatal("sum overflow")
	}
}
