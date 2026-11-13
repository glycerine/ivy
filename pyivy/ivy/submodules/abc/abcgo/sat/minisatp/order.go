package minisatp

// varOrder is the activity heap and deterministic random selection from
// MiniSat-p's Heap.h and VarOrder.h. Index zero is the heap sentinel.
type varOrder struct {
	heap  []int
	index []int
	seed  float64
	s     *solver
}

func newVarOrder(s *solver) varOrder {
	o := varOrder{heap: []int{-1}, index: make([]int, len(s.assign)), seed: 91648253, s: s}
	for i := range s.assign {
		o.insert(i)
	}
	return o
}

func (o *varOrder) up(i int) {
	x := o.heap[i]
	for i/2 != 0 && o.s.activity[x] > o.s.activity[o.heap[i/2]] {
		o.heap[i] = o.heap[i/2]
		o.index[o.heap[i]] = i
		i /= 2
	}
	o.heap[i], o.index[x] = x, i
}

func (o *varOrder) insert(v int) {
	o.heap = append(o.heap, v)
	o.index[v] = len(o.heap) - 1
	o.up(o.index[v])
}

func (o *varOrder) pop() int {
	r := o.heap[1]
	x := o.heap[len(o.heap)-1]
	o.heap = o.heap[:len(o.heap)-1]
	o.index[r] = 0
	if len(o.heap) == 1 {
		return r
	}
	i := 1
	for i*2 < len(o.heap) {
		child := i * 2
		if child+1 < len(o.heap) && o.s.activity[o.heap[child+1]] > o.s.activity[o.heap[child]] {
			child++
		}
		if o.s.activity[o.heap[child]] <= o.s.activity[x] {
			break
		}
		o.heap[i] = o.heap[child]
		o.index[o.heap[i]] = i
		i = child
	}
	o.heap[i], o.index[x] = x, i
	return r
}

func (o *varOrder) random() float64 {
	o.seed *= 1389796
	o.seed -= float64(int(o.seed/2147483647)) * 2147483647
	return o.seed / 2147483647
}

func (o *varOrder) selectVar() int {
	if o.random() < 0.02 && len(o.heap) > 1 {
		next := int(o.random() * float64(len(o.s.assign)))
		if o.s.assign[next] == 0 {
			return next
		}
	}
	for len(o.heap) > 1 {
		v := o.pop()
		if o.s.assign[v] == 0 {
			return v
		}
	}
	return -1
}
