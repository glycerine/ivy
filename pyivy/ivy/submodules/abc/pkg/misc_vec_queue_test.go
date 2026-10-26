package abc

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"slices"
	"testing"
)

func TestFloatQueueNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/vec_queue.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Mode                     int
		Initial, Updated, Popped []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 4 {
		t.Fatal(len(rows))
	}
	for _, r := range rows {
		q, _ := NewFloatPriorityQueue(0, 64)
		costs := make([]float32, 32)
		for i := range costs {
			costs[i] = float32(i * 7 % 11)
			if r.Mode == 1 {
				costs[i] = 1
			}
		}
		if r.Mode == 3 {
			costs[7] = float32(math.Inf(1))
			costs[2] = float32(math.Inf(-1))
			costs[31] = float32(math.Copysign(0, -1))
			costs[5] = 0
			costs[9] = float32(math.Inf(1))
			costs[1] = float32(math.Inf(-1))
		}
		if r.Mode != 0 {
			if e = q.SetPriorities(costs); e != nil {
				t.Fatal(e)
			}
		}
		for _, id := range []int{7, 2, 31, 5, 9, 1} {
			if e = q.Push(id); e != nil {
				t.Fatal(e)
			}
		}
		if !slices.Equal(q.IDs(), r.Initial) {
			t.Fatal(r.Mode, q.IDs(), r.Initial)
		}
		if r.Mode == 2 {
			q.SetCost(5, 30)
			q.SetCost(31, -4)
		}
		if !slices.Equal(q.IDs(), r.Updated) {
			t.Fatal(r.Mode, q.IDs(), r.Updated)
		}
		for _, want := range r.Popped {
			if e = q.Check(); e != nil {
				t.Fatal(e)
			}
			got, e := q.Pop()
			if e != nil || got != want {
				t.Fatal(r.Mode, want, got, e)
			}
		}
		if e = q.Check(); e != nil {
			t.Fatal(e)
		}
	}
}
func TestFloatQueueValidationAndOwnership(t *testing.T) {
	var q FloatPriorityQueue
	if e := q.Push(7); e != nil {
		t.Fatal(e)
	}
	if top, cost, ok := q.Top(); !ok || top != 7 || cost != 7 {
		t.Fatal(top, cost, ok)
	}
	q.Clear()
	if id, _, ok := q.Top(); id != -1 || ok {
		t.Fatal(id, ok)
	}
	if _, e := q.Pop(); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	p, _ := NewFloatPriorityQueue(0, 8)
	costs := []float32{1, 2, 3, 4}
	p.SetPriorities(costs)
	p.Push(1)
	p.Push(3)
	costs[1] = 1000
	if id, _, _ := p.Top(); id != 3 {
		t.Fatal("borrowed costs")
	}
	clone := p.Clone()
	clone.SetCost(1, 100)
	if id, _, _ := p.Top(); id != 3 {
		t.Fatal("aliased clone")
	}
	ids := p.IDs()
	ids[0] = 99
	if e := p.Check(); e != nil {
		t.Fatal(e)
	}
	before := p.IDs()
	for _, op := range []func() error{func() error { return p.SetPriorities([]float32{0}) }, func() error { return p.SetPriorities([]float32{0, float32(math.NaN()), 2, 3}) }, func() error { return p.SetCost(1, float32(math.NaN())) }, func() error { return p.Push(1) }, func() error { return p.Push(4) }, func() error { return p.Push(-1) }} {
		if e := op(); !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
		if !slices.Equal(before, p.IDs()) {
			t.Fatal("changed on failure")
		}
	}
	if e := p.Push(8); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if e := p.SetPriorities(make([]float32, 9)); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if e := p.SetPriorities(nil); e != nil {
		t.Fatal(e)
	}
	if id, _, _ := p.Top(); id != 3 {
		t.Fatal(id)
	}
	if e := p.SetCost(1, 2); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := NewFloatPriorityQueue(-1, 0); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := NewFloatPriorityQueue(1, -1); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if _, e := NewFloatPriorityQueue(9, 8); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	var nilQueue *FloatPriorityQueue
	if nilQueue.Len() != 0 || nilQueue.Contains(0) || nilQueue.Clone() != nil {
		t.Fatal("nil access")
	}
	if e := nilQueue.Push(0); !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	p.Clear()
	if e := p.Check(); e != nil {
		t.Fatal(e)
	}
	p.Push(7)
	p.order[7] = 0
	if e := p.Check(); !errors.Is(e, ErrInvariant) {
		t.Fatal(e)
	}
}

func TestFloatQueueFormat(t *testing.T) {
	q, err := NewFloatPriorityQueue(0, 16)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{1, 3, 2} {
		if err = q.Push(id); err != nil {
			t.Fatal(err)
		}
	}
	want := "3    \n3    \n\n1    2    \n1    2    \n\n"
	if got := q.Format(); got != want {
		t.Fatalf("format mismatch\n got %q\nwant %q", got, want)
	}
	if (*FloatPriorityQueue)(nil).Format() != "" {
		t.Fatal("nil format")
	}
}
func FuzzFloatQueue(f *testing.F) {
	f.Add([]byte{0, 7, 0, 2, 0, 31, 2, 7, 1, 0})
	f.Add([]byte{4, 1, 0, 5, 2, 5, 1, 0})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 512 {
			data = data[:512]
		}
		q, _ := NewFloatPriorityQueue(0, 64)
		present := make([]bool, 64)
		costs := make([]float32, 64)
		custom := false
		check := func() {
			t.Helper()
			if e := q.Check(); e != nil {
				t.Fatal(e)
			}
			count := 0
			maximum := float32(math.Inf(-1))
			for id, yes := range present {
				if q.Contains(id) != yes {
					t.Fatal("membership")
				}
				if yes {
					count++
					value := float32(id)
					if custom {
						value = costs[id]
					}
					if value > maximum {
						maximum = value
					}
				}
			}
			id, priority, ok := q.Top()
			if q.Len() != count || ok != (count > 0) {
				t.Fatal("size")
			}
			if ok && (!present[id] || priority != maximum) {
				t.Fatal("maximum", id, priority, maximum)
			}
		}
		for i := 0; i+1 < len(data); i += 2 {
			op, id := data[i]%7, int(data[i+1]%64)
			switch op {
			case 0:
				if !present[id] {
					if e := q.Push(id); e != nil {
						t.Fatal(e)
					}
					present[id] = true
				}
			case 1:
				if q.Len() > 0 {
					top, _, _ := q.Top()
					got, e := q.Pop()
					if e != nil || got != top {
						t.Fatal(e)
					}
					present[got] = false
				}
			case 2:
				if custom {
					costs[id] = float32(int8(data[i]))
					if e := q.SetCost(id, costs[id]); e != nil {
						t.Fatal(e)
					}
				}
			case 3:
				q.Clear()
				clear(present)
			case 4:
				custom = true
				for j := range costs {
					costs[j] = float32(int8(data[(i+j)%len(data)]))
				}
				if e := q.SetPriorities(costs); e != nil {
					t.Fatal(e)
				}
			case 5:
				custom = false
				q.SetPriorities(nil)
			case 6:
				clone := q.Clone()
				if e := clone.Check(); e != nil {
					t.Fatal(e)
				}
				clone.Clear()
				if !slices.Equal(q.IDs(), q.Clone().IDs()) {
					t.Fatal("clone")
				}
			}
			check()
		}
	})
}
