package abc

// andTable is the structural hash of AND nodes: an open-addressing table
// with linear probing, keyed by the pair of signed fanin ids (andHashKey).
// It does what map[[2]int]*Object did, without Go map overhead; deletions
// shift entries back, so there are no tombstones.
//
// Each slot keeps its key next to its node (16 bytes, four slots to a cache
// line), so a probe touches one line; separate key and node arrays cost two
// misses per probe. Keys are stored as int32 halves: object IDs, and their
// complements for inverted edges, fit in 32 bits.
type andTable struct {
	slots []andSlot
	count int
}

type andSlot struct {
	a, b int32
	v    *Object
}

func andTableHash(a, b int32) uint64 {
	h := uint64(int64(a))*0x9E3779B97F4A7C15 ^ uint64(int64(b))*0xC2B2AE3D27D4EB4F
	return h ^ h>>29
}

func andTableKey(k [2]int) (int32, int32) {
	a, b := int32(k[0]), int32(k[1])
	if int(a) != k[0] || int(b) != k[1] {
		panic("abc: AND hash key exceeds 32 bits")
	}
	return a, b
}

func (t *andTable) len() int { return t.count }

// reserve sizes the table for n entries without rehashing on the way.
func (t *andTable) reserve(n int) {
	size := 16
	for size < 2*n {
		size <<= 1
	}
	if size > len(t.slots) {
		t.rehash(size)
	}
}

func (t *andTable) rehash(size int) {
	old := t.slots
	t.slots, t.count = make([]andSlot, size), 0
	for _, s := range old {
		if s.v != nil {
			t.insert(s.a, s.b, s.v)
		}
	}
}

func (t *andTable) get(k [2]int) *Object {
	if t.count == 0 {
		return nil
	}
	a, b := andTableKey(k)
	slots := t.slots
	mask := uint64(len(slots) - 1)
	for i := andTableHash(a, b) & mask; ; i = (i + 1) & mask {
		s := &slots[i]
		if s.v == nil {
			return nil
		}
		if s.a == a && s.b == b {
			return s.v
		}
	}
}

func (t *andTable) set(k [2]int, o *Object) {
	if o == nil {
		t.del(k)
		return
	}
	a, b := andTableKey(k)
	if 2*(t.count+1) > len(t.slots) {
		t.rehash(max(16, 2*len(t.slots)))
	}
	t.insert(a, b, o)
}

func (t *andTable) insert(a, b int32, o *Object) {
	slots := t.slots
	mask := uint64(len(slots) - 1)
	for i := andTableHash(a, b) & mask; ; i = (i + 1) & mask {
		s := &slots[i]
		if s.v == nil {
			*s = andSlot{a, b, o}
			t.count++
			return
		}
		if s.a == a && s.b == b {
			s.v = o
			return
		}
	}
}

func (t *andTable) del(k [2]int) {
	if t.count == 0 {
		return
	}
	a, b := andTableKey(k)
	slots := t.slots
	mask := uint64(len(slots) - 1)
	i := andTableHash(a, b) & mask
	for {
		if slots[i].v == nil {
			return
		}
		if slots[i].a == a && slots[i].b == b {
			break
		}
		i = (i + 1) & mask
	}
	// backward-shift deletion
	for j := (i + 1) & mask; slots[j].v != nil; j = (j + 1) & mask {
		home := andTableHash(slots[j].a, slots[j].b) & mask
		// move j to i when its home is not cyclically within (i, j]
		if (j > i && (home <= i || home > j)) || (j < i && home <= i && home > j) {
			slots[i] = slots[j]
			i = j
		}
	}
	slots[i] = andSlot{}
	t.count--
}

func (t *andTable) clear() {
	clear(t.slots)
	t.count = 0
}
