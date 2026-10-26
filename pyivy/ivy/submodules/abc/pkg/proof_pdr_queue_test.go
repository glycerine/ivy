package abc

import (
	"slices"
	"testing"
)

func TestPdrQueueOrderAndOwnership(t *testing.T) {
	cube := pdrTestCube(t, []PdrLit{0}, []PdrLit{1})
	var q, other PdrQueue
	rows := [][2]int{{2, 10}, {0, 10}, {1, 4}, {1, 4}, {1, -1}, {2, 0}}
	entries := make([]*PdrObligation, len(rows))
	for i, row := range rows {
		o, err := NewPdrObligation(row[0], row[1], cube, nil)
		if err != nil {
			t.Fatal(err)
		}
		entries[i] = o
		if err := q.Push(o); err != nil {
			t.Fatal(err)
		}
		if err := q.Check(); err != nil {
			t.Fatal(err)
		}
		if q.Push(o) == nil || other.Push(o) == nil || o.Reschedule(1, 1) == nil {
			t.Fatal("queued obligation mutated or linked twice")
		}
	}
	for _, index := range []int{1, 4, 2, 3, 5, 0} {
		if q.Peek() != entries[index] {
			t.Fatal("priority order or equal-key stability")
		}
		if o := q.Pop(); o != entries[index] || o.owner != nil || o.link != nil {
			t.Fatal("pop ownership")
		}
	}
	if !q.Empty() || q.Pop() != nil || q.Len() != 0 || q.MaximumLen() != 6 || q.PushCount() != 6 {
		t.Fatal("queue counters")
	}
	o := entries[0]
	if err := o.Reschedule(3, -2); err != nil {
		t.Fatal(err)
	}
	if err := other.Push(o); err != nil {
		t.Fatal(err)
	}
	other.Clear()
	if o.owner != nil || other.Len() != 0 {
		t.Fatal("clear ownership")
	}
	next, err := NewPdrObligation(2, 0, cube, nil)
	if err != nil {
		t.Fatal(err)
	}
	predecessor, err := NewPdrObligation(1, -1, cube, next)
	if err != nil {
		t.Fatal(err)
	}
	q.Push(predecessor)
	q.Clear()
	if predecessor.Next() != next || predecessor.State() != cube || next.State() != cube {
		t.Fatal("clear destroyed witness chain")
	}
	if q.Push(nil) == nil {
		t.Fatal("nil obligation")
	}
	if _, err := NewPdrObligation(-1, 0, cube, nil); err == nil {
		t.Fatal("negative frame")
	}
	if _, err := NewPdrObligation(0, 0, nil, nil); err == nil {
		t.Fatal("nil cube")
	}
}
func TestPdrQueueInvariantDetection(t *testing.T) {
	cube := pdrTestCube(t, nil, nil)
	o, _ := NewPdrObligation(1, 0, cube, nil)
	var q PdrQueue
	q.Push(o)
	o.link = o
	if q.Check() == nil {
		t.Fatal("queue cycle")
	}
	o.link = nil
	o.owner = nil
	if q.Check() == nil {
		t.Fatal("owner mismatch")
	}
	o.owner = &q
	q.length++
	if q.Check() == nil {
		t.Fatal("queue count")
	}
	q.length--
	if q.Check() != nil {
		t.Fatal("restored queue")
	}
}
func FuzzPdrQueue(f *testing.F) {
	f.Add([]byte{2, 3, 1, 4, 0, 5, 1, 4})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			return
		}
		cube := pdrTestCube(t, nil, nil)
		var q PdrQueue
		model := make([]*PdrObligation, 0)
		for i := 0; i+1 < len(data); i += 2 {
			o, err := NewPdrObligation(int(data[i]%8), int(int8(data[i+1])), cube, nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := q.Push(o); err != nil {
				t.Fatal(err)
			}
			model = append(model, o)
		}
		slices.SortStableFunc(model, func(a, b *PdrObligation) int {
			if a.frame < b.frame {
				return -1
			}
			if a.frame > b.frame {
				return 1
			}
			if a.priority < b.priority {
				return -1
			}
			if a.priority > b.priority {
				return 1
			}
			return 0
		})
		for _, expected := range model {
			if q.Pop() != expected {
				t.Fatal("priority model disagreement")
			}
			if err := q.Check(); err != nil {
				t.Fatal(err)
			}
		}
		if !q.Empty() {
			t.Fatal("leftover obligation")
		}
	})
}
