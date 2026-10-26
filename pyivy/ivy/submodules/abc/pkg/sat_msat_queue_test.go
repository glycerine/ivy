package abc

import (
	"errors"
	"slices"
	"testing"
)

func TestMsatQueueFIFOEpochAndCapacity(t *testing.T) {
	queue, err := NewMsatQueue(3)
	if err != nil || queue.Len() != 0 || queue.Capacity() != 3 || queue.Pending() == nil {
		t.Fatal(queue, err)
	}
	for _, literal := range []int{7, -1, 12} {
		if err = queue.Insert(literal); err != nil {
			t.Fatal(err)
		}
	}
	if queue.Len() != 3 || !slices.Equal(queue.Pending(), []int{7, -1, 12}) {
		t.Fatal(queue.Len(), queue.Pending())
	}
	for index, want := range []int{7, -1, 12} {
		got, found, extractErr := queue.Extract()
		if extractErr != nil || !found || got != want || queue.Len() != 2-index {
			t.Fatal(index, got, found, extractErr, queue.Len())
		}
	}
	if got, found, err := queue.Extract(); err != nil || found || got != 0 {
		t.Fatal("empty", got, found, err)
	}
	if err = queue.Insert(99); !errors.Is(err, ErrLimit) {
		t.Fatal("extracted slots reused before clear", err)
	}
	if err = queue.Clear(); err != nil || queue.Len() != 0 || queue.Capacity() != 3 {
		t.Fatal(err, queue.Len(), queue.Capacity())
	}
	if err = queue.Insert(99); err != nil {
		t.Fatal(err)
	}
	if got, found, err := queue.Extract(); err != nil || !found || got != 99 {
		t.Fatal(got, found, err)
	}
}

func TestMsatQueueOwnershipLifecycleAndErrors(t *testing.T) {
	if queue, err := NewMsatQueue(-1); queue != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(queue, err)
	}
	empty, err := NewMsatQueue(0)
	if err != nil {
		t.Fatal(err)
	}
	if err = empty.Insert(1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	queue, _ := NewMsatQueue(2)
	queue.Insert(3)
	queue.Insert(4)
	pending := queue.Pending()
	pending[0] = 9
	if !slices.Equal(queue.Pending(), []int{3, 4}) {
		t.Fatal("pending aliases queue")
	}
	queue.Close()
	queue.Close()
	if queue.Len() != 0 || queue.Capacity() != 0 || queue.Pending() != nil {
		t.Fatal("closed queue exposes storage")
	}
	if err = queue.Insert(1); !errors.Is(err, ErrInvariant) {
		t.Fatal("closed insert", err)
	}
	if _, _, err = queue.Extract(); !errors.Is(err, ErrInvariant) {
		t.Fatal("closed extract", err)
	}
	if err = queue.Clear(); !errors.Is(err, ErrInvariant) {
		t.Fatal("closed clear", err)
	}
	if err = (*MsatQueue)(nil).Insert(1); !errors.Is(err, ErrDimension) {
		t.Fatal("nil insert", err)
	}
}

func FuzzMsatQueueOperations(f *testing.F) {
	for _, seed := range [][]byte{{}, {0}, {1, 7, 2, 1}, {1, 255, 1, 2, 0, 3}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, operations []byte) {
		const capacity = 8
		queue, err := NewMsatQueue(capacity)
		if err != nil {
			t.Fatal(err)
		}
		storage := make([]int, capacity)
		first, last := 0, 0
		for _, operation := range operations {
			switch operation % 3 {
			case 0:
				literal := int(int8(operation))
				err = queue.Insert(literal)
				if last == capacity {
					if !errors.Is(err, ErrLimit) {
						t.Fatal("insert limit", err)
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					storage[last], last = literal, last+1
				}
			case 1:
				got, found, extractErr := queue.Extract()
				wantFound := first < last
				want := 0
				if wantFound {
					want, first = storage[first], first+1
				}
				if extractErr != nil || found != wantFound || got != want {
					t.Fatal("extract", got, found, extractErr, want, wantFound)
				}
			case 2:
				if err = queue.Clear(); err != nil {
					t.Fatal(err)
				}
				clear(storage)
				first, last = 0, 0
			}
			if queue.Len() != last-first || !slices.Equal(queue.Pending(), storage[first:last]) {
				t.Fatal(queue.Len(), queue.Pending(), first, last, storage[first:last])
			}
		}
	})
}
