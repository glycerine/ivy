package aig

// strashTable is the structural hash table: an open-addressing map from an
// ordered fanin pair to its node, with linear probing and backward-shift
// deletion. It replaces a Go map on the hottest path of AIG construction.
type strashTable struct {
	slots []strashSlot // power-of-two length, or nil
	n     int
}

type strashSlot struct {
	key   edgePair // key.a == InvalidEdge marks an empty slot
	value Edge
}

func strashHash(k edgePair) uint32 {
	h := uint32(k.a)*0x9E3779B1 ^ uint32(k.b)*0x85EBCA77
	return h ^ h>>15
}

func (t *strashTable) len() int { return t.n }

func (t *strashTable) get(k edgePair) (Edge, bool) {
	if t.n == 0 {
		return 0, false
	}
	mask := uint32(len(t.slots) - 1)
	for i := strashHash(k) & mask; ; i = (i + 1) & mask {
		s := &t.slots[i]
		if s.key == k {
			return s.value, true
		}
		if s.key.a == InvalidEdge {
			return 0, false
		}
	}
}

func (t *strashTable) set(k edgePair, v Edge) {
	if 2*(t.n+1) > len(t.slots) {
		t.grow(max(64, 2*len(t.slots)))
	}
	mask := uint32(len(t.slots) - 1)
	for i := strashHash(k) & mask; ; i = (i + 1) & mask {
		s := &t.slots[i]
		if s.key == k {
			s.value = v
			return
		}
		if s.key.a == InvalidEdge {
			s.key, s.value = k, v
			t.n++
			return
		}
	}
}

func (t *strashTable) del(k edgePair) {
	if t.n == 0 {
		return
	}
	mask := uint32(len(t.slots) - 1)
	i := strashHash(k) & mask
	for t.slots[i].key != k {
		if t.slots[i].key.a == InvalidEdge {
			return
		}
		i = (i + 1) & mask
	}
	// Shift back later members of the probe run into the hole.
	for j := (i + 1) & mask; t.slots[j].key.a != InvalidEdge; j = (j + 1) & mask {
		home := strashHash(t.slots[j].key) & mask
		if (j-home)&mask >= (j-i)&mask {
			t.slots[i] = t.slots[j]
			i = j
		}
	}
	t.slots[i] = strashSlot{key: edgePair{InvalidEdge, InvalidEdge}}
	t.n--
}

func (t *strashTable) clear() {
	for i := range t.slots {
		t.slots[i] = strashSlot{key: edgePair{InvalidEdge, InvalidEdge}}
	}
	t.n = 0
}

// reserve sizes the table for n entries without rehashing on the way.
func (t *strashTable) reserve(n int) {
	if size := 64; 2*n > len(t.slots) {
		for size < 2*n {
			size *= 2
		}
		t.grow(size)
	}
}

func (t *strashTable) grow(size int) {
	old := t.slots
	t.slots = make([]strashSlot, size)
	t.n = 0
	for i := range t.slots {
		t.slots[i].key = edgePair{InvalidEdge, InvalidEdge}
	}
	for _, s := range old {
		if s.key.a != InvalidEdge {
			t.set(s.key, s.value)
		}
	}
}
