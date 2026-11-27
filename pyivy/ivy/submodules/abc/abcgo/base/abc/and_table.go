package abc

// andTable is the structural hash of AND nodes: an open-addressing table
// with linear probing, keyed by the pair of signed fanin ids (andHashKey).
// It does what map[[2]int]*Object did, without Go map overhead; deletions
// shift entries back, so there are no tombstones.
type andTable struct {
	keys  [][2]int
	vals  []*Object
	count int
}

func andTableHash(k [2]int) uint64 {
	h := uint64(k[0])*0x9E3779B97F4A7C15 ^ uint64(k[1])*0xC2B2AE3D27D4EB4F
	return h ^ h>>29
}

func (t *andTable) len() int { return t.count }

// reserve sizes the table for n entries without rehashing on the way.
func (t *andTable) reserve(n int) {
	size := 16
	for size < 2*n {
		size <<= 1
	}
	if size > len(t.vals) {
		t.rehash(size)
	}
}

func (t *andTable) rehash(size int) {
	keys, vals := t.keys, t.vals
	t.keys, t.vals, t.count = make([][2]int, size), make([]*Object, size), 0
	for i, v := range vals {
		if v != nil {
			t.set(keys[i], v)
		}
	}
}

func (t *andTable) get(k [2]int) *Object {
	if t.count == 0 {
		return nil
	}
	mask := uint64(len(t.vals) - 1)
	for i := andTableHash(k) & mask; ; i = (i + 1) & mask {
		v := t.vals[i]
		if v == nil {
			return nil
		}
		if t.keys[i] == k {
			return v
		}
	}
}

func (t *andTable) set(k [2]int, o *Object) {
	if o == nil {
		t.del(k)
		return
	}
	if 2*(t.count+1) > len(t.vals) {
		t.rehash(max(16, 2*len(t.vals)))
	}
	mask := uint64(len(t.vals) - 1)
	for i := andTableHash(k) & mask; ; i = (i + 1) & mask {
		if t.vals[i] == nil {
			t.keys[i], t.vals[i] = k, o
			t.count++
			return
		}
		if t.keys[i] == k {
			t.vals[i] = o
			return
		}
	}
}

func (t *andTable) del(k [2]int) {
	if t.count == 0 {
		return
	}
	mask := uint64(len(t.vals) - 1)
	i := andTableHash(k) & mask
	for {
		if t.vals[i] == nil {
			return
		}
		if t.keys[i] == k {
			break
		}
		i = (i + 1) & mask
	}
	// backward-shift deletion
	for j := (i + 1) & mask; t.vals[j] != nil; j = (j + 1) & mask {
		home := andTableHash(t.keys[j]) & mask
		// move j to i when its home is not cyclically within (i, j]
		if (j > i && (home <= i || home > j)) || (j < i && home <= i && home > j) {
			t.keys[i], t.vals[i] = t.keys[j], t.vals[j]
			i = j
		}
	}
	t.vals[i], t.keys[i] = nil, [2]int{}
	t.count--
}

func (t *andTable) clear() {
	clear(t.vals)
	t.count = 0
}
