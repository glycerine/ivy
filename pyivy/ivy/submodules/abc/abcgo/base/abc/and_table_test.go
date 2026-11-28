package abc

import (
	"math/rand"
	"testing"
)

// The interleaved table must behave as a map under insert, overwrite and
// backward-shift deletion, including complemented (negative) key halves.
func TestAndTableMatchesMap(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	var table andTable
	want := map[[2]int]*Object{}
	objects := make([]*Object, 64)
	for i := range objects {
		objects[i] = &Object{ID: i}
	}
	key := func() [2]int {
		a, b := r.Intn(300), r.Intn(300)
		if r.Intn(2) == 0 {
			a = ^a
		}
		if r.Intn(2) == 0 {
			b = ^b
		}
		return [2]int{a, b}
	}
	for step := 0; step < 200000; step++ {
		k := key()
		switch r.Intn(3) {
		case 0, 1:
			o := objects[r.Intn(len(objects))]
			table.set(k, o)
			want[k] = o
		default:
			table.del(k)
			delete(want, k)
		}
		if step%1000 == 0 {
			table.reserve(r.Intn(4000))
		}
		probe := key()
		if got := table.get(probe); got != want[probe] {
			t.Fatalf("step %d: get(%v) = %v, want %v", step, probe, got, want[probe])
		}
		if table.len() != len(want) {
			t.Fatalf("step %d: len %d, want %d", step, table.len(), len(want))
		}
	}
	for k, o := range want {
		if table.get(k) != o {
			t.Fatalf("final get(%v) wrong", k)
		}
	}
}
