package gia

// jfCutAccounting tracks temporary changes to the mapping references. best
// returns the current best packed cut for an object ID.
type jfCutAccounting struct {
	refs    []int32
	best    func(int32) []int32
	touched []int32
}

// ref adjusts reference counts and returns the area of newly referenced or
// dereferenced cones. With record set, touched IDs are kept for restoration.
func (a *jfCutAccounting) ref(cut []int32, add, edges, record bool) int32 {
	count := Jf_CutCost(cut)
	if edges {
		count = count<<4 | Jf_CutSize(cut)
	}
	for i := int32(1); i <= Jf_CutSize(cut); i++ {
		id := Jf_CutVar(cut, i)
		// The original Jf_CutForEachVar macro stops when its assigned ID is zero.
		if id == 0 {
			break
		}
		old := a.refs[id]
		if add {
			a.refs[id]++
		} else {
			a.refs[id]--
		}
		crossed := add && old == 0 || !add && a.refs[id] == 0
		if crossed {
			next := a.best(id)
			if !Jf_CutIsTriv(next, id) {
				count += a.ref(next, add, edges, record)
			}
		}
		if record {
			a.touched = append(a.touched, id)
		}
	}
	return count
}
func (a *jfCutAccounting) area(cut []int32, edges bool) int32 {
	a.touched = a.touched[:0]
	result := a.ref(cut, true, edges, true)
	for _, id := range a.touched {
		a.refs[id]--
	}
	return result
}
func (a *jfCutAccounting) checkMffcRec(cut []int32, limit int) bool {
	for i := int32(1); i <= Jf_CutSize(cut); i++ {
		id := Jf_CutVar(cut, i)
		if id == 0 {
			break
		}
		a.refs[id]--
		recurse := false
		var next []int32
		if a.refs[id] == 0 {
			next = a.best(id)
			recurse = !Jf_CutIsTriv(next, id)
		}
		a.touched = append(a.touched, id)
		if len(a.touched) >= limit {
			return false
		}
		if recurse && !a.checkMffcRec(next, limit) {
			return false
		}
	}
	return true
}
func (a *jfCutAccounting) checkMffc(cut []int32, limit int) bool {
	a.touched = a.touched[:0]
	result := a.checkMffcRec(cut, limit)
	for _, id := range a.touched {
		a.refs[id]++
	}
	return result
}
