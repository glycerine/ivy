package aig

// edgeTable maps node IDs to edges as a dense slice: ABC keeps choice
// equivalences in the pEquivs array, and the cycle check of choice
// derivation reads them for every visited node. A nil table has no
// entries; an ID can hold InvalidEdge explicitly (present, as a map would).
type edgeTable struct {
	vals []Edge
	has  []bool
	n    int
}

func newEdgeTable() *edgeTable { return &edgeTable{} }

func (t *edgeTable) get(id int) (Edge, bool) {
	if t == nil || id < 0 || id >= len(t.has) || !t.has[id] {
		return InvalidEdge, false
	}
	return t.vals[id], true
}

// at is the entry for id, InvalidEdge when absent.
func (t *edgeTable) at(id int) Edge {
	e, _ := t.get(id)
	return e
}

func (t *edgeTable) set(id int, e Edge) {
	if id >= len(t.has) {
		n := max(id+1, 2*len(t.has))
		vals := make([]Edge, n)
		copy(vals, t.vals)
		has := make([]bool, n)
		copy(has, t.has)
		t.vals, t.has = vals, has
	}
	if !t.has[id] {
		t.has[id] = true
		t.n++
	}
	t.vals[id] = e
}

func (t *edgeTable) len() int {
	if t == nil {
		return 0
	}
	return t.n
}

// edgeTableOf builds a table from id, edge pairs (tests).
func edgeTableOf(pairs ...any) *edgeTable {
	t := newEdgeTable()
	for i := 0; i+1 < len(pairs); i += 2 {
		t.set(pairs[i].(int), pairs[i+1].(Edge))
	}
	return t
}
